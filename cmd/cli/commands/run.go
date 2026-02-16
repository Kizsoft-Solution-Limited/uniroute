package commands

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/Kizsoft-Solution-Limited/uniroute/pkg/color"
	"github.com/spf13/cobra"
)

var (
	runPort string
	runDir  string
)

var runCmd = &cobra.Command{
	Use:   "run [flags] -- <command> [args...]",
	Short: "Run your normal command and auto-start the tunnel (Option 3)",
	Long: `Run your usual dev command; we start it and the tunnel together. Port is auto-detected from the project.

  Terminal:  uniroute run -- php artisan serve
  Terminal:  uniroute run -- npm run dev
  Terminal:  uniroute run -- rails s

  We run your command and start the tunnel. Port is taken from:
  1) Your command (e.g. php artisan serve --port=8080  -> tunnel to 8080)
  2) Our --port flag
  3) Auto-detected from project (Laravel 8000, Vite 5173, etc.)

Examples:
  uniroute run -- php artisan serve --port=8080   # Tunnel to 8080 (from your command)
  uniroute run -- php artisan serve               # Tunnel to 8000 (Laravel default)
  uniroute run -- npm run dev -- --port 3000     # Tunnel to 3000 (from your command)
  uniroute run -- rails s -p 3001                 # Tunnel to 3001`,
	RunE: runRun,
}

func init() {
	runCmd.Flags().StringVarP(&runPort, "port", "p", "", "Tunnel port (default: auto-detected from project)")
	runCmd.Flags().StringVar(&runDir, "dir", "", "Project directory (default: current directory)")
}

func runRun(cmd *cobra.Command, args []string) error {
	// Parse command: everything after "--", or all args
	var userCmd []string
	for i, a := range args {
		if a == "--" {
			userCmd = args[i+1:]
			break
		}
	}
	if len(userCmd) == 0 {
		userCmd = args
	}
	if len(userCmd) == 0 {
		return fmt.Errorf("provide a command after -- (e.g. uniroute run -- php artisan serve)")
	}

	dir := runDir
	if dir == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("get working directory: %w", err)
		}
		dir = cwd
	}
	dir, err := filepath.Abs(dir)
	if err != nil {
		return fmt.Errorf("resolve directory: %w", err)
	}

	port := runPort
	if port == "" {
		port = parsePortFromCommand(userCmd)
	}
	if port == "" {
		project, err := detectProject(dir)
		if err != nil {
			return fmt.Errorf("%w\nUse --port or pass port in your command (e.g. php artisan serve --port=8080)", err)
		}
		port = project.Port
	}

	fmt.Println(color.Cyan("Run (Option 3): your command + tunnel"))
	fmt.Println(color.Gray("Command: " + strings.Join(userCmd, " ")))
	fmt.Println(color.Gray("Tunnel port: " + port + " (public URL will appear in tunnel UI)"))
	fmt.Println()

	userExec := exec.Command(userCmd[0], userCmd[1:]...)
	userExec.Dir = dir
	userExec.Stdout = os.Stdout
	userExec.Stderr = os.Stderr
	userExec.Stdin = os.Stdin

	if err := userExec.Start(); err != nil {
		return fmt.Errorf("start command: %w", err)
	}

	defer func() {
		if userExec.Process != nil {
			_ = userExec.Process.Kill()
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		if userExec.Process != nil {
			_ = userExec.Process.Kill()
		}
		os.Exit(0)
	}()

	time.Sleep(2 * time.Second)

	tunnelPort = port
	tunnelProtocol = "http"
	tunnelServerURL = getTunnelServerURL()
	return runBuiltInTunnel(cmd, args)
}

// parsePortFromCommand extracts port from common flags in the user's command.
// Supports: --port=8080, --port 8080, -p 8080, --port:8080 (Laravel-style).
func parsePortFromCommand(args []string) string {
	for i, a := range args {
		a = strings.TrimSpace(a)
		// --port=8080 or --port:8080
		if strings.HasPrefix(a, "--port=") || strings.HasPrefix(a, "--port:") {
			port := strings.TrimPrefix(a, "--port=")
			port = strings.TrimPrefix(port, "--port:")
			if _, err := strconv.Atoi(port); err == nil {
				return port
			}
		}
		// --port 8080 or -p 8080 (next arg is the port)
		if (a == "--port" || a == "-p") && i+1 < len(args) {
			port := strings.TrimSpace(args[i+1])
			if _, err := strconv.Atoi(port); err == nil {
				return port
			}
		}
	}
	return ""
}
