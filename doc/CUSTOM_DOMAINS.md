# Custom Domains in UniRoute

## Overview

Custom domains allow you to use your own domain name (e.g., `example.com`) instead of the default random subdomain (e.g., `abc123.localhost:8055`).

## Supported Tunnel Types

**Currently, custom domains only work with HTTP tunnels.**

### HTTP Tunnels ✅
- **Fully supported**: Custom domains work perfectly with HTTP tunnels
- Uses hostname-based routing (checks Host header)
- Example: `https://example.com` → routes to your HTTP tunnel

### TCP/TLS/UDP Tunnels ❌
- **Not supported**: These protocols use port-based routing
- TCP/TLS tunnels get allocated ports (e.g., `tunnel-server:20000`)
- UDP tunnels get allocated ports (e.g., `tunnel-server:20001`)
- Custom domains cannot be used because routing is based on port numbers, not hostnames

## How Custom Domains Work

### 1. Add Domain to Account
```bash
# Add domain (not assigned to any tunnel yet)
uniroute domain example.com
```

### 2. Assign Domain to Tunnel
```bash
# Assign to tunnel by subdomain
uniroute domain example.com abc123

# Or assign by tunnel ID
uniroute domain example.com --tunnel-id <tunnel-id>
```

### 3. Configure DNS
Add a CNAME record in your DNS provider:
```
Type: CNAME
Name: example.com (or @ for root domain)
Target: tunnel.uniroute.co
```

### 4. Verify DNS
```bash
uniroute domain verify example.com
```

### 5. Use Your Domain
Once DNS is configured and verified, your custom domain will automatically route to your HTTP tunnel:
- `https://example.com` → Your HTTP tunnel
- Works with all HTTP requests (GET, POST, etc.)

## Technical Details

### How It Works Internally

1. **Domain Storage**: Custom domain is stored in the database with the tunnel's subdomain
2. **DNS Routing**: Your DNS provider routes `example.com` → `tunnel.uniroute.co`
3. **Request Routing**: When a request arrives with `Host: example.com`:
   - Server checks `GetTunnelByCustomDomain(hostname)`
   - Finds the tunnel's subdomain
   - Routes to the active tunnel connection

### Code Flow

1. **HTTP Request Arrives**: `handleHTTPRequest()` receives request with `Host: example.com`
2. **Subdomain Check**: First tries to find tunnel by subdomain (fails for custom domains)
3. **Custom Domain Lookup**: Calls `repository.GetTunnelByCustomDomain(hostname)`
4. **Tunnel Resolution**: Gets tunnel's subdomain from database
5. **Route to Tunnel**: Uses subdomain to find active tunnel connection
6. **Forward Request**: Forwards HTTP request through the tunnel

## Limitations

- **HTTP Only**: Custom domains only work with HTTP tunnels
- **TCP/TLS/UDP**: These protocols use port-based routing, not hostname-based
- **DNS Required**: Must configure CNAME record pointing to `tunnel.uniroute.co`
- **Verification**: DNS must be verified before domain works

## Future Enhancements

To support custom domains for TCP/TLS/UDP tunnels, you would need:
- Port-based domain mapping (e.g., `example.com:20000` → TCP tunnel)
- Or SRV record support for service discovery
- Different routing mechanism than HTTP hostname-based routing

## Examples

### HTTP Tunnel with Custom Domain
```bash
# 1. Create HTTP tunnel
uniroute http 3000 myapp

# 2. Add and assign custom domain
uniroute domain example.com myapp

# 3. Configure DNS: CNAME example.com → tunnel.uniroute.co

# 4. Verify DNS
uniroute domain verify example.com

# 5. Access via custom domain
curl https://example.com
```

### Resume Domain Assignment
```bash
# Resume by subdomain
uniroute domain resume myapp

# Resume by domain name
uniroute domain resume example.com
```

## Domain Management Commands

```bash
# List all domains
uniroute domain list

# Show domain details
uniroute domain show example.com

# Verify DNS configuration
uniroute domain verify example.com

# Remove domain
uniroute domain remove example.com
```
