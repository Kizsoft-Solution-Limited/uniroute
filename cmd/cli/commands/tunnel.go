package commands

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/Kizsoft-Solution-Limited/uniroute/internal/tunnel"
	"github.com/Kizsoft-Solution-Limited/uniroute/pkg/color"
	"github.com/Kizsoft-Solution-Limited/uniroute/pkg/logger"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
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
  uniroute tunnel --protocol udp --port 53    # UDP tunnel for DNS
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
      },
      {
        "name": "dns",
        "protocol": "udp",
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
	customDomain    string
	clearSaved      bool
	listSaved       bool
	startAll        bool
	createConfig    bool
	forceNew        bool
)

func init() {
	tunnelCmd.Flags().StringVarP(&tunnelPort, "port", "p", "8084", "Local port to tunnel")
	tunnelCmd.Flags().StringVar(&tunnelProtocol, "protocol", "http", "Tunnel protocol: http, tcp, tls, or udp")
	tunnelCmd.Flags().StringVar(&tunnelHost, "host", "", "Request specific host/subdomain")
	tunnelCmd.Flags().StringVar(&customDomain, "domain", "", "Set custom domain for tunnel (requires DNS configuration)")
	
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
		errStr := err.Error()
		// Check if this is an authentication error - automatically log out
		if strings.Contains(strings.ToLower(errStr), "token") ||
		   strings.Contains(strings.ToLower(errStr), "authentication") ||
		   strings.Contains(strings.ToLower(errStr), "expired") ||
		   strings.Contains(strings.ToLower(errStr), "invalid") {
			// Clear expired/invalid token
			clearExpiredToken()
			log.Warn().Msg("Authentication failed - token cleared. Please run 'uniroute auth login' to authenticate again")
			return fmt.Errorf("authentication failed: %w\n\nPlease run 'uniroute auth login' to authenticate again", err)
		}
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
	if tunnelProtocol != "http" && tunnelProtocol != "tcp" && tunnelProtocol != "tls" && tunnelProtocol != "udp" {
		return fmt.Errorf("invalid protocol '%s'. Must be: http, tcp, tls, or udp", tunnelProtocol)
	}

	// Determine local URL based on protocol
	var localURL string
	if tunnelProtocol == "http" {
		localURL = fmt.Sprintf("http://localhost:%s", tunnelPort)
	} else {
		// For TCP/TLS, use host:port format
		localURL = fmt.Sprintf("localhost:%s", tunnelPort)
	}

	// Use Bubble Tea for UI if available, otherwise fall back to ANSI approach
	// For now, we'll use Bubble Tea
	// Note: "Starting UniRoute Tunnel..." is now shown in the Bubble Tea header

	// Create tunnel client with protocol and host
	client := tunnel.NewTunnelClientWithOptions(tunnelServerURL, localURL, tunnelProtocol, tunnelHost, log)
	
	// Set auth token if user is authenticated (for automatic tunnel-user association)
	token := getAuthToken()
	if token != "" {
		client.SetToken(token)
		log.Debug().Msg("Auth token set - tunnel will be automatically associated with your account")
	}
	
	// Request handler is set up in Bubble Tea model

	// If --new flag is set, clear resume info and set forceNew flag to force new tunnel
	if forceNew {
		client.ClearResumeInfo() // This also sets forceNew = true internally
		client.SetForceNew(true)  // Explicitly set forceNew flag
		log.Info().Msg("--new flag set: forcing new tunnel creation (not resuming or auto-finding)")
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
		errStr := err.Error()
		// Check if this is an authentication error - automatically log out
		if strings.Contains(strings.ToLower(errStr), "token") ||
		   strings.Contains(strings.ToLower(errStr), "authentication") ||
		   strings.Contains(strings.ToLower(errStr), "expired") ||
		   strings.Contains(strings.ToLower(errStr), "invalid") {
			// Clear expired/invalid token
			clearExpiredToken()
			log.Warn().Msg("Authentication failed - token cleared. Please run 'uniroute auth login' to authenticate again")
			return fmt.Errorf("authentication failed: %w\n\nPlease run 'uniroute auth login' to authenticate again", err)
		}
		return fmt.Errorf("failed to connect to tunnel server: %w", err)
	}

	// Get tunnel info
	info := client.GetTunnelInfo()
	if info == nil {
		return fmt.Errorf("failed to get tunnel information")
	}

	// Set custom domain if provided
	if customDomain != "" {
		if err := setCustomDomain(info.ID, customDomain, token); err != nil {
			log.Warn().Err(err).Str("domain", customDomain).Msg("Failed to set custom domain (tunnel will still work with subdomain)")
		} else {
			log.Info().Str("domain", customDomain).Msg("Custom domain set successfully")
		}
	}

	// Get account display
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

	// Use Bubble Tea for UI
	return runTunnelWithBubbleTea(client, info, accountDisplay, tunnelServerURL, localURL)
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
			{
				Name:      "dns",
				Protocol:  "udp",
				LocalAddr: "localhost:53",
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
	fmt.Println("Edit the file to configure your tunnels, then run:")
	fmt.Printf("  %s\n", color.Cyan("uniroute tunnel --all"))
	return nil
}

// listLocalTunnels lists tunnels from the local config file
func listLocalTunnels() {
	log := logger.New()
	configManager := tunnel.NewConfigManager(log)
	config, err := configManager.Load()
	if err != nil || config == nil {
		fmt.Println(color.Gray("📁 Local Configuration Tunnels: No configuration file found"))
		fmt.Println()
		return
	}

	if len(config.Tunnels) == 0 {
		fmt.Println(color.Gray("📁 Local Configuration Tunnels: No tunnels defined"))
		fmt.Println()
		return
	}

	fmt.Println(color.Bold("📁 Local Configuration Tunnels"))
	fmt.Println()

	// Table header
	fmt.Printf("  %-20s %-10s %-25s %-10s\n",
		color.Bold("Name"),
		color.Bold("Protocol"),
		color.Bold("Local Address"),
		color.Bold("Status"))
	fmt.Println()

	// Table rows
	for _, tc := range config.Tunnels {
		status := color.Red("Disabled")
		if tc.Enabled {
			status = color.Green("Enabled")
		}
		fmt.Printf("  %-20s %-10s %-25s %s\n",
			tc.Name,
			tc.Protocol,
			tc.LocalAddr,
			status)
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

	// Get tunnel server URL
	serverURL := getTunnelServerURL()
	if serverURL == "" {
		serverURL = "tunnel.uniroute.co"
	}

	// Make API request to list tunnels
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", fmt.Sprintf("http://%s/api/tunnels", serverURL), nil)
	if err != nil {
		fmt.Println(color.Gray("🌐 Server Tunnels: Unable to fetch (server unreachable)"))
		fmt.Println()
		return
	}

	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(color.Gray("🌐 Server Tunnels: Unable to fetch (connection error)"))
		fmt.Println()
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Println(color.Gray("🌐 Server Tunnels: Unable to fetch (authentication error)"))
		fmt.Println()
		return
	}

	var result struct {
		Tunnels []struct {
			ID        string `json:"id"`
			Subdomain string `json:"subdomain"`
			Protocol  string `json:"protocol"`
			Status    string `json:"status"`
		} `json:"tunnels"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Println(color.Gray("🌐 Server Tunnels: Unable to parse response"))
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
	fmt.Printf("  %-20s %-10s %-30s %-10s\n",
		color.Bold("Subdomain"),
		color.Bold("Protocol"),
		color.Bold("Public URL"),
		color.Bold("Status"))
	fmt.Println()

	// Table rows
	for _, t := range result.Tunnels {
		publicURL := fmt.Sprintf("http://%s.localhost:8055", t.Subdomain)
		status := color.Green("Active")
		if t.Status != "active" {
			status = color.Gray(t.Status)
		}
		fmt.Printf("  %-20s %-10s %-30s %s\n",
			t.Subdomain,
			t.Protocol,
			publicURL,
			status)
	}
	fmt.Println()
}

// showTunnelUsage shows usage examples and help
func showTunnelUsage() {
	fmt.Println(color.Bold("📖 Usage Examples"))
	fmt.Println()
	fmt.Printf("  %s\n", color.Cyan("uniroute tunnel"))
	fmt.Println("    Create a new HTTP tunnel on port 8084 (default)")
	fmt.Println()
	fmt.Printf("  %s\n", color.Cyan("uniroute tunnel --port 3000"))
	fmt.Println("    Create a tunnel for local server on port 3000")
	fmt.Println()
	fmt.Printf("  %s\n", color.Cyan("uniroute tunnel --protocol tcp --port 3306"))
	fmt.Println("    Create a TCP tunnel for MySQL")
	fmt.Println()
	fmt.Printf("  %s\n", color.Cyan("uniroute tunnel --host myapp"))
	fmt.Println("    Request a specific subdomain (myapp.localhost:8055)")
	fmt.Println()
	fmt.Printf("  %s\n", color.Cyan("uniroute tunnel --host myapp --new"))
	fmt.Println("    Create new tunnel with specific subdomain (fails if subdomain taken)")
	fmt.Println()
	fmt.Printf("  %s\n", color.Cyan("uniroute tunnel --all"))
	fmt.Println("    Start all enabled tunnels from ~/.uniroute/tunnels.json")
	fmt.Println()
	fmt.Printf("  %s\n", color.Cyan("uniroute tunnel --init"))
	fmt.Println("    Create an example tunnel configuration file")
	fmt.Println()
	fmt.Printf("  %s\n", color.Cyan("uniroute tunnel --list"))
	fmt.Println("    List all tunnels (local config + server subdomains)")
	fmt.Println()
	fmt.Printf("  %s\n", color.Cyan("uniroute tunnel --resume abc123"))
	fmt.Println("    Resume a specific subdomain tunnel")
	fmt.Println()
	fmt.Printf("  %s\n", color.Cyan("uniroute tunnel --domain example.com"))
	fmt.Println("    Set custom domain for tunnel (requires DNS configuration)")
	fmt.Println()
	fmt.Printf("  %s\n", color.Cyan("uniroute tunnel --new"))
	fmt.Println("    Force creating a new tunnel (don't resume saved state)")
	fmt.Println()

	fmt.Println(color.Bold("💡 Tips:"))
	fmt.Printf("   • %s\n", color.Gray("Tunnels automatically resume: Run the same command to get the same subdomain"))
	fmt.Printf("   • %s\n", color.Gray("Saved state: ~/.uniroute/tunnel-state.json (auto-saved on tunnel creation)"))
	fmt.Printf("   • %s\n", color.Gray("Local config: ~/.uniroute/tunnels.json (for multiple tunnels)"))
	fmt.Printf("   • %s\n", color.Gray("Use 'uniroute auth login' to authenticate and see server tunnels"))
	fmt.Println()
}

// setCustomDomain sets a custom domain for a tunnel via API
func setCustomDomain(tunnelID, domain, token string) error {
	// Get API URL from auth config or environment
	var apiURL string
	if envURL := os.Getenv("UNIROUTE_API_URL"); envURL != "" {
		apiURL = envURL
	} else {
		// Try to load from auth config
	homeDir, err := os.UserHomeDir()
		if err == nil {
	configPath := filepath.Join(homeDir, ".uniroute", "auth.json")
			if data, err := os.ReadFile(configPath); err == nil {
	var authConfig struct {
					ServerURL string `json:"server_url"`
				}
				if err := json.Unmarshal(data, &authConfig); err == nil && authConfig.ServerURL != "" {
					apiURL = authConfig.ServerURL
				}
			}
		}
		if apiURL == "" {
			apiURL = "https://api.uniroute.co"
		}
	}
	if apiURL == "" {
		apiURL = "https://api.uniroute.co"
	}

	// Ensure URL has protocol
	if !strings.HasPrefix(apiURL, "http://") && !strings.HasPrefix(apiURL, "https://") {
		apiURL = "https://" + apiURL
	}

	// Create request
	reqBody := map[string]string{"domain": domain}
	jsonData, err := json.Marshal(reqBody)
		if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	// Use /auth/tunnels for JWT auth (frontend/CLI), /v1/tunnels for API keys
	endpoint := fmt.Sprintf("%s/auth/tunnels/%s/domain", apiURL, tunnelID)
	req, err := http.NewRequest("POST", endpoint, strings.NewReader(string(jsonData)))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err == nil {
			if msg, ok := errResp["message"].(string); ok {
				return fmt.Errorf("API error: %s", msg)
			}
			if errMsg, ok := errResp["error"].(string); ok {
				return fmt.Errorf("API error: %s", errMsg)
			}
		}
		return fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	return nil
}
