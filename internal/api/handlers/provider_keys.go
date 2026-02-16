package handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/Kizsoft-Solution-Limited/uniroute/internal/security"
	"github.com/Kizsoft-Solution-Limited/uniroute/pkg/errors"
)

// ProviderKeyValidator validates a provider API key with an outbound call to the provider.
type ProviderKeyValidator interface {
	ValidateKey(ctx context.Context, provider string, apiKey string) error
}

// ProviderKeyHandler handles provider key management (BYOK)
type ProviderKeyHandler struct {
	providerKeyService *security.ProviderKeyService
	keyValidator       ProviderKeyValidator // optional; if set, TestProviderKey calls the provider API
}

// NewProviderKeyHandler creates a new provider key handler. keyValidator may be nil (test endpoint will only check key existence).
func NewProviderKeyHandler(providerKeyService *security.ProviderKeyService, keyValidator ProviderKeyValidator) *ProviderKeyHandler {
	return &ProviderKeyHandler{
		providerKeyService: providerKeyService,
		keyValidator:       keyValidator,
	}
}

// AddProviderKeyRequest represents a request to add a provider key
type AddProviderKeyRequest struct {
	Provider string `json:"provider" binding:"required"` // 'openai', 'anthropic', 'google'
	APIKey   string `json:"api_key" binding:"required"`  // Plaintext API key (will be encrypted)
}

// AddProviderKey handles POST /auth/provider-keys (user route - users manage their own provider keys BYOK)
func (h *ProviderKeyHandler) AddProviderKey(c *gin.Context) {
	var req AddProviderKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   errors.ErrInvalidRequest.Error(),
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

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid user ID",
		})
		return
	}

	// Add provider key (will encrypt it)
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

// ListProviderKeys handles GET /auth/provider-keys (user route - users manage their own provider keys BYOK)
func (h *ProviderKeyHandler) ListProviderKeys(c *gin.Context) {
	// Get user ID from context
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

	// List provider keys (without decrypting)
	keys, err := h.providerKeyService.ListProviderKeys(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Return keys without sensitive data
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

// UpdateProviderKeyRequest represents a request to update a provider key
type UpdateProviderKeyRequest struct {
	APIKey string `json:"api_key" binding:"required"` // Plaintext API key (will be encrypted)
}

// UpdateProviderKey handles PUT /auth/provider-keys/:provider (user route - users manage their own provider keys BYOK)
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

	// Get user ID from context
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

	// Update provider key (will encrypt it)
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

// DeleteProviderKey handles DELETE /auth/provider-keys/:provider (user route - users manage their own provider keys BYOK)
func (h *ProviderKeyHandler) DeleteProviderKey(c *gin.Context) {
	provider := c.Param("provider")
	if provider == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "provider parameter is required",
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

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid user ID",
		})
		return
	}

	// Delete provider key
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

// TestProviderKey handles POST /auth/provider-keys/:provider/test (user route - users manage their own provider keys BYOK)
func (h *ProviderKeyHandler) TestProviderKey(c *gin.Context) {
	provider := c.Param("provider")
	if provider == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "provider parameter is required",
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

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid user ID",
		})
		return
	}

	// Get and test provider key
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

