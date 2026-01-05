package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	version = "1.0.0"
	rootCmd = &cobra.Command{
		Use:   "uniroute",
		Short: "UniRoute - Unified gateway for every AI model",
		Long: `UniRoute is a unified gateway platform that routes, secures, and manages 
traffic to any LLM (cloud or local) with one unified API.

One unified gateway for every AI model. Route, secure, and manage traffic 
to any LLM—cloud or local—with one unified platform.`,
		Version: version,
	}
)

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(authCmd)
	rootCmd.AddCommand(projectsCmd)
	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(tunnelCmd)
	rootCmd.AddCommand(keysCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(logsCmd)
	rootCmd.AddCommand(upgradeCmd)
}

// SetVersion sets the version for the CLI
func SetVersion(v string) {
	version = v
	rootCmd.Version = v
}

// GetVersion returns the current version
func GetVersion() string {
	return version
}

// PrintError prints an error message and exits
func PrintError(err error) {
	fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	os.Exit(1)
}

