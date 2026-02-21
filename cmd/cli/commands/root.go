package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	version = "1.0.0"
	rootAll bool
	rootCmd = &cobra.Command{
		Use:   "uniroute",
		Short: "UniRoute - Unified gateway for every AI model",
		Long: `UniRoute is a unified gateway platform that routes, secures, and manages 
traffic to any LLM (cloud or local) with one unified API.

One unified gateway for every AI model. Route, secure, and manage traffic 
to any LLM—cloud or local—with one unified platform.`,
		Version: version,
		RunE: func(cmd *cobra.Command, args []string) error {
			if rootAll {
				return runAllTunnels(cmd, args)
			}
			if isLocalMode() {
				fmt.Println("🚀 UniRoute CLI - Local Development Mode")
				fmt.Println()
				fmt.Println("Local mode detected. Available commands:")
				fmt.Println("  uniroute tunnel --all    Start all configured tunnels")
				fmt.Println("  uniroute status          Check gateway server status")
				fmt.Println("  uniroute auth login      Authenticate (optional in local mode)")
				fmt.Println()
				fmt.Println("Gateway: http://localhost:8084")
				fmt.Println("Tunnel Server: localhost:8080")
				fmt.Println()
				fmt.Println("Run 'uniroute --help' for all commands")
				return nil
			}
			return cmd.Help()
		},
	}
)

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.Flags().BoolVar(&rootAll, "all", false, "Start all configured tunnels")
	rootCmd.AddCommand(authCmd)
	rootCmd.AddCommand(projectsCmd)
	rootCmd.AddCommand(devCmd)
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(tunnelCmd)
	rootCmd.AddCommand(httpCmd)
	rootCmd.AddCommand(tcpCmd)
	rootCmd.AddCommand(tlsCmd)
	rootCmd.AddCommand(udpCmd)
	rootCmd.AddCommand(resumeCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(keysCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(logsCmd)
	rootCmd.AddCommand(upgradeCmd)
	rootCmd.AddCommand(domainCmd)
}

func SetVersion(v string) {
	version = v
	rootCmd.Version = v
}

func GetVersion() string {
	return version
}

func PrintError(err error) {
	fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	os.Exit(1)
}

