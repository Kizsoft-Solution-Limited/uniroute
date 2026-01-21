package gateway

import (
	"context"
	"strings"
	"time"

	"github.com/Kizsoft-Solution-Limited/uniroute/internal/providers"
)

// RoutingStrategy defines how requests are routed to providers
type RoutingStrategy interface {
	SelectProvider(ctx context.Context, req providers.ChatRequest, availableProviders []providers.Provider) (providers.Provider, error)
}

// StrategyType represents the type of routing strategy
type StrategyType string

const (
	StrategyModelBased   StrategyType = "model"    // Select by model (default)
	StrategyCostBased    StrategyType = "cost"     // Select cheapest provider
	StrategyLatencyBased StrategyType = "latency"  // Select fastest provider
	StrategyLoadBalanced StrategyType = "balanced" // Round-robin load balancing
	StrategyCustom       StrategyType = "custom"   // Custom rules
)

// ModelBasedStrategy selects provider based on model name
type ModelBasedStrategy struct{}

func (s *ModelBasedStrategy) SelectProvider(ctx context.Context, req providers.ChatRequest, availableProviders []providers.Provider) (providers.Provider, error) {
	if len(availableProviders) == 0 {
		return nil, ErrNoProviders
	}

	modelLower := strings.ToLower(req.Model)

	// Find provider that supports the requested model
	// First try exact match
	for _, provider := range availableProviders {
		models := provider.GetModels()
		for _, model := range models {
			if strings.ToLower(model) == modelLower {
				return provider, nil
			}
		}
	}

	// Try partial match (for cases like "llama2" matching "llama2:7b" or "mistral" matching "mistral:latest")
	for _, provider := range availableProviders {
		models := provider.GetModels()
		for _, model := range models {
			providerModelLower := strings.ToLower(model)
			// Check if requested model name is contained in provider model name or vice versa
			// This handles cases like "llama2" matching "llama2:7b" or "mistral" matching "mistral:latest"
			if strings.Contains(providerModelLower, modelLower) || strings.Contains(modelLower, providerModelLower) {
				return provider, nil
			}
		}
	}

	// Special handling for local models: if model name suggests local (llama, mistral, etc.), prefer local provider
	if strings.Contains(modelLower, "llama") || strings.Contains(modelLower, "mistral") || 
	   strings.Contains(modelLower, "phi") || strings.Contains(modelLower, "codellama") ||
	   strings.Contains(modelLower, "neural") || strings.Contains(modelLower, "orca") {
		for _, provider := range availableProviders {
			if provider.Name() == "local" {
				return provider, nil
			}
		}
	}

	// Fallback to first available provider
	return availableProviders[0], nil
}

// CostBasedStrategy selects the cheapest provider for the model
type CostBasedStrategy struct {
	costCalculator *CostCalculator
}

func NewCostBasedStrategy(calculator *CostCalculator) *CostBasedStrategy {
	return &CostBasedStrategy{
		costCalculator: calculator,
	}
}

func (s *CostBasedStrategy) SelectProvider(ctx context.Context, req providers.ChatRequest, availableProviders []providers.Provider) (providers.Provider, error) {
	if len(availableProviders) == 0 {
		return nil, ErrNoProviders
	}

	var cheapestProvider providers.Provider
	lowestCost := 999999.0

	for _, provider := range availableProviders {
		// Check if provider supports the model
		supportsModel := false
		for _, model := range provider.GetModels() {
			if model == req.Model {
				supportsModel = true
				break
			}
		}
		if !supportsModel {
			continue
		}

		// Estimate cost for this request
		cost := s.costCalculator.EstimateCost(provider.Name(), req.Model, req.Messages)
		if cost < lowestCost {
			lowestCost = cost
			cheapestProvider = provider
		}
	}

	if cheapestProvider == nil {
		// Fallback to first provider
		return availableProviders[0], nil
	}

	return cheapestProvider, nil
}

// LatencyBasedStrategy selects the provider with lowest latency
type LatencyBasedStrategy struct {
	latencyTracker *LatencyTracker
}

func NewLatencyBasedStrategy(tracker *LatencyTracker) *LatencyBasedStrategy {
	return &LatencyBasedStrategy{
		latencyTracker: tracker,
	}
}

func (s *LatencyBasedStrategy) SelectProvider(ctx context.Context, req providers.ChatRequest, availableProviders []providers.Provider) (providers.Provider, error) {
	if len(availableProviders) == 0 {
		return nil, ErrNoProviders
	}

	var fastestProvider providers.Provider
	lowestLatency := time.Duration(999999) * time.Second

	for _, provider := range availableProviders {
		// Check if provider supports the model
		supportsModel := false
		for _, model := range provider.GetModels() {
			if model == req.Model {
				supportsModel = true
				break
			}
		}
		if !supportsModel {
			continue
		}

		// Get average latency for this provider
		latency := s.latencyTracker.GetAverageLatency(provider.Name())
		if latency < lowestLatency {
			lowestLatency = latency
			fastestProvider = provider
		}
	}

	if fastestProvider == nil {
		// Fallback to first provider
		return availableProviders[0], nil
	}

	return fastestProvider, nil
}

// LoadBalancedStrategy uses round-robin load balancing
type LoadBalancedStrategy struct {
	counter int
}

func NewLoadBalancedStrategy() *LoadBalancedStrategy {
	return &LoadBalancedStrategy{
		counter: 0,
	}
}

func (s *LoadBalancedStrategy) SelectProvider(ctx context.Context, req providers.ChatRequest, availableProviders []providers.Provider) (providers.Provider, error) {
	if len(availableProviders) == 0 {
		return nil, ErrNoProviders
	}

	// Filter providers that support the model
	supportedProviders := make([]providers.Provider, 0)
	for _, provider := range availableProviders {
		for _, model := range provider.GetModels() {
			if model == req.Model {
				supportedProviders = append(supportedProviders, provider)
				break
			}
		}
	}

	if len(supportedProviders) == 0 {
		// No provider supports the model, use all providers
		supportedProviders = availableProviders
	}

	// Round-robin selection
	selected := supportedProviders[s.counter%len(supportedProviders)]
	s.counter++

	return selected, nil
}

// CustomStrategy uses custom routing rules
type CustomStrategy struct {
	rules []RoutingRule
}

type RoutingRule struct {
	Condition func(req providers.ChatRequest) bool
	Provider  string
	Priority  int // Higher priority = checked first
}

func NewCustomStrategy(rules []RoutingRule) *CustomStrategy {
	return &CustomStrategy{
		rules: rules,
	}
}

func (s *CustomStrategy) SelectProvider(ctx context.Context, req providers.ChatRequest, availableProviders []providers.Provider) (providers.Provider, error) {
	if len(availableProviders) == 0 {
		return nil, ErrNoProviders
	}

	// If no custom rules defined, fallback to model-based selection
	if s.rules == nil || len(s.rules) == 0 {
		strategy := &ModelBasedStrategy{}
		return strategy.SelectProvider(ctx, req, availableProviders)
	}

	// Sort rules by priority (highest first)
	// For now, we'll check in order (assuming rules are pre-sorted)

	// Check custom rules first
	for _, rule := range s.rules {
		if rule.Condition != nil && rule.Condition(req) {
			// Find provider by name
			for _, provider := range availableProviders {
				if provider.Name() == rule.Provider {
					return provider, nil
				}
			}
		}
	}

	// Fallback to model-based selection if no rules match
	strategy := &ModelBasedStrategy{}
	return strategy.SelectProvider(ctx, req, availableProviders)
}

var ErrNoProviders = &RouterError{Message: "no providers available"}

type RouterError struct {
	Message string
}

func (e *RouterError) Error() string {
	return e.Message
}
