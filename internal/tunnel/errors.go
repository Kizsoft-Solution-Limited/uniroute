package tunnel

import "errors"

var (
	ErrDuplicateRequestID    = errors.New("duplicate request ID")
	ErrRequestNotFound        = errors.New("request not found")
	ErrRequestTimeout        = errors.New("request timeout")
	ErrResponseChannelTimeout = errors.New("response channel timeout")

	ErrInvalidToken      = errors.New("invalid tunnel token")
	ErrTokenExpired      = errors.New("tunnel token expired")
	ErrUnauthorized      = errors.New("unauthorized")
	ErrTokenRequired     = errors.New("tunnel token required")

	ErrConnectionClosed = errors.New("connection closed")
	ErrConnectionLost   = errors.New("connection lost")

	ErrTunnelNotFound    = errors.New("tunnel not found")
	ErrTunnelInactive    = errors.New("tunnel is inactive")
	ErrSubdomainTaken    = errors.New("subdomain already taken")
	ErrInvalidSubdomain  = errors.New("invalid subdomain")
	ErrInvalidLocalURL   = errors.New("invalid local URL")

	ErrInvalidRequest = errors.New("invalid request")
	ErrDatabaseError  = errors.New("database error")
)

