package providers

import (
	"context"
	"fmt"
	"strings"

	"github.com/rs/zerolog"
)

type KeyValidator struct {
	logger zerolog.Logger
}

func NewKeyValidator(logger zerolog.Logger) *KeyValidator {
	return &KeyValidator{logger: logger}
}

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
