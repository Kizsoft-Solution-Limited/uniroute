package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/Kizsoft-Solution-Limited/uniroute/internal/api/middleware"
)

func TestIPWhitelistMiddleware_AllowedIP(t *testing.T) {
	gin.SetMode(gin.TestMode)

	allowedIPs := []string{"192.168.1.100", "10.0.0.1"}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)
	c.Request.RemoteAddr = "192.168.1.100:12345"

	middleware := middleware.IPWhitelistMiddleware(allowedIPs)
	middleware(c)

	if c.Writer.Status() == http.StatusForbidden {
		t.Error("Request from allowed IP should not be forbidden")
	}
}

func TestIPWhitelistMiddleware_BlockedIP(t *testing.T) {
	gin.SetMode(gin.TestMode)

	allowedIPs := []string{"192.168.1.100", "10.0.0.1"}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)
	c.Request.RemoteAddr = "192.168.1.200:12345" // Not in whitelist

	middleware := middleware.IPWhitelistMiddleware(allowedIPs)
	middleware(c)

	if c.Writer.Status() != http.StatusForbidden {
		t.Errorf("Expected status %d, got %d", http.StatusForbidden, c.Writer.Status())
	}
}

func TestIPWhitelistMiddleware_EmptyWhitelist(t *testing.T) {
	gin.SetMode(gin.TestMode)

	allowedIPs := []string{} // Empty whitelist

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)
	c.Request.RemoteAddr = "192.168.1.100:12345"

	middleware := middleware.IPWhitelistMiddleware(allowedIPs)
	middleware(c)

	// Empty whitelist should allow all
	if c.Writer.Status() == http.StatusForbidden {
		t.Error("Empty whitelist should allow all IPs")
	}
}

func TestIPWhitelistMiddleware_NilWhitelist(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var allowedIPs []string = nil

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)
	c.Request.RemoteAddr = "192.168.1.100:12345"

	middleware := middleware.IPWhitelistMiddleware(allowedIPs)
	middleware(c)

	// Nil whitelist should allow all
	if c.Writer.Status() == http.StatusForbidden {
		t.Error("Nil whitelist should allow all IPs")
	}
}

