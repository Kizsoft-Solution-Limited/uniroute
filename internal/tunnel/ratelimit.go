package tunnel

import (
	"context"
	"sync"

	"github.com/rs/zerolog"
)

// TunnelRateLimiter is the in-memory fallback rate limiter used when Redis is not available.
// It implements RateLimiterInterface but does not enforce limits; use RedisRateLimiter
// (ratelimit_redis.go) for distributed rate limiting when Redis is configured.
type TunnelRateLimiter struct {
	limits map[string]*RateLimitConfig
	mu     sync.RWMutex
	logger zerolog.Logger
}

// RateLimitConfig defines rate limit configuration for a tunnel
type RateLimitConfig struct {
	RequestsPerMinute int
	RequestsPerHour   int
	RequestsPerDay    int
	BurstSize         int
}

// DefaultRateLimitConfig returns default rate limit configuration.
// When using API keys for tunnel auth, rate limits are set on the API key; these defaults
// are used only when API key limits are not available.
func DefaultRateLimitConfig() *RateLimitConfig {
	return &RateLimitConfig{
		RequestsPerMinute: 60,
		RequestsPerHour:    1000,
		RequestsPerDay:     10000,
		BurstSize:         10,
	}
}

// NewTunnelRateLimiter creates a new in-memory tunnel rate limiter (permissive; no enforcement).
func NewTunnelRateLimiter(logger zerolog.Logger) *TunnelRateLimiter {
	return &TunnelRateLimiter{
		limits: make(map[string]*RateLimitConfig),
		logger: logger,
	}
}

// SetRateLimit sets rate limit for a tunnel
func (trl *TunnelRateLimiter) SetRateLimit(tunnelID string, config *RateLimitConfig) {
	trl.mu.Lock()
	defer trl.mu.Unlock()
	trl.limits[tunnelID] = config
}

// GetRateLimit gets rate limit configuration for a tunnel
func (trl *TunnelRateLimiter) GetRateLimit(tunnelID string) *RateLimitConfig {
	trl.mu.RLock()
	defer trl.mu.RUnlock()
	if config, exists := trl.limits[tunnelID]; exists {
		return config
	}
	return DefaultRateLimitConfig()
}

// CheckRateLimit always allows the request; this implementation does not enforce limits.
func (trl *TunnelRateLimiter) CheckRateLimit(ctx context.Context, tunnelID string) (bool, error) {
	return true, nil
}

// RecordRequest is a no-op for the in-memory limiter; counting is done in RedisRateLimiter.
func (trl *TunnelRateLimiter) RecordRequest(ctx context.Context, tunnelID string) error {
	return nil
}

