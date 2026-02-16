package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/Kizsoft-Solution-Limited/uniroute/pkg/color"
	"github.com/spf13/cobra"
)

// devProject holds detected dev server command and port
type devProject struct {
	Name    string   // "Node (Vite)", "Laravel", etc.
	Command string   // executable
	Args    []string // arguments
	Port    string   // port for tunnel
}

var (
	devPort      string // override port
	devDir       string // run from directory (default: cwd)
	devNoTunnel  bool   // start dev server only, no tunnel
	devAttach    bool   // tunnel only: use your own dev server, we just add the public URL
)

var devCmd = &cobra.Command{
	Use:   "dev",
	Short: "Dev server + tunnel, or tunnel-only (attach to your running app)",
	Long: `Three ways to get a public URL:

  1) One command (we start your dev server + tunnel):
     uniroute dev
     -> We run php artisan serve / npm run dev / etc. and the tunnel (auto-detected).

  2) Your normal command + attach (two terminals):
     Terminal 1:  php artisan serve     # or npm run dev, rails s, etc.
     Terminal 2:  uniroute dev --attach  # tunnel only; we auto-detect port.

  3) Your normal command as-is; we run it and add the tunnel (one terminal):
     uniroute run -- php artisan serve
     uniroute run -- npm run dev
     -> You type your usual command; we run it and start the tunnel (port auto-detected).

Supported projects (auto-detected from current directory):
  - Node / Vite / Next / React  package.json + "dev" script  -> port 3000/5173/3002
  - PHP Laravel                 composer.json + artisan      -> port 8000
  - Python Django               manage.py                     -> port 8000
  - Python Flask                requirements.txt + flask      -> port 5000
  - Python FastAPI              requirements.txt + fastapi    -> port 8000 (uvicorn)
  - Go                          go.mod                         -> port 8080 (go run .)
  - Ruby on Rails               Gemfile + config.ru          -> port 3000

Examples:
  uniroute dev                    # Option 1: we start dev server + tunnel
  uniroute dev --attach           # Option 2: tunnel only (you run your command in another terminal)
  uniroute run -- php artisan serve   # Option 3: your command + tunnel (all languages)
  uniroute run -- npm run dev
  uniroute dev --dir ./api        # Target a subdirectory
  uniroute dev --no-tunnel        # Start dev server only (no tunnel)`,
	RunE: runDev,
}

func init() {
	devCmd.Flags().StringVarP(&devPort, "port", "p", "", "Port for tunnel (overrides auto-detected port)")
	devCmd.Flags().StringVar(&devDir, "dir", "", "Project directory (default: current directory)")
	devCmd.Flags().BoolVar(&devNoTunnel, "no-tunnel", false, "Start dev server only, do not start tunnel")
	devCmd.Flags().BoolVar(&devAttach, "attach", false, "Tunnel only; you run your dev server (e.g. php artisan serve) in another terminal")
}

func runDev(cmd *cobra.Command, args []string) error {
	dir := devDir
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

	project, err := detectProject(dir)
	if err != nil {
		if devAttach && devPort != "" {
			project = &devProject{Name: "custom", Port: devPort}
			err = nil
		} else if devAttach {
			return fmt.Errorf("%w\nUse --port to specify the port (e.g. uniroute dev --attach --port 8000)", err)
		} else {
			return err
		}
	}

	port := devPort
	if port == "" {
		port = project.Port
	}

	// Attach mode: only start tunnel (user runs their own dev server)
	if devAttach {
		fmt.Println(color.Cyan("Tunnel only (attach mode). Make sure your dev server is running (e.g. php artisan serve or npm run dev)."))
		fmt.Println(color.Gray("Port: " + port + " (public URL will appear in tunnel UI)"))
		fmt.Println()
		tunnelPort = port
		tunnelProtocol = "http"
		tunnelServerURL = getTunnelServerURL()
		return runBuiltInTunnel(cmd, args)
	}

	fmt.Println(color.Cyan("Detected: " + project.Name))
	fmt.Println(color.Gray("Command: " + project.Command + " " + strings.Join(project.Args, " ")))
	fmt.Println(color.Gray("Tunnel port: " + port + " (public URL will appear in tunnel UI)"))
	fmt.Println()

	devCmdExec := exec.Command(project.Command, project.Args...)
	devCmdExec.Dir = dir
	devCmdExec.Stdout = os.Stdout
	devCmdExec.Stderr = os.Stderr
	devCmdExec.Stdin = os.Stdin

	if err := devCmdExec.Start(); err != nil {
		return fmt.Errorf("start dev server: %w", err)
	}

	defer func() {
		if devCmdExec.Process != nil {
			_ = devCmdExec.Process.Kill()
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		if devCmdExec.Process != nil {
			_ = devCmdExec.Process.Kill()
		}
		os.Exit(0)
	}()

	time.Sleep(2 * time.Second)

	if devNoTunnel {
		fmt.Println(color.Green("Dev server running. Press Ctrl+C to stop."))
		_ = devCmdExec.Wait()
		return nil
	}

	tunnelPort = port
	tunnelProtocol = "http"
	tunnelServerURL = getTunnelServerURL()
	return runBuiltInTunnel(cmd, args)
}

func detectProject(dir string) (*devProject, error) {
	// Node: package.json with dev script
	if b, err := os.ReadFile(filepath.Join(dir, "package.json")); err == nil {
		var pkg struct {
			Scripts         map[string]string `json:"scripts"`
			DevDependencies map[string]string `json:"devDependencies"`
		}
		if json.Unmarshal(b, &pkg) == nil && pkg.Scripts["dev"] != "" {
			port := "3000"
			if pkg.DevDependencies["vite"] != "" {
				port = detectVitePort(dir)
			} else if pkg.DevDependencies["next"] != "" {
				port = "3000"
			}
			return &devProject{
				Name:    "Node (npm run dev)",
				Command: "npm",
				Args:    []string{"run", "dev"},
				Port:    port,
			}, nil
		}
	}

	// PHP Laravel: composer.json + artisan
	if _, err := os.Stat(filepath.Join(dir, "composer.json")); err == nil {
		if _, err := os.Stat(filepath.Join(dir, "artisan")); err == nil {
			return &devProject{
				Name:    "PHP Laravel",
				Command: "php",
				Args:    []string{"artisan", "serve"},
				Port:    "8000",
			}, nil
		}
	}

	// Python Django: manage.py
	if _, err := os.Stat(filepath.Join(dir, "manage.py")); err == nil {
		return &devProject{
			Name:    "Python Django",
			Command: "python",
			Args:    []string{"manage.py", "runserver"},
			Port:    "8000",
		}, nil
	}

	// Python Flask: requirements.txt or pyproject.toml with flask (no manage.py)
	if _, err := os.Stat(filepath.Join(dir, "manage.py")); err != nil {
		if containsDependency(dir, "flask") {
			return &devProject{
				Name:    "Python Flask",
				Command: "flask",
				Args:    []string{"run"},
				Port:    "5000",
			}, nil
		}
	}

	// Python FastAPI: requirements.txt or pyproject.toml with fastapi
	if containsDependency(dir, "fastapi") {
		appSpec := "main:app"
		if _, err := os.Stat(filepath.Join(dir, "app.py")); err == nil {
			appSpec = "app:app"
		}
		return &devProject{
			Name:    "Python FastAPI",
			Command: "uvicorn",
			Args:    []string{appSpec, "--reload"},
			Port:    "8000",
		}, nil
	}

	// Go: go.mod
	if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
		return &devProject{
			Name:    "Go",
			Command: "go",
			Args:    []string{"run", "."},
			Port:    "8080",
		}, nil
	}

	// Ruby on Rails: Gemfile + config.ru
	if _, err := os.Stat(filepath.Join(dir, "Gemfile")); err == nil {
		if _, err := os.Stat(filepath.Join(dir, "config.ru")); err == nil {
			return &devProject{
				Name:    "Ruby on Rails",
				Command: "rails",
				Args:    []string{"s"},
				Port:    "3000",
			}, nil
		}
	}

	return nil, fmt.Errorf("no supported project found in %s\nSupported: Node, Laravel, Django, Flask, FastAPI, Go, Rails (see uniroute dev --help)", dir)
}

// containsDependency checks requirements.txt, pyproject.toml, or Pipfile for a package name.
func containsDependency(dir, pkg string) bool {
	for _, name := range []string{"requirements.txt", "requirements-dev.txt", "pyproject.toml", "Pipfile"} {
		b, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			continue
		}
		content := strings.ToLower(string(b))
		if strings.Contains(content, pkg) {
			return true
		}
	}
	return false
}

func detectVitePort(dir string) string {
	for _, name := range []string{"vite.config.ts", "vite.config.js", "vite.config.mjs"} {
		path := filepath.Join(dir, name)
		b, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		content := string(b)
		if strings.Contains(content, "port:") {
			if strings.Contains(content, "3002") {
				return "3002"
			}
			if strings.Contains(content, "5173") {
				return "5173"
			}
		}
	}
	return "5173"
}
