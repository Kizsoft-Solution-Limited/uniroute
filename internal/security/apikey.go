package security

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

type APIKeyService struct {
	secret string
	keys   map[string]string // In-memory storage (fallback when database is not available)
}

func NewAPIKeyService(secret string) *APIKeyService {
	return &APIKeyService{
		secret: secret,
		keys:   make(map[string]string),
	}
}

func (s *APIKeyService) GenerateAPIKey() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	key := "ur_" + hex.EncodeToString(bytes)

	hashed, err := bcrypt.GenerateFromPassword([]byte(key), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash API key: %w", err)
	}

	s.keys[key] = string(hashed)

	return key, nil
}

func (s *APIKeyService) ValidateAPIKey(key string) bool {
	hashed, exists := s.keys[key]
	if !exists {
		return false
	}

	err := bcrypt.CompareHashAndPassword([]byte(hashed), []byte(key))
	return err == nil
}

func ExtractAPIKey(authHeader string) string {
	if authHeader == "" {
		return ""
	}

	parts := strings.Fields(authHeader)
	if len(parts) >= 2 {
		if strings.ToLower(parts[0]) == "bearer" {
			return parts[1]
		}
	}
	if len(parts) > 0 {
		return parts[0]
	}
	return authHeader
}

