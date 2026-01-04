# 🚀 UniRoute Quick Start Guide

## Phase 1: Getting Started

### Prerequisites

1. **Go 1.21+** installed
2. **Ollama** installed and running (for local LLM support)
   ```bash
   # Install Ollama: https://ollama.ai
   # Start Ollama
   ollama serve
   
   # Pull a model (optional, for testing)
   ollama pull llama2
   ```

### Setup

**Option 1: Download Pre-built Binary (Recommended for CLI)**

```bash
# Download CLI for your platform
# macOS (Apple Silicon):
curl -L https://github.com/Kizsoft-Solution-Limited/uniroute/releases/latest/download/uniroute-darwin-arm64 -o uniroute
chmod +x uniroute
sudo mv uniroute /usr/local/bin/

# Verify
uniroute --version
```

**Option 2: Build from Source**

1. **Clone and navigate to project:**
   ```bash
   git clone https://github.com/Kizsoft-Solution-Limited/uniroute.git
   cd uniroute
   ```

2. **Install dependencies:**
   ```bash
   make deps
   # or
   go mod download
   ```

3. **Build binaries:**
   ```bash
   make build
   ```

4. **Set up environment (optional):**
   ```bash
   cp .env.example .env
   # Edit .env if needed (defaults work for local development)
   ```

See [CLI_INSTALLATION.md](./CLI_INSTALLATION.md) for all installation options.

### Running the Server

```bash
make dev
```

The server will:
- Start on `http://localhost:8084`
- Generate a default API key (save this!)
- Register the local LLM provider
- Be ready to accept requests

### Testing

1. **Health Check:**
   ```bash
   curl http://localhost:8084/health
   # Should return: {"status":"ok"}
   ```

2. **Chat Request (with API key from server output):**
   ```bash
   curl -X POST http://localhost:8084/v1/chat \
     -H "Authorization: Bearer YOUR_API_KEY_HERE" \
     -H "Content-Type: application/json" \
     -d '{
       "model": "llama2",
       "messages": [
         {"role": "user", "content": "Hello! Say hi back."}
       ]
     }'
   ```

### Sharing Your Server

**Option 1: Using cloudflared (Recommended - 100% Free, No Installation)**
```bash
# Install cloudflared (one-time)
# macOS: brew install cloudflared
# Linux: Download from https://github.com/cloudflare/cloudflared/releases

# In another terminal
cloudflared tunnel --url http://localhost:8084
# Returns: https://random-subdomain.trycloudflare.com
# Share the public URL with others
# ✅ 100% free, no signup required, no time limits
```

**Option 2: Using ngrok (Free tier, requires signup)**
```bash
# Install ngrok: https://ngrok.com/download
# In another terminal
ngrok http 8084
# Returns: https://abc123.ngrok-free.app
# Share the public URL with others
```

**Option 3: Built-in UniRoute Tunnel (Requires CLI Installation)**
```bash
# Download CLI (see CLI_INSTALLATION.md)
# Or build: make build

# Expose your local app (works with any app, any port)
uniroute tunnel --port 8084
# Returns: http://{subdomain}.uniroute.dev -> http://localhost:8084
# ✅ Features: Subdomain persistence, auto-reconnection, statistics
# ✅ Uses public UniRoute tunnel server (no setup needed!)
```

**Option 4: Local network**
```bash
# Access from other machines on same network
http://YOUR_LOCAL_IP:8084
```

> **Note**: For the built-in tunnel, you need to build the CLI first. See [CLI_INSTALLATION.md](./CLI_INSTALLATION.md) for detailed instructions.

### Project Structure

```
uniroute/
├── cmd/gateway/          # Main application entry point
├── internal/
│   ├── api/              # HTTP handlers and middleware
│   ├── gateway/          # Request routing logic
│   ├── providers/        # LLM provider implementations
│   ├── security/         # Authentication
│   └── config/           # Configuration
├── pkg/                  # Shared utilities
└── Makefile              # Common commands
```

### Common Commands

```bash
make dev          # Start development server
make build        # Build binary
make test         # Run tests
make fmt          # Format code
make lint         # Run linters
make vet          # Run go vet
make security     # Security scan
```

### Next Steps

- ✅ Phase 1 complete: Local LLM provider working
- ⏭️ Phase 2: Add security enhancements (JWT, rate limiting)
- ⏭️ Phase 3: Add cloud providers (OpenAI, Anthropic, Google)

See [START_HERE.md](./START_HERE.md) for complete documentation.

