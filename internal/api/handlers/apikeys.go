package handlers

import (
	"net/http"
	"time"

	"github.com/Kizsoft-Solution-Limited/uniroute/internal/security"
	"github.com/Kizsoft-Solution-Limited/uniroute/pkg/errors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// APIKeyHandler handles API key management
type APIKeyHandler struct {
	apiKeyService *security.APIKeyServiceV2
}

// NewAPIKeyHandler creates a new API key handler
func NewAPIKeyHandler(apiKeyService *security.APIKeyServiceV2) *APIKeyHandler {
	return &APIKeyHandler{
		apiKeyService: apiKeyService,
	}
}

// CreateAPIKeyRequest represents a request to create an API key
type CreateAPIKeyRequest struct {
	Name               string     `json:"name" binding:"required"`
	RateLimitPerMinute int        `json:"rate_limit_per_minute"`
	RateLimitPerDay    int        `json:"rate_limit_per_day"`
	ExpiresAt          *time.Time `json:"expires_at,omitempty"`
}

// CreateAPIKey handles POST /auth/api-keys (user route - users manage their own keys)
func (h *APIKeyHandler) CreateAPIKey(c *gin.Context) {
	var req CreateAPIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   errors.ErrInvalidRequest.Error(),
			"details": err.Error(),
		})
		return
	}

	// Get user ID from context (set by JWT middleware)
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

	// Set defaults
	if req.RateLimitPerMinute == 0 {
		req.RateLimitPerMinute = 60
	}
	if req.RateLimitPerDay == 0 {
		req.RateLimitPerDay = 10000
	}

	// Create API key
	key, apiKey, err := h.apiKeyService.CreateAPIKey(
		c.Request.Context(),
		userID,
		req.Name,
		req.RateLimitPerMinute,
		req.RateLimitPerDay,
		req.ExpiresAt,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":         apiKey.ID,
		"key":        key, // Only returned once!
		"name":       apiKey.Name,
		"created_at": apiKey.CreatedAt,
		"expires_at": apiKey.ExpiresAt,
		"message":    "Save this key - it will not be shown again",
	})
}

// ListAPIKeys handles GET /auth/api-keys (user route - users manage their own keys)
func (h *APIKeyHandler) ListAPIKeys(c *gin.Context) {
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

	// List API keys for user
	keys, err := h.apiKeyService.ListAPIKeysByUser(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Format response (don't expose sensitive data)
	keyList := make([]map[string]interface{}, len(keys))
	for i, key := range keys {
		keyList[i] = map[string]interface{}{
			"id":                    key.ID.String(),
			"name":                  key.Name,
			"rate_limit_per_minute": key.RateLimitPerMinute,
			"rate_limit_per_day":    key.RateLimitPerDay,
			"created_at":            key.CreatedAt.Format(time.RFC3339),
			"expires_at":            nil,
			"is_active":             key.IsActive,
		}
		if key.ExpiresAt != nil {
			keyList[i]["expires_at"] = key.ExpiresAt.Format(time.RFC3339)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"keys": keyList,
	})
}

// RevokeAPIKey handles DELETE /auth/api-keys/:id (user route - users manage their own keys)
func (h *APIKeyHandler) RevokeAPIKey(c *gin.Context) {
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

	// Parse API key ID
	idStr := c.Param("id")
	keyID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid API key ID",
		})
		return
	}

	// Verify the key belongs to the user
	keys, err := h.apiKeyService.ListAPIKeysByUser(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Check if key exists and belongs to user
	found := false
	for _, key := range keys {
		if key.ID == keyID {
			found = true
			break
		}
	}

	if !found {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "API key not found",
		})
		return
	}

	// Delete the API key
	if err := h.apiKeyService.DeleteAPIKey(c.Request.Context(), keyID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "API key revoked successfully",
	})
}
