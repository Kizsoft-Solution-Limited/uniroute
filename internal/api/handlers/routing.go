package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/Kizsoft-Solution-Limited/uniroute/internal/gateway"
	"github.com/Kizsoft-Solution-Limited/uniroute/internal/providers"
	"github.com/Kizsoft-Solution-Limited/uniroute/pkg/errors"
	"github.com/rs/zerolog"
)

// RoutingHandler handles routing configuration requests
type RoutingHandler struct {
	Router *gateway.Router
	Logger zerolog.Logger
}

// NewRoutingHandler creates a new routing handler
func NewRoutingHandler(router *gateway.Router, logger zerolog.Logger) *RoutingHandler {
	return &RoutingHandler{
		Router: router,
		Logger: logger,
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

	// Set strategy
	strategyType := gateway.StrategyType(req.Strategy)
	h.Router.SetStrategyType(strategyType)

	c.JSON(http.StatusOK, gin.H{
		"message": "routing strategy updated",
		"strategy": req.Strategy,
	})
}

// GetRoutingStrategy handles GET /admin/routing/strategy
func (h *RoutingHandler) GetRoutingStrategy(c *gin.Context) {
	// For now, return current strategy (we'll need to track this)
	c.JSON(http.StatusOK, gin.H{
		"strategy": "model", // Default
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

