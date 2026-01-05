package security_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/Kizsoft-Solution-Limited/uniroute/internal/security"
	"github.com/Kizsoft-Solution-Limited/uniroute/internal/storage"
)

// mockProviderKeyRepository is a mock implementation of ProviderKeyRepositoryInterface
type mockProviderKeyRepository struct {
	keys map[string]*storage.UserProviderKey
}

func newMockProviderKeyRepository() *mockProviderKeyRepository {
	return &mockProviderKeyRepository{
		keys: make(map[string]*storage.UserProviderKey),
	}
}

func (m *mockProviderKeyRepository) Create(ctx context.Context, key *storage.UserProviderKey) error {
	keyStr := key.UserID.String() + ":" + key.Provider
	m.keys[keyStr] = key
	return nil
}

func (m *mockProviderKeyRepository) FindByUserAndProvider(ctx context.Context, userID uuid.UUID, provider string) (*storage.UserProviderKey, error) {
	keyStr := userID.String() + ":" + provider
	key, exists := m.keys[keyStr]
	if !exists {
		return nil, nil
	}
	return key, nil
}

func (m *mockProviderKeyRepository) ListByUserID(ctx context.Context, userID uuid.UUID) ([]*storage.UserProviderKey, error) {
	var result []*storage.UserProviderKey
	for _, key := range m.keys {
		if key.UserID == userID {
			result = append(result, key)
		}
	}
	return result, nil
}

func (m *mockProviderKeyRepository) Update(ctx context.Context, key *storage.UserProviderKey) error {
	keyStr := key.UserID.String() + ":" + key.Provider
	if _, exists := m.keys[keyStr]; exists {
		m.keys[keyStr] = key
	}
	return nil
}

func (m *mockProviderKeyRepository) Delete(ctx context.Context, userID uuid.UUID, provider string) error {
	keyStr := userID.String() + ":" + provider
	delete(m.keys, keyStr)
	return nil
}

func (m *mockProviderKeyRepository) DeleteByID(ctx context.Context, userID uuid.UUID, keyID uuid.UUID) error {
	for keyStr, key := range m.keys {
		if key.ID == keyID && key.UserID == userID {
			delete(m.keys, keyStr)
			return nil
		}
	}
	return nil
}

func TestNewProviderKeyService(t *testing.T) {
	repo := newMockProviderKeyRepository()
	encryptionKey := "test-encryption-key-min-32-chars-long"

	service, err := security.NewProviderKeyService(repo, encryptionKey)
	if err != nil {
		t.Fatalf("NewProviderKeyService failed: %v", err)
	}

	if service == nil {
		t.Fatal("NewProviderKeyService returned nil")
	}

	// Test that service was created successfully (encryption key is private)
	if service == nil {
		t.Fatal("Service should not be nil")
	}
}

func TestProviderKeyService_AddProviderKey(t *testing.T) {
	repo := newMockProviderKeyRepository()
	encryptionKey := "test-encryption-key-min-32-chars-long"
	service, _ := security.NewProviderKeyService(repo, encryptionKey)

	userID := uuid.New()
	provider := "openai"
	apiKey := "sk-test-key-12345"

	err := service.AddProviderKey(context.Background(), userID, provider, apiKey)
	if err != nil {
		t.Fatalf("AddProviderKey failed: %v", err)
	}

	// Verify key was stored
	stored, err := repo.FindByUserAndProvider(context.Background(), userID, provider)
	if err != nil {
		t.Fatalf("FindByUserAndProvider failed: %v", err)
	}
	if stored == nil {
		t.Fatal("Provider key was not stored")
	}

	// Verify key is encrypted
	if stored.APIKeyEncrypted == apiKey {
		t.Error("API key was not encrypted")
	}

	// Verify we can decrypt it
	decrypted, err := service.GetProviderKey(context.Background(), userID, provider)
	if err != nil {
		t.Fatalf("GetProviderKey failed: %v", err)
	}
	if decrypted != apiKey {
		t.Errorf("Decrypted key mismatch: expected %s, got %s", apiKey, decrypted)
	}
}

func TestProviderKeyService_AddProviderKey_InvalidProvider(t *testing.T) {
	repo := newMockProviderKeyRepository()
	encryptionKey := "test-encryption-key-min-32-chars-long"
	service, _ := security.NewProviderKeyService(repo, encryptionKey)

	userID := uuid.New()
	provider := "invalid-provider"
	apiKey := "sk-test-key-12345"

	err := service.AddProviderKey(context.Background(), userID, provider, apiKey)
	if err == nil {
		t.Error("Expected error for invalid provider, got nil")
	}
}

func TestProviderKeyService_GetProviderKey(t *testing.T) {
	repo := newMockProviderKeyRepository()
	encryptionKey := "test-encryption-key-min-32-chars-long"
	service, _ := security.NewProviderKeyService(repo, encryptionKey)

	userID := uuid.New()
	provider := "anthropic"
	apiKey := "sk-ant-test-key-12345"

	// Add key first
	err := service.AddProviderKey(context.Background(), userID, provider, apiKey)
	if err != nil {
		t.Fatalf("AddProviderKey failed: %v", err)
	}

	// Get key
	retrieved, err := service.GetProviderKey(context.Background(), userID, provider)
	if err != nil {
		t.Fatalf("GetProviderKey failed: %v", err)
	}

	if retrieved != apiKey {
		t.Errorf("Retrieved key mismatch: expected %s, got %s", apiKey, retrieved)
	}
}

func TestProviderKeyService_GetProviderKey_NotFound(t *testing.T) {
	repo := newMockProviderKeyRepository()
	encryptionKey := "test-encryption-key-min-32-chars-long"
	service, _ := security.NewProviderKeyService(repo, encryptionKey)

	userID := uuid.New()
	provider := "openai"

	retrieved, err := service.GetProviderKey(context.Background(), userID, provider)
	if err != nil {
		t.Fatalf("GetProviderKey failed: %v", err)
	}

	if retrieved != "" {
		t.Errorf("Expected empty string for non-existent key, got %s", retrieved)
	}
}

func TestProviderKeyService_ListProviderKeys(t *testing.T) {
	repo := newMockProviderKeyRepository()
	encryptionKey := "test-encryption-key-min-32-chars-long"
	service, _ := security.NewProviderKeyService(repo, encryptionKey)

	userID := uuid.New()

	// Add multiple keys
	service.AddProviderKey(context.Background(), userID, "openai", "sk-openai-123")
	service.AddProviderKey(context.Background(), userID, "anthropic", "sk-ant-123")
	service.AddProviderKey(context.Background(), userID, "google", "AIza-123")

	// List keys
	keys, err := service.ListProviderKeys(context.Background(), userID)
	if err != nil {
		t.Fatalf("ListProviderKeys failed: %v", err)
	}

	if len(keys) != 3 {
		t.Errorf("Expected 3 keys, got %d", len(keys))
	}

	// Verify providers
	providers := make(map[string]bool)
	for _, key := range keys {
		providers[key.Provider] = true
	}

	if !providers["openai"] || !providers["anthropic"] || !providers["google"] {
		t.Error("Not all providers found in list")
	}
}

func TestProviderKeyService_UpdateProviderKey(t *testing.T) {
	repo := newMockProviderKeyRepository()
	encryptionKey := "test-encryption-key-min-32-chars-long"
	service, _ := security.NewProviderKeyService(repo, encryptionKey)

	userID := uuid.New()
	provider := "openai"
	oldKey := "sk-old-key-123"
	newKey := "sk-new-key-456"

	// Add initial key
	err := service.AddProviderKey(context.Background(), userID, provider, oldKey)
	if err != nil {
		t.Fatalf("AddProviderKey failed: %v", err)
	}

	// Update key
	err = service.UpdateProviderKey(context.Background(), userID, provider, newKey)
	if err != nil {
		t.Fatalf("UpdateProviderKey failed: %v", err)
	}

	// Verify update
	retrieved, err := service.GetProviderKey(context.Background(), userID, provider)
	if err != nil {
		t.Fatalf("GetProviderKey failed: %v", err)
	}

	if retrieved != newKey {
		t.Errorf("Key was not updated: expected %s, got %s", newKey, retrieved)
	}
}

func TestProviderKeyService_UpdateProviderKey_CreateIfNotExists(t *testing.T) {
	repo := newMockProviderKeyRepository()
	encryptionKey := "test-encryption-key-min-32-chars-long"
	service, _ := security.NewProviderKeyService(repo, encryptionKey)

	userID := uuid.New()
	provider := "openai"
	apiKey := "sk-new-key-456"

	// Update non-existent key (should create it)
	err := service.UpdateProviderKey(context.Background(), userID, provider, apiKey)
	if err != nil {
		t.Fatalf("UpdateProviderKey failed: %v", err)
	}

	// Verify key was created
	retrieved, err := service.GetProviderKey(context.Background(), userID, provider)
	if err != nil {
		t.Fatalf("GetProviderKey failed: %v", err)
	}

	if retrieved != apiKey {
		t.Errorf("Key was not created: expected %s, got %s", apiKey, retrieved)
	}
}

func TestProviderKeyService_DeleteProviderKey(t *testing.T) {
	repo := newMockProviderKeyRepository()
	encryptionKey := "test-encryption-key-min-32-chars-long"
	service, _ := security.NewProviderKeyService(repo, encryptionKey)

	userID := uuid.New()
	provider := "openai"
	apiKey := "sk-test-key-123"

	// Add key
	err := service.AddProviderKey(context.Background(), userID, provider, apiKey)
	if err != nil {
		t.Fatalf("AddProviderKey failed: %v", err)
	}

	// Delete key
	err = service.DeleteProviderKey(context.Background(), userID, provider)
	if err != nil {
		t.Fatalf("DeleteProviderKey failed: %v", err)
	}

	// Verify deletion
	retrieved, err := service.GetProviderKey(context.Background(), userID, provider)
	if err != nil {
		t.Fatalf("GetProviderKey failed: %v", err)
	}

	if retrieved != "" {
		t.Error("Key was not deleted")
	}
}

func TestProviderKeyService_EncryptionDecryption(t *testing.T) {
	repo := newMockProviderKeyRepository()
	encryptionKey := "test-encryption-key-min-32-chars-long"
	service, _ := security.NewProviderKeyService(repo, encryptionKey)

	userID := uuid.New()
	provider := "openai"
	originalKey := "sk-test-key-very-long-key-value-123456789"

	// Add key
	err := service.AddProviderKey(context.Background(), userID, provider, originalKey)
	if err != nil {
		t.Fatalf("AddProviderKey failed: %v", err)
	}

	// Get stored key (encrypted)
	stored, err := repo.FindByUserAndProvider(context.Background(), userID, provider)
	if err != nil {
		t.Fatalf("FindByUserAndProvider failed: %v", err)
	}

	// Verify encryption
	if stored.APIKeyEncrypted == originalKey {
		t.Error("Key was not encrypted")
	}

	// Decrypt and verify
	decrypted, err := service.GetProviderKey(context.Background(), userID, provider)
	if err != nil {
		t.Fatalf("GetProviderKey failed: %v", err)
	}

	if decrypted != originalKey {
		t.Errorf("Decryption failed: expected %s, got %s", originalKey, decrypted)
	}
}

func TestProviderKeyService_MultipleUsers(t *testing.T) {
	repo := newMockProviderKeyRepository()
	encryptionKey := "test-encryption-key-min-32-chars-long"
	service, _ := security.NewProviderKeyService(repo, encryptionKey)

	user1ID := uuid.New()
	user2ID := uuid.New()
	provider := "openai"

	// Add keys for different users
	service.AddProviderKey(context.Background(), user1ID, provider, "sk-user1-key")
	service.AddProviderKey(context.Background(), user2ID, provider, "sk-user2-key")

	// Verify isolation
	key1, _ := service.GetProviderKey(context.Background(), user1ID, provider)
	key2, _ := service.GetProviderKey(context.Background(), user2ID, provider)

	if key1 == key2 {
		t.Error("User keys are not isolated")
	}

	if key1 != "sk-user1-key" {
		t.Errorf("User1 key mismatch: expected sk-user1-key, got %s", key1)
	}

	if key2 != "sk-user2-key" {
		t.Errorf("User2 key mismatch: expected sk-user2-key, got %s", key2)
	}
}

