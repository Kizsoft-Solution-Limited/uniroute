package commands

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the UniRoute gateway server",
	Long: `Start the UniRoute gateway server on the configured port.

The server will:
- Start on the configured port (default: 8084)
- Register available LLM providers
- Be ready to accept requests

Example:
  uniroute start
  uniroute start --port 8080
  uniroute start --config .env.production`,
	RunE: runStart,
}

var (
	startPort     string
	startConfig   string
	startDetached bool
)

func init() {
	startCmd.Flags().StringVarP(&startPort, "port", "p", "", "Port to run the server on (overrides config)")
	startCmd.Flags().StringVarP(&startConfig, "config", "c", "", "Path to config file (.env)")
	startCmd.Flags().BoolVarP(&startDetached, "detached", "d", false, "Run in detached mode (background)")
}

func runStart(cmd *cobra.Command, args []string) error {
	// Determine port to use
	port := startPort
	if port == "" {
		port = "8084" // default
	}

	// Check if port is already in use
	if isPortInUse(port) {
		fmt.Printf("⚠️  Port %s is already in use.\n", port)
		fmt.Println()
		fmt.Println("Options:")
		fmt.Printf("  1. Use a different port: uniroute start --port 8085\n")
		fmt.Println("  2. Stop the existing server first")
		fmt.Println("  3. Check what's using the port:")
		fmt.Printf("     lsof -i :%s\n", port)
		fmt.Println()
		return fmt.Errorf("port %s is already in use", port)
	}

	// Find the gateway binary
	executable, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to find executable: %w", err)
	}

	// Get the directory of the CLI binary
	cliDir := filepath.Dir(executable)
	
	// Look for gateway binary in same directory or ../bin
	gatewayBinary := filepath.Join(cliDir, "uniroute-gateway")
	if _, err := os.Stat(gatewayBinary); os.IsNotExist(err) {
		// Try bin directory
		gatewayBinary = filepath.Join(filepath.Dir(cliDir), "bin", "uniroute-gateway")
		if _, err := os.Stat(gatewayBinary); os.IsNotExist(err) {
			// Try current directory
			gatewayBinary = "./bin/uniroute-gateway"
			if _, err := os.Stat(gatewayBinary); os.IsNotExist(err) {
				return fmt.Errorf("gateway binary not found. Please build it first: make build")
			}
		}
	}

	// Build command
	gatewayCmd := exec.Command(gatewayBinary)

	// Set environment variables
	if startPort != "" {
		gatewayCmd.Env = append(os.Environ(), fmt.Sprintf("PORT=%s", startPort))
	}
	if startConfig != "" {
		gatewayCmd.Env = append(gatewayCmd.Env, fmt.Sprintf("ENV_FILE=%s", startConfig))
	}

	// Set output
	gatewayCmd.Stdout = os.Stdout
	gatewayCmd.Stderr = os.Stderr
	gatewayCmd.Stdin = os.Stdin

	if startDetached {
		// Run in background
		gatewayCmd.Stdout = nil
		gatewayCmd.Stderr = nil
		gatewayCmd.Stdin = nil
		if err := gatewayCmd.Start(); err != nil {
			return fmt.Errorf("failed to start gateway: %w", err)
		}
		fmt.Printf("Gateway started in background (PID: %d)\n", gatewayCmd.Process.Pid)
		return nil
	}

	// Run in foreground
	fmt.Println("Starting UniRoute Gateway...")
	if startPort != "" {
		fmt.Printf("Port: %s\n", startPort)
	}
	fmt.Println("Press Ctrl+C to stop")
	fmt.Println()
	
	if err := gatewayCmd.Run(); err != nil {
		// Check if it's a port error
		if strings.Contains(err.Error(), "address already in use") || strings.Contains(err.Error(), "bind") {
			fmt.Println()
			fmt.Printf("❌ Port %s is already in use.\n", port)
			fmt.Println()
			fmt.Println("Try:")
			fmt.Printf("  uniroute start --port 8085\n")
			return fmt.Errorf("port conflict")
		}
		return fmt.Errorf("gateway exited with error: %w", err)
	}

	return nil
}

// isPortInUse checks if a port is already in use
func isPortInUse(port string) bool {
	timeout := time.Second
	conn, err := net.DialTimeout("tcp", net.JoinHostPort("localhost", port), timeout)
	if err != nil {
		return false
	}
	if conn != nil {
		conn.Close()
		return true
	}
	return false
}
