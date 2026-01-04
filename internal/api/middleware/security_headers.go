package middleware

import (
	"github.com/gin-gonic/gin"
)

// SecurityHeadersMiddleware adds security headers to responses
func SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// X-Frame-Options: Prevent clickjacking
		c.Header("X-Frame-Options", "DENY")

		// X-Content-Type-Options: Prevent MIME type sniffing
		c.Header("X-Content-Type-Options", "nosniff")

		// X-XSS-Protection: Enable XSS filter
		c.Header("X-XSS-Protection", "1; mode=block")

		// Content-Security-Policy: Restrict resource loading
		c.Header("Content-Security-Policy", "default-src 'self'")

		// Strict-Transport-Security: Force HTTPS (if using HTTPS)
		// Only set if request is HTTPS
		if c.Request.TLS != nil {
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}

		// Referrer-Policy: Control referrer information
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		c.Next()
	}
}

