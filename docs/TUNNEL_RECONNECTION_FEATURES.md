# Tunnel Reconnection & Resume Features

## ✅ Implemented Features

### 1. Automatic Reconnection
- **Location**: `internal/tunnel/client.go`
- **Features**:
  - Automatic reconnection on connection loss
  - Exponential backoff (5s → 60s max)
  - Continuous retry until successful
  - Connection status monitoring

### 2. Subdomain Persistence
- **Location**: `internal/tunnel/client.go`
- **Features**:
  - Saves subdomain and tunnel ID after first connection
  - Automatically resumes same tunnel on reconnect
  - No need to create new tunnel each time
  - Preserves your public URL

### 3. Request Queuing
- **Location**: `internal/tunnel/client.go`
- **Features**:
  - Queues requests during disconnection
  - Automatically processes queued requests after reconnection
  - Prevents request loss during network issues

### 4. Server-Side Resume Support
- **Location**: `internal/tunnel/server.go`
- **Features**:
  - Detects resume requests (subdomain/tunnelID in InitMessage)
  - Reuses existing tunnel instead of creating new one
  - Updates WebSocket connection for resumed tunnel
  - Maintains tunnel statistics

## How It Works

### Client Reconnection Flow

```
Connection Lost
  ↓
Detect Disconnection
  ↓
Save Current Subdomain/TunnelID
  ↓
Exponential Backoff Retry
  ↓
Reconnect with Saved Subdomain/TunnelID
  ↓
Server Resumes Existing Tunnel
  ↓
Process Queued Requests
  ↓
Continue Normal Operation
```

### Resume Request Format

```json
{
  "type": "init",
  "version": "1.0",
  "local_url": "http://localhost:8084",
  "subdomain": "abc123def4",  // Optional: for resume
  "tunnel_id": "tunnel-id-123"  // Optional: for resume
}
```

## Usage

### Automatic Behavior

The tunnel client **automatically**:
1. Saves subdomain/tunnel ID after first connection
2. Attempts to resume same tunnel on reconnect
3. Falls back to new tunnel if resume fails
4. Processes queued requests after reconnection

### User Experience

```bash
# First connection
$ ./bin/uniroute tunnel --built-in --port 8084 --server localhost:8080
🌐 Starting UniRoute tunnel...
Session Status                online
Subdomain                     abc123def4
Forwarding                    http://abc123def4.localhost:8080 -> http://localhost:8084

# Network interruption...
⚠️  Connection lost, attempting to reconnect...

# Automatic reconnection (same subdomain!)
Reconnected successfully - resumed same tunnel
Subdomain                     abc123def4  # Same as before!
Forwarding                    http://abc123def4.localhost:8080 -> http://localhost:8084
```

## Benefits

1. **Consistent URLs**: Your public URL stays the same across reconnections
2. **No Request Loss**: Queued requests are processed after reconnection
3. **Seamless Recovery**: Automatic reconnection without user intervention
4. **Better UX**: Users don't need to update URLs after network issues

## Technical Details

### Reconnection Logic

```go
// Exponential backoff
backoff := 5 * time.Second
maxBackoff := 60 * time.Second

for {
    time.Sleep(backoff)
    if err := tc.Connect(); err == nil {
        // Success - resumed same tunnel
        return
    }
    backoff *= 2
    if backoff > maxBackoff {
        backoff = maxBackoff
    }
}
```

### Resume Detection

```go
// Client sends resume request
if tc.subdomain != "" || tc.tunnelID != "" {
    initMsg.Subdomain = tc.subdomain
    initMsg.TunnelID = tc.tunnelID
}

// Server detects and resumes
if initMsg.Subdomain != "" {
    existingTunnel = ts.tunnels[initMsg.Subdomain]
    if existingTunnel != nil {
        // Resume existing tunnel
        existingTunnel.WSConn = ws
        // ... update connection
    }
}
```

## Testing

### Test Reconnection

1. Start tunnel client
2. Note the subdomain
3. Stop tunnel server (simulate network loss)
4. Restart tunnel server
5. Client should automatically reconnect with same subdomain

### Test Request Queuing

1. Start tunnel client
2. Make HTTP request to tunnel
3. Stop tunnel server mid-request
4. Restart tunnel server
5. Queued request should be processed

## Status

✅ **Fully Implemented and Working**

- Automatic reconnection ✅
- Subdomain persistence ✅
- Request queuing ✅
- Server-side resume ✅
- User-friendly messages ✅

The tunnel now provides a robust, production-ready experience with automatic recovery from network issues!

