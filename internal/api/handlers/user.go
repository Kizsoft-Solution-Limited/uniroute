package handlers

import (
	"net/http"
	"strconv"

	"github.com/Kizsoft-Solution-Limited/uniroute/internal/storage"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

type UserHandler struct {
	userRepo *storage.UserRepository
	logger   zerolog.Logger
}

func NewUserHandler(userRepo *storage.UserRepository, logger zerolog.Logger) *UserHandler {
	return &UserHandler{
		userRepo: userRepo,
		logger:   logger,
	}
}

type UpdateProfileRequest struct {
	Name string `json:"name" binding:"omitempty,max=255"`
	// Note: Email and roles are explicitly NOT included to prevent unauthorized changes
}

func (h *UserHandler) HandleUpdateProfile(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = h.userRepo.UpdateUserProfile(c.Request.Context(), userID, req.Name)
	if err != nil {
		if err == storage.ErrUserNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		h.logger.Error().Err(err).Str("user_id", userID.String()).Msg("Failed to update user profile")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
		return
	}

	user, err := h.userRepo.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to get updated user")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get updated user"})
		return
	}

	roles := user.Roles
	if len(roles) == 0 {
		roles = []string{"user"}
	}

	c.JSON(http.StatusOK, UserResponse{
		ID:            user.ID.String(),
		Email:         user.Email,
		Name:          user.Name,
		EmailVerified: user.EmailVerified,
		Roles:         roles,
		CreatedAt:     user.CreatedAt,
	})
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=8"`
}

func (h *UserHandler) HandleChangePassword(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.userRepo.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		if err == storage.ErrUserNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		h.logger.Error().Err(err).Str("user_id", userID.String()).Msg("Failed to get user")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user"})
		return
	}

	if err := h.userRepo.VerifyPassword(user.PasswordHash, req.CurrentPassword); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Current password is incorrect"})
		return
	}

	if err := h.userRepo.UpdatePassword(c.Request.Context(), userID, req.NewPassword); err != nil {
		h.logger.Error().Err(err).Str("user_id", userID.String()).Msg("Failed to update password")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password changed successfully"})
}

func (h *UserHandler) HandleGetUser(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	user, err := h.userRepo.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		if err == storage.ErrUserNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		h.logger.Error().Err(err).Msg("Failed to get user")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user"})
		return
	}

	roles := user.Roles
	if len(roles) == 0 {
		roles = []string{"user"}
	}

	c.JSON(http.StatusOK, gin.H{
		"id":             user.ID.String(),
		"email":          user.Email,
		"name":           user.Name,
		"email_verified": user.EmailVerified,
		"roles":          roles,
		"created_at":     user.CreatedAt,
	})
}

type UpdateUserRolesRequest struct {
	Roles []string `json:"roles" binding:"required,min=1,dive,oneof=user admin"`
}

func (h *UserHandler) HandleUpdateUserRoles(c *gin.Context) {
	userIDParam := c.Param("id")
	userID, err := uuid.Parse(userIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var req UpdateUserRolesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(req.Roles) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "At least one role is required"})
		return
	}

	hasUserRole := false
	for _, role := range req.Roles {
		if role == "user" {
			hasUserRole = true
			break
		}
	}
	if !hasUserRole {
		req.Roles = append(req.Roles, "user")
	}

	err = h.userRepo.UpdateUserRoles(c.Request.Context(), userID, req.Roles)
	if err != nil {
		if err == storage.ErrUserNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		h.logger.Error().Err(err).Str("user_id", userID.String()).Msg("Failed to update user roles")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user roles"})
		return
	}

	user, err := h.userRepo.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to get updated user")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get updated user"})
		return
	}

	roles := user.Roles
	if len(roles) == 0 {
		roles = []string{"user"}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User roles updated successfully",
		"user": UserResponse{
			ID:            user.ID.String(),
			Email:         user.Email,
			Name:          user.Name,
			EmailVerified: user.EmailVerified,
			Roles:         roles,
			CreatedAt:     user.CreatedAt,
		},
	})
}

func (h *UserHandler) HandleListUsers(c *gin.Context) {
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

	users, err := h.userRepo.ListUsers(c.Request.Context(), limit, offset)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to list users")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list users"})
		return
	}

	total, err := h.userRepo.CountUsers(c.Request.Context())
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to count users")
		// Continue without total count
		total = -1
	}

	// Convert to response format (exclude password hash)
	response := make([]UserResponse, 0, len(users))
	for _, user := range users {
		roles := user.Roles
		if len(roles) == 0 {
			roles = []string{"user"}
		}
		response = append(response, UserResponse{
			ID:            user.ID.String(),
			Email:         user.Email,
			Name:          user.Name,
			EmailVerified: user.EmailVerified,
			Roles:         roles,
			CreatedAt:     user.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"users": response,
		"limit": limit,
		"offset": offset,
		"count": len(response),
		"total": total,
	})
}
