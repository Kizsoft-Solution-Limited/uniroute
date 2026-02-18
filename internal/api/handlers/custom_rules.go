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

type CustomRulesHandler struct {
	router         *gateway.Router
	ruleRepo       *storage.CustomRoutingRulesRepository
	costCalculator *gateway.CostCalculator
	latencyTracker *gateway.LatencyTracker
	logger         zerolog.Logger
}

func NewCustomRulesHandler(router *gateway.Router, ruleRepo *storage.CustomRoutingRulesRepository, logger zerolog.Logger) *CustomRulesHandler {
	return &CustomRulesHandler{
		router:         router,
		ruleRepo:       ruleRepo,
		costCalculator: router.GetCostCalculator(),
		latencyTracker: router.GetLatencyTracker(),
		logger:         logger,
	}
}

func (h *CustomRulesHandler) GetCustomRules(c *gin.Context) {
	ctx := c.Request.Context()

	isAdminRequest := c.FullPath() != "" && len(c.FullPath()) >= 6 && c.FullPath()[:6] == "/admin"

	var userID *uuid.UUID
	if !isAdminRequest {
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

	var rules []*storage.CustomRoutingRule
	var err error

	if isAdminRequest {
		rules, err = h.ruleRepo.GetActiveRules(ctx)
	} else if userID != nil {
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

	rulesMaps := make([]map[string]interface{}, len(req.Rules))
	for i, rule := range req.Rules {
		rulesMaps[i] = rule.ToInterface()
	}

	isAdminRequest := c.FullPath() != "" && c.FullPath()[:6] == "/admin"

	var saveUserID *uuid.UUID
	if !isAdminRequest {
		saveUserID = userID
	} else {
		saveUserID = nil
	}

	if err := h.ruleRepo.SaveRulesForUser(ctx, rulesMaps, saveUserID, updatedBy); err != nil {
		h.logger.Error().Err(err).Msg("Failed to save custom routing rules")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to save custom routing rules",
		})
		return
	}

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

type CustomRuleRequest struct {
	Name          string                 `json:"name" binding:"required"`
	ConditionType string                 `json:"condition_type" binding:"required"` // 'model', 'cost_threshold', 'latency_threshold'
	ConditionValue map[string]interface{} `json:"condition_value" binding:"required"`
	ProviderName  string                 `json:"provider_name" binding:"required"`
	Priority      int                    `json:"priority"`
	Enabled       bool                   `json:"enabled"`
	Description   string                 `json:"description"`
}

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

func (h *CustomRulesHandler) reloadRouterRules(ctx context.Context) error {
	rules, err := h.ruleRepo.GetActiveRules(ctx)
	if err != nil {
		return err
	}

	routingRules := make([]gateway.RoutingRule, 0, len(rules))
	for _, rule := range rules {
		routingRule := gateway.RoutingRule{
			Provider:  rule.ProviderName,
			Priority:  rule.Priority,
			Condition: h.buildCondition(rule),
		}
		routingRules = append(routingRules, routingRule)
	}

	customStrategy := gateway.NewCustomStrategy(routingRules)
	h.router.SetCustomStrategy(customStrategy)

	return nil
}

func (h *CustomRulesHandler) buildCondition(rule *storage.CustomRoutingRule) func(req providers.ChatRequest) bool {
	return func(req providers.ChatRequest) bool {
		switch rule.ConditionType {
		case "model":
			if model, ok := rule.ConditionValue["model"].(string); ok {
				return req.Model == model
			}
		case "cost_threshold":
			if maxCost, ok := rule.ConditionValue["max_cost"].(float64); ok {
				estimatedCost := h.costCalculator.EstimateCost(rule.ProviderName, req.Model, req.Messages)
				return estimatedCost <= maxCost
			}
			return false
		case "latency_threshold":
			if maxLatencyMs, ok := rule.ConditionValue["max_latency_ms"].(float64); ok {
				avgLatency := h.latencyTracker.GetAverageLatency(rule.ProviderName)
				return avgLatency.Milliseconds() <= int64(maxLatencyMs)
			}
			return false
		}
		return false
	}
}

