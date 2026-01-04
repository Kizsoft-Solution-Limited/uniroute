package tunnel

import (
	"context"
	"time"

	"github.com/rs/zerolog"
)

// RequestLogger handles logging tunnel requests to database
type RequestLogger struct {
	repository *TunnelRepository
	logger     zerolog.Logger
	enabled    bool
}

// NewRequestLogger creates a new request logger
func NewRequestLogger(repo *TunnelRepository, logger zerolog.Logger) *RequestLogger {
	return &RequestLogger{
		repository: repo,
		logger:     logger,
		enabled:    repo != nil,
	}
}

// LogRequest logs a tunnel request to the database
func (rl *RequestLogger) LogRequest(ctx context.Context, req *TunnelRequestLog) error {
	if !rl.enabled {
		return nil
	}

	// Async logging to avoid blocking request flow
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := rl.repository.CreateTunnelRequest(ctx, req); err != nil {
			rl.logger.Error().Err(err).Str("request_id", req.RequestID).Msg("Failed to log tunnel request")
		}
	}()

	return nil
}

// TunnelRequestLog is defined in repository.go

