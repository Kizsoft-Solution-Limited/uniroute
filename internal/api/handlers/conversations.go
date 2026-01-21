package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/Kizsoft-Solution-Limited/uniroute/internal/storage"
	"github.com/Kizsoft-Solution-Limited/uniroute/pkg/errors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ConversationHandler handles conversation-related requests
type ConversationHandler struct {
	convRepo *storage.ConversationRepository
}

// NewConversationHandler creates a new conversation handler
func NewConversationHandler(convRepo *storage.ConversationRepository) *ConversationHandler {
	return &ConversationHandler{
		convRepo: convRepo,
	}
}

// CreateConversationRequest represents a request to create a conversation
type CreateConversationRequest struct {
	Title *string `json:"title"`
	Model *string `json:"model"`
}

// UpdateConversationRequest represents a request to update a conversation
type UpdateConversationRequest struct {
	Title *string `json:"title"`
	Model *string `json:"model"`
}

// CreateConversation handles POST /auth/conversations
func (h *ConversationHandler) CreateConversation(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "unauthorized",
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

	var req CreateConversationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   errors.ErrInvalidRequest.Error(),
			"details": err.Error(),
		})
		return
	}

	conv, err := h.convRepo.CreateConversation(c.Request.Context(), userID, req.Title, req.Model)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to create conversation",
		})
		return
	}

	c.JSON(http.StatusOK, conv)
}

// ListConversations handles GET /auth/conversations
func (h *ConversationHandler) ListConversations(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "unauthorized",
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

	// Get pagination parameters
	limit := 50 // Default limit
	offset := 0

	if limitStr := c.Query("limit"); limitStr != "" {
		var parsedLimit int
		if _, err := fmt.Sscanf(limitStr, "%d", &parsedLimit); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}
	if offsetStr := c.Query("offset"); offsetStr != "" {
		var parsedOffset int
		if _, err := fmt.Sscanf(offsetStr, "%d", &parsedOffset); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	conversations, err := h.convRepo.ListConversations(c.Request.Context(), userID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to list conversations",
		})
		return
	}

	c.JSON(http.StatusOK, conversations)
}

// GetConversation handles GET /auth/conversations/:id
func (h *ConversationHandler) GetConversation(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "unauthorized",
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

	conversationID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid conversation ID",
		})
		return
	}

	conv, err := h.convRepo.GetConversation(c.Request.Context(), conversationID, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "conversation not found",
		})
		return
	}

	// Get messages for this conversation
	messages, err := h.convRepo.GetMessages(c.Request.Context(), conversationID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to get messages",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"conversation": conv,
		"messages":     messages,
	})
}

// UpdateConversation handles PUT /auth/conversations/:id
func (h *ConversationHandler) UpdateConversation(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "unauthorized",
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

	conversationID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid conversation ID",
		})
		return
	}

	var req UpdateConversationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   errors.ErrInvalidRequest.Error(),
			"details": err.Error(),
		})
		return
	}

	err = h.convRepo.UpdateConversation(c.Request.Context(), conversationID, userID, req.Title, req.Model)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "conversation not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to update conversation",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "conversation updated",
	})
}

// DeleteConversation handles DELETE /auth/conversations/:id
func (h *ConversationHandler) DeleteConversation(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "unauthorized",
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

	conversationID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid conversation ID",
		})
		return
	}

	err = h.convRepo.DeleteConversation(c.Request.Context(), conversationID, userID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "conversation not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to delete conversation",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "conversation deleted",
	})
}
