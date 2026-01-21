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
		// Note: CSP is set more permissively in development to allow CORS
		// In production, this should be more restrictive
		// Swagger UI needs relaxed CSP to load external resources
		if gin.Mode() == gin.ReleaseMode {
			// Check if this is the Swagger UI route (handle with or without trailing slash)
			path := c.Request.URL.Path
			if path == "/swagger" || path == "/swagger/" || path == "/swagger.json" {
				// Relaxed CSP for Swagger UI to allow CDN resources and inline scripts
				// This allows Swagger UI to load from unpkg.com CDN and execute inline scripts
				// connect-src includes unpkg.com for source map loading
				c.Header("Content-Security-Policy", 
					"default-src 'self'; "+
					"script-src 'self' 'unsafe-inline' 'unsafe-eval' https://unpkg.com; "+
					"style-src 'self' 'unsafe-inline' https://unpkg.com; "+
					"font-src 'self' data: https://unpkg.com; "+
					"img-src 'self' data: blob: https:; "+
					"media-src 'self' data: blob:; "+
					"connect-src 'self' https://unpkg.com")
			} else {
				c.Header("Content-Security-Policy", "default-src 'self'")
			}
		}

		// Strict-Transport-Security: Force HTTPS (if using HTTPS)
		// Only set if request is HTTPS
		if c.Request.TLS != nil {
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}

		// Referrer-Policy: Control referrer information
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		// Permissions-Policy: Control browser features (microphone, camera, etc.)
		// Allow microphone for voice recording feature (self = same origin only)
		// Format: feature=(allowed-origins), use * for all origins or 'self' for same origin
		c.Header("Permissions-Policy", "microphone=(self), camera=(), geolocation=()")

		c.Next()
	}
}
