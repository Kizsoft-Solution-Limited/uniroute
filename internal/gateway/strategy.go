package gateway

import (
	"context"
	"strings"
	"time"

	"github.com/Kizsoft-Solution-Limited/uniroute/internal/providers"
)

type RoutingStrategy interface {
	SelectProvider(ctx context.Context, req providers.ChatRequest, availableProviders []providers.Provider) (providers.Provider, error)
}

type StrategyType string

const (
	StrategyModelBased   StrategyType = "model"
	StrategyCostBased    StrategyType = "cost"
	StrategyLatencyBased StrategyType = "latency"
	StrategyLoadBalanced StrategyType = "balanced"
	StrategyCustom       StrategyType = "custom"
)

type ModelBasedStrategy struct{}

func (s *ModelBasedStrategy) SelectProvider(ctx context.Context, req providers.ChatRequest, availableProviders []providers.Provider) (providers.Provider, error) {
	if len(availableProviders) == 0 {
		return nil, ErrNoProviders
	}

	modelLower := strings.ToLower(req.Model)
	for _, provider := range availableProviders {
		models := provider.GetModels()
		for _, model := range models {
			if strings.ToLower(model) == modelLower {
				return provider, nil
			}
		}
	}
	for _, provider := range availableProviders {
		models := provider.GetModels()
		for _, model := range models {
			providerModelLower := strings.ToLower(model)
			if strings.Contains(providerModelLower, modelLower) || strings.Contains(modelLower, providerModelLower) {
				return provider, nil
			}
		}
	}
	if strings.Contains(modelLower, "llama") || strings.Contains(modelLower, "mistral") || 
	   strings.Contains(modelLower, "phi") || strings.Contains(modelLower, "codellama") ||
	   strings.Contains(modelLower, "neural") || strings.Contains(modelLower, "orca") {
		for _, provider := range availableProviders {
			if provider.Name() == "local" {
				return provider, nil
			}
		}
	}
	return availableProviders[0], nil
}

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
		cost := s.costCalculator.EstimateCost(provider.Name(), req.Model, req.Messages)
		if cost < lowestCost {
			lowestCost = cost
			cheapestProvider = provider
		}
	}

	if cheapestProvider == nil {
		return availableProviders[0], nil
	}

	return cheapestProvider, nil
}

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
		latency := s.latencyTracker.GetAverageLatency(provider.Name())
		if latency < lowestLatency {
			lowestLatency = latency
			fastestProvider = provider
		}
	}

	if fastestProvider == nil {
		return availableProviders[0], nil
	}

	return fastestProvider, nil
}

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
		supportedProviders = availableProviders
	}
	selected := supportedProviders[s.counter%len(supportedProviders)]
	s.counter++

	return selected, nil
}

type CustomStrategy struct {
	rules []RoutingRule
}

type RoutingRule struct {
	Condition func(req providers.ChatRequest) bool
	Provider  string
	Priority  int
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

	if s.rules == nil || len(s.rules) == 0 {
		strategy := &ModelBasedStrategy{}
		return strategy.SelectProvider(ctx, req, availableProviders)
	}
	for _, rule := range s.rules {
		if rule.Condition != nil && rule.Condition(req) {
			for _, provider := range availableProviders {
				if provider.Name() == rule.Provider {
					return provider, nil
				}
			}
		}
	}
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
