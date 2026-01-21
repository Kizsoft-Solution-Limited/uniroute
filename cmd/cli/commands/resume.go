package commands

import (
	"strings"

	"github.com/Kizsoft-Solution-Limited/uniroute/internal/tunnel"
	"github.com/Kizsoft-Solution-Limited/uniroute/pkg/logger"
	"github.com/spf13/cobra"
)

var resumeCmd = &cobra.Command{
	Use:   "resume [subdomain]",
	Short: "Resume a tunnel by subdomain (shortcut for 'tunnel --resume')",
	Long: `Resume an existing tunnel by its subdomain.

This is a shortcut command for: uniroute tunnel --resume [subdomain]

When resuming, it will use the same LocalURL (port) as when the tunnel was created,
unless you specify a different port with --port.

Examples:
  uniroute resume abc123              # Resume tunnel with subdomain abc123 (uses saved port)
  uniroute resume abc123 --port 8080  # Resume with specific port (overrides saved port)
  uniroute resume abc123 --protocol tcp --port 3306  # Resume with different protocol`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Set resume subdomain
		resumeSubdomain = args[0]
		
		// Try to load saved state to get the LocalURL (port) from when tunnel was created
		// This ensures resume uses the same port as the original tunnel
		log := logger.New()
		persistence := tunnel.NewTunnelPersistence(log)
		if state, err := persistence.Load(); err == nil && state != nil {
			if state.Subdomain == resumeSubdomain && state.LocalURL != "" {
				// Extract port from saved LocalURL
				// Format: http://localhost:8009 or localhost:8009
				savedLocalURL := state.LocalURL
				if strings.HasPrefix(savedLocalURL, "http://") {
					savedLocalURL = strings.TrimPrefix(savedLocalURL, "http://")
				} else if strings.HasPrefix(savedLocalURL, "https://") {
					savedLocalURL = strings.TrimPrefix(savedLocalURL, "https://")
				}
				
				// Extract port (format: localhost:8009 or 127.0.0.1:8009)
				if strings.Contains(savedLocalURL, ":") {
					parts := strings.Split(savedLocalURL, ":")
					if len(parts) == 2 {
						savedPort := parts[1]
						// Check if user explicitly set --port flag
						// If port is still the default (8084), use saved port instead
						// This ensures resume uses the same port as when tunnel was created
						portFlag := cmd.Flags().Lookup("port")
						userSetPort := portFlag != nil && portFlag.Changed
						
						if !userSetPort {
							// User didn't specify --port, use saved port from when tunnel was created
							tunnelPort = savedPort
						}
					}
				}
			}
		}
		
		// Call the tunnel command handler
		return runBuiltInTunnel(cmd, args)
	},
}

func init() {
	// Add port and protocol flags to resume command
	resumeCmd.Flags().StringVarP(&tunnelPort, "port", "p", "8084", "Local port to tunnel")
	resumeCmd.Flags().StringVar(&tunnelProtocol, "protocol", "http", "Tunnel protocol: http, tcp, or tls")
	resumeCmd.Flags().StringVar(&tunnelHost, "host", "", "Request specific host/subdomain")
	resumeCmd.Flags().StringVarP(&tunnelServerURL, "server", "s", "tunnel.uniroute.co", "Tunnel server URL")
	
	rootCmd.AddCommand(resumeCmd)
}
