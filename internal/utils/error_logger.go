package utils

import (
	"context"
	"fmt"
	"runtime"
	"strings"

	"github.com/Kizsoft-Solution-Limited/uniroute/internal/storage"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

type ErrorLogger struct {
	errorLogRepo *storage.ErrorLogRepository
	logger       zerolog.Logger
}

func NewErrorLogger(errorLogRepo *storage.ErrorLogRepository, logger zerolog.Logger) *ErrorLogger {
	return &ErrorLogger{
		errorLogRepo: errorLogRepo,
		logger:       logger,
	}
}

func (e *ErrorLogger) LogError(ctx context.Context, err error, contextData map[string]interface{}, userID *uuid.UUID) {
	if err == nil {
		return
	}

	stackTrace := getStackTrace()

	pc, file, line, ok := runtime.Caller(1)
	var callerInfo string
	if ok {
		fn := runtime.FuncForPC(pc)
		if fn != nil {
			callerInfo = fmt.Sprintf("%s:%d in %s", file, line, fn.Name())
		} else {
			callerInfo = fmt.Sprintf("%s:%d", file, line)
		}
	}

	if contextData == nil {
		contextData = make(map[string]interface{})
	}
	contextData["caller"] = callerInfo
	contextData["error_type"] = "backend"

	errorLog := &storage.ErrorLog{
		UserID:     userID,
		ErrorType:  "exception",
		Message:    err.Error(),
		StackTrace: &stackTrace,
		Context:    contextData,
		Severity:    "error",
		Resolved:   false,
	}

	go func() {
		if err := e.errorLogRepo.CreateErrorLog(ctx, errorLog); err != nil {
			e.logger.Error().Err(err).Msg("Failed to log error to database")
		} else {
			e.logger.Debug().
				Str("error_id", errorLog.ID.String()).
				Str("message", err.Error()).
				Msg("Error logged to database")
		}
	}()

	e.logger.Error().
		Err(err).
		Fields(contextData).
		Str("caller", callerInfo).
		Msg("Backend error occurred")
}

func (e *ErrorLogger) LogErrorWithContext(ctx context.Context, err error, message string, contextData map[string]interface{}, userID *uuid.UUID) {
	if contextData == nil {
		contextData = make(map[string]interface{})
	}
	contextData["custom_message"] = message
	e.LogError(ctx, err, contextData, userID)
}

func (e *ErrorLogger) LogWarning(ctx context.Context, message string, contextData map[string]interface{}, userID *uuid.UUID) {
	if contextData == nil {
		contextData = make(map[string]interface{})
	}
	contextData["error_type"] = "backend"

	errorLog := &storage.ErrorLog{
		UserID:    userID,
		ErrorType: "message",
		Message:   message,
		Context:   contextData,
		Severity:  "warning",
		Resolved:  false,
	}

	go func() {
		if err := e.errorLogRepo.CreateErrorLog(ctx, errorLog); err != nil {
			e.logger.Error().Err(err).Msg("Failed to log warning to database")
		}
	}()

	e.logger.Warn().
		Fields(contextData).
		Msg(message)
}

func getStackTrace() string {
	buf := make([]byte, 4096)
	n := runtime.Stack(buf, false)
	stack := string(buf[:n])
	lines := strings.Split(stack, "\n")
	if len(lines) > 10 {
		lines = lines[:10]
	}
	return strings.Join(lines, "\n")
}


