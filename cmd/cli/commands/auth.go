package commands

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Authenticate with UniRoute",
	Long: `Authenticate with the UniRoute server to manage your projects, API keys, and tunnels.

Your login session persists across CLI restarts, computer reboots, and system shutdowns.
You only need to log in once, and you'll stay logged in until you explicitly log out
or your token expires.

Commands:
  login    Login to your UniRoute account (session persists across restarts)
  logout   Logout and clear saved credentials
  status   Show current authentication status`,
}

var authLoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login to UniRoute",
	Long: `Login to your UniRoute account using email and password.

The server URL is determined by (in priority order):
  1. --server flag (if provided)
  2. UNIROUTE_API_URL environment variable
  3. Previously saved server URL from auth config
  4. Auto-detected local mode (http://localhost:8084)
  5. Default production server (https://api.uniroute.co)

Examples:
  # Use environment variable (recommended)
  export UNIROUTE_API_URL=http://localhost:8084
  uniroute auth login

  # Explicit server flag
  uniroute auth login --server http://localhost:8084

  # Auto-detect (will use localhost if local mode detected)
  uniroute auth login`,
	RunE: runAuthLogin,
}

var authLogoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Logout from UniRoute",
	Long:  `Logout and clear saved authentication credentials.`,
	RunE:  runAuthLogout,
}

var authStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show authentication status",
	Long:  `Show your current authentication status and user information.`,
	RunE:  runAuthStatus,
}

var (
	authEmail    string
	authPassword string
	authServer   string
)

func init() {
	authCmd.AddCommand(authLoginCmd)
	authCmd.AddCommand(authLogoutCmd)
	authCmd.AddCommand(authStatusCmd)

	authLoginCmd.Flags().StringVarP(&authEmail, "email", "e", "", "Email address")
	authLoginCmd.Flags().StringVarP(&authPassword, "password", "p", "", "Password")
	authLoginCmd.Flags().StringVarP(&authServer, "server", "s", "", "UniRoute server URL (overrides UNIROUTE_API_URL env var)")
}

// AuthConfig stores authentication information
type AuthConfig struct {
	Token     string `json:"token"`
	Email     string `json:"email"`
	ServerURL string `json:"server_url"`
	ExpiresAt string `json:"expires_at,omitempty"`
}

// getConfigPath returns the path to the config file
func getConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	configDir := filepath.Join(homeDir, ".uniroute")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return "", err
	}
	return filepath.Join(configDir, "auth.json"), nil
}

// loadAuthConfig loads authentication config from file
func loadAuthConfig() (*AuthConfig, error) {
	configPath, err := getConfigPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // No config file, not logged in
		}
		return nil, err
	}

	var config AuthConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// saveAuthConfig saves authentication config to file
func saveAuthConfig(config *AuthConfig) error {
	configPath, err := getConfigPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0600)
}

// getAuthToken returns the current auth token
// The token persists across CLI sessions, computer restarts, etc.
func getAuthToken() string {
	config, err := loadAuthConfig()
	if err != nil || config == nil {
		return ""
	}
	
	// Check if token has expired (if expiration info is available)
	if config.ExpiresAt != "" {
		expiresAt, err := time.Parse(time.RFC3339, config.ExpiresAt)
		if err == nil && time.Now().After(expiresAt) {
			// Token expired, clear it
			_ = runAuthLogout(nil, nil)
			return ""
		}
	}
	
	return config.Token
}

// isAuthenticated checks if user is currently authenticated
func isAuthenticated() bool {
	return getAuthToken() != ""
}

// clearExpiredToken clears the auth config if token appears to be expired
// This is called when API returns 401 Unauthorized
func clearExpiredToken() {
	configPath, err := getConfigPath()
	if err != nil {
		return
	}
	_ = os.Remove(configPath) // Ignore errors
}

// getServerURL returns the configured server URL
// Priority: 1. Environment variable (UNIROUTE_API_URL), 2. Saved auth config, 3. Auto-detect local mode, 4. Default
func getServerURL() string {
	// Priority 1: Environment variable (highest priority)
	if envURL := os.Getenv("UNIROUTE_API_URL"); envURL != "" {
		return envURL
	}
	
	// Priority 2: Saved auth config
	config, err := loadAuthConfig()
	if err == nil && config != nil && config.ServerURL != "" {
		return config.ServerURL
	}
	
	// Priority 3: Auto-detect local mode
	if isLocalMode() {
		return "http://localhost:8084"
	}
	
	// Priority 4: Default (only used if nothing else is configured)
	return "https://api.uniroute.co"
}

// isLocalMode detects if we're running in local development mode
// Checks environment variable or attempts to detect from localhost connectivity
func isLocalMode() bool {
	// Check explicit environment variable
	if env := os.Getenv("UNIROUTE_ENV"); env == "local" || env == "development" || env == "dev" {
		return true
	}
	
	// Check if API URL is set to localhost
	if apiURL := os.Getenv("UNIROUTE_API_URL"); apiURL != "" {
		return strings.Contains(apiURL, "localhost") || strings.Contains(apiURL, "127.0.0.1")
	}
	
	// Check if tunnel URL is explicitly set to localhost
	if tunnelURL := os.Getenv("UNIROUTE_TUNNEL_URL"); tunnelURL != "" {
		return strings.Contains(tunnelURL, "localhost") || strings.Contains(tunnelURL, "127.0.0.1")
	}
	
	// Try to detect: check if localhost:8055 (tunnel server), localhost:8080 (tunnel server), or localhost:8084 (gateway) is reachable
	// This helps detect local mode even without auth config
	client := &http.Client{Timeout: 1 * time.Second}
	
	// Check tunnel server on port 8055 (preferred)
	resp, err := client.Get("http://localhost:8055/health")
	if err == nil {
		resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			return true
		}
	}
	
	// Check tunnel server on port 8080 (fallback)
	resp, err = client.Get("http://localhost:8080/health")
	if err == nil {
		resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			return true
		}
	}
	
	// Check gateway server (localhost:8084) as fallback
	resp, err = client.Get("http://localhost:8084/health")
	if err == nil {
		resp.Body.Close()
		return resp.StatusCode == http.StatusOK
	}
	
	return false
}

// getTunnelServerURL returns the tunnel server URL
// Priority: 1. Environment variable (UNIROUTE_TUNNEL_URL), 2. Auto-detect local mode, 3. Default
func getTunnelServerURL() string {
	// Priority 1: Environment variable (highest priority)
	if envURL := os.Getenv("UNIROUTE_TUNNEL_URL"); envURL != "" {
		return envURL
	}
	
	// Priority 2: Auto-detect local mode
	if isLocalMode() {
		// Check which port is available (prefer 8055, fallback to 8080)
		client := &http.Client{Timeout: 1 * time.Second}
		if resp, err := client.Get("http://localhost:8055/health"); err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return "localhost:8055"
			}
		}
		// Fallback to 8080 if 8055 is not available
		return "localhost:8080"
	}
	
	// Priority 3: Default (only used if nothing else is configured)
	return "tunnel.uniroute.co"
}

func runAuthLogin(cmd *cobra.Command, args []string) error {
	// Check if user is already logged in
	if isAuthenticated() {
		config, err := loadAuthConfig()
		if err == nil && config != nil {
			// Verify token is still valid by checking expiration
			tokenValid := true
			if config.ExpiresAt != "" {
				expiresAt, err := time.Parse(time.RFC3339, config.ExpiresAt)
				if err == nil && time.Now().After(expiresAt) {
					tokenValid = false
					// Clear expired token
					_ = runAuthLogout(nil, nil)
				}
			}
			
			if tokenValid {
				fmt.Println("✅ You are already logged in!")
				fmt.Printf("   Email: %s\n", config.Email)
				fmt.Printf("   Server: %s\n", config.ServerURL)
				if config.ExpiresAt != "" {
					expiresAt, err := time.Parse(time.RFC3339, config.ExpiresAt)
					if err == nil {
						timeUntilExpiry := time.Until(expiresAt)
						days := int(timeUntilExpiry.Hours() / 24)
						if days > 0 {
							fmt.Printf("   Expires: %d days remaining\n", days)
						} else {
							hours := int(timeUntilExpiry.Hours())
							fmt.Printf("   Expires: %d hours remaining\n", hours)
						}
					}
				} else {
					fmt.Printf("   Status: ✅ Active (no expiration)\n")
				}
				fmt.Println()
				fmt.Println("   To log in with a different account, run 'uniroute auth logout' first")
				return nil
			}
		}
	}

	// Get email if not provided
	if authEmail == "" {
		fmt.Print("Email: ")
		fmt.Scanln(&authEmail)
		if authEmail == "" {
			return fmt.Errorf("email is required")
		}
	}

	// Get password if not provided
	if authPassword == "" {
		fmt.Print("Password: ")
		// Read password without echoing to terminal (cross-platform)
		passwordBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
		if err != nil {
			return fmt.Errorf("failed to read password: %w", err)
		}
		fmt.Println() // Print newline after password input
		authPassword = string(passwordBytes)
		if authPassword == "" {
			return fmt.Errorf("password is required")
		}
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Build login request
	body := map[string]interface{}{
		"email":    authEmail,
		"password": authPassword,
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	// Determine server URL: flag > env var > saved config > auto-detect > default
	serverURL := authServer
	if serverURL == "" {
		// No flag provided, use getServerURL() which checks env var, config, and auto-detects
		serverURL = getServerURL()
	}
	
	// Create request
	loginURL := fmt.Sprintf("%s/auth/login", serverURL)
	req, err := http.NewRequest("POST", loginURL, strings.NewReader(string(jsonBody)))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to server: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("login failed: %s", string(respBody))
	}

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	token, ok := result["token"].(string)
	if !ok {
		return fmt.Errorf("invalid response: token not found")
	}

	// Save config
	config := &AuthConfig{
		Token:     token,
		Email:     authEmail,
		ServerURL: serverURL,
	}

	if expiresAt, ok := result["expires_at"].(string); ok {
		config.ExpiresAt = expiresAt
	}

	if err := saveAuthConfig(config); err != nil {
		return fmt.Errorf("failed to save credentials: %w", err)
	}

	fmt.Println("✅ Successfully logged in!")
	fmt.Printf("   Email: %s\n", authEmail)
	
	// Show helpful message about environment variables if using default
	if os.Getenv("UNIROUTE_API_URL") == "" && serverURL == "https://api.uniroute.co" {
		fmt.Printf("   💡 Tip: Set UNIROUTE_API_URL env var to avoid hardcoded defaults\n")
	}
	
	fmt.Printf("   Run 'uniroute auth logout' to log out\n")

	return nil
}

func runAuthLogout(cmd *cobra.Command, args []string) error {
	configPath, err := getConfigPath()
	if err != nil {
		return err
	}

	if err := os.Remove(configPath); err != nil {
		if os.IsNotExist(err) {
			fmt.Println("✅ Already logged out")
			return nil
		}
		return fmt.Errorf("failed to remove config: %w", err)
	}

	fmt.Println("✅ Successfully logged out")
	return nil
}

func runAuthStatus(cmd *cobra.Command, args []string) error {
	config, err := loadAuthConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if config == nil {
		fmt.Println("❌ Not logged in")
		fmt.Println("   Run 'uniroute auth login' to authenticate")
		return nil
	}

	// Check if token is still valid
	token := getAuthToken()
	if token == "" {
		fmt.Println("❌ Session expired")
		fmt.Println("   Run 'uniroute auth login' to authenticate again")
		return nil
	}

	fmt.Println("✅ Logged in (session persists across restarts)")
	fmt.Printf("   Email: %s\n", config.Email)
	fmt.Printf("   Server: %s\n", config.ServerURL)
	if config.ExpiresAt != "" {
		expiresAt, err := time.Parse(time.RFC3339, config.ExpiresAt)
		if err == nil {
			if time.Now().After(expiresAt) {
				fmt.Printf("   Status: ⚠️  Token expired\n")
			} else {
				timeUntilExpiry := time.Until(expiresAt)
				days := int(timeUntilExpiry.Hours() / 24)
				if days > 0 {
					fmt.Printf("   Expires: %s (%d days remaining)\n", config.ExpiresAt, days)
				} else {
					hours := int(timeUntilExpiry.Hours())
					fmt.Printf("   Expires: %s (%d hours remaining)\n", config.ExpiresAt, hours)
				}
			}
		} else {
			fmt.Printf("   Expires: %s\n", config.ExpiresAt)
		}
	} else {
		fmt.Printf("   Status: ✅ Active (no expiration)\n")
	}

	return nil
}
