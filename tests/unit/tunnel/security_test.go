package tunnel_test

import (
	"net/http/httptest"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"

	"github.com/Kizsoft-Solution-Limited/uniroute/internal/tunnel"
)

var (
	NewSecurityMiddleware = tunnel.NewSecurityMiddleware
)

func TestNewSecurityMiddleware(t *testing.T) {
	logger := zerolog.Nop()
	middleware := NewSecurityMiddleware(logger)

	assert.NotNil(t, middleware)
	// Note: allowedOrigins is unexported, so we can't test it directly
	// We test the exported behavior instead
}

func TestSecurityMiddleware_AddAllowedOrigin(t *testing.T) {
	logger := zerolog.Nop()
	middleware := NewSecurityMiddleware(logger)

	middleware.AddAllowedOrigin("https://example.com")
	// Test through ValidateOrigin instead of accessing unexported field
	assert.True(t, middleware.ValidateOrigin("https://example.com"))
}

func TestSecurityMiddleware_ValidateOrigin(t *testing.T) {
	logger := zerolog.Nop()
	middleware := NewSecurityMiddleware(logger)

	// No origins configured - should allow all
	assert.True(t, middleware.ValidateOrigin("https://example.com"))

	// Add origin
	middleware.AddAllowedOrigin("https://example.com")
	assert.True(t, middleware.ValidateOrigin("https://example.com"))
	assert.False(t, middleware.ValidateOrigin("https://other.com"))
}

func TestSecurityMiddleware_AddSecurityHeaders(t *testing.T) {
	logger := zerolog.Nop()
	middleware := NewSecurityMiddleware(logger)

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	middleware.AddSecurityHeaders(w, req)

	assert.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"))
	assert.Equal(t, "DENY", w.Header().Get("X-Frame-Options"))
	assert.Equal(t, "1; mode=block", w.Header().Get("X-XSS-Protection"))
}

func TestSecurityMiddleware_ValidateRequest(t *testing.T) {
	logger := zerolog.Nop()
	middleware := NewSecurityMiddleware(logger)

	// Valid request
	req := httptest.NewRequest("GET", "/test", nil)
	err := middleware.ValidateRequest(req)
	assert.NoError(t, err)

	// Invalid method
	req = httptest.NewRequest("INVALID", "/test", nil)
	err = middleware.ValidateRequest(req)
	assert.Error(t, err)
}

func TestSecurityMiddleware_SanitizePath(t *testing.T) {
	logger := zerolog.Nop()
	middleware := NewSecurityMiddleware(logger)

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"normal path", "/api/test", "/api/test"},
		{"path traversal", "/api/../test", "/api/test"},
		{"double slash", "/api//test", "/api/test"},
		{"null byte", "/api/test\x00", "/api/test"},
		{"combined", "/api/../test//path\x00", "/api/test/path"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := middleware.SanitizePath(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
