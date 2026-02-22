package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/Kizsoft-Solution-Limited/uniroute/internal/storage"
	"github.com/rs/zerolog"
)

type AnalyticsHandler struct {
	requestRepo *storage.RequestRepository
	logger      zerolog.Logger
}

func NewAnalyticsHandler(requestRepo *storage.RequestRepository, logger zerolog.Logger) *AnalyticsHandler {
	return &AnalyticsHandler{
		requestRepo: requestRepo,
		logger:      logger,
	}
}

func (h *AnalyticsHandler) GetUsageStats(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	var userID *uuid.UUID
	if exists {
		if id, ok := userIDStr.(string); ok {
			if parsed, err := uuid.Parse(id); err == nil {
				userID = &parsed
			}
		}
	}

	startTime := time.Now().AddDate(0, 0, -30)
	endTime := time.Now()

	if startStr := c.Query("start_time"); startStr != "" {
		if parsed, err := time.Parse(time.RFC3339, startStr); err == nil {
			startTime = parsed
		}
	}

	if endStr := c.Query("end_time"); endStr != "" {
		if parsed, err := time.Parse(time.RFC3339, endStr); err == nil {
			endTime = parsed
		}
	}

	stats, err := h.requestRepo.GetUsageStats(c.Request.Context(), userID, startTime, endTime)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to get usage stats")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to get usage statistics",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"period": gin.H{
			"start": startTime.Format(time.RFC3339),
			"end":   endTime.Format(time.RFC3339),
		},
		"total_requests":    stats.TotalRequests,
		"total_tokens":      stats.TotalTokens,
		"total_cost":        stats.TotalCost,
		"average_latency_ms": stats.AverageLatencyMs,
		"requests_by_provider": stats.RequestsByProvider,
		"requests_by_model":    stats.RequestsByModel,
		"cost_by_provider":     stats.CostByProvider,
	})
}

func (h *AnalyticsHandler) GetUsageStatsAdmin(c *gin.Context) {
	startTime := time.Now().AddDate(0, 0, -30)
	endTime := time.Now()

	if startStr := c.Query("start_time"); startStr != "" {
		if parsed, err := time.Parse(time.RFC3339, startStr); err == nil {
			startTime = parsed
		}
	}

	if endStr := c.Query("end_time"); endStr != "" {
		if parsed, err := time.Parse(time.RFC3339, endStr); err == nil {
			endTime = parsed
		}
	}

	stats, err := h.requestRepo.GetUsageStats(c.Request.Context(), nil, startTime, endTime)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to get usage stats (admin)")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to get usage statistics",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"period": gin.H{
			"start": startTime.Format(time.RFC3339),
			"end":   endTime.Format(time.RFC3339),
		},
		"total_requests":     stats.TotalRequests,
		"total_tokens":       stats.TotalTokens,
		"total_cost":        stats.TotalCost,
		"average_latency_ms": stats.AverageLatencyMs,
		"requests_by_provider": stats.RequestsByProvider,
		"requests_by_model":   stats.RequestsByModel,
		"cost_by_provider":   stats.CostByProvider,
	})
}

func (h *AnalyticsHandler) GetRequests(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	var userID *uuid.UUID
	if exists {
		if id, ok := userIDStr.(string); ok {
			if parsed, err := uuid.Parse(id); err == nil {
				userID = &parsed
			}
		}
	}

	limit := 50
	if limitStr := c.Query("limit"); limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	offset := 0
	if offsetStr := c.Query("offset"); offsetStr != "" {
		if parsed, err := strconv.Atoi(offsetStr); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	requests, err := h.requestRepo.GetRequests(c.Request.Context(), userID, limit, offset)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to get requests")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to get requests",
		})
		return
	}

	response := make([]map[string]interface{}, 0, len(requests))
	for _, req := range requests {
		response = append(response, map[string]interface{}{
			"id":            req.ID.String(),
			"provider":      req.Provider,
			"model":         req.Model,
			"input_tokens":  req.InputTokens,
			"output_tokens": req.OutputTokens,
			"total_tokens":  req.TotalTokens,
			"cost":          req.Cost,
			"latency_ms":    req.LatencyMs,
			"status_code":   req.StatusCode,
			"created_at":   req.CreatedAt.Format(time.RFC3339),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"requests": response,
		"limit":    limit,
		"offset":   offset,
		"count":    len(response),
	})
}

