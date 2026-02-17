# Installation

Install UniRoute CLI on your system. Choose the method that works best for you.

## One-Line Install (Recommended)

The easiest way to install UniRoute is using our installation script:

```bash
curl -fsSL https://raw.githubusercontent.com/Kizsoft-Solution-Limited/uniroute/main/scripts/install.sh | bash
```

This script:
- Auto-detects your OS and architecture
- Downloads the latest release
- Installs to `/usr/local/bin` (or `~/.local/bin` if no sudo)
- Adds to your PATH

## Manual Installation

### macOS

```bash
# Download for macOS
curl -LO https://github.com/Kizsoft-Solution-Limited/uniroute/releases/latest/download/uniroute-darwin-amd64.tar.gz

# Extract
tar -xzf uniroute-darwin-amd64.tar.gz

# Move to PATH
sudo mv uniroute /usr/local/bin/
```

### Linux

```bash
# Download for Linux
curl -LO https://github.com/Kizsoft-Solution-Limited/uniroute/releases/latest/download/uniroute-linux-amd64.tar.gz

# Extract
tar -xzf uniroute-linux-amd64.tar.gz

# Move to PATH
sudo mv uniroute /usr/local/bin/
```

### Windows

```powershell
# Download for Windows
Invoke-WebRequest -Uri "https://github.com/Kizsoft-Solution-Limited/uniroute/releases/latest/download/uniroute-windows-amd64.zip" -OutFile "uniroute.zip"

# Extract
Expand-Archive uniroute.zip

# Add to PATH (manually or via PowerShell)
$env:Path += ";C:\path\to\uniroute"
```

## Build from Source

If you want to build from source:

```bash
# Clone repository
git clone https://github.com/Kizsoft-Solution-Limited/uniroute.git
cd uniroute

# Build
make build

# Install
sudo make install
```

## Verify Installation

```bash
uniroute --version
```

You should see the version number. If you get a "command not found" error, make sure UniRoute is in your PATH.

## Next Steps

- [Authentication](/docs/authentication) - Set up your account
- [Getting Started](/docs/getting-started) - Create your first tunnel
