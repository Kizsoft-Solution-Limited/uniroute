package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/Kizsoft-Solution-Limited/uniroute/internal/gateway"
	"github.com/Kizsoft-Solution-Limited/uniroute/pkg/errors"
	"github.com/rs/zerolog"
)

type ProviderHandler struct {
	Router *gateway.Router
	Logger zerolog.Logger
}

func NewProviderHandler(router *gateway.Router, logger zerolog.Logger) *ProviderHandler {
	return &ProviderHandler{
		Router: router,
		Logger: logger,
	}
}

func (h *ProviderHandler) ListProviders(c *gin.Context) {
	ctx := c.Request.Context()
	var providerDetails []map[string]interface{}

	if userIDVal, exists := c.Get("user_id"); exists {
		if userIDStr, ok := userIDVal.(string); ok {
			if userID, err := uuid.Parse(userIDStr); err == nil {
				providerDetails = h.Router.ListProviderDetailsForUser(ctx, &userID)
			}
		}
	}
	if providerDetails == nil {
		names := h.Router.ListProviders()
		providerDetails = make([]map[string]interface{}, 0, len(names))
		for _, name := range names {
			provider, err := h.Router.GetProvider(name)
			if err != nil {
				continue
			}
			healthy := provider.HealthCheck(ctx) == nil
			providerDetails = append(providerDetails, map[string]interface{}{
				"name":    name,
				"healthy": healthy,
				"models":  provider.GetModels(),
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"providers": providerDetails,
	})
}

func (h *ProviderHandler) GetProviderHealth(c *gin.Context) {
	providerName := c.Param("name")
	
	provider, err := h.Router.GetProvider(providerName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": errors.ErrProviderNotFound.Error(),
		})
		return
	}

	err = provider.HealthCheck(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"provider": providerName,
			"healthy":  false,
			"error":    err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"provider": providerName,
		"healthy":  true,
		"models":   provider.GetModels(),
	})
}

