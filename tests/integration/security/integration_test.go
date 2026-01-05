package integration

import (
	"context"
	"testing"
	"time"

	"github.com/Kizsoft-Solution-Limited/uniroute/internal/storage"
	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/Kizsoft-Solution-Limited/uniroute/internal/security"
)

// getTestRedisClient returns a Redis client for integration testing
func getTestRedisClient(t *testing.T) *storage.RedisClient {
	redisURL := "redis://localhost:6379/15" // Use DB 15 for testing
	logger := zerolog.Nop()

	client, err := storage.NewRedisClient(redisURL, logger)
	if err != nil {
		t.Skipf("Skipping integration test: Redis not available at %s: %v", redisURL, err)
	}

	// Clear test database
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	client.Client().FlushDB(ctx)

	return client
}

// getTestPostgresClient returns a PostgreSQL client for integration testing
func getTestPostgresClient(t *testing.T) *storage.PostgresClient {
	databaseURL := "postgres://postgres:postgres@localhost/uniroute_test?sslmode=disable"
	logger := zerolog.Nop()

	client, err := storage.NewPostgresClient(databaseURL, logger)
	if err != nil {
		t.Skipf("Skipping integration test: PostgreSQL not available at %s: %v", databaseURL, err)
	}

	// Run migrations
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Create tables if they don't exist
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
		CREATE INDEX IF NOT EXISTS idx_api_keys_user_id ON api_keys(user_id);
		CREATE INDEX IF NOT EXISTS idx_api_keys_is_active ON api_keys(is_active);
	`)
	if err != nil {
		t.Fatalf("Failed to create test tables: %v", err)
	}

	// Clear test data
	pool.Exec(ctx, "TRUNCATE TABLE api_keys CASCADE")
	pool.Exec(ctx, "TRUNCATE TABLE users CASCADE")

	return client
}

// TestAPIKeyServiceV2_Integration tests API key service with real PostgreSQL
func TestAPIKeyServiceV2_Integration(t *testing.T) {
	postgresClient := getTestPostgresClient(t)
	defer postgresClient.Close()

	repo := storage.NewAPIKeyRepository(postgresClient.Pool())
	service := security.NewAPIKeyServiceV2(repo, "test-secret-key")
	ctx := context.Background()

	// Create a test user
	userID := uuid.New()

	// Create API key
	key, apiKey, err := service.CreateAPIKey(ctx, userID, "Integration Test Key", 60, 10000, nil)
	if err != nil {
		t.Fatalf("CreateAPIKey failed: %v", err)
	}

	if key == "" {
		t.Error("Generated key should not be empty")
	}
	if apiKey == nil {
		t.Fatal("API key record should not be nil")
	}

	// Validate the key
	validatedKey, err := service.ValidateAPIKey(ctx, key)
	if err != nil {
		t.Fatalf("ValidateAPIKey failed: %v", err)
	}
	if validatedKey == nil {
		t.Fatal("Validated key should not be nil")
	}
	if validatedKey.ID != apiKey.ID {
		t.Errorf("Expected key ID %s, got %s", apiKey.ID, validatedKey.ID)
	}

	// Test invalid key
	_, err = service.ValidateAPIKey(ctx, "ur_invalidkey123456789012345678901234567890")
	if err == nil {
		t.Error("Invalid key should return error")
	}
}

// TestAPIKeyServiceV2_Integration_Expiration tests API key expiration with real database
func TestAPIKeyServiceV2_Integration_Expiration(t *testing.T) {
	postgresClient := getTestPostgresClient(t)
	defer postgresClient.Close()

	repo := storage.NewAPIKeyRepository(postgresClient.Pool())
	service := security.NewAPIKeyServiceV2(repo, "test-secret-key")
	ctx := context.Background()

	userID := uuid.New()
	expiresAt := time.Now().Add(-1 * time.Hour) // Expired

	// Create expired API key
	key, _, err := service.CreateAPIKey(ctx, userID, "Expired Key", 60, 10000, &expiresAt)
	if err != nil {
		t.Fatalf("CreateAPIKey failed: %v", err)
	}

	// Validation should fail for expired key
	_, err = service.ValidateAPIKey(ctx, key)
	if err == nil {
		t.Error("Expired key should return error")
	}
}

// TestAPIKeyServiceV2_Integration_Inactive tests inactive API key handling
func TestAPIKeyServiceV2_Integration_Inactive(t *testing.T) {
	postgresClient := getTestPostgresClient(t)
	defer postgresClient.Close()

	repo := storage.NewAPIKeyRepository(postgresClient.Pool())
	service := security.NewAPIKeyServiceV2(repo, "test-secret-key")
	ctx := context.Background()

	userID := uuid.New()

	// Create API key
	key, apiKey, err := service.CreateAPIKey(ctx, userID, "Test Key", 60, 10000, nil)
	if err != nil {
		t.Fatalf("CreateAPIKey failed: %v", err)
	}

	// Deactivate the key
	apiKey.IsActive = false
	err = repo.Update(ctx, apiKey)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	// Validation should fail for inactive key
	_, err = service.ValidateAPIKey(ctx, key)
	if err == nil {
		t.Error("Inactive key should return error")
	}
}

// TestRateLimiter_Integration tests rate limiter with real Redis
func TestRateLimiter_Integration(t *testing.T) {
	redisClient := getTestRedisClient(t)
	defer redisClient.Close()

	limiter := security.NewRateLimiter(redisClient)
	ctx := context.Background()
	key := "integration-test-key"
	limitPerMinute := 5
	limitPerDay := 100

	// Make requests up to the limit
	for i := 0; i < limitPerMinute; i++ {
		allowed, err := limiter.CheckRateLimit(ctx, key, limitPerMinute, limitPerDay)
		if err != nil {
			t.Fatalf("CheckRateLimit returned error: %v", err)
		}
		if !allowed {
			t.Errorf("Request %d should be allowed", i+1)
		}
	}

	// Next request should exceed limit
	allowed, err := limiter.CheckRateLimit(ctx, key, limitPerMinute, limitPerDay)
	if err != nil {
		t.Fatalf("CheckRateLimit returned error: %v", err)
	}
	if allowed {
		t.Error("Request should be denied when exceeding minute limit")
	}

	// Check remaining requests
	minRemaining, dayRemaining, err := limiter.GetRemainingRequests(ctx, key, limitPerMinute, limitPerDay)
	if err != nil {
		t.Fatalf("GetRemainingRequests returned error: %v", err)
	}
	if minRemaining != 0 {
		t.Errorf("Expected 0 minute remaining, got %d", minRemaining)
	}
	if dayRemaining != int64(limitPerDay-limitPerMinute) {
		t.Errorf("Expected %d day remaining, got %d", limitPerDay-limitPerMinute, dayRemaining)
	}
}

// TestRateLimiter_Integration_DayLimit tests daily rate limit with real Redis
func TestRateLimiter_Integration_DayLimit(t *testing.T) {
	redisClient := getTestRedisClient(t)
	defer redisClient.Close()

	limiter := security.NewRateLimiter(redisClient)
	ctx := context.Background()
	key := "integration-test-day-key"
	limitPerMinute := 1000 // High minute limit
	limitPerDay := 5       // Low day limit

	// Make requests up to the day limit
	for i := 0; i < limitPerDay; i++ {
		allowed, err := limiter.CheckRateLimit(ctx, key, limitPerMinute, limitPerDay)
		if err != nil {
			t.Fatalf("CheckRateLimit returned error: %v", err)
		}
		if !allowed {
			t.Errorf("Request %d should be allowed", i+1)
		}
	}

	// Next request should exceed day limit
	allowed, err := limiter.CheckRateLimit(ctx, key, limitPerMinute, limitPerDay)
	if err != nil {
		t.Fatalf("CheckRateLimit returned error: %v", err)
	}
	if allowed {
		t.Error("Request should be denied when exceeding day limit")
	}
}

// TestFullFlow_Integration tests the complete flow: API key creation -> validation -> rate limiting
func TestFullFlow_Integration(t *testing.T) {
	// Setup
	postgresClient := getTestPostgresClient(t)
	defer postgresClient.Close()

	redisClient := getTestRedisClient(t)
	defer redisClient.Close()

	// Create services
	apiKeyRepo := storage.NewAPIKeyRepository(postgresClient.Pool())
	apiKeyService := security.NewAPIKeyServiceV2(apiKeyRepo, "test-secret-key")
	rateLimiter := security.NewRateLimiter(redisClient)

	ctx := context.Background()
	userID := uuid.New()

	// Step 1: Create API key
	key, apiKey, err := apiKeyService.CreateAPIKey(ctx, userID, "Full Flow Test", 10, 100, nil)
	if err != nil {
		t.Fatalf("CreateAPIKey failed: %v", err)
	}

	// Step 2: Validate API key
	validatedKey, err := apiKeyService.ValidateAPIKey(ctx, key)
	if err != nil {
		t.Fatalf("ValidateAPIKey failed: %v", err)
	}
	if validatedKey.ID != apiKey.ID {
		t.Errorf("Key IDs don't match: %s != %s", validatedKey.ID, apiKey.ID)
	}

	// Step 3: Use rate limiter with API key
	rateLimitKey := "key:" + key
	allowed, err := rateLimiter.CheckRateLimit(ctx, rateLimitKey, apiKey.RateLimitPerMinute, apiKey.RateLimitPerDay)
	if err != nil {
		t.Fatalf("CheckRateLimit failed: %v", err)
	}
	if !allowed {
		t.Error("First request should be allowed")
	}

	// Step 4: Check remaining requests
	minRemaining, dayRemaining, err := rateLimiter.GetRemainingRequests(ctx, rateLimitKey, apiKey.RateLimitPerMinute, apiKey.RateLimitPerDay)
	if err != nil {
		t.Fatalf("GetRemainingRequests failed: %v", err)
	}
	if minRemaining != int64(apiKey.RateLimitPerMinute-1) {
		t.Errorf("Expected %d minute remaining, got %d", apiKey.RateLimitPerMinute-1, minRemaining)
	}
	if dayRemaining != int64(apiKey.RateLimitPerDay-1) {
		t.Errorf("Expected %d day remaining, got %d", apiKey.RateLimitPerDay-1, dayRemaining)
	}
}
