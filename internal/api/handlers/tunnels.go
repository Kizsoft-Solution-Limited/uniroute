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

type TunnelHandler struct {
	repository    *tunnel.TunnelRepository
	domainManager *tunnel.DomainManager
	logger        zerolog.Logger
}

func NewTunnelHandler(repository *tunnel.TunnelRepository, logger zerolog.Logger) *TunnelHandler {
	return &TunnelHandler{
		repository: repository,
		logger:     logger,
	}
}

func (h *TunnelHandler) SetDomainManager(manager *tunnel.DomainManager) {
	h.domainManager = manager
}

func (h *TunnelHandler) ListTunnels(c *gin.Context) {
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

	tunnels, err := h.repository.ListTunnelsByUser(c.Request.Context(), userID, "")
	if err != nil {
		h.logger.Error().Err(err).Str("user_id", userID.String()).Msg("Failed to list tunnels")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to list tunnels",
		})
		return
	}

	if tunnels == nil {
		tunnels = []*tunnel.Tunnel{}
	}

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
		protocol := t.Protocol
		if protocol == "" {
			protocol = "http"
		}
		tunnelList[i]["protocol"] = protocol
	}

	c.JSON(http.StatusOK, gin.H{
		"tunnels": tunnelList,
	})
}

func (h *TunnelHandler) GetTunnel(c *gin.Context) {
	idStr := c.Param("id")

	if idStr == "stats" {
		h.logger.Debug().Str("path", c.Request.URL.Path).Str("id_param", idStr).Msg("GetTunnel: rejected 'stats' as tunnel ID - should use /tunnels/stats route")
		c.JSON(http.StatusNotFound, gin.H{
			"error": "route not found. Use /tunnels/stats for statistics",
		})
		c.Abort()
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

	tunnelID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid tunnel ID",
		})
		return
	}

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
	protocol := t.Protocol
	if protocol == "" {
		protocol = "http"
	}
	response["protocol"] = protocol

	c.JSON(http.StatusOK, gin.H{
		"tunnel": response,
	})
}

func (h *TunnelHandler) DisconnectTunnel(c *gin.Context) {
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

	idStr := c.Param("id")
	tunnelID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid tunnel ID",
		})
		return
	}

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

func (h *TunnelHandler) AssociateTunnel(c *gin.Context) {
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

	idStr := c.Param("id")
	tunnelID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid tunnel ID",
		})
		return
	}

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

func (h *TunnelHandler) SetCustomDomain(c *gin.Context) {
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

	idStr := c.Param("id")
	tunnelID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid tunnel ID",
		})
		return
	}

	var req struct {
		Domain string `json:"domain"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})
		return
	}

	req.Domain = strings.TrimSpace(req.Domain)

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

	if req.Domain != "" && h.domainManager != nil {
		if err := h.domainManager.ValidateCustomDomain(c.Request.Context(), req.Domain); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "domain validation failed",
				"message": err.Error(),
			})
			return
		}
	}

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

func (h *TunnelHandler) GetTunnelStats(c *gin.Context) {
	h.logger.Debug().Str("path", c.Request.URL.Path).Msg("GetTunnelStats handler called (admin only)")
	hours := 24
	if hoursStr := c.Query("hours"); hoursStr != "" {
		if parsed, err := strconv.Atoi(hoursStr); err == nil && parsed > 0 && parsed <= 168 {
			hours = parsed
		}
	}

	var intervalHours float64
	if hours >= 168 {
		intervalHours = 24
	} else if hours == 6 {
		intervalHours = 0.5
	} else {
		intervalHours = 1.0
	}

	endTime := time.Now()
	var startTime time.Time

	if hours >= 168 {
		year, month, day := endTime.Date()
		todayStart := time.Date(year, month, day, 0, 0, 0, 0, endTime.Location())
		startTime = todayStart.AddDate(0, 0, -6)
	} else {
		startTime = endTime.Add(-time.Duration(hours) * time.Hour)
	}

	h.logger.Debug().
		Time("start_time", startTime).
		Time("end_time", endTime).
		Int("hours", hours).
		Float64("interval_hours", intervalHours).
		Msg("GetTunnelStats: fetching stats for all tunnels (admin)")

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
