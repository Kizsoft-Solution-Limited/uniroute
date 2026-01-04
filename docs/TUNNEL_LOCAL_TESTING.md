# Tunnel Local Testing Guide

## Quick Start

### 1. Build the Tunnel Server

```bash
# Build all binaries (including tunnel server)
make build

# Or build just the tunnel server
CGO_ENABLED=0 go build -o bin/uniroute-tunnel-server ./cmd/tunnel-server
```

### 2. Start the Tunnel Server

```bash
# Start on default port 8080
./bin/uniroute-tunnel-server

# Or specify a port
./bin/uniroute-tunnel-server --port 8080

# With debug logging
./bin/uniroute-tunnel-server --port 8080 --log-level debug
```

### 3. Verify Server is Running

```bash
# Check health endpoint
curl http://localhost:8080/health

# Expected response:
# {"status":"ok","tunnels":0}

# Check root endpoint
curl http://localhost:8080/

# Expected: HTML page showing tunnel server status
```

### 4. Start Your Local Gateway

In a new terminal:

```bash
# Start the UniRoute gateway
./bin/uniroute-gateway

# Or use the CLI
./bin/uniroute start
```

### 5. Connect Tunnel Client

In another terminal:

```bash
# Connect using built-in tunnel
./bin/uniroute tunnel --built-in --port 8084 --server localhost:8080

# You should see:
# 🌐 Starting UniRoute tunnel...
#    Local URL: http://localhost:8084
#    Tunnel Server: localhost:8080
# 
# Connecting to tunnel server...
# 
# Session Status                online
# Tunnel Server                 UniRoute (built-in)
# Version                       1.0.0
# Forwarding                    http://{subdomain}.localhost:8080 -> http://localhost:8084
```

### 6. Test the Tunnel

Once connected, you'll get a subdomain like `abc123.localhost:8080`. Test it:

```bash
# Replace {subdomain} with your actual subdomain
curl http://{subdomain}.localhost:8080/health

# Or test the gateway endpoint
curl http://{subdomain}.localhost:8080/v1/health
```

## Testing Scenarios

### Scenario 1: Basic Request Forwarding

1. Start tunnel server: `./bin/uniroute-tunnel-server --port 8080`
2. Start gateway: `./bin/uniroute-gateway` (on port 8084)
3. Connect tunnel: `./bin/uniroute tunnel --built-in --port 8084 --server localhost:8080`
4. Make request: `curl http://{subdomain}.localhost:8080/v1/health`

**Expected**: Request is forwarded through tunnel and returns gateway response

### Scenario 2: Request Tracking

1. Enable debug logging: `./bin/uniroute-tunnel-server --port 8080 --log-level debug`
2. Connect tunnel and make requests
3. Check logs for request IDs and tracking

**Expected**: Each request has unique ID, responses are matched correctly

### Scenario 3: Disconnection Handling

1. Connect tunnel
2. Make a request
3. Kill the tunnel client (Ctrl+C)
4. Restart tunnel client
5. Make another request

**Expected**: Queued requests are processed after reconnection

### Scenario 4: Multiple Tunnels

1. Start tunnel server
2. Connect multiple clients with different local URLs
3. Each gets unique subdomain
4. Test each tunnel independently

**Expected**: Each tunnel works independently with its own subdomain

## Troubleshooting

### Issue: "no such file or directory"

**Solution**: Build the binary first:
```bash
make build
# or
CGO_ENABLED=0 go build -o bin/uniroute-tunnel-server ./cmd/tunnel-server
```

### Issue: Port already in use

**Solution**: Use a different port:
```bash
./bin/uniroute-tunnel-server --port 8081
```

### Issue: Connection refused

**Solution**: 
1. Verify tunnel server is running: `curl http://localhost:8080/health`
2. Check firewall settings
3. Verify local gateway is running on the specified port

### Issue: Request timeout

**Solution**:
1. Check local gateway is accessible: `curl http://localhost:8084/health`
2. Verify tunnel client is connected
3. Check tunnel server logs for errors

## Monitoring

### Check Active Tunnels

```bash
# Health endpoint shows tunnel count
curl http://localhost:8080/health

# Root endpoint shows active tunnels
curl http://localhost:8080/
```

### View Logs

Tunnel server logs show:
- Tunnel connections/disconnections
- Request forwarding
- Errors and warnings
- Request tracking (with debug level)

## Next Steps

Once local testing is successful:

1. ✅ Verify request/response matching
2. ✅ Test request queuing
3. ✅ Test multiple tunnels
4. ✅ Test error handling
5. ✅ Test authentication (if enabled)

Ready for production deployment! 🚀

