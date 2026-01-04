package providers

import (
	"testing"

	"github.com/rs/zerolog"
)

func TestNewGoogleProvider(t *testing.T) {
	provider := NewGoogleProvider("test-key", "", zerolog.Nop())
	if provider == nil {
		t.Fatal("NewGoogleProvider returned nil")
	}
	if provider.Name() != "google" {
		t.Errorf("Expected name 'google', got '%s'", provider.Name())
	}
}

func TestGoogleProvider_Name(t *testing.T) {
	provider := NewGoogleProvider("test-key", "", zerolog.Nop())
	if provider.Name() != "google" {
		t.Errorf("Expected name 'google', got '%s'", provider.Name())
	}
}

func TestGoogleProvider_GetModels(t *testing.T) {
	provider := NewGoogleProvider("test-key", "", zerolog.Nop())
	models := provider.GetModels()
	if len(models) == 0 {
		t.Error("Expected at least one model")
	}

	// Check for common models
	found := false
	for _, model := range models {
		if model == "gemini-pro" || model == "gemini-1.5-pro" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected to find common Google models")
	}
}

func TestGoogleProvider_Chat_NoAPIKey(t *testing.T) {
	provider := NewGoogleProvider("", "", zerolog.Nop())
	_, err := provider.Chat(nil, ChatRequest{
		Model: "gemini-pro",
		Messages: []Message{
			{Role: "user", Content: "test"},
		},
	})
	if err == nil {
		t.Error("Expected error when API key is not configured")
	}
}
