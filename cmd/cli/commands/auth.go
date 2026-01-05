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

Commands:
  login    Login to your UniRoute account
  logout   Logout and clear saved credentials
  status   Show current authentication status`,
}

var authLoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login to UniRoute",
	Long: `Login to your UniRoute account using email and password.

Example:
  uniroute auth login
  uniroute auth login --email user@example.com
  uniroute auth login --server https://api.uniroute.dev`,
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
	authLoginCmd.Flags().StringVarP(&authServer, "server", "s", "https://api.uniroute.dev", "UniRoute server URL")
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
func getAuthToken() string {
	config, err := loadAuthConfig()
	if err != nil || config == nil {
		return ""
	}
	return config.Token
}

// getServerURL returns the configured server URL
func getServerURL() string {
	config, err := loadAuthConfig()
	if err != nil || config == nil {
		return "https://api.uniroute.dev"
	}
	if config.ServerURL != "" {
		return config.ServerURL
	}
	return "https://api.uniroute.dev"
}

func runAuthLogin(cmd *cobra.Command, args []string) error {
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

	// Create request
	loginURL := fmt.Sprintf("%s/auth/login", authServer)
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
		ServerURL: authServer,
	}

	if expiresAt, ok := result["expires_at"].(string); ok {
		config.ExpiresAt = expiresAt
	}

	if err := saveAuthConfig(config); err != nil {
		return fmt.Errorf("failed to save credentials: %w", err)
	}

	fmt.Println("✅ Successfully logged in!")
	fmt.Printf("   Email: %s\n", authEmail)
	fmt.Printf("   Server: %s\n", authServer)

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

	fmt.Println("✅ Logged in")
	fmt.Printf("   Email: %s\n", config.Email)
	fmt.Printf("   Server: %s\n", config.ServerURL)
	if config.ExpiresAt != "" {
		fmt.Printf("   Expires: %s\n", config.ExpiresAt)
	}

	return nil
}
