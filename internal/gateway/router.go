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

type RoutingStrategyServiceInterface interface {
	GetDefaultRoutingStrategy(ctx context.Context) (string, error)
	IsRoutingStrategyLocked(ctx context.Context) (bool, error)
}

type UserRoutingStrategyServiceInterface interface {
	GetUserRoutingStrategy(ctx context.Context, userID uuid.UUID) (string, error)
}

type CustomRulesServiceInterface interface {
	GetActiveRulesForUser(ctx context.Context, userID *uuid.UUID) ([]CustomRule, error)
}

type CustomRule struct {
	ConditionType  string
	ConditionValue map[string]interface{}
	ProviderName   string
	Priority       int
}

type Router struct {
	providers                  map[string]providers.Provider
	defaultProvider            providers.Provider
	strategy                   RoutingStrategy
	currentStrategyType        StrategyType
	costCalculator             *CostCalculator
	latencyTracker             *LatencyTracker
	providerKeyService         ProviderKeyServiceInterface
	serverProviderKeys         ServerProviderKeys
	routingStrategyService     RoutingStrategyServiceInterface
	userRoutingStrategyService UserRoutingStrategyServiceInterface
	customRulesService         CustomRulesServiceInterface
}

type ProviderKeyServiceInterface interface {
	GetProviderKey(ctx context.Context, userID uuid.UUID, provider string) (string, error)
}

type ServerProviderKeys struct {
	OpenAI    string
	Anthropic string
	Google    string
}

func NewRouter() *Router {
	return &Router{
		providers:           make(map[string]providers.Provider),
		strategy:            &ModelBasedStrategy{},
		currentStrategyType: StrategyModelBased,
		costCalculator:      NewCostCalculator(),
		latencyTracker:      NewLatencyTracker(100),
		providerKeyService:  nil,
		serverProviderKeys:  ServerProviderKeys{},
	}
}

func (r *Router) SetProviderKeyService(service ProviderKeyServiceInterface) {
	r.providerKeyService = service
}

func (r *Router) SetRoutingStrategyService(service RoutingStrategyServiceInterface) {
	r.routingStrategyService = service
}

func (r *Router) SetUserRoutingStrategyService(service UserRoutingStrategyServiceInterface) {
	r.userRoutingStrategyService = service
}

func (r *Router) SetCustomRulesService(service CustomRulesServiceInterface) {
	r.customRulesService = service
}

func (r *Router) SetServerProviderKeys(keys ServerProviderKeys) {
	r.serverProviderKeys = keys
}

func (r *Router) SetStrategy(strategy RoutingStrategy) {
	r.strategy = strategy
}

func (r *Router) SetCustomStrategy(customStrategy *CustomStrategy) {
	r.strategy = customStrategy
	r.currentStrategyType = StrategyCustom
}

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
		newStrategy = NewCustomStrategy(nil)
	default:
		newStrategy = &ModelBasedStrategy{}
		strategyType = StrategyModelBased
	}

	r.strategy = newStrategy
	r.currentStrategyType = strategyType
}

func (r *Router) GetStrategyType() StrategyType {
	if r.currentStrategyType != "" {
		return r.currentStrategyType
	}
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

func (r *Router) GetStrategyForUser(ctx context.Context, userID *uuid.UUID) StrategyType {
	if r.routingStrategyService != nil {
		locked, err := r.routingStrategyService.IsRoutingStrategyLocked(ctx)
		if err == nil && locked {
			return r.GetStrategyType()
		}
	}
	if userID != nil && r.userRoutingStrategyService != nil {
		userStrategy, err := r.userRoutingStrategyService.GetUserRoutingStrategy(ctx, *userID)
		if err == nil && userStrategy != "" {
			strategyType := StrategyType(userStrategy)
			switch strategyType {
			case StrategyModelBased, StrategyCostBased, StrategyLatencyBased, StrategyLoadBalanced, StrategyCustom:
				return strategyType
			}
		}
	}
	return r.GetStrategyType()
}

func (r *Router) GetStrategyInstanceForUser(ctx context.Context, userID *uuid.UUID) RoutingStrategy {
	strategyType := r.GetStrategyForUser(ctx, userID)
	switch strategyType {
	case StrategyCostBased:
		return NewCostBasedStrategy(r.costCalculator)
	case StrategyLatencyBased:
		return NewLatencyBasedStrategy(r.latencyTracker)
	case StrategyLoadBalanced:
		return NewLoadBalancedStrategy()
	case StrategyCustom:
		if r.customRulesService != nil && userID != nil {
			customRules, err := r.customRulesService.GetActiveRulesForUser(ctx, userID)
			if err == nil && len(customRules) > 0 {
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
		if r.strategy != nil {
			if customStrategy, ok := r.strategy.(*CustomStrategy); ok && customStrategy != nil {
				return customStrategy
			}
		}
		return NewCustomStrategy(nil)
	case StrategyModelBased:
		fallthrough
	default:
		return &ModelBasedStrategy{}
	}
}

func (r *Router) buildCustomRuleCondition(rule CustomRule) func(req providers.ChatRequest) bool {
	return func(req providers.ChatRequest) bool {
		switch rule.ConditionType {
		case "model":
			if model, ok := rule.ConditionValue["model"].(string); ok {
				return req.Model == model
			}
		case "cost_threshold":
			if maxCost, ok := rule.ConditionValue["max_cost"].(float64); ok {
				estimatedCost := r.costCalculator.EstimateCost(rule.ProviderName, req.Model, req.Messages)
				return estimatedCost <= maxCost
			}
			return false
		case "latency_threshold":
			if maxLatencyMs, ok := rule.ConditionValue["max_latency_ms"].(float64); ok {
				avgLatency := r.latencyTracker.GetAverageLatency(rule.ProviderName)
				return avgLatency.Milliseconds() <= int64(maxLatencyMs)
			}
			return false
		}
		return false
	}
}

func (r *Router) RegisterProvider(provider providers.Provider) {
	r.providers[provider.Name()] = provider
	if r.defaultProvider == nil {
		r.defaultProvider = provider
	}
}

func (r *Router) Route(ctx context.Context, req providers.ChatRequest, userID *uuid.UUID) (*providers.ChatResponse, error) {
	if len(r.providers) == 0 {
		return nil, fmt.Errorf("no providers available")
	}
	allProviders := r.getAllProviders()
	strategy := r.GetStrategyInstanceForUser(ctx, userID)
	selectedProvider, err := strategy.SelectProvider(ctx, req, allProviders)
	if err != nil {
		return nil, fmt.Errorf("failed to select provider: %w", err)
	}
	availableProviders := r.getAvailableProviders(ctx, userID)
	if !r.providerInList(availableProviders, selectedProvider.Name()) {
		availableProviders = []providers.Provider{selectedProvider}
	}
	providersToTry := []providers.Provider{selectedProvider}
	for _, provider := range availableProviders {
		if provider.Name() != selectedProvider.Name() {
			providersToTry = append(providersToTry, provider)
		}
	}
	var lastErr error
	for _, provider := range providersToTry {
		start := time.Now()
		resp, err := provider.Chat(ctx, req)
		latency := time.Since(start)
		r.latencyTracker.RecordLatency(provider.Name(), latency)
		if err == nil {
			resp.Provider = provider.Name()
			resp.LatencyMs = latency.Milliseconds()
			if resp.Usage.TotalTokens > 0 {
				resp.Cost = r.costCalculator.CalculateActualCost(provider.Name(), resp.Model, resp.Usage)
			}
			return resp, nil
		}
		lastErr = err
	}
	if lastErr != nil {
		return nil, fmt.Errorf("all providers failed, last error: %w", lastErr)
	}

	return nil, fmt.Errorf("no providers available")
}

func (r *Router) RouteStream(ctx context.Context, req providers.ChatRequest, userID *uuid.UUID) (<-chan providers.StreamChunk, <-chan error) {
	chunkChan := make(chan providers.StreamChunk, 10)
	errChan := make(chan error, 1)

	go func() {
		defer close(chunkChan)
		defer close(errChan)

		if len(r.providers) == 0 {
			errChan <- fmt.Errorf("no providers available")
			return
		}
		allProviders := r.getAllProviders()
		strategy := r.GetStrategyInstanceForUser(ctx, userID)
		selectedProvider, err := strategy.SelectProvider(ctx, req, allProviders)
		if err != nil {
			errChan <- fmt.Errorf("failed to select provider: %w", err)
			return
		}
		availableProviders := r.getAvailableProviders(ctx, userID)
		if !r.providerInList(availableProviders, selectedProvider.Name()) {
			availableProviders = []providers.Provider{selectedProvider}
		}
		_, ok := selectedProvider.(providers.StreamingProvider)
		if !ok {
			resp, err := selectedProvider.Chat(ctx, req)
			if err != nil {
				errChan <- err
				return
			}
			if len(resp.Choices) > 0 {
				content := ""
				switch c := resp.Choices[0].Message.Content.(type) {
				case string:
					content = c
				default:
					content = fmt.Sprintf("%v", c)
				}

				chunkChan <- providers.StreamChunk{
					ID:      resp.ID,
					Content: content,
					Done:    true,
					Usage:   &resp.Usage,
				}
			}
			return
		}

		providersToTry := []providers.Provider{selectedProvider}
		modelLower := strings.ToLower(req.Model)
		ollamaStyle := strings.Contains(modelLower, ":")
		if !ollamaStyle {
			for _, provider := range availableProviders {
				if provider.Name() != selectedProvider.Name() {
					providersToTry = append(providersToTry, provider)
				}
			}
		}

		var lastErr error
		for _, provider := range providersToTry {
			streamingProvider, ok := provider.(providers.StreamingProvider)
			if !ok {
				continue
			}
			streamChunks, streamErrs := streamingProvider.ChatStream(ctx, req)
			sentAnyChunk := false
		readLoop:
			for {
				select {
				case chunk, ok := <-streamChunks:
					if !ok {
						break readLoop
					}
					chunk.Provider = provider.Name()
					chunkChan <- chunk
					sentAnyChunk = true
					if chunk.Done {
						return
					}
				case err, ok := <-streamErrs:
					if !ok {
						break readLoop
					}
					lastErr = err
					break readLoop
				case <-ctx.Done():
					errChan <- ctx.Err()
					return
				}
			}

			if sentAnyChunk {
				chunkChan <- providers.StreamChunk{Content: "", Done: true, Provider: provider.Name()}
				return
			}
			if lastErr == nil {
				return
			}
		}
		if lastErr != nil {
			errChan <- fmt.Errorf("all providers failed, last error: %w", lastErr)
		} else {
			errChan <- fmt.Errorf("no streaming providers available")
		}
	}()

	return chunkChan, errChan
}

func (r *Router) getAllProviders() []providers.Provider {
	out := make([]providers.Provider, 0, len(r.providers))
	for _, p := range r.providers {
		out = append(out, p)
	}
	return out
}

func (r *Router) getAvailableProviders(ctx context.Context, userID *uuid.UUID) []providers.Provider {
	available := make([]providers.Provider, 0)
	if userID != nil && r.providerKeyService != nil {
		userProviders := r.getUserProviders(ctx, *userID)
		for _, provider := range userProviders {
			if err := provider.HealthCheck(ctx); err == nil {
				available = append(available, provider)
			}
		}
		if len(available) > 0 {
			return available
		}
	}
	for _, provider := range r.providers {
		if err := provider.HealthCheck(ctx); err == nil {
			available = append(available, provider)
		}
	}
	return available
}

func (r *Router) providerInList(list []providers.Provider, name string) bool {
	for _, p := range list {
		if p.Name() == name {
			return true
		}
	}
	return false
}

func (r *Router) getUserProviders(ctx context.Context, userID uuid.UUID) []providers.Provider {
	userProviders := make([]providers.Provider, 0)
	providersToCheck := []string{"openai", "anthropic", "google"}
	for _, providerName := range providersToCheck {
		apiKey, err := r.providerKeyService.GetProviderKey(ctx, userID, providerName)
		if err != nil || apiKey == "" {
			continue
		}
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

func (r *Router) selectProvider(model string) providers.Provider {
	modelLower := strings.ToLower(model)
	for _, provider := range r.providers {
		models := provider.GetModels()
		for _, providerModel := range models {
			if strings.ToLower(providerModel) == modelLower ||
				strings.Contains(modelLower, strings.ToLower(provider.Name())) {
				return provider
			}
		}
	}
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
	return r.defaultProvider
}

func (r *Router) GetProvider(name string) (providers.Provider, error) {
	provider, exists := r.providers[name]
	if !exists {
		return nil, errors.ErrProviderNotFound
	}
	return provider, nil
}

func (r *Router) ListProviders() []string {
	names := make([]string, 0, len(r.providers))
	for name := range r.providers {
		names = append(names, name)
	}
	return names
}

// ListProviderDetailsForUser returns provider details (name, healthy, models) for both
// server-registered providers and any BYOK providers the user has a key for.
func (r *Router) ListProviderDetailsForUser(ctx context.Context, userID *uuid.UUID) []map[string]interface{} {
	seen := make(map[string]bool)
	out := make([]map[string]interface{}, 0)

	add := func(p providers.Provider) {
		name := p.Name()
		if seen[name] {
			return
		}
		seen[name] = true
		healthy := p.HealthCheck(ctx) == nil
		out = append(out, map[string]interface{}{
			"name":    name,
			"healthy": healthy,
			"models":  p.GetModels(),
		})
	}

	for _, p := range r.providers {
		add(p)
	}
	if userID != nil && r.providerKeyService != nil {
		for _, p := range r.getUserProviders(ctx, *userID) {
			add(p)
		}
	}
	return out
}

func (r *Router) GetCostCalculator() *CostCalculator {
	return r.costCalculator
}

func (r *Router) GetLatencyTracker() *LatencyTracker {
	return r.latencyTracker
}
