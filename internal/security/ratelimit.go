package security

import (
	"context"
	"fmt"
	"time"

	"github.com/Kizsoft-Solution-Limited/uniroute/internal/storage"
	"github.com/redis/go-redis/v9"
)

type RateLimiter struct {
	redis *storage.RedisClient
}

func NewRateLimiter(redis *storage.RedisClient) *RateLimiter {
	return &RateLimiter{
		redis: redis,
	}
}

func (r *RateLimiter) CheckRateLimit(ctx context.Context, key string, limitPerMinute, limitPerDay int) (bool, error) {
	client := r.redis.Client()

	// Check per-minute limit
	minuteKey := fmt.Sprintf("ratelimit:minute:%s", key)
	minuteCount, err := client.Incr(ctx, minuteKey).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check rate limit: %w", err)
	}

	if minuteCount == 1 {
		// Set expiration for minute window
		client.Expire(ctx, minuteKey, time.Minute)
	}

	if minuteCount > int64(limitPerMinute) {
		return false, nil
	}

	// Check per-day limit
	dayKey := fmt.Sprintf("ratelimit:day:%s:%s", key, time.Now().Format("2006-01-02"))
	dayCount, err := client.Incr(ctx, dayKey).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check daily rate limit: %w", err)
	}

	if dayCount == 1 {
		// Set expiration for day window (25 hours to be safe)
		client.Expire(ctx, dayKey, 25*time.Hour)
	}

	if dayCount > int64(limitPerDay) {
		return false, nil
	}

	return true, nil
}

func (r *RateLimiter) GetRemainingRequests(ctx context.Context, key string, limitPerMinute, limitPerDay int) (minuteRemaining, dayRemaining int64, err error) {
	client := r.redis.Client()

	minuteKey := fmt.Sprintf("ratelimit:minute:%s", key)
	minuteCount, err := client.Get(ctx, minuteKey).Int64()
	if err == redis.Nil {
		minuteCount = 0
	} else if err != nil {
		return 0, 0, err
	}

	dayKey := fmt.Sprintf("ratelimit:day:%s:%s", key, time.Now().Format("2006-01-02"))
	dayCount, err := client.Get(ctx, dayKey).Int64()
	if err == redis.Nil {
		dayCount = 0
	} else if err != nil {
		return 0, 0, err
	}

	minuteRemaining = int64(limitPerMinute) - minuteCount
	if minuteRemaining < 0 {
		minuteRemaining = 0
	}

	dayRemaining = int64(limitPerDay) - dayCount
	if dayRemaining < 0 {
		dayRemaining = 0
	}

	return minuteRemaining, dayRemaining, nil
}

