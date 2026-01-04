package security

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

// APIKeyService handles API key generation and validation
type APIKeyService struct {
	secret string
	keys   map[string]string // In-memory storage for Phase 1 (will use DB in Phase 2)
}

// NewAPIKeyService creates a new API key service
func NewAPIKeyService(secret string) *APIKeyService {
	return &APIKeyService{
		secret: secret,
		keys:   make(map[string]string),
	}
}

// GenerateAPIKey generates a new API key
func (s *APIKeyService) GenerateAPIKey() (string, error) {
	// Generate random bytes
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	// Create API key with prefix
	key := "ur_" + hex.EncodeToString(bytes)

	// Hash the key for storage
	hashed, err := bcrypt.GenerateFromPassword([]byte(key), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash API key: %w", err)
	}

	// Store hashed key (in-memory for Phase 1)
	s.keys[key] = string(hashed)

	return key, nil
}

// ValidateAPIKey validates an API key
func (s *APIKeyService) ValidateAPIKey(key string) bool {
	// Check if key exists
	hashed, exists := s.keys[key]
	if !exists {
		return false
	}

	// Verify the key matches the hash
	err := bcrypt.CompareHashAndPassword([]byte(hashed), []byte(key))
	return err == nil
}

// ExtractAPIKey extracts API key from Authorization header
func ExtractAPIKey(authHeader string) string {
	if authHeader == "" {
		return ""
	}

	// Support "Bearer <key>" or just "<key>"
	parts := strings.Fields(authHeader) // Fields handles multiple spaces
	if len(parts) >= 2 {
		// If it starts with "Bearer", return the second part
		if strings.ToLower(parts[0]) == "bearer" {
			return parts[1]
		}
	}
	// If no "Bearer" prefix, return the whole string (or first part)
	if len(parts) > 0 {
		return parts[0]
	}
	return authHeader
}

