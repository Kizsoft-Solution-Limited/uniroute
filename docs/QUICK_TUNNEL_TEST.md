# Quick Tunnel Test Guide

## Step-by-Step Testing

### 1. Start Tunnel Server (Terminal 1)

```bash
./bin/uniroute-tunnel-server --port 8080
```

You should see:
```
{"level":"info","port":8080,"time":"...","message":"Tunnel server starting"}
```

### 2. Start Gateway (Terminal 2)

```bash
./bin/uniroute-gateway
```

Or:
```bash
./bin/uniroute start
```

### 3. Connect Tunnel Client (Terminal 3)

```bash
./bin/uniroute tunnel --built-in --port 8084 --server localhost:8080
```

You should see:
```
🌐 Starting UniRoute tunnel...
   Local URL: http://localhost:8084
   Tunnel Server: localhost:8080

Connecting to tunnel server...

Session Status                online
Tunnel Server                 UniRoute (built-in)
Version                       1.0.0
Forwarding                    http://abc123.localhost:8080 -> http://localhost:8084

Press Ctrl+C to stop
```

### 4. Test the Tunnel (Terminal 4)

Replace `abc123` with your actual subdomain:

```bash
# Test health endpoint
curl http://abc123.localhost:8080/health

# Test gateway endpoint
curl http://abc123.localhost:8080/v1/health
```

## Troubleshooting

### Connection Refused

**Problem**: `dial tcp [::1]:8080: connect: connection refused`

**Solution**: Make sure tunnel server is running:
```bash
# Check if running
ps aux | grep uniroute-tunnel-server

# Start if not running
./bin/uniroute-tunnel-server --port 8080
```

### Port Already in Use

**Problem**: Port 8080 is already in use

**Solution**: Use a different port:
```bash
# Start server on different port
./bin/uniroute-tunnel-server --port 8081

# Connect with different port
./bin/uniroute tunnel --built-in --port 8084 --server localhost:8081
```

### Gateway Not Running

**Problem**: Tunnel connects but requests fail

**Solution**: Make sure gateway is running on port 8084:
```bash
# Check gateway
curl http://localhost:8084/health

# Start gateway if needed
./bin/uniroute-gateway
```

## Quick Test Script

Save this as `test-tunnel.sh`:

```bash
#!/bin/bash

echo "Starting tunnel server..."
./bin/uniroute-tunnel-server --port 8080 > /tmp/tunnel-server.log 2>&1 &
TUNNEL_PID=$!

sleep 2

echo "Starting gateway..."
./bin/uniroute-gateway > /tmp/gateway.log 2>&1 &
GATEWAY_PID=$!

sleep 2

echo "Connecting tunnel..."
./bin/uniroute tunnel --built-in --port 8084 --server localhost:8080 &
TUNNEL_CLIENT_PID=$!

sleep 3

echo "Tunnel setup complete!"
echo "Tunnel Server PID: $TUNNEL_PID"
echo "Gateway PID: $GATEWAY_PID"
echo "Tunnel Client PID: $TUNNEL_CLIENT_PID"
echo ""
echo "Check logs:"
echo "  tail -f /tmp/tunnel-server.log"
echo "  tail -f /tmp/gateway.log"
echo ""
echo "To stop: kill $TUNNEL_PID $GATEWAY_PID $TUNNEL_CLIENT_PID"
```

Make it executable:
```bash
chmod +x test-tunnel.sh
./test-tunnel.sh
```

