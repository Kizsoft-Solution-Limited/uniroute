package middleware

import (
	"github.com/gin-gonic/gin"
)

func SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-XSS-Protection", "1; mode=block")

		if gin.Mode() == gin.ReleaseMode {
			path := c.Request.URL.Path
			if path == "/swagger" || path == "/swagger/" || path == "/swagger.json" {
				c.Header("Content-Security-Policy",
					"default-src 'self'; "+
						"script-src 'self' 'unsafe-inline' 'unsafe-eval' https://unpkg.com; "+
						"style-src 'self' 'unsafe-inline' https://unpkg.com; "+
						"font-src 'self' data: https://unpkg.com; "+
						"img-src 'self' data: blob: https:; "+
						"media-src 'self' data: blob:; "+
						"connect-src 'self' https://unpkg.com")
			} else {
				c.Header("Content-Security-Policy",
					"default-src 'self'; "+
						"media-src 'self' data: blob:; "+
						"img-src 'self' data: blob: https:; "+
						"connect-src 'self'")
			}
		}

		if c.Request.TLS != nil {
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}

		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Permissions-Policy", "microphone=(self), camera=(), geolocation=()")

		c.Next()
	}
}
