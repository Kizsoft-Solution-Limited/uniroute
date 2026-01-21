package integration

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/Kizsoft-Solution-Limited/uniroute/internal/api/middleware"
	"github.com/Kizsoft-Solution-Limited/uniroute/internal/security"
	"github.com/Kizsoft-Solution-Limited/uniroute/internal/storage"
	"github.com/Kizsoft-Solution-Limited/uniroute/tests/testutil"
)

// getTestPostgresClient returns a PostgreSQL client for integration testing
func getTestPostgresClient(t *testing.T) *storage.PostgresClient {
	config := testutil.GetTestConfig(t)
	logger := zerolog.Nop()
	
	client, err := storage.NewPostgresClient(config.PostgresURL, logger)
	if err != nil {
		t.Skipf("Skipping integration test: PostgreSQL not available: %v", err)
	}
	
	// Run migrations
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	pool := client.Pool()
	_, err = pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			email VARCHAR(255) UNIQUE NOT NULL,
			password_hash VARCHAR(255) NOT NULL,
			created_at TIMESTAMP DEFAULT NOW(),
			updated_at TIMESTAMP DEFAULT NOW()
		);
		
		CREATE TABLE IF NOT EXISTS api_keys (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id UUID REFERENCES users(id) ON DELETE CASCADE,
			lookup_hash VARCHAR(64) UNIQUE NOT NULL,
			verification_hash TEXT NOT NULL,
			name VARCHAR(255),
			rate_limit_per_minute INTEGER DEFAULT 60,
			rate_limit_per_day INTEGER DEFAULT 10000,
			created_at TIMESTAMP DEFAULT NOW(),
			expires_at TIMESTAMP,
			is_active BOOLEAN DEFAULT true
		);
		
		CREATE INDEX IF NOT EXISTS idx_api_keys_lookup_hash ON api_keys(lookup_hash);
	`)
	if err != nil {
		t.Fatalf("Failed to create test tables: %v", err)
	}
	
	pool.Exec(ctx, "TRUNCATE TABLE api_keys CASCADE")
	pool.Exec(ctx, "TRUNCATE TABLE users CASCADE")
	
	return client
}

// getTestRedisClient returns a Redis client for integration testing
func getTestRedisClient(t *testing.T) *storage.RedisClient {
	config := testutil.GetTestConfig(t)
	logger := zerolog.Nop()
	
	client, err := storage.NewRedisClient(config.RedisURL, logger)
	if err != nil {
		t.Skipf("Skipping integration test: Redis not available: %v", err)
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	client.Client().FlushDB(ctx)
	
	return client
}

// TestAuthMiddleware_Integration tests API key authentication with real database
func TestAuthMiddleware_Integration(t *testing.T) {
	testutil.SkipIfShort(t)
	gin.SetMode(gin.TestMode)
	
	postgresClient := getTestPostgresClient(t)
	defer postgresClient.Close()
	
	apiKeyRepo := storage.NewAPIKeyRepository(postgresClient.Pool())
	apiKeyService := security.NewAPIKeyServiceV2(apiKeyRepo, "test-secret-key")
	ctx := context.Background()
	
	// Create an API key
	userID := uuid.New()
	key, _, err := apiKeyService.CreateAPIKey(ctx, userID, "Integration Test", 60, 10000, nil)
	if err != nil {
		t.Fatalf("CreateAPIKey failed: %v", err)
	}
	
	// Test with valid key
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)
	c.Request.Header.Set("Authorization", "Bearer "+key)
	
	middleware := middleware.AuthMiddleware(apiKeyService)
	middleware(c)
	
	if c.Writer.Status() == http.StatusUnauthorized {
		t.Error("Request with valid API key should be authorized")
	}
	
	// Check context values
	apiKeyID, exists := c.Get("api_key_id")
	if !exists {
		t.Error("api_key_id should be set in context")
	}
	if apiKeyID == "" {
		t.Error("api_key_id should not be empty")
	}
	
	// Test with invalid key
	w2 := httptest.NewRecorder()
	c2, _ := gin.CreateTestContext(w2)
	c2.Request = httptest.NewRequest("GET", "/test", nil)
	c2.Request.Header.Set("Authorization", "Bearer ur_invalidkey123456789012345678901234567890")
	
	middleware(c2)
	
	if c2.Writer.Status() != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, c2.Writer.Status())
	}
}

// TestRateLimitMiddleware_Integration tests rate limiting with real Redis
func TestRateLimitMiddleware_Integration(t *testing.T) {
	testutil.SkipIfShort(t)
	gin.SetMode(gin.TestMode)
	
	redisClient := getTestRedisClient(t)
	defer redisClient.Close()
	
	rateLimiter := security.NewRateLimiter(redisClient)
	limitPerMinute := 3
	limitPerDay := 100
	
	// Create middleware with fixed limits
	middleware := middleware.RateLimitMiddleware(rateLimiter, func(identifier string) (int, int) {
		return limitPerMinute, limitPerDay
	})
	
	// Make requests up to the limit
	for i := 0; i < limitPerMinute; i++ {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/test", nil)
		
		middleware(c)
		
		if c.Writer.Status() == http.StatusTooManyRequests {
			t.Errorf("Request %d should be allowed", i+1)
		}
	}
	
	// Next request should exceed limit
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)
	
	middleware(c)
	
	if c.Writer.Status() != http.StatusTooManyRequests {
		t.Errorf("Expected status %d, got %d", http.StatusTooManyRequests, c.Writer.Status())
	}
	
	// Check rate limit headers
	remaining := c.Writer.Header().Get("X-RateLimit-Remaining-PerMinute")
	if remaining == "" {
		t.Error("X-RateLimit-Remaining-PerMinute header should be present")
	}
}

// TestFullMiddlewareFlow_Integration tests the complete middleware chain
func TestFullMiddlewareFlow_Integration(t *testing.T) {
	testutil.SkipIfShort(t)
	gin.SetMode(gin.TestMode)
	
	postgresClient := getTestPostgresClient(t)
	defer postgresClient.Close()
	
	redisClient := getTestRedisClient(t)
	defer redisClient.Close()
	
	// Setup services
	apiKeyRepo := storage.NewAPIKeyRepository(postgresClient.Pool())
	apiKeyService := security.NewAPIKeyServiceV2(apiKeyRepo, "test-secret-key")
	rateLimiter := security.NewRateLimiter(redisClient)
	
	ctx := context.Background()
	userID := uuid.New()
	
	// Create API key
	key, apiKey, err := apiKeyService.CreateAPIKey(ctx, userID, "Full Flow Test", 5, 100, nil)
	if err != nil {
		t.Fatalf("CreateAPIKey failed: %v", err)
	}
	
	// Create router with middleware
	r := gin.New()
	r.Use(middleware.SecurityHeadersMiddleware())
	
	api := r.Group("/v1")
	api.Use(middleware.AuthMiddleware(apiKeyService))
	api.Use(middleware.RateLimitMiddleware(rateLimiter, func(identifier string) (int, int) {
		// Use the API key's limits
		return apiKey.RateLimitPerMinute, apiKey.RateLimitPerDay
	}))
	
	api.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})
	
	// Test successful request
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/v1/test", nil)
	req.Header.Set("Authorization", "Bearer "+key)
	r.ServeHTTP(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
	
	// Test rate limit exceeded
	for i := 0; i < apiKey.RateLimitPerMinute; i++ {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/v1/test", nil)
		req.Header.Set("Authorization", "Bearer "+key)
		r.ServeHTTP(w, req)
	}
	
	// Next request should be rate limited
	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest("GET", "/v1/test", nil)
	req2.Header.Set("Authorization", "Bearer "+key)
	r.ServeHTTP(w2, req2)
	
	if w2.Code != http.StatusTooManyRequests {
		t.Errorf("Expected status %d, got %d", http.StatusTooManyRequests, w2.Code)
	}
}

