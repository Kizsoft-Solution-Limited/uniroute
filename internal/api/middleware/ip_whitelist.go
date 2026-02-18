package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func IPWhitelistMiddleware(allowedIPs []string) gin.HandlerFunc {
	allowedMap := make(map[string]bool)
	for _, ip := range allowedIPs {
		allowedMap[strings.TrimSpace(ip)] = true
	}

	return func(c *gin.Context) {
		if len(allowedIPs) == 0 {
			c.Next()
			return
		}

		clientIP := c.ClientIP()

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
