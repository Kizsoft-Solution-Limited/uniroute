package handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/Kizsoft-Solution-Limited/uniroute/internal/security"
	"github.com/Kizsoft-Solution-Limited/uniroute/pkg/errors"
)

type ProviderKeyValidator interface {
	ValidateKey(ctx context.Context, provider string, apiKey string) error
}

type ProviderKeyHandler struct {
	providerKeyService *security.ProviderKeyService
	keyValidator       ProviderKeyValidator // optional; if set, TestProviderKey calls the provider API
}

func NewProviderKeyHandler(providerKeyService *security.ProviderKeyService, keyValidator ProviderKeyValidator) *ProviderKeyHandler {
	return &ProviderKeyHandler{
		providerKeyService: providerKeyService,
		keyValidator:       keyValidator,
	}
}

type AddProviderKeyRequest struct {
	Provider string `json:"provider" binding:"required"` // 'openai', 'anthropic', 'google'
	APIKey   string `json:"api_key" binding:"required"`  // Plaintext API key (will be encrypted)
}

func (h *ProviderKeyHandler) AddProviderKey(c *gin.Context) {
	var req AddProviderKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   errors.ErrInvalidRequest.Error(),
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

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid user ID",
		})
		return
	}

	if err := h.providerKeyService.AddProviderKey(c.Request.Context(), userID, req.Provider, req.APIKey); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":  "Provider key added successfully",
		"provider": req.Provider,
	})
}

func (h *ProviderKeyHandler) ListProviderKeys(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "user not authenticated",
		})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid user ID",
		})
		return
	}

	keys, err := h.providerKeyService.ListProviderKeys(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	response := make([]map[string]interface{}, len(keys))
	for i, key := range keys {
		response[i] = map[string]interface{}{
			"id":         key.ID,
			"provider":   key.Provider,
			"is_active":  key.IsActive,
			"created_at": key.CreatedAt,
			"updated_at": key.UpdatedAt,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"keys": response,
	})
}

type UpdateProviderKeyRequest struct {
	APIKey string `json:"api_key" binding:"required"` // Plaintext API key (will be encrypted)
}

func (h *ProviderKeyHandler) UpdateProviderKey(c *gin.Context) {
	provider := c.Param("provider")
	if provider == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "provider parameter is required",
		})
		return
	}

	var req UpdateProviderKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   errors.ErrInvalidRequest.Error(),
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

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid user ID",
		})
		return
	}

	if err := h.providerKeyService.UpdateProviderKey(c.Request.Context(), userID, provider, req.APIKey); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Provider key updated successfully",
		"provider": provider,
	})
}

func (h *ProviderKeyHandler) DeleteProviderKey(c *gin.Context) {
	provider := c.Param("provider")
	if provider == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "provider parameter is required",
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

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid user ID",
		})
		return
	}

	if err := h.providerKeyService.DeleteProviderKey(c.Request.Context(), userID, provider); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Provider key deleted successfully",
		"provider": provider,
	})
}

func (h *ProviderKeyHandler) TestProviderKey(c *gin.Context) {
	provider := c.Param("provider")
	if provider == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "provider parameter is required",
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

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid user ID",
		})
		return
	}

	apiKey, err := h.providerKeyService.GetProviderKey(c.Request.Context(), userID, provider)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to retrieve provider key",
			"details": err.Error(),
		})
		return
	}

	if apiKey == "" {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "provider key not found",
		})
		return
	}

	if h.keyValidator != nil {
		if err := h.keyValidator.ValidateKey(c.Request.Context(), provider, apiKey); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "provider key test failed",
				"details": err.Error(),
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Provider key test successful",
		"provider": provider,
		"status":   "valid",
	})
}

