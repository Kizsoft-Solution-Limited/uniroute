package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/Kizsoft-Solution-Limited/uniroute/internal/gateway"
	"github.com/Kizsoft-Solution-Limited/uniroute/internal/monitoring"
	"github.com/Kizsoft-Solution-Limited/uniroute/internal/providers"
	"github.com/Kizsoft-Solution-Limited/uniroute/internal/security"
	"github.com/Kizsoft-Solution-Limited/uniroute/internal/storage"
	"github.com/Kizsoft-Solution-Limited/uniroute/pkg/errors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"
)

// ChatHandler handles chat completion requests
type ChatHandler struct {
	router      *gateway.Router
	requestRepo *storage.RequestRepository
	convRepo    *storage.ConversationRepository
	logger      zerolog.Logger
	upgrader    websocket.Upgrader
	jwtService  *security.JWTService // For WebSocket authentication
}

// NewChatHandler creates a new chat handler
func NewChatHandler(router *gateway.Router, requestRepo *storage.RequestRepository, convRepo *storage.ConversationRepository, logger zerolog.Logger) *ChatHandler {
	return &ChatHandler{
		router:      router,
		requestRepo: requestRepo,
		convRepo:    convRepo,
		logger:      logger,
		jwtService:  nil, // Will be set if JWT service is available
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				// Allow all origins for now (can be restricted in production)
				return true
			},
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
	}
}

// SetJWTService sets the JWT service for WebSocket authentication
func (h *ChatHandler) SetJWTService(jwtService *security.JWTService) {
	h.jwtService = jwtService
}

// ChatRequestWithConversation extends ChatRequest with optional conversation_id
type ChatRequestWithConversation struct {
	providers.ChatRequest
	ConversationID *string `json:"conversation_id,omitempty"` // Optional: save to conversation
}

// HandleChat handles POST /v1/chat and POST /auth/chat requests
func (h *ChatHandler) HandleChat(c *gin.Context) {
	var reqWithConv ChatRequestWithConversation
	if err := c.ShouldBindJSON(&reqWithConv); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   errors.ErrInvalidRequest.Error(),
			"details": err.Error(),
		})
		return
	}

	req := reqWithConv.ChatRequest

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

		// Save messages to conversation if conversation_id is provided
		if reqWithConv.ConversationID != nil && h.convRepo != nil && userID != nil {
			conversationID, err := uuid.Parse(*reqWithConv.ConversationID)
			if err == nil {
				// Save user message (last message in the request)
				if len(req.Messages) > 0 {
					lastUserMsg := req.Messages[len(req.Messages)-1]
					_, _ = h.convRepo.AddMessage(c.Request.Context(), conversationID, lastUserMsg.Role, lastUserMsg.Content, nil)
				}

				// Save assistant response
				if resp != nil && len(resp.Choices) > 0 {
					assistantMsg := resp.Choices[0].Message
					metadata := map[string]interface{}{
						"tokens":  resp.Usage.TotalTokens,
						"cost":    resp.Cost,
						"provider": resp.Provider,
						"latency_ms": latency.Milliseconds(),
					}
					_, _ = h.convRepo.AddMessage(c.Request.Context(), conversationID, assistantMsg.Role, assistantMsg.Content, metadata)
				}
			}
		}
	}

	// Track request asynchronously (don't block response)
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

	// Record Prometheus metrics for monitoring
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

// HandleChatStream handles streaming chat requests via Server-Sent Events (SSE)
func (h *ChatHandler) HandleChatStream(c *gin.Context) {
	var reqWithConv ChatRequestWithConversation
	if err := c.ShouldBindJSON(&reqWithConv); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   errors.ErrInvalidRequest.Error(),
			"details": err.Error(),
		})
		return
	}

	req := reqWithConv.ChatRequest

	// Validate required fields
	if req.Model == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "model is required",
		})
		return
	}

	if len(req.Messages) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "messages is required",
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

	// Set up SSE headers
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no") // Disable nginx buffering

	// Track request start time
	startTime := time.Now()

	// Get streaming channels from router
	chunkChan, errChan := h.router.RouteStream(c.Request.Context(), req, userID)

	// Track response data for final metrics
	var responseID string
	var fullContent strings.Builder
	var finalUsage *providers.Usage
	var provider string = "unknown"
	var status string = "success"

	// Stream chunks via SSE
	c.Stream(func(w io.Writer) bool {
		select {
		case chunk, ok := <-chunkChan:
			if !ok {
				// Channel closed, send final metrics and close
				latency := time.Since(startTime)

				// Track request asynchronously
				if h.requestRepo != nil {
					go func() {
						requestRecord := &storage.Request{
							ID:            uuid.New(),
							APIKeyID:      apiKeyID,
							UserID:        userID,
							Provider:      provider,
							Model:         req.Model,
							RequestType:   "chat_stream",
							InputTokens:   func() int { if finalUsage != nil { return finalUsage.PromptTokens }; return 0 }(),
							OutputTokens:  func() int { if finalUsage != nil { return finalUsage.CompletionTokens }; return 0 }(),
							TotalTokens:   func() int { if finalUsage != nil { return finalUsage.TotalTokens }; return 0 }(),
							Cost:          0, // Will be calculated if needed
							LatencyMs:     int(latency.Milliseconds()),
							StatusCode:    http.StatusOK,
							ErrorMessage:  nil,
							CreatedAt:     time.Now(),
						}

						ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
						defer cancel()

						if err := h.requestRepo.Create(ctx, requestRecord); err != nil {
							h.logger.Error().Err(err).Msg("Failed to track streaming request")
						}
					}()
				}

				// Record Prometheus metrics
				if finalUsage != nil {
					monitoring.RecordRequest(provider, req.Model, status, latency.Seconds())
					monitoring.RecordTokens(provider, req.Model, "input", finalUsage.PromptTokens)
					monitoring.RecordTokens(provider, req.Model, "output", finalUsage.CompletionTokens)
				}

				// Save to conversation if needed
				if reqWithConv.ConversationID != nil && h.convRepo != nil && userID != nil {
					conversationID, err := uuid.Parse(*reqWithConv.ConversationID)
					if err == nil {
						// Save user message
						if len(req.Messages) > 0 {
							lastUserMsg := req.Messages[len(req.Messages)-1]
							_, _ = h.convRepo.AddMessage(c.Request.Context(), conversationID, lastUserMsg.Role, lastUserMsg.Content, nil)
						}

						// Save assistant response
						metadata := map[string]interface{}{
							"tokens":     func() int { if finalUsage != nil { return finalUsage.TotalTokens }; return 0 }(),
							"cost":       0,
							"provider":   provider,
							"latency_ms": latency.Milliseconds(),
						}
						_, _ = h.convRepo.AddMessage(c.Request.Context(), conversationID, "assistant", fullContent.String(), metadata)
					}
				}

				return false // Close stream
			}

			// Update tracking data
			if chunk.ID != "" {
				responseID = chunk.ID
			}
			if chunk.Content != "" {
				fullContent.WriteString(chunk.Content)
			}
			if chunk.Usage != nil {
				finalUsage = chunk.Usage
			}
			if chunk.Provider != "" {
				provider = chunk.Provider
			}

			// Send chunk as SSE event
			chunkJSON, err := json.Marshal(chunk)
			if err != nil {
				h.logger.Error().Err(err).Msg("Failed to marshal stream chunk")
				return false
			}

			// Write SSE format: "data: {...}\n\n"
			fmt.Fprintf(w, "data: %s\n\n", chunkJSON)
			// Note: io.Writer doesn't have Flush(), but gin's Stream handles flushing

			// Continue streaming if not done
			return !chunk.Done

		case err, ok := <-errChan:
			if !ok {
				return false // Channel closed
			}

			// Send error as SSE event
			status = "error"
			errorChunk := providers.StreamChunk{
				ID:      responseID,
				Content: "",
				Done:    true,
				Error:   err.Error(),
			}

			errorJSON, _ := json.Marshal(errorChunk)
			fmt.Fprintf(w, "data: %s\n\n", errorJSON)
			// Note: io.Writer doesn't have Flush(), but gin's Stream handles flushing

			// Track error request
			if h.requestRepo != nil {
				go func() {
					errorMsg := err.Error()
					requestRecord := &storage.Request{
						ID:           uuid.New(),
						APIKeyID:     apiKeyID,
						UserID:       userID,
						Provider:     "unknown",
						Model:        req.Model,
						RequestType:  "chat_stream",
						StatusCode:   http.StatusInternalServerError,
						ErrorMessage: &errorMsg,
						CreatedAt:    time.Now(),
					}

					ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
					defer cancel()

					if err := h.requestRepo.Create(ctx, requestRecord); err != nil {
						h.logger.Error().Err(err).Msg("Failed to track streaming error")
					}
				}()
			}

			return false // Close stream on error

		case <-c.Request.Context().Done():
			// Client disconnected
			return false
		}
	})
}

// WebSocketMessage represents a WebSocket message for chat streaming
type WebSocketMessage struct {
	Type    string                 `json:"type"` // "request", "ping", "pong"
	Request *ChatRequestWithConversation `json:"request,omitempty"`
	Data    map[string]interface{} `json:"data,omitempty"`
}

// WebSocketResponse represents a WebSocket response message
type WebSocketResponse struct {
	Type    string                 `json:"type"` // "chunk", "done", "error"
	ID      string                 `json:"id,omitempty"`
	Content string                 `json:"content,omitempty"`
	Done    bool                   `json:"done"`
	Usage   *providers.Usage       `json:"usage,omitempty"`
	Error   string                 `json:"error,omitempty"`
	Provider string                `json:"provider,omitempty"`
}

// HandleChatWebSocket handles WebSocket-based chat streaming
func (h *ChatHandler) HandleChatWebSocket(c *gin.Context) {
	// Get user ID and API key info from context (set by middleware before WebSocket upgrade)
	// The middleware runs before the handler, so context should be populated
	var userID *uuid.UUID
	var apiKeyID *uuid.UUID

	// Try to get user ID from context (set by JWT or API key middleware)
	if userIDStr, exists := c.Get("user_id"); exists {
		if idStr, ok := userIDStr.(string); ok {
			if parsed, err := uuid.Parse(idStr); err == nil {
				userID = &parsed
			}
		}
	}

	// Try to get API key ID from context (set by API key middleware)
	if apiKeyIDStr, exists := c.Get("api_key_id"); exists {
		if idStr, ok := apiKeyIDStr.(string); ok {
			if parsed, err := uuid.Parse(idStr); err == nil {
				apiKeyID = &parsed
			}
		}
	}

	// Also try to extract token from query parameter as fallback for JWT
	// (API keys are handled by middleware, but JWT tokens might be in query param)
	token := c.Query("token")
	if userID == nil && h.jwtService != nil && token != "" {
		// Only validate as JWT if it doesn't look like an API key (starts with "ur_")
		if !strings.HasPrefix(token, "ur_") {
			claims, err := h.jwtService.ValidateToken(token)
			if err == nil {
				parsedUserID, err := uuid.Parse(claims.UserID)
				if err == nil {
					userID = &parsedUserID
				}
			}
		}
	}
	
	// Note: API key authentication is handled by middleware before WebSocket upgrade
	// The middleware now supports both Authorization header and ?token=ur_xxx query parameter
	// The apiKeyID and userID should already be set in context for /v1/chat/ws endpoint

	// Upgrade connection to WebSocket
	ws, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to upgrade WebSocket connection")
		return
	}
	defer ws.Close()

	// Set read/write deadlines
	ws.SetReadDeadline(time.Now().Add(60 * time.Second))
	ws.SetWriteDeadline(time.Now().Add(10 * time.Second))
	ws.SetPongHandler(func(string) error {
		ws.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	// Start ping ticker to keep connection alive
	pingTicker := time.NewTicker(30 * time.Second)
	defer pingTicker.Stop()

	// Read initial request
	var wsMsg WebSocketMessage
	if err := ws.ReadJSON(&wsMsg); err != nil {
		h.logger.Error().Err(err).Msg("Failed to read WebSocket message")
		return
	}

	// Validate message type
	if wsMsg.Type != "request" || wsMsg.Request == nil {
		h.sendWebSocketError(ws, "Invalid message type. Expected 'request' with chat request")
		return
	}

	reqWithConv := *wsMsg.Request
	req := reqWithConv.ChatRequest

	// Validate required fields
	if req.Model == "" {
		h.sendWebSocketError(ws, "model is required")
		return
	}

	if len(req.Messages) == 0 {
		h.sendWebSocketError(ws, "messages is required")
		return
	}

	// apiKeyID and userID are already set above from context (set by middleware)
	// The middleware now supports both Authorization header and ?token=ur_xxx query parameter

	// Track request start time
	startTime := time.Now()

	// Get streaming channels from router
	chunkChan, errChan := h.router.RouteStream(c.Request.Context(), req, userID)

	// Track response data for final metrics
	var responseID string
	var fullContent strings.Builder
	var finalUsage *providers.Usage
	var provider string = "unknown"
	var status string = "success"
	defer pingTicker.Stop()

	// Handle ping/pong in a separate goroutine
	go func() {
		for {
			select {
			case <-pingTicker.C:
				ws.SetWriteDeadline(time.Now().Add(10 * time.Second))
				// Send ping frame (WebSocket protocol ping, not JSON)
				if err := ws.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
					return
				}
			}
		}
	}()

	// Stream chunks via WebSocket
	done := false
	for !done {
		select {
		case chunk, ok := <-chunkChan:
			if !ok {
				// Channel closed, send final metrics and close
				latency := time.Since(startTime)

				// Track request asynchronously
				if h.requestRepo != nil {
					go func() {
						requestRecord := &storage.Request{
							ID:            uuid.New(),
							APIKeyID:      apiKeyID,
							UserID:        userID,
							Provider:      provider,
							Model:         req.Model,
							RequestType:   "chat_websocket",
							InputTokens:   func() int { if finalUsage != nil { return finalUsage.PromptTokens }; return 0 }(),
							OutputTokens:  func() int { if finalUsage != nil { return finalUsage.CompletionTokens }; return 0 }(),
							TotalTokens:   func() int { if finalUsage != nil { return finalUsage.TotalTokens }; return 0 }(),
							Cost:          0,
							LatencyMs:     int(latency.Milliseconds()),
							StatusCode:    http.StatusOK,
							ErrorMessage:  nil,
							CreatedAt:     time.Now(),
						}

						ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
						defer cancel()

						if err := h.requestRepo.Create(ctx, requestRecord); err != nil {
							h.logger.Error().Err(err).Msg("Failed to track WebSocket streaming request")
						}
					}()
				}

				// Record Prometheus metrics
				if finalUsage != nil {
					monitoring.RecordRequest(provider, req.Model, status, latency.Seconds())
					monitoring.RecordTokens(provider, req.Model, "input", finalUsage.PromptTokens)
					monitoring.RecordTokens(provider, req.Model, "output", finalUsage.CompletionTokens)
				}

				// Save to conversation if needed
				if reqWithConv.ConversationID != nil && h.convRepo != nil && userID != nil {
					conversationID, err := uuid.Parse(*reqWithConv.ConversationID)
					if err == nil {
						// Save user message
						if len(req.Messages) > 0 {
							lastUserMsg := req.Messages[len(req.Messages)-1]
							_, _ = h.convRepo.AddMessage(c.Request.Context(), conversationID, lastUserMsg.Role, lastUserMsg.Content, nil)
						}

						// Save assistant response
						metadata := map[string]interface{}{
							"tokens":     func() int { if finalUsage != nil { return finalUsage.TotalTokens }; return 0 }(),
							"cost":       0,
							"provider":   provider,
							"latency_ms": latency.Milliseconds(),
						}
						_, _ = h.convRepo.AddMessage(c.Request.Context(), conversationID, "assistant", fullContent.String(), metadata)
					}
				}

				done = true
				break
			}

			// Update tracking data
			if chunk.ID != "" {
				responseID = chunk.ID
			}
			if chunk.Content != "" {
				fullContent.WriteString(chunk.Content)
			}
			if chunk.Usage != nil {
				finalUsage = chunk.Usage
			}
			if chunk.Provider != "" {
				provider = chunk.Provider
			}

			// Send chunk via WebSocket
			ws.SetWriteDeadline(time.Now().Add(10 * time.Second))
			response := WebSocketResponse{
				ID:       responseID, // Use tracked responseID
				Content:  chunk.Content,
				Done:     chunk.Done,
				Usage:    chunk.Usage,
				Provider: chunk.Provider,
			}

			if err := ws.WriteJSON(response); err != nil {
				h.logger.Error().Err(err).Msg("Failed to write WebSocket message")
				done = true
				break
			}

			if chunk.Done {
				done = true
				break
			}

		case err, ok := <-errChan:
			if !ok {
				done = true
				break
			}

			// Send error via WebSocket
			status = "error"
			h.sendWebSocketError(ws, err.Error())

			// Track error request
			if h.requestRepo != nil {
				go func() {
					errorMsg := err.Error()
					requestRecord := &storage.Request{
						ID:           uuid.New(),
						APIKeyID:     apiKeyID,
						UserID:       userID,
						Provider:     "unknown",
						Model:        req.Model,
						RequestType:  "chat_websocket",
						StatusCode:   http.StatusInternalServerError,
						ErrorMessage: &errorMsg,
						CreatedAt:    time.Now(),
					}

					ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
					defer cancel()

					if err := h.requestRepo.Create(ctx, requestRecord); err != nil {
						h.logger.Error().Err(err).Msg("Failed to track WebSocket streaming error")
					}
				}()
			}

			done = true
			break

		case <-c.Request.Context().Done():
			// Client disconnected
			done = true
			break
		}
	}
}

// sendWebSocketError sends an error message via WebSocket
func (h *ChatHandler) sendWebSocketError(ws *websocket.Conn, errorMsg string) {
	ws.SetWriteDeadline(time.Now().Add(10 * time.Second))
	response := WebSocketResponse{
		Error: errorMsg,
		Done:  true,
	}
	ws.WriteJSON(response)
}
