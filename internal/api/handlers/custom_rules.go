package handlers

import (
	"context"
	"net/http"

	"github.com/Kizsoft-Solution-Limited/uniroute/internal/gateway"
	"github.com/Kizsoft-Solution-Limited/uniroute/internal/providers"
	"github.com/Kizsoft-Solution-Limited/uniroute/internal/storage"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// CustomRulesHandler handles custom routing rules configuration
type CustomRulesHandler struct {
	router     *gateway.Router
	ruleRepo   *storage.CustomRoutingRulesRepository
	logger     zerolog.Logger
}

// NewCustomRulesHandler creates a new custom rules handler
func NewCustomRulesHandler(router *gateway.Router, ruleRepo *storage.CustomRoutingRulesRepository, logger zerolog.Logger) *CustomRulesHandler {
	return &CustomRulesHandler{
		router:   router,
		ruleRepo: ruleRepo,
		logger:   logger,
	}
}

// GetCustomRules handles GET /admin/routing/custom-rules
func (h *CustomRulesHandler) GetCustomRules(c *gin.Context) {
	ctx := c.Request.Context()

	rules, err := h.ruleRepo.GetActiveRules(ctx)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to get custom routing rules")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve custom routing rules",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"rules": rules,
		"count": len(rules),
	})
}

// SetCustomRules handles POST /admin/routing/custom-rules
func (h *CustomRulesHandler) SetCustomRules(c *gin.Context) {
	var req struct {
		Rules []CustomRuleRequest `json:"rules" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request",
			"details": err.Error(),
		})
		return
	}

	// Get user ID
	userID, exists := c.Get("user_id")
	var updatedBy *uuid.UUID
	if exists {
		if uid, ok := userID.(uuid.UUID); ok {
			updatedBy = &uid
		}
	}

	ctx := c.Request.Context()

	// Convert rules to map slice
	rulesMaps := make([]map[string]interface{}, len(req.Rules))
	for i, rule := range req.Rules {
		rulesMaps[i] = rule.ToInterface()
	}

	// Save rules
	if err := h.ruleRepo.SaveRules(ctx, rulesMaps, updatedBy); err != nil {
		h.logger.Error().Err(err).Msg("Failed to save custom routing rules")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to save custom routing rules",
		})
		return
	}

	// Reload router with new rules
	if err := h.reloadRouterRules(ctx); err != nil {
		h.logger.Warn().Err(err).Msg("Failed to reload router rules, but rules were saved")
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Custom routing rules updated successfully",
		"count":   len(req.Rules),
	})
}

// CustomRuleRequest represents a custom routing rule
type CustomRuleRequest struct {
	Name          string                 `json:"name" binding:"required"`
	ConditionType string                 `json:"condition_type" binding:"required"` // 'model', 'cost_threshold', 'latency_threshold'
	ConditionValue map[string]interface{} `json:"condition_value" binding:"required"`
	ProviderName  string                 `json:"provider_name" binding:"required"`
	Priority      int                    `json:"priority"`
	Enabled       bool                   `json:"enabled"`
	Description   string                 `json:"description"`
}

// ToInterface converts CustomRuleRequest to map[string]interface{}
func (r *CustomRuleRequest) ToInterface() map[string]interface{} {
	return map[string]interface{}{
		"name":           r.Name,
		"condition_type": r.ConditionType,
		"condition_value": r.ConditionValue,
		"provider_name":  r.ProviderName,
		"priority":       r.Priority,
		"enabled":        r.Enabled,
		"description":    r.Description,
	}
}

// reloadRouterRules reloads custom rules into the router
func (h *CustomRulesHandler) reloadRouterRules(ctx context.Context) error {
	rules, err := h.ruleRepo.GetActiveRules(ctx)
	if err != nil {
		return err
	}

	// Convert database rules to gateway routing rules
	routingRules := make([]gateway.RoutingRule, 0, len(rules))
	for _, rule := range rules {
		routingRule := gateway.RoutingRule{
			Provider: rule.ProviderName,
			Priority: rule.Priority,
			Condition: h.buildCondition(rule),
		}
		routingRules = append(routingRules, routingRule)
	}

	// Update router's custom strategy
	customStrategy := gateway.NewCustomStrategy(routingRules)
	h.router.SetCustomStrategy(customStrategy)

	return nil
}

// buildCondition creates a condition function from a rule
func (h *CustomRulesHandler) buildCondition(rule *storage.CustomRoutingRule) func(req providers.ChatRequest) bool {
	return func(req providers.ChatRequest) bool {
		switch rule.ConditionType {
		case "model":
			if model, ok := rule.ConditionValue["model"].(string); ok {
				return req.Model == model
			}
		case "cost_threshold":
			// This would require cost calculation - for now, return false
			// TODO: Implement cost-based condition
			return false
		case "latency_threshold":
			// This would require latency tracking - for now, return false
			// TODO: Implement latency-based condition
			return false
		}
		return false
	}
}

