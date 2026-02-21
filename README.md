# 🚀 UniRoute

**One unified gateway for every AI model. Route, secure, and manage traffic to any LLM—cloud or local—with one unified platform. Built-in tunneling for exposing local services to the internet.**

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8.svg)](https://golang.org/)
[![Status](https://img.shields.io/badge/Status-In%20Development-orange)](https://github.com/Kizsoft-Solution-Limited/uniroute)

**Open Source • Self-Hostable • Managed Service Available**

---

## 📖 Overview

UniRoute is a unified gateway platform that routes, secures, and manages traffic to any LLM (cloud or local) with one unified API. A single entry point that intelligently routes requests to the best available model.

### Unified API

UniRoute provides a **single, consistent API interface** for all LLM providers. Instead of learning different APIs for OpenAI, Anthropic, Google, and local models, you use one unified endpoint:

- **Single Endpoint**: `/v1/chat` works with all providers
- **Consistent Format**: Same request/response format across all providers
- **Automatic Routing**: UniRoute intelligently routes to the best available model
- **Provider Abstraction**: Switch providers without changing your code
- **Multi-Provider Support**: Use multiple providers simultaneously with failover

This unified approach means you can:
- Build applications that work with any LLM provider
- Switch between providers without code changes
- Use local models (Ollama, vLLM) alongside cloud providers
- Implement intelligent routing based on cost, latency, or availability
- Scale across multiple providers automatically

### Why UniRoute?

- ✅ **Open Source** - Full transparency, community-driven
- ✅ **Self-Hostable** - Deploy anywhere, full control (100% free)
- ✅ **Managed Service** - UniRoute handles provider keys, unified billing (pay-as-you-go)
- ✅ **Local LLM First** - Priority support for local models (Ollama, vLLM) - Free & Private
- ✅ **Multi-Provider** - Support for OpenAI, Anthropic, Google, Local LLMs, and more
- ✅ **Shareable** - Host and share your gateway with others (built-in tunneling)
- ✅ **Intelligent Routing** - Load balancing, failover, cost-based routing
- ✅ **Enterprise Security** - API keys, rate limiting, Zero Trust support

---

## 🚀 Quick Start

### Prerequisites

- Go 1.21+
- PostgreSQL 15+ (or [Supabase](https://supabase.com) free tier)
- Redis 7+ (or [Upstash](https://upstash.com) free tier)
- SMTP Server (for email verification and password reset) - [Mailtrap](https://mailtrap.io) (free tier) recommended for development
- Docker (optional, for containerized deployment)

### Installation

**Option 1: One-Line Install (Easiest)** ⭐

```bash
# One-line installation (auto-detects OS and architecture)
curl -fsSL https://raw.githubusercontent.com/Kizsoft-Solution-Limited/uniroute/main/scripts/install.sh | bash

# Verify
uniroute --version

# Login to your account

## Email/Password Login
```bash
uniroute auth login
# Or with email flag
uniroute auth login --email user@example.com
```

## Choosing the server (hosted vs local)
When you don't pass `--server`, `--local`, or `--live`, the CLI picks the server in this order: **`UNIROUTE_API_URL`** env var → **saved server** from your last login → **hosted** (https://app.uniroute.co). New installs default to hosted.

Switch explicitly:
- **Hosted:** `uniroute auth login --live` (or `uniroute auth login --server https://app.uniroute.co`)
- **Local:** `uniroute auth login --local` (or `uniroute auth login --server http://localhost:8084`)

## API Key Login (Recommended for Automation)
API keys provide longer sessions without expiration, making them ideal for automation and CI/CD:
```bash
# Login with API key
uniroute auth login --api-key ur_xxxxxxxxxxxxx
# or using short flag
uniroute auth login -k ur_xxxxxxxxxxxxx
# use hosted explicitly
uniroute auth login -k ur_xxx --live
# use local server
uniroute auth login -k ur_xxx --local
```

**Benefits of API Key Login:**
- ✅ No expiration (unlike JWT tokens)
- ✅ Perfect for automation and scripts
- ✅ Longer sessions for CI/CD pipelines
- ✅ Same authentication as email/password

# Start using UniRoute!
uniroute projects list
```

**Option 2: Manual Download**

```bash
# Download for your platform (see CLI_INSTALLATION.md for all platforms)
# macOS (Apple Silicon):
curl -L https://github.com/Kizsoft-Solution-Limited/uniroute/releases/latest/download/uniroute-darwin-arm64 -o uniroute
chmod +x uniroute
sudo mv uniroute /usr/local/bin/
```

**CLI Environment Variables** (Recommended for local development):

```bash
# Set API server URL (when you don't use --server/--local/--live: env overrides saved config, then default is hosted)
# Use BASE_URL from .env or set explicitly:
export UNIROUTE_API_URL=${BASE_URL:-http://localhost:8084}

# Set tunnel server URL (default: auto-detects local mode or uses tunnel.uniroute.co)
export UNIROUTE_TUNNEL_URL=localhost:8080

# Enable local development mode
export UNIROUTE_ENV=local

# Then login (will use localhost automatically)

## Email/Password Login
```bash
uniroute auth login
```

## API Key Login (Longer Sessions)
```bash
uniroute auth login --api-key ur_xxxxxxxxxxxxx
# or
uniroute auth login -k ur_xxxxxxxxxxxxx
```

**Note:** API keys provide longer sessions without expiration, making them ideal for automation.
```

**Option 2: Build from Source**

```bash
# Clone the repository
git clone https://github.com/Kizsoft-Solution-Limited/uniroute.git
cd uniroute

# Install dependencies
go mod download

# Install OAuth2 package (required for Google/GitHub/X login)
go get golang.org/x/oauth2

# Build binaries
make build

# Set up environment
cp .env.example .env
# Edit .env with your configuration (see Environment Variables below)

# Run database migrations
make migrate

# Start the server
make dev
```

### Environment Variables

Create a `.env` file in the root directory with the following variables:

```bash
# Server Configuration
PORT=8084
ENV=development
FRONTEND_URL=http://localhost:3000

# Database
DATABASE_URL=postgresql://user:password@localhost:5432/uniroute?sslmode=disable

# Redis (for rate limiting)
REDIS_URL=redis://localhost:6379

# Security
API_KEY_SECRET=your-secret-key-min-32-chars
JWT_SECRET=your-jwt-secret-min-32-chars
PROVIDER_KEY_ENCRYPTION_KEY=your-encryption-key-32-chars

# SMTP Configuration (for email verification and password reset)
SMTP_HOST=sandbox.smtp.mailtrap.io
SMTP_PORT=2525
SMTP_USERNAME=your-mailtrap-username
SMTP_PASSWORD=your-mailtrap-password
SMTP_FROM=noreply@uniroute.co

# Optional: Cloud Provider API Keys (server-level, fallback for BYOK)
OPENAI_API_KEY=your-openai-key
ANTHROPIC_API_KEY=your-anthropic-key
GOOGLE_API_KEY=your-google-key

# Optional: OAuth Configuration (for Google, GitHub, and X/Twitter login)
GOOGLE_OAUTH_CLIENT_ID=your-google-oauth-client-id
GOOGLE_OAUTH_CLIENT_SECRET=your-google-oauth-client-secret
GITHUB_OAUTH_CLIENT_ID=your-github-oauth-client-id
GITHUB_OAUTH_CLIENT_SECRET=your-github-oauth-client-secret
X_OAUTH_CLIENT_ID=your-x-oauth-client-id
X_OAUTH_CLIENT_SECRET=your-x-oauth-client-secret

# Optional: IP Whitelist (comma-separated)
IP_WHITELIST=127.0.0.1,::1

# Optional: CORS Origins (comma-separated, overrides defaults)
CORS_ORIGINS=http://localhost:3000,https://app.uniroute.co

# Optional: Tunnel Allowed Origins (comma-separated, overrides defaults)
TUNNEL_ORIGINS=http://localhost,https://tunnel.uniroute.co,.uniroute.co

# Base URL Configuration
# Base URL for the API server (used in documentation and examples)
# For local development: http://localhost:8084
# For production: https://app.uniroute.co (or your domain)
BASE_URL=http://localhost:8084

# Tunnel Server Configuration
# Base domain for tunnel subdomains (e.g., "uniroute.co" or "yourdomain.com")
# If not set, defaults to "uniroute.co"
TUNNEL_BASE_DOMAIN=uniroute.co

# Localhost domain for local development tunnels
# Default: "localhost" (results in subdomain.localhost:port)
TUNNEL_LOCALHOST_DOMAIN=localhost

# Base port for TCP/UDP tunnel allocation
# Default: 20000 (tunnels will use ports starting from this number)
TUNNEL_TCP_PORT_BASE=20000

# Website URL for links in error pages and UI
# Default: https://uniroute.co
WEBSITE_URL=https://uniroute.co
```

#### SMTP Configuration

UniRoute requires SMTP configuration for email verification and password reset functionality. For development, we recommend using [Mailtrap](https://mailtrap.io) (free tier):

1. **Sign up for Mailtrap** (free tier available)
2. **Get your SMTP credentials** from the Mailtrap dashboard
3. **Add to `.env` file**:
   ```bash
   SMTP_HOST=sandbox.smtp.mailtrap.io
   SMTP_PORT=2525
   SMTP_USERNAME=your-mailtrap-username
   SMTP_PASSWORD=your-mailtrap-password
   SMTP_FROM=noreply@uniroute.co
   ```

**For Production:**
- Use a production SMTP service (SendGrid, AWS SES, Mailgun, etc.)
- Update `SMTP_HOST`, `SMTP_PORT`, `SMTP_USERNAME`, and `SMTP_PASSWORD` accordingly
- Common ports: `587` (STARTTLS) or `465` (TLS)

#### OAuth Configuration

UniRoute supports OAuth authentication with Google, GitHub, and X (Twitter). To enable:

1. **Google OAuth:**
   - Go to [Google Cloud Console](https://console.cloud.google.com/)
   - Create OAuth 2.0 credentials
   - Add authorized redirect URI: `{BASE_URL}/auth/google/callback` (e.g., `http://localhost:8084/auth/google/callback` for local dev, or `https://app.uniroute.co/auth/google/callback` for production)
   - Add to `.env`:
     ```bash
     GOOGLE_OAUTH_CLIENT_ID=your-client-id
     GOOGLE_OAUTH_CLIENT_SECRET=your-client-secret
     ```

2. **GitHub OAuth:**
   - Go to [GitHub Developer Settings](https://github.com/settings/developers)
   - Create a new OAuth App
   - Set Authorization callback URL: `{BACKEND_URL}/auth/github/callback` (e.g., `http://localhost:8084/auth/github/callback` for local dev)
   - Add to `.env`:
     ```bash
     GITHUB_OAUTH_CLIENT_ID=your-client-id
     GITHUB_OAUTH_CLIENT_SECRET=your-client-secret
     ```

3. **X (Twitter) OAuth:**
   - Go to [Twitter Developer Portal](https://developer.twitter.com/)
   - Create an OAuth 2.0 app
   - Add callback URL: `{BACKEND_URL}/auth/x/callback` (e.g., `http://localhost:8084/auth/x/callback` for local dev)
   - Add to `.env`:
     ```bash
     X_OAUTH_CLIENT_ID=your-client-id
     X_OAUTH_CLIENT_SECRET=your-client-secret
     ```

**Note:** 
- OAuth providers redirect to your backend server, which then redirects to the frontend with the authentication token
- **OAuth users do NOT need email verification** - OAuth providers (Google, GitHub, X) already verify user emails, so users are automatically marked as verified upon OAuth login/registration

See [CLI Reference](https://uniroute.co/docs/cli) for detailed CLI installation and usage instructions, including the interactive UI.

### Verify Installation

```bash
# Check health endpoint (replace with your BASE_URL if different)
curl ${BASE_URL:-http://localhost:8084}/health
```

### First Request (Local LLM)

```bash
# With Ollama running locally
# Replace ${BASE_URL} with your actual base URL (e.g., http://localhost:8084)
curl -X POST ${BASE_URL:-http://localhost:8084}/v1/chat \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "llama2",
    "messages": [
      {"role": "user", "content": "Hello!"}
    ]
  }'
```

### Expose to Internet

**Option 1: Using cloudflared (100% free, no signup) - Recommended**
```bash
# Replace with your BASE_URL port (default: 8084)
cloudflared tunnel --url ${BASE_URL:-http://localhost:8084}
# Returns: https://random-subdomain.trycloudflare.com
```

**Option 2: Using cloudflared (100% free, no signup)**
```bash
# Replace 8084 with your actual port from BASE_URL
cloudflared tunnel --url ${BASE_URL:-http://localhost:8084}
# Returns: https://random-subdomain.trycloudflare.com
```

**Option 3: Built-in UniRoute tunnel (requires CLI installation)** ⭐ Recommended
```bash
# Download CLI (see CLI Reference in docs)
# Or build: make build

# Expose your local app (any port, any app)
# Replace 8084 with your actual port from BASE_URL
uniroute tunnel --port 8084
# Returns: http://{subdomain}.${TUNNEL_BASE_DOMAIN:-uniroute.co} -> ${BASE_URL}

# Works with any local application, not just UniRoute!

# Multiple protocols supported
uniroute tunnel --protocol http --port 8080   # HTTP tunnel
uniroute tunnel --protocol tcp --port 3306    # TCP tunnel (MySQL)
uniroute tunnel --protocol tls --port 5432    # TLS tunnel (PostgreSQL)
uniroute tunnel --protocol udp --port 53      # UDP tunnel (DNS)

# Shortcut commands
uniroute http 8080    # HTTP tunnel (shortcut)
uniroute tcp 3306     # TCP tunnel (shortcut)
uniroute tls 5432     # TLS tunnel (shortcut)
uniroute udp 53       # UDP tunnel (shortcut)

# Custom subdomain support (shortcut syntax)
uniroute http 8080 myapp              # Request specific subdomain (myapp.uniroute.co) - shortcut
uniroute http 8080 myapp --new        # Create new tunnel with specific subdomain - shortcut
uniroute tcp 3306 mydb                # TCP tunnel with specific subdomain - shortcut
uniroute tcp 3306 mydb --new          # TCP tunnel with subdomain and force new - shortcut

# Custom subdomain support (flag syntax - also works)
uniroute tunnel --host myapp          # Request specific subdomain (myapp.uniroute.co)
uniroute tunnel --host myapp --new    # Create new tunnel with specific subdomain
uniroute http 8080 --host myapp       # HTTP tunnel with specific subdomain
uniroute http 8080 --host myapp --new # HTTP tunnel with subdomain and force new

# Custom domain support
uniroute domain example.com                    # Add domain to account (no tunnel assignment)
uniroute domain example.com abc123             # Add domain AND assign to tunnel by subdomain (shortcut)
uniroute domain example.com --subdomain abc123  # Add domain AND assign to tunnel (flag syntax)
uniroute domain example.com --tunnel-id <id>    # Add domain AND assign to specific tunnel

# Domain management commands
uniroute domain list                           # List all your custom domains
uniroute domain show example.com               # Show domain details and status
uniroute domain verify example.com             # Verify DNS configuration
uniroute domain resume                         # Resume last used domain assignment
uniroute domain resume abc123                  # Resume domain assignment by subdomain
uniroute domain resume example.com              # Resume domain assignment by domain
uniroute domain remove example.com             # Remove domain from account

# Start multiple tunnels at once
uniroute tunnel --all  # Starts all configured tunnels from ~/.uniroute/tunnels.json

# Resume previous tunnel
uniroute tunnel --resume  # Automatically resumes last tunnel
uniroute resume abc123    # Resume tunnel by subdomain (shortcut)
```

**Dev server + tunnel (one command or your normal command):**

Three ways to run your dev server and get a public URL. Works with Laravel, Vue, React, Django, Rails, and more.

```bash
# Option 1: We start your dev server + tunnel (port auto-detected)
uniroute dev

# Option 2: You run your server; we only add the tunnel (two terminals)
# Terminal 1:  php artisan serve   (or npm run dev, rails s, etc.)
# Terminal 2:  uniroute dev --attach

# Option 3: Your exact command; we run it and add the tunnel (port from your command or project)
uniroute run -- php artisan serve
uniroute run -- php artisan serve --port=8080   # tunnel to 8080
uniroute run -- npm run dev
```

Port is taken from: (1) your command (e.g. `--port=8080`), (2) our `--port` flag, or (3) auto-detected from the project. Supported: **Node** (Vite/Next/React), **PHP** (Laravel), **Python** (Django, Flask, FastAPI), **Go**, **Ruby** (Rails). **Custom domains** work with dev/run: assign a domain to the tunnel and traffic to your domain goes to the same HTTP tunnel. See **Tunnels → Dev & Run** in the [Docs UI](/docs/tunnels/dev-run) or run `uniroute dev --help` and `uniroute run --help`.

**Tunnel Features:**
- ✅ HTTP, TCP, TLS, and UDP protocol support
- ✅ Persistent tunnels (survive CLI restarts)
- ✅ Multiple tunnels support
- ✅ Custom subdomains
- ✅ Custom domains (bring your own domain)
- ✅ Domain management (list, show, verify, remove)
- ✅ Domain resume functionality
- ✅ Automatic reconnection
- ✅ Tunnel state management

### Custom Domain Management

UniRoute supports using your own custom domains instead of random subdomains. You can manage domains through the CLI or dashboard.

**Adding and Assigning Domains:**

```bash
# Add domain to your account (not assigned to any tunnel yet)
uniroute domain example.com

# Add domain AND assign to tunnel in one command
uniroute domain example.com abc123              # By subdomain (shortcut)
uniroute domain example.com --subdomain abc123  # By subdomain (flag)
uniroute domain example.com --tunnel-id <id>   # By tunnel ID
```

**Domain Management Commands:**

```bash
# List all your domains
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

**DNS Configuration:**

After adding a domain, you need to configure DNS:

1. **Add CNAME Record** in your DNS provider:
   ```
   Type: CNAME
   Name: example.com (or @ for root domain)
   Target: tunnel.uniroute.co
   ```

2. **Verify DNS** configuration:
   ```bash
   uniroute domain verify example.com
   ```
   Or use the dashboard at `https://app.uniroute.co/dashboard/domains`

3. **Once verified**, your domain is ready to use!

**Domain Resume Feature:**

When you assign a domain to a tunnel, the assignment is automatically saved. You can resume it later:

```bash
# First time: assign domain to tunnel
uniroute domain billspot.co abc123

# Later: resume the same assignment
uniroute domain resume abc123
# or
uniroute domain resume billspot.co
```

The resume feature:
- ✅ Saves domain-to-tunnel assignments automatically
- ✅ Allows resuming by domain name or subdomain
- ✅ Automatically looks up current tunnel ID
- ✅ Works across CLI sessions (persistent storage)

**Domain Management in Dashboard:**

You can also manage domains through the web dashboard:
- View all domains: `/dashboard/domains`
- Add new domains
- Verify DNS configuration
- Delete domains

Both CLI and dashboard use the same backend system, so domains created via CLI appear in the dashboard and vice versa.

---

## ✨ Features

### Core Features

- **Unified API Interface** - Single endpoint for all LLM providers
- **Intelligent Routing** - Model selection based on cost, latency, availability
- **Load Balancing** - Distribute traffic across multiple instances
- **Automatic Failover** - Seamless switching when providers fail
- **Security & Access Control** - API keys, JWT, rate limiting, IP whitelisting
- **User Authentication** - Registration, login, email verification, password reset
- **OAuth Authentication** - Login/register with Google, GitHub, and X (Twitter)
- **Email Service** - SMTP integration for verification and password reset emails
- **Monitoring & Analytics** - Usage tracking, cost tracking, performance metrics
- **Error Logging** - Frontend error tracking and admin error log management
- **Multi-Provider Support** - OpenAI, Anthropic, Google, Local LLMs
- **Tunneling** - Built-in tunneling for exposing local services
- **CLI Tool** - Command-line interface for tunnel management and authentication
- **Developer Experience** - CLI tool, SDKs, OpenAPI docs

### Supported Providers

- 🏠 **Local LLMs (Ollama, vLLM)** ⭐ Priority - Free, private, self-hosted
- 🤖 OpenAI (GPT-4, GPT-3.5, etc.)
- 🧠 Anthropic (Claude)
- 🔷 Google (Gemini)
- 📊 Cohere
- ➕ Custom providers

---

## 🏗️ Architecture

### System Architecture

UniRoute follows a **layered architecture** with clear separation of concerns:

```
┌─────────────────────────────────────────────────────────────┐
│                    Client Layer                             │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐  │
│  │ Web App  │  │   CLI    │  │   SDK    │  │   API    │  │
│  │ (Vue.js) │  │  Tool    │  │  Clients │  │  Users   │  │
│  └────┬─────┘  └────┬─────┘  └────┬─────┘  └────┬─────┘  │
└───────┼─────────────┼─────────────┼─────────────┼────────┘
        │             │             │             │
        └─────────────┬──┴──────────────┴─────────────┘
                   │
                   ▼
┌─────────────────────────────────────────────────────────────┐
│              Presentation Layer (API Gateway)               │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  HTTP API (Gin Framework)                            │  │
│  │  ├── REST Endpoints (/v1/chat, /auth/*, etc.)       │  │
│  │  ├── WebSocket (Tunnel Server)                       │  │
│  │  └── Middleware (CORS, Auth, Rate Limit, Security)  │  │
│  └──────────────────────────────────────────────────────┘  │
└───────────────────────────┬─────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│                  Business Logic Layer                       │
│  ┌──────────────────┐  ┌──────────────────┐               │
│  │  Gateway Router  │  │  OAuth Service   │               │
│  │  ├── Routing     │  │  ├── Google      │               │
│  │  ├── Strategy    │  │  └── X/Twitter   │               │
│  │  └── Failover    │  └──────────────────┘               │
│  └──────────────────┘  ┌──────────────────┐               │
│  ┌──────────────────┐  │  Security Layer  │               │
│  │  Cost Calculator │  │  ├── JWT         │               │
│  │  Latency Tracker │  │  ├── API Keys    │               │
│  └──────────────────┘  │  └── Rate Limit  │               │
└─────────────────────────┼───────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────┐
│                    Data Access Layer                        │
│  ┌──────────────────┐  ┌──────────────────┐               │
│  │  Repositories    │  │  Provider Layer  │               │
│  │  ├── User        │  │  ├── OpenAI      │               │
│  │  ├── API Key     │  │  ├── Anthropic   │               │
│  │  ├── Tunnel      │  │  ├── Google      │               │
│  │  └── Request     │  │  └── Local LLM   │               │
│  └──────────────────┘  └──────────────────┘               │
└─────────────────────────┬───────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────┐
│                    Infrastructure Layer                     │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐    │
│  │  PostgreSQL  │  │    Redis     │  │   External   │    │
│  │  (Database)  │  │  (Cache/RL)  │  │   Services   │    │
│  └──────────────┘  └──────────────┘  └──────────────┘    │
└─────────────────────────────────────────────────────────────┘
```

### Architecture Principles

1. **Layered Architecture**: Clear separation between presentation, business logic, and data access
2. **Interface-Based Design**: Providers, repositories, and services use interfaces for flexibility
3. **Dependency Injection**: Services are injected, not hardcoded
4. **Repository Pattern**: Data access abstracted through repositories
5. **Strategy Pattern**: Routing strategies (cost, latency, balanced, custom)
6. **Middleware Pattern**: Cross-cutting concerns (auth, rate limiting, CORS)
7. **Configuration via Environment**: All settings via environment variables, no hardcoded values

### Project Structure

```
uniroute/
├── cmd/
│   ├── gateway/          # Gateway server entry point
│   ├── tunnel-server/    # Tunnel server entry point
│   └── cli/              # CLI tool entry point
├── internal/
│   ├── api/              # Presentation layer
│   │   ├── handlers/     # HTTP request handlers
│   │   ├── middleware/   # HTTP middleware (auth, CORS, rate limit)
│   │   └── router.go     # Route definitions
│   ├── gateway/          # Business logic - routing
│   │   ├── router.go     # Main routing logic
│   │   ├── strategy.go   # Routing strategies
│   │   ├── cost_calculator.go
│   │   └── latency_tracker.go
│   ├── providers/        # LLM provider implementations
│   │   ├── interface.go  # Provider interface
│   │   ├── openai.go
│   │   ├── anthropic.go
│   │   ├── google.go
│   │   └── local.go
│   ├── oauth/            # OAuth authentication
│   │   └── service.go    # Google & X OAuth service
│   ├── security/         # Security layer
│   │   ├── jwt.go        # JWT authentication
│   │   ├── apikey.go     # API key management
│   │   ├── apikey_v2.go  # Database-backed API keys
│   │   └── ratelimit.go  # Rate limiting
│   ├── storage/          # Data access layer
│   │   ├── postgres.go   # PostgreSQL client
│   │   ├── redis.go      # Redis client
│   │   ├── models.go     # Data models
│   │   └── *_repository.go  # Repository implementations
│   ├── tunnel/           # Tunnel functionality
│   │   ├── server.go     # Tunnel server
│   │   ├── client.go     # Tunnel client
│   │   └── repository.go # Tunnel data access
│   ├── email/            # Email service
│   └── config/           # Configuration management
├── frontend/             # Vue.js frontend application
│   ├── src/
│   │   ├── views/        # Page components
│   │   ├── components/   # Reusable components
│   │   ├── services/     # API clients
│   │   └── stores/       # State management
├── migrations/           # Database migrations
├── tests/                # Test suites
└── pkg/                  # Shared packages
    ├── logger/
    ├── errors/
    └── version/
```

### Data Flow

1. **Request Flow**:
   ```
   Client → API Gateway → Middleware (Auth/Rate Limit) → Handler → 
   Gateway Router → Provider → LLM API → Response → Client
   ```

2. **OAuth Flow** (No email verification required - OAuth providers verify):
   ```
   User clicks "Login with Google/GitHub/X" → Frontend calls /auth/google, /auth/github, or /auth/x → 
   Backend returns auth URL → User authorizes → OAuth provider redirects to 
   /auth/google/callback, /auth/github/callback, or /auth/x/callback → Backend creates/logs in user 
   (email auto-verified) → Redirects to frontend with JWT token → User authenticated
   ```

3. **Tunnel Flow**:
   ```
   CLI connects → Tunnel Server (WebSocket) → Tunnel created → 
   Public URL assigned → Traffic forwarded → Local service responds → 
   Response sent back through tunnel
```

---

## 💻 Technology Stack

- **Backend**: Go 1.21+
- **Frontend**: Vue.js 3 + TypeScript + Tailwind CSS
- **API Framework**: Gin
- **Database**: PostgreSQL + Redis
- **Authentication**: JWT + API Keys + OAuth2 (Google, GitHub, X/Twitter)
- **Email Service**: SMTP (Mailtrap, SendGrid, AWS SES, etc.)
- **Tunneling**: Built-in WebSocket-based tunnel server
- **Monitoring**: Prometheus + Grafana
- **Logging**: Structured logging (zerolog)

---

## 📚 Documentation

- **[Interactive Documentation](https://uniroute.co/docs)** - Full documentation with guides, API reference, and examples
- **[CLI Reference](https://uniroute.co/docs/cli)** - 📦 CLI installation and usage guide
- **[Tunnel Documentation](https://uniroute.co/docs/tunnels)** - 🔌 Tunnel configuration, protocols, and custom domains
- **[Custom Domains Guide](https://uniroute.co/docs/tunnels/custom-domains)** - 🌐 Custom domain setup and management
- **API Documentation**: Interactive Swagger UI available at `${BASE_URL}/swagger` when the server is running (default: `http://localhost:8084/swagger` for local development)
- **Postman Collection**: Import `UniRoute.postman_collection.json` for ready-to-use API requests

---

## 🔐 Security

**⚠️ CRITICAL: Security First**

UniRoute implements enterprise-grade security:

- ✅ Input validation and sanitization
- ✅ API keys with bcrypt hashing
- ✅ JWT authentication with strong secrets
- ✅ Rate limiting (per-key, per-IP)
- ✅ HTTPS/TLS enforcement
- ✅ Security headers (CSP, HSTS, etc.)
- ✅ Parameterized queries (SQL injection prevention)
- ✅ Encrypted secrets at rest

See the [Security overview](https://uniroute.co/docs/security) in the documentation for the complete checklist.

---

## 🐳 Deployment

### Docker Compose

```bash
docker-compose up -d
```

### Coolify

1. Create new application in Coolify
2. Connect GitHub repository
3. Set environment variables
4. Deploy automatically

For Coolify setup (frontend, backend, tunnel from one repo) and other options, see the [Deployment Guide](https://uniroute.co/docs/deployment) in the docs UI.

---

## 🧪 Development

### Project Structure

```
uniroute/
├── cmd/
│   └── gateway/          # Application entry point
├── internal/
│   ├── api/              # HTTP handlers & middleware
│   ├── gateway/          # Routing logic
│   ├── providers/        # LLM provider implementations
│   ├── security/         # Auth & rate limiting
│   ├── storage/          # Database clients
│   └── monitoring/       # Metrics & analytics
├── migrations/           # Database migrations
└── pkg/                  # Shared packages (reusable utilities)
```

### Makefile Commands

```bash
make dev          # Start development server
make build        # Build binary
make test         # Run tests
make migrate      # Run database migrations
make lint         # Run linters
make security     # Run security scans
make fmt          # Format code
make vet          # Run go vet
```

---

## 🧹 Code Quality & Best Practices

### Clean Code Principles

We follow **clean code principles** to ensure maintainability, readability, and reusability:

#### 1. **Single Responsibility Principle (SRP)**
Each function, struct, and package should have one clear purpose.

```go
// ✅ Good: Single responsibility
func ValidateAPIKey(key string) error {
    // Only validates API key format
}

// ❌ Bad: Multiple responsibilities
func ProcessRequestAndValidateAndLog(key string, req Request) error {
    // Too many responsibilities
}
```

#### 2. **DRY (Don't Repeat Yourself)**
Extract common patterns into reusable functions and packages.

```go
// ✅ Good: Reusable error handling
func HandleProviderError(err error, provider string) error {
    return fmt.Errorf("provider %s: %w", provider, err)
}

// ❌ Bad: Repeated error handling
func CallOpenAI() error {
    if err != nil {
        return fmt.Errorf("provider openai: %v", err)
    }
}
func CallAnthropic() error {
    if err != nil {
        return fmt.Errorf("provider anthropic: %v", err)
    }
}
```

#### 3. **Interface-Based Design**
Use interfaces for abstraction and testability.

```go
// ✅ Good: Provider interface for reusability
type Provider interface {
    Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error)
    HealthCheck(ctx context.Context) error
}

// All providers implement the same interface
type OpenAIProvider struct {}
type AnthropicProvider struct {}
type LocalLLMProvider struct {}
```

#### 4. **Composition Over Inheritance**
Use composition and embedding for code reuse.

```go
// ✅ Good: Composition with base functionality
type BaseProvider struct {
    client *http.Client
    logger *zerolog.Logger
}

type OpenAIProvider struct {
    BaseProvider
    apiKey string
}

func (p *OpenAIProvider) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
    // Reuse BaseProvider's client and logger
    return p.makeRequest(ctx, req)
}
```

#### 5. **Error Handling**
Consistent, informative error handling.

```go
// ✅ Good: Wrapped errors with context
var (
    ErrProviderUnavailable = errors.New("provider unavailable")
    ErrRateLimitExceeded   = errors.New("rate limit exceeded")
)

func CallProvider(ctx context.Context, req Request) error {
    if err := provider.Call(ctx, req); err != nil {
        return fmt.Errorf("failed to call provider: %w", err)
    }
    return nil
}
```

### Reusable Code Patterns

#### 1. **Provider Interface Pattern**
All LLM providers implement a common interface for easy swapping:

```go
// pkg/providers/interface.go
type Provider interface {
    Name() string
    Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error)
    Stream(ctx context.Context, req ChatRequest) (<-chan StreamChunk, error)
    HealthCheck(ctx context.Context) error
    GetModels() []string
}

// Easy to add new providers
type CustomProvider struct {
    // implements Provider interface
}
```

#### 2. **Middleware Pattern**
Reusable middleware for cross-cutting concerns:

```go
// internal/api/middleware/auth.go
func AuthMiddleware(apiKeyService *APIKeyService) gin.HandlerFunc {
    return func(c *gin.Context) {
        key := extractAPIKey(c)
        if err := apiKeyService.Validate(key); err != nil {
            c.JSON(401, gin.H{"error": "unauthorized"})
            c.Abort()
            return
        }
        c.Next()
    }
}

// Reusable across all routes
router.Use(AuthMiddleware(apiKeyService))
router.Use(RateLimitMiddleware(rateLimiter))
router.Use(LoggingMiddleware(logger))
```

#### 3. **Repository Pattern**
Abstract data access for testability and reusability:

```go
// internal/storage/repository.go
type APIKeyRepository interface {
    FindByKeyHash(ctx context.Context, hash string) (*APIKey, error)
    Create(ctx context.Context, key *APIKey) error
    Update(ctx context.Context, key *APIKey) error
    Delete(ctx context.Context, id string) error
}

// PostgreSQL implementation
type PostgresAPIKeyRepository struct {
    db *sql.DB
}

// Redis cache implementation
type CachedAPIKeyRepository struct {
    repo APIKeyRepository
    cache *redis.Client
}
```

#### 4. **Strategy Pattern**
Pluggable algorithms for routing, load balancing, etc.:

```go
// internal/gateway/strategy.go
type RoutingStrategy interface {
    SelectProvider(ctx context.Context, req Request, providers []Provider) (Provider, error)
}

// Different strategies
type CostBasedStrategy struct {}
type LatencyBasedStrategy struct {}
type RoundRobinStrategy struct {}
type FailoverStrategy struct {}

// Easy to swap strategies
router := NewRouter(CostBasedStrategy{})
```

#### 5. **Factory Pattern**
Centralized creation of providers and services:

```go
// internal/providers/factory.go
type ProviderFactory struct {
    config *Config
    logger *zerolog.Logger
}

func (f *ProviderFactory) CreateProvider(name string) (Provider, error) {
    switch name {
    case "openai":
        return NewOpenAIProvider(f.config.OpenAI, f.logger)
    case "anthropic":
        return NewAnthropicProvider(f.config.Anthropic, f.logger)
    case "local":
        return NewLocalLLMProvider(f.config.Local, f.logger)
    default:
        return nil, ErrUnknownProvider
    }
}
```

### Code Organization Guidelines

#### Package Structure
- **`pkg/`**: Reusable packages that can be imported by other projects
- **`internal/`**: Private application code, not importable by external projects
- **`cmd/`**: Application entry points

#### Naming Conventions
```go
// ✅ Good naming
type APIKeyService struct {}        // Clear, descriptive
func ValidateRequest(req Request)   // Verb + noun
const MaxRetries = 3                // Constants in PascalCase

// ❌ Bad naming
type Svc struct {}                  // Abbreviation unclear
func Do(req Request)                // Too generic
const max = 3                       // Should be exported if used elsewhere
```

#### Function Guidelines
- **Keep functions small**: Max 50 lines, ideally 20-30
- **One level of abstraction**: Don't mix high-level and low-level logic
- **Descriptive names**: Function name should describe what it does
- **Avoid side effects**: Functions should do one thing and do it well

```go
// ✅ Good: Small, focused function
func CalculateCost(tokens int, model string) (float64, error) {
    rate, err := GetModelRate(model)
    if err != nil {
        return 0, err
    }
    return float64(tokens) * rate, nil
}

// ❌ Bad: Too many responsibilities
func ProcessEverything(req Request) (*Response, error) {
    // 200 lines of mixed logic
}
```

### Testing Best Practices

#### Unit Tests
- Test one thing at a time
- Use table-driven tests for multiple scenarios
- Mock external dependencies

```go
// ✅ Good: Table-driven test
func TestCalculateCost(t *testing.T) {
    tests := []struct {
        name     string
        tokens   int
        model    string
        want     float64
        wantErr  bool
    }{
        {"gpt-4", 1000, "gpt-4", 0.03, false},
        {"invalid model", 1000, "invalid", 0, true},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := CalculateCost(tt.tokens, tt.model)
            // assertions...
        })
    }
}
```

### Code Review Checklist

Before submitting code, ensure:

- [ ] **Single Responsibility**: Each function/struct has one clear purpose
- [ ] **DRY**: No code duplication; common logic extracted
- [ ] **Interfaces**: Used for abstraction and testability
- [ ] **Error Handling**: All errors handled appropriately
- [ ] **Tests**: Unit tests for new functionality
- [ ] **Documentation**: Public functions/types documented
- [ ] **Naming**: Clear, descriptive names
- [ ] **Formatting**: Code formatted with `gofmt`
- [ ] **Linting**: No linter warnings
- [ ] **Security**: No security vulnerabilities introduced

### Reusability Checklist

When writing code, ask:

1. **Can this be reused?** - Extract to `pkg/` if reusable across projects
2. **Is this abstracted?** - Use interfaces for flexibility
3. **Is this configurable?** - Avoid hardcoded values
4. **Is this testable?** - Dependencies injected, not hardcoded
5. **Is this documented?** - Clear documentation for future developers

### Example: Reusable Provider Pattern

```go
// internal/providers/base.go - Reusable base provider
type BaseProvider struct {
    name     string
    client   *http.Client
    logger   *zerolog.Logger
    metrics  *prometheus.CounterVec
}

func (b *BaseProvider) makeRequest(ctx context.Context, url string, body interface{}) (*http.Response, error) {
    // Reusable HTTP request logic
    // Logging, metrics, retries, etc.
}

// internal/providers/openai.go - Specific implementation
type OpenAIProvider struct {
    BaseProvider
    apiKey string
    baseURL string
}

func (p *OpenAIProvider) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
    // Uses BaseProvider.makeRequest() for HTTP calls
    return p.makeRequest(ctx, p.baseURL+"/chat", req)
}
```

---

## 📝 Code Standards

### Go Style Guide
- Follow [Effective Go](https://go.dev/doc/effective_go) guidelines
- Use `gofmt` for formatting (enforced in CI)
- Follow [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)

### Required Tools
```bash
# Install development tools
go install golang.org/x/tools/cmd/goimports@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install github.com/securego/gosec/v2/cmd/gosec@latest

# Run before committing
make fmt      # Format code
make lint     # Run linters
make vet      # Run go vet
make security # Security scan
```

---

## 📊 API Examples

### Chat Completion (Local LLM)

```bash
# Using local Ollama instance
# Replace ${BASE_URL} with your actual base URL
curl -X POST ${BASE_URL:-http://localhost:8084}/v1/chat \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "llama2",
    "messages": [
      {"role": "user", "content": "Explain quantum computing"}
    ],
    "temperature": 0.7,
    "max_tokens": 1000
  }'
```

### Response

```json
{
  "id": "chat-123",
  "model": "llama2",
  "provider": "local",
  "choices": [{
    "message": {
      "role": "assistant",
      "content": "Quantum computing is a type of computation..."
    }
  }],
  "usage": {
    "prompt_tokens": 5,
    "completion_tokens": 10,
    "total_tokens": 15
  },
  "cost": 0.0,
  "latency_ms": 250
}
```

### Share Your Gateway Server

```bash
# 1. Start Ollama
ollama serve

# 2. Start UniRoute
make dev

# 3. Expose to internet (using cloudflared - 100% free, no signup)
# Replace with your BASE_URL
cloudflared tunnel --url ${BASE_URL:-http://localhost:8084}
# Share: https://random-subdomain.trycloudflare.com

# 4. Others can now use your gateway!
curl -X POST https://random-subdomain.trycloudflare.com/v1/chat \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"model": "llama2", "messages": [...]}'
```

---

## 🤝 Contributing

Contributions are welcome! We prioritize **clean, reusable code**. Please follow these guidelines:

### Contribution Workflow

1. **Fork the repository**
2. **Create a feature branch** (`git checkout -b feature/amazing-feature`)
3. **Write clean, reusable code**:
   - Follow Single Responsibility Principle
   - Extract common patterns into reusable functions
   - Use interfaces for abstraction
   - Keep functions small and focused
4. **Write tests** for new features (aim for >80% coverage)
5. **Update documentation** (code comments, README, etc.)
6. **Run quality checks**:
   ```bash
   make fmt      # Format code
   make lint     # Run linters
   make test     # Run tests
   make security # Security scan
   ```
7. **Commit using conventional commits**:
   ```
   feat: add new provider support
   fix: resolve rate limiting bug
   refactor: extract common error handling
   docs: update API documentation
   ```
8. **Push to the branch** (`git push origin feature/amazing-feature`)
9. **Open a Pull Request** with clear description

### Code Contribution Guidelines

#### Before You Start
- Check existing code patterns and follow them
- Look for similar functionality that can be reused
- Discuss major changes in an issue first

#### While Coding
- **Reuse existing code**: Check `pkg/` and `internal/` for reusable utilities
- **Create reusable components**: If you write something that could be reused, put it in `pkg/`
- **Use interfaces**: Abstract dependencies for testability
- **Write small functions**: Max 50 lines, ideally 20-30
- **Document public APIs**: All exported functions/types need godoc comments

#### Code Review Focus
- **Reusability**: Can this code be reused elsewhere?
- **Maintainability**: Is the code easy to understand and modify?
- **Testability**: Can this be easily tested?
- **Performance**: Are there any obvious performance issues?
- **Security**: Any security concerns?

### Example: Good Contribution

```go
// ✅ Good: Reusable, well-documented, testable
// Package providers contains reusable provider implementations.
package providers

// Provider defines the interface for all LLM providers.
// This interface allows easy swapping of providers and testing.
type Provider interface {
    // Chat sends a chat request to the provider.
    Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error)
    
    // HealthCheck verifies the provider is available.
    HealthCheck(ctx context.Context) error
}

// NewProviderFactory creates a new provider factory with the given config.
// This factory pattern allows centralized provider creation.
func NewProviderFactory(config *Config) *ProviderFactory {
    return &ProviderFactory{config: config}
}
```

### Example: Bad Contribution

```go
// ❌ Bad: Not reusable, no documentation, hard to test
func doStuff(req map[string]interface{}) map[string]interface{} {
    // 200 lines of mixed logic
    // Hardcoded values
    // No error handling
    // Can't be reused
}
```

---

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

## 🔗 Links

## 💝 Donate

UniRoute is an open-source project built with ❤️ by the community. If you find it useful, please consider supporting the project:

- ⭐ **Star the repository** on GitHub
- 🐛 **Report bugs** and suggest features
- 💻 **Contribute code** via pull requests
- ☕ **Buy us a coffee** - [Donate](https://buy.polar.sh/polar_cl_h5uF0bHhXXF6EO8Mx3tVP1Ry1G4wNWn4V8phg3rStVs)

Your support helps us continue developing and maintaining UniRoute!

---

## 🔗 Links

- **GitHub**: https://github.com/Kizsoft-Solution-Limited/uniroute
- **API Documentation**: `http://localhost:8084/swagger` (when server is running)
- **Issues**: https://github.com/Kizsoft-Solution-Limited/uniroute/issues

---

## ✨ Current Features

- ✅ **Unified API** - Single endpoint for all LLM providers
- ✅ **Multi-Provider Support** - OpenAI, Anthropic, Google, Local LLMs (Ollama, vLLM)
- ✅ **User Authentication** - Registration, login, email verification, password reset
- ✅ **OAuth Authentication** - Login/register with Google, GitHub, and X (Twitter)
- ✅ **Email Service** - SMTP integration for verification and password reset emails
- ✅ **Security & Rate Limiting** - API keys, JWT, progressive rate limiting, IP whitelisting
- ✅ **Intelligent Routing** - Cost-based, latency-based, and failover routing
- ✅ **Custom Routing Rules** - User-specific routing strategies and custom rules
- ✅ **BYOK (Bring Your Own Keys)** - User-provided provider API keys with encryption
- ✅ **Monitoring & Analytics** - Usage tracking, cost tracking, performance metrics
- ✅ **Tunneling** - Built-in tunnel server and CLI for exposing local services (HTTP, TCP, TLS)
- ✅ **CLI Tool** - Full-featured command-line interface for tunnel management and authentication
- ✅ **Error Logging** - Frontend error tracking with admin dashboard
- ✅ **Developer Experience** - CLI tool, SDKs, built-in tunneling, OpenAPI docs

---

## ⭐ Star History

If you find UniRoute useful, please consider giving it a star ⭐ on GitHub!

---

## 🙏 Acknowledgments

- Built with modern Go best practices and clean architecture
- Built with ❤️ by the open-source community

---

## 📋 Quick Reference: Clean Code Checklist

### Before Committing Code

```bash
# Run all checks
make fmt      # Format with gofmt
make lint     # Run linters
make test     # Run tests
make vet      # Run go vet
make security # Security scan
```

### Code Quality Questions

Ask yourself:
- ✅ **Single Purpose?** Does this function/struct do one thing?
- ✅ **Reusable?** Can this be extracted and reused?
- ✅ **Testable?** Can I easily write tests for this?
- ✅ **Documented?** Is the code self-explanatory or documented?
- ✅ **No Duplication?** Is this logic already implemented elsewhere?
- ✅ **Interface-Based?** Am I using interfaces for abstraction?
- ✅ **Error Handling?** Are all errors handled appropriately?
- ✅ **Small Functions?** Is each function < 50 lines?

### Reusability Patterns

| Pattern | Use Case | Location |
|---------|----------|----------|
| **Interface** | Provider abstraction | `internal/providers/interface.go` |
| **Middleware** | Cross-cutting concerns | `internal/api/middleware/` |
| **Repository** | Data access abstraction | `internal/storage/repository.go` |
| **Strategy** | Pluggable algorithms | `internal/gateway/strategy.go` |
| **Factory** | Object creation | `internal/providers/factory.go` |
| **Base Struct** | Shared functionality | `internal/providers/base.go` |

### Common Reusable Components

- **Error Handling**: `pkg/errors/` - Standardized error types
- **Logging**: `pkg/logger/` - Structured logging utilities
- **HTTP Client**: `internal/providers/base.go` - Reusable HTTP client
- **Validation**: `internal/api/validators/` - Request validation
- **Metrics**: `internal/monitoring/metrics.go` - Prometheus metrics

---

**Made with ❤️ by [Kizsoft Solution Limited](https://github.com/Kizsoft-Solution-Limited)**

