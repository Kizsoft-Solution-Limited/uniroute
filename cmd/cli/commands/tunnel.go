package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/Kizsoft-Solution-Limited/uniroute/internal/tunnel"
	"github.com/Kizsoft-Solution-Limited/uniroute/pkg/color"
	"github.com/Kizsoft-Solution-Limited/uniroute/pkg/logger"
	versioncheck "github.com/Kizsoft-Solution-Limited/uniroute/pkg/version"
	"github.com/rs/zerolog"
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

Supported protocols:
  - http: HTTP/HTTPS tunneling (default)
  - tcp:  Raw TCP tunneling
  - tls:  TLS-encrypted TCP tunneling

Examples:
  uniroute tunnel                    # Create HTTP tunnel (auto-resumes if available)
  uniroute tunnel --port 8084        # Tunnel specific port
  uniroute tunnel --protocol tcp --port 3306  # TCP tunnel for MySQL
  uniroute tunnel --protocol tls --port 5432  # TLS tunnel for PostgreSQL
  uniroute tunnel --host myapp       # Request specific subdomain
  uniroute tunnel --all              # Start all enabled tunnels from config
  uniroute tunnel --init             # Create example tunnel configuration file
  uniroute tunnel --resume abc123    # Resume specific subdomain
  uniroute tunnel --list             # List all tunnels (local config + server subdomains)
  uniroute tunnel --clear            # Clear saved tunnel state

Configuration File:
  Create ~/.uniroute/tunnels.json to define multiple tunnels:
  {
    "version": "1.0",
    "tunnels": [
      {
        "name": "web",
        "protocol": "http",
        "local_addr": "localhost:8080",
        "enabled": true
      },
      {
        "name": "mysql",
        "protocol": "tcp",
        "local_addr": "localhost:3306",
        "enabled": true
      }
    ]
  }
  Then run: uniroute --all or uniroute tunnel --all

The --list flag shows:
  • Local Configuration Tunnels: Tunnels defined in ~/.uniroute/tunnels.json
  • Server Tunnels: Your active subdomain tunnels on the server (requires authentication)`,
	RunE: runTunnel,
}

var (
	tunnelPort      string
	tunnelProtocol  string
	tunnelHost      string
	tunnelServerURL string
	resumeSubdomain string
	clearSaved       bool
	listSaved       bool
	startAll        bool
	createConfig    bool
	forceNew        bool
)

func init() {
	tunnelCmd.Flags().StringVarP(&tunnelPort, "port", "p", "8084", "Local port to tunnel")
	tunnelCmd.Flags().StringVar(&tunnelProtocol, "protocol", "http", "Tunnel protocol: http, tcp, or tls")
	tunnelCmd.Flags().StringVar(&tunnelHost, "host", "", "Request specific host/subdomain")
	
	// Auto-detect local mode and set default tunnel server URL
	defaultTunnelServer := getTunnelServerURL()
	
	// Advanced option - hide from help by default (users can still use it)
	tunnelCmd.Flags().StringVarP(&tunnelServerURL, "server", "s", defaultTunnelServer, "Tunnel server URL (default: auto-detected based on environment)")
	tunnelCmd.Flags().MarkHidden("server") // Hide from help output
	
	tunnelCmd.Flags().StringVar(&resumeSubdomain, "resume", "", "Resume a specific subdomain (or use saved one if not specified)")
	tunnelCmd.Flags().BoolVar(&clearSaved, "clear", false, "Clear saved tunnel state")
	tunnelCmd.Flags().BoolVar(&listSaved, "list", false, "List saved tunnel state")
	tunnelCmd.Flags().BoolVar(&startAll, "all", false, "Start all enabled tunnels from configuration file")
	tunnelCmd.Flags().BoolVar(&createConfig, "init", false, "Create example tunnel configuration file")
	tunnelCmd.Flags().BoolVar(&forceNew, "new", false, "Force creating a new tunnel (don't resume saved state)")
}

func runTunnel(cmd *cobra.Command, args []string) error {
	log := logger.New()
	persistence := tunnel.NewTunnelPersistence(log)

	// Handle --init: create example config file
	if createConfig {
		return createExampleConfig(log)
	}

	// Handle list/clear commands
	if listSaved {
		return listAllTunnels()
	}

	if clearSaved {
		if err := persistence.Clear(); err != nil {
			return fmt.Errorf("failed to clear saved state: %w", err)
		}
		fmt.Println(color.Green("✓ Cleared saved tunnel state"))
		return nil
	}

	// Handle --all flag: start all enabled tunnels from config
	if startAll {
		return runAllTunnels(cmd, args)
	}

	// Always use built-in tunnel (default behavior)
	return runBuiltInTunnel(cmd, args)
}

// runAllTunnels starts all enabled tunnels from configuration file
func runAllTunnels(cmd *cobra.Command, args []string) error {
	log := logger.New()
	configManager := tunnel.NewConfigManager(log)

	// Load enabled tunnels from config
	tunnels, err := configManager.GetEnabledTunnels()
	if err != nil {
		return fmt.Errorf("failed to load tunnel configuration: %w", err)
	}

	if len(tunnels) == 0 {
		configPath := configManager.GetConfigPath()
		fmt.Println(color.Yellow("⚠️  No enabled tunnels found in configuration"))
		fmt.Println()
		fmt.Printf("Create a tunnel configuration file at: %s\n", color.Cyan(configPath))
		fmt.Println()
		fmt.Println(color.Bold("Quick Setup:"))
		fmt.Println()
		fmt.Printf("1. Create the directory:\n")
		fmt.Printf("   %s\n", color.Gray("mkdir -p ~/.uniroute"))
		fmt.Println()
		fmt.Printf("2. Create the config file:\n")
		fmt.Printf("   %s\n", color.Gray("nano ~/.uniroute/tunnels.json"))
		fmt.Println()
		fmt.Printf("3. Add your tunnel configuration:\n")
		fmt.Println(color.Gray(`{
  "version": "1.0",
  "tunnels": [
    {
      "name": "web",
      "protocol": "http",
      "local_addr": "localhost:8080",
      "enabled": true
    },
    {
      "name": "mysql",
      "protocol": "tcp",
      "local_addr": "localhost:3306",
      "enabled": true
    }
  ]
}`))
		fmt.Println()
		fmt.Printf("4. Start all tunnels:\n")
		fmt.Printf("   %s\n", color.Cyan("uniroute --all"))
		fmt.Println()
		fmt.Printf("For more details, see: %s\n", color.Gray("docs/TUNNEL_CONFIG.md"))
		return nil
	}

	fmt.Println(color.Cyan(fmt.Sprintf("🚀 Starting %d tunnel(s)...", len(tunnels))))
	fmt.Println()

	// Start each tunnel in a goroutine
	var wg sync.WaitGroup
	errors := make(chan error, len(tunnels))

	for _, tunnelConfig := range tunnels {
		wg.Add(1)
		go func(tc tunnel.TunnelConfig) {
			defer wg.Done()
			if err := startTunnelFromConfig(tc); err != nil {
				errors <- fmt.Errorf("tunnel '%s': %w", tc.Name, err)
			}
		}(tunnelConfig)
	}

	// Wait for all tunnels to start
	wg.Wait()
	close(errors)

	// Check for errors
	var hasErrors bool
	for err := range errors {
		fmt.Println(color.Red(fmt.Sprintf("❌ %s", err.Error())))
		hasErrors = true
	}

	if hasErrors {
		return fmt.Errorf("some tunnels failed to start")
	}

	fmt.Println()
	fmt.Println(color.Green("✓ All tunnels started successfully"))
	fmt.Println(color.Gray("Press Ctrl+C to stop all tunnels"))

	// Wait for interrupt
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	fmt.Println()
	fmt.Println(color.Yellow("Shutting down all tunnels..."))
	return nil
}

// startTunnelFromConfig starts a single tunnel from configuration
func startTunnelFromConfig(tc tunnel.TunnelConfig) error {
	log := logger.New()

	// Determine local URL based on protocol
	var localURL string
	if tc.Protocol == "http" {
		if !strings.HasPrefix(tc.LocalAddr, "http://") && !strings.HasPrefix(tc.LocalAddr, "https://") {
			localURL = "http://" + tc.LocalAddr
		} else {
			localURL = tc.LocalAddr
		}
	} else {
		// For TCP/TLS, use host:port format
		localURL = tc.LocalAddr
	}

	// Use server URL from config, or auto-detect, or use default
	serverURL := tc.ServerURL
	if serverURL == "" {
		if tunnelServerURL != "" && tunnelServerURL != "tunnel.uniroute.co" {
			serverURL = tunnelServerURL
		} else {
			serverURL = getTunnelServerURL()
		}
	}

	fmt.Printf("   %s %s (%s) -> %s\n",
		color.Green("✓"),
		color.Bold(tc.Name),
		color.Gray(tc.Protocol),
		color.Cyan(localURL))

	// Create tunnel client with protocol from config
	client := tunnel.NewTunnelClientWithOptions(serverURL, localURL, tc.Protocol, tc.Host, log)

	// Connect to tunnel server
	if err := client.Connect(); err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}

	// Get tunnel info
	info := client.GetTunnelInfo()
	if info == nil {
		return fmt.Errorf("failed to get tunnel information")
	}

	fmt.Printf("      %s %s\n", color.Gray("Public URL:"), color.Cyan(info.PublicURL))

	// Keep tunnel running (in real implementation, this would be managed)
	// For now, just log that it's started
	return nil
}

// runBuiltInTunnel uses the built-in tunnel client
func runBuiltInTunnel(cmd *cobra.Command, args []string) error {
	// Auto-detect tunnel server URL if not explicitly set or if using default production URL
	// Always re-check to ensure we use local mode when appropriate
	if tunnelServerURL == "" || 
		tunnelServerURL == "tunnel.uniroute.co" || 
		tunnelServerURL == "https://tunnel.uniroute.co" {
		tunnelServerURL = getTunnelServerURL()
	}
	
	// Check if using public server - require authentication
	// Allow localhost for development/testing without auth
	isLocalTunnel := tunnelServerURL == "localhost:8055" ||
		tunnelServerURL == "http://localhost:8055" ||
		tunnelServerURL == "localhost:8080" ||
		tunnelServerURL == "http://localhost:8080" ||
		strings.Contains(tunnelServerURL, "localhost") ||
		strings.Contains(tunnelServerURL, "127.0.0.1")
	
	if !isLocalTunnel &&
		(tunnelServerURL == "tunnel.uniroute.co" ||
			tunnelServerURL == "https://tunnel.uniroute.co" ||
			strings.Contains(tunnelServerURL, ".uniroute.co")) {
		token := getAuthToken()
		if token == "" {
			return fmt.Errorf("authentication required for public tunnel server\nRun 'uniroute auth login' first")
		}
	}

	log := logger.New()

	// Validate protocol
	if tunnelProtocol != "http" && tunnelProtocol != "tcp" && tunnelProtocol != "tls" {
		return fmt.Errorf("invalid protocol '%s'. Must be: http, tcp, or tls", tunnelProtocol)
	}

	// Determine local URL based on protocol
	var localURL string
	if tunnelProtocol == "http" {
		localURL = fmt.Sprintf("http://localhost:%s", tunnelPort)
	} else {
		// For TCP/TLS, use host:port format
		localURL = fmt.Sprintf("localhost:%s", tunnelPort)
	}

	fmt.Println()
	fmt.Println(color.Cyan("Starting UniRoute Tunnel..."))
	fmt.Println()
	fmt.Printf("   %s %s\n", color.Gray("Protocol:"), color.Bold(tunnelProtocol))
	fmt.Printf("   %s %s\n", color.Gray("Local URL:"), color.Bold(localURL))
	if tunnelHost != "" {
		fmt.Printf("   %s %s\n", color.Gray("Host:"), color.Bold(tunnelHost))
	}
	fmt.Printf("   %s %s\n", color.Gray("Tunnel Server:"), color.Bold(tunnelServerURL))
	fmt.Println()

	// Create tunnel client with protocol and host
	client := tunnel.NewTunnelClientWithOptions(tunnelServerURL, localURL, tunnelProtocol, tunnelHost, log)
	
	// Set auth token if user is authenticated (for automatic tunnel-user association)
	token := getAuthToken()
	if token != "" {
		client.SetToken(token)
		log.Debug().Msg("Auth token set - tunnel will be automatically associated with your account")
	}
	
	// Set up request handler for displaying HTTP requests (will be set after initial output)
	var requestEventsMu sync.Mutex
	var requestEvents []tunnel.RequestEvent
	maxRequestEvents := 20 // Keep last 20 requests
	
	// Request handler will be set after we print the initial output
	var requestHandler tunnel.RequestEventHandler

	// If --new flag is set, clear resume info to force new tunnel
	if forceNew {
		client.ClearResumeInfo()
		log.Info().Msg("--new flag set: forcing new tunnel creation (not resuming)")
	}

	// If --resume flag is set, use that subdomain instead of saved state
	if resumeSubdomain != "" {
		// Load saved state to get tunnel ID if available
		persistence := tunnel.NewTunnelPersistence(log)
		if state, err := persistence.Load(); err == nil && state != nil {
			if state.Subdomain == resumeSubdomain {
				// Use saved tunnel ID if subdomain matches
				client.SetResumeInfo(resumeSubdomain, state.TunnelID)
				log.Info().
					Str("subdomain", resumeSubdomain).
					Str("tunnel_id", state.TunnelID).
					Msg("Resuming tunnel with saved tunnel ID")
			} else {
				// Just use subdomain, let server find the tunnel
				client.SetResumeInfo(resumeSubdomain, "")
				log.Info().
					Str("subdomain", resumeSubdomain).
					Msg("Resuming tunnel by subdomain")
			}
		} else {
			// No saved state, just use subdomain
			client.SetResumeInfo(resumeSubdomain, "")
			log.Info().
				Str("subdomain", resumeSubdomain).
				Msg("Resuming tunnel by subdomain (no saved state)")
		}
	}

	// Connect to tunnel server
	if err := client.Connect(); err != nil {
		return fmt.Errorf("failed to connect to tunnel server: %w", err)
	}

	// Get tunnel info
	info := client.GetTunnelInfo()
	if info == nil {
		return fmt.Errorf("failed to get tunnel information")
	}
	
	// Get tunnel ID for association if needed
	tunnelID := info.ID

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

	// Initial connection stats (needed before printing)
	initialStats := &tunnel.ConnectionStats{
		Total: 0,
		Open:  1,
		RT1:   0.00,
		RT5:   0.00,
		P50:   0.00,
		P90:   0.00,
	}
	
	// Track line numbers for dynamic fields as we print
	// We'll count lines from the start to know where each dynamic field is
	var dynamicLineNumbers = struct {
		sync.Mutex
		sessionStatus int // Line number for Session Status
		latency       int // Line number for Latency
		connections   int // Line number for Connections stats
	}{}
	
	lineCounter := 0
	countLine := func() int {
		lineCounter++
		return lineCounter - 1
	}

	// Session Status with dynamic emoji (will be updated)
	fmt.Printf("Session Status                %s %s\n", color.Green("●"), color.Green("online"))
	dynamicLineNumbers.sessionStatus = countLine()

	fmt.Printf("Account                       %s\n", color.Gray(accountDisplay))
	countLine()
	currentVersion := GetVersion()
	fmt.Printf("Version                       %s\n", color.Gray(currentVersion))
	countLine()

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
	countLine()
	fmt.Printf("Region                        %s\n", color.Gray("Local"))
	countLine()
	
	// Latency (will be updated)
	fmt.Printf("Latency                       %s\n", color.Gray(latencyStr))
	dynamicLineNumbers.latency = countLine()
	
	fmt.Printf("Web Interface                 %s\n", color.Cyan("http://127.0.0.1:4040"))
	countLine()
	fmt.Println()
	countLine()
	fmt.Printf("Connections                   ttl     opn     rt1     rt5     p50     p90\n")
	countLine()
	
	// Connections stats (will be updated)
	fmt.Printf("                              %s      %s      %s      %s      %s      %s\n",
		color.Gray(fmt.Sprintf("%d", initialStats.Total)),
		color.Gray(fmt.Sprintf("%d", initialStats.Open)),
		color.Gray(fmt.Sprintf("%.2f", initialStats.RT1)),
		color.Gray(fmt.Sprintf("%.2f", initialStats.RT5)),
		color.Gray(fmt.Sprintf("%.2f", initialStats.P50)),
		color.Gray(fmt.Sprintf("%.2f", initialStats.P90)))
	dynamicLineNumbers.connections = countLine()
	
	fmt.Println()
	countLine()
	fmt.Printf("🌍 Public URL:\n")
	countLine()
	fmt.Printf("   %s\n", color.Cyan(info.PublicURL))
	countLine()
	fmt.Println()
	countLine()
	fmt.Printf("🔗 Forwarding:\n")
	countLine()
	fmt.Printf("   %s %s %s\n",
		color.Cyan(info.PublicURL),
		color.Gray("->"),
		color.Bold(localURL))
	countLine()
	fmt.Println()
	countLine()
	fmt.Printf("🆔 Tunnel ID:\n")
	countLine()
	fmt.Printf("   %s\n", color.Gray(tunnelID))
	countLine()
	if token != "" {
		fmt.Printf("   %s\n", color.Green("✓ Authenticated - tunnel will be automatically associated with your account"))
		countLine()
		fmt.Printf("   %s\n", color.Gray("   If tunnel doesn't appear in dashboard, ensure tunnel server has JWT_SECRET configured"))
		countLine()
	} else {
		fmt.Printf("   %s\n", color.Yellow("⚠ Not logged in - tunnel will not appear in dashboard"))
		countLine()
		fmt.Printf("   %s\n", color.Gray("   Run 'uniroute auth login' and recreate tunnel to auto-associate"))
		countLine()
		fmt.Printf("   %s\n", color.Gray(fmt.Sprintf("   Or manually associate: curl -X POST %s/auth/tunnels/%s/associate -H \"Authorization: Bearer YOUR_TOKEN\"", getServerURL(), tunnelID)))
		countLine()
	}
	fmt.Println()
	countLine()
	fmt.Printf("HTTP Requests\n")
	countLine()
	fmt.Printf("-------------\n")
	countLine()
	// HTTP Requests will be displayed here as they come in
	// Format: TIME METHOD PATH STATUS_CODE STATUS_TEXT
	
	fmt.Println()
	countLine()
	fmt.Printf("📋 %s\n", color.Gray("To see database save logs, check tunnel server output (run with --log-level debug)"))
	countLine()
	fmt.Println()
	countLine()
	// Ensure quote is displayed and flushed
	quoteLine := color.Yellow(fmt.Sprintf("💬 %s", funnyQuote))
	fmt.Println(quoteLine)
	countLine()
	os.Stdout.Sync() // Flush output to ensure quote is visible
	fmt.Println()
	countLine()
	fmt.Println(color.Gray("Press Ctrl+C to stop"))
	countLine()
	updateInfoMu.Lock()
	hasUpdate := updateInfo != nil && updateInfo.UpdateAvailable
	updateInfoMu.Unlock()
	if hasUpdate {
		fmt.Println(color.Gray("Press Ctrl+U to upgrade"))
		countLine()
	}
	fmt.Println()
	
	// Save total line count - this is where we are now (after all output)
	totalLines := lineCounter
	
	// Function to format and display a request event
	formatRequestEvent := func(event tunnel.RequestEvent) {
		// Get timezone abbreviation (WAT, EST, etc.)
		loc, _ := time.LoadLocation("Africa/Lagos") // Default to WAT, can be made configurable
		eventTime := event.Time.In(loc)
		timeStr := eventTime.Format("15:04:05.000 MST")
		
		// Truncate path if too long (max 50 chars for display)
		path := event.Path
		if len(path) > 50 {
			path = path[:47] + "..."
		}
		
		// Format status code and text
		statusCodeStr := fmt.Sprintf("%d", event.StatusCode)
		statusText := event.StatusText
		// Extract status text (e.g., "200 OK" -> "OK")
		if strings.Contains(statusText, " ") {
			parts := strings.SplitN(statusText, " ", 2)
			if len(parts) > 1 {
				statusText = parts[1]
			}
		}
		// Truncate status text if too long (max 10 chars)
		if len(statusText) > 10 {
			statusText = statusText[:7] + "..."
		}
		
		// Color code based on status
		var statusColor func(string) string
		if event.StatusCode >= 200 && event.StatusCode < 300 {
			statusColor = color.Green
		} else if event.StatusCode >= 300 && event.StatusCode < 400 {
			statusColor = color.Yellow
		} else if event.StatusCode >= 400 && event.StatusCode < 500 {
			statusColor = color.Red
		} else if event.StatusCode >= 500 {
			statusColor = color.Red
		} else {
			statusColor = color.Gray
		}
		
		// Format the request line (similar to ngrok format)
		// Format: TIME(20) METHOD(7) PATH(50) STATUS_CODE(3) STATUS_TEXT(10)
		// Total width: ~90 chars
		methodPadded := event.Method
		if len(methodPadded) < 7 {
			methodPadded = methodPadded + strings.Repeat(" ", 7-len(methodPadded))
		}
		
		pathPadded := path
		if len(pathPadded) < 50 {
			pathPadded = pathPadded + strings.Repeat(" ", 50-len(pathPadded))
		}
		
		statusCodePadded := statusCodeStr
		if len(statusCodePadded) < 3 {
			statusCodePadded = strings.Repeat(" ", 3-len(statusCodePadded)) + statusCodePadded
		}
		
		requestLine := fmt.Sprintf("%s %s %s %s %s",
			color.Gray(timeStr),
			color.Cyan(methodPadded),
			pathPadded,
			statusColor(statusCodePadded),
			statusColor(statusText))
		
		// Print the request line cleanly
		fmt.Printf("%s\n", requestLine)
		os.Stdout.Sync()
	}
	
	// Set up request handler now that we know where to display requests
	requestHandler = func(event tunnel.RequestEvent) {
		requestEventsMu.Lock()
		requestEvents = append(requestEvents, event)
		// Keep only last N requests
		if len(requestEvents) > maxRequestEvents {
			requestEvents = requestEvents[len(requestEvents)-maxRequestEvents:]
		}
		requestEventsMu.Unlock()
		
		// Format and display the request
		formatRequestEvent(event)
	}
	client.SetRequestHandler(requestHandler)
	
	// Function to update a specific dynamic line
	// We know the line numbers from when we printed them
	updateLine := func(which string, newContent string) {
		var targetLine int
		switch which {
		case "status":
			targetLine = dynamicLineNumbers.sessionStatus
		case "latency":
			targetLine = dynamicLineNumbers.latency
		case "connections":
			targetLine = dynamicLineNumbers.connections
		default:
			return
		}
		
		// Calculate lines to move up from current position (end of output)
		linesToMove := totalLines - targetLine
		
		if linesToMove > 0 {
			// Move up, update, move back down
			fmt.Printf("\033[%dA\r%s\033[K\033[%dB", linesToMove, newContent, linesToMove)
			os.Stdout.Sync()
		}
	}
	
	// Track current values for dynamic fields
	var (
		currentStatusLine  = fmt.Sprintf("Session Status                %s %s", color.Green("●"), color.Green("online"))
		currentLatencyLine = fmt.Sprintf("Latency                       %s", color.Gray(latencyStr))
		currentStatsLine   = fmt.Sprintf("                              %s      %s      %s      %s      %s      %s",
			color.Gray(fmt.Sprintf("%d", initialStats.Total)),
			color.Gray(fmt.Sprintf("%d", initialStats.Open)),
			color.Gray(fmt.Sprintf("%.2f", initialStats.RT1)),
			color.Gray(fmt.Sprintf("%.2f", initialStats.RT5)),
			color.Gray(fmt.Sprintf("%.2f", initialStats.P50)),
			color.Gray(fmt.Sprintf("%.2f", initialStats.P90)))
	)

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Track if we should stop printing updates
	var stopMu sync.Mutex
	stopUpdates := false

	// Monitor connection status, latency, and stats, update display dynamically
	go func() {
		lastStatusState := "online" // Track full status state: "online", "reconnecting", or "offline"
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

				// Get connection status first (most reliable check)
				// This is a simple state check and won't be affected by network requests
				connectionStatus := client.GetConnectionStatus()
				isConnected := client.IsConnected() // Used for stats fetching
				currentLatency := client.GetLatency()

				// Only fetch connection stats if we're actually connected (not reconnecting/offline)
				// Use a timeout to prevent stats fetch from blocking or causing false status changes
				var currentStats *tunnel.ConnectionStats
				if connectionStatus == "online" && isConnected && info != nil {
					// Fetch stats with timeout to prevent blocking
					statsChan := make(chan *tunnel.ConnectionStats, 1)
					errChan := make(chan error, 1)
					go func() {
						stats, err := client.GetConnectionStats(tunnelServerURL, info.ID)
						if err != nil {
							errChan <- err
							return
						}
						statsChan <- stats
					}()
					
					select {
					case stats := <-statsChan:
						currentStats = stats
					case <-errChan:
						// Stats fetch failed - keep last stats (don't change status)
						currentStats = lastStats
					case <-time.After(2 * time.Second):
						// Stats fetch timed out - keep last stats (don't change status)
						currentStats = lastStats
					}
				} else {
					currentStats = lastStats
				}

				// Update latency display if it changed significantly
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
						// Update latency line
						currentLatencyLine = fmt.Sprintf("Latency                       %s", color.Gray(latencyStr))
						updateLine("latency", currentLatencyLine)
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
						// Update connections stats line
						currentStatsLine = fmt.Sprintf("                              %s      %s      %s      %s      %s      %s",
							color.Gray(fmt.Sprintf("%d", currentStats.Total)),
							color.Gray(fmt.Sprintf("%d", currentStats.Open)),
							color.Gray(fmt.Sprintf("%.2f", currentStats.RT1)),
							color.Gray(fmt.Sprintf("%.2f", currentStats.RT5)),
							color.Gray(fmt.Sprintf("%.2f", currentStats.P50)),
							color.Gray(fmt.Sprintf("%.2f", currentStats.P90)))
						updateLine("connections", currentStatsLine)
						lastStats = currentStats
					}
				}

				// Check if status changed (connected, reconnecting, or offline)
				// Only update if status actually changed to avoid false positives
				currentStatusState := connectionStatus
				statusChanged := currentStatusState != lastStatusState

				if statusChanged {
					// Check context before printing status changes
					select {
					case <-ctx.Done():
						return
					default:
						wasOnline := lastStatusState == "online"
						
						// Update Session Status
						switch connectionStatus {
						case "online":
							currentStatusLine = fmt.Sprintf("Session Status                %s %s", color.Green("●"), color.Green("online"))
						case "reconnecting":
							currentStatusLine = fmt.Sprintf("Session Status                %s %s", color.Yellow("●"), color.Yellow("reconnecting"))
						default:
							currentStatusLine = fmt.Sprintf("Session Status                %s %s", color.Red("●"), color.Red("offline"))
						}
						
						updateLine("status", currentStatusLine)
						
						// Show status change messages
						if statusChanged {
							if !wasOnline && connectionStatus == "online" {
								fmt.Print("\r\033[K")
								fmt.Println(color.Green("✓ Reconnected successfully!"))
							} else if wasOnline && (connectionStatus == "reconnecting" || connectionStatus == "offline") {
								fmt.Print("\r\033[K")
								fmt.Println(color.Yellow("⚠️  Connection lost, attempting to reconnect..."))
							}
						}
						
						lastStatusState = currentStatusState
					}
				}
			}
		}
	}()

	// Set up signal handling FIRST (before raw terminal mode)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

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
	// This goroutine will forward Ctrl+C to the signal channel as a fallback
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
					// Ctrl+U is byte 21 (0x15) - trigger upgrade
					if b[0] == 21 {
						select {
						case keyChan <- true:
						default:
						}
					}
					// Ctrl+C is byte 3 (0x03) - forward to signal channel as fallback
					// In raw mode, Ctrl+C might not trigger SIGINT, so we handle it manually
					if b[0] == 3 {
						select {
						case sigChan <- os.Interrupt:
						default:
						}
					}
				}
			}
		}()
	}

	// Use a loop to handle both signals and keyboard input
	for {
		select {
		case <-sigChan:
			// Ctrl+C or SIGTERM received - shutdown
			goto shutdown
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
				// No update available - continue waiting
				fmt.Println()
				fmt.Println(color.Green("✓ You're already using the latest version"))
				fmt.Println()
				// Continue loop to wait for Ctrl+C
			}
		case <-ctx.Done():
			// Context cancelled - shutdown
			goto shutdown
		}
	}

shutdown:

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

// listAllTunnels lists all tunnels (both from config and from server)
func listAllTunnels() error {
	fmt.Println()
	fmt.Println(color.Cyan("📋 UniRoute Tunnels"))
	fmt.Println()

	// List tunnels from local config file
	listLocalTunnels()

	// List tunnels from server (if authenticated)
	listServerTunnels()

	// Show usage
	showTunnelUsage()

	return nil
}

// createExampleConfig creates an example tunnel configuration file
func createExampleConfig(log zerolog.Logger) error {
	configManager := tunnel.NewConfigManager(log)
	configPath := configManager.GetConfigPath()

	// Check if config file already exists
	if _, err := os.Stat(configPath); err == nil {
		fmt.Println(color.Yellow(fmt.Sprintf("⚠️  Configuration file already exists at: %s", configPath)))
		fmt.Println()
		fmt.Println("To view the current configuration:")
		fmt.Printf("  %s\n", color.Cyan("cat "+configPath))
		fmt.Println()
		fmt.Println("To edit the configuration:")
		fmt.Printf("  %s\n", color.Cyan("nano "+configPath))
		return nil
	}

	// Create example config
	exampleConfig := &tunnel.TunnelConfigFile{
		Version: "1.0",
		Tunnels: []tunnel.TunnelConfig{
			{
				Name:      "web",
				Protocol:  "http",
				LocalAddr: "localhost:8080",
				Enabled:   true,
			},
			{
				Name:      "api",
				Protocol:  "http",
				LocalAddr: "localhost:3000",
				Enabled:   true,
			},
			{
				Name:      "mysql",
				Protocol:  "tcp",
				LocalAddr: "localhost:3306",
				Enabled:   false, // Disabled by default
			},
			{
				Name:      "postgres",
				Protocol:  "tls",
				LocalAddr: "localhost:5432",
				Enabled:   false, // Disabled by default
			},
		},
	}

	// Save config
	if err := configManager.Save(exampleConfig); err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}

	fmt.Println(color.Green("✓ Created example tunnel configuration file"))
	fmt.Println()
	fmt.Printf("Location: %s\n", color.Cyan(configPath))
	fmt.Println()
	fmt.Println(color.Bold("Example configuration:"))
	fmt.Println(color.Gray(`{
  "version": "1.0",
  "tunnels": [
    {
      "name": "web",
      "protocol": "http",
      "local_addr": "localhost:8080",
      "enabled": true
    },
    {
      "name": "api",
      "protocol": "http",
      "local_addr": "localhost:3000",
      "enabled": true
    },
    {
      "name": "mysql",
      "protocol": "tcp",
      "local_addr": "localhost:3306",
      "enabled": false
    },
    {
      "name": "postgres",
      "protocol": "tls",
      "local_addr": "localhost:5432",
      "enabled": false
    }
  ]
}`))
	fmt.Println()
	fmt.Println(color.Bold("Next steps:"))
	fmt.Printf("1. Edit the config file: %s\n", color.Cyan("nano "+configPath))
	fmt.Printf("2. Enable/disable tunnels by setting \"enabled\": true/false\n")
	fmt.Printf("3. Add your own tunnels to the \"tunnels\" array\n")
	fmt.Printf("4. Start all enabled tunnels: %s\n", color.Cyan("uniroute --all"))
	fmt.Println()
	fmt.Printf("For more details, see: %s\n", color.Gray("docs/TUNNEL_CONFIG.md"))

	return nil
}

// listLocalTunnels lists tunnels from the local config file
func listLocalTunnels() {
	log := logger.New()
	configManager := tunnel.NewConfigManager(log)

	config, err := configManager.Load()
	if err != nil {
		fmt.Println(color.Yellow("⚠️  Could not load tunnel configuration"))
		return
	}

	if len(config.Tunnels) == 0 {
		fmt.Println(color.Gray("📁 Local Configuration: No tunnels configured"))
		fmt.Println()
		return
	}

	fmt.Println(color.Bold("📁 Local Configuration Tunnels"))
	fmt.Println()

	// Table header
	fmt.Printf("   %-20s %-10s %-25s %-8s\n",
		color.Bold("Name"),
		color.Bold("Protocol"),
		color.Bold("Local Address"),
		color.Bold("Status"))
	fmt.Printf("   %s\n", color.Gray(strings.Repeat("-", 70)))

	// List all tunnels (enabled and disabled)
	for _, tc := range config.Tunnels {
		status := color.Red("disabled")
		if tc.Enabled {
			status = color.Green("enabled")
		}

		fmt.Printf("   %-20s %-10s %-25s %s\n",
			color.Bold(tc.Name),
			color.Gray(tc.Protocol),
			color.Cyan(tc.LocalAddr),
			status)

		// Show additional info if available
		if tc.Host != "" {
			fmt.Printf("      %s %s\n", color.Gray("Host:"), color.Gray(tc.Host))
		}
		if tc.ServerURL != "" {
			fmt.Printf("      %s %s\n", color.Gray("Server:"), color.Gray(tc.ServerURL))
		}
	}

	fmt.Println()
}

// listServerTunnels lists tunnels from the server (subdomain tunnels)
func listServerTunnels() {
	token := getAuthToken()
	if token == "" {
		fmt.Println(color.Gray("🌐 Server Tunnels: Not authenticated (run 'uniroute auth login' to see your subdomain tunnels)"))
		fmt.Println()
		return
	}

	serverURL := getServerURL()
	apiURL := fmt.Sprintf("%s/v1/tunnels", serverURL)

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		fmt.Println(color.Yellow("⚠️  Could not create request to list server tunnels"))
		fmt.Println()
		return
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(color.Yellow("⚠️  Could not connect to server to list tunnels"))
		fmt.Println()
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		// Clear expired token
		clearExpiredToken()
		fmt.Println(color.Yellow("⚠️  Authentication expired (run 'uniroute auth login' to re-authenticate)"))
		fmt.Println()
		return
	}

	if resp.StatusCode != http.StatusOK {
		fmt.Println(color.Yellow(fmt.Sprintf("⚠️  Server returned status %d", resp.StatusCode)))
		fmt.Println()
		return
	}

	var result struct {
		Tunnels []struct {
			ID           string `json:"id"`
			Subdomain    string `json:"subdomain"`
			PublicURL    string `json:"public_url"`
			LocalURL     string `json:"local_url"`
			Status       string `json:"status"`
			RequestCount int64  `json:"request_count"`
			CreatedAt    string `json:"created_at"`
			LastActive   string `json:"last_active,omitempty"`
		} `json:"tunnels"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Println(color.Yellow("⚠️  Could not parse server response"))
		fmt.Println()
		return
	}

	if len(result.Tunnels) == 0 {
		fmt.Println(color.Gray("🌐 Server Tunnels: No active tunnels on server"))
		fmt.Println()
		return
	}

	fmt.Println(color.Bold("🌐 Server Tunnels (Subdomain Tunnels)"))
	fmt.Println()

	// Table header
	fmt.Printf("   %-20s %-10s %-40s %-10s %-8s\n",
		color.Bold("Subdomain"),
		color.Bold("Status"),
		color.Bold("Public URL"),
		color.Bold("Requests"),
		color.Bold("Last Active"))
	fmt.Printf("   %s\n", color.Gray(strings.Repeat("-", 95)))

	// List tunnels
	for _, t := range result.Tunnels {
		statusColor := color.Green
		if t.Status != "active" {
			statusColor = color.Red
		}

		lastActive := "Never"
		if t.LastActive != "" {
			// Parse and format timestamp
			if parsed, err := time.Parse(time.RFC3339, t.LastActive); err == nil {
				lastActive = parsed.Format("2006-01-02 15:04")
			}
		}

		fmt.Printf("   %-20s %-10s %-40s %-10s %-8s\n",
			color.Cyan(t.Subdomain),
			statusColor(t.Status),
			color.Cyan(t.PublicURL),
			color.Gray(fmt.Sprintf("%d", t.RequestCount)),
			color.Gray(lastActive))

		// Show local URL if available
		if t.LocalURL != "" {
			fmt.Printf("      %s %s\n", color.Gray("Local:"), color.Gray(t.LocalURL))
		}
	}

	fmt.Println()
}

// showTunnelUsage shows usage examples and help
func showTunnelUsage() {
	fmt.Println(color.Bold("📖 Usage Examples"))
	fmt.Println()

	fmt.Println(color.Cyan("1. Create a new tunnel:"))
	fmt.Printf("   %s\n", color.Gray("uniroute http 8080"))
	fmt.Printf("   %s\n", color.Gray("uniroute tcp 3306"))
	fmt.Printf("   %s\n", color.Gray("uniroute tls 5432"))
	fmt.Println()
	fmt.Println(color.Cyan("1b. Force new tunnel (don't resume):"))
	fmt.Printf("   %s\n", color.Gray("uniroute http 8080 --new"))
	fmt.Printf("   %s\n", color.Gray("uniroute tunnel --new --port 8080"))
	fmt.Println()

	fmt.Println(color.Cyan("2. Start all enabled tunnels from config:"))
	fmt.Printf("   %s\n", color.Gray("uniroute tunnel --all"))
	fmt.Println()

	fmt.Println(color.Cyan("3. Resume a specific tunnel:"))
	fmt.Printf("   %s\n", color.Gray("uniroute resume abc123"))
	fmt.Printf("   %s\n", color.Gray("uniroute resume abc123 --port 8080"))
	fmt.Printf("   %s\n", color.Gray("uniroute resume abc123 --protocol tcp --port 3306"))
	fmt.Printf("   %s\n", color.Gray("# Or just run the same command - it auto-resumes"))
	fmt.Printf("   %s\n", color.Gray("uniroute http 8080  # Automatically resumes if saved state exists"))
	fmt.Println()

	fmt.Println(color.Cyan("4. List all tunnels:"))
	fmt.Printf("   %s\n", color.Gray("uniroute list"))
	fmt.Println()

	fmt.Println(color.Cyan("5. Clear saved tunnel state:"))
	fmt.Printf("   %s\n", color.Gray("uniroute tunnel --clear"))
	fmt.Println()

	fmt.Println(color.Cyan("6. Configure multiple tunnels:"))
	fmt.Println()
	fmt.Println(color.Gray("   Step 1: Create config directory"))
	fmt.Printf("   %s\n", color.Gray("   mkdir -p ~/.uniroute"))
	fmt.Println()
	fmt.Println(color.Gray("   Step 2: Create config file"))
	fmt.Printf("   %s\n", color.Gray("   nano ~/.uniroute/tunnels.json"))
	fmt.Println()
	fmt.Println(color.Gray("   Step 3: Add tunnel configurations"))
	fmt.Println(color.Gray(`   {
     "version": "1.0",
     "tunnels": [
       {
         "name": "web",
         "protocol": "http",
         "local_addr": "localhost:8080",
         "enabled": true
       },
       {
         "name": "mysql",
         "protocol": "tcp",
         "local_addr": "localhost:3306",
         "enabled": true
       }
     ]
   }`))
	fmt.Println()
	fmt.Println(color.Gray("   Step 4: Start all tunnels"))
	fmt.Printf("   %s\n", color.Cyan("   uniroute --all"))
	fmt.Println()
	fmt.Printf("   %s\n", color.Gray("   See docs/TUNNEL_CONFIG.md for complete guide"))
	fmt.Println()

	fmt.Println(color.Cyan("7. View tunnel details:"))
	fmt.Printf("   %s\n", color.Gray("uniroute tunnel --list"))
	fmt.Printf("   %s\n", color.Gray("# Shows both local config tunnels and server subdomain tunnels"))
	fmt.Println()

	fmt.Println(color.Bold("💡 Tips:"))
	fmt.Printf("   • %s\n", color.Gray("Tunnels automatically resume: Run the same command to get the same subdomain"))
	fmt.Printf("   • %s\n", color.Gray("Saved state: ~/.uniroute/tunnel-state.json (auto-saved on tunnel creation)"))
	fmt.Printf("   • %s\n", color.Gray("Local config: ~/.uniroute/tunnels.json (for multiple tunnels)"))
	fmt.Printf("   • %s\n", color.Gray("Use 'uniroute auth login' to authenticate and see server tunnels"))
	fmt.Printf("   • %s\n", color.Gray("Find your subdomain: uniroute tunnel --list"))
	fmt.Println()
}
