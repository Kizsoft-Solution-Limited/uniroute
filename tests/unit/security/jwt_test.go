package security_test

import (
	"testing"
	"time"

	"github.com/Kizsoft-Solution-Limited/uniroute/internal/security"
)

func TestNewJWTService(t *testing.T) {
	service := security.NewJWTService("test-secret-key")
	if service == nil {
		t.Fatal("NewJWTService returned nil")
	}
}

func TestJWTService_GenerateToken(t *testing.T) {
	service := security.NewJWTService("test-secret-key-min-32-chars")

	token, err := service.GenerateToken("user123", "test@example.com", []string{"user"}, 1*time.Hour)
	if err != nil {
		t.Fatalf("GenerateToken returned error: %v", err)
	}
	if token == "" {
		t.Error("Generated token should not be empty")
	}
	if len(token) < 50 {
		t.Error("Generated token should be reasonably long")
	}
}

func TestJWTService_ValidateToken(t *testing.T) {
	service := security.NewJWTService("test-secret-key-min-32-chars")

	// Generate a valid token
	userID := "user123"
	email := "test@example.com"
	token, err := service.GenerateToken(userID, email, []string{"user"}, 1*time.Hour)
	if err != nil {
		t.Fatalf("GenerateToken returned error: %v", err)
	}

	// Validate the token
	claims, err := service.ValidateToken(token)
	if err != nil {
		t.Fatalf("ValidateToken returned error: %v", err)
	}
	if claims == nil {
		t.Fatal("Claims should not be nil")
	}
	if claims.UserID != userID {
		t.Errorf("Expected UserID %s, got %s", userID, claims.UserID)
	}
	if claims.Email != email {
		t.Errorf("Expected Email %s, got %s", email, claims.Email)
	}
}

func TestJWTService_ValidateToken_Expired(t *testing.T) {
	service := security.NewJWTService("test-secret-key-min-32-chars")

	// Generate an expired token
	token, err := service.GenerateToken("user123", "test@example.com", []string{"user"}, -1*time.Hour)
	if err != nil {
		t.Fatalf("GenerateToken returned error: %v", err)
	}

	// Validate should fail
	claims, err := service.ValidateToken(token)
	if err == nil {
		t.Error("ValidateToken should return error for expired token")
	}
	if err != security.ErrExpiredToken {
		t.Errorf("Expected ErrExpiredToken, got %v", err)
	}
	if claims != nil {
		t.Error("Claims should be nil for expired token")
	}
}

func TestJWTService_ValidateToken_Invalid(t *testing.T) {
	service := security.NewJWTService("test-secret-key-min-32-chars")

	tests := []struct {
		name  string
		token string
	}{
		{
			name:  "Empty token",
			token: "",
		},
		{
			name:  "Invalid format",
			token: "invalid.token.here",
		},
		{
			name:  "Wrong secret",
			token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiMTIzIn0.invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := service.ValidateToken(tt.token)
			if err == nil {
				t.Error("ValidateToken should return error for invalid token")
			}
			if err != security.ErrInvalidToken && err != security.ErrExpiredToken {
				t.Errorf("Expected ErrInvalidToken or ErrExpiredToken, got %v", err)
			}
			if claims != nil {
				t.Error("Claims should be nil for invalid token")
			}
		})
	}
}

func TestJWTService_ValidateToken_WrongSecret(t *testing.T) {
	service1 := security.NewJWTService("secret-key-1-min-32-chars")
	service2 := security.NewJWTService("secret-key-2-min-32-chars")

	// Generate token with service1
	token, err := service1.GenerateToken("user123", "test@example.com", []string{"user"}, 1*time.Hour)
	if err != nil {
		t.Fatalf("GenerateToken returned error: %v", err)
	}

	// Try to validate with service2 (different secret)
	claims, err := service2.ValidateToken(token)
	if err == nil {
		t.Error("ValidateToken should return error for token signed with different secret")
	}
	if claims != nil {
		t.Error("Claims should be nil for token with wrong secret")
	}
}
