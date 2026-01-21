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
	router         *gateway.Router
	ruleRepo       *storage.CustomRoutingRulesRepository
	costCalculator *gateway.CostCalculator
	latencyTracker *gateway.LatencyTracker
	logger         zerolog.Logger
}

// NewCustomRulesHandler creates a new custom rules handler
func NewCustomRulesHandler(router *gateway.Router, ruleRepo *storage.CustomRoutingRulesRepository, logger zerolog.Logger) *CustomRulesHandler {
	return &CustomRulesHandler{
		router:         router,
		ruleRepo:       ruleRepo,
		costCalculator: router.GetCostCalculator(),
		latencyTracker: router.GetLatencyTracker(),
		logger:         logger,
	}
}

// GetCustomRules handles GET /admin/routing/custom-rules (admin) or GET /auth/routing/custom-rules (user)
func (h *CustomRulesHandler) GetCustomRules(c *gin.Context) {
	ctx := c.Request.Context()

	// Check if this is an admin request (admin routes) or user request
	isAdminRequest := c.FullPath() != "" && len(c.FullPath()) >= 6 && c.FullPath()[:6] == "/admin"
	
	var userID *uuid.UUID
	if !isAdminRequest {
		// User request: get user-specific rules
		userIDStr, exists := c.Get("user_id")
		if exists {
			if uid, ok := userIDStr.(uuid.UUID); ok {
				userID = &uid
			} else if idStr, ok := userIDStr.(string); ok {
				if uid, err := uuid.Parse(idStr); err == nil {
					userID = &uid
				}
			}
		}
	}
	// Admin request: userID stays nil to get global rules

	var rules []*storage.CustomRoutingRule
	var err error
	
	if isAdminRequest {
		// Get global/admin rules
		rules, err = h.ruleRepo.GetActiveRules(ctx)
	} else if userID != nil {
		// Get user-specific rules
		rules, err = h.ruleRepo.GetActiveRulesForUser(ctx, userID)
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "user not authenticated",
		})
		return
	}
	
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
		"user_specific": !isAdminRequest,
	})
}

// SetCustomRules handles POST /admin/routing/custom-rules (admin) or POST /auth/routing/custom-rules (user)
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

	// Get user ID from context
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "user not authenticated",
		})
		return
	}

	var userID *uuid.UUID
	var updatedBy *uuid.UUID
	
	if uid, ok := userIDStr.(uuid.UUID); ok {
		userID = &uid
		updatedBy = &uid
	} else if idStr, ok := userIDStr.(string); ok {
		if uid, err := uuid.Parse(idStr); err == nil {
			userID = &uid
			updatedBy = &uid
		}
	}

	if userID == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid user ID",
		})
		return
	}

	ctx := c.Request.Context()

	// Convert rules to map slice
	rulesMaps := make([]map[string]interface{}, len(req.Rules))
	for i, rule := range req.Rules {
		rulesMaps[i] = rule.ToInterface()
	}

	// Check if this is an admin request (admin routes should save as global rules)
	// For now, we'll check if the route path contains "/admin/"
	isAdminRequest := c.FullPath() != "" && c.FullPath()[:6] == "/admin"
	
	var saveUserID *uuid.UUID
	if !isAdminRequest {
		// User request: save as user-specific rules
		saveUserID = userID
	} else {
		// Admin request: save as global rules (user_id = NULL)
		saveUserID = nil
	}

	// Save rules (user-specific or global)
	if err := h.ruleRepo.SaveRulesForUser(ctx, rulesMaps, saveUserID, updatedBy); err != nil {
		h.logger.Error().Err(err).Msg("Failed to save custom routing rules")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to save custom routing rules",
		})
		return
	}

	// Reload router with new rules (only for global/admin rules)
	if isAdminRequest {
		if err := h.reloadRouterRules(ctx); err != nil {
			h.logger.Warn().Err(err).Msg("Failed to reload router rules, but rules were saved")
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "Custom routing rules updated successfully",
		"count":        len(req.Rules),
		"user_specific": !isAdminRequest,
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
			// Check if estimated cost is below threshold
			if maxCost, ok := rule.ConditionValue["max_cost"].(float64); ok {
				// Estimate cost for the request
				estimatedCost := h.costCalculator.EstimateCost(rule.ProviderName, req.Model, req.Messages)
				return estimatedCost <= maxCost
			}
			return false
		case "latency_threshold":
			// Check if average latency is below threshold
			if maxLatencyMs, ok := rule.ConditionValue["max_latency_ms"].(float64); ok {
				avgLatency := h.latencyTracker.GetAverageLatency(rule.ProviderName)
				return avgLatency.Milliseconds() <= int64(maxLatencyMs)
			}
			return false
		}
		return false
	}
}

