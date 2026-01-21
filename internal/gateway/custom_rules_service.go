package gateway

import (
	"context"

	"github.com/Kizsoft-Solution-Limited/uniroute/internal/providers"
	"github.com/Kizsoft-Solution-Limited/uniroute/internal/storage"
	"github.com/google/uuid"
)

// CustomRulesServiceAdapter adapts storage.CustomRoutingRulesRepository to CustomRulesServiceInterface
type CustomRulesServiceAdapter struct {
	repo           *storage.CustomRoutingRulesRepository
	costCalculator *CostCalculator
	latencyTracker *LatencyTracker
}

// NewCustomRulesServiceAdapter creates a new adapter
func NewCustomRulesServiceAdapter(repo *storage.CustomRoutingRulesRepository, costCalculator *CostCalculator, latencyTracker *LatencyTracker) *CustomRulesServiceAdapter {
	return &CustomRulesServiceAdapter{
		repo:           repo,
		costCalculator: costCalculator,
		latencyTracker: latencyTracker,
	}
}

// GetActiveRulesForUser implements CustomRulesServiceInterface
func (a *CustomRulesServiceAdapter) GetActiveRulesForUser(ctx context.Context, userID *uuid.UUID) ([]CustomRule, error) {
	rules, err := a.repo.GetActiveRulesForUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Convert storage rules to gateway rules
	gatewayRules := make([]CustomRule, 0, len(rules))
	for _, rule := range rules {
		gatewayRules = append(gatewayRules, CustomRule{
			ConditionType:  rule.ConditionType,
			ConditionValue: rule.ConditionValue,
			ProviderName:   rule.ProviderName,
			Priority:       rule.Priority,
		})
	}

	return gatewayRules, nil
}

// BuildRoutingRules converts CustomRule slice to RoutingRule slice with condition functions
func (a *CustomRulesServiceAdapter) BuildRoutingRules(rules []CustomRule) []RoutingRule {
	routingRules := make([]RoutingRule, 0, len(rules))
	for _, rule := range rules {
		condition := a.buildCondition(rule)
		routingRules = append(routingRules, RoutingRule{
			Provider:  rule.ProviderName,
			Priority:  rule.Priority,
			Condition: condition,
		})
	}
	return routingRules
}

// buildCondition creates a condition function from a rule
func (a *CustomRulesServiceAdapter) buildCondition(rule CustomRule) func(req providers.ChatRequest) bool {
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
				estimatedCost := a.costCalculator.EstimateCost(rule.ProviderName, req.Model, req.Messages)
				return estimatedCost <= maxCost
			}
			return false
		case "latency_threshold":
			// Check if average latency is below threshold
			if maxLatencyMs, ok := rule.ConditionValue["max_latency_ms"].(float64); ok {
				avgLatency := a.latencyTracker.GetAverageLatency(rule.ProviderName)
				return avgLatency.Milliseconds() <= int64(maxLatencyMs)
			}
			return false
		}
		return false
	}
}
