package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

var tlsCmd = &cobra.Command{
	Use:   "tls [port] [host]",
	Short: "Create TLS tunnel (shortcut for 'tunnel --protocol tls')",
	Long: `Create a TLS tunnel to expose your local TLS service.

This is a shortcut command for: uniroute tunnel --protocol tls --port [port]

Examples:
  uniroute tls 5432              # Tunnel PostgreSQL with TLS on port 5432 (auto-resumes if available)
  uniroute tls 443               # Tunnel HTTPS service on port 443
  uniroute tls 5432 --new        # Force new tunnel (don't resume)
  uniroute tls 5432 mydb         # Request specific subdomain (mydb.uniroute.co) - shortcut
  uniroute tls 5432 mydb --new   # Create new tunnel with specific subdomain - shortcut
  uniroute tls 5432 --host mydb  # Request specific subdomain (mydb.uniroute.co) - flag syntax
  uniroute tls 5432 --host mydb --new  # Create new tunnel with specific subdomain - flag syntax`,
	Args: cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		tunnelProtocol = "tls"
		if len(args) == 0 {
			return fmt.Errorf("port is required for TLS tunnel")
		}
		tunnelPort = args[0]
		if len(args) > 1 && tunnelHost == "" {
			tunnelHost = args[1]
		}
		return runBuiltInTunnel(cmd, args)
	},
}

func init() {
	tlsCmd.Flags().BoolVar(&forceNew, "new", false, "Force creating a new tunnel (don't resume saved state)")
	tlsCmd.Flags().StringVar(&tunnelHost, "host", "", "Request specific host/subdomain (reserved subdomain)")
}
