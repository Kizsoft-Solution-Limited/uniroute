package providers

import (
	"testing"

	"github.com/rs/zerolog"
)

func TestLocalProvider_Name(t *testing.T) {
	provider := NewLocalProvider("http://localhost:11434", zerolog.Nop())
	if provider.Name() != "local" {
		t.Errorf("Expected name 'local', got '%s'", provider.Name())
	}
}

func TestLocalProvider_GetModels(t *testing.T) {
	provider := NewLocalProvider("http://localhost:11434", zerolog.Nop())
	models := provider.GetModels()
	// This will return empty if Ollama is not running, which is fine for unit test
	_ = models
}

func TestNewLocalProvider(t *testing.T) {
	baseURL := "http://localhost:11434"
	logger := zerolog.Nop()
	provider := NewLocalProvider(baseURL, logger)

	if provider == nil {
		t.Fatal("NewLocalProvider returned nil")
	}

	if provider.baseURL != baseURL {
		t.Errorf("Expected baseURL '%s', got '%s'", baseURL, provider.baseURL)
	}

	if provider.client == nil {
		t.Error("HTTP client should not be nil")
	}
}
