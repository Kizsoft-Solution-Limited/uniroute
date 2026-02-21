# Getting Started

Get up and running with UniRoute in minutes. This guide will walk you through the basics.

## Quick Start

### 1. Install UniRoute CLI

```bash
# One-line installation
curl -fsSL https://raw.githubusercontent.com/Kizsoft-Solution-Limited/uniroute/main/scripts/install.sh | bash

# Verify installation
uniroute --version
```

### 2. Authenticate

```bash
# Login (default: hosted UniRoute for new installs)
uniroute auth login

# Use hosted explicitly
uniroute auth login --live

# Use local server (e.g. self-hosted)
uniroute auth login --local
```

When you don't pass `--server`, `--local`, or `--live`, the CLI uses `UNIROUTE_API_URL` (if set), then your last saved server, then the hosted server.

### 3. Create Your First Tunnel

```bash
# Expose a local web server (shortcut - recommended)
uniroute http 8080

# You'll get a public URL like: http://abc123.localhost:8055
```

### 4. Make Your First API Request

```bash
# Using curl
curl -X POST https://app.uniroute.co/v1/chat \
  -H "Authorization: Bearer ur_your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4o",
    "messages": [{"role": "user", "content": "Hello!"}]
  }'
```

## What's Next?

- Learn about [Tunnels](/docs/tunnels) - Expose local services
- Explore [API Reference](/docs/api) - Make AI requests
- Configure [Routing](/docs/routing) - Optimize provider selection
- Review [Security](/docs/security) - Secure your setup
