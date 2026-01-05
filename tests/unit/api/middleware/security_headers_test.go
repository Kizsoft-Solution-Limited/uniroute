package middleware_test

import (
	"crypto/tls"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/Kizsoft-Solution-Limited/uniroute/internal/api/middleware"
)

func TestSecurityHeadersMiddleware(t *testing.T) {
	// Set to ReleaseMode because CSP is only set in ReleaseMode
	gin.SetMode(gin.ReleaseMode)
	defer gin.SetMode(gin.TestMode) // Restore after test

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)

	middleware := middleware.SecurityHeadersMiddleware()
	middleware(c)

	// Check security headers
	headers := map[string]string{
		"X-Frame-Options":           "DENY",
		"X-Content-Type-Options":     "nosniff",
		"X-XSS-Protection":           "1; mode=block",
		"Content-Security-Policy":    "default-src 'self'",
		"Referrer-Policy":            "strict-origin-when-cross-origin",
	}

	for header, expectedValue := range headers {
		actualValue := c.Writer.Header().Get(header)
		if actualValue != expectedValue {
			t.Errorf("Expected %s header '%s', got '%s'", header, expectedValue, actualValue)
		}
	}
}

func TestSecurityHeadersMiddleware_HSTS(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)
	c.Request.TLS = &tls.ConnectionState{} // Simulate HTTPS

	middleware := middleware.SecurityHeadersMiddleware()
	middleware(c)

	// HSTS should be set for HTTPS
	hsts := c.Writer.Header().Get("Strict-Transport-Security")
	if hsts == "" {
		t.Error("Strict-Transport-Security header should be set for HTTPS requests")
	}
	if hsts != "max-age=31536000; includeSubDomains" {
		t.Errorf("Expected HSTS header 'max-age=31536000; includeSubDomains', got '%s'", hsts)
	}
}

func TestSecurityHeadersMiddleware_NoHSTSForHTTP(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)
	// No TLS, so HTTP

	middleware := middleware.SecurityHeadersMiddleware()
	middleware(c)

	// HSTS should not be set for HTTP
	hsts := c.Writer.Header().Get("Strict-Transport-Security")
	if hsts != "" {
		t.Error("Strict-Transport-Security header should not be set for HTTP requests")
	}
}

