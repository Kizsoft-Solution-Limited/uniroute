package handlers

import (
	"net/http"
	"time"

	"github.com/Kizsoft-Solution-Limited/uniroute/internal/storage"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ErrorLogHandler handles error logging requests
type ErrorLogHandler struct {
	errorLogRepo *storage.ErrorLogRepository
}

// NewErrorLogHandler creates a new error log handler
func NewErrorLogHandler(errorLogRepo *storage.ErrorLogRepository) *ErrorLogHandler {
	return &ErrorLogHandler{
		errorLogRepo: errorLogRepo,
	}
}

// ErrorLogRequest represents the error log request from frontend
type ErrorLogRequest struct {
	ErrorType  string                 `json:"error_type" binding:"required"` // 'exception', 'message', 'network', 'server'
	Message    string                 `json:"message" binding:"required"`
	StackTrace *string                `json:"stack_trace,omitempty"`
	URL        *string                `json:"url,omitempty"`
	Context    map[string]interface{} `json:"context,omitempty"`
	Severity   string                 `json:"severity,omitempty"` // 'error', 'warning', 'info'
}

// HandleLogError handles POST /api/errors/log
// This endpoint receives error reports from the frontend
func (h *ErrorLogHandler) HandleLogError(c *gin.Context) {
	var req ErrorLogRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request",
			"details": err.Error(),
		})
		return
	}

	// Get user ID from context (if authenticated)
	var userID *uuid.UUID
	if userIDValue, exists := c.Get("user_id"); exists {
		if uid, ok := userIDValue.(uuid.UUID); ok {
			userID = &uid
		}
	}

	// Get client IP
	ipAddress := c.ClientIP()

	// Get user agent
	userAgent := c.GetHeader("User-Agent")

	// Get current URL from request
	currentURL := c.GetHeader("Referer")
	if currentURL == "" {
		currentURL = c.Request.URL.String()
	}

	// Set default severity if not provided
	severity := req.Severity
	if severity == "" {
		severity = "error"
	}

	// Create error log
	errorLog := &storage.ErrorLog{
		UserID:     userID,
		ErrorType:  req.ErrorType,
		Message:    req.Message,
		StackTrace: req.StackTrace,
		URL:        &currentURL,
		UserAgent:  &userAgent,
		IPAddress:  &ipAddress,
		Context:    req.Context,
		Severity:   severity,
		Resolved:   false,
	}

	// Save to database
	if err := h.errorLogRepo.CreateErrorLog(c.Request.Context(), errorLog); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to log error",
		})
		return
	}

	// Return success
	c.JSON(http.StatusOK, gin.H{
		"message": "Error logged successfully",
		"id":      errorLog.ID,
	})
}

// GetErrorLogsRequest represents query parameters for getting error logs
type GetErrorLogsRequest struct {
	UserID    *string `form:"user_id"`
	ErrorType string  `form:"error_type"`
	Severity  string  `form:"severity"`
	Resolved  *bool   `form:"resolved"`
	Limit     int     `form:"limit"`
	Offset    int     `form:"offset"`
}

// HandleGetErrorLogs handles GET /admin/errors (admin only)
func (h *ErrorLogHandler) HandleGetErrorLogs(c *gin.Context) {
	var req GetErrorLogsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid query parameters",
			"details": err.Error(),
		})
		return
	}

	// Build filters
	filters := storage.ErrorLogFilters{
		ErrorType: req.ErrorType,
		Severity:  req.Severity,
		Resolved:  req.Resolved,
		Limit:     req.Limit,
	}

	// Parse user ID if provided
	if req.UserID != nil && *req.UserID != "" {
		userID, err := uuid.Parse(*req.UserID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid user_id format",
			})
			return
		}
		filters.UserID = &userID
	}

	// Set default limit
	if filters.Limit <= 0 {
		filters.Limit = 50
	}
	if filters.Limit > 200 {
		filters.Limit = 200 // Max limit
	}

	// Get error logs
	logs, err := h.errorLogRepo.GetErrorLogs(c.Request.Context(), filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve error logs",
		})
		return
	}

	// Convert to response format with properly formatted dates
	errorLogsResponse := make([]map[string]interface{}, 0, len(logs))
	for _, log := range logs {
		responseLog := map[string]interface{}{
			"id":         log.ID.String(),
			"error_type": log.ErrorType,
			"message":    log.Message,
			"severity":   log.Severity,
			"resolved":   log.Resolved,
			"created_at": log.CreatedAt.Format(time.RFC3339),
		}

		if log.UserID != nil {
			responseLog["user_id"] = log.UserID.String()
		}
		if log.StackTrace != nil {
			responseLog["stack_trace"] = *log.StackTrace
		}
		if log.URL != nil {
			responseLog["url"] = *log.URL
		}
		if log.UserAgent != nil {
			responseLog["user_agent"] = *log.UserAgent
		}
		if log.IPAddress != nil {
			responseLog["ip_address"] = *log.IPAddress
		}
		if log.Context != nil && len(log.Context) > 0 {
			responseLog["context"] = log.Context
		}

		errorLogsResponse = append(errorLogsResponse, responseLog)
	}

	c.JSON(http.StatusOK, gin.H{
		"errors": errorLogsResponse,
		"count":  len(errorLogsResponse),
		"limit":  filters.Limit,
	})
}

// HandleMarkResolved handles PATCH /admin/errors/:id/resolve (admin only)
func (h *ErrorLogHandler) HandleMarkResolved(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid error ID",
		})
		return
	}

	if err := h.errorLogRepo.MarkResolved(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to mark error as resolved",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Error marked as resolved",
		"id":      id,
	})
}
