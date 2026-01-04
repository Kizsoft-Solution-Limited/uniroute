package errors

import "fmt"

// Common errors
var (
	ErrProviderUnavailable = fmt.Errorf("provider unavailable")
	ErrInvalidAPIKey       = fmt.Errorf("invalid API key")
	ErrUnauthorized        = fmt.Errorf("unauthorized")
	ErrInvalidRequest      = fmt.Errorf("invalid request")
	ErrProviderNotFound    = fmt.Errorf("provider not found")
	ErrRateLimitExceeded   = fmt.Errorf("rate limit exceeded")
)
