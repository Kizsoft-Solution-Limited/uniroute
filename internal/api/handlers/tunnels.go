package handlers

import (
	"net/http"
	"time"

	"github.com/Kizsoft-Solution-Limited/uniroute/internal/tunnel"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// TunnelHandler handles tunnel management requests
type TunnelHandler struct {
	repository *tunnel.TunnelRepository
	logger     zerolog.Logger
}

// NewTunnelHandler creates a new tunnel handler
func NewTunnelHandler(repository *tunnel.TunnelRepository, logger zerolog.Logger) *TunnelHandler {
	return &TunnelHandler{
		repository: repository,
		logger:     logger,
	}
}

// ListTunnels handles GET /v1/tunnels
func (h *TunnelHandler) ListTunnels(c *gin.Context) {
	// Get user ID from context (set by JWT or API key middleware)
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

	// List tunnels for user (we need to add this method to repository)
	// For now, we'll get all active tunnels and filter by user
	// TODO: Add ListTunnelsByUser method to repository
	tunnels, err := h.repository.ListTunnelsByUser(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to list tunnels")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to list tunnels",
		})
		return
	}

	// Format response
	tunnelList := make([]map[string]interface{}, len(tunnels))
	for i, t := range tunnels {
		tunnelList[i] = map[string]interface{}{
			"id":           t.ID,
			"subdomain":    t.Subdomain,
			"public_url":   t.PublicURL,
			"local_url":    t.LocalURL,
			"status":       t.Status,
			"request_count": t.RequestCount,
			"created_at":   t.CreatedAt.Format(time.RFC3339),
		}
		if !t.LastActive.IsZero() {
			tunnelList[i]["last_active"] = t.LastActive.Format(time.RFC3339)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"tunnels": tunnelList,
	})
}

// GetTunnel handles GET /v1/tunnels/:id
func (h *TunnelHandler) GetTunnel(c *gin.Context) {
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

	// Parse tunnel ID
	idStr := c.Param("id")
	tunnelID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid tunnel ID",
		})
		return
	}

	// Get tunnel
	t, err := h.repository.GetTunnelByID(c.Request.Context(), tunnelID)
	if err != nil {
		if err == tunnel.ErrTunnelNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "tunnel not found",
			})
		} else {
			h.logger.Error().Err(err).Msg("Failed to get tunnel")
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to get tunnel",
			})
		}
		return
	}

	// Verify tunnel belongs to user
	if t.UserID == "" {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "access denied",
		})
		return
	}
	
	tunnelUserID, err := uuid.Parse(t.UserID)
	if err != nil || tunnelUserID != userID {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "access denied",
		})
		return
	}

	// Format response
	response := map[string]interface{}{
		"id":            t.ID,
		"subdomain":     t.Subdomain,
		"public_url":    t.PublicURL,
		"local_url":     t.LocalURL,
		"status":        t.Status,
		"request_count": t.RequestCount,
		"created_at":    t.CreatedAt.Format(time.RFC3339),
	}
	if !t.LastActive.IsZero() {
		response["last_active"] = t.LastActive.Format(time.RFC3339)
	}

	c.JSON(http.StatusOK, gin.H{
		"tunnel": response,
	})
}

// DisconnectTunnel handles POST /v1/tunnels/:id/disconnect
func (h *TunnelHandler) DisconnectTunnel(c *gin.Context) {
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

	// Parse tunnel ID
	idStr := c.Param("id")
	tunnelID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid tunnel ID",
		})
		return
	}

	// Get tunnel to verify ownership
	t, err := h.repository.GetTunnelByID(c.Request.Context(), tunnelID)
	if err != nil {
		if err == tunnel.ErrTunnelNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "tunnel not found",
			})
		} else {
			h.logger.Error().Err(err).Msg("Failed to get tunnel")
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to get tunnel",
			})
		}
		return
	}

	// Verify tunnel belongs to user
	if t.UserID == "" {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "access denied",
		})
		return
	}
	
	tunnelUserID, err := uuid.Parse(t.UserID)
	if err != nil || tunnelUserID != userID {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "access denied",
		})
		return
	}

	// Update tunnel status to inactive
	if err := h.repository.UpdateTunnelStatus(c.Request.Context(), t.ID, "inactive"); err != nil {
		h.logger.Error().Err(err).Msg("Failed to disconnect tunnel")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to disconnect tunnel",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "tunnel disconnected successfully",
	})
}

