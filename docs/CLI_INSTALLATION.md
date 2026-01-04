# CLI Installation Guide

## Overview

The UniRoute CLI (`uniroute`) connects to the **public UniRoute server** to manage your projects, API keys, tunnels, and more. Everything is managed in one place - no need to run your own server!

**Key Features:**
- ✅ **Single Source of Truth** - All your data is managed by UniRoute
- ✅ **User Authentication** - Login once, access all your projects
- ✅ **Project Management** - View and manage all your projects
- ✅ **API Key Management** - Create and manage API keys for your projects
- ✅ **Tunnel Management** - Create tunnels to expose local apps
- ✅ **No Server Setup** - Everything connects to the public UniRoute server

## Installation Methods

### Method 1: Quick Download Script (Easiest) ⭐

```bash
# Download and run the install script
curl -fsSL https://raw.githubusercontent.com/Kizsoft-Solution-Limited/uniroute/main/scripts/download-cli.sh | bash

# Or download the script first
curl -L https://raw.githubusercontent.com/Kizsoft-Solution-Limited/uniroute/main/scripts/download-cli.sh -o download-cli.sh
chmod +x download-cli.sh
./download-cli.sh
```

### Method 2: Download Pre-built Binary (Manual)

Download the latest release for your platform:

**macOS (Apple Silicon):**
```bash
# Download
curl -L https://github.com/Kizsoft-Solution-Limited/uniroute/releases/latest/download/uniroute-darwin-arm64 -o uniroute

# Make executable
chmod +x uniroute

# Move to PATH (optional)
sudo mv uniroute /usr/local/bin/
```

**macOS (Intel):**
```bash
curl -L https://github.com/Kizsoft-Solution-Limited/uniroute/releases/latest/download/uniroute-darwin-amd64 -o uniroute
chmod +x uniroute
sudo mv uniroute /usr/local/bin/
```

**Linux (AMD64):**
```bash
curl -L https://github.com/Kizsoft-Solution-Limited/uniroute/releases/latest/download/uniroute-linux-amd64 -o uniroute
chmod +x uniroute
sudo mv uniroute /usr/local/bin/
```

**Linux (ARM64):**
```bash
curl -L https://github.com/Kizsoft-Solution-Limited/uniroute/releases/latest/download/uniroute-linux-arm64 -o uniroute
chmod +x uniroute
sudo mv uniroute /usr/local/bin/
```

**Windows:**
```powershell
# Download
Invoke-WebRequest -Uri https://github.com/Kizsoft-Solution-Limited/uniroute/releases/latest/download/uniroute-windows-amd64.exe -OutFile uniroute.exe

# Add to PATH or use directly
.\uniroute.exe tunnel
```

**Verify installation:**
```bash
uniroute --version
```

### Method 2: Build from Source

If pre-built binaries aren't available or you want to build yourself:

```bash
# Clone the repository
git clone https://github.com/Kizsoft-Solution-Limited/uniroute.git
cd uniroute

# Build the CLI
make build

# The binary will be at: ./bin/uniroute
```

### Method 3: Go Install

Once the project is published, you'll be able to install directly:

```bash
go install github.com/Kizsoft-Solution-Limited/uniroute/cmd/cli@latest
```

## Getting Started

### 1. Authenticate

First, login to your UniRoute account:

```bash
uniroute auth login
# Enter your email and password when prompted
```

Or with flags:
```bash
uniroute auth login --email user@example.com --password yourpassword
```

### 2. View Your Projects

See all your projects:
```bash
uniroute projects list
```

### 3. Manage API Keys

Create API keys for your projects:
```bash
uniroute keys create --name "My API Key"
```

## Using the Tunnel Command

### Prerequisites

1. **Authentication**: You must be logged in (see above)
   - Run `uniroute auth login` first
   - Your tunnels are associated with your account

2. **Tunnel Server**: Uses the public UniRoute tunnel server automatically
   - Default: `tunnel.uniroute.dev`
   - No setup needed!

3. **Local Application**: Any local application you want to expose
   - Your local web server (e.g., `http://localhost:3000`)
   - Your local API (e.g., `http://localhost:8080`)
   - Any service running on your machine

### Basic Usage

```bash
# Tunnel your local application (default port 8084)
# Automatically resumes your previous subdomain if available!
uniroute tunnel

# Tunnel a specific port
uniroute tunnel --port 3000

# Resume a specific subdomain
uniroute tunnel --resume abc123

# Use custom tunnel server (if you have your own)
uniroute tunnel --port 8084 --server your-tunnel-server.com

# List saved tunnel state
uniroute tunnel --list

# Clear saved tunnel state (create new subdomain next time)
uniroute tunnel --clear
```

### Resuming Tunnels

**Automatic Resume (Recommended):**
```bash
# Just run the tunnel command again - it automatically resumes your saved subdomain!
uniroute tunnel --port 8084
# Uses the same subdomain from your last session
```

**Manual Resume:**
```bash
# Resume a specific subdomain
uniroute tunnel --resume abc123 --port 8084

# Check what subdomain is saved
uniroute tunnel --list
```

**Create New Tunnel:**
```bash
# Clear saved state to get a new subdomain
uniroute tunnel --clear
uniroute tunnel --port 8084  # Creates new subdomain
```

### Common Use Cases

**1. Expose a Local Web App:**
```bash
# Your app runs on localhost:3000
uniroute tunnel --port 3000

# Get public URL: http://abc123.uniroute.dev
# Share with others!
```

**2. Expose a Local API:**
```bash
# Your API runs on localhost:8080
uniroute tunnel --port 8080

# Access from anywhere: http://xyz789.uniroute.dev
```

**3. Expose UniRoute Gateway (if running locally):**
```bash
# If you're running UniRoute gateway locally
./bin/uniroute-gateway  # Runs on port 8084

# In another terminal, expose it
uniroute tunnel --port 8084

# Now your local UniRoute is accessible publicly!
```

### Full Example

```bash
# Start your local application (any app, any port)
# Example: A Node.js app on port 3000
npm start  # Runs on http://localhost:3000

# In another terminal, create tunnel
uniroute tunnel --port 3000

# You'll see output like:
# Tunnel Connected Successfully!
# Public URL: http://abc123.uniroute.dev
# Forwarding: http://abc123.uniroute.dev -> http://localhost:3000
```

### Using Your Own Tunnel Server (Advanced)

If you want to run your own tunnel server:

```bash
# Terminal 1: Start tunnel server
./bin/uniroute-tunnel-server --port 8080

# Terminal 2: Create tunnel to your local app
uniroute tunnel --port 3000 --server localhost:8080
```

## All CLI Commands

### Authentication
```bash
uniroute auth login          # Login to your account
uniroute auth logout         # Logout
uniroute auth status         # Check login status
```

### Projects
```bash
uniroute projects list       # List all your projects
uniroute projects show --id PROJECT_ID  # Show project details
```

### API Keys
```bash
uniroute keys create         # Create a new API key (requires login)
uniroute keys create --name "My Key"  # Create named key
```

### Tunnels
```bash
uniroute tunnel              # Create tunnel (default port 8084)
uniroute tunnel --port 3000  # Tunnel specific port
uniroute tunnel --list       # List your tunnels
```

### Status
```bash
uniroute status              # Check public server status
```

> **Note**: Most commands require authentication. Run `uniroute auth login` first!

## Troubleshooting

### "Command not found: uniroute"

If you get this error, make sure:
1. You've built the CLI: `make build`
2. The binary exists: `ls -la bin/uniroute`
3. You're using the full path: `./bin/uniroute` or it's in your PATH

### "Tunnel server connection failed"

Check that:
1. The tunnel server is running: `curl http://localhost:8080/health`
2. The server URL is correct: `--server localhost:8080`
3. No firewall is blocking the connection

### "Port already in use"

If you see port conflicts:
1. Check what's using the port: `lsof -i :8084`
2. Use a different port: `--port 8085`
3. Stop the conflicting service

## Alternative: Use cloudflared (No Installation Required)

If you don't want to install the CLI, you can use cloudflared instead:

```bash
# Install cloudflared (one-time)
# macOS: brew install cloudflared
# Linux: Download from https://github.com/cloudflare/cloudflared/releases

# Use it directly
cloudflared tunnel --url http://localhost:8084
```

This is 100% free and requires no signup, but doesn't provide the same features as the built-in tunnel (like subdomain persistence, reconnection, etc.).

## Next Steps

- See [QUICKSTART.md](./QUICKSTART.md) for getting started
- See [TUNNEL_LOCAL_TESTING.md](./TUNNEL_LOCAL_TESTING.md) for detailed tunnel testing
- See [README.md](./README.md) for full documentation

