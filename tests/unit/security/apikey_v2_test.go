package security_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/Kizsoft-Solution-Limited/uniroute/internal/security"
	"github.com/Kizsoft-Solution-Limited/uniroute/internal/storage"
)

// Ensure mockAPIKeyRepository implements APIKeyRepositoryInterface
var _ storage.APIKeyRepositoryInterface = (*mockAPIKeyRepository)(nil)

// mockAPIKeyRepository is a mock repository for testing
type mockAPIKeyRepository struct {
	keys map[string]*storage.APIKey
}

func newMockAPIKeyRepository() *mockAPIKeyRepository {
	return &mockAPIKeyRepository{
		keys: make(map[string]*storage.APIKey),
	}
}

func (m *mockAPIKeyRepository) Create(ctx context.Context, key *storage.APIKey) error {
	m.keys[key.LookupHash] = key
	return nil
}

func (m *mockAPIKeyRepository) FindByLookupHash(ctx context.Context, lookupHash string) (*storage.APIKey, error) {
	key, ok := m.keys[lookupHash]
	if !ok {
		return nil, nil
	}
	// Filter inactive keys (matching real repository behavior)
	if !key.IsActive {
		return nil, nil
	}
	// Filter expired keys (matching real repository behavior)
	if key.ExpiresAt != nil && key.ExpiresAt.Before(time.Now()) {
		return nil, nil
	}
	return key, nil
}

func (m *mockAPIKeyRepository) ListByUserID(ctx context.Context, userID uuid.UUID) ([]*storage.APIKey, error) {
	var keys []*storage.APIKey
	for _, key := range m.keys {
		if key.UserID == userID {
			keys = append(keys, key)
		}
	}
	return keys, nil
}

func (m *mockAPIKeyRepository) Update(ctx context.Context, key *storage.APIKey) error {
	m.keys[key.LookupHash] = key
	return nil
}

func (m *mockAPIKeyRepository) Delete(ctx context.Context, id uuid.UUID) error {
	for hash, key := range m.keys {
		if key.ID == id {
			key.IsActive = false
			m.keys[hash] = key
			break
		}
	}
	return nil
}

func TestNewAPIKeyServiceV2(t *testing.T) {
	repo := newMockAPIKeyRepository()
	service := security.NewAPIKeyServiceV2(repo, "test-secret")
	if service == nil {
		t.Fatal("NewAPIKeyServiceV2 returned nil")
	}
}

func TestAPIKeyServiceV2_CreateAPIKey(t *testing.T) {
	repo := newMockAPIKeyRepository()
	service := security.NewAPIKeyServiceV2(repo, "test-secret")
	ctx := context.Background()

	userID := uuid.New()
	name := "Test Key"
	rateLimitPerMinute := 60
	rateLimitPerDay := 10000

	key, apiKey, err := service.CreateAPIKey(ctx, userID, name, rateLimitPerMinute, rateLimitPerDay, nil)
	if err != nil {
		t.Fatalf("CreateAPIKey returned error: %v", err)
	}
	if key == "" {
		t.Error("Generated key should not be empty")
	}
	if len(key) < 10 {
		t.Error("Generated key should be reasonably long")
	}
	if key[:3] != "ur_" {
		t.Error("API key should have 'ur_' prefix")
	}
	if apiKey == nil {
		t.Fatal("API key record should not be nil")
	}
	if apiKey.Name != name {
		t.Errorf("Expected name %s, got %s", name, apiKey.Name)
	}
	if apiKey.RateLimitPerMinute != rateLimitPerMinute {
		t.Errorf("Expected rate limit per minute %d, got %d", rateLimitPerMinute, apiKey.RateLimitPerMinute)
	}
	if apiKey.RateLimitPerDay != rateLimitPerDay {
		t.Errorf("Expected rate limit per day %d, got %d", rateLimitPerDay, apiKey.RateLimitPerDay)
	}
	if apiKey.UserID != userID {
		t.Errorf("Expected user ID %s, got %s", userID, apiKey.UserID)
	}
	if !apiKey.IsActive {
		t.Error("API key should be active by default")
	}
}

func TestAPIKeyServiceV2_CreateAPIKey_WithExpiration(t *testing.T) {
	repo := newMockAPIKeyRepository()
	service := security.NewAPIKeyServiceV2(repo, "test-secret")
	ctx := context.Background()

	userID := uuid.New()
	expiresAt := time.Now().Add(24 * time.Hour)

	key, apiKey, err := service.CreateAPIKey(ctx, userID, "Test Key", 60, 10000, &expiresAt)
	if err != nil {
		t.Fatalf("CreateAPIKey returned error: %v", err)
	}
	if key == "" {
		t.Error("Generated key should not be empty")
	}
	if apiKey.ExpiresAt == nil {
		t.Error("ExpiresAt should be set")
	}
	if !apiKey.ExpiresAt.Equal(expiresAt) {
		t.Error("ExpiresAt should match provided value")
	}
}

func TestAPIKeyServiceV2_ValidateAPIKey(t *testing.T) {
	repo := newMockAPIKeyRepository()
	service := security.NewAPIKeyServiceV2(repo, "test-secret")
	ctx := context.Background()

	userID := uuid.New()

	// Create an API key
	key, _, err := service.CreateAPIKey(ctx, userID, "Test Key", 60, 10000, nil)
	if err != nil {
		t.Fatalf("CreateAPIKey returned error: %v", err)
	}

	// Validate the key
	apiKey, err := service.ValidateAPIKey(ctx, key)
	if err != nil {
		t.Fatalf("ValidateAPIKey returned error: %v", err)
	}
	if apiKey == nil {
		t.Fatal("API key should not be nil")
	}
	if apiKey.UserID != userID {
		t.Errorf("Expected user ID %s, got %s", userID, apiKey.UserID)
	}
}

func TestAPIKeyServiceV2_ValidateAPIKey_Invalid(t *testing.T) {
	repo := newMockAPIKeyRepository()
	service := security.NewAPIKeyServiceV2(repo, "test-secret")
	ctx := context.Background()

	tests := []struct {
		name string
		key  string
	}{
		{
			name: "Non-existent key",
			key:  "ur_nonexistentkey123456789012345678901234567890",
		},
		{
			name: "Invalid format",
			key:  "invalid-key",
		},
		{
			name: "Empty key",
			key:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			apiKey, err := service.ValidateAPIKey(ctx, tt.key)
			if err == nil {
				t.Error("ValidateAPIKey should return error for invalid key")
			}
			if apiKey != nil {
				t.Error("API key should be nil for invalid key")
			}
		})
	}
}

func TestAPIKeyServiceV2_ValidateAPIKey_Inactive(t *testing.T) {
	repo := newMockAPIKeyRepository()
	service := security.NewAPIKeyServiceV2(repo, "test-secret")
	ctx := context.Background()

	userID := uuid.New()

	// Create an API key
	key, apiKey, err := service.CreateAPIKey(ctx, userID, "Test Key", 60, 10000, nil)
	if err != nil {
		t.Fatalf("CreateAPIKey returned error: %v", err)
	}

	// Deactivate the key
	apiKey.IsActive = false
	repo.Update(ctx, apiKey)

	// Validation should fail
	_, err = service.ValidateAPIKey(ctx, key)
	if err == nil {
		t.Error("ValidateAPIKey should return error for inactive key")
	}
}

func TestAPIKeyServiceV2_ValidateAPIKey_Expired(t *testing.T) {
	repo := newMockAPIKeyRepository()
	service := security.NewAPIKeyServiceV2(repo, "test-secret")
	ctx := context.Background()

	userID := uuid.New()
	expiresAt := time.Now().Add(-1 * time.Hour) // Expired

	// Create an expired API key
	key, apiKey, err := service.CreateAPIKey(ctx, userID, "Test Key", 60, 10000, &expiresAt)
	if err != nil {
		t.Fatalf("CreateAPIKey returned error: %v", err)
	}

	// The repository should filter out expired keys
	// But for testing, we'll manually set it
	apiKey.ExpiresAt = &expiresAt
	repo.Update(ctx, apiKey)

	// Validation should fail (repository should return nil for expired keys)
	// Note: The actual repository implementation filters expired keys
	_, err = service.ValidateAPIKey(ctx, key)
	if err == nil {
		t.Error("ValidateAPIKey should return error for expired key")
	}
}

