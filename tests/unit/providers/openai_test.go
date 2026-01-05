package providers_test

import (
	"testing"

	"github.com/rs/zerolog"

	"github.com/Kizsoft-Solution-Limited/uniroute/internal/providers"
)

func TestNewOpenAIProvider(t *testing.T) {
	provider := providers.NewOpenAIProvider("test-key", "", zerolog.Nop())
	if provider == nil {
		t.Fatal("NewOpenAIProvider returned nil")
	}
	if provider.Name() != "openai" {
		t.Errorf("Expected name 'openai', got '%s'", provider.Name())
	}
	// Note: apiKey is unexported, so we can't test it directly
}

func TestOpenAIProvider_Name(t *testing.T) {
	provider := providers.NewOpenAIProvider("test-key", "", zerolog.Nop())
	if provider.Name() != "openai" {
		t.Errorf("Expected name 'openai', got '%s'", provider.Name())
	}
}

func TestOpenAIProvider_GetModels(t *testing.T) {
	provider := providers.NewOpenAIProvider("test-key", "", zerolog.Nop())
	models := provider.GetModels()
	if len(models) == 0 {
		t.Error("Expected at least one model")
	}

	// Check for common models
	found := false
	for _, model := range models {
		if model == "gpt-4" || model == "gpt-3.5-turbo" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected to find common OpenAI models")
	}
}

func TestOpenAIProvider_Chat_NoAPIKey(t *testing.T) {
	provider := providers.NewOpenAIProvider("", "", zerolog.Nop())
	_, err := provider.Chat(nil, providers.ChatRequest{
		Model: "gpt-4",
		Messages: []providers.Message{
			{Role: "user", Content: "test"},
		},
	})
	if err == nil {
		t.Error("Expected error when API key is not configured")
	}
}
