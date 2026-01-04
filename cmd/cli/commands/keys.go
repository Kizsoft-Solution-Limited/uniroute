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

var keysCmd = &cobra.Command{
	Use:   "keys",
	Short: "Manage API keys",
	Long: `Manage API keys for the UniRoute gateway.

Commands:
  create  Create a new API key
  list    List all API keys (requires JWT)
  revoke  Revoke an API key (requires JWT)`,
}

var keysCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new API key",
	Long: `Create a new API key for accessing the UniRoute gateway.

Example:
  uniroute keys create
  uniroute keys create --name "My API Key"
  uniroute keys create --url http://localhost:8084 --jwt-token YOUR_JWT`,
	RunE: runKeysCreate,
}

var (
	keysURL      string
	keysName     string
	keysJWTToken string
)

func init() {
	keysCmd.AddCommand(keysCreateCmd)

	keysCreateCmd.Flags().StringVarP(&keysURL, "url", "u", "", "Gateway server URL (default: public UniRoute server)")
	keysCreateCmd.Flags().StringVarP(&keysName, "name", "n", "", "Name for the API key")
	keysCreateCmd.Flags().StringVarP(&keysJWTToken, "jwt-token", "t", "", "JWT token for authentication (required for database-backed keys)")
}

func runKeysCreate(cmd *cobra.Command, args []string) error {
	// Use public server by default
	serverURL := keysURL
	if serverURL == "" {
		serverURL = getServerURL()
	}

	// Get auth token (from login or JWT flag)
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

	// Build request body
	body := map[string]interface{}{}
	if keysName != "" {
		body["name"] = keysName
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create request
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/admin/api-keys", serverURL), strings.NewReader(string(jsonBody)))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

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

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("server returned error: %s", string(respBody))
	}

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		// Not JSON, print raw
		fmt.Println("API Key created:")
		fmt.Println(string(respBody))
		return nil
	}

	fmt.Println("✅ API Key created successfully!")
	if apiKey, ok := result["api_key"].(string); ok {
		fmt.Printf("\nAPI Key: %s\n", apiKey)
		fmt.Println("\n⚠️  Save this key securely - it won't be shown again!")
	}
	if id, ok := result["id"].(string); ok {
		fmt.Printf("ID: %s\n", id)
	}

	return nil
}
