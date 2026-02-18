package security

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/Kizsoft-Solution-Limited/uniroute/internal/storage"
)

type APIKeyServiceV2 struct {
	repo   storage.APIKeyRepositoryInterface
	secret string
}

func NewAPIKeyServiceV2(repo storage.APIKeyRepositoryInterface, secret string) *APIKeyServiceV2 {
	return &APIKeyServiceV2{
		repo:   repo,
		secret: secret,
	}
}

func (s *APIKeyServiceV2) CreateAPIKey(ctx context.Context, userID uuid.UUID, name string, rateLimitPerMinute, rateLimitPerDay int, expiresAt *time.Time) (string, *storage.APIKey, error) {
	// Generate random bytes
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", nil, fmt.Errorf("failed to generate random bytes: %w", err)
	}

	// Create API key with prefix
	key := "ur_" + hex.EncodeToString(bytes)

	// Create lookup hash (SHA256) for fast database lookup
	lookupHash := s.hashForLookup(key)

	// Hash the key with bcrypt for secure storage
	bcryptHash, err := bcrypt.GenerateFromPassword([]byte(key), bcrypt.DefaultCost)
	if err != nil {
		return "", nil, fmt.Errorf("failed to hash API key: %w", err)
	}

	// Create API key record
	apiKey := &storage.APIKey{
		ID:                 uuid.New(),
		UserID:             userID,
		LookupHash:         lookupHash,         // SHA256 for fast lookup
		VerificationHash:   string(bcryptHash), // bcrypt for verification
		Name:               name,
		RateLimitPerMinute: rateLimitPerMinute,
		RateLimitPerDay:    rateLimitPerDay,
		CreatedAt:          time.Now(),
		ExpiresAt:          expiresAt,
		IsActive:           true,
	}

	// Store in database
	if err := s.repo.Create(ctx, apiKey); err != nil {
		return "", nil, fmt.Errorf("failed to store API key: %w", err)
	}

	return key, apiKey, nil
}

func (s *APIKeyServiceV2) ValidateAPIKey(ctx context.Context, key string) (*storage.APIKey, error) {
	// Create lookup hash to find the key
	lookupHash := s.hashForLookup(key)

	// Find by lookup hash
	apiKey, err := s.repo.FindByLookupHash(ctx, lookupHash)
	if err != nil {
		return nil, fmt.Errorf("failed to find API key: %w", err)
	}
	if apiKey == nil {
		return nil, fmt.Errorf("API key not found")
	}

	// Verify the key matches the stored bcrypt hash
	err = bcrypt.CompareHashAndPassword([]byte(apiKey.VerificationHash), []byte(key))
	if err != nil {
		return nil, fmt.Errorf("invalid API key")
	}

	return apiKey, nil
}

func (s *APIKeyServiceV2) ListAPIKeysByUser(ctx context.Context, userID uuid.UUID) ([]*storage.APIKey, error) {
	return s.repo.ListByUserID(ctx, userID)
}

func (s *APIKeyServiceV2) DeleteAPIKey(ctx context.Context, keyID uuid.UUID) error {
	return s.repo.Delete(ctx, keyID)
}

func (s *APIKeyServiceV2) hashForLookup(key string) string {
	hash := sha256.Sum256([]byte(key))
	return hex.EncodeToString(hash[:])
}
