package security

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"time"

	"github.com/google/uuid"
	"github.com/Kizsoft-Solution-Limited/uniroute/internal/storage"
)

// ProviderKeyService handles encryption and management of user provider keys
type ProviderKeyService struct {
	repo          ProviderKeyRepositoryInterface
	encryptionKey []byte // Master encryption key (should be from config)
}

// ProviderKeyRepositoryInterface defines the interface for provider key repository
type ProviderKeyRepositoryInterface interface {
	Create(ctx context.Context, key *storage.UserProviderKey) error
	FindByUserAndProvider(ctx context.Context, userID uuid.UUID, provider string) (*storage.UserProviderKey, error)
	ListByUserID(ctx context.Context, userID uuid.UUID) ([]*storage.UserProviderKey, error)
	Update(ctx context.Context, key *storage.UserProviderKey) error
	Delete(ctx context.Context, userID uuid.UUID, provider string) error
	DeleteByID(ctx context.Context, userID uuid.UUID, keyID uuid.UUID) error
}

// NewProviderKeyService creates a new provider key service
// encryptionKey should be a string that will be derived to 32-byte key for AES-256
func NewProviderKeyService(repo ProviderKeyRepositoryInterface, encryptionKey string) (*ProviderKeyService, error) {
	// Derive 32-byte key from string (using SHA256)
	keyBytes := deriveEncryptionKey(encryptionKey)
	
	return &ProviderKeyService{
		repo:          repo,
		encryptionKey: keyBytes,
	}, nil
}

// deriveEncryptionKey derives a 32-byte key from a string using SHA256
func deriveEncryptionKey(key string) []byte {
	hash := sha256.Sum256([]byte(key))
	return hash[:]
}

// AddProviderKey adds or updates a provider API key for a user
func (s *ProviderKeyService) AddProviderKey(ctx context.Context, userID uuid.UUID, provider string, apiKey string) error {
	// Validate provider
	if !isValidProvider(provider) {
		return fmt.Errorf("invalid provider: %s", provider)
	}

	// Encrypt the API key
	encrypted, err := s.encrypt(apiKey)
	if err != nil {
		return fmt.Errorf("failed to encrypt API key: %w", err)
	}

	// Create or update the key
	key := &storage.UserProviderKey{
		ID:              uuid.New(),
		UserID:          userID,
		Provider:        provider,
		APIKeyEncrypted: encrypted,
		IsActive:        true,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	return s.repo.Create(ctx, key)
}

// GetProviderKey retrieves and decrypts a provider API key for a user
func (s *ProviderKeyService) GetProviderKey(ctx context.Context, userID uuid.UUID, provider string) (string, error) {
	key, err := s.repo.FindByUserAndProvider(ctx, userID, provider)
	if err != nil {
		return "", fmt.Errorf("failed to find provider key: %w", err)
	}
	if key == nil {
		return "", nil // No key found
	}

	// Decrypt the API key
	decrypted, err := s.decrypt(key.APIKeyEncrypted)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt API key: %w", err)
	}

	return decrypted, nil
}

// ListProviderKeys lists all provider keys for a user (without decrypting)
func (s *ProviderKeyService) ListProviderKeys(ctx context.Context, userID uuid.UUID) ([]*storage.UserProviderKey, error) {
	return s.repo.ListByUserID(ctx, userID)
}

// UpdateProviderKey updates a provider API key
func (s *ProviderKeyService) UpdateProviderKey(ctx context.Context, userID uuid.UUID, provider string, apiKey string) error {
	// Get existing key
	existing, err := s.repo.FindByUserAndProvider(ctx, userID, provider)
	if err != nil {
		return fmt.Errorf("failed to find provider key: %w", err)
	}
	if existing == nil {
		// Create new if doesn't exist
		return s.AddProviderKey(ctx, userID, provider, apiKey)
	}

	// Encrypt the new API key
	encrypted, err := s.encrypt(apiKey)
	if err != nil {
		return fmt.Errorf("failed to encrypt API key: %w", err)
	}

	// Update
	existing.APIKeyEncrypted = encrypted
	existing.UpdatedAt = time.Now()
	existing.IsActive = true

	return s.repo.Update(ctx, existing)
}

// DeleteProviderKey removes a provider key (soft delete)
func (s *ProviderKeyService) DeleteProviderKey(ctx context.Context, userID uuid.UUID, provider string) error {
	return s.repo.Delete(ctx, userID, provider)
}

// DeleteProviderKeyByID removes a provider key by ID
func (s *ProviderKeyService) DeleteProviderKeyByID(ctx context.Context, userID uuid.UUID, keyID uuid.UUID) error {
	return s.repo.DeleteByID(ctx, userID, keyID)
}

// encrypt encrypts a plaintext API key using AES-256-GCM
func (s *ProviderKeyService) encrypt(plaintext string) (string, error) {
	block, err := aes.NewCipher(s.encryptionKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// decrypt decrypts an encrypted API key
func (s *ProviderKeyService) decrypt(encrypted string) (string, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(s.encryptionKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// isValidProvider validates that the provider name is supported
func isValidProvider(provider string) bool {
	validProviders := map[string]bool{
		"openai":    true,
		"anthropic": true,
		"google":    true,
	}
	return validProviders[provider]
}

