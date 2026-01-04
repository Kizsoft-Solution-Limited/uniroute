package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// IPWhitelistMiddleware creates middleware for IP whitelisting
func IPWhitelistMiddleware(allowedIPs []string) gin.HandlerFunc {
	// Convert to map for faster lookup
	allowedMap := make(map[string]bool)
	for _, ip := range allowedIPs {
		allowedMap[strings.TrimSpace(ip)] = true
	}

	return func(c *gin.Context) {
		// If no IPs configured, allow all
		if len(allowedIPs) == 0 {
			c.Next()
			return
		}

		clientIP := c.ClientIP()

		// Check if IP is allowed
		if !allowedMap[clientIP] {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "IP address not allowed",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
