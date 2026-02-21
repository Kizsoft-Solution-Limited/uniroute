package commands

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check the status of the UniRoute gateway server",
	Long: `Check if the UniRoute gateway server is running and healthy.

Example:
  uniroute status
  uniroute status --url http://localhost:8084`,
	RunE: runStatus,
}

var statusURL string

func init() {
	statusCmd.Flags().StringVarP(&statusURL, "url", "u", "", "Gateway server URL (default: public UniRoute server)")
}

func runStatus(cmd *cobra.Command, args []string) error {
	serverURL := statusURL
	if serverURL == "" {
		serverURL = getServerURL()
	}

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	healthURL := fmt.Sprintf("%s/health", serverURL)
	resp, err := client.Get(healthURL)
	if err != nil {
		fmt.Printf("❌ Server is not responding at %s\n", serverURL)
		fmt.Printf("   Error: %v\n", err)
		return fmt.Errorf("server unreachable")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("❌ Server returned status %d\n", resp.StatusCode)
		return fmt.Errorf("server unhealthy")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	var health map[string]interface{}
	if err := json.Unmarshal(body, &health); err != nil {
		fmt.Printf("✅ Server is running at %s\n", serverURL)
		fmt.Printf("   Response: %s\n", string(body))
		return nil
	}

	fmt.Printf("✅ Server is running at %s\n", serverURL)
	fmt.Printf("   Status: %v\n", health["status"])

	providersURL := fmt.Sprintf("%s/v1/providers", serverURL)
	providersResp, err := client.Get(providersURL)
	if err == nil && providersResp.StatusCode == http.StatusOK {
		defer providersResp.Body.Close()
		providersBody, _ := io.ReadAll(providersResp.Body)
		var providers []string
		if json.Unmarshal(providersBody, &providers) == nil {
			fmt.Printf("   Providers: %v\n", providers)
		}
	}

	return nil
}

