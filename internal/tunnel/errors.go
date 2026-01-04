package tunnel

import "errors"

var (
	// Request tracking errors
	ErrDuplicateRequestID    = errors.New("duplicate request ID")
	ErrRequestNotFound        = errors.New("request not found")
	ErrRequestTimeout        = errors.New("request timeout")
	ErrResponseChannelTimeout = errors.New("response channel timeout")

	// Authentication errors
	ErrInvalidToken      = errors.New("invalid tunnel token")
	ErrTokenExpired      = errors.New("tunnel token expired")
	ErrUnauthorized      = errors.New("unauthorized")
	ErrTokenRequired     = errors.New("tunnel token required")

	// Connection errors
	ErrConnectionClosed = errors.New("connection closed")
	ErrConnectionLost   = errors.New("connection lost")

	// Tunnel errors
	ErrTunnelNotFound    = errors.New("tunnel not found")
	ErrTunnelInactive    = errors.New("tunnel is inactive")
	ErrSubdomainTaken    = errors.New("subdomain already taken")
	ErrInvalidSubdomain  = errors.New("invalid subdomain")
	ErrInvalidLocalURL   = errors.New("invalid local URL")

	// Request validation errors
	ErrInvalidRequest = errors.New("invalid request")

	// Database errors
	ErrDatabaseError = errors.New("database error")
)

