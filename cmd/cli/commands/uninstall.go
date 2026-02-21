package commands

import (
	"fmt"
	"os"

	"github.com/Kizsoft-Solution-Limited/uniroute/internal/tunnel"
	"github.com/Kizsoft-Solution-Limited/uniroute/pkg/color"
	"github.com/spf13/cobra"
)

var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Uninstall UniRoute CLI and remove config",
	Long: `Remove the UniRoute CLI binary and config directory (~/.uniroute).

Config (auth, tunnel state) is removed. If the binary was installed to a system path
(e.g. /usr/local/bin), you may need to run the uninstall script with sudo, or remove
the binary manually:

  curl -fsSL https://raw.githubusercontent.com/Kizsoft-Solution-Limited/uniroute/main/scripts/install.sh | bash -s uninstall`,
	RunE: runUninstall,
}

func init() {
	rootCmd.AddCommand(uninstallCmd)
}

func runUninstall(cmd *cobra.Command, args []string) error {
	configDir := tunnel.GetConfigDir()
	if err := os.RemoveAll(configDir); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove config: %w", err)
	}
	fmt.Printf("   %s Config removed: %s\n", color.Green("✓"), color.Gray(configDir))

	execPath, err := os.Executable()
	if err != nil {
		fmt.Println()
		fmt.Println(color.Green("Config removed. To remove the binary, run:"))
		fmt.Println("  curl -fsSL https://raw.githubusercontent.com/Kizsoft-Solution-Limited/uniroute/main/scripts/install.sh | bash -s uninstall")
		return nil
	}
	if err := os.Remove(execPath); err != nil {
		fmt.Printf("   %s Binary not removed (try with sudo): %s\n", color.Yellow("!"), color.Gray(execPath))
		fmt.Println()
		fmt.Println(color.Yellow("To remove the binary, run:"))
		fmt.Println("  curl -fsSL https://raw.githubusercontent.com/Kizsoft-Solution-Limited/uniroute/main/scripts/install.sh | bash -s uninstall")
		return nil
	}
	fmt.Printf("   %s Binary removed: %s\n", color.Green("✓"), color.Gray(execPath))
	fmt.Println()
	fmt.Println(color.Green("UniRoute CLI uninstalled."))
	fmt.Println(color.Gray("The 'uniroute' command will stop working after this process exits."))
	return nil
}
