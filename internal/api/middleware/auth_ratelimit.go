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

func AuthRateLimitMiddleware(authRateLimiter *security.AuthRateLimiter, maxAttempts int) gin.HandlerFunc {
	return func(c *gin.Context) {
		var identifier string
		if email := extractEmailFromRequest(c); email != "" {
			identifier = "email:" + email
		} else {
			identifier = "ip:" + c.ClientIP()
		}

		allowed, waitTime, blocked, err := authRateLimiter.CheckAuthRateLimit(c.Request.Context(), identifier, maxAttempts)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "rate limit check failed",
			})
			c.Abort()
			return
		}

		if !allowed {
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

		waitTime, blocked, _ = authRateLimiter.GetWaitTime(c.Request.Context(), identifier)
		c.Header("X-RateLimit-Limit", strconv.Itoa(maxAttempts))
		if waitTime > 0 {
			c.Header("X-RateLimit-Remaining", "0")
			c.Header("Retry-After", strconv.Itoa(waitTime))
		}

		c.Next()
	}
}

func extractEmailFromRequest(c *gin.Context) string {
	if c.Request.Body == nil {
		if email := c.PostForm("email"); email != "" {
			return email
		}
		return ""
	}

	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil || len(bodyBytes) == 0 {
		if email := c.PostForm("email"); email != "" {
			return email
		}
		return ""
	}

	c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	var body map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &body); err == nil {
		if email, ok := body["email"].(string); ok && email != "" {
			return email
		}
	}
	
	return ""
}

