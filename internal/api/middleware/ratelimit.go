package middleware

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/Kizsoft-Solution-Limited/uniroute/internal/security"
	"github.com/Kizsoft-Solution-Limited/uniroute/internal/storage"
	"github.com/Kizsoft-Solution-Limited/uniroute/pkg/errors"
)

func RateLimitMiddleware(rateLimiter *security.RateLimiter, getLimits func(string) (int, int)) gin.HandlerFunc {
	return func(c *gin.Context) {
		identifier := getIdentifier(c)

		var limitPerMinute, limitPerDay int
		if apiKeyRecord, exists := c.Get("api_key_record"); exists {
			if keyRecord, ok := apiKeyRecord.(*storage.APIKey); ok {
				limitPerMinute = keyRecord.RateLimitPerMinute
				limitPerDay = keyRecord.RateLimitPerDay
			}
		}

		if limitPerMinute == 0 || limitPerDay == 0 {
			limitPerMinute, limitPerDay = getLimits(identifier)
		}

		allowed, err := rateLimiter.CheckRateLimit(c.Request.Context(), identifier, limitPerMinute, limitPerDay)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "rate limit check failed",
			})
			c.Abort()
			return
		}

		if !allowed {
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

		minRemaining, dayRemaining, _ := rateLimiter.GetRemainingRequests(c.Request.Context(), identifier, limitPerMinute, limitPerDay)
		c.Header("X-RateLimit-Limit-PerMinute", strconv.Itoa(limitPerMinute))
		c.Header("X-RateLimit-Limit-PerDay", strconv.Itoa(limitPerDay))
		c.Header("X-RateLimit-Remaining-PerMinute", strconv.FormatInt(minRemaining, 10))
		c.Header("X-RateLimit-Remaining-PerDay", strconv.FormatInt(dayRemaining, 10))

		c.Next()
	}
}

func getIdentifier(c *gin.Context) string {
	if apiKey, exists := c.Get("api_key"); exists {
		if key, ok := apiKey.(string); ok {
			return "key:" + key
		}
	}
	return "ip:" + c.ClientIP()
}

