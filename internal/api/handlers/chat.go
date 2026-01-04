package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/Kizsoft-Solution-Limited/uniroute/internal/gateway"
	"github.com/Kizsoft-Solution-Limited/uniroute/internal/monitoring"
	"github.com/Kizsoft-Solution-Limited/uniroute/internal/providers"
	"github.com/Kizsoft-Solution-Limited/uniroute/internal/storage"
	"github.com/Kizsoft-Solution-Limited/uniroute/pkg/errors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// ChatHandler handles chat completion requests
type ChatHandler struct {
	router      *gateway.Router
	requestRepo *storage.RequestRepository
	logger      zerolog.Logger
}

// NewChatHandler creates a new chat handler
func NewChatHandler(router *gateway.Router, requestRepo *storage.RequestRepository, logger zerolog.Logger) *ChatHandler {
	return &ChatHandler{
		router:      router,
		requestRepo: requestRepo,
		logger:      logger,
	}
}

// HandleChat handles POST /v1/chat requests
func (h *ChatHandler) HandleChat(c *gin.Context) {
	var req providers.ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   errors.ErrInvalidRequest.Error(),
			"details": err.Error(),
		})
		return
	}

	// Validate required fields
	if req.Model == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "model is required",
		})
		return
	}

	if len(req.Messages) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "messages are required",
		})
		return
	}

	// Get API key info from context (for tracking)
	apiKeyIDStr, _ := c.Get("api_key_id")
	userIDStr, _ := c.Get("user_id")

	var apiKeyID *uuid.UUID
	var userID *uuid.UUID

	if id, ok := apiKeyIDStr.(string); ok {
		if parsed, err := uuid.Parse(id); err == nil {
			apiKeyID = &parsed
		}
	}
	if id, ok := userIDStr.(string); ok {
		if parsed, err := uuid.Parse(id); err == nil {
			userID = &parsed
		}
	}

	// Track request start time
	startTime := time.Now()

	// Route request to appropriate provider (with user ID for BYOK)
	resp, err := h.router.Route(c.Request.Context(), req, userID)
	latency := time.Since(startTime)

	// Determine status
	statusCode := http.StatusOK
	status := "success"
	var errorMsg *string
	provider := "unknown"

	if err != nil {
		statusCode = http.StatusInternalServerError
		status = "error"
		msg := err.Error()
		errorMsg = &msg
		c.JSON(statusCode, gin.H{
			"error": err.Error(),
		})
	} else {
		provider = resp.Provider
		c.JSON(http.StatusOK, resp)
	}

	// Phase 5: Track request (async, don't block response)
	if h.requestRepo != nil {
		go func() {
			requestRecord := &storage.Request{
				ID:          uuid.New(),
				APIKeyID:    apiKeyID,
				UserID:      userID,
				Provider:    provider,
				Model:       req.Model,
				RequestType: "chat",
				InputTokens: func() int {
					if resp != nil {
						return resp.Usage.PromptTokens
					}
					return 0
				}(),
				OutputTokens: func() int {
					if resp != nil {
						return resp.Usage.CompletionTokens
					}
					return 0
				}(),
				TotalTokens: func() int {
					if resp != nil {
						return resp.Usage.TotalTokens
					}
					return 0
				}(),
				Cost: func() float64 {
					if resp != nil {
						return resp.Cost
					}
					return 0
				}(),
				LatencyMs:    int(latency.Milliseconds()),
				StatusCode:   statusCode,
				ErrorMessage: errorMsg,
				CreatedAt:    time.Now(),
			}

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			if err := h.requestRepo.Create(ctx, requestRecord); err != nil {
				h.logger.Error().Err(err).Msg("Failed to track request")
			}
		}()
	}

	// Phase 5: Record Prometheus metrics
	if err == nil && resp != nil {
		monitoring.RecordRequest(resp.Provider, resp.Model, status, latency.Seconds())
		monitoring.RecordTokens(resp.Provider, resp.Model, "input", resp.Usage.PromptTokens)
		monitoring.RecordTokens(resp.Provider, resp.Model, "output", resp.Usage.CompletionTokens)
		if resp.Cost > 0 {
			monitoring.RecordCost(resp.Provider, resp.Model, resp.Cost)
		}
	} else {
		monitoring.RecordRequest("unknown", req.Model, status, latency.Seconds())
	}
}
