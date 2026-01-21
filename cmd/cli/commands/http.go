package commands

import (
	"github.com/spf13/cobra"
)

var httpCmd = &cobra.Command{
	Use:   "http [port]",
	Short: "Create HTTP tunnel (shortcut for 'tunnel --protocol http')",
	Long: `Create an HTTP tunnel to expose your local web server.

This is a shortcut command for: uniroute tunnel --protocol http --port [port]

Examples:
  uniroute http 8080        # Tunnel port 8080 via HTTP (auto-resumes if available)
  uniroute http             # Tunnel default port 8084 via HTTP
  uniroute http 8080 --new  # Force new tunnel (don't resume)`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Set protocol to http
		tunnelProtocol = "http"
		
		// Set port if provided
		if len(args) > 0 {
			tunnelPort = args[0]
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
}
