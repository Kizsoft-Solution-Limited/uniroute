package handlers

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Kizsoft-Solution-Limited/uniroute/internal/email"
	"github.com/Kizsoft-Solution-Limited/uniroute/internal/security"
	"github.com/Kizsoft-Solution-Limited/uniroute/internal/storage"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// AuthHandler handles authentication requests
type AuthHandler struct {
	userRepo        *storage.UserRepository
	jwtService      *security.JWTService
	emailService    *email.EmailService
	authRateLimiter *security.AuthRateLimiter
	frontendURL     string
	logger          zerolog.Logger
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(userRepo *storage.UserRepository, jwtService *security.JWTService, emailService *email.EmailService, authRateLimiter *security.AuthRateLimiter, frontendURL string, logger zerolog.Logger) *AuthHandler {
	return &AuthHandler{
		userRepo:        userRepo,
		jwtService:      jwtService,
		emailService:    emailService,
		authRateLimiter: authRateLimiter,
		frontendURL:     frontendURL,
		logger:          logger,
	}
}

// RegisterRequest represents a registration request
type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	Name     string `json:"name,omitempty"`
}

// LoginRequest represents a login request
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// AuthResponse represents an authentication response
type AuthResponse struct {
	Token string        `json:"token"`
	User  *UserResponse `json:"user"`
}

// UserResponse represents a user in the response
type UserResponse struct {
	ID            string    `json:"id"`
	Email         string    `json:"email"`
	Name          string    `json:"name,omitempty"`
	EmailVerified bool      `json:"email_verified"`
	Roles         []string  `json:"roles"` // Array of roles: ['user'], ['admin'], or ['user', 'admin']
	CreatedAt     time.Time `json:"created_at"`
}

// HandleRegister handles user registration
// SECURITY: This endpoint ALWAYS sets user role to ['user'] - it cannot be overridden
func (h *AuthHandler) HandleRegister(c *gin.Context) {
	// Security: First check raw request for any role/roles fields
	// This prevents hackers from trying to set admin role during registration
	var rawRequest map[string]interface{}
	if err := c.ShouldBindJSON(&rawRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Reject registration if roles or role field is present
	if roles, exists := rawRequest["roles"]; exists {
		h.logger.Warn().
			Str("ip", c.ClientIP()).
			Interface("attempted_roles", roles).
			Msg("Security: Registration attempt with roles field - rejected")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Roles cannot be set during registration",
		})
		return
	}
	if role, exists := rawRequest["role"]; exists {
		h.logger.Warn().
			Str("ip", c.ClientIP()).
			Interface("attempted_role", role).
			Msg("Security: Registration attempt with role field - rejected")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Role cannot be set during registration",
		})
		return
	}

	// Extract only allowed fields to prevent any role injection
	// RegisterRequest struct doesn't include roles field, so it's safe
	var req RegisterRequest
	if email, ok := rawRequest["email"].(string); ok {
		req.Email = email
	}
	if password, ok := rawRequest["password"].(string); ok {
		req.Password = password
	}
	if name, ok := rawRequest["name"].(string); ok {
		req.Name = name
	}

	// Validate required fields
	if req.Email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email is required"})
		return
	}
	if req.Password == "" || len(req.Password) < 8 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Password is required and must be at least 8 characters"})
		return
	}

	// Create user - roles are ALWAYS set to ['user'] in the database
	// This is enforced at the database level and cannot be overridden
	// Even if someone tries to pass roles, it will be ignored
	user, err := h.userRepo.CreateUser(c.Request.Context(), req.Email, req.Password, req.Name)
	if err != nil {
		if err == storage.ErrUserAlreadyExists {
			c.JSON(http.StatusConflict, gin.H{"error": "User already exists"})
			return
		}
		h.logger.Error().Err(err).Str("email", req.Email).Msg("Failed to create user")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	// Generate email verification token
	verificationToken := uuid.New().String()
	expiresAt := time.Now().Add(24 * time.Hour) // Token expires in 24 hours

	// Create verification token
	err = h.userRepo.CreateEmailVerificationToken(c.Request.Context(), user.ID, verificationToken, expiresAt)
	if err != nil {
		h.logger.Error().Err(err).Str("email", user.Email).Msg("Failed to create email verification token")
		// Continue anyway - user is created, they can request resend
	}

	// Send verification email
	if h.emailService != nil {
		smtpConfig := h.emailService.GetConfig()
		if configured, ok := smtpConfig["configured"].(bool); ok && configured {
			userName := user.Name
			if userName == "" {
				userName = user.Email
			}
			err = h.emailService.SendVerificationEmail(user.Email, userName, verificationToken, h.frontendURL)
			if err != nil {
				h.logger.Error().Err(err).Str("email", user.Email).Msg("Failed to send verification email")
			}
		}
	}

	// Don't auto-login - require email verification first
	// But return success message
	c.JSON(http.StatusCreated, gin.H{
		"message": "Registration successful. Please check your email to verify your account.",
		"user": &UserResponse{
			ID:            user.ID.String(),
			Email:         user.Email,
			Name:          user.Name,
			EmailVerified: user.EmailVerified,
			CreatedAt:     user.CreatedAt,
		},
	})
}

// HandleLogin handles user login
func (h *AuthHandler) HandleLogin(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error().Err(err).Msg("Failed to bind login request")
		// Provide more helpful error message
		if err.Error() == "EOF" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid request body. Please ensure the request contains valid JSON with email and password fields.",
			})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		return
	}

	// Get identifier for rate limiting (use email for more accurate tracking)
	// This prevents issues with shared IPs (NAT, VPN, etc.)
	identifier := "email:" + req.Email
	ipIdentifier := "ip:" + c.ClientIP()

	// Get user by email
	user, err := h.userRepo.GetUserByEmail(c.Request.Context(), req.Email)
	if err != nil {
		if err == storage.ErrUserNotFound {
			// Record failed attempt on both email and IP
			if h.authRateLimiter != nil {
				h.authRateLimiter.RecordFailedAttempt(c.Request.Context(), identifier)
				h.authRateLimiter.RecordFailedAttempt(c.Request.Context(), ipIdentifier)
			}
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
			return
		}
		h.logger.Error().Err(err).Msg("Failed to get user")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to authenticate"})
		return
	}

	// Verify password
	if err := h.userRepo.VerifyPassword(user.PasswordHash, req.Password); err != nil {
		// Record failed attempt on both email and IP
		if h.authRateLimiter != nil {
			h.authRateLimiter.RecordFailedAttempt(c.Request.Context(), identifier)
			h.authRateLimiter.RecordFailedAttempt(c.Request.Context(), ipIdentifier)
		}
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	// Check if email is verified - require verification before login
	// Handle both false and NULL (for existing users before migration)
	if !user.EmailVerified {
		// Automatically send verification email
		if h.emailService != nil {
			smtpConfig := h.emailService.GetConfig()
			if configured, ok := smtpConfig["configured"].(bool); ok && configured {
				// Generate new verification token
				verificationToken := uuid.New().String()
				expiresAt := time.Now().Add(24 * time.Hour) // Token expires in 24 hours

				// Create verification token
				err = h.userRepo.CreateEmailVerificationToken(c.Request.Context(), user.ID, verificationToken, expiresAt)
				if err != nil {
					h.logger.Error().Err(err).Str("email", user.Email).Msg("Failed to create email verification token during login")
				} else {
					// Send verification email
					userName := user.Name
					if userName == "" {
						userName = user.Email
					}

					err = h.emailService.SendVerificationEmail(user.Email, userName, verificationToken, h.frontendURL)
					if err != nil {
						h.logger.Error().Err(err).Str("email", user.Email).Msg("Failed to auto-send verification email on login")
					}
				}
			}
		}

		// Don't record as failed attempt - credentials are correct, just need verification
		c.JSON(http.StatusForbidden, gin.H{
			"error":      "Email not verified",
			"code":       "EMAIL_NOT_VERIFIED",
			"message":    "Please verify your email address before logging in. A verification link has been sent to your email.",
			"email":      user.Email,
			"can_resend": true,
		})
		return
	}

	// Success - reset rate limit for both email and IP
	if h.authRateLimiter != nil {
		h.authRateLimiter.RecordSuccess(c.Request.Context(), identifier)
		h.authRateLimiter.RecordSuccess(c.Request.Context(), ipIdentifier)
	}

	// Get user roles (default to ["user"] if not set)
	roles := user.Roles
	if len(roles) == 0 {
		roles = []string{"user"}
	}

	// Generate JWT token
	token, err := h.jwtService.GenerateToken(user.ID.String(), user.Email, roles, 24*time.Hour)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to generate token")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, AuthResponse{
		Token: token,
		User: &UserResponse{
			ID:            user.ID.String(),
			Email:         user.Email,
			Name:          user.Name,
			EmailVerified: user.EmailVerified,
			Roles:         roles,
			CreatedAt:     user.CreatedAt,
		},
	})
}

// HandleLogout handles user logout
func (h *AuthHandler) HandleLogout(c *gin.Context) {
	// In a stateless JWT system, logout is handled client-side by removing the token
	// We could implement token blacklisting here if needed
	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

// HandleProfile handles getting user profile
func (h *AuthHandler) HandleProfile(c *gin.Context) {
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

	// Get user from database
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

// HandleRefresh handles token refresh
func (h *AuthHandler) HandleRefresh(c *gin.Context) {
	// Get user ID, email, and role from JWT claims (set by JWT middleware)
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	email, exists := c.Get("user_email")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Get roles from context (set by JWT middleware), default to ["user"]
	roles, exists := c.Get("user_roles")
	rolesSlice := []string{"user"}
	if exists {
		if rolesArray, ok := roles.([]string); ok && len(rolesArray) > 0 {
			rolesSlice = rolesArray
		}
	}

	// Generate new token
	token, err := h.jwtService.GenerateToken(userIDStr.(string), email.(string), rolesSlice, 24*time.Hour)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to generate token")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

// PasswordResetRequest represents a password reset request
type PasswordResetRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// PasswordResetConfirmRequest represents a password reset confirmation
type PasswordResetConfirmRequest struct {
	Token    string `json:"token" binding:"required"`
	Password string `json:"password" binding:"required,min=8"`
}

// HandlePasswordResetRequest handles password reset requests
func (h *AuthHandler) HandlePasswordResetRequest(c *gin.Context) {
	var req PasswordResetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Use email-based tracking for more accurate rate limiting
	// This prevents issues with shared IPs (NAT, VPN, corporate networks)
	identifier := "email:" + req.Email
	ipIdentifier := "ip:" + c.ClientIP()

	// Get user by email
	user, err := h.userRepo.GetUserByEmail(c.Request.Context(), req.Email)
	if err != nil {
		// Don't reveal if user exists or not (security best practice)
		// Record failed attempt on both email and IP
		if h.authRateLimiter != nil {
			h.authRateLimiter.RecordFailedAttempt(c.Request.Context(), identifier)
			h.authRateLimiter.RecordFailedAttempt(c.Request.Context(), ipIdentifier)
		}
		c.JSON(http.StatusOK, gin.H{"message": "If the email exists, a password reset link has been sent"})
		return
	}

	// Generate reset token (in production, use crypto/rand)
	token := uuid.New().String()
	expiresAt := time.Now().Add(1 * time.Hour) // Token expires in 1 hour

	// Create reset token
	err = h.userRepo.CreatePasswordResetToken(c.Request.Context(), user.ID, token, expiresAt)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to create password reset token")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create reset token"})
		return
	}

	// Send password reset email
	if h.emailService != nil {
		userName := user.Name
		if userName == "" {
			userName = user.Email
		}
		err = h.emailService.SendPasswordResetEmail(user.Email, userName, token, h.frontendURL)
		if err != nil {
			h.logger.Error().Err(err).Msg("Failed to send password reset email")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send password reset email"})
			return
		}
	} else {
		h.logger.Warn().Msg("Email service not configured - password reset email not sent")
		// In development, log the token for testing
		if gin.Mode() == gin.DebugMode {
			h.logger.Info().Str("token", token).Str("email", user.Email).Msg("Password reset token generated (email service not configured)")
		}
	}

	// Success - reset rate limit for both email and IP
	if h.authRateLimiter != nil {
		h.authRateLimiter.RecordSuccess(c.Request.Context(), identifier)
		h.authRateLimiter.RecordSuccess(c.Request.Context(), ipIdentifier)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "If the email exists, a password reset link has been sent",
	})
}

// HandlePasswordResetConfirm handles password reset confirmation
func (h *AuthHandler) HandlePasswordResetConfirm(c *gin.Context) {
	var req PasswordResetConfirmRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get identifier for rate limiting
	identifier := "ip:" + c.ClientIP()

	// Get reset token
	resetToken, err := h.userRepo.GetPasswordResetToken(c.Request.Context(), req.Token)
	if err != nil {
		// Record failed attempt
		if h.authRateLimiter != nil {
			h.authRateLimiter.RecordFailedAttempt(c.Request.Context(), identifier)
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid or expired token"})
		return
	}

	// Update password
	err = h.userRepo.UpdatePassword(c.Request.Context(), resetToken.UserID, req.Password)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to update password")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password"})
		return
	}

	// Mark token as used
	err = h.userRepo.MarkPasswordResetTokenAsUsed(c.Request.Context(), req.Token)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to mark token as used")
		// Don't fail the request if this fails
	}

	// Success - reset rate limit
	if h.authRateLimiter != nil {
		h.authRateLimiter.RecordSuccess(c.Request.Context(), identifier)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password reset successfully"})
}

// VerifyEmailRequest represents an email verification request
type VerifyEmailRequest struct {
	Token string `json:"token" binding:"required"`
}

// HandleVerifyEmail handles email verification
func (h *AuthHandler) HandleVerifyEmail(c *gin.Context) {
	var req VerifyEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get identifier for rate limiting
	identifier := "ip:" + c.ClientIP()

	// Get verification token
	verificationToken, err := h.userRepo.GetEmailVerificationToken(c.Request.Context(), req.Token)
	if err != nil {
		// Record failed attempt
		if h.authRateLimiter != nil {
			h.authRateLimiter.RecordFailedAttempt(c.Request.Context(), identifier)
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid or expired token"})
		return
	}

	// Verify email
	err = h.userRepo.VerifyEmail(c.Request.Context(), verificationToken.UserID)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to verify email")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify email"})
		return
	}

	// Mark token as used
	err = h.userRepo.MarkEmailVerificationTokenAsUsed(c.Request.Context(), req.Token)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to mark token as used")
		// Don't fail the request if this fails
	}

	// Success - reset rate limit
	if h.authRateLimiter != nil {
		h.authRateLimiter.RecordSuccess(c.Request.Context(), identifier)
	}

	// Generate JWT token for auto-login after verification
	user, err := h.userRepo.GetUserByID(c.Request.Context(), verificationToken.UserID)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to get user after verification")
		c.JSON(http.StatusOK, gin.H{"message": "Email verified successfully"})
		return
	}

	// Send welcome email after successful verification
	if h.emailService != nil {
		dashboardURL := fmt.Sprintf("%s/dashboard", strings.TrimSuffix(h.frontendURL, "/"))
		if err := h.emailService.SendWelcomeEmail(user.Email, user.Name, dashboardURL); err != nil {
			// Log error but don't fail the verification
			h.logger.Error().Err(err).Str("email", user.Email).Msg("Failed to send welcome email")
		}
	}

	// Get user roles (default to ["user"] if not set)
	roles := user.Roles
	if len(roles) == 0 {
		roles = []string{"user"}
	}

	token, err := h.jwtService.GenerateToken(user.ID.String(), user.Email, roles, 24*time.Hour)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to generate token")
		c.JSON(http.StatusOK, gin.H{"message": "Email verified successfully"})
		return
	}

	c.JSON(http.StatusOK, AuthResponse{
		Token: token,
		User: &UserResponse{
			ID:            user.ID.String(),
			Email:         user.Email,
			Name:          user.Name,
			EmailVerified: true,
			CreatedAt:     user.CreatedAt,
		},
	})
}

// ResendVerificationRequest represents a resend verification request
type ResendVerificationRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// HandleResendVerification handles resending verification email
func (h *AuthHandler) HandleResendVerification(c *gin.Context) {
	h.logger.Info().
		Str("ip", c.ClientIP()).
		Str("method", c.Request.Method).
		Str("path", c.Request.URL.Path).
		Msg("Resend verification request received")

	var req ResendVerificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn().
			Err(err).
			Str("ip", c.ClientIP()).
			Msg("Failed to bind resend verification request")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info().
		Str("email", req.Email).
		Msg("Processing resend verification request")

	// Use email-based tracking for more accurate rate limiting
	identifier := "email:" + req.Email
	ipIdentifier := "ip:" + c.ClientIP()

	// Get user by email
	user, err := h.userRepo.GetUserByEmail(c.Request.Context(), req.Email)
	if err != nil {
		// Don't reveal if user exists or not (security best practice)
		// Record failed attempt on both email and IP
		if h.authRateLimiter != nil {
			h.authRateLimiter.RecordFailedAttempt(c.Request.Context(), identifier)
			h.authRateLimiter.RecordFailedAttempt(c.Request.Context(), ipIdentifier)
		}
		c.JSON(http.StatusOK, gin.H{"message": "If the email exists and is not verified, a verification link has been sent"})
		return
	}

	// Check if already verified
	if user.EmailVerified {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email is already verified"})
		return
	}

	// Generate new verification token
	verificationToken := uuid.New().String()
	expiresAt := time.Now().Add(24 * time.Hour) // Token expires in 24 hours

	// Create verification token
	err = h.userRepo.CreateEmailVerificationToken(c.Request.Context(), user.ID, verificationToken, expiresAt)
	if err != nil {
		h.logger.Error().Err(err).Str("email", user.Email).Msg("Failed to create email verification token")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create verification token"})
		return
	}

	// Send verification email
	if h.emailService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Email service not available"})
		return
	}

	smtpConfig := h.emailService.GetConfig()
	if configured, ok := smtpConfig["configured"].(bool); !ok || !configured {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "SMTP not configured"})
		return
	}

	userName := user.Name
	if userName == "" {
		userName = user.Email
	}

	err = h.emailService.SendVerificationEmail(user.Email, userName, verificationToken, h.frontendURL)
	if err != nil {
		h.logger.Error().Err(err).Str("email", user.Email).Msg("Failed to send verification email")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send verification email"})
		return
	}

	// Success - reset rate limit for both email and IP
	if h.authRateLimiter != nil {
		h.authRateLimiter.RecordSuccess(c.Request.Context(), identifier)
		h.authRateLimiter.RecordSuccess(c.Request.Context(), ipIdentifier)
	}

	c.JSON(http.StatusOK, gin.H{"message": "If the email exists and is not verified, a verification link has been sent"})
}
