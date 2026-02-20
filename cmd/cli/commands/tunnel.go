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

	defaultTunnelServer := getTunnelServerURL()
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

	if createConfig {
		return createExampleConfig(log)
	}

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

	if startAll {
		return runAllTunnels(cmd, args)
	}

	return runBuiltInTunnel(cmd, args)
}

func runAllTunnels(cmd *cobra.Command, args []string) error {
	log := logger.New()
	configManager := tunnel.NewConfigManager(log)

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

	wg.Wait()
	close(errors)

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

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	fmt.Println()
	fmt.Println(color.Yellow("Shutting down all tunnels..."))
	return nil
}

func startTunnelFromConfig(tc tunnel.TunnelConfig) error {
	log := logger.New()

	var localURL string
	if tc.Protocol == "http" {
		if !strings.HasPrefix(tc.LocalAddr, "http://") && !strings.HasPrefix(tc.LocalAddr, "https://") {
			localURL = "http://" + tc.LocalAddr
		} else {
			localURL = tc.LocalAddr
		}
	} else {
		localURL = tc.LocalAddr
	}

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

	client := tunnel.NewTunnelClientWithOptions(serverURL, localURL, tc.Protocol, tc.Host, log)

	if err := client.Connect(); err != nil {
		errStr := err.Error()
		if strings.Contains(strings.ToLower(errStr), "token") ||
		   strings.Contains(strings.ToLower(errStr), "authentication") ||
		   strings.Contains(strings.ToLower(errStr), "expired") ||
		   strings.Contains(strings.ToLower(errStr), "invalid") {
			clearExpiredToken()
			log.Warn().Msg("Authentication failed - token cleared. Please run 'uniroute auth login' to authenticate again")
			return fmt.Errorf("authentication failed: %w\n\nPlease run 'uniroute auth login' to authenticate again", err)
		}
		return fmt.Errorf("failed to connect: %w", err)
	}

	info := client.GetTunnelInfo()
	if info == nil {
		return fmt.Errorf("failed to get tunnel information")
	}

	fmt.Printf("      %s %s\n", color.Gray("Public URL:"), color.Cyan(info.PublicURL))

	return nil
}

func runBuiltInTunnel(cmd *cobra.Command, args []string) error {
	if tunnelServerURL == "" || 
		tunnelServerURL == "tunnel.uniroute.co" || 
		tunnelServerURL == "https://tunnel.uniroute.co" {
		tunnelServerURL = getTunnelServerURL()
	}

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

	if tunnelProtocol != "http" && tunnelProtocol != "tcp" && tunnelProtocol != "tls" && tunnelProtocol != "udp" {
		return fmt.Errorf("invalid protocol '%s'. Must be: http, tcp, tls, or udp", tunnelProtocol)
	}

	var localURL string
	if tunnelProtocol == "http" {
		localURL = fmt.Sprintf("http://localhost:%s", tunnelPort)
	} else {
		localURL = fmt.Sprintf("localhost:%s", tunnelPort)
	}

	client := tunnel.NewTunnelClientWithOptions(tunnelServerURL, localURL, tunnelProtocol, tunnelHost, log)

	token := getAuthToken()
	if token != "" {
		client.SetToken(token)
		log.Debug().Msg("Auth token set - tunnel will be automatically associated with your account")
	}

	if forceNew {
		client.ClearResumeInfo()
		client.SetForceNew(true)
		log.Info().Msg("--new flag set: forcing new tunnel creation (not resuming or auto-finding)")
	}

	if resumeSubdomain != "" {
		persistence := tunnel.NewTunnelPersistence(log)
		if state, err := persistence.Load(); err == nil && state != nil {
			if state.Subdomain == resumeSubdomain {
				client.SetResumeInfo(resumeSubdomain, state.TunnelID)
				log.Info().
					Str("subdomain", resumeSubdomain).
					Str("tunnel_id", state.TunnelID).
					Msg("Resuming tunnel with saved tunnel ID")
			} else {
				client.SetResumeInfo(resumeSubdomain, "")
				log.Info().
					Str("subdomain", resumeSubdomain).
					Msg("Resuming tunnel by subdomain")
			}
		} else {
			client.SetResumeInfo(resumeSubdomain, "")
			log.Info().
				Str("subdomain", resumeSubdomain).
				Msg("Resuming tunnel by subdomain (no saved state)")
		}
	}

	if err := client.Connect(); err != nil {
		errStr := err.Error()
		if strings.Contains(strings.ToLower(errStr), "token") ||
		   strings.Contains(strings.ToLower(errStr), "authentication") ||
		   strings.Contains(strings.ToLower(errStr), "expired") ||
		   strings.Contains(strings.ToLower(errStr), "invalid") {
			clearExpiredToken()
			log.Warn().Msg("Authentication failed - token cleared. Please run 'uniroute auth login' to authenticate again")
			return fmt.Errorf("authentication failed: %w\n\nPlease run 'uniroute auth login' to authenticate again", err)
		}
		return fmt.Errorf("failed to connect to tunnel server: %w", err)
	}

	info := client.GetTunnelInfo()
	if info == nil {
		return fmt.Errorf("failed to get tunnel information")
	}

	if customDomain != "" {
		if err := setCustomDomain(info.ID, customDomain, token); err != nil {
			log.Warn().Err(err).Str("domain", customDomain).Msg("Failed to set custom domain (tunnel will still work with subdomain)")
		} else {
			log.Info().Str("domain", customDomain).Msg("Custom domain set successfully")
		}
	}

	accountDisplay := "Free"
	configPath := filepath.Join(tunnel.GetConfigDir(), "auth.json")
	if data, err := os.ReadFile(configPath); err == nil {
		var authConfig struct {
			Email string `json:"email"`
		}
		if err := json.Unmarshal(data, &authConfig); err == nil && authConfig.Email != "" {
			accountDisplay = authConfig.Email + " (Plan: Free)"
		}
	}

	return runTunnelWithBubbleTea(client, info, accountDisplay, tunnelServerURL, localURL)
}

func listAllTunnels() error {
	fmt.Println()
	fmt.Println(color.Cyan("📋 UniRoute Tunnels"))
	fmt.Println()

	listLocalTunnels()
	listServerTunnels()
	showTunnelUsage()

	return nil
}

func createExampleConfig(log zerolog.Logger) error {
	configManager := tunnel.NewConfigManager(log)
	configPath := configManager.GetConfigPath()

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

	fmt.Printf("  %-20s %-10s %-25s %-10s\n",
		color.Bold("Name"),
		color.Bold("Protocol"),
		color.Bold("Local Address"),
		color.Bold("Status"))
	fmt.Println()

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

func listServerTunnels() {
	token := getAuthToken()
	if token == "" {
		fmt.Println(color.Gray("🌐 Server Tunnels: Not authenticated (run 'uniroute auth login' to see your subdomain tunnels)"))
		fmt.Println()
		return
	}

	serverURL := getTunnelServerURL()
	if serverURL == "" {
		serverURL = "tunnel.uniroute.co"
	}

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

	fmt.Printf("  %-20s %-10s %-30s %-10s\n",
		color.Bold("Subdomain"),
		color.Bold("Protocol"),
		color.Bold("Public URL"),
		color.Bold("Status"))
	fmt.Println()

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

func setCustomDomain(tunnelID, domain, token string) error {
	var apiURL string
	if envURL := os.Getenv("UNIROUTE_API_URL"); envURL != "" {
		apiURL = envURL
	} else {
		configPath := filepath.Join(tunnel.GetConfigDir(), "auth.json")
		if data, err := os.ReadFile(configPath); err == nil {
			var authConfig struct {
				ServerURL string `json:"server_url"`
			}
			if err := json.Unmarshal(data, &authConfig); err == nil && authConfig.ServerURL != "" {
				apiURL = authConfig.ServerURL
			}
		}
		if apiURL == "" {
			apiURL = "https://app.uniroute.co"
		}
	}
	if apiURL == "" {
		apiURL = "https://app.uniroute.co"
	}

	if !strings.HasPrefix(apiURL, "http://") && !strings.HasPrefix(apiURL, "https://") {
		apiURL = "https://" + apiURL
	}

	reqBody := map[string]string{"domain": domain}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}
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
