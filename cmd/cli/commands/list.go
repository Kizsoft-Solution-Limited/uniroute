package commands

import (
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all tunnels (shortcut for 'tunnel --list')",
	Long: `List all tunnels from local configuration and server.

This is a shortcut command for: uniroute tunnel --list

Shows:
  • Local Configuration Tunnels: Tunnels defined in ~/.uniroute/tunnels.json
  • Server Tunnels: Your active subdomain tunnels on the server (requires authentication)`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return listAllTunnels()
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
