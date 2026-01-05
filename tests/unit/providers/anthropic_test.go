package providers_test

import (
	"testing"

	"github.com/rs/zerolog"

	"github.com/Kizsoft-Solution-Limited/uniroute/internal/providers"
)

func TestNewAnthropicProvider(t *testing.T) {
	provider := providers.NewAnthropicProvider("test-key", "", zerolog.Nop())
	if provider == nil {
		t.Fatal("NewAnthropicProvider returned nil")
	}
	if provider.Name() != "anthropic" {
		t.Errorf("Expected name 'anthropic', got '%s'", provider.Name())
	}
}

func TestAnthropicProvider_Name(t *testing.T) {
	provider := providers.NewAnthropicProvider("test-key", "", zerolog.Nop())
	if provider.Name() != "anthropic" {
		t.Errorf("Expected name 'anthropic', got '%s'", provider.Name())
	}
}

func TestAnthropicProvider_GetModels(t *testing.T) {
	provider := providers.NewAnthropicProvider("test-key", "", zerolog.Nop())
	models := provider.GetModels()
	if len(models) == 0 {
		t.Error("Expected at least one model")
	}

	// Check for common models
	found := false
	for _, model := range models {
		if model == "claude-3-5-sonnet-20241022" || model == "claude-3-haiku-20240307" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected to find common Anthropic models")
	}
}

func TestAnthropicProvider_Chat_NoAPIKey(t *testing.T) {
	provider := providers.NewAnthropicProvider("", "", zerolog.Nop())
	_, err := provider.Chat(nil, providers.ChatRequest{
		Model: "claude-3-5-sonnet-20241022",
		Messages: []providers.Message{
			{Role: "user", Content: "test"},
		},
	})
	if err == nil {
		t.Error("Expected error when API key is not configured")
	}
}
