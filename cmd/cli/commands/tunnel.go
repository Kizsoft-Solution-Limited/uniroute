package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"os/exec"

	"github.com/Kizsoft-Solution-Limited/uniroute/internal/tunnel"
	"github.com/Kizsoft-Solution-Limited/uniroute/pkg/color"
	"github.com/Kizsoft-Solution-Limited/uniroute/pkg/logger"
	versioncheck "github.com/Kizsoft-Solution-Limited/uniroute/pkg/version"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var tunnelCmd = &cobra.Command{
	Use:   "tunnel",
	Short: "Expose local server to internet",
	Long: `Expose your local application to the internet with a public URL.

This command uses the built-in UniRoute tunnel server to create a secure tunnel
to your local server. The tunnel automatically resumes your previous subdomain
if available.

Examples:
  uniroute tunnel                    # Create tunnel (auto-resumes if available)
  uniroute tunnel --port 8084        # Tunnel specific port
  uniroute tunnel --resume abc123    # Resume specific subdomain
  uniroute tunnel --list             # List saved tunnel state
  uniroute tunnel --clear            # Clear saved tunnel state`,
	RunE: runTunnel,
}

var (
	tunnelPort      string
	tunnelServerURL string
	resumeSubdomain string
	clearSaved      bool
	listSaved       bool
)

func init() {
	tunnelCmd.Flags().StringVarP(&tunnelPort, "port", "p", "8084", "Local port to tunnel")
	tunnelCmd.Flags().StringVarP(&tunnelServerURL, "server", "s", "tunnel.uniroute.dev", "Tunnel server URL (default: public UniRoute server)")
	tunnelCmd.Flags().StringVar(&resumeSubdomain, "resume", "", "Resume a specific subdomain (or use saved one if not specified)")
	tunnelCmd.Flags().BoolVar(&clearSaved, "clear", false, "Clear saved tunnel state")
	tunnelCmd.Flags().BoolVar(&listSaved, "list", false, "List saved tunnel state")
}

func runTunnel(cmd *cobra.Command, args []string) error {
	log := logger.New()
	persistence := tunnel.NewTunnelPersistence(log)

	// Handle list/clear commands
	if listSaved {
		state, err := persistence.Load()
		if err != nil || state == nil {
			fmt.Println(color.Gray("No saved tunnel state found"))
			return nil
		}
		fmt.Println(color.Cyan("📋 Saved Tunnel State:"))
		fmt.Printf("   %s %s\n", color.Bold("Subdomain:"), color.Cyan(state.Subdomain))
		fmt.Printf("   %s %s\n", color.Bold("Public URL:"), color.Cyan(state.PublicURL))
		fmt.Printf("   %s %s\n", color.Bold("Local URL:"), color.Gray(state.LocalURL))
		fmt.Printf("   %s %s\n", color.Bold("Server URL:"), color.Gray(state.ServerURL))
		fmt.Printf("   %s %s\n", color.Bold("Last Used:"), color.Gray(state.LastUsed.Format("2006-01-02 15:04:05")))
		return nil
	}

	if clearSaved {
		if err := persistence.Clear(); err != nil {
			return fmt.Errorf("failed to clear saved state: %w", err)
		}
		fmt.Println(color.Green("✓ Cleared saved tunnel state"))
		return nil
	}

	// Always use built-in tunnel (default behavior)
	return runBuiltInTunnel(cmd, args)
}

// runBuiltInTunnel uses the built-in tunnel client
func runBuiltInTunnel(cmd *cobra.Command, args []string) error {
	// Check if using public server - require authentication
	// Allow localhost for development/testing without auth
	if tunnelServerURL != "localhost:8080" &&
		tunnelServerURL != "http://localhost:8080" &&
		(tunnelServerURL == "tunnel.uniroute.dev" ||
			tunnelServerURL == "https://tunnel.uniroute.dev" ||
			strings.Contains(tunnelServerURL, ".uniroute.dev")) {
		token := getAuthToken()
		if token == "" {
			return fmt.Errorf("authentication required for public tunnel server\nRun 'uniroute auth login' first")
		}
	}

	log := logger.New()

	localURL := fmt.Sprintf("http://localhost:%s", tunnelPort)

	fmt.Println()
	fmt.Println(color.Cyan("Starting UniRoute Tunnel..."))
	fmt.Println()
	fmt.Printf("   %s %s\n", color.Gray("Local URL:"), color.Bold(localURL))
	fmt.Printf("   %s %s\n", color.Gray("Tunnel Server:"), color.Bold(tunnelServerURL))
	fmt.Println()
	fmt.Println(color.Yellow("🔌 Connecting to tunnel server..."))
	fmt.Println()

	// Create tunnel client
	client := tunnel.NewTunnelClient(tunnelServerURL, localURL, log)

	// Connect to tunnel server
	if err := client.Connect(); err != nil {
		return fmt.Errorf("failed to connect to tunnel server: %w", err)
	}

	// Get tunnel info
	info := client.GetTunnelInfo()
	if info == nil {
		return fmt.Errorf("failed to get tunnel information")
	}

	// Get random funny quote
	funnyQuote := tunnel.GetRandomQuote()

	// Get initial latency (will be 0 until first heartbeat)
	initialLatency := client.GetLatency()
	latencyStr := "0ms"
	if initialLatency > 0 {
		latencyStr = fmt.Sprintf("%dms", initialLatency)
	}

	// Get user email from auth config
	accountDisplay := "Free"
	homeDir, err := os.UserHomeDir()
	if err == nil {
		configPath := filepath.Join(homeDir, ".uniroute", "auth.json")
		if data, err := os.ReadFile(configPath); err == nil {
			var authConfig struct {
				Email string `json:"email"`
			}
			if err := json.Unmarshal(data, &authConfig); err == nil && authConfig.Email != "" {
				accountDisplay = authConfig.Email + " (Plan: Free)"
			}
		}
	}

	// ngrok-style output
	fmt.Println()
	fmt.Println(color.Green("Tunnel Connected Successfully!"))
	fmt.Println()

	// Session Status with dynamic emoji
	fmt.Printf("Session Status                %s %s\n", color.Green("●"), color.Green("online"))

	fmt.Printf("Account                       %s\n", color.Gray(accountDisplay))
	currentVersion := GetVersion()
	fmt.Printf("Version                       %s", color.Gray(currentVersion))

	// Check for updates in background (non-blocking)
	var updateInfo *versioncheck.VersionInfo
	var updateInfoMu sync.Mutex
	go func() {
		versionURL := os.Getenv("UNIROUTE_VERSION_URL")
		if versionURL == "" {
			versionURL = "https://api.github.com/repos/Kizsoft-Solution-Limited/uniroute/releases/latest"
		}
		checker := versioncheck.NewChecker(versionURL)
		info, _ := checker.CheckForUpdate(currentVersion)
		if info != nil && info.UpdateAvailable {
			updateInfoMu.Lock()
			updateInfo = info
			updateInfoMu.Unlock()
			// Display update notification (will be shown after version line)
		}
	}()

	fmt.Println()
	fmt.Printf("Region                        %s\n", color.Gray("Local"))
	fmt.Printf("Latency                       %s\n", color.Gray(latencyStr))
	fmt.Printf("Web Interface                 %s\n", color.Cyan("http://127.0.0.1:4040"))
	fmt.Println()
	fmt.Printf("Connections                   ttl     opn     rt1     rt5     p50     p90\n")

	// Initial connection stats (will be updated)
	initialStats := &tunnel.ConnectionStats{
		Total: 0,
		Open:  1,
		RT1:   0.00,
		RT5:   0.00,
		P50:   0.00,
		P90:   0.00,
	}
	fmt.Printf("                              %s      %s      %s      %s      %s      %s\n",
		color.Gray(fmt.Sprintf("%d", initialStats.Total)),
		color.Gray(fmt.Sprintf("%d", initialStats.Open)),
		color.Gray(fmt.Sprintf("%.2f", initialStats.RT1)),
		color.Gray(fmt.Sprintf("%.2f", initialStats.RT5)),
		color.Gray(fmt.Sprintf("%.2f", initialStats.P50)),
		color.Gray(fmt.Sprintf("%.2f", initialStats.P90)))
	fmt.Println()
	fmt.Printf("🌍 Public URL:\n")
	fmt.Printf("   %s\n", color.Cyan(info.PublicURL))
	fmt.Println()
	fmt.Printf("🔗 Forwarding:\n")
	fmt.Printf("   %s %s %s\n",
		color.Cyan(info.PublicURL),
		color.Gray("->"),
		color.Bold(localURL))
	fmt.Println()
	fmt.Println(color.Yellow(fmt.Sprintf("💬 %s", funnyQuote)))
	fmt.Println()
	fmt.Println(color.Gray("Press Ctrl+C to stop"))
	updateInfoMu.Lock()
	hasUpdate := updateInfo != nil && updateInfo.UpdateAvailable
	updateInfoMu.Unlock()
	if hasUpdate {
		fmt.Println(color.Gray("Press Ctrl+U to upgrade"))
	}
	fmt.Println()

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Track if we should stop printing updates
	var stopMu sync.Mutex
	stopUpdates := false

	// Monitor connection status, latency, and stats, update display dynamically
	go func() {
		lastStatus := true
		lastLatency := int64(-1)
		lastStats := initialStats
		abs := func(x int64) int64 {
			if x < 0 {
				return -x
			}
			return x
		}

		ticker := time.NewTicker(3 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				// Stop immediately when context is cancelled
				stopMu.Lock()
				stopUpdates = true
				stopMu.Unlock()
				return
			case <-ticker.C:
				// Check if we should stop
				stopMu.Lock()
				shouldStop := stopUpdates
				stopMu.Unlock()
				if shouldStop {
					return
				}

				// Check if context is cancelled before processing
				select {
				case <-ctx.Done():
					stopMu.Lock()
					stopUpdates = true
					stopMu.Unlock()
					return
				default:
				}

				isConnected := client.IsConnected()
				currentLatency := client.GetLatency()

				// Fetch connection stats from server if connected
				var currentStats *tunnel.ConnectionStats
				if isConnected && info != nil {
					if stats, err := client.GetConnectionStats(tunnelServerURL, info.ID); err == nil {
						currentStats = stats
					} else {
						// If fetch fails, keep last stats
						currentStats = lastStats
					}
				} else {
					currentStats = lastStats
				}

				// Update latency display if it changed significantly (use \r to overwrite same line)
				if lastLatency == -1 || (currentLatency != lastLatency && abs(currentLatency-lastLatency) > 5) {
					// Check if we should stop before printing
					stopMu.Lock()
					shouldStop := stopUpdates
					stopMu.Unlock()
					if shouldStop {
						return
					}

					// Check if context is cancelled before printing
					select {
					case <-ctx.Done():
						stopMu.Lock()
						stopUpdates = true
						stopMu.Unlock()
						return
					default:
						latencyStr := "0ms"
						if currentLatency > 0 {
							latencyStr = fmt.Sprintf("%dms", currentLatency)
						}
						// Update latency on the same line without creating new lines
						// Move cursor up 4 lines to the latency line, update it, then move back
						fmt.Print("\033[4A")  // Move up 4 lines to latency line
						fmt.Print("\r\033[K") // Clear to end of line
						fmt.Printf("Latency                       %s", color.Gray(latencyStr))
						fmt.Print("\033[4B") // Move back down 4 lines
						lastLatency = currentLatency
					}
				}

				// Update connection stats if they changed (overwrite the connections line)
				if currentStats != nil && (currentStats.Total != lastStats.Total ||
					currentStats.RT1 != lastStats.RT1 || currentStats.RT5 != lastStats.RT5 ||
					currentStats.P50 != lastStats.P50 || currentStats.P90 != lastStats.P90) {
					// Check if we should stop before printing
					stopMu.Lock()
					shouldStop := stopUpdates
					stopMu.Unlock()
					if shouldStop {
						return
					}

					// Check context before printing
					select {
					case <-ctx.Done():
						stopMu.Lock()
						stopUpdates = true
						stopMu.Unlock()
						return
					default:
						// Move cursor up one line, clear it, and update
						fmt.Printf("\r\033[K                              %s      %s      %s      %s      %s      %s",
							color.Gray(fmt.Sprintf("%d", currentStats.Total)),
							color.Gray(fmt.Sprintf("%d", currentStats.Open)),
							color.Gray(fmt.Sprintf("%.2f", currentStats.RT1)),
							color.Gray(fmt.Sprintf("%.2f", currentStats.RT5)),
							color.Gray(fmt.Sprintf("%.2f", currentStats.P50)),
							color.Gray(fmt.Sprintf("%.2f", currentStats.P90)))
						lastStats = currentStats
					}
				}

				if isConnected != lastStatus {
					// Check context before printing status changes
					select {
					case <-ctx.Done():
						return
					default:
						if !isConnected {
							// Connection lost
							fmt.Println()
							fmt.Printf("Session Status                %s %s\n", color.Red("●"), color.Red("offline"))
							fmt.Println(color.Yellow("⚠️  Connection lost, attempting to reconnect..."))
							fmt.Println()
						} else {
							// Reconnected
							fmt.Println()
							fmt.Printf("Session Status                %s %s\n", color.Green("●"), color.Green("online"))
							fmt.Println(color.Green("✓ Reconnected successfully!"))
							fmt.Println()
						}
						lastStatus = isConnected
					}
				}
			}
		}
	}()

	// Set up terminal for raw input (to detect Ctrl+U)
	// Only enable if stdin is a terminal
	var oldState *term.State
	if term.IsTerminal(int(os.Stdin.Fd())) {
		var err error
		oldState, err = term.MakeRaw(int(os.Stdin.Fd()))
		if err == nil {
			defer term.Restore(int(os.Stdin.Fd()), oldState)
		}
	}

	// Channel for keyboard input (Ctrl+U)
	keyChan := make(chan bool, 1)

	// Read keyboard input in background (only if terminal)
	if term.IsTerminal(int(os.Stdin.Fd())) {
		go func() {
			b := make([]byte, 1)
			for {
				select {
				case <-ctx.Done():
					return
				default:
					n, err := os.Stdin.Read(b)
					if err != nil || n == 0 {
						return
					}
					// Ctrl+U is byte 21 (0x15)
					if b[0] == 21 {
						select {
						case keyChan <- true:
						default:
						}
					}
				}
			}
		}()
	}

	// Wait for interrupt signal or Ctrl+U
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	select {
	case <-sigChan:
		// Normal shutdown
	case <-keyChan:
		// Ctrl+U pressed - trigger upgrade
		updateInfoMu.Lock()
		hasUpdate := updateInfo != nil && updateInfo.UpdateAvailable
		currentUpdateInfo := updateInfo
		updateInfoMu.Unlock()

		if hasUpdate && currentUpdateInfo != nil {
			// Restore terminal before running upgrade
			if oldState != nil {
				term.Restore(int(os.Stdin.Fd()), oldState)
			}

			fmt.Println()
			fmt.Println()
			fmt.Println(color.Cyan("Upgrading UniRoute CLI..."))
			fmt.Println()

			// Run upgrade command
			upgradeCmd := exec.Command(os.Args[0], "upgrade")
			upgradeCmd.Stdout = os.Stdout
			upgradeCmd.Stderr = os.Stderr
			upgradeCmd.Stdin = os.Stdin
			if err := upgradeCmd.Run(); err != nil {
				fmt.Println(color.Red(fmt.Sprintf("Upgrade failed: %v", err)))
				fmt.Println()
				fmt.Println(color.Yellow("You can manually upgrade by running: uniroute upgrade"))
			}

			// Close tunnel after upgrade
			fmt.Println()
			fmt.Println(color.Yellow("Closing tunnel..."))
			stopMu.Lock()
			stopUpdates = true
			stopMu.Unlock()
			cancel()
			if err := client.Close(); err != nil {
				return err
			}
			return nil
		} else {
			// No update available
			fmt.Println()
			fmt.Println(color.Green("✓ You're already using the latest version"))
			fmt.Println()
			// Continue running tunnel - wait for Ctrl+C
			<-sigChan
		}
	}

	// Mark that we should stop updates immediately
	stopMu.Lock()
	stopUpdates = true
	stopMu.Unlock()

	// Cancel context to stop monitoring goroutine immediately
	cancel()

	// Clear any pending output (latency line) - move cursor to start, clear line, newline
	fmt.Print("\r\033[K")
	fmt.Println()

	// Give goroutine a moment to exit
	time.Sleep(300 * time.Millisecond)

	// Close tunnel connection
	fmt.Println()
	fmt.Println(color.Yellow("Shutting down tunnel..."))
	if err := client.Close(); err != nil {
		fmt.Println(color.Red(fmt.Sprintf("Error closing tunnel: %v", err)))
		return err
	}
	fmt.Println(color.Green("Tunnel closed successfully"))
	return nil
}
