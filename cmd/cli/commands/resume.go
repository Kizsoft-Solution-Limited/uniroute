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
		resumeSubdomain = args[0]
		log := logger.New()
		persistence := tunnel.NewTunnelPersistence(log)
		if state, err := persistence.Load(); err == nil && state != nil {
			if state.Subdomain == resumeSubdomain && state.LocalURL != "" {
				savedLocalURL := state.LocalURL
				if strings.HasPrefix(savedLocalURL, "http://") {
					savedLocalURL = strings.TrimPrefix(savedLocalURL, "http://")
				} else if strings.HasPrefix(savedLocalURL, "https://") {
					savedLocalURL = strings.TrimPrefix(savedLocalURL, "https://")
				}
				if strings.Contains(savedLocalURL, ":") {
					parts := strings.Split(savedLocalURL, ":")
					if len(parts) == 2 {
						savedPort := parts[1]
						portFlag := cmd.Flags().Lookup("port")
						userSetPort := portFlag != nil && portFlag.Changed
						if !userSetPort {
							tunnelPort = savedPort
						}
					}
				}
			}
		}
		return runBuiltInTunnel(cmd, args)
	},
}

func init() {
	resumeCmd.Flags().StringVarP(&tunnelPort, "port", "p", "8084", "Local port to tunnel")
	resumeCmd.Flags().StringVar(&tunnelProtocol, "protocol", "http", "Tunnel protocol: http, tcp, or tls")
	resumeCmd.Flags().StringVar(&tunnelHost, "host", "", "Request specific host/subdomain")
	resumeCmd.Flags().StringVarP(&tunnelServerURL, "server", "s", "tunnel.uniroute.co", "Tunnel server URL")
	
	rootCmd.AddCommand(resumeCmd)
}
