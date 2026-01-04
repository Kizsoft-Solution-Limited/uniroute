package tunnel

import (
	"net/http"
	"strings"

	"github.com/rs/zerolog"
)

// SecurityMiddleware handles security-related HTTP headers and validation
type SecurityMiddleware struct {
	allowedOrigins map[string]bool
	logger         zerolog.Logger
}

// NewSecurityMiddleware creates a new security middleware
func NewSecurityMiddleware(logger zerolog.Logger) *SecurityMiddleware {
	return &SecurityMiddleware{
		allowedOrigins: make(map[string]bool),
		logger:         logger,
	}
}

// AddAllowedOrigin adds an allowed origin for CORS
func (sm *SecurityMiddleware) AddAllowedOrigin(origin string) {
	sm.allowedOrigins[origin] = true
}

// ValidateOrigin validates the Origin header
func (sm *SecurityMiddleware) ValidateOrigin(origin string) bool {
	if len(sm.allowedOrigins) == 0 {
		// No restrictions if no origins configured
		return true
	}
	return sm.allowedOrigins[origin]
}

// AddSecurityHeaders adds security headers to HTTP responses
func (sm *SecurityMiddleware) AddSecurityHeaders(w http.ResponseWriter, r *http.Request) {
	// CORS headers
	origin := r.Header.Get("Origin")
	if origin != "" && sm.ValidateOrigin(origin) {
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Max-Age", "3600")
	}

	// Security headers
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("X-XSS-Protection", "1; mode=block")
	w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

	// Handle preflight requests
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}
}

// ValidateRequest validates incoming HTTP requests
func (sm *SecurityMiddleware) ValidateRequest(r *http.Request) error {
	// Validate method
	allowedMethods := map[string]bool{
		http.MethodGet:     true,
		http.MethodPost:    true,
		http.MethodPut:     true,
		http.MethodDelete:  true,
		http.MethodPatch:   true,
		http.MethodOptions: true,
		http.MethodHead:    true,
	}

	if !allowedMethods[r.Method] {
		return ErrInvalidRequest
	}

	// Validate path length
	if len(r.URL.Path) > 2048 {
		return ErrInvalidRequest
	}

	// Validate header size (basic check)
	totalHeaderSize := 0
	for k, v := range r.Header {
		totalHeaderSize += len(k)
		for _, val := range v {
			totalHeaderSize += len(val)
		}
	}
	if totalHeaderSize > 8192 { // 8KB limit
		return ErrInvalidRequest
	}

	return nil
}

// SanitizePath sanitizes the request path
func (sm *SecurityMiddleware) SanitizePath(path string) string {
	// Remove path traversal attempts
	path = strings.ReplaceAll(path, "..", "")
	path = strings.ReplaceAll(path, "//", "/")
	
	// Remove null bytes
	path = strings.ReplaceAll(path, "\x00", "")
	
	return path
}

