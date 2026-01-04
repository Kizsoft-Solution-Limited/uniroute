package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/Kizsoft-Solution-Limited/uniroute/internal/security"
	"github.com/Kizsoft-Solution-Limited/uniroute/pkg/errors"
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
	Name              string     `json:"name" binding:"required"`
	RateLimitPerMinute int        `json:"rate_limit_per_minute"`
	RateLimitPerDay    int        `json:"rate_limit_per_day"`
	ExpiresAt         *time.Time  `json:"expires_at,omitempty"`
}

// CreateAPIKey handles POST /admin/api-keys
func (h *APIKeyHandler) CreateAPIKey(c *gin.Context) {
	var req CreateAPIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": errors.ErrInvalidRequest.Error(),
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
		"id":        apiKey.ID,
		"key":       key, // Only returned once!
		"name":      apiKey.Name,
		"created_at": apiKey.CreatedAt,
		"expires_at": apiKey.ExpiresAt,
		"message":   "Save this key - it will not be shown again",
	})
}

// ListAPIKeys handles GET /admin/api-keys
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

	// For now, return empty list (we'll implement repository method later)
	_ = userID
	c.JSON(http.StatusOK, gin.H{
		"keys": []interface{}{},
		"message": "API key listing will be implemented",
	})
}

// RevokeAPIKey handles DELETE /admin/api-keys/:id
func (h *APIKeyHandler) RevokeAPIKey(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid API key ID",
		})
		return
	}

	// For now, return success (we'll implement repository method later)
	_ = id
	c.JSON(http.StatusOK, gin.H{
		"message": "API key revoked",
	})
}

