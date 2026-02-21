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

On Mac and Linux, a downloaded binary is **not executable** until you run `chmod +x`. Do that before running or moving the file.

**If the file is in your Downloads folder** (e.g. `uniroute-darwin-amd64` or `uniroute-darwin-arm64`):

```bash
cd ~/Downloads
chmod +x uniroute-darwin-amd64   # or uniroute-darwin-arm64 for Apple Silicon
sudo mv uniroute-darwin-amd64 /usr/local/bin/uniroute
uniroute --version
```

### macOS

**Apple Silicon (M1/M2/M3):**
```bash
curl -LO https://github.com/Kizsoft-Solution-Limited/uniroute/releases/latest/download/uniroute-darwin-arm64
chmod +x uniroute-darwin-arm64
sudo mv uniroute-darwin-arm64 /usr/local/bin/uniroute
```

**Intel:**
```bash
curl -LO https://github.com/Kizsoft-Solution-Limited/uniroute/releases/latest/download/uniroute-darwin-amd64
chmod +x uniroute-darwin-amd64
sudo mv uniroute-darwin-amd64 /usr/local/bin/uniroute
```

### Linux

```bash
curl -LO https://github.com/Kizsoft-Solution-Limited/uniroute/releases/latest/download/uniroute-linux-amd64
chmod +x uniroute-linux-amd64
sudo mv uniroute-linux-amd64 /usr/local/bin/uniroute
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
