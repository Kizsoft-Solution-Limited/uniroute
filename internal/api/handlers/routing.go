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

type RoutingHandler struct {
	Router              *gateway.Router
	SettingsRepository  *storage.SystemSettingsRepository
	UserRepository      *storage.UserRepository
	Logger              zerolog.Logger
}

func NewRoutingHandler(router *gateway.Router, settingsRepo *storage.SystemSettingsRepository, userRepo *storage.UserRepository, logger zerolog.Logger) *RoutingHandler {
	return &RoutingHandler{
		Router:             router,
		SettingsRepository: settingsRepo,
		UserRepository:     userRepo,
		Logger:             logger,
	}
}

type SetRoutingStrategyRequest struct {
	Strategy string `json:"strategy" binding:"required"` // "model", "cost", "latency", "balanced", "custom"
}

func (h *RoutingHandler) SetRoutingStrategy(c *gin.Context) {
	var req SetRoutingStrategyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": errors.ErrInvalidRequest.Error(),
			"details": err.Error(),
		})
		return
	}

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

	userID, exists := c.Get("user_id")
	var updatedBy *uuid.UUID
	if exists {
		if uid, ok := userID.(uuid.UUID); ok {
			updatedBy = &uid
		}
	}

	if h.SettingsRepository != nil {
		ctx := c.Request.Context()
		if err := h.SettingsRepository.SetDefaultRoutingStrategy(ctx, req.Strategy, updatedBy); err != nil {
			h.Logger.Error().
				Err(err).
				Str("strategy", req.Strategy).
				Msg("Failed to save default routing strategy to database")
		} else {
			h.Logger.Info().
				Str("strategy", req.Strategy).
				Msg("Saved default routing strategy to database")
		}
	}

	strategyType := gateway.StrategyType(req.Strategy)
	h.Router.SetStrategyType(strategyType)

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

func (h *RoutingHandler) GetRoutingStrategy(c *gin.Context) {
	ctx := c.Request.Context()
	var currentStrategy string

	if h.SettingsRepository != nil {
		dbStrategy, err := h.SettingsRepository.GetDefaultRoutingStrategy(ctx)
		if err == nil && dbStrategy != "" {
			currentStrategy = dbStrategy
			strategyType := gateway.StrategyType(currentStrategy)
			h.Router.SetStrategyType(strategyType)
		}
	}

	if currentStrategy == "" {
		currentStrategy = string(h.Router.GetStrategyType())
	}

	var isLocked bool
	if h.SettingsRepository != nil {
		locked, err := h.SettingsRepository.IsRoutingStrategyLocked(ctx)
		if err == nil {
			isLocked = locked
		}
	}

	if userIDVal, exists := c.Get("user_id"); exists {
		var userID *uuid.UUID
		switch v := userIDVal.(type) {
		case string:
			if parsed, err := uuid.Parse(v); err == nil {
				userID = &parsed
			}
		case uuid.UUID:
			uid := v
			userID = &uid
		}
		if userID != nil {
			currentStrategy = string(h.Router.GetStrategyForUser(ctx, userID))
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"strategy":             currentStrategy,
		"is_locked":            isLocked,
		"available_strategies": []string{"model", "cost", "latency", "balanced", "custom"},
	})
}

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

	estimates := make(map[string]float64)
	calculator := h.Router.GetCostCalculator()
	ctx := c.Request.Context()

	var providersToCheck []gateway.ProviderEntry
	if userIDVal, exists := c.Get("user_id"); exists {
		var userID *uuid.UUID
		switch v := userIDVal.(type) {
		case string:
			if parsed, err := uuid.Parse(v); err == nil {
				userID = &parsed
			}
		case uuid.UUID:
			uid := v
			userID = &uid
		}
		if userID != nil {
			providersToCheck = h.Router.GetProvidersForUser(ctx, userID)
		}
	}
	if len(providersToCheck) == 0 {
		for _, name := range h.Router.ListProviders() {
			p, err := h.Router.GetProvider(name)
			if err != nil {
				continue
			}
			providersToCheck = append(providersToCheck, gateway.ProviderEntry{Name: name, Provider: p})
		}
	}

	for _, entry := range providersToCheck {
		supports := false
		for _, model := range entry.Provider.GetModels() {
			if model == req.Model {
				supports = true
				break
			}
		}
		if !supports {
			continue
		}
		cost := calculator.EstimateCost(entry.Name, req.Model, req.Messages)
		estimates[entry.Name] = cost
	}

	c.JSON(http.StatusOK, gin.H{
		"model":    req.Model,
		"estimates": estimates,
	})
}

func (h *RoutingHandler) GetLatencyStats(c *gin.Context) {
	stats := make(map[string]interface{})
	ctx := c.Request.Context()
	tracker := h.Router.GetLatencyTracker()

	var providerNames []string
	if userIDVal, exists := c.Get("user_id"); exists {
		var userID *uuid.UUID
		switch v := userIDVal.(type) {
		case string:
			if parsed, err := uuid.Parse(v); err == nil {
				userID = &parsed
			}
		case uuid.UUID:
			uid := v
			userID = &uid
		}
		if userID != nil {
			entries := h.Router.GetProvidersForUser(ctx, userID)
			for _, e := range entries {
				providerNames = append(providerNames, e.Name)
			}
		}
	}
	if len(providerNames) == 0 {
		providerNames = h.Router.ListProviders()
	}
	for _, providerName := range providerNames {
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

func (h *RoutingHandler) GetUserRoutingStrategy(c *gin.Context) {
	ctx := c.Request.Context()

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

	var isLocked bool
	if h.SettingsRepository != nil {
		locked, err := h.SettingsRepository.IsRoutingStrategyLocked(ctx)
		if err == nil {
			isLocked = locked
		}
	}

	var userStrategy string
	if h.UserRepository != nil {
		pref, err := h.UserRepository.GetUserRoutingStrategy(ctx, userID)
		if err == nil {
			userStrategy = pref
		}
	}

	var defaultStrategy string
	if h.SettingsRepository != nil {
		def, err := h.SettingsRepository.GetDefaultRoutingStrategy(ctx)
		if err == nil {
			defaultStrategy = def
		}
	}
	if defaultStrategy == "" {
		defaultStrategy = "model"
	}

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

func (h *RoutingHandler) SetUserRoutingStrategy(c *gin.Context) {
	var req SetRoutingStrategyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": errors.ErrInvalidRequest.Error(),
			"details": err.Error(),
		})
		return
	}

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

func (h *RoutingHandler) ClearUserRoutingStrategy(c *gin.Context) {
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

func (h *RoutingHandler) SetRoutingStrategyLock(c *gin.Context) {
	var req struct {
		Locked *bool `json:"locked"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": errors.ErrInvalidRequest.Error(),
			"details": err.Error(),
		})
		return
	}

	if req.Locked == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": errors.ErrInvalidRequest.Error(),
			"details": "locked field is required",
		})
		return
	}
	locked := *req.Locked

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

	if err := h.SettingsRepository.SetRoutingStrategyLock(ctx, locked, updatedBy); err != nil {
		h.Logger.Error().
			Err(err).
			Bool("locked", locked).
			Msg("Failed to set routing strategy lock")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to set routing strategy lock",
		})
		return
	}

	h.Logger.Info().
		Bool("locked", locked).
		Msg("Routing strategy lock updated")

	action := "unlocked"
	if locked {
		action = "locked"
	}

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("routing strategy %s successfully", action),
		"locked":  locked,
	})
}

