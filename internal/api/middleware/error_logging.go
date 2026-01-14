package middleware

import (
	"runtime"

	"github.com/Kizsoft-Solution-Limited/uniroute/internal/storage"
	"github.com/Kizsoft-Solution-Limited/uniroute/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// ErrorLoggingMiddleware creates middleware that logs errors to database
func ErrorLoggingMiddleware(errorLogRepo *storage.ErrorLogRepository, logger zerolog.Logger) gin.HandlerFunc {
	errorLogger := utils.NewErrorLogger(errorLogRepo, logger)

	return func(c *gin.Context) {
		// Continue with request
		c.Next()

		// Check if there are any errors
		if len(c.Errors) > 0 {
			// Get user ID from context if available
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

			// Log each error
			for _, err := range c.Errors {
				contextData := map[string]interface{}{
					"path":       c.Request.URL.Path,
					"method":     c.Request.Method,
					"status":     c.Writer.Status(),
					"client_ip":  c.ClientIP(),
					"user_agent": c.GetHeader("User-Agent"),
				}

				// Add request body if available (for debugging, but be careful with sensitive data)
				if c.Request.Body != nil && c.Request.ContentLength < 1024 { // Only for small requests
					// Note: Body is already read, so we can't read it again
					// This is just for reference
				}

				errorLogger.LogError(c.Request.Context(), err.Err, contextData, userID)
			}
		}

		// Also catch panics
		defer func() {
			if r := recover(); r != nil {
				// Get stack trace
				buf := make([]byte, 4096)
				runtime.Stack(buf, false)

				// Get user ID
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

				// Create panic error
				var panicErr error
				if err, ok := r.(error); ok {
					panicErr = err
				} else {
					panicErr = &PanicError{Value: r}
				}

				errorLogger.LogError(c.Request.Context(), panicErr, contextData, userID)

				// Re-panic to let Gin's recovery handle it
				panic(r)
			}
		}()
	}
}

// PanicError wraps non-error panic values
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


