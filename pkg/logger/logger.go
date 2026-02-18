package logger

import (
	"os"

	"github.com/rs/zerolog"
)

func New() zerolog.Logger {
	return zerolog.New(os.Stdout).
		With().
		Timestamp().
		Logger().
		Level(zerolog.InfoLevel)
}

func NewDebug() zerolog.Logger {
	return zerolog.New(os.Stdout).
		With().
		Timestamp().
		Logger().
		Level(zerolog.DebugLevel)
}
