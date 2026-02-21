# CLI Reference

Complete reference for the UniRoute command-line interface.

## Installation

```bash
curl -fsSL https://raw.githubusercontent.com/Kizsoft-Solution-Limited/uniroute/main/scripts/install.sh | bash
```

## Authentication

```bash
# Login (default: hosted; use --live for hosted, --local for local server)
uniroute auth login

# Logout
uniroute auth logout

# Check status
uniroute auth status
```

When you don't pass `--server`, `--local`, or `--live`, the CLI uses `UNIROUTE_API_URL` (if set), then saved server from last login, then hosted (https://app.uniroute.co).

## Tunnels

### Shortcuts (Recommended)

```bash
uniroute http 8080    # HTTP tunnel
uniroute tcp 3306     # TCP tunnel
uniroute tls 5432     # TLS tunnel
uniroute udp 53       # UDP tunnel
```

### Create Tunnels

```bash
# Create HTTP tunnel
uniroute http 8080

# Create TCP tunnel
uniroute tcp 3306

# Create TLS tunnel
uniroute tls 5432

# Create UDP tunnel
uniroute udp 53

# Force new tunnel (don't resume)
uniroute http 8080 --new

# Create tunnel with specific subdomain (shortcut syntax - recommended)
uniroute http 8080 myapp
uniroute http 8080 myapp --new
uniroute tcp 3306 mydb
uniroute tcp 3306 mydb --new
uniroute tls 5432 mydb
uniroute tls 5432 mydb --new
uniroute udp 53 dns
uniroute udp 53 dns --new

# Create tunnel with specific subdomain (flag syntax - also works)
uniroute http 8080 --host myapp
uniroute http 8080 --host myapp --new
uniroute tcp 3306 --host mydb
uniroute tcp 3306 --host mydb --new
uniroute tls 5432 --host mydb
uniroute tls 5432 --host mydb --new
uniroute udp 53 --host dns
uniroute udp 53 --host dns --new

# Start all tunnels from config
uniroute tunnel --all
```

### Manage Tunnels

```bash
# List all tunnels (shortcut - recommended)
uniroute list

# Resume specific tunnel (shortcut - recommended)
uniroute resume abc123

# Create tunnel configuration
uniroute tunnel --init

# Clear saved tunnel state
uniroute tunnel --clear
```

### Creating Tunnels (Shortcuts)

```bash
# HTTP tunnel (shortcut - recommended)
uniroute http 8080

# TCP tunnel (shortcut - recommended)
uniroute tcp 3306

# TLS tunnel (shortcut - recommended)
uniroute tls 5432

# Request specific subdomain (shortcut syntax - recommended)
uniroute http 8080 myapp
uniroute http 8080 myapp --new
uniroute tcp 3306 mydb
uniroute tcp 3306 mydb --new

# Request specific subdomain (flag syntax - also works)
uniroute http 8080 --host myapp
uniroute http 8080 --host myapp --new

# Set custom domain (shortcut - recommended)
uniroute domain example.com
uniroute domain example.com abc123              # Shortcut: domain + subdomain
uniroute domain example.com --subdomain abc123  # Flag syntax: domain + subdomain flag
```

> 💡 **Tip:** Tunnels automatically resume their previous subdomain when you run the same command again.

## API Keys

```bash
# List keys
uniroute keys list

# Create key
uniroute keys create --name "My App"

# Delete key
uniroute keys delete <key-id>
```

## Domain Management

### Adding and Assigning Domains

```bash
# Add domain to account (not assigned to any tunnel yet)
uniroute domain example.com

# Add domain AND assign to tunnel by subdomain (shortcut syntax - recommended)
uniroute domain example.com abc123

# Add domain AND assign to tunnel (flag syntax - also works)
uniroute domain example.com --subdomain abc123

# Add domain AND assign to specific tunnel by tunnel ID
uniroute domain example.com --tunnel-id <id>
```

### Domain Management Commands

```bash
# List all your custom domains
uniroute domain list

# Show details for a specific domain
uniroute domain show example.com

# Verify DNS configuration
uniroute domain verify example.com

# Resume domain assignment (restore previous assignment)
uniroute domain resume                    # Resume last used assignment
uniroute domain resume abc123             # Resume by subdomain
uniroute domain resume example.com        # Resume by domain name

# Remove domain from account
uniroute domain remove example.com
```

> 💡 **Tip:** Domain assignments are automatically saved. Use `uniroute domain resume` to restore them later.

## Global Options

```bash
--help, -h     Show help
--version, -v  Show version
```

## Next Steps

- [Getting Started](/docs/getting-started) - Learn the basics
- [Tunnels](/docs/tunnels) - Tunnel commands
