package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

var tcpCmd = &cobra.Command{
	Use:   "tcp [port] [host]",
	Short: "Create TCP tunnel (shortcut for 'tunnel --protocol tcp')",
	Long: `Create a TCP tunnel to expose your local TCP service.

This is a shortcut command for: uniroute tunnel --protocol tcp --port [port]

Examples:
  uniroute tcp 3306              # Tunnel MySQL on port 3306 (auto-resumes if available)
  uniroute tcp 5432              # Tunnel PostgreSQL on port 5432
  uniroute tcp 3306 --new        # Force new tunnel (don't resume)
  uniroute tcp 3306 mydb         # Request specific subdomain (mydb.uniroute.co) - shortcut
  uniroute tcp 3306 mydb --new   # Create new tunnel with specific subdomain - shortcut
  uniroute tcp 3306 --host mydb  # Request specific subdomain (mydb.uniroute.co) - flag syntax
  uniroute tcp 3306 --host mydb --new  # Create new tunnel with specific subdomain - flag syntax`,
	Args: cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		tunnelProtocol = "tcp"
		if len(args) == 0 {
			return fmt.Errorf("port is required for TCP tunnel")
		}
		tunnelPort = args[0]
		if len(args) > 1 && tunnelHost == "" {
			tunnelHost = args[1]
		}
		return runBuiltInTunnel(cmd, args)
	},
}

func init() {
	tcpCmd.Flags().BoolVar(&forceNew, "new", false, "Force creating a new tunnel (don't resume saved state)")
	tcpCmd.Flags().StringVar(&tunnelHost, "host", "", "Request specific host/subdomain (reserved subdomain)")
}
