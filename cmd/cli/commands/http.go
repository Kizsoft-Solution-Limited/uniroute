package commands

import (
	"github.com/spf13/cobra"
)

var httpCmd = &cobra.Command{
	Use:   "http [port] [host]",
	Short: "Create HTTP tunnel (shortcut for 'tunnel --protocol http')",
	Long: `Create an HTTP tunnel to expose your local web server.

This is a shortcut command for: uniroute tunnel --protocol http --port [port]

Examples:
  uniroute http 8080                    # Tunnel port 8080 via HTTP (auto-resumes if available)
  uniroute http                         # Tunnel default port 8084 via HTTP
  uniroute http 8080 --new              # Force new tunnel (don't resume)
  uniroute http 8080 myapp              # Request specific subdomain (myapp.uniroute.co) - shortcut
  uniroute http 8080 myapp --new        # Create new tunnel with specific subdomain - shortcut
  uniroute http 8080 --host myapp       # Request specific subdomain (myapp.uniroute.co) - flag syntax
  uniroute http 8080 --host myapp --new # Create new tunnel with specific subdomain - flag syntax`,
	Args: cobra.MaximumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Set protocol to http
		tunnelProtocol = "http"
		
		// Set port if provided (first argument)
		if len(args) > 0 {
			tunnelPort = args[0]
		}
		
		// Set host if provided as second positional argument (shortcut syntax)
		// Only use if --host flag is not set (flag takes precedence)
		if len(args) > 1 && tunnelHost == "" {
			tunnelHost = args[1]
		}
		
		// Reset tunnel server URL to force auto-detection
		// This ensures local mode is detected even if tunnel command was initialized earlier
		tunnelServerURL = ""
		
		// Call the tunnel command handler
		return runBuiltInTunnel(cmd, args)
	},
}

func init() {
	httpCmd.Flags().BoolVar(&forceNew, "new", false, "Force creating a new tunnel (don't resume saved state)")
	httpCmd.Flags().StringVar(&tunnelHost, "host", "", "Request specific host/subdomain (reserved subdomain)")
}
