package commands

import (
	"strings"

	"github.com/Kizsoft-Solution-Limited/uniroute/internal/tunnel"
	"github.com/Kizsoft-Solution-Limited/uniroute/pkg/logger"
	"github.com/spf13/cobra"
)

var resumeCmd = &cobra.Command{
	Use:   "resume [subdomain]",
	Short: "Resume a tunnel by subdomain (no --protocol or --port needed)",
	Long: `Resume an existing tunnel by its subdomain.

This is a shortcut command for: uniroute tunnel --resume [subdomain]

Uses saved protocol and port for that subdomain, so no --protocol or --port needed.

Examples:
  uniroute resume 84b13ab7                  # Resume (uses saved protocol and port)
  uniroute resume 84b13ab7 --port 8080      # Resume with different port
  uniroute resume abc123 --protocol tcp     # Override protocol`,
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
				protocolFlag := cmd.Flags().Lookup("protocol")
				userSetProtocol := protocolFlag != nil && protocolFlag.Changed
				if !userSetProtocol && state.Protocol != "" {
					tunnelProtocol = state.Protocol
				}
			}
		}
		return runBuiltInTunnel(cmd, args)
	},
}

func init() {
	resumeCmd.Flags().StringVarP(&tunnelPort, "port", "p", "8084", "Local port to tunnel")
	resumeCmd.Flags().StringVar(&tunnelProtocol, "protocol", "http", "Tunnel protocol: http, tcp, tls, or udp")
	resumeCmd.Flags().StringVar(&tunnelHost, "host", "", "Request specific host/subdomain")
	resumeCmd.Flags().StringVarP(&tunnelServerURL, "server", "s", "tunnel.uniroute.co", "Tunnel server URL")
}
