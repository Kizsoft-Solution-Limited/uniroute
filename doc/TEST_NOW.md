# Test Tunnels Now! 🚀

Since I can't run network tests from the sandbox, here's how to test manually:

## Quick Start (3 Steps)

### 1. Set Environment Variable
```bash
# Tunnel server default port is 8055 (not 8084, which is the gateway)
export UNIROUTE_TUNNEL_URL="localhost:8055"
```

### 2. Run Automated Test Script
```bash
cd /Users/adikekizito/GoDev/uniroute
./scripts/run_all_tests.sh
```

This will automatically:
- ✅ Test TCP tunnel
- ✅ Test UDP tunnel  
- ✅ Test TLS tunnel
- ✅ Show results for each

---

## Manual Testing (Step by Step)

### Test 1: TCP Tunnel

**Terminal 1:**
```bash
nc -l 3306
```

**Terminal 2:**
```bash
cd /Users/adikekizito/GoDev/uniroute
export UNIROUTE_TUNNEL_URL="localhost:8055"
./cli tcp 3306 testtcp --new
```

**Look for output like:**
```
🌍 Public URL:   testtcp.localhost:20000
```

**Terminal 3:**
```bash
# Use the port from Terminal 2 (e.g., 20000)
nc localhost 20000
# Type something - it should appear in Terminal 1
```

---

### Test 2: UDP Tunnel

**Terminal 1:**
```bash
nc -u -l 53
```

**Terminal 2:**
```bash
export UNIROUTE_TUNNEL_URL="localhost:8055"
./cli udp 53 testudp --new
```

**Terminal 3:**
```bash
# Use the allocated port from Terminal 2
echo "Hello UDP" | nc -u localhost <port>
# Check Terminal 1 - should see "Hello UDP"
```

---

### Test 3: TLS Tunnel

**One-time setup:**
```bash
openssl req -x509 -newkey rsa:2048 -keyout /tmp/test.key -out /tmp/test.crt -days 365 -nodes -subj "/CN=localhost"
```

**Terminal 1:**
```bash
openssl s_server -accept 5432 -cert /tmp/test.crt -key /tmp/test.key
```

**Terminal 2:**
```bash
export UNIROUTE_TUNNEL_URL="localhost:8055"
./cli tls 5432 testtls --new
```

**Terminal 3:**
```bash
# Use the allocated port from Terminal 2
openssl s_client -connect localhost:<port>
```

---

## What to Check

For each tunnel:
- ✅ Tunnel connects (no errors in CLI)
- ✅ Port is allocated (shown in output)
- ✅ Can connect from client
- ✅ Data flows correctly
- ✅ Reconnection works

---

## Troubleshooting

**"operation not permitted" or "websocket: bad handshake"**
- Make sure tunnel server is running on port 8055: `lsof -i :8055`
- Port 8084 is the gateway, not the tunnel server
- Start tunnel server: `go run ./cmd/tunnel-server/main.go -port 8055`

**"Connection refused"**
- Check tunnel server is running
- Verify `UNIROUTE_TUNNEL_URL` is set correctly

**"Port already in use"**
- Change local port: `./cli tcp 3307` instead of `3306`

---

## Files Created

- `scripts/run_all_tests.sh` - Automated test script
- `doc/TEST_NOW.md` - This file
- `doc/QUICK_TEST_LOCAL.md` - Detailed guide
- `doc/START_TESTING.md` - Quick reference

**Ready to test!** 🎉
