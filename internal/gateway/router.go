package gateway

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/Kizsoft-Solution-Limited/uniroute/internal/providers"
	"github.com/Kizsoft-Solution-Limited/uniroute/pkg/errors"
)

// Router routes requests to appropriate providers with failover support
type Router struct {
	providers            map[string]providers.Provider
	defaultProvider      providers.Provider
	strategy             RoutingStrategy
	costCalculator       *CostCalculator
	latencyTracker       *LatencyTracker
	providerKeyService   ProviderKeyServiceInterface // BYOK: For user-specific provider keys
	serverProviderKeys   ServerProviderKeys          // Server-level keys (fallback)
}

// ProviderKeyServiceInterface defines interface for getting user provider keys
type ProviderKeyServiceInterface interface {
	GetProviderKey(ctx context.Context, userID uuid.UUID, provider string) (string, error)
}

// ServerProviderKeys holds server-level provider keys (fallback)
type ServerProviderKeys struct {
	OpenAI    string
	Anthropic string
	Google    string
}

// NewRouter creates a new router
func NewRouter() *Router {
	return &Router{
		providers:          make(map[string]providers.Provider),
		strategy:           &ModelBasedStrategy{}, // Default strategy
		costCalculator:     NewCostCalculator(),
		latencyTracker:     NewLatencyTracker(100),
		providerKeyService: nil, // Will be set if BYOK is enabled
		serverProviderKeys: ServerProviderKeys{}, // Will be set from config
	}
}

// SetProviderKeyService sets the provider key service for BYOK
func (r *Router) SetProviderKeyService(service ProviderKeyServiceInterface) {
	r.providerKeyService = service
}

// SetServerProviderKeys sets server-level provider keys (fallback)
func (r *Router) SetServerProviderKeys(keys ServerProviderKeys) {
	r.serverProviderKeys = keys
}

// SetStrategy sets the routing strategy
func (r *Router) SetStrategy(strategy RoutingStrategy) {
	r.strategy = strategy
}

// SetStrategyType sets the routing strategy by type
func (r *Router) SetStrategyType(strategyType StrategyType) {
	switch strategyType {
	case StrategyCostBased:
		r.strategy = NewCostBasedStrategy(r.costCalculator)
	case StrategyLatencyBased:
		r.strategy = NewLatencyBasedStrategy(r.latencyTracker)
	case StrategyLoadBalanced:
		r.strategy = NewLoadBalancedStrategy()
	case StrategyModelBased:
		r.strategy = &ModelBasedStrategy{}
	default:
		r.strategy = &ModelBasedStrategy{}
	}
}

// RegisterProvider registers a provider with the router
func (r *Router) RegisterProvider(provider providers.Provider) {
	r.providers[provider.Name()] = provider
	// Set first registered provider as default (usually local)
	if r.defaultProvider == nil {
		r.defaultProvider = provider
	}
}

// Route routes a request to the appropriate provider with failover
// userID is optional - if provided, will use user's provider keys (BYOK)
func (r *Router) Route(ctx context.Context, req providers.ChatRequest, userID *uuid.UUID) (*providers.ChatResponse, error) {
	if len(r.providers) == 0 {
		return nil, fmt.Errorf("no providers available")
	}

	// Get available providers (healthy ones)
	availableProviders := r.getAvailableProviders(ctx, userID)

	if len(availableProviders) == 0 {
		return nil, fmt.Errorf("no healthy providers available")
	}

	// Phase 4: Use routing strategy to select provider
	selectedProvider, err := r.strategy.SelectProvider(ctx, req, availableProviders)
	if err != nil {
		return nil, fmt.Errorf("failed to select provider: %w", err)
	}

	// Try selected provider with failover
	providersToTry := []providers.Provider{selectedProvider}

	// Add other providers as fallbacks (excluding the selected one)
	for _, provider := range availableProviders {
		if provider.Name() != selectedProvider.Name() {
			providersToTry = append(providersToTry, provider)
		}
	}

	// Try each provider until one succeeds
	var lastErr error
	for _, provider := range providersToTry {
		start := time.Now()
		resp, err := provider.Chat(ctx, req)
		latency := time.Since(start)

		// Record latency
		r.latencyTracker.RecordLatency(provider.Name(), latency)

		if err == nil {
			// Add Phase 4 metadata
			resp.Provider = provider.Name()
			resp.LatencyMs = latency.Milliseconds()

			// Calculate actual cost if we have usage data
			if resp.Usage.TotalTokens > 0 {
				resp.Cost = r.costCalculator.CalculateActualCost(provider.Name(), resp.Model, resp.Usage)
			}
			return resp, nil
		}
		lastErr = err
	}

	// All providers failed
	if lastErr != nil {
		return nil, fmt.Errorf("all providers failed, last error: %w", lastErr)
	}

	return nil, fmt.Errorf("no providers available")
}

// getAvailableProviders returns list of healthy providers
// If userID is provided, uses user's provider keys (BYOK), otherwise uses server-level keys
func (r *Router) getAvailableProviders(ctx context.Context, userID *uuid.UUID) []providers.Provider {
	available := make([]providers.Provider, 0)
	
	// BYOK: If user has provider keys, create providers with user's keys
	if userID != nil && r.providerKeyService != nil {
		userProviders := r.getUserProviders(ctx, *userID)
		for _, provider := range userProviders {
			if err := provider.HealthCheck(ctx); err == nil {
				available = append(available, provider)
			}
		}
		// If user has providers, return them (don't fall back to server-level)
		if len(available) > 0 {
			return available
		}
	}
	
	// Fallback to server-level providers
	for _, provider := range r.providers {
		if err := provider.HealthCheck(ctx); err == nil {
			available = append(available, provider)
		}
	}
	return available
}

// getUserProviders creates providers using user's API keys (BYOK)
func (r *Router) getUserProviders(ctx context.Context, userID uuid.UUID) []providers.Provider {
	userProviders := make([]providers.Provider, 0)
	
	// Try to get user's keys for each provider
	providersToCheck := []string{"openai", "anthropic", "google"}
	
	for _, providerName := range providersToCheck {
		apiKey, err := r.providerKeyService.GetProviderKey(ctx, userID, providerName)
		if err != nil || apiKey == "" {
			continue // User doesn't have key for this provider
		}
		
		// Create provider with user's key (using zerolog.Nop() for now)
		var provider providers.Provider
		switch providerName {
		case "openai":
			provider = providers.NewOpenAIProvider(apiKey, "", zerolog.Nop())
		case "anthropic":
			provider = providers.NewAnthropicProvider(apiKey, "", zerolog.Nop())
		case "google":
			provider = providers.NewGoogleProvider(apiKey, "", zerolog.Nop())
		}
		
		if provider != nil {
			userProviders = append(userProviders, provider)
		}
	}
	
	return userProviders
}

// selectProvider selects the best provider for a given model
func (r *Router) selectProvider(model string) providers.Provider {
	// Model-to-provider mapping
	modelLower := strings.ToLower(model)

	// Check each provider's models
	for _, provider := range r.providers {
		models := provider.GetModels()
		for _, providerModel := range models {
			if strings.ToLower(providerModel) == modelLower ||
				strings.Contains(modelLower, strings.ToLower(provider.Name())) {
				return provider
			}
		}
	}

	// Default provider selection based on model prefix
	if strings.HasPrefix(modelLower, "gpt") {
		if provider, ok := r.providers["openai"]; ok {
			return provider
		}
	}
	if strings.HasPrefix(modelLower, "claude") {
		if provider, ok := r.providers["anthropic"]; ok {
			return provider
		}
	}
	if strings.HasPrefix(modelLower, "gemini") {
		if provider, ok := r.providers["google"]; ok {
			return provider
		}
	}

	// Fallback to default provider (usually local)
	return r.defaultProvider
}

// GetProvider returns a provider by name
func (r *Router) GetProvider(name string) (providers.Provider, error) {
	provider, exists := r.providers[name]
	if !exists {
		return nil, errors.ErrProviderNotFound
	}
	return provider, nil
}

// ListProviders returns all registered providers
func (r *Router) ListProviders() []string {
	names := make([]string, 0, len(r.providers))
	for name := range r.providers {
		names = append(names, name)
	}
	return names
}

// GetCostCalculator returns the cost calculator
func (r *Router) GetCostCalculator() *CostCalculator {
	return r.costCalculator
}

// GetLatencyTracker returns the latency tracker
func (r *Router) GetLatencyTracker() *LatencyTracker {
	return r.latencyTracker
}
