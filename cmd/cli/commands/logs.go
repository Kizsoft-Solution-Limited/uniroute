package commands

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/spf13/cobra"
)

var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "View live logs from the gateway server",
	Long: `View live logs from the UniRoute gateway server.

This command connects to the server and streams logs in real-time.

Example:
  uniroute logs
  uniroute logs --url http://localhost:8084
  uniroute logs --follow`,
	RunE: runLogs,
}

var (
	logsURL    string
	logsFollow bool
)

func init() {
	logsCmd.Flags().StringVarP(&logsURL, "url", "u", "http://localhost:8084", "Gateway server URL")
	logsCmd.Flags().BoolVarP(&logsFollow, "follow", "f", true, "Follow log output (default: true)")
}

func runLogs(cmd *cobra.Command, args []string) error {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	healthURL := fmt.Sprintf("%s/health", logsURL)
	resp, err := client.Get(healthURL)
	if err != nil {
		return fmt.Errorf("server is not running at %s: %w", logsURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned status %d", resp.StatusCode)
	}

	fmt.Printf("📋 Server is running at %s\n", logsURL)
	fmt.Println()
	fmt.Println("Note: Live log streaming will be available in a future update.")
	fmt.Println("For now, check server logs directly or use:")
	fmt.Println("  - Docker: docker logs <container>")
	fmt.Println("  - Systemd: journalctl -u uniroute -f")
	fmt.Println("  - Direct: Check the terminal where you ran 'uniroute start'")

	body, _ := io.ReadAll(resp.Body)
	if len(body) > 0 {
		fmt.Printf("\nServer health: %s\n", string(body))
	}

	return nil
}

