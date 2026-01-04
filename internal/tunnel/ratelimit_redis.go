package tunnel

import (
	"context"
	"fmt"
	"time"

	"github.com/Kizsoft-Solution-Limited/uniroute/internal/storage"
	"github.com/rs/zerolog"
)

// RedisRateLimiter implements distributed rate limiting using Redis
type RedisRateLimiter struct {
	redisClient *storage.RedisClient
	logger      zerolog.Logger
	configs     map[string]*RateLimitConfig
	mu          interface{} // sync.RWMutex (avoid import)
}

// NewRedisRateLimiter creates a new Redis-based rate limiter
func NewRedisRateLimiter(redisClient *storage.RedisClient, logger zerolog.Logger) *RedisRateLimiter {
	return &RedisRateLimiter{
		redisClient: redisClient,
		logger:      logger,
		configs:     make(map[string]*RateLimitConfig),
	}
}

// SetRateLimit sets rate limit configuration for a tunnel
func (rrl *RedisRateLimiter) SetRateLimit(tunnelID string, config *RateLimitConfig) {
	rrl.configs[tunnelID] = config
}

// GetRateLimit gets rate limit configuration for a tunnel
func (rrl *RedisRateLimiter) GetRateLimit(tunnelID string) *RateLimitConfig {
	if config, exists := rrl.configs[tunnelID]; exists {
		return config
	}
	return DefaultRateLimitConfig()
}

// CheckRateLimit checks if a request should be allowed using Redis
func (rrl *RedisRateLimiter) CheckRateLimit(ctx context.Context, tunnelID string) (bool, error) {
	if rrl.redisClient == nil {
		// Fallback to permissive if Redis not available
		return true, nil
	}

	config := rrl.GetRateLimit(tunnelID)
	now := time.Now()

	// Check per-minute limit
	minuteKey := fmt.Sprintf("tunnel:ratelimit:%s:minute:%d", tunnelID, now.Unix()/60)
	minuteCount, err := rrl.redisClient.Client().Incr(ctx, minuteKey).Result()
	if err != nil {
		return true, fmt.Errorf("failed to check rate limit: %w", err)
	}

	if minuteCount == 1 {
		// Set expiration on first request
		rrl.redisClient.Client().Expire(ctx, minuteKey, 1*time.Minute)
	}

	if minuteCount > int64(config.RequestsPerMinute) {
		rrl.logger.Warn().
			Str("tunnel_id", tunnelID).
			Int64("count", minuteCount).
			Int("limit", config.RequestsPerMinute).
			Msg("Rate limit exceeded (per minute)")
		return false, nil
	}

	// Check per-hour limit
	hourKey := fmt.Sprintf("tunnel:ratelimit:%s:hour:%d", tunnelID, now.Unix()/3600)
	hourCount, err := rrl.redisClient.Client().Incr(ctx, hourKey).Result()
	if err != nil {
		return true, fmt.Errorf("failed to check rate limit: %w", err)
	}

	if hourCount == 1 {
		rrl.redisClient.Client().Expire(ctx, hourKey, 1*time.Hour)
	}

	if hourCount > int64(config.RequestsPerHour) {
		rrl.logger.Warn().
			Str("tunnel_id", tunnelID).
			Int64("count", hourCount).
			Int("limit", config.RequestsPerHour).
			Msg("Rate limit exceeded (per hour)")
		return false, nil
	}

	// Check per-day limit
	dayKey := fmt.Sprintf("tunnel:ratelimit:%s:day:%d", tunnelID, now.Unix()/86400)
	dayCount, err := rrl.redisClient.Client().Incr(ctx, dayKey).Result()
	if err != nil {
		return true, fmt.Errorf("failed to check rate limit: %w", err)
	}

	if dayCount == 1 {
		rrl.redisClient.Client().Expire(ctx, dayKey, 24*time.Hour)
	}

	if dayCount > int64(config.RequestsPerDay) {
		rrl.logger.Warn().
			Str("tunnel_id", tunnelID).
			Int64("count", dayCount).
			Int("limit", config.RequestsPerDay).
			Msg("Rate limit exceeded (per day)")
		return false, nil
	}

	return true, nil
}

// RecordRequest records a request for rate limiting (already done in CheckRateLimit)
func (rrl *RedisRateLimiter) RecordRequest(ctx context.Context, tunnelID string) error {
	// Already recorded in CheckRateLimit via Incr
	return nil
}

// GetRemainingRequests returns remaining requests for a tunnel
func (rrl *RedisRateLimiter) GetRemainingRequests(ctx context.Context, tunnelID string) (int, int, int, error) {
	if rrl.redisClient == nil {
		return -1, -1, -1, fmt.Errorf("Redis not available")
	}

	config := rrl.GetRateLimit(tunnelID)
	now := time.Now()

	minuteKey := fmt.Sprintf("tunnel:ratelimit:%s:minute:%d", tunnelID, now.Unix()/60)
	hourKey := fmt.Sprintf("tunnel:ratelimit:%s:hour:%d", tunnelID, now.Unix()/3600)
	dayKey := fmt.Sprintf("tunnel:ratelimit:%s:day:%d", tunnelID, now.Unix()/86400)

	minuteCount, _ := rrl.redisClient.Client().Get(ctx, minuteKey).Int()
	hourCount, _ := rrl.redisClient.Client().Get(ctx, hourKey).Int()
	dayCount, _ := rrl.redisClient.Client().Get(ctx, dayKey).Int()

	remainingMinute := config.RequestsPerMinute - minuteCount
	remainingHour := config.RequestsPerHour - hourCount
	remainingDay := config.RequestsPerDay - dayCount

	return remainingMinute, remainingHour, remainingDay, nil
}

