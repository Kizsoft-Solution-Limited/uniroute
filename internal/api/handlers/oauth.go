package handlers

import (
	"net/http"
	"time"

	"github.com/Kizsoft-Solution-Limited/uniroute/internal/oauth"
	"github.com/Kizsoft-Solution-Limited/uniroute/internal/security"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// OAuthHandler handles OAuth authentication
type OAuthHandler struct {
	oauthService *oauth.OAuthService
	jwtService   *security.JWTService
	frontendURL  string
}

// NewOAuthHandler creates a new OAuth handler
func NewOAuthHandler(oauthService *oauth.OAuthService, jwtService *security.JWTService, frontendURL string) *OAuthHandler {
	return &OAuthHandler{
		oauthService: oauthService,
		jwtService:   jwtService,
		frontendURL:  frontendURL,
	}
}

// HandleGoogleAuth initiates Google OAuth flow
func (h *OAuthHandler) HandleGoogleAuth(c *gin.Context) {
	if !h.oauthService.IsGoogleConfigured() {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Google OAuth not configured",
		})
		return
	}

	// Generate state token for CSRF protection
	state := uuid.New().String()
	c.SetCookie("oauth_state", state, 600, "/", "", false, true) // 10 minutes, httpOnly

	authURL, err := h.oauthService.GetGoogleAuthURL(state)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate auth URL",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"auth_url": authURL,
		"state":    state,
	})
}

// HandleGoogleCallback handles Google OAuth callback
func (h *OAuthHandler) HandleGoogleCallback(c *gin.Context) {
	code := c.Query("code")
	state := c.Query("state")

	// Verify state token
	cookieState, err := c.Cookie("oauth_state")
	if err != nil || cookieState != state {
		c.Redirect(http.StatusFound, h.frontendURL+"/login?error=invalid_state")
		return
	}

	// Clear state cookie
	c.SetCookie("oauth_state", "", -1, "/", "", false, true)

	if code == "" {
		c.Redirect(http.StatusFound, h.frontendURL+"/login?error=no_code")
		return
	}

	// Handle OAuth callback
	user, err := h.oauthService.HandleGoogleCallback(c.Request.Context(), code)
	if err != nil {
		c.Redirect(http.StatusFound, h.frontendURL+"/login?error=oauth_failed")
		return
	}

	// Generate JWT token
	token, err := h.jwtService.GenerateToken(user.ID.String(), user.Email, user.Roles, 7*24*time.Hour)
	if err != nil {
		c.Redirect(http.StatusFound, h.frontendURL+"/login?error=token_generation_failed")
		return
	}

	// Redirect to frontend with token
	c.Redirect(http.StatusFound, h.frontendURL+"/auth/callback?token="+token+"&provider=google")
}

// HandleXAuth initiates X (Twitter) OAuth flow
func (h *OAuthHandler) HandleXAuth(c *gin.Context) {
	if !h.oauthService.IsXConfigured() {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "X OAuth not configured",
		})
		return
	}

	// Generate state token for CSRF protection
	state := uuid.New().String()
	c.SetCookie("oauth_state", state, 600, "/", "", false, true) // 10 minutes, httpOnly

	authURL, err := h.oauthService.GetXAuthURL(state)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate auth URL",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"auth_url": authURL,
		"state":    state,
	})
}

// HandleXCallback handles X (Twitter) OAuth callback
func (h *OAuthHandler) HandleXCallback(c *gin.Context) {
	code := c.Query("code")
	state := c.Query("state")

	// Verify state token
	cookieState, err := c.Cookie("oauth_state")
	if err != nil || cookieState != state {
		c.Redirect(http.StatusFound, h.frontendURL+"/login?error=invalid_state")
		return
	}

	// Clear state cookie
	c.SetCookie("oauth_state", "", -1, "/", "", false, true)

	if code == "" {
		c.Redirect(http.StatusFound, h.frontendURL+"/login?error=no_code")
		return
	}

	// Handle OAuth callback
	user, err := h.oauthService.HandleXCallback(c.Request.Context(), code)
	if err != nil {
		c.Redirect(http.StatusFound, h.frontendURL+"/login?error=oauth_failed")
		return
	}

	// Generate JWT token
	token, err := h.jwtService.GenerateToken(user.ID.String(), user.Email, user.Roles, 7*24*time.Hour)
	if err != nil {
		c.Redirect(http.StatusFound, h.frontendURL+"/login?error=token_generation_failed")
		return
	}

	// Redirect to frontend with token
	c.Redirect(http.StatusFound, h.frontendURL+"/auth/callback?token="+token+"&provider=x")
}

// HandleGithubAuth initiates GitHub OAuth flow
func (h *OAuthHandler) HandleGithubAuth(c *gin.Context) {
	if !h.oauthService.IsGithubConfigured() {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "GitHub OAuth not configured",
		})
		return
	}

	// Generate state token for CSRF protection
	state := uuid.New().String()
	c.SetCookie("oauth_state", state, 600, "/", "", false, true) // 10 minutes, httpOnly

	authURL, err := h.oauthService.GetGithubAuthURL(state)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate auth URL",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"auth_url": authURL,
		"state":    state,
	})
}

// HandleGithubCallback handles GitHub OAuth callback
func (h *OAuthHandler) HandleGithubCallback(c *gin.Context) {
	code := c.Query("code")
	state := c.Query("state")

	// Verify state token
	cookieState, err := c.Cookie("oauth_state")
	if err != nil || cookieState != state {
		c.Redirect(http.StatusFound, h.frontendURL+"/login?error=invalid_state")
		return
	}

	// Clear state cookie
	c.SetCookie("oauth_state", "", -1, "/", "", false, true)

	if code == "" {
		c.Redirect(http.StatusFound, h.frontendURL+"/login?error=no_code")
		return
	}

	// Handle OAuth callback
	user, err := h.oauthService.HandleGithubCallback(c.Request.Context(), code)
	if err != nil {
		c.Redirect(http.StatusFound, h.frontendURL+"/login?error=oauth_failed")
		return
	}

	// Generate JWT token
	token, err := h.jwtService.GenerateToken(user.ID.String(), user.Email, user.Roles, 7*24*time.Hour)
	if err != nil {
		c.Redirect(http.StatusFound, h.frontendURL+"/login?error=token_generation_failed")
		return
	}

	// Redirect to frontend with token
	c.Redirect(http.StatusFound, h.frontendURL+"/auth/callback?token="+token+"&provider=github")
}
