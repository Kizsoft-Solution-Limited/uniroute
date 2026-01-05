package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/Kizsoft-Solution-Limited/uniroute/internal/api/middleware"
	"github.com/Kizsoft-Solution-Limited/uniroute/internal/security"
)

func TestJWTAuthMiddleware_ValidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	jwtService := security.NewJWTService("test-secret-key-min-32-chars")
	token, err := jwtService.GenerateToken("user123", "test@example.com", []string{"user"}, 1*time.Hour)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)
	c.Request.Header.Set("Authorization", "Bearer "+token)

	middleware := middleware.JWTAuthMiddleware(jwtService)
	middleware(c)

	if c.Writer.Status() == http.StatusUnauthorized {
		t.Error("Request should be authorized with valid token")
	}

	// Check that user info is set in context
	userID, exists := c.Get("user_id")
	if !exists {
		t.Error("user_id should be set in context")
	}
	if userID != "user123" {
		t.Errorf("Expected user_id 'user123', got '%s'", userID)
	}

	email, exists := c.Get("user_email")
	if !exists {
		t.Error("user_email should be set in context")
	}
	if email != "test@example.com" {
		t.Errorf("Expected email 'test@example.com', got '%s'", email)
	}
}

func TestJWTAuthMiddleware_NoToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	jwtService := security.NewJWTService("test-secret-key-min-32-chars")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)

	middleware := middleware.JWTAuthMiddleware(jwtService)
	middleware(c)

	if c.Writer.Status() != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, c.Writer.Status())
	}
}

func TestJWTAuthMiddleware_InvalidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	jwtService := security.NewJWTService("test-secret-key-min-32-chars")

	tests := []struct {
		name  string
		token string
	}{
		{
			name:  "Invalid format",
			token: "invalid.token.here",
		},
		{
			name:  "Empty token",
			token: "",
		},
		{
			name:  "No Bearer prefix",
			token: "just-a-token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/test", nil)
			
			authHeader := "Bearer " + tt.token
			if tt.token == "" {
				authHeader = ""
			} else if tt.name == "No Bearer prefix" {
				authHeader = tt.token
			}
			c.Request.Header.Set("Authorization", authHeader)

			middleware := middleware.JWTAuthMiddleware(jwtService)
			middleware(c)

			if c.Writer.Status() != http.StatusUnauthorized {
				t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, c.Writer.Status())
			}
		})
	}
}

func TestJWTAuthMiddleware_ExpiredToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	jwtService := security.NewJWTService("test-secret-key-min-32-chars")
	token, err := jwtService.GenerateToken("user123", "test@example.com", []string{"user"}, -1*time.Hour) // Expired
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)
	c.Request.Header.Set("Authorization", "Bearer "+token)

	middleware := middleware.JWTAuthMiddleware(jwtService)
	middleware(c)

	if c.Writer.Status() != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, c.Writer.Status())
	}
}

