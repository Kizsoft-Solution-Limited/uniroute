package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/Kizsoft-Solution-Limited/uniroute/internal/storage"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// OAuthService handles OAuth authentication
type OAuthService struct {
	googleConfig *oauth2.Config
	xConfig      *oauth2.Config
	githubConfig *oauth2.Config
	userRepo     *storage.UserRepository
	frontendURL  string
}

// NewOAuthService creates a new OAuth service
// backendURL: The backend server URL (e.g., http://localhost:8084) for OAuth callbacks
// frontendURL: The frontend URL for redirecting after successful OAuth
func NewOAuthService(googleClientID, googleClientSecret, xClientID, xClientSecret, githubClientID, githubClientSecret, backendURL, frontendURL string, userRepo *storage.UserRepository) *OAuthService {
	service := &OAuthService{
		userRepo:    userRepo,
		frontendURL: frontendURL,
	}

	// Configure Google OAuth
	// OAuth providers redirect to the backend, which then redirects to frontend
	if googleClientID != "" && googleClientSecret != "" {
		service.googleConfig = &oauth2.Config{
			ClientID:     googleClientID,
			ClientSecret: googleClientSecret,
			RedirectURL: backendURL + "/auth/google/callback",
			Scopes:       []string{"openid", "profile", "email"},
			Endpoint:     google.Endpoint,
		}
	}

	// Configure X (Twitter) OAuth 2.0
	if xClientID != "" && xClientSecret != "" {
		service.xConfig = &oauth2.Config{
			ClientID:     xClientID,
			ClientSecret: xClientSecret,
			RedirectURL:  backendURL + "/auth/x/callback",
			Scopes:       []string{"tweet.read", "users.read", "offline.access"},
			Endpoint: oauth2.Endpoint{
				AuthURL:  "https://twitter.com/i/oauth2/authorize",
				TokenURL: "https://api.twitter.com/2/oauth2/token",
			},
		}
	}

	// Configure GitHub OAuth
	if githubClientID != "" && githubClientSecret != "" {
		service.githubConfig = &oauth2.Config{
			ClientID:     githubClientID,
			ClientSecret: githubClientSecret,
			RedirectURL:  backendURL + "/auth/github/callback",
			Scopes:       []string{"user:email"},
			Endpoint: oauth2.Endpoint{
				AuthURL:  "https://github.com/login/oauth/authorize",
				TokenURL: "https://github.com/login/oauth/access_token",
			},
		}
	}

	return service
}

// GetGoogleAuthURL returns the Google OAuth authorization URL
func (s *OAuthService) GetGoogleAuthURL(state string) (string, error) {
	if s.googleConfig == nil {
		return "", fmt.Errorf("Google OAuth not configured")
	}
	return s.googleConfig.AuthCodeURL(state, oauth2.AccessTypeOnline), nil
}

// GetXAuthURL returns the X (Twitter) OAuth authorization URL
func (s *OAuthService) GetXAuthURL(state string) (string, error) {
	if s.xConfig == nil {
		return "", fmt.Errorf("X OAuth not configured")
	}
	return s.xConfig.AuthCodeURL(state, oauth2.AccessTypeOnline), nil
}

// GoogleUserInfo represents Google user information
type GoogleUserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
}

// XUserInfo represents X (Twitter) user information
type XUserInfo struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Name     string `json:"name"`
	Email    string `json:"email"`
}

// HandleGoogleCallback handles Google OAuth callback
func (s *OAuthService) HandleGoogleCallback(ctx context.Context, code string) (*storage.User, error) {
	if s.googleConfig == nil {
		return nil, fmt.Errorf("Google OAuth not configured")
	}

	// Exchange code for token
	token, err := s.googleConfig.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange token: %w", err)
	}

	// Get user info from Google
	client := s.googleConfig.Client(ctx, token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var userInfo GoogleUserInfo
	if err := json.Unmarshal(body, &userInfo); err != nil {
		return nil, fmt.Errorf("failed to parse user info: %w", err)
	}

	if userInfo.Email == "" {
		return nil, fmt.Errorf("email not provided by Google")
	}

	// Check if user exists
	user, err := s.userRepo.GetUserByEmail(ctx, userInfo.Email)
	if err != nil {
		if err == storage.ErrUserNotFound {
			// Create new user with OAuth
			// Generate a random password (won't be used for OAuth users)
			randomPassword := uuid.New().String()
			user, err = s.userRepo.CreateUser(ctx, userInfo.Email, randomPassword, userInfo.Name)
			if err != nil {
				return nil, fmt.Errorf("failed to create user: %w", err)
			}
			// OAuth providers (Google) already verify emails, so mark as verified immediately
			// No email verification needed for OAuth users
			if err := s.userRepo.UpdateUserEmailVerified(ctx, user.ID, true); err != nil {
				return nil, fmt.Errorf("failed to verify email: %w", err)
			}
			user.EmailVerified = true
		} else {
			return nil, fmt.Errorf("failed to get user: %w", err)
		}
	}

	return user, nil
}

// HandleXCallback handles X (Twitter) OAuth callback
func (s *OAuthService) HandleXCallback(ctx context.Context, code string) (*storage.User, error) {
	if s.xConfig == nil {
		return nil, fmt.Errorf("X OAuth not configured")
	}

	// Exchange code for token
	token, err := s.xConfig.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange token: %w", err)
	}

	// Get user info from X API v2
	client := s.xConfig.Client(ctx, token)
	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.twitter.com/2/users/me?user.fields=profile_image_url", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// X API v2 returns data in a different format
	var xResponse struct {
		Data struct {
			ID       string `json:"id"`
			Name     string `json:"name"`
			Username string `json:"username"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &xResponse); err != nil {
		return nil, fmt.Errorf("failed to parse user info: %w", err)
	}

	// X doesn't provide email in basic OAuth 2.0 flow
	// Use username@x.oauth as a placeholder email
	email := fmt.Sprintf("%s@x.oauth", xResponse.Data.Username)

	// Check if user exists
	user, err := s.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		if err == storage.ErrUserNotFound {
			// Create new user with OAuth
			randomPassword := uuid.New().String()
			user, err = s.userRepo.CreateUser(ctx, email, randomPassword, xResponse.Data.Name)
			if err != nil {
				return nil, fmt.Errorf("failed to create user: %w", err)
			}
			// OAuth providers verify users, so mark email as verified immediately
			// No email verification needed for OAuth users
			if err := s.userRepo.UpdateUserEmailVerified(ctx, user.ID, true); err != nil {
				return nil, fmt.Errorf("failed to verify email: %w", err)
			}
			user.EmailVerified = true
		} else {
			return nil, fmt.Errorf("failed to get user: %w", err)
		}
	}

	return user, nil
}

// IsGoogleConfigured checks if Google OAuth is configured
func (s *OAuthService) IsGoogleConfigured() bool {
	return s.googleConfig != nil
}

// IsXConfigured checks if X OAuth is configured
func (s *OAuthService) IsXConfigured() bool {
	return s.xConfig != nil
}

// GetGithubAuthURL returns the GitHub OAuth authorization URL
func (s *OAuthService) GetGithubAuthURL(state string) (string, error) {
	if s.githubConfig == nil {
		return "", fmt.Errorf("GitHub OAuth not configured")
	}
	return s.githubConfig.AuthCodeURL(state, oauth2.AccessTypeOnline), nil
}

// IsGithubConfigured checks if GitHub OAuth is configured
func (s *OAuthService) IsGithubConfigured() bool {
	return s.githubConfig != nil
}

// GithubUserInfo represents GitHub user information
type GithubUserInfo struct {
	ID        int    `json:"id"`
	Login     string `json:"login"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url"`
}

// HandleGithubCallback handles GitHub OAuth callback
func (s *OAuthService) HandleGithubCallback(ctx context.Context, code string) (*storage.User, error) {
	if s.githubConfig == nil {
		return nil, fmt.Errorf("GitHub OAuth not configured")
	}

	// Exchange code for token
	token, err := s.githubConfig.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange token: %w", err)
	}

	// Get user info from GitHub
	client := s.githubConfig.Client(ctx, token)
	
	// First, get user profile
	resp, err := client.Get("https://api.github.com/user")
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var userInfo GithubUserInfo
	if err := json.Unmarshal(body, &userInfo); err != nil {
		return nil, fmt.Errorf("failed to parse user info: %w", err)
	}

	// If email is not in the user profile, fetch it from the emails endpoint
	if userInfo.Email == "" {
		emailsResp, err := client.Get("https://api.github.com/user/emails")
		if err == nil {
			defer emailsResp.Body.Close()
			emailsBody, err := io.ReadAll(emailsResp.Body)
			if err == nil {
				var emails []struct {
					Email    string `json:"email"`
					Primary  bool   `json:"primary"`
					Verified bool   `json:"verified"`
				}
				if err := json.Unmarshal(emailsBody, &emails); err == nil {
					// Find primary email
					for _, email := range emails {
						if email.Primary && email.Verified {
							userInfo.Email = email.Email
							break
						}
					}
					// If no primary verified email, use first verified email
					if userInfo.Email == "" {
						for _, email := range emails {
							if email.Verified {
								userInfo.Email = email.Email
								break
							}
						}
					}
				}
			}
		}
	}

	// Use login as fallback email if still no email
	if userInfo.Email == "" {
		userInfo.Email = fmt.Sprintf("%s@github.oauth", userInfo.Login)
	}

	// Use login as name if name is not provided
	if userInfo.Name == "" {
		userInfo.Name = userInfo.Login
	}

	// Check if user exists
	user, err := s.userRepo.GetUserByEmail(ctx, userInfo.Email)
	if err != nil {
		if err == storage.ErrUserNotFound {
			// Create new user with OAuth
			// Generate a random password (won't be used for OAuth users)
			randomPassword := uuid.New().String()
			user, err = s.userRepo.CreateUser(ctx, userInfo.Email, randomPassword, userInfo.Name)
			if err != nil {
				return nil, fmt.Errorf("failed to create user: %w", err)
			}
			// OAuth providers (GitHub) already verify emails, so mark as verified immediately
			// No email verification needed for OAuth users
			if err := s.userRepo.UpdateUserEmailVerified(ctx, user.ID, true); err != nil {
				return nil, fmt.Errorf("failed to verify email: %w", err)
			}
			user.EmailVerified = true
		} else {
			return nil, fmt.Errorf("failed to get user: %w", err)
		}
	}

	return user, nil
}
