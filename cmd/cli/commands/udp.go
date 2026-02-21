package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

var udpCmd = &cobra.Command{
	Use:   "udp [port] [host]",
	Short: "Create UDP tunnel (shortcut for 'tunnel --protocol udp')",
	Long: `Create a UDP tunnel to expose your local UDP service.

This is a shortcut command for: uniroute tunnel --protocol udp --port [port]

Examples:
  uniroute udp 53              # Tunnel DNS server on port 53 (auto-resumes if available)
  uniroute udp 1194            # Tunnel OpenVPN on port 1194
  uniroute udp 53 --new        # Force new tunnel (don't resume)
  uniroute udp 53 dns          # Request specific subdomain (dns.uniroute.co) - shortcut
  uniroute udp 53 dns --new     # Create new tunnel with specific subdomain - shortcut
  uniroute udp 53 --host dns   # Request specific subdomain (dns.uniroute.co) - flag syntax
  uniroute udp 53 --host dns --new  # Create new tunnel with specific subdomain - flag syntax`,
	Args: cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		tunnelProtocol = "udp"
		if len(args) == 0 {
			return fmt.Errorf("port is required for UDP tunnel")
		}
		tunnelPort = args[0]
		if len(args) > 1 && tunnelHost == "" {
			tunnelHost = args[1]
		}
		return runBuiltInTunnel(cmd, args)
	},
}

func init() {
	udpCmd.Flags().BoolVar(&forceNew, "new", false, "Force creating a new tunnel (don't resume saved state)")
	udpCmd.Flags().StringVar(&tunnelHost, "host", "", "Request specific host/subdomain (reserved subdomain)")
}
