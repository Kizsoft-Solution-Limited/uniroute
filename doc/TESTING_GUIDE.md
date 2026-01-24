# Tunnel Testing Guide

## Testing Strategy

### Phase 1: Local CLI Testing (Quick Smoke Tests)
Test each tunnel type locally to ensure basic functionality works before deploying to live server.

### Phase 2: Live Server Testing (Full Integration)
Test on live server for real-world scenarios and production validation.

---

## Phase 1: Local CLI Testing

### Prerequisites
- Local tunnel server running (or connect to dev server)
- CLI authenticated (`uniroute auth login`)
- Test services running locally

### 1. TCP Tunnel Test

**Start a simple TCP server:**
```bash
# Terminal 1: Start a simple TCP echo server
nc -l 3306
# Or use a MySQL test instance if available
```

**Start TCP tunnel:**
```bash
# Terminal 2: Start tunnel
uniroute tcp 3306 testtcp
```

**Test connection:**
```bash
# Terminal 3: Connect to tunnel
# Get the public URL from CLI output (e.g., tunnel-server:20000)
nc <tunnel-server-host> <allocated-port>
# Type something, should echo back
```

**What to check:**
- ✅ Tunnel connects successfully
- ✅ Port is allocated
- ✅ Can connect via public port
- ✅ Data flows both ways
- ✅ Reconnection works (close/reopen tunnel)

### 2. TLS Tunnel Test

**Start a TLS service:**
```bash
# Terminal 1: Start a simple TLS server (or use PostgreSQL with TLS)
# Example with openssl:
openssl s_server -accept 5432 -cert server.crt -key server.key
```

**Start TLS tunnel:**
```bash
# Terminal 2: Start tunnel
uniroute tls 5432 testtls
```

**Test connection:**
```bash
# Terminal 3: Connect with TLS client
openssl s_client -connect <tunnel-server-host>:<allocated-port>
```

**What to check:**
- ✅ Tunnel connects successfully
- ✅ TLS handshake works
- ✅ Encrypted data flows
- ✅ Certificate validation (if applicable)

### 3. UDP Tunnel Test

**Start a UDP service:**
```bash
# Terminal 1: Start a simple UDP server
nc -u -l 53
```

**Start UDP tunnel:**
```bash
# Terminal 2: Start tunnel
uniroute udp 53 testudp
```

**Test connection:**
```bash
# Terminal 3: Send UDP packet
echo "test" | nc -u <tunnel-server-host> <allocated-port>
```

**What to check:**
- ✅ Tunnel connects successfully
- ✅ UDP port is allocated
- ✅ Packets are received
- ✅ Bidirectional UDP works

### 4. HTTP Tunnel (Already Tested ✅)
```bash
uniroute http 3000 testhttp
# Visit http://testhttp.localhost:8055
```

---

## Phase 2: Live Server Testing

After local tests pass, test on live server for:
- Real network conditions
- DNS resolution
- Production-like environment
- Rate limiting
- Authentication
- Custom domains (HTTP only)

### Live Server Test Checklist

#### HTTP Tunnel
- [ ] Create tunnel: `uniroute http 3000 myapp`
- [ ] Access via subdomain: `https://myapp.uniroute.co`
- [ ] Test with custom domain: `uniroute domain example.com myapp`
- [ ] Verify DNS: `uniroute domain verify example.com`
- [ ] Access via custom domain: `https://example.com`
- [ ] Test rate limiting
- [ ] Test authentication (unauthenticated users)
- [ ] Test reconnection

#### TCP Tunnel
- [ ] Create tunnel: `uniroute tcp 3306 mydb`
- [ ] Connect via allocated port: `mysql -h <server> -P <port>`
- [ ] Test data flow
- [ ] Test reconnection
- [ ] Test rate limiting

#### TLS Tunnel
- [ ] Create tunnel: `uniroute tls 5432 mydb`
- [ ] Connect with TLS client
- [ ] Test encrypted connection
- [ ] Test reconnection
- [ ] Test rate limiting

#### UDP Tunnel
- [ ] Create tunnel: `uniroute udp 53 dns`
- [ ] Send UDP packets
- [ ] Test bidirectional communication
- [ ] Test reconnection
- [ ] Test rate limiting

---

## Quick Test Commands

### Test All Protocols Locally (Quick Smoke Test)
```bash
# HTTP (already tested)
uniroute http 3000 testhttp

# TCP
uniroute tcp 3306 testtcp

# TLS  
uniroute tls 5432 testtls

# UDP
uniroute udp 53 testudp
```

### Test Reconnection
```bash
# Start tunnel
uniroute tcp 3306 testtcp

# Close tunnel (Ctrl+C)
# Wait a few seconds
# Resume tunnel
uniroute resume testtcp
```

### Test Rate Limiting
```bash
# Create tunnel with API key that has rate limits
uniroute http 3000 myapp

# Make many requests quickly
for i in {1..100}; do curl http://myapp.localhost:8055; done
# Should hit rate limit
```

---

## What to Look For

### ✅ Success Indicators
- Tunnel connects without errors
- Port is allocated correctly
- Can connect from external client
- Data flows correctly
- Reconnection works
- Rate limiting applies
- Authentication works

### ❌ Failure Indicators
- Connection errors
- Port allocation failures
- Data not flowing
- Reconnection fails
- Rate limits not applied
- Authentication issues

---

## Recommended Testing Order

1. **Local HTTP** ✅ (Already done)
2. **Local TCP** (Quick test - 5 min)
3. **Local TLS** (Quick test - 5 min)
4. **Local UDP** (Quick test - 5 min)
5. **Live Server HTTP** (Full test - 10 min)
6. **Live Server TCP** (Full test - 10 min)
7. **Live Server TLS** (Full test - 10 min)
8. **Live Server UDP** (Full test - 10 min)

**Total time: ~1 hour for comprehensive testing**

---

## Notes

- **Custom domains only work with HTTP tunnels** (as documented)
- **TCP/TLS/UDP use port-based routing**, not hostname-based
- **Rate limiting applies to all tunnel types** (recently implemented)
- **Authentication applies to all tunnel types** (recently implemented)
