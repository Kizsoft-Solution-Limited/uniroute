package commands

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/Kizsoft-Solution-Limited/uniroute/pkg/color"
	versioncheck "github.com/Kizsoft-Solution-Limited/uniroute/pkg/version"
	"github.com/spf13/cobra"
)

var (
	upgradeCmd     *cobra.Command
	upgradeAutoYes bool
)

func init() {
	upgradeCmd = &cobra.Command{
		Use:   "upgrade",
		Short: "Upgrade UniRoute CLI to the latest version",
		Long: `Upgrade UniRoute CLI to the latest version.

This command checks for updates and upgrades the CLI tool.
On macOS/Linux, it uses the same installation method you used initially.
On Windows, it downloads the latest release.`,
		RunE: runUpgrade,
	}
	upgradeCmd.Flags().BoolVarP(&upgradeAutoYes, "yes", "y", false, "Auto-confirm upgrade without prompting")
}

func runUpgrade(cmd *cobra.Command, args []string) error {
	currentVersion := GetVersion()
	
	fmt.Println()
	fmt.Println(color.Cyan("Checking for updates..."))
	fmt.Println()
	
	// Check for updates
	// Default to GitHub releases API, but can be configured via env var
	versionURL := os.Getenv("UNIROUTE_VERSION_URL")
	if versionURL == "" {
		versionURL = "https://api.github.com/repos/Kizsoft-Solution-Limited/uniroute/releases/latest"
	}
	
	checker := versioncheck.NewChecker(versionURL)
	info, err := checker.CheckForUpdate(currentVersion)
	if err != nil {
		fmt.Println(color.Yellow("⚠️  Could not check for updates. Please check your internet connection."))
		return nil
	}
	
	if !info.UpdateAvailable {
		fmt.Println(color.Green(fmt.Sprintf("✓ You're using the latest version: %s", currentVersion)))
		return nil
	}
	
	fmt.Println(color.Yellow(fmt.Sprintf("📦 New version available: %s (current: %s)", info.LatestVersion, currentVersion)))
	fmt.Println()
	
	// Determine upgrade command based on OS and installation method
	var upgradeCmd []string
	var upgradeInstructions string
	
	switch runtime.GOOS {
	case "darwin": // macOS
		// Check if installed via Homebrew
		if _, err := exec.LookPath("brew"); err == nil {
			// Check if uniroute is a brew formula
			brewList, _ := exec.Command("brew", "list", "--formula").Output()
			if strings.Contains(string(brewList), "uniroute") {
				upgradeCmd = []string{"brew", "upgrade", "uniroute"}
				upgradeInstructions = "Run: brew upgrade uniroute"
			}
		}
		
		// If not Homebrew, check if installed via go install
		if upgradeCmd == nil {
			upgradeInstructions = "Run: go install github.com/Kizsoft-Solution-Limited/uniroute/cmd/cli@latest"
		}
		
	case "linux":
		// Check installation method
		if _, err := exec.LookPath("snap"); err == nil {
			// Check if installed via snap
			snapList, _ := exec.Command("snap", "list", "uniroute").Output()
			if strings.Contains(string(snapList), "uniroute") {
				upgradeCmd = []string{"sudo", "snap", "refresh", "uniroute"}
				upgradeInstructions = "Run: sudo snap refresh uniroute"
			}
		}
		
		// Check if installed via package manager
		if upgradeCmd == nil {
			// Try common package managers
			if _, err := exec.LookPath("apt"); err == nil {
				upgradeInstructions = "Run: sudo apt update && sudo apt upgrade uniroute"
			} else if _, err := exec.LookPath("yum"); err == nil {
				upgradeInstructions = "Run: sudo yum update uniroute"
			} else {
				// Default to go install
				upgradeInstructions = "Run: go install github.com/Kizsoft-Solution-Limited/uniroute/cmd/cli@latest"
			}
		}
		
	case "windows":
		// Windows - download latest release
		upgradeInstructions = fmt.Sprintf("Download the latest release from: %s", info.ReleaseURL)
		if info.ReleaseURL == "" {
			upgradeInstructions = "Visit: https://github.com/Kizsoft-Solution-Limited/uniroute/releases/latest"
		}
	}
	
	fmt.Println(color.Cyan("To upgrade:"))
	if upgradeCmd != nil {
		fmt.Printf("   %s\n", color.Bold(strings.Join(upgradeCmd, " ")))
		fmt.Println()
		
		// Auto-confirm if --yes flag is set (used when called from tunnel UI)
		shouldRun := upgradeAutoYes
		if !shouldRun {
			fmt.Print("Run the command above? (y/n): ")
			var response string
			fmt.Scanln(&response)
			shouldRun = strings.ToLower(response) == "y" || strings.ToLower(response) == "yes"
		}
		
		if shouldRun {
			cmd := exec.Command(upgradeCmd[0], upgradeCmd[1:]...)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("upgrade failed: %w", err)
			}
			fmt.Println()
			fmt.Println(color.Green("✓ Upgrade completed successfully!"))
			return nil
		}
	} else {
		fmt.Printf("   %s\n", color.Bold(upgradeInstructions))
	}
	
	if info.ReleaseURL != "" {
		fmt.Println()
		fmt.Printf("   %s %s\n", color.Gray("Release notes:"), color.Cyan(info.ReleaseURL))
	}
	
	return nil
}

