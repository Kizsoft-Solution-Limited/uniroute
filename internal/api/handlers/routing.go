package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/Kizsoft-Solution-Limited/uniroute/internal/gateway"
	"github.com/Kizsoft-Solution-Limited/uniroute/internal/providers"
	"github.com/Kizsoft-Solution-Limited/uniroute/internal/storage"
	"github.com/Kizsoft-Solution-Limited/uniroute/pkg/errors"
	"github.com/rs/zerolog"
)

// RoutingHandler handles routing configuration requests
type RoutingHandler struct {
	Router              *gateway.Router
	SettingsRepository  *storage.SystemSettingsRepository
	UserRepository      *storage.UserRepository
	Logger              zerolog.Logger
}

// NewRoutingHandler creates a new routing handler
func NewRoutingHandler(router *gateway.Router, settingsRepo *storage.SystemSettingsRepository, userRepo *storage.UserRepository, logger zerolog.Logger) *RoutingHandler {
	return &RoutingHandler{
		Router:             router,
		SettingsRepository: settingsRepo,
		UserRepository:     userRepo,
		Logger:             logger,
	}
}

// SetRoutingStrategyRequest represents a request to set routing strategy
type SetRoutingStrategyRequest struct {
	Strategy string `json:"strategy" binding:"required"` // "model", "cost", "latency", "balanced", "custom"
}

// SetRoutingStrategy handles POST /admin/routing/strategy
func (h *RoutingHandler) SetRoutingStrategy(c *gin.Context) {
	var req SetRoutingStrategyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": errors.ErrInvalidRequest.Error(),
			"details": err.Error(),
		})
		return
	}

	// Validate strategy
	validStrategies := map[string]bool{
		"model":    true,
		"cost":     true,
		"latency":  true,
		"balanced": true,
		"custom":   true,
	}
	if !validStrategies[req.Strategy] {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid strategy. Must be one of: model, cost, latency, balanced, custom",
		})
		return
	}

	// Get user ID from context (set by JWT middleware)
	userID, exists := c.Get("user_id")
	var updatedBy *uuid.UUID
	if exists {
		if uid, ok := userID.(uuid.UUID); ok {
			updatedBy = &uid
		}
	}

	// Save to database if repository is available (set as DEFAULT strategy)
	if h.SettingsRepository != nil {
		ctx := c.Request.Context()
		if err := h.SettingsRepository.SetDefaultRoutingStrategy(ctx, req.Strategy, updatedBy); err != nil {
			h.Logger.Error().
				Err(err).
				Str("strategy", req.Strategy).
				Msg("Failed to save default routing strategy to database")
			// Continue anyway - update in-memory router
		} else {
			h.Logger.Info().
				Str("strategy", req.Strategy).
				Msg("Saved default routing strategy to database")
		}
	}

	// Set strategy in router (in-memory default)
	strategyType := gateway.StrategyType(req.Strategy)
	h.Router.SetStrategyType(strategyType)

	// Verify strategy was set correctly
	actualStrategy := h.Router.GetStrategyType()
	h.Logger.Info().
		Str("requested_strategy", req.Strategy).
		Str("actual_strategy", string(actualStrategy)).
		Msg("Routing strategy updated")

	c.JSON(http.StatusOK, gin.H{
		"message": "routing strategy updated",
		"strategy": string(actualStrategy), // Return actual strategy, not requested
	})
}

// GetRoutingStrategy handles GET /admin/routing/strategy
func (h *RoutingHandler) GetRoutingStrategy(c *gin.Context) {
	ctx := c.Request.Context()
	var currentStrategy string

	// Try to get from database first (if repository is available)
	if h.SettingsRepository != nil {
		dbStrategy, err := h.SettingsRepository.GetDefaultRoutingStrategy(ctx)
		if err == nil && dbStrategy != "" {
			currentStrategy = dbStrategy
			// Sync router with database value
			strategyType := gateway.StrategyType(currentStrategy)
			h.Router.SetStrategyType(strategyType)
		}
	}

	// Fallback to router's in-memory value
	if currentStrategy == "" {
		currentStrategy = string(h.Router.GetStrategyType())
	}

	// Get lock status - always include in response
	var isLocked bool = false // Default to unlocked
	if h.SettingsRepository != nil {
		locked, err := h.SettingsRepository.IsRoutingStrategyLocked(ctx)
		if err != nil {
			// Keep default false if error
		} else {
			isLocked = locked
		}
	}
	
	c.JSON(http.StatusOK, gin.H{
		"strategy":            currentStrategy,
		"is_locked":           isLocked, // Always include this field
		"available_strategies": []string{
			"model",
			"cost",
			"latency",
			"balanced",
			"custom",
		},
	})
}

// GetCostEstimate handles POST /v1/routing/estimate-cost
func (h *RoutingHandler) GetCostEstimate(c *gin.Context) {
	var req struct {
		Model    string                  `json:"model" binding:"required"`
		Messages []providers.Message `json:"messages" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": errors.ErrInvalidRequest.Error(),
		})
		return
	}

	// Get cost estimates for all providers
	estimates := make(map[string]float64)
	calculator := gateway.NewCostCalculator()

	// Estimate for each provider that supports the model
	providers := h.Router.ListProviders()
	for _, providerName := range providers {
		provider, err := h.Router.GetProvider(providerName)
		if err != nil {
			continue
		}

		// Check if provider supports model
		supports := false
		for _, model := range provider.GetModels() {
			if model == req.Model {
				supports = true
				break
			}
		}
		if !supports {
			continue
		}

		cost := calculator.EstimateCost(providerName, req.Model, req.Messages)
		estimates[providerName] = cost
	}

	c.JSON(http.StatusOK, gin.H{
		"model":    req.Model,
		"estimates": estimates,
	})
}

// GetLatencyStats handles GET /v1/routing/latency
func (h *RoutingHandler) GetLatencyStats(c *gin.Context) {
	// Get latency stats for all providers
	stats := make(map[string]interface{})
	
	tracker := h.Router.GetLatencyTracker()
	providers := h.Router.ListProviders()
	for _, providerName := range providers {
		avg, min, max, count := tracker.GetLatencyStats(providerName)
		stats[providerName] = gin.H{
			"average_ms": avg.Milliseconds(),
			"min_ms":     min.Milliseconds(),
			"max_ms":     max.Milliseconds(),
			"samples":    count,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"latency_stats": stats,
	})
}

// GetUserRoutingStrategy handles GET /auth/routing/strategy (user-facing)
func (h *RoutingHandler) GetUserRoutingStrategy(c *gin.Context) {
	ctx := c.Request.Context()
	
	// Get user ID from context
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "user not authenticated",
		})
		return
	}

	userID, ok := userIDStr.(uuid.UUID)
	if !ok {
		// Try to parse as string
		if idStr, ok := userIDStr.(string); ok {
			var err error
			userID, err = uuid.Parse(idStr)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "invalid user ID",
				})
				return
			}
		} else {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid user ID",
			})
			return
		}
	}

	// Check if strategy is locked
	var isLocked bool
	if h.SettingsRepository != nil {
		locked, err := h.SettingsRepository.IsRoutingStrategyLocked(ctx)
		if err == nil {
			isLocked = locked
		}
	}

	// Get user's strategy preference
	var userStrategy string
	if h.UserRepository != nil {
		pref, err := h.UserRepository.GetUserRoutingStrategy(ctx, userID)
		if err == nil {
			userStrategy = pref
		}
	}

	// Get default strategy
	var defaultStrategy string
	if h.SettingsRepository != nil {
		def, err := h.SettingsRepository.GetDefaultRoutingStrategy(ctx)
		if err == nil {
			defaultStrategy = def
		}
	}
	if defaultStrategy == "" {
		defaultStrategy = "model" // Fallback
	}

	// Determine effective strategy
	effectiveStrategy := userStrategy
	if effectiveStrategy == "" || isLocked {
		effectiveStrategy = defaultStrategy
	}

	c.JSON(http.StatusOK, gin.H{
		"strategy":           effectiveStrategy,
		"user_strategy":       userStrategy, // NULL if using default
		"default_strategy":    defaultStrategy,
		"is_locked":          isLocked,
		"available_strategies": []string{
			"model",
			"cost",
			"latency",
			"balanced",
			"custom",
		},
	})
}

// SetUserRoutingStrategy handles PUT /auth/routing/strategy (user-facing)
func (h *RoutingHandler) SetUserRoutingStrategy(c *gin.Context) {
	var req SetRoutingStrategyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": errors.ErrInvalidRequest.Error(),
			"details": err.Error(),
		})
		return
	}

	// Validate strategy
	validStrategies := map[string]bool{
		"model":    true,
		"cost":     true,
		"latency":  true,
		"balanced": true,
		"custom":   true,
	}
	if !validStrategies[req.Strategy] {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid strategy. Must be one of: model, cost, latency, balanced, custom",
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

	userID, ok := userIDStr.(uuid.UUID)
	if !ok {
		if idStr, ok := userIDStr.(string); ok {
			var err error
			userID, err = uuid.Parse(idStr)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "invalid user ID",
				})
				return
			}
		} else {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid user ID",
			})
			return
		}
	}

	ctx := c.Request.Context()

	// Check if strategy is locked
	if h.SettingsRepository != nil {
		locked, err := h.SettingsRepository.IsRoutingStrategyLocked(ctx)
		if err == nil && locked {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "routing strategy is locked by administrator",
				"message": "The routing strategy is locked by an administrator. You cannot override the default strategy.",
				"locked":  true,
			})
			return
		}
	}

	// Save user preference
	if h.UserRepository == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "user repository not available",
		})
		return
	}

	strategyPtr := &req.Strategy
	if err := h.UserRepository.SetUserRoutingStrategy(ctx, userID, strategyPtr); err != nil {
		h.Logger.Error().
			Err(err).
			Str("user_id", userID.String()).
			Str("strategy", req.Strategy).
			Msg("Failed to save user routing strategy")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to save routing strategy",
		})
		return
	}

	h.Logger.Info().
		Str("user_id", userID.String()).
		Str("strategy", req.Strategy).
		Msg("User routing strategy updated")

	c.JSON(http.StatusOK, gin.H{
		"message": "routing strategy updated",
		"strategy": req.Strategy,
	})
}

// ClearUserRoutingStrategy handles DELETE /auth/routing/strategy (user-facing)
// Clears user's routing strategy preference to use default
func (h *RoutingHandler) ClearUserRoutingStrategy(c *gin.Context) {
	// Get user ID from context
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "user not authenticated",
		})
		return
	}

	userID, ok := userIDStr.(uuid.UUID)
	if !ok {
		if idStr, ok := userIDStr.(string); ok {
			var err error
			userID, err = uuid.Parse(idStr)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "invalid user ID",
				})
				return
			}
		} else {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid user ID",
			})
			return
		}
	}

	ctx := c.Request.Context()

	// Check if strategy is locked
	if h.SettingsRepository != nil {
		locked, err := h.SettingsRepository.IsRoutingStrategyLocked(ctx)
		if err == nil && locked {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "routing strategy is locked by administrator",
				"message": "The routing strategy is locked by an administrator. You cannot override the default strategy.",
				"locked":  true,
			})
			return
		}
	}

	// Clear user preference (set to nil)
	if h.UserRepository == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "user repository not available",
		})
		return
	}

	if err := h.UserRepository.SetUserRoutingStrategy(ctx, userID, nil); err != nil {
		h.Logger.Error().
			Err(err).
			Str("user_id", userID.String()).
			Msg("Failed to clear user routing strategy")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to clear routing strategy",
		})
		return
	}

	h.Logger.Info().
		Str("user_id", userID.String()).
		Msg("User routing strategy cleared (using default)")

	c.JSON(http.StatusOK, gin.H{
		"message": "routing strategy cleared, now using default",
	})
}

// SetRoutingStrategyLock handles POST /admin/routing/strategy/lock
func (h *RoutingHandler) SetRoutingStrategyLock(c *gin.Context) {
	var req struct {
		Locked bool `json:"locked" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": errors.ErrInvalidRequest.Error(),
			"details": err.Error(),
		})
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	var updatedBy *uuid.UUID
	if exists {
		if uid, ok := userID.(uuid.UUID); ok {
			updatedBy = &uid
		} else if idStr, ok := userID.(string); ok {
			if parsed, err := uuid.Parse(idStr); err == nil {
				updatedBy = &parsed
			}
		}
	}

	ctx := c.Request.Context()

	if h.SettingsRepository == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "settings repository not available",
		})
		return
	}

	if err := h.SettingsRepository.SetRoutingStrategyLock(ctx, req.Locked, updatedBy); err != nil {
		h.Logger.Error().
			Err(err).
			Bool("locked", req.Locked).
			Msg("Failed to set routing strategy lock")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to set routing strategy lock",
		})
		return
	}

	h.Logger.Info().
		Bool("locked", req.Locked).
		Msg("Routing strategy lock updated")

	action := "unlocked"
	if req.Locked {
		action = "locked"
	}

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("routing strategy %s successfully", action),
		"locked":  req.Locked,
	})
}

