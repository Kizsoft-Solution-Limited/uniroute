package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

var tcpCmd = &cobra.Command{
	Use:   "tcp [port]",
	Short: "Create TCP tunnel (shortcut for 'tunnel --protocol tcp')",
	Long: `Create a TCP tunnel to expose your local TCP service.

This is a shortcut command for: uniroute tunnel --protocol tcp --port [port]

Examples:
  uniroute tcp 3306         # Tunnel MySQL on port 3306 (auto-resumes if available)
  uniroute tcp 5432         # Tunnel PostgreSQL on port 5432
  uniroute tcp 3306 --new   # Force new tunnel (don't resume)`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Set protocol to tcp
		tunnelProtocol = "tcp"
		
		// Set port (required for TCP)
		if len(args) == 0 {
			return fmt.Errorf("port is required for TCP tunnel")
		}
		tunnelPort = args[0]
		
		// Call the tunnel command handler
		return runBuiltInTunnel(cmd, args)
	},
}

func init() {
	tcpCmd.Flags().BoolVar(&forceNew, "new", false, "Force creating a new tunnel (don't resume saved state)")
}
