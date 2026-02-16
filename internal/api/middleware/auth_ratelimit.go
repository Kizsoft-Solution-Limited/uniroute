package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/Kizsoft-Solution-Limited/uniroute/internal/security"
)

// AuthRateLimitMiddleware creates middleware for progressive rate limiting on auth endpoints
// It uses email-based tracking when available, falling back to IP-based tracking
func AuthRateLimitMiddleware(authRateLimiter *security.AuthRateLimiter, maxAttempts int) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Try to get email from request body for more accurate tracking
		// This prevents shared IP issues (NAT, VPN, etc.)
		var identifier string
		
		// Check if we can extract email from request
		if email := extractEmailFromRequest(c); email != "" {
			identifier = "email:" + email
		} else {
			// Fall back to IP-based tracking
			identifier = "ip:" + c.ClientIP()
		}

		// Check rate limit
		allowed, waitTime, blocked, err := authRateLimiter.CheckAuthRateLimit(c.Request.Context(), identifier, maxAttempts)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "rate limit check failed",
			})
			c.Abort()
			return
		}

		if !allowed {
			// Calculate retry after time
			retryAfter := time.Now().Add(time.Duration(waitTime) * time.Second)

			c.Header("X-RateLimit-Limit", strconv.Itoa(maxAttempts))
			c.Header("X-RateLimit-Remaining", "0")
			c.Header("Retry-After", strconv.FormatInt(retryAfter.Unix(), 10))
			c.Header("X-RateLimit-Reset", strconv.FormatInt(retryAfter.Unix(), 10))

			var errorMsg string
			if blocked {
				errorMsg = fmt.Sprintf("Too many failed attempts. Account temporarily blocked. Please try again in %d minutes.", waitTime/60)
			} else {
				errorMsg = fmt.Sprintf("Too many attempts. Please wait %d seconds before trying again.", waitTime)
			}

			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":      errorMsg,
				"retry_after": waitTime,
				"blocked":    blocked,
			})
			c.Abort()
			return
		}

		// Add rate limit headers
		waitTime, blocked, _ = authRateLimiter.GetWaitTime(c.Request.Context(), identifier)
		c.Header("X-RateLimit-Limit", strconv.Itoa(maxAttempts))
		if waitTime > 0 {
			c.Header("X-RateLimit-Remaining", "0")
			c.Header("Retry-After", strconv.Itoa(waitTime))
		}

		c.Next()
	}
}

// extractEmailFromRequest reads and restores the body so the handler can still read it.
func extractEmailFromRequest(c *gin.Context) string {
	// Check if body is already consumed
	if c.Request.Body == nil {
		// Try form data as fallback
		if email := c.PostForm("email"); email != "" {
			return email
		}
		return ""
	}
	
	// Read the request body
	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil || len(bodyBytes) == 0 {
		// Try form data as fallback
		if email := c.PostForm("email"); email != "" {
			return email
		}
		return ""
	}
	
	// Restore the body so the handler can read it
	c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	
	// Try to parse JSON to extract email
	var body map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &body); err == nil {
		if email, ok := body["email"].(string); ok && email != "" {
			return email
		}
	}
	
	return ""
}

