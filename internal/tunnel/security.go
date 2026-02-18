package tunnel

import (
	"net/http"
	"strings"

	"github.com/rs/zerolog"
)

type SecurityMiddleware struct {
	allowedOrigins map[string]bool
	logger         zerolog.Logger
}

func NewSecurityMiddleware(logger zerolog.Logger) *SecurityMiddleware {
	return &SecurityMiddleware{
		allowedOrigins: make(map[string]bool),
		logger:         logger,
	}
}

func (sm *SecurityMiddleware) AddAllowedOrigin(origin string) {
	sm.allowedOrigins[origin] = true
}

func (sm *SecurityMiddleware) ValidateOrigin(origin string) bool {
	if len(sm.allowedOrigins) == 0 {
		return true
	}
	return sm.allowedOrigins[origin]
}

func (sm *SecurityMiddleware) AddSecurityHeaders(w http.ResponseWriter, r *http.Request) {
	origin := r.Header.Get("Origin")
	if len(sm.allowedOrigins) == 0 {
		if origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		}
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Max-Age", "3600")
	} else if origin != "" && sm.ValidateOrigin(origin) {
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Max-Age", "3600")
	}

	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("X-XSS-Protection", "1; mode=block")
	w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}
}

func (sm *SecurityMiddleware) ValidateRequest(r *http.Request) error {
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

	if len(r.URL.Path) > 2048 {
		return ErrInvalidRequest
	}

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
	path = strings.ReplaceAll(path, "..", "")
	path = strings.ReplaceAll(path, "//", "/")
	path = strings.ReplaceAll(path, "\\", "/")

	path = strings.ReplaceAll(path, "\x00", "")
	path = strings.Map(func(r rune) rune {
		if r < 32 && r != '\t' && r != '\n' && r != '\r' {
			return -1 // Remove control characters except tab, newline, carriage return
		}
		return r
	}, path)

	path = strings.Trim(path, "/")
	if path == "" {
		path = "/"
	} else if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	if len(path) > 2048 {
		path = path[:2048]
	}
	
	return path
}

