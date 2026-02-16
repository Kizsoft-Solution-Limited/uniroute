package providers

import (
	"context"
	"fmt"
	"strings"

	"github.com/rs/zerolog"
)

// KeyValidator validates provider API keys by calling each provider's HealthCheck.
type KeyValidator struct {
	logger zerolog.Logger
}

// NewKeyValidator returns a validator that uses provider HealthCheck to test keys.
func NewKeyValidator(logger zerolog.Logger) *KeyValidator {
	return &KeyValidator{logger: logger}
}

// ValidateKey checks that the given API key is valid for the provider by making a minimal API call.
func (v *KeyValidator) ValidateKey(ctx context.Context, provider string, apiKey string) error {
	if apiKey == "" {
		return fmt.Errorf("API key is empty")
	}
	provider = strings.ToLower(strings.TrimSpace(provider))
	switch provider {
	case "openai":
		p := NewOpenAIProvider(apiKey, "", v.logger)
		return p.HealthCheck(ctx)
	case "anthropic":
		p := NewAnthropicProvider(apiKey, "", v.logger)
		return p.HealthCheck(ctx)
	case "google":
		p := NewGoogleProvider(apiKey, "", v.logger)
		return p.HealthCheck(ctx)
	default:
		return fmt.Errorf("unsupported provider: %s", provider)
	}
}
