package tunnel

import "context"

// RateLimiterInterface defines the interface for rate limiters
type RateLimiterInterface interface {
	CheckRateLimit(ctx context.Context, tunnelID string) (bool, error)
	RecordRequest(ctx context.Context, tunnelID string) error
	SetRateLimit(tunnelID string, config *RateLimitConfig)
	GetRateLimit(tunnelID string) *RateLimitConfig
}

