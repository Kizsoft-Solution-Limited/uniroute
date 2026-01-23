package tunnel

import (
	"context"
	"sync"

	"github.com/rs/zerolog"
)

// TunnelRateLimiter handles rate limiting per tunnel
type TunnelRateLimiter struct {
	limits    map[string]*RateLimitConfig
	mu        sync.RWMutex
	logger    zerolog.Logger
	redis     interface{} // Will integrate with Redis later
}

// RateLimitConfig defines rate limit configuration for a tunnel
type RateLimitConfig struct {
	RequestsPerMinute int
	RequestsPerHour   int
	RequestsPerDay    int
	BurstSize         int
}

// DefaultRateLimitConfig returns default rate limit configuration
// NOTE: When using API keys for tunnel authentication, users set rate limits when creating the API key.
// Ideally, tunnels should use the API key's rate limits (RateLimitPerMinute, RateLimitPerDay).
// These defaults are only used as fallback when API key rate limits are not available.
func DefaultRateLimitConfig() *RateLimitConfig {
	return &RateLimitConfig{
		RequestsPerMinute: 60,   // Default fallback - users should set this in API key creation
		RequestsPerHour:   1000, // Default fallback - users should set this in API key creation
		RequestsPerDay:    10000, // Default fallback - users should set this in API key creation
		BurstSize:         10,   // Default fallback
	}
}

// NewTunnelRateLimiter creates a new tunnel rate limiter
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

// CheckRateLimit checks if a request should be allowed (simplified in-memory version)
// TODO: Integrate with Redis for distributed rate limiting
func (trl *TunnelRateLimiter) CheckRateLimit(ctx context.Context, tunnelID string) (bool, error) {
	// Use simple in-memory rate limiting
	// TODO: Integrate with Redis for distributed rate limiting
	config := trl.GetRateLimit(tunnelID)
	
	// Simple check - in production, use token bucket or sliding window
	// For now, just return true (no rate limiting)
	_ = config
	return true, nil
}

// RecordRequest records a request for rate limiting purposes
func (trl *TunnelRateLimiter) RecordRequest(ctx context.Context, tunnelID string) error {
	// TODO: Record in Redis for distributed rate limiting
	return nil
}

