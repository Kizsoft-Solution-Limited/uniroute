# Quick Local Tunnel Testing Guide

## Prerequisites

1. **Tunnel server running** (or connect to dev server)
   ```bash
   # If running locally, start tunnel server:
   go run ./cmd/tunnel-server/main.go -port 8080
   ```

2. **CLI authenticated** (if using production server)
   ```bash
   uniroute auth login
   ```

3. **Netcat installed** (for testing TCP/UDP)
   ```bash
   # macOS: usually pre-installed
   # Linux: sudo apt-get install netcat
   ```

---

## Test 1: TCP Tunnel (5 minutes)

### Step 1: Start TCP Server (Terminal 1)
```bash
# Simple echo server on port 3306
nc -l 3306
```

### Step 2: Start TCP Tunnel (Terminal 2)
```bash
uniroute tcp 3306 testtcp
```

**Look for output like:**
```
🌍 Public URL:   testtcp.localhost:20000
```

### Step 3: Test Connection (Terminal 3)
```bash
# Connect to the allocated port (check CLI output for exact port)
nc localhost 20000

# Type something, press Enter
# It should appear in Terminal 1 (the TCP server)
```

**✅ Success:** Data flows from Terminal 3 → Tunnel → Terminal 1

---

## Test 2: UDP Tunnel (5 minutes)

### Step 1: Start UDP Server (Terminal 1)
```bash
nc -u -l 53
```

### Step 2: Start UDP Tunnel (Terminal 2)
```bash
uniroute udp 53 testudp
```

**Look for output like:**
```
🌍 Public URL:   testudp.localhost:20001
```

### Step 3: Send UDP Packet (Terminal 3)
```bash
# Send test packet (use the allocated port from CLI)
echo "Hello UDP" | nc -u localhost 20001
```

**✅ Success:** Packet appears in Terminal 1 (UDP server)

---

## Test 3: TLS Tunnel (5 minutes)

### Step 1: Create Test Certificate (One-time setup)
```bash
# Create self-signed certificate for testing
openssl req -x509 -newkey rsa:2048 -keyout /tmp/test.key -out /tmp/test.crt -days 365 -nodes -subj "/CN=localhost"
```

### Step 2: Start TLS Server (Terminal 1)
```bash
openssl s_server -accept 5432 -cert /tmp/test.crt -key /tmp/test.key
```

### Step 3: Start TLS Tunnel (Terminal 2)
```bash
uniroute tls 5432 testtls
```

**Look for output like:**
```
🌍 Public URL:   testtls.localhost:20002
```

### Step 4: Test TLS Connection (Terminal 3)
```bash
# Connect with TLS client (use allocated port)
openssl s_client -connect localhost:20002 -verify_return_error
```

**✅ Success:** TLS handshake completes, encrypted connection established

---

## Test 4: HTTP Tunnel (Already Tested ✅)

```bash
# Terminal 1: Start web server
python3 -m http.server 3000

# Terminal 2: Start tunnel
uniroute http 3000 testhttp

# Terminal 3: Visit in browser
# http://testhttp.localhost:8055
```

---

## Quick Test All (One Command)

Run this to test all protocols quickly:

```bash
# Terminal 1: Start all test servers
nc -l 3306 &  # TCP
nc -u -l 53 &  # UDP
# TLS requires cert setup first

# Terminal 2: Start all tunnels
uniroute tcp 3306 testtcp &
uniroute udp 53 testudp &
uniroute http 3000 testhttp &
```

---

## Troubleshooting

### "Connection refused"
- Check tunnel server is running
- Verify port is not already in use
- Check firewall settings

### "Authentication required"
- Run `uniroute auth login`
- Or use local tunnel server (localhost)

### "Port already in use"
- Change the local port (e.g., `uniroute tcp 3307` instead of `3306`)
- Or stop the service using that port

### "Tunnel not connecting"
- Check tunnel server logs
- Verify network connectivity
- Check if tunnel server URL is correct

---

## What to Verify

For each tunnel type, check:
- ✅ Tunnel connects successfully (no errors in CLI)
- ✅ Port is allocated (shown in CLI output)
- ✅ Can connect from external client
- ✅ Data flows correctly (bidirectional)
- ✅ Reconnection works (close/reopen tunnel)
- ✅ Rate limiting applies (if configured)
- ✅ Authentication works (if required)

---

## Next Steps

After local testing passes:
1. Test on live server
2. Test with custom domains (HTTP only)
3. Test rate limiting
4. Test authentication scenarios
5. Test reconnection/resume
