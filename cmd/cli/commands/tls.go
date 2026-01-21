package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

var tlsCmd = &cobra.Command{
	Use:   "tls [port]",
	Short: "Create TLS tunnel (shortcut for 'tunnel --protocol tls')",
	Long: `Create a TLS tunnel to expose your local TLS service.

This is a shortcut command for: uniroute tunnel --protocol tls --port [port]

Examples:
  uniroute tls 5432         # Tunnel PostgreSQL with TLS on port 5432 (auto-resumes if available)
  uniroute tls 443          # Tunnel HTTPS service on port 443
  uniroute tls 5432 --new   # Force new tunnel (don't resume)`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Set protocol to tls
		tunnelProtocol = "tls"
		
		// Set port (required for TLS)
		if len(args) == 0 {
			return fmt.Errorf("port is required for TLS tunnel")
		}
		tunnelPort = args[0]
		
		// Call the tunnel command handler
		return runBuiltInTunnel(cmd, args)
	},
}

func init() {
	tlsCmd.Flags().BoolVar(&forceNew, "new", false, "Force creating a new tunnel (don't resume saved state)")
}
