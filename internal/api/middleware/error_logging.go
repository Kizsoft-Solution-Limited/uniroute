package middleware

import (
	"runtime"

	"github.com/Kizsoft-Solution-Limited/uniroute/internal/storage"
	"github.com/Kizsoft-Solution-Limited/uniroute/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

func ErrorLoggingMiddleware(errorLogRepo *storage.ErrorLogRepository, logger zerolog.Logger) gin.HandlerFunc {
	errorLogger := utils.NewErrorLogger(errorLogRepo, logger)

	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) > 0 {
			var userID *uuid.UUID
			if userIDValue, exists := c.Get("user_id"); exists {
				if uid, ok := userIDValue.(uuid.UUID); ok {
					userID = &uid
				} else if idStr, ok := userIDValue.(string); ok {
					if parsed, err := uuid.Parse(idStr); err == nil {
						userID = &parsed
					}
				}
			}

			for _, err := range c.Errors {
				contextData := map[string]interface{}{
					"path":       c.Request.URL.Path,
					"method":     c.Request.Method,
					"status":     c.Writer.Status(),
					"client_ip":  c.ClientIP(),
					"user_agent": c.GetHeader("User-Agent"),
				}

				errorLogger.LogError(c.Request.Context(), err.Err, contextData, userID)
			}
		}

		defer func() {
			if r := recover(); r != nil {
				buf := make([]byte, 4096)
				runtime.Stack(buf, false)

				var userID *uuid.UUID
				if userIDValue, exists := c.Get("user_id"); exists {
					if uid, ok := userIDValue.(uuid.UUID); ok {
						userID = &uid
					}
				}

				contextData := map[string]interface{}{
					"path":       c.Request.URL.Path,
					"method":     c.Request.Method,
					"panic":      true,
					"recovered":  r,
					"client_ip":  c.ClientIP(),
					"user_agent": c.GetHeader("User-Agent"),
				}

				var panicErr error
				if err, ok := r.(error); ok {
					panicErr = err
				} else {
					panicErr = &PanicError{Value: r}
				}

				errorLogger.LogError(c.Request.Context(), panicErr, contextData, userID)

				panic(r)
			}
		}()
	}
}

type PanicError struct {
	Value interface{}
}

func (e *PanicError) Error() string {
	return "panic: " + stringify(e.Value)
}

func stringify(v interface{}) string {
	if s, ok := v.(string); ok {
		return s
	}
	return "unknown panic value"
}


