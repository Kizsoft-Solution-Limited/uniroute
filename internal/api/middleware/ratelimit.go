package middleware

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/Kizsoft-Solution-Limited/uniroute/internal/security"
	"github.com/Kizsoft-Solution-Limited/uniroute/internal/storage"
	"github.com/Kizsoft-Solution-Limited/uniroute/pkg/errors"
)

// RateLimitMiddleware creates middleware for rate limiting
func RateLimitMiddleware(rateLimiter *security.RateLimiter, getLimits func(string) (int, int)) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get identifier (API key or IP)
		identifier := getIdentifier(c)

		// Get limits for this identifier
		// First check if API key record is in context (set by AuthMiddleware)
		// and extract rate limits from it
		var limitPerMinute, limitPerDay int
		if apiKeyRecord, exists := c.Get("api_key_record"); exists {
			if keyRecord, ok := apiKeyRecord.(*storage.APIKey); ok {
				// Use API key's configured rate limits
				limitPerMinute = keyRecord.RateLimitPerMinute
				limitPerDay = keyRecord.RateLimitPerDay
			}
		}
		
		// If limits not found in context (or are 0), use getLimits function as fallback
		if limitPerMinute == 0 || limitPerDay == 0 {
			limitPerMinute, limitPerDay = getLimits(identifier)
		}

		// Check rate limit
		allowed, err := rateLimiter.CheckRateLimit(c.Request.Context(), identifier, limitPerMinute, limitPerDay)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "rate limit check failed",
			})
			c.Abort()
			return
		}

		if !allowed {
			// Get remaining requests for response headers
			minRemaining, dayRemaining, _ := rateLimiter.GetRemainingRequests(c.Request.Context(), identifier, limitPerMinute, limitPerDay)
			
			c.Header("X-RateLimit-Limit-PerMinute", strconv.Itoa(limitPerMinute))
			c.Header("X-RateLimit-Limit-PerDay", strconv.Itoa(limitPerDay))
			c.Header("X-RateLimit-Remaining-PerMinute", strconv.FormatInt(minRemaining, 10))
			c.Header("X-RateLimit-Remaining-PerDay", strconv.FormatInt(dayRemaining, 10))
			
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": errors.ErrRateLimitExceeded.Error(),
			})
			c.Abort()
			return
		}

		// Add rate limit headers to successful responses
		minRemaining, dayRemaining, _ := rateLimiter.GetRemainingRequests(c.Request.Context(), identifier, limitPerMinute, limitPerDay)
		c.Header("X-RateLimit-Limit-PerMinute", strconv.Itoa(limitPerMinute))
		c.Header("X-RateLimit-Limit-PerDay", strconv.Itoa(limitPerDay))
		c.Header("X-RateLimit-Remaining-PerMinute", strconv.FormatInt(minRemaining, 10))
		c.Header("X-RateLimit-Remaining-PerDay", strconv.FormatInt(dayRemaining, 10))

		c.Next()
	}
}

// getIdentifier gets the identifier for rate limiting (API key or IP)
func getIdentifier(c *gin.Context) string {
	// Prefer API key if available
	if apiKey, exists := c.Get("api_key"); exists {
		if key, ok := apiKey.(string); ok {
			return "key:" + key
		}
	}

	// Fall back to IP address
	return "ip:" + c.ClientIP()
}

