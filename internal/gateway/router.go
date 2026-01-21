package gateway

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Kizsoft-Solution-Limited/uniroute/internal/providers"
	"github.com/Kizsoft-Solution-Limited/uniroute/pkg/errors"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// RoutingStrategyServiceInterface defines interface for getting routing strategy
type RoutingStrategyServiceInterface interface {
	GetDefaultRoutingStrategy(ctx context.Context) (string, error)
	IsRoutingStrategyLocked(ctx context.Context) (bool, error)
}

// UserRoutingStrategyServiceInterface defines interface for getting user routing strategy
type UserRoutingStrategyServiceInterface interface {
	GetUserRoutingStrategy(ctx context.Context, userID uuid.UUID) (string, error)
}

// CustomRulesServiceInterface defines interface for loading custom routing rules
type CustomRulesServiceInterface interface {
	GetActiveRulesForUser(ctx context.Context, userID *uuid.UUID) ([]CustomRule, error)
}

// CustomRule represents a custom routing rule (simplified for router)
type CustomRule struct {
	ConditionType  string
	ConditionValue map[string]interface{}
	ProviderName   string
	Priority       int
}

// Router routes requests to appropriate providers with failover support
type Router struct {
	providers                  map[string]providers.Provider
	defaultProvider            providers.Provider
	strategy                   RoutingStrategy
	currentStrategyType        StrategyType // Track current strategy type explicitly (default)
	costCalculator             *CostCalculator
	latencyTracker             *LatencyTracker
	providerKeyService         ProviderKeyServiceInterface // BYOK: For user-specific provider keys
	serverProviderKeys         ServerProviderKeys          // Server-level keys (fallback)
	routingStrategyService     RoutingStrategyServiceInterface
	userRoutingStrategyService UserRoutingStrategyServiceInterface
	customRulesService         CustomRulesServiceInterface // For loading user-specific custom rules
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
		providers:           make(map[string]providers.Provider),
		strategy:            &ModelBasedStrategy{}, // Default strategy
		currentStrategyType: StrategyModelBased,    // Default strategy type
		costCalculator:      NewCostCalculator(),
		latencyTracker:      NewLatencyTracker(100),
		providerKeyService:  nil,                  // Will be set if BYOK is enabled
		serverProviderKeys:  ServerProviderKeys{}, // Will be set from config
	}
}

// SetProviderKeyService sets the provider key service for BYOK
func (r *Router) SetProviderKeyService(service ProviderKeyServiceInterface) {
	r.providerKeyService = service
}

// SetRoutingStrategyService sets the routing strategy service (for getting default strategy)
func (r *Router) SetRoutingStrategyService(service RoutingStrategyServiceInterface) {
	r.routingStrategyService = service
}

// SetUserRoutingStrategyService sets the user routing strategy service (for getting user preferences)
func (r *Router) SetUserRoutingStrategyService(service UserRoutingStrategyServiceInterface) {
	r.userRoutingStrategyService = service
}

// SetCustomRulesService sets the custom rules service (for loading user-specific custom rules)
func (r *Router) SetCustomRulesService(service CustomRulesServiceInterface) {
	r.customRulesService = service
}

// SetServerProviderKeys sets server-level provider keys (fallback)
func (r *Router) SetServerProviderKeys(keys ServerProviderKeys) {
	r.serverProviderKeys = keys
}

// SetStrategy sets the routing strategy
func (r *Router) SetStrategy(strategy RoutingStrategy) {
	r.strategy = strategy
}

// SetCustomStrategy sets a custom routing strategy with rules
func (r *Router) SetCustomStrategy(customStrategy *CustomStrategy) {
	r.strategy = customStrategy
	r.currentStrategyType = StrategyCustom
}

// SetStrategyType sets the routing strategy by type
func (r *Router) SetStrategyType(strategyType StrategyType) {
	var newStrategy RoutingStrategy
	switch strategyType {
	case StrategyCostBased:
		newStrategy = NewCostBasedStrategy(r.costCalculator)
	case StrategyLatencyBased:
		newStrategy = NewLatencyBasedStrategy(r.latencyTracker)
	case StrategyLoadBalanced:
		newStrategy = NewLoadBalancedStrategy()
	case StrategyModelBased:
		newStrategy = &ModelBasedStrategy{}
	case StrategyCustom:
		// Custom strategy with empty rules falls back to model-based
		// Admin can configure custom rules via API in the future
		newStrategy = NewCustomStrategy(nil) // nil rules = fallback to model-based
	default:
		newStrategy = &ModelBasedStrategy{}
		strategyType = StrategyModelBased
	}

	// Set both the strategy and the type
	r.strategy = newStrategy
	r.currentStrategyType = strategyType
}

// GetStrategyType returns the current routing strategy type (default)
func (r *Router) GetStrategyType() StrategyType {
	// Use explicit field first (more reliable)
	if r.currentStrategyType != "" {
		return r.currentStrategyType
	}

	// Fallback to type assertion if field is not set
	if r.strategy == nil {
		return StrategyModelBased
	}

	switch r.strategy.(type) {
	case *CostBasedStrategy:
		return StrategyCostBased
	case *LatencyBasedStrategy:
		return StrategyLatencyBased
	case *LoadBalancedStrategy:
		return StrategyLoadBalanced
	case *ModelBasedStrategy:
		return StrategyModelBased
	case *CustomStrategy:
		return StrategyCustom
	default:
		return StrategyModelBased
	}
}

// GetStrategyForUser returns the routing strategy for a specific user
// Checks: user preference → default → fallback
func (r *Router) GetStrategyForUser(ctx context.Context, userID *uuid.UUID) StrategyType {
	// 1. Check if strategy is locked (admin override)
	if r.routingStrategyService != nil {
		locked, err := r.routingStrategyService.IsRoutingStrategyLocked(ctx)
		if err == nil && locked {
			// Strategy is locked, use default
			return r.GetStrategyType()
		}
	}

	// 2. If user ID provided, check user preference
	if userID != nil && r.userRoutingStrategyService != nil {
		userStrategy, err := r.userRoutingStrategyService.GetUserRoutingStrategy(ctx, *userID)
		if err == nil && userStrategy != "" {
			// User has a preference, use it
			strategyType := StrategyType(userStrategy)
			// Validate strategy type
			switch strategyType {
			case StrategyModelBased, StrategyCostBased, StrategyLatencyBased, StrategyLoadBalanced, StrategyCustom:
				return strategyType
			}
		}
	}

	// 3. Fall back to default
	return r.GetStrategyType()
}

// GetStrategyInstanceForUser returns the routing strategy instance for a specific user
// This is used by Route() to get the actual strategy object
func (r *Router) GetStrategyInstanceForUser(ctx context.Context, userID *uuid.UUID) RoutingStrategy {
	strategyType := r.GetStrategyForUser(ctx, userID)

	// Create strategy instance based on type
	switch strategyType {
	case StrategyCostBased:
		return NewCostBasedStrategy(r.costCalculator)
	case StrategyLatencyBased:
		return NewLatencyBasedStrategy(r.latencyTracker)
	case StrategyLoadBalanced:
		return NewLoadBalancedStrategy()
	case StrategyCustom:
		// Custom strategy - load user-specific rules if available
		if r.customRulesService != nil && userID != nil {
			customRules, err := r.customRulesService.GetActiveRulesForUser(ctx, userID)
			if err == nil && len(customRules) > 0 {
				// Convert to routing rules with condition functions
				// We need to build conditions, so we'll create a temporary adapter
				// For now, we'll use a simple approach: create routing rules directly
				routingRules := make([]RoutingRule, 0, len(customRules))
				for _, rule := range customRules {
					condition := r.buildCustomRuleCondition(rule)
					routingRules = append(routingRules, RoutingRule{
						Provider:  rule.ProviderName,
						Priority:  rule.Priority,
						Condition: condition,
					})
				}
				return NewCustomStrategy(routingRules)
			}
		}
		// Fallback: use global custom rules if set, otherwise model-based
		if r.strategy != nil {
			if customStrategy, ok := r.strategy.(*CustomStrategy); ok && customStrategy != nil {
				return customStrategy
			}
		}
		// No rules available, fallback to model-based
		return NewCustomStrategy(nil) // nil rules = fallback to model-based
	case StrategyModelBased:
		fallthrough
	default:
		return &ModelBasedStrategy{}
	}
}

// buildCustomRuleCondition creates a condition function from a CustomRule
func (r *Router) buildCustomRuleCondition(rule CustomRule) func(req providers.ChatRequest) bool {
	return func(req providers.ChatRequest) bool {
		switch rule.ConditionType {
		case "model":
			if model, ok := rule.ConditionValue["model"].(string); ok {
				return req.Model == model
			}
		case "cost_threshold":
			// Check if estimated cost is below threshold
			if maxCost, ok := rule.ConditionValue["max_cost"].(float64); ok {
				// Estimate cost for the request
				estimatedCost := r.costCalculator.EstimateCost(rule.ProviderName, req.Model, req.Messages)
				return estimatedCost <= maxCost
			}
			return false
		case "latency_threshold":
			// Check if average latency is below threshold
			if maxLatencyMs, ok := rule.ConditionValue["max_latency_ms"].(float64); ok {
				avgLatency := r.latencyTracker.GetAverageLatency(rule.ProviderName)
				return avgLatency.Milliseconds() <= int64(maxLatencyMs)
			}
			return false
		}
		return false
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

	// Use routing strategy to select provider (user-specific if available)
	strategy := r.GetStrategyInstanceForUser(ctx, userID)
	selectedProvider, err := strategy.SelectProvider(ctx, req, availableProviders)
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
			// Add routing metadata to response
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
