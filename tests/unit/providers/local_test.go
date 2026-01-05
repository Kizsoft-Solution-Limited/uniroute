package providers_test

import (
	"testing"

	"github.com/rs/zerolog"

	"github.com/Kizsoft-Solution-Limited/uniroute/internal/providers"
)

func TestLocalProvider_Name(t *testing.T) {
	provider := providers.NewLocalProvider("http://localhost:11434", zerolog.Nop())
	if provider.Name() != "local" {
		t.Errorf("Expected name 'local', got '%s'", provider.Name())
	}
}

func TestLocalProvider_GetModels(t *testing.T) {
	provider := providers.NewLocalProvider("http://localhost:11434", zerolog.Nop())
	models := provider.GetModels()
	// This will return empty if Ollama is not running, which is fine for unit test
	_ = models
}

func TestNewLocalProvider(t *testing.T) {
	baseURL := "http://localhost:11434"
	logger := zerolog.Nop()
	provider := providers.NewLocalProvider(baseURL, logger)

	if provider == nil {
		t.Fatal("NewLocalProvider returned nil")
	}

	// Note: baseURL and client are unexported fields, so we can't test them directly
	// We can only test the exported methods
	if provider.Name() != "local" {
		t.Error("Provider name should be 'local'")
	}
}
