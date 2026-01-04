# Tunnel Persistence & Resume Features

## ✅ New Features

### 1. Automatic Subdomain Persistence
- **Location**: `internal/tunnel/persistence.go`, `internal/tunnel/client.go`
- **Features**:
  - Automatically saves tunnel state to `~/.uniroute/tunnel-state.json`
  - Saves subdomain, tunnel ID, public URL, local URL, server URL
  - Automatically loads saved state on next tunnel start
  - Attempts to resume same tunnel automatically

### 2. Resume Same Tunnel After Restart
- **Location**: `internal/tunnel/client.go`
- **Features**:
  - Client automatically loads saved subdomain/tunnel ID
  - Attempts to resume same tunnel when reconnecting
  - Falls back to new tunnel if resume fails
  - Works across client restarts (not just reconnections)

### 3. CLI Commands for Tunnel Management
- **Location**: `cmd/cli/commands/tunnel.go`
- **New Flags**:
  - `--resume <subdomain>`: Resume a specific subdomain
  - `--list`: List saved tunnel state
  - `--clear`: Clear saved tunnel state

## How It Works

### Automatic Persistence Flow

```
User Starts Tunnel
  ↓
Client Connects → Gets Subdomain "abc123"
  ↓
Client Saves State to ~/.uniroute/tunnel-state.json
  ↓
User Closes Tunnel (Ctrl+C)
  ↓
User Starts Tunnel Again
  ↓
Client Loads Saved State
  ↓
Client Attempts to Resume "abc123"
  ↓
Server Resumes Existing Tunnel (if still active)
  OR
Server Creates New Tunnel (if old one expired)
```

### Saved State Format

```json
{
  "tunnel_id": "6ba51eefa0592d03eede6653c75511bc",
  "subdomain": "f43958c6b098",
  "public_url": "http://f43958c6b098.localhost:8080",
  "local_url": "http://localhost:8084",
  "server_url": "localhost:8080",
  "created_at": "2026-01-04T17:55:46Z",
  "last_used": "2026-01-04T17:55:46Z"
}
```

## Usage Examples

### Automatic Resume (Default Behavior)

```bash
# First time - creates new tunnel
$ ./bin/uniroute tunnel --built-in --port 8084
Session Status                online
Subdomain                     abc123def4
Forwarding                    http://abc123def4.localhost:8080 -> http://localhost:8084

# User closes (Ctrl+C)

# Second time - automatically resumes same tunnel
$ ./bin/uniroute tunnel --built-in --port 8084
Loaded saved tunnel state - will attempt to resume
Session Status                online
Subdomain                     abc123def4  # Same subdomain!
Forwarding                    http://abc123def4.localhost:8080 -> http://localhost:8084
```

### List Saved Tunnel State

```bash
$ ./bin/uniroute tunnel --list
📋 Saved Tunnel State:
   Subdomain: f43958c6b098
   Public URL: http://f43958c6b098.localhost:8080
   Local URL: http://localhost:8084
   Server URL: localhost:8080
   Last Used: 2026-01-04 17:55:46
```

### Clear Saved State

```bash
$ ./bin/uniroute tunnel --clear
✓ Cleared saved tunnel state

# Next tunnel start will create a new subdomain
```

### Resume Specific Subdomain

```bash
$ ./bin/uniroute tunnel --built-in --resume abc123def4
Attempting to resume subdomain: abc123def4
Session Status                online
Subdomain                     abc123def4
```

## Benefits

1. **Consistent URLs**: Your public URL stays the same across restarts
2. **Better UX**: No need to update URLs after restarting tunnel
3. **Flexibility**: Can clear saved state to get new subdomain
4. **Transparency**: Can view saved state with `--list`
5. **Automatic**: Works automatically, no extra steps needed

## File Location

Saved tunnel state is stored in:
- **macOS/Linux**: `~/.uniroute/tunnel-state.json`
- **Windows**: `%USERPROFILE%\.uniroute\tunnel-state.json`

## Smart Resume Logic

The client will:
1. ✅ Load saved state if it exists
2. ✅ Only use saved state if server URL matches (prevents wrong server resume)
3. ✅ Attempt to resume saved subdomain
4. ✅ Fall back to new tunnel if resume fails (server doesn't have it)
5. ✅ Save new tunnel state after successful connection

## Status

✅ **Fully Implemented**

- Automatic persistence ✅
- Resume after restart ✅
- CLI commands (--list, --clear) ✅
- Smart server URL matching ✅
- Graceful fallback ✅

Now users can close and restart their tunnel while keeping the same subdomain! 🎉

