package logger

import (
	"os"

	"github.com/rs/zerolog"
)

// New creates a new logger instance
func New() zerolog.Logger {
	return zerolog.New(os.Stdout).
		With().
		Timestamp().
		Logger().
		Level(zerolog.InfoLevel)
}

// NewDebug creates a logger with debug level
func NewDebug() zerolog.Logger {
	return zerolog.New(os.Stdout).
		With().
		Timestamp().
		Logger().
		Level(zerolog.DebugLevel)
}
