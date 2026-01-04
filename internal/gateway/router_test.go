package gateway

import (
	"context"
	"testing"

	"github.com/Kizsoft-Solution-Limited/uniroute/internal/providers"
	"github.com/Kizsoft-Solution-Limited/uniroute/pkg/errors"
)

// mockProvider is a mock provider for testing
type mockProvider struct {
	name      string
	available bool
	models    []string // Custom models list
}

func (m *mockProvider) Name() string {
	return m.name
}

func (m *mockProvider) Chat(ctx context.Context, req providers.ChatRequest) (*providers.ChatResponse, error) {
	if !m.available {
		return nil, errors.ErrProviderUnavailable
	}
	return &providers.ChatResponse{
		ID:    "test-123",
		Model: req.Model,
		Choices: []providers.Choice{
			{
				Message: providers.Message{
					Role:    "assistant",
					Content: "Test response",
				},
			},
		},
	}, nil
}

func (m *mockProvider) HealthCheck(ctx context.Context) error {
	if !m.available {
		return errors.ErrProviderUnavailable
	}
	return nil
}

func (m *mockProvider) GetModels() []string {
	if len(m.models) > 0 {
		return m.models
	}
	return []string{"test-model"}
}

func TestNewRouter(t *testing.T) {
	router := NewRouter()
	if router == nil {
		t.Fatal("NewRouter returned nil")
	}
	if router.providers == nil {
		t.Error("Providers map should be initialized")
	}
}

func TestRouter_RegisterProvider(t *testing.T) {
	router := NewRouter()
	provider := &mockProvider{name: "test", available: true}

	router.RegisterProvider(provider)

	if router.defaultProvider == nil {
		t.Error("Default provider should be set")
	}

	if router.providers["test"] == nil {
		t.Error("Provider should be registered")
	}
}

func TestRouter_GetProvider(t *testing.T) {
	router := NewRouter()
	provider := &mockProvider{name: "test", available: true}
	router.RegisterProvider(provider)

	got, err := router.GetProvider("test")
	if err != nil {
		t.Errorf("GetProvider returned error: %v", err)
	}
	if got == nil {
		t.Error("GetProvider returned nil")
	}

	_, err = router.GetProvider("nonexistent")
	if err != errors.ErrProviderNotFound {
		t.Errorf("Expected ErrProviderNotFound, got %v", err)
	}
}

func TestRouter_Route(t *testing.T) {
	router := NewRouter()
	provider := &mockProvider{name: "test", available: true}
	router.RegisterProvider(provider)

	req := providers.ChatRequest{
		Model: "test-model",
		Messages: []providers.Message{
			{Role: "user", Content: "Hello"},
		},
	}

	resp, err := router.Route(context.Background(), req, nil)
	if err != nil {
		t.Errorf("Route returned error: %v", err)
	}
	if resp == nil {
		t.Fatal("Route returned nil response")
	}
	if resp.ID != "test-123" {
		t.Errorf("Expected ID 'test-123', got '%s'", resp.ID)
	}
}

func TestRouter_Route_NoProvider(t *testing.T) {
	router := NewRouter()

	req := providers.ChatRequest{
		Model: "test-model",
		Messages: []providers.Message{
			{Role: "user", Content: "Hello"},
		},
	}

	_, err := router.Route(context.Background(), req)
	if err == nil {
		t.Error("Expected error when no providers are registered")
	}
}

func TestRouter_Route_UnavailableProvider(t *testing.T) {
	router := NewRouter()
	provider := &mockProvider{name: "test", available: false}
	router.RegisterProvider(provider)

	req := providers.ChatRequest{
		Model: "test-model",
		Messages: []providers.Message{
			{Role: "user", Content: "Hello"},
		},
	}

	_, err := router.Route(context.Background(), req)
	if err == nil {
		t.Error("Expected error when provider is unavailable")
	}
}

// Phase 3: Multi-provider tests

func TestRouter_ListProviders(t *testing.T) {
	router := NewRouter()
	router.RegisterProvider(&mockProvider{name: "provider1", available: true})
	router.RegisterProvider(&mockProvider{name: "provider2", available: true})

	providers := router.ListProviders()
	if len(providers) != 2 {
		t.Errorf("Expected 2 providers, got %d", len(providers))
	}
}

func TestRouter_Route_Failover(t *testing.T) {
	router := NewRouter()
	// Primary provider (unavailable)
	router.RegisterProvider(&mockProvider{name: "primary", available: false})
	// Backup provider (available)
	router.RegisterProvider(&mockProvider{name: "backup", available: true})

	req := providers.ChatRequest{
		Model: "test-model",
		Messages: []providers.Message{
			{Role: "user", Content: "Hello"},
		},
	}

	resp, err := router.Route(context.Background(), req, nil)
	if err != nil {
		t.Errorf("Route should succeed with failover, got error: %v", err)
	}
	if resp == nil {
		t.Error("Route should return response from backup provider")
	}
}

func TestRouter_SelectProvider_ByModel(t *testing.T) {
	router := NewRouter()

	// Create providers with specific models
	openAIProvider := &mockProvider{
		name:      "openai",
		available: true,
		models:    []string{"gpt-4", "gpt-3.5-turbo"},
	}

	localProvider := &mockProvider{
		name:      "local",
		available: true,
		models:    []string{"llama2", "mistral"},
	}

	router.RegisterProvider(localProvider)
	router.RegisterProvider(openAIProvider)

	// Test GPT model selection
	req := providers.ChatRequest{
		Model: "gpt-4",
		Messages: []providers.Message{
			{Role: "user", Content: "Hello"},
		},
	}

	// Should select OpenAI provider
	selected := router.selectProvider(req.Model)
	if selected == nil {
		t.Fatal("Should select a provider")
	}
	if selected.Name() != "openai" {
		t.Errorf("Expected OpenAI provider for GPT model, got %s", selected.Name())
	}
}
