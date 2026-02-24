package commands

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

const dateOnlyLayout = "2006-01-02"

var keysCmd = &cobra.Command{
	Use:   "keys",
	Short: "Manage API keys",
	Long: `Manage API keys for the UniRoute gateway.

Commands:
  create  Create a new API key
  list    List all API keys (requires authentication)
  revoke  Revoke (disable) an API key - key stops working but stays in list
  delete  Delete (remove) an API key permanently`,
}

var keysCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new API key",
	Long: `Create a new API key for accessing the UniRoute gateway.

Example:
  uniroute keys create
  uniroute keys create --name "My API Key"
  uniroute keys create --name "Long-lived"   (no expiration = never expires)
  uniroute keys create --expires-at 2030-12-31
  uniroute keys create --url http://localhost:8084 --jwt-token YOUR_JWT`,
	RunE: runKeysCreate,
}

var keysListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all API keys",
	Long: `List all API keys for your account.

Example:
  uniroute keys list
  uniroute keys list --url http://localhost:8084`,
	RunE: runKeysList,
}

var keysRevokeCmd = &cobra.Command{
	Use:   "revoke",
	Short: "Revoke (disable) an API key",
	Long: `Revoke (disable) an API key by ID. The key stops working but remains in the list as inactive.

Example:
  uniroute keys revoke <key-id>
  uniroute keys revoke <key-id> --url http://localhost:8084`,
	Args:  cobra.ExactArgs(1),
	RunE:  runKeysRevoke,
}

var keysDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete (remove) an API key permanently",
	Long: `Delete (remove) an API key permanently by ID. The key is removed from the list; cannot be undone.

Example:
  uniroute keys delete <key-id>
  uniroute keys delete <key-id> --url http://localhost:8084`,
	Args:  cobra.ExactArgs(1),
	RunE:  runKeysDelete,
}

var (
	keysURL             string
	keysName            string
	keysJWTToken        string
	keysExpiresAt       string
	keysRateLimitMin    int
	keysRateLimitDay    int
)

func init() {
	keysCmd.AddCommand(keysCreateCmd)
	keysCmd.AddCommand(keysListCmd)
	keysCmd.AddCommand(keysRevokeCmd)
	keysCmd.AddCommand(keysDeleteCmd)

	keysCreateCmd.Flags().StringVarP(&keysURL, "url", "u", "", "Gateway server URL (default: public UniRoute server)")
	keysCreateCmd.Flags().StringVarP(&keysName, "name", "n", "", "Name for the API key")
	keysCreateCmd.Flags().StringVarP(&keysJWTToken, "jwt-token", "t", "", "JWT token for authentication (required for database-backed keys)")
	keysCreateCmd.Flags().StringVar(&keysExpiresAt, "expires-at", "", "Expiration date (YYYY-MM-DD); omit for no expiration (key lasts until revoked)")
	keysCreateCmd.Flags().IntVar(&keysRateLimitMin, "rate-limit-minute", 0, "Rate limit per minute (default 60)")
	keysCreateCmd.Flags().IntVar(&keysRateLimitDay, "rate-limit-day", 0, "Rate limit per day (default 10000)")

	keysListCmd.Flags().StringVarP(&keysURL, "url", "u", "", "Gateway server URL (default: public UniRoute server)")
	keysListCmd.Flags().StringVarP(&keysJWTToken, "jwt-token", "t", "", "JWT token for authentication")

	keysRevokeCmd.Flags().StringVarP(&keysURL, "url", "u", "", "Gateway server URL (default: public UniRoute server)")
	keysRevokeCmd.Flags().StringVarP(&keysJWTToken, "jwt-token", "t", "", "JWT token for authentication")

	keysDeleteCmd.Flags().StringVarP(&keysURL, "url", "u", "", "Gateway server URL (default: public UniRoute server)")
	keysDeleteCmd.Flags().StringVarP(&keysJWTToken, "jwt-token", "t", "", "JWT token for authentication")
}

func runKeysCreate(cmd *cobra.Command, args []string) error {
	serverURL := keysURL
	if serverURL == "" {
		serverURL = getServerURL()
	}

	token := keysJWTToken
	if token == "" {
		token = getAuthToken()
		if token == "" {
			return fmt.Errorf("not authenticated. Run 'uniroute auth login' or use --jwt-token")
		}
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	body := map[string]interface{}{}
	if keysName != "" {
		body["name"] = keysName
	}
	if keysRateLimitMin > 0 {
		body["rate_limit_per_minute"] = keysRateLimitMin
	}
	if keysRateLimitDay > 0 {
		body["rate_limit_per_day"] = keysRateLimitDay
	}
	if keysExpiresAt != "" {
		var expiresAt time.Time
		if t, err := time.Parse(time.RFC3339, keysExpiresAt); err == nil {
			expiresAt = t
		} else if t, err := time.Parse(dateOnlyLayout, keysExpiresAt); err == nil {
			expiresAt = time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 0, time.UTC)
		} else {
			return fmt.Errorf("invalid --expires-at: use YYYY-MM-DD or RFC3339")
		}
		body["expires_at"] = expiresAt.Format(time.RFC3339)
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/auth/api-keys", serverURL), strings.NewReader(string(jsonBody)))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to server: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("server returned error: %s", string(respBody))
	}

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		fmt.Println("API Key created:")
		fmt.Println(string(respBody))
		return nil
	}

	fmt.Println("✅ API Key created successfully!")
	if apiKey, ok := result["api_key"].(string); ok {
		fmt.Printf("\nAPI Key: %s\n", apiKey)
		fmt.Println("\n⚠️  Save this key securely - it won't be shown again!")
	} else if key, ok := result["key"].(string); ok {
		fmt.Printf("\nAPI Key: %s\n", key)
		fmt.Println("\n⚠️  Save this key securely - it won't be shown again!")
	}
	if id, ok := result["id"].(string); ok {
		fmt.Printf("ID: %s\n", id)
	}

	return nil
}

func runKeysList(cmd *cobra.Command, args []string) error {
	serverURL := keysURL
	if serverURL == "" {
		serverURL = getServerURL()
	}

	token := keysJWTToken
	if token == "" {
		token = getAuthToken()
		if token == "" {
			return fmt.Errorf("not authenticated. Run 'uniroute auth login' or use --jwt-token")
		}
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/auth/api-keys", serverURL), nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

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
		return fmt.Errorf("server returned error: %s", string(respBody))
	}

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		fmt.Println("API Keys:")
		fmt.Println(string(respBody))
		return nil
	}

	keys, ok := result["keys"].([]interface{})
	if !ok {
		fmt.Println("No API keys found")
		return nil
	}

	if len(keys) == 0 {
		fmt.Println("No API keys found. Create one with 'uniroute keys create'")
		return nil
	}

	fmt.Println("Your API Keys:")
	fmt.Println(strings.Repeat("-", 80))
	for i, key := range keys {
		keyMap, ok := key.(map[string]interface{})
		if !ok {
			continue
		}

		fmt.Printf("\n%d. ", i+1)
		if name, ok := keyMap["name"].(string); ok && name != "" {
			fmt.Printf("%s\n", name)
		} else {
			fmt.Printf("(Unnamed)\n")
		}

		if id, ok := keyMap["id"].(string); ok {
			fmt.Printf("   ID: %s\n", id)
		}
		if createdAt, ok := keyMap["created_at"].(string); ok {
			fmt.Printf("   Created: %s\n", createdAt)
		}
		if expiresAt, ok := keyMap["expires_at"].(string); ok && expiresAt != "" {
			fmt.Printf("   Expires: %s\n", expiresAt)
		} else {
			fmt.Printf("   Expires: Never\n")
		}
		if isActive, ok := keyMap["is_active"].(bool); ok {
			if isActive {
				fmt.Printf("   Status: Active\n")
			} else {
				fmt.Printf("   Status: Inactive\n")
			}
		}
		if rateLimitPerMinute, ok := keyMap["rate_limit_per_minute"].(float64); ok {
			fmt.Printf("   Rate Limit: %.0f/min", rateLimitPerMinute)
			if rateLimitPerDay, ok := keyMap["rate_limit_per_day"].(float64); ok {
				fmt.Printf(", %.0f/day\n", rateLimitPerDay)
			} else {
				fmt.Println()
			}
		}
	}
	fmt.Println()

	return nil
}

func runKeysRevoke(cmd *cobra.Command, args []string) error {
	keyID := args[0]

	serverURL := keysURL
	if serverURL == "" {
		serverURL = getServerURL()
	}

	token := keysJWTToken
	if token == "" {
		token = getAuthToken()
		if token == "" {
			return fmt.Errorf("not authenticated. Run 'uniroute auth login' or use --jwt-token")
		}
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/auth/api-keys/%s", serverURL, keyID), nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to server: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("server returned error: %s", string(respBody))
	}

	fmt.Printf("✅ API key %s revoked successfully\n", keyID)
	return nil
}

func runKeysDelete(cmd *cobra.Command, args []string) error {
	keyID := args[0]

	serverURL := keysURL
	if serverURL == "" {
		serverURL = getServerURL()
	}

	token := keysJWTToken
	if token == "" {
		token = getAuthToken()
		if token == "" {
			return fmt.Errorf("not authenticated. Run 'uniroute auth login' or use --jwt-token")
		}
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/auth/api-keys/%s/permanent", serverURL, keyID), nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to server: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("server returned error: %s", string(respBody))
	}

	fmt.Printf("✅ API key %s deleted permanently\n", keyID)
	return nil
}
