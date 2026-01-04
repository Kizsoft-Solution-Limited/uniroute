package security

import (
	"context"
	"testing"
	"time"

	"github.com/Kizsoft-Solution-Limited/uniroute/internal/storage"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

// mockRedisClient is a mock Redis client for testing
type mockRedisClient struct {
	client *redis.Client
}

func newMockRedisClient() (*storage.RedisClient, error) {
	// Use a real Redis client pointing to a test database
	// In a real scenario, you'd use a test Redis instance or a mock
	opts := &redis.Options{
		Addr: "localhost:6379",
		DB:   15, // Use DB 15 for testing
	}
	client := redis.NewClient(opts)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	// Clear test database
	client.FlushDB(ctx)

	// Use NewRedisClient with a constructed URL
	// Since we can't access unexported fields, we'll use the public API
	logger := zerolog.Nop()
	return storage.NewRedisClient("redis://localhost:6379/15", logger)
}

func TestNewRateLimiter(t *testing.T) {
	redisClient, err := newMockRedisClient()
	if err != nil {
		t.Skipf("Skipping test: Redis not available: %v", err)
	}
	defer redisClient.Close()

	limiter := NewRateLimiter(redisClient)
	if limiter == nil {
		t.Fatal("NewRateLimiter returned nil")
	}
}

func TestRateLimiter_CheckRateLimit_WithinLimit(t *testing.T) {
	redisClient, err := newMockRedisClient()
	if err != nil {
		t.Skipf("Skipping test: Redis not available: %v", err)
	}
	defer redisClient.Close()

	limiter := NewRateLimiter(redisClient)
	ctx := context.Background()

	// Test within limit
	allowed, err := limiter.CheckRateLimit(ctx, "test-key-1", 10, 100)
	if err != nil {
		t.Fatalf("CheckRateLimit returned error: %v", err)
	}
	if !allowed {
		t.Error("Request should be allowed within limit")
	}
}

func TestRateLimiter_CheckRateLimit_ExceedsMinuteLimit(t *testing.T) {
	redisClient, err := newMockRedisClient()
	if err != nil {
		t.Skipf("Skipping test: Redis not available: %v", err)
	}
	defer redisClient.Close()

	limiter := NewRateLimiter(redisClient)
	ctx := context.Background()
	key := "test-key-2"
	limitPerMinute := 5

	// Make requests up to the limit
	for i := 0; i < limitPerMinute; i++ {
		allowed, err := limiter.CheckRateLimit(ctx, key, limitPerMinute, 100)
		if err != nil {
			t.Fatalf("CheckRateLimit returned error: %v", err)
		}
		if !allowed {
			t.Errorf("Request %d should be allowed", i+1)
		}
	}

	// Next request should exceed limit
	allowed, err := limiter.CheckRateLimit(ctx, key, limitPerMinute, 100)
	if err != nil {
		t.Fatalf("CheckRateLimit returned error: %v", err)
	}
	if allowed {
		t.Error("Request should be denied when exceeding minute limit")
	}
}

func TestRateLimiter_CheckRateLimit_ExceedsDayLimit(t *testing.T) {
	redisClient, err := newMockRedisClient()
	if err != nil {
		t.Skipf("Skipping test: Redis not available: %v", err)
	}
	defer redisClient.Close()

	limiter := NewRateLimiter(redisClient)
	ctx := context.Background()
	key := "test-key-3"
	limitPerDay := 5

	// Make requests up to the day limit
	for i := 0; i < limitPerDay; i++ {
		allowed, err := limiter.CheckRateLimit(ctx, key, 100, limitPerDay)
		if err != nil {
			t.Fatalf("CheckRateLimit returned error: %v", err)
		}
		if !allowed {
			t.Errorf("Request %d should be allowed", i+1)
		}
	}

	// Next request should exceed day limit
	allowed, err := limiter.CheckRateLimit(ctx, key, 100, limitPerDay)
	if err != nil {
		t.Fatalf("CheckRateLimit returned error: %v", err)
	}
	if allowed {
		t.Error("Request should be denied when exceeding day limit")
	}
}

func TestRateLimiter_GetRemainingRequests(t *testing.T) {
	redisClient, err := newMockRedisClient()
	if err != nil {
		t.Skipf("Skipping test: Redis not available: %v", err)
	}
	defer redisClient.Close()

	limiter := NewRateLimiter(redisClient)
	ctx := context.Background()
	key := "test-key-4"
	limitPerMinute := 10
	limitPerDay := 100

	// Initially, all requests should be remaining
	minRemaining, dayRemaining, err := limiter.GetRemainingRequests(ctx, key, limitPerMinute, limitPerDay)
	if err != nil {
		t.Fatalf("GetRemainingRequests returned error: %v", err)
	}
	if minRemaining != int64(limitPerMinute) {
		t.Errorf("Expected %d minute remaining, got %d", limitPerMinute, minRemaining)
	}
	if dayRemaining != int64(limitPerDay) {
		t.Errorf("Expected %d day remaining, got %d", limitPerDay, dayRemaining)
	}

	// Make some requests
	for i := 0; i < 3; i++ {
		_, err := limiter.CheckRateLimit(ctx, key, limitPerMinute, limitPerDay)
		if err != nil {
			t.Fatalf("CheckRateLimit returned error: %v", err)
		}
	}

	// Check remaining
	minRemaining, dayRemaining, err = limiter.GetRemainingRequests(ctx, key, limitPerMinute, limitPerDay)
	if err != nil {
		t.Fatalf("GetRemainingRequests returned error: %v", err)
	}
	if minRemaining != int64(limitPerMinute-3) {
		t.Errorf("Expected %d minute remaining, got %d", limitPerMinute-3, minRemaining)
	}
	if dayRemaining != int64(limitPerDay-3) {
		t.Errorf("Expected %d day remaining, got %d", limitPerDay-3, dayRemaining)
	}
}
