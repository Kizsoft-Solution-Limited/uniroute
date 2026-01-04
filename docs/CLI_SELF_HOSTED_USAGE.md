# Using CLI with Self-Hosted UniRoute

## 🎯 Overview

The UniRoute CLI works with **both hosted and self-hosted** instances. You can easily switch between them or use both.

---

## 🔄 How It Works

### Default Behavior

By default, the CLI connects to the **hosted UniRoute service** at `https://api.uniroute.dev`.

### Self-Hosted Usage

When you self-host UniRoute, you can use the CLI with your own instance in two ways:

1. **Per-command** - Use `--url` or `--server` flags for each command
2. **Persistent** - Login once with `--server` flag to save your server URL

---

## 📋 CLI Commands for Self-Hosted

### 1. Authentication (`auth login`)

**Login to your self-hosted instance:**

```bash
# Login to self-hosted instance
uniroute auth login --server http://localhost:8084

# Or with full URL
uniroute auth login --server https://your-uniroute.example.com

# Check current auth status
uniroute auth status
# Shows: Server: http://localhost:8084
```

**What happens:**
- CLI saves the server URL in `~/.uniroute/auth.json`
- All future commands use this server URL by default
- You can switch back to hosted by logging in again with `--server https://api.uniroute.dev`

---

### 2. Check Status (`status`)

**Check your self-hosted instance:**

```bash
# Using saved server (from auth login)
uniroute status

# Or specify server directly
uniroute status --url http://localhost:8084

# Check hosted server
uniroute status --url https://api.uniroute.dev
```

**Output:**
```
✅ Server is running at http://localhost:8084
   Status: ok
   Providers: [local, openai, anthropic]
```

---

### 3. API Key Management (`keys`)

**Create API keys on your self-hosted instance:**

```bash
# Using saved server (from auth login)
uniroute keys create --name "My API Key"

# Or specify server directly
uniroute keys create --url http://localhost:8084 --name "My API Key"

# With JWT token (if not logged in)
uniroute keys create --url http://localhost:8084 --jwt-token YOUR_JWT

# List keys
uniroute keys list --url http://localhost:8084
```

**Note:** For self-hosted, you may need to:
- Use JWT token if database-backed keys are enabled
- Or use in-memory keys (Phase 1 mode) which don't require auth

---

### 4. Tunnel (`tunnel`)

**Use tunnel with self-hosted tunnel server:**

```bash
# Connect to your self-hosted tunnel server
uniroute tunnel --port 8084 --server localhost:8080

# Or with full URL
uniroute tunnel --port 8084 --server https://tunnel.your-domain.com

# Using hosted tunnel server (default)
uniroute tunnel --port 8084
# Uses: tunnel.uniroute.dev (requires auth)
```

**Authentication:**
- **Localhost**: No auth required (for development)
- **Public servers**: Requires authentication (`uniroute auth login`)

---

### 5. Projects (`projects`)

**Manage projects on self-hosted:**

```bash
# List projects (using saved server)
uniroute projects list

# Or specify server
uniroute projects list --server http://localhost:8084

# Show project details
uniroute projects show PROJECT_ID --server http://localhost:8084
```

---

## 🔀 Switching Between Hosted and Self-Hosted

### Switch to Self-Hosted

```bash
# Login to your self-hosted instance
uniroute auth login --server http://localhost:8084

# All commands now use your self-hosted instance
uniroute status
uniroute keys create
uniroute projects list
```

### Switch Back to Hosted

```bash
# Login to hosted service
uniroute auth login --server https://api.uniroute.dev

# All commands now use hosted service
uniroute status
uniroute keys create
uniroute projects list
```

### Use Both (Per-Command)

```bash
# Use hosted for some commands
uniroute status --url https://api.uniroute.dev
uniroute keys create --url https://api.uniroute.dev

# Use self-hosted for others
uniroute status --url http://localhost:8084
uniroute keys create --url http://localhost:8084
```

---

## 📁 Configuration Storage

### Auth Config Location

The CLI stores authentication and server URL in:
```
~/.uniroute/auth.json
```

**File structure:**
```json
{
  "token": "jwt-token-here",
  "email": "user@example.com",
  "server_url": "http://localhost:8084",
  "expires_at": "2024-12-31T23:59:59Z"
}
```

### Tunnel State Location

Tunnel state is stored in:
```
~/.uniroute/tunnel-state.json
```

**File structure:**
```json
{
  "tunnel_id": "abc123",
  "subdomain": "xyz789",
  "public_url": "http://xyz789.localhost:8080",
  "local_url": "http://localhost:8084",
  "server_url": "localhost:8080",
  "created_at": "2024-01-01T00:00:00Z",
  "last_used": "2024-01-01T12:00:00Z"
}
```

---

## 🎯 Common Scenarios

### Scenario 1: Development with Local Instance

```bash
# Start your self-hosted UniRoute
./bin/uniroute-gateway

# In another terminal, use CLI
uniroute auth login --server http://localhost:8084
uniroute status
uniroute keys create --name "Dev Key"
```

### Scenario 2: Production Self-Hosted

```bash
# Login to production instance
uniroute auth login --server https://api.yourcompany.com

# Use CLI normally
uniroute status
uniroute keys create --name "Production Key"
uniroute projects list
```

### Scenario 3: Using Both Hosted and Self-Hosted

```bash
# Work with hosted (default)
uniroute auth login --server https://api.uniroute.dev
uniroute keys create --name "Hosted Key"

# Work with self-hosted (per-command)
uniroute keys create --url http://localhost:8084 --name "Local Key"
```

### Scenario 4: Team Using Self-Hosted

```bash
# Each team member logs into shared self-hosted instance
uniroute auth login --server https://api.yourcompany.com

# Everyone uses CLI normally
uniroute status
uniroute keys create
uniroute projects list
```

---

## 🔐 Authentication for Self-Hosted

### Option 1: JWT Token (Database-Backed)

If your self-hosted instance uses database-backed API keys (Phase 2+):

1. **Get JWT token** from your self-hosted instance
   - Via web dashboard
   - Via API: `POST /auth/login`

2. **Login with CLI:**
   ```bash
   uniroute auth login --server http://localhost:8084
   # Enter email and password
   ```

3. **Or use JWT directly:**
   ```bash
   uniroute keys create --url http://localhost:8084 --jwt-token YOUR_JWT
   ```

### Option 2: In-Memory Keys (Phase 1)

If your self-hosted instance uses in-memory keys (Phase 1):

- **No authentication needed** for API key creation
- Just use `--url` flag:
  ```bash
  uniroute status --url http://localhost:8084
  ```

---

## 🚀 Quick Start: Self-Hosted

### Step 1: Start Your Self-Hosted Instance

```bash
# Start UniRoute gateway
./bin/uniroute-gateway

# Or with Docker
docker run -p 8084:8084 uniroute-gateway
```

### Step 2: Configure CLI

```bash
# Login to your instance
uniroute auth login --server http://localhost:8084

# Or just use --url for each command
uniroute status --url http://localhost:8084
```

### Step 3: Use CLI

```bash
# Check status
uniroute status

# Create API key
uniroute keys create --name "My Key"

# Use the API key in your app
curl -X POST http://localhost:8084/v1/chat \
  -H "Authorization: Bearer ur_your-key" \
  -d '{"model": "gpt-4", "messages": [...]}'
```

---

## 📊 Command Reference

### All Commands Support `--url` or `--server`

| Command | Flag | Default | Example |
|---------|------|---------|---------|
| `auth login` | `--server` | `https://api.uniroute.dev` | `--server http://localhost:8084` |
| `status` | `--url` | Saved server | `--url http://localhost:8084` |
| `keys create` | `--url` | Saved server | `--url http://localhost:8084` |
| `keys list` | `--url` | Saved server | `--url http://localhost:8084` |
| `projects list` | `--server` | Saved server | `--server http://localhost:8084` |
| `tunnel` | `--server` | `tunnel.uniroute.dev` | `--server localhost:8080` |

---

## 🔍 Troubleshooting

### CLI Can't Connect to Self-Hosted

```bash
# Check if server is running
uniroute status --url http://localhost:8084

# Check if port is correct
curl http://localhost:8084/health

# Check firewall/network
ping your-server.com
```

### Wrong Server URL

```bash
# Check current server
uniroute auth status

# Update server URL
uniroute auth login --server http://correct-url:8084
```

### Authentication Issues

```bash
# Check if logged in
uniroute auth status

# Re-login
uniroute auth logout
uniroute auth login --server http://localhost:8084

# Use JWT token directly
uniroute keys create --url http://localhost:8084 --jwt-token YOUR_JWT
```

---

## 📝 Summary

**Self-Hosted Usage:**
1. ✅ Use `--url` or `--server` flags to point to your instance
2. ✅ Or login once with `--server` to save it
3. ✅ All commands work the same way
4. ✅ Can switch between hosted and self-hosted easily
5. ✅ Configuration stored in `~/.uniroute/auth.json`

**The CLI is designed to work seamlessly with both hosted and self-hosted instances!** 🚀

