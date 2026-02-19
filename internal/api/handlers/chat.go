package handlers

import (
	"bytes"
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

type ChatHandler struct {
	router      *gateway.Router
	requestRepo *storage.RequestRepository
	convRepo    *storage.ConversationRepository
	logger      zerolog.Logger
	upgrader    websocket.Upgrader
	jwtService  *security.JWTService // For WebSocket authentication
}

func NewChatHandler(router *gateway.Router, requestRepo *storage.RequestRepository, convRepo *storage.ConversationRepository, logger zerolog.Logger) *ChatHandler {
	return &ChatHandler{
		router:      router,
		requestRepo: requestRepo,
		convRepo:    convRepo,
		logger:      logger,
		jwtService:  nil, // Will be set if JWT service is available
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
	}
}

func (h *ChatHandler) SetJWTService(jwtService *security.JWTService) {
	h.jwtService = jwtService
}

type ChatRequestWithConversation struct {
	providers.ChatRequest
	ConversationID *string `json:"conversation_id,omitempty"`
}

type chatStreamRequest struct {
	ConversationID *string               `json:"conversation_id,omitempty"`
	Model          string                `json:"model"`
	Messages       []providers.Message   `json:"messages"`
	Temperature    float64               `json:"temperature,omitempty"`
	MaxTokens      int                   `json:"max_tokens,omitempty"`
}

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

	startTime := time.Now()
	resp, err := h.router.Route(c.Request.Context(), req, userID)
	latency := time.Since(startTime)

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

		if reqWithConv.ConversationID != nil && h.convRepo != nil && userID != nil {
			conversationID, err := uuid.Parse(*reqWithConv.ConversationID)
			if err == nil {
				if len(req.Messages) > 0 {
					lastUserMsg := req.Messages[len(req.Messages)-1]
					if _, addErr := h.convRepo.AddMessage(c.Request.Context(), conversationID, lastUserMsg.Role, lastUserMsg.Content, nil); addErr != nil {
						h.logger.Warn().Err(addErr).Str("conversation_id", conversationID.String()).Msg("Failed to save user message to conversation")
					}
				}
				if resp != nil && len(resp.Choices) > 0 {
					assistantMsg := resp.Choices[0].Message
					metadata := map[string]interface{}{
						"tokens":  resp.Usage.TotalTokens,
						"cost":    resp.Cost,
						"provider": resp.Provider,
						"latency_ms": latency.Milliseconds(),
					}
					if _, addErr := h.convRepo.AddMessage(c.Request.Context(), conversationID, assistantMsg.Role, assistantMsg.Content, metadata); addErr != nil {
						h.logger.Warn().Err(addErr).Str("conversation_id", conversationID.String()).Msg("Failed to save assistant message to conversation")
					}
				}
			}
		}
	}

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

func (h *ChatHandler) HandleChatStream(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	c.Request.Body = io.NopCloser(bytes.NewReader(body))

	var streamReq chatStreamRequest
	if err := json.Unmarshal(body, &streamReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   errors.ErrInvalidRequest.Error(),
			"details": err.Error(),
		})
		return
	}
	if streamReq.ConversationID == nil {
		var raw map[string]interface{}
		if json.Unmarshal(body, &raw) == nil {
			if v, ok := raw["conversation_id"]; ok {
				if s, ok := v.(string); ok && s != "" {
					streamReq.ConversationID = &s
				}
			}
		}
	}

	req := providers.ChatRequest{
		Model:       streamReq.Model,
		Messages:    streamReq.Messages,
		Temperature: streamReq.Temperature,
		MaxTokens:   streamReq.MaxTokens,
	}
	reqWithConv := ChatRequestWithConversation{
		ChatRequest:    req,
		ConversationID: streamReq.ConversationID,
	}

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

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no") // Disable nginx buffering

	startTime := time.Now()
	chunkChan, errChan := h.router.RouteStream(c.Request.Context(), req, userID)

	var responseID string
	var fullContent strings.Builder
	var finalUsage *providers.Usage
	var provider string = "unknown"
	var status string = "success"

	c.Stream(func(w io.Writer) bool {
		select {
		case chunk, ok := <-chunkChan:
			if !ok {
				latency := time.Since(startTime)

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

				if finalUsage != nil {
					monitoring.RecordRequest(provider, req.Model, status, latency.Seconds())
					monitoring.RecordTokens(provider, req.Model, "input", finalUsage.PromptTokens)
					monitoring.RecordTokens(provider, req.Model, "output", finalUsage.CompletionTokens)
				}

				if reqWithConv.ConversationID != nil && h.convRepo != nil && userID != nil {
					conversationID, err := uuid.Parse(*reqWithConv.ConversationID)
					if err == nil {
						ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
						defer cancel()
						if len(req.Messages) > 0 {
							lastUserMsg := req.Messages[len(req.Messages)-1]
							if _, addErr := h.convRepo.AddMessage(ctx, conversationID, lastUserMsg.Role, lastUserMsg.Content, nil); addErr != nil {
								h.logger.Warn().Err(addErr).Str("conversation_id", conversationID.String()).Msg("Failed to save user message to conversation")
							}
						}
						metadata := map[string]interface{}{
							"tokens":     func() int { if finalUsage != nil { return finalUsage.TotalTokens }; return 0 }(),
							"cost":       0,
							"provider":   provider,
							"latency_ms": latency.Milliseconds(),
						}
						if _, addErr := h.convRepo.AddMessage(ctx, conversationID, "assistant", fullContent.String(), metadata); addErr != nil {
							h.logger.Warn().Err(addErr).Str("conversation_id", conversationID.String()).Msg("Failed to save assistant message to conversation")
						}
					} else {
						h.logger.Warn().Err(err).Str("conversation_id_raw", *reqWithConv.ConversationID).Msg("Invalid conversation_id, not saving messages")
					}
				} else {
					if reqWithConv.ConversationID == nil {
						h.logger.Debug().Msg("Chat stream: no conversation_id in request, messages not persisted")
					} else if h.convRepo == nil {
						h.logger.Debug().Msg("Chat stream: convRepo nil, messages not persisted")
					} else if userID == nil {
						h.logger.Debug().Msg("Chat stream: user_id nil, messages not persisted")
					}
				}

				return false // Close stream
			}

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

			chunkJSON, err := json.Marshal(chunk)
			if err != nil {
				h.logger.Error().Err(err).Msg("Failed to marshal stream chunk")
				return false
			}

			fmt.Fprintf(w, "data: %s\n\n", chunkJSON)
			// Note: io.Writer doesn't have Flush(), but gin's Stream handles flushing

			return !chunk.Done

		case err, ok := <-errChan:
			if !ok {
				return false // Channel closed
			}

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
			return false
		}
	})
}

type WebSocketMessage struct {
	Type    string                 `json:"type"` // "request", "ping", "pong"
	Request *ChatRequestWithConversation `json:"request,omitempty"`
	Data    map[string]interface{} `json:"data,omitempty"`
}

type WebSocketResponse struct {
	Type    string                 `json:"type"` // "chunk", "done", "error"
	ID      string                 `json:"id,omitempty"`
	Content string                 `json:"content,omitempty"`
	Done    bool                   `json:"done"`
	Usage   *providers.Usage       `json:"usage,omitempty"`
	Error   string                 `json:"error,omitempty"`
	Provider string                `json:"provider,omitempty"`
}

func (h *ChatHandler) HandleChatWebSocket(c *gin.Context) {
	var userID *uuid.UUID
	var apiKeyID *uuid.UUID

	if userIDStr, exists := c.Get("user_id"); exists {
		if idStr, ok := userIDStr.(string); ok {
			if parsed, err := uuid.Parse(idStr); err == nil {
				userID = &parsed
			}
		}
	}

	if apiKeyIDStr, exists := c.Get("api_key_id"); exists {
		if idStr, ok := apiKeyIDStr.(string); ok {
			if parsed, err := uuid.Parse(idStr); err == nil {
				apiKeyID = &parsed
			}
		}
	}

	token := c.Query("token")
	if userID == nil && h.jwtService != nil && token != "" {
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

	ws, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to upgrade WebSocket connection")
		return
	}
	defer ws.Close()

	ws.SetReadDeadline(time.Now().Add(60 * time.Second))
	ws.SetWriteDeadline(time.Now().Add(10 * time.Second))
	ws.SetPongHandler(func(string) error {
		ws.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	pingTicker := time.NewTicker(30 * time.Second)
	defer pingTicker.Stop()

	var wsMsg WebSocketMessage
	if err := ws.ReadJSON(&wsMsg); err != nil {
		h.logger.Error().Err(err).Msg("Failed to read WebSocket message")
		return
	}

	if wsMsg.Type != "request" || wsMsg.Request == nil {
		h.sendWebSocketError(ws, "Invalid message type. Expected 'request' with chat request")
		return
	}

	reqWithConv := *wsMsg.Request
	req := reqWithConv.ChatRequest

	if req.Model == "" {
		h.sendWebSocketError(ws, "model is required")
		return
	}

	if len(req.Messages) == 0 {
		h.sendWebSocketError(ws, "messages is required")
		return
	}

	startTime := time.Now()
	chunkChan, errChan := h.router.RouteStream(c.Request.Context(), req, userID)

	var responseID string
	var fullContent strings.Builder
	var finalUsage *providers.Usage
	var provider string = "unknown"
	var status string = "success"
	defer pingTicker.Stop()

	go func() {
		for {
			select {
			case <-pingTicker.C:
				ws.SetWriteDeadline(time.Now().Add(10 * time.Second))
				if err := ws.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
					return
				}
			}
		}
	}()

	done := false
	for !done {
		select {
		case chunk, ok := <-chunkChan:
			if !ok {
				latency := time.Since(startTime)

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

				if finalUsage != nil {
					monitoring.RecordRequest(provider, req.Model, status, latency.Seconds())
					monitoring.RecordTokens(provider, req.Model, "input", finalUsage.PromptTokens)
					monitoring.RecordTokens(provider, req.Model, "output", finalUsage.CompletionTokens)
				}

				if reqWithConv.ConversationID != nil && h.convRepo != nil && userID != nil {
					conversationID, err := uuid.Parse(*reqWithConv.ConversationID)
					if err == nil {
						if len(req.Messages) > 0 {
							lastUserMsg := req.Messages[len(req.Messages)-1]
							_, _ = h.convRepo.AddMessage(c.Request.Context(), conversationID, lastUserMsg.Role, lastUserMsg.Content, nil)
						}

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

			status = "error"
			h.sendWebSocketError(ws, err.Error())

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
			done = true
			break
		}
	}
}

func (h *ChatHandler) sendWebSocketError(ws *websocket.Conn, errorMsg string) {
	ws.SetWriteDeadline(time.Now().Add(10 * time.Second))
	response := WebSocketResponse{
		Error: errorMsg,
		Done:  true,
	}
	ws.WriteJSON(response)
}
