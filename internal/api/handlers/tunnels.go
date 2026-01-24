package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Kizsoft-Solution-Limited/uniroute/internal/tunnel"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// TunnelHandler handles tunnel management requests
type TunnelHandler struct {
	repository    *tunnel.TunnelRepository
	domainManager *tunnel.DomainManager
	logger        zerolog.Logger
}

// NewTunnelHandler creates a new tunnel handler
func NewTunnelHandler(repository *tunnel.TunnelRepository, logger zerolog.Logger) *TunnelHandler {
	return &TunnelHandler{
		repository: repository,
		logger:     logger,
	}
}

// SetDomainManager sets the domain manager for custom domain validation
func (h *TunnelHandler) SetDomainManager(manager *tunnel.DomainManager) {
	h.domainManager = manager
}

// ListTunnels handles GET /v1/tunnels or GET /auth/tunnels
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

	// List tunnels for user
	// This includes tunnels with matching user_id
	// Pass empty string for protocol filter to get all tunnels (not filtering by protocol)
	tunnels, err := h.repository.ListTunnelsByUser(c.Request.Context(), userID, "")
	if err != nil {
		h.logger.Error().Err(err).Str("user_id", userID.String()).Msg("Failed to list tunnels")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to list tunnels",
		})
		return
	}

	// Return empty array if no tunnels found (not an error)
	if tunnels == nil {
		tunnels = []*tunnel.Tunnel{}
	}

	// Format response
	tunnelList := make([]map[string]interface{}, len(tunnels))
	for i, t := range tunnels {
		tunnelList[i] = map[string]interface{}{
			"id":            t.ID,
			"subdomain":     t.Subdomain,
			"public_url":    t.PublicURL,
			"local_url":     t.LocalURL,
			"status":        t.Status,
			"request_count": t.RequestCount,
			"created_at":    t.CreatedAt.Format(time.RFC3339),
		}
		if !t.LastActive.IsZero() {
			tunnelList[i]["last_active"] = t.LastActive.Format(time.RFC3339)
		}
		if t.CustomDomain != "" {
			tunnelList[i]["custom_domain"] = t.CustomDomain
		}
		// Include protocol (default to "http" if not set for backward compatibility)
		protocol := t.Protocol
		if protocol == "" {
			protocol = "http" // Default for backward compatibility
		}
		tunnelList[i]["protocol"] = protocol
	}

	c.JSON(http.StatusOK, gin.H{
		"tunnels": tunnelList,
	})
}

// GetTunnel handles GET /v1/tunnels/:id
func (h *TunnelHandler) GetTunnel(c *gin.Context) {
	// Parse tunnel ID first to check for reserved names
	idStr := c.Param("id")

	// Reject reserved route names (prevent conflict with /tunnels/stats)
	// This check must happen BEFORE any other processing
	if idStr == "stats" {
		h.logger.Debug().Str("path", c.Request.URL.Path).Str("id_param", idStr).Msg("GetTunnel: rejected 'stats' as tunnel ID - should use /tunnels/stats route")
		c.JSON(http.StatusNotFound, gin.H{
			"error": "route not found. Use /tunnels/stats for statistics",
		})
		c.Abort()
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
	if t.CustomDomain != "" {
		response["custom_domain"] = t.CustomDomain
	}
	// Include protocol (default to "http" if not set for backward compatibility)
	protocol := t.Protocol
	if protocol == "" {
		protocol = "http" // Default for backward compatibility
	}
	response["protocol"] = protocol

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

// AssociateTunnel handles POST /auth/tunnels/:id/associate
// Associates a CLI-created tunnel (without user_id) with the authenticated user
func (h *TunnelHandler) AssociateTunnel(c *gin.Context) {
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

	// Get tunnel to verify it exists and doesn't already belong to another user
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

	// If tunnel already has a user_id, verify it's the same user
	if t.UserID != "" {
		tunnelUserID, err := uuid.Parse(t.UserID)
		if err == nil && tunnelUserID == userID {
			// Already associated with this user
			c.JSON(http.StatusOK, gin.H{
				"message": "tunnel already associated with your account",
			})
			return
		}
		// Belongs to another user
		c.JSON(http.StatusForbidden, gin.H{
			"error": "tunnel belongs to another user",
		})
		return
	}

	// Associate tunnel with user
	if err := h.repository.AssociateTunnelWithUser(c.Request.Context(), tunnelID, userID); err != nil {
		if err == tunnel.ErrTunnelNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "tunnel not found",
			})
		} else {
			h.logger.Error().Err(err).Msg("Failed to associate tunnel")
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to associate tunnel",
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "tunnel associated successfully",
	})
}

// SetCustomDomain handles POST/PUT /v1/tunnels/:id/domain
// Sets or updates a custom domain for a tunnel
func (h *TunnelHandler) SetCustomDomain(c *gin.Context) {
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

	// Parse request body
	var req struct {
		Domain string `json:"domain"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})
		return
	}

	// Trim whitespace from domain
	req.Domain = strings.TrimSpace(req.Domain)

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

	// Validate custom domain only if provided (empty string clears the domain)
	if req.Domain != "" && h.domainManager != nil {
		if err := h.domainManager.ValidateCustomDomain(c.Request.Context(), req.Domain); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "domain validation failed",
				"message": err.Error(),
			})
			return
		}
	}

	// Update custom domain in database (empty string clears the domain)
	if err := h.repository.UpdateTunnelCustomDomain(c.Request.Context(), tunnelID, req.Domain); err != nil {
		h.logger.Error().Err(err).Msg("Failed to update custom domain")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to update custom domain",
		})
		return
	}

	message := "custom domain updated successfully"
	if req.Domain == "" {
		message = "custom domain removed successfully"
	}

	c.JSON(http.StatusOK, gin.H{
		"message": message,
		"domain":  req.Domain,
	})
}

// GetTunnelStats handles GET /admin/tunnels/stats (admin only)
// Returns active tunnel count over time for charting - shows ALL tunnels for admins
func (h *TunnelHandler) GetTunnelStats(c *gin.Context) {
	h.logger.Debug().Str("path", c.Request.URL.Path).Msg("GetTunnelStats handler called (admin only)")

	// This endpoint is admin-only (enforced by AdminMiddleware in router)
	// Admin middleware ensures user has admin role, so we can safely get all tunnels

	// Parse time range (default: last 24 hours)
	hours := 24
	if hoursStr := c.Query("hours"); hoursStr != "" {
		if parsed, err := strconv.Atoi(hoursStr); err == nil && parsed > 0 && parsed <= 168 {
			hours = parsed
		}
	}

	// For time-series views with trend/slope:
	// - 6h: Time-series data with 30-minute intervals (12 points) to show trend
	// - 24h: Time-series data with 1-hour intervals (24 points) to show trend
	// - 7d: One point per day (daily aggregation)
	var intervalHours float64
	if hours >= 168 {
		// For 7-day views: use daily intervals (24 hours) - one point per day
		intervalHours = 24
	} else if hours == 6 {
		// For 6h view: use 30-minute intervals (0.5 hours) for trend visualization
		intervalHours = 0.5
	} else {
		// For 24h view: use 1-hour intervals for trend visualization
		intervalHours = 1.0
	}

	// Calculate time range
	endTime := time.Now()
	var startTime time.Time

	if hours >= 168 {
		// For 7d view: show data from the last 7 days from NOW (rolling window)
		// Start from 7 days ago at midnight, including today
		year, month, day := endTime.Date()
		todayStart := time.Date(year, month, day, 0, 0, 0, 0, endTime.Location())
		// Go back 6 days (to get 7 days total including today)
		startTime = todayStart.AddDate(0, 0, -6)
	} else {
		// For 6h and 24h views: show data from the last N hours from NOW (rolling window)
		// This shows all data from the past 6 or 24 hours
		startTime = endTime.Add(-time.Duration(hours) * time.Hour)
	}

	h.logger.Debug().
		Time("start_time", startTime).
		Time("end_time", endTime).
		Int("hours", hours).
		Float64("interval_hours", intervalHours).
		Msg("GetTunnelStats: fetching stats for all tunnels (admin)")

	// Log the calculated time range for debugging
	if hours >= 168 {
		h.logger.Debug().
			Msg("GetTunnelStats: 7d view - using daily intervals (one point per day)")
	} else if hours == 24 {
		h.logger.Debug().
			Msg("GetTunnelStats: 24h view - showing last 24 hours of data")
	} else if hours == 6 {
		h.logger.Debug().
			Msg("GetTunnelStats: 6h view - showing last 6 hours of data")
	}

	// Get tunnel statistics over time for ALL tunnels (uuid.Nil = all users/admin view)
	stats, err := h.repository.GetTunnelStatsOverTime(c.Request.Context(), uuid.Nil, startTime, endTime, intervalHours)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to get tunnel stats (all tunnels)")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to get tunnel statistics",
		})
		return
	}

	h.logger.Debug().
		Int("data_points", len(stats)).
		Float64("interval_hours", intervalHours).
		Int("hours", hours).
		Msg("GetTunnelStats: stats retrieved successfully")

	// For 6h/24h views, we expect multiple points (time-series data for trend visualization)
	// For 7d view, we expect one point per day
	if hours < 168 {
		h.logger.Debug().
			Int("data_points", len(stats)).
			Int("hours", hours).
			Float64("interval_hours", intervalHours).
			Msg("GetTunnelStats: time-series data for 6h/24h view (expecting multiple points)")
	} else {
		h.logger.Debug().
			Int("data_points", len(stats)).
			Msg("GetTunnelStats: daily aggregated data for 7d view")
	}

	c.JSON(http.StatusOK, gin.H{
		"period": gin.H{
			"start": startTime.Format(time.RFC3339),
			"end":   endTime.Format(time.RFC3339),
		},
		"interval_hours": intervalHours,
		"data":           stats,
	})
}
