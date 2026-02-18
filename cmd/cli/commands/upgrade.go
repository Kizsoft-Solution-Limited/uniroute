package commands

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/Kizsoft-Solution-Limited/uniroute/pkg/color"
	versioncheck "github.com/Kizsoft-Solution-Limited/uniroute/pkg/version"
	"github.com/spf13/cobra"
)

var (
	upgradeAutoYes bool
	upgradeCmd = &cobra.Command{
		Use:   "upgrade",
		Short: "Upgrade UniRoute CLI to the latest version",
		Long: `Upgrade UniRoute CLI to the latest version.

This command checks for updates and upgrades the CLI tool.
On macOS/Linux, it uses the same installation method you used initially.
On Windows, it downloads the latest release.`,
		RunE: runUpgrade,
	}
)

func init() {
	upgradeCmd.Flags().BoolVarP(&upgradeAutoYes, "yes", "y", false, "Auto-confirm upgrade without prompting")
}

func runUpgrade(cmd *cobra.Command, args []string) error {
	currentVersion := GetVersion()
	
	fmt.Println()
	fmt.Println(color.Cyan("Checking for updates..."))
	fmt.Println()
	
	// Default to GitHub releases API; override with UNIROUTE_VERSION_URL
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
	
	var upgradeCmd []string
	var upgradeInstructions string
	
	goInstallVersion := info.LatestVersion
	if goInstallVersion == "" {
		goInstallVersion = "latest"
	} else if !strings.HasPrefix(goInstallVersion, "v") {
		goInstallVersion = "v" + goInstallVersion
	}
	goInstallCmd := []string{"go", "install", fmt.Sprintf("github.com/Kizsoft-Solution-Limited/uniroute/cmd/cli@%s", goInstallVersion)}
	goInstallInstructions := fmt.Sprintf("Run: go install github.com/Kizsoft-Solution-Limited/uniroute/cmd/cli@%s", goInstallVersion)

	switch runtime.GOOS {
	case "darwin":
		if _, err := exec.LookPath("brew"); err == nil {
			brewList, _ := exec.Command("brew", "list", "--formula").Output()
			if strings.Contains(string(brewList), "uniroute") {
				upgradeCmd = []string{"brew", "upgrade", "uniroute"}
				upgradeInstructions = "Run: brew upgrade uniroute"
			}
		}
		if upgradeCmd == nil {
			if _, err := exec.LookPath("go"); err == nil {
				upgradeCmd = goInstallCmd
				upgradeInstructions = goInstallInstructions
			} else {
				upgradeInstructions = goInstallInstructions
			}
		}

	case "linux":
		if _, err := exec.LookPath("snap"); err == nil {
			snapList, _ := exec.Command("snap", "list", "uniroute").Output()
			if strings.Contains(string(snapList), "uniroute") {
				upgradeCmd = []string{"sudo", "snap", "refresh", "uniroute"}
				upgradeInstructions = "Run: sudo snap refresh uniroute"
			}
		}
		if upgradeCmd == nil {
			if _, err := exec.LookPath("go"); err == nil {
				upgradeCmd = goInstallCmd
				upgradeInstructions = goInstallInstructions
			} else if _, err := exec.LookPath("apt"); err == nil {
				upgradeInstructions = "Run: sudo apt update && sudo apt upgrade uniroute"
			} else if _, err := exec.LookPath("yum"); err == nil {
				upgradeInstructions = "Run: sudo yum update uniroute"
			} else {
				upgradeInstructions = goInstallInstructions
			}
		}

	case "windows":
		if _, err := exec.LookPath("go"); err == nil {
			upgradeCmd = goInstallCmd
			upgradeInstructions = goInstallInstructions
		} else {
			upgradeInstructions = fmt.Sprintf("Download the latest release from: %s", info.ReleaseURL)
			if info.ReleaseURL == "" {
				upgradeInstructions = "Visit: https://github.com/Kizsoft-Solution-Limited/uniroute/releases/latest"
			}
		}
	}
	
	fmt.Println(color.Cyan("To upgrade:"))
	if upgradeCmd != nil {
		fmt.Printf("   %s\n", color.Bold(strings.Join(upgradeCmd, " ")))
		fmt.Println()
		
		// --yes skips prompt (used when called from tunnel UI)
		shouldRun := upgradeAutoYes
		if !shouldRun {
			fmt.Print("Run the command above? (y/n): ")
			var response string
			fmt.Scanln(&response)
			shouldRun = strings.ToLower(response) == "y" || strings.ToLower(response) == "yes"
		}
		
		if shouldRun {
			c := exec.Command(upgradeCmd[0], upgradeCmd[1:]...)
			c.Stdout = os.Stdout
			c.Stderr = os.Stderr
			if err := c.Run(); err != nil {
				return fmt.Errorf("upgrade failed: %w", err)
			}
			fmt.Println()
			fmt.Println(color.Green("✓ Upgrade completed successfully!"))
			if len(upgradeCmd) > 0 && upgradeCmd[0] == "go" {
				currentExe, _ := os.Executable()
				currentExe, _ = filepath.EvalSymlinks(currentExe)
				if out, err := exec.Command("go", "env", "GOPATH").Output(); err == nil {
					gopath := strings.TrimSpace(string(out))
					if gopath != "" {
						newBin := filepath.Join(gopath, "bin", "uniroute")
						if currentExe != newBin {
							if _, err := os.Stat(newBin); err == nil {
								fmt.Println(color.Yellow("  You're running the CLI from a different location. To use the new version, run:"))
								fmt.Printf("  %s\n", color.Bold(fmt.Sprintf("sudo cp %s %s", newBin, currentExe)))
							}
						}
					}
				}
			}
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

