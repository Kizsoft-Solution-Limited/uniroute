package handlers

import (
	"net/http"
	"strconv"

	"github.com/Kizsoft-Solution-Limited/uniroute/internal/storage"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// UserHandler handles user management operations
type UserHandler struct {
	userRepo *storage.UserRepository
	logger   zerolog.Logger
}

// NewUserHandler creates a new user handler
func NewUserHandler(userRepo *storage.UserRepository, logger zerolog.Logger) *UserHandler {
	return &UserHandler{
		userRepo: userRepo,
		logger:   logger,
	}
}

// UpdateProfileRequest represents a profile update request
type UpdateProfileRequest struct {
	Name string `json:"name" binding:"omitempty,max=255"`
	// Note: Email and roles are explicitly NOT included to prevent unauthorized changes
}

// HandleUpdateProfile handles updating user profile (name only, no role changes)
func (h *UserHandler) HandleUpdateProfile(c *gin.Context) {
	// Get user ID from JWT claims (set by JWT middleware)
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

	// Update only the name - roles are explicitly excluded for security
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

	// Get updated user
	user, err := h.userRepo.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to get updated user")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get updated user"})
		return
	}

	// Get user roles (default to ["user"] if not set)
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

// UpdateUserRolesRequest represents a request to update user roles (admin only)
type UpdateUserRolesRequest struct {
	Roles []string `json:"roles" binding:"required,min=1,dive,oneof=user admin"`
}

// HandleUpdateUserRoles handles updating user roles (admin only)
func (h *UserHandler) HandleUpdateUserRoles(c *gin.Context) {
	// Get target user ID from URL parameter
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

	// Validate that at least one role is provided
	if len(req.Roles) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "At least one role is required"})
		return
	}

	// Ensure 'user' role is always included (users should always have base user role)
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

	// Update user roles
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

	// Get updated user
	user, err := h.userRepo.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to get updated user")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get updated user"})
		return
	}

	// Get user roles (default to ["user"] if not set)
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

// HandleListUsers handles listing all users (admin only)
func (h *UserHandler) HandleListUsers(c *gin.Context) {
	// Parse pagination
	limit := 50 // default
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

	// Get total count
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
