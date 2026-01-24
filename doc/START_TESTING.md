# Start Testing Now! 🚀

You're authenticated and ready to test. Follow these steps:

## Test 1: TCP Tunnel

### Terminal 1: Start TCP Server
```bash
nc -l 3306
```

### Terminal 2: Start TCP Tunnel
```bash
cd /Users/adikekizito/GoDev/uniroute
uniroute tcp 3306 testtcp
```

**Look for this in the output:**
```
🌍 Public URL:   testtcp.localhost:20000
```
*(Note the port number - it might be different, like 20000, 20001, etc.)*

### Terminal 3: Test Connection
```bash
# Replace 20000 with the actual port from Terminal 2
nc localhost 20000

# Type something and press Enter
# It should appear in Terminal 1!
```

**✅ Success:** Data flows from Terminal 3 → Tunnel → Terminal 1

---

## Test 2: UDP Tunnel

### Terminal 1: Start UDP Server
```bash
nc -u -l 53
```

### Terminal 2: Start UDP Tunnel
```bash
uniroute udp 53 testudp
```

**Note the allocated port from the output**

### Terminal 3: Send UDP Packet
```bash
# Replace 20001 with the actual port from Terminal 2
echo "Hello UDP" | nc -u localhost 20001
```

**✅ Success:** Packet appears in Terminal 1

---

## Test 3: TLS Tunnel

### One-time setup: Create test certificate
```bash
openssl req -x509 -newkey rsa:2048 -keyout /tmp/test.key -out /tmp/test.crt -days 365 -nodes -subj "/CN=localhost"
```

### Terminal 1: Start TLS Server
```bash
openssl s_server -accept 5432 -cert /tmp/test.crt -key /tmp/test.key
```

### Terminal 2: Start TLS Tunnel
```bash
uniroute tls 5432 testtls
```

**Note the allocated port**

### Terminal 3: Test TLS Connection
```bash
# Replace 20002 with the actual port
openssl s_client -connect localhost:20002
```

**✅ Success:** TLS handshake completes

---

## Quick Commands Reference

```bash
# Check auth status
uniroute auth status

# Start TCP tunnel
uniroute tcp 3306 testtcp

# Start UDP tunnel
uniroute udp 53 testudp

# Start TLS tunnel
uniroute tls 5432 testtls

# Start HTTP tunnel (already tested)
uniroute http 3000 testhttp
```

---

## What to Check

For each tunnel:
- ✅ Tunnel connects (no errors)
- ✅ Port allocated (shown in CLI)
- ✅ Can connect from client
- ✅ Data flows correctly
- ✅ Reconnection works

---

## Troubleshooting

**"Connection refused"**
- Make sure test server is running in Terminal 1
- Check the port number matches

**"Port already in use"**
- Change local port: `uniroute tcp 3307` instead of `3306`

**"Tunnel not connecting"**
- Check tunnel server is running on localhost:8084
- Verify authentication: `uniroute auth status`
