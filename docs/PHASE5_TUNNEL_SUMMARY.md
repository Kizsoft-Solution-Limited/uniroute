# Phase 5: Custom Tunnel - Domain Management Summary

## Overview

Phase 5 focuses on domain and subdomain management. **Note: SSL/TLS is handled by Coolify**, so we don't need to implement ACME/Let's Encrypt certificate management in the tunnel server.

## ✅ Completed Features

### 1. Domain Manager
- **Location**: `internal/tunnel/domain.go`
- **Features**:
  - Subdomain allocation with conflict checking
  - Public URL generation (HTTP/HTTPS based on domain)
  - Custom domain validation
  - DNS validation (TXT, CNAME records)
  - DNS propagation waiting

### 2. Subdomain Pool
- **Location**: `internal/tunnel/domain.go`
- **Features**:
  - Thread-safe subdomain allocation
  - Subdomain release/reuse
  - Random subdomain generation (8 characters)

### 3. DNS Validator
- **Location**: `internal/tunnel/domain.go`
- **Features**:
  - TXT record validation (for domain ownership)
  - CNAME record validation
  - DNS propagation waiting

### 4. Server Integration
- **Location**: `internal/tunnel/server.go`
- **Features**:
  - Domain manager integration
  - Automatic subdomain allocation
  - Public URL generation with HTTPS support (when base domain set)
  - Coolify-compatible (SSL handled by reverse proxy)

## Architecture

### Domain Management Flow

```
Client Connects → Domain Manager
  ├─ Allocate Subdomain (check conflicts)
  ├─ Generate Public URL
  │   ├─ With base domain → HTTPS (Coolify handles SSL)
  │   └─ Without base domain → HTTP localhost
  └─ Return to Client
```

### SSL/TLS Handling

**Important**: SSL/TLS termination is handled by **Coolify** (or any reverse proxy):
- Tunnel server runs on HTTP internally
- Coolify handles Let's Encrypt certificates
- Coolify terminates SSL and forwards to tunnel server
- No ACME client needed in tunnel server

## New Files Created

- `internal/tunnel/domain.go` - Domain and subdomain management
- `internal/tunnel/domain_test.go` - Domain management tests

## Files Modified

- `internal/tunnel/server.go` - Integrated domain manager
- `cmd/tunnel-server/main.go` - Added domain manager configuration

## Removed Files

- `internal/tunnel/ssl.go` - **Removed** (not needed - Coolify handles SSL)

## Configuration

### Environment Variables

```bash
# Base domain for tunnel subdomains (e.g., uniroute.dev)
TUNNEL_BASE_DOMAIN=uniroute.dev

# Database (for tunnel persistence)
DATABASE_URL=postgres://user:pass@localhost/uniroute

# Redis (for rate limiting)
REDIS_URL=redis://localhost:6379
```

### Example Setup

```bash
# Set base domain
export TUNNEL_BASE_DOMAIN=uniroute.dev

# Start tunnel server
./bin/uniroute-tunnel-server --port 8080
```

## Usage

### Basic Subdomain Allocation

```go
domainManager := tunnel.NewDomainManager("uniroute.dev", logger)
subdomain, err := domainManager.AllocateSubdomain(ctx, repository)
// Returns: "a1b2c3d4" → Public URL: https://a1b2c3d4.uniroute.dev
```

### Custom Domain Validation

```go
err := domainManager.ValidateCustomDomain(ctx, "example.com")
// Validates DNS resolution and format
```

### DNS Validation

```go
validator := tunnel.NewDNSValidator(logger)

// Validate TXT record
valid, err := validator.ValidateTXTRecord(ctx, "_acme-challenge.example.com", "expected-value")

// Validate CNAME
valid, err := validator.ValidateCNAMERecord(ctx, "tunnel.example.com", "target.uniroute.dev")
```

## Testing

### Unit Tests

```bash
# Run domain tests
go test ./internal/tunnel -v -run TestDomain

# Run all Phase 5 tests
go test ./internal/tunnel -v
```

### Test Coverage

- ✅ Domain manager creation
- ✅ Public URL generation (HTTP/HTTPS)
- ✅ Subdomain pool allocation
- ✅ Subdomain release
- ✅ Custom domain validation
- ✅ DNS validation (basic)

## Coolify Integration

### How It Works

1. **Tunnel Server**: Runs on HTTP (port 8080)
2. **Coolify**: 
   - Handles SSL/TLS termination
   - Manages Let's Encrypt certificates
   - Routes HTTPS traffic to tunnel server
3. **DNS**: Points to Coolify's load balancer
4. **Result**: Secure HTTPS tunnels without certificate management in tunnel server

### Coolify Configuration

```yaml
# docker-compose.yml or Coolify service
services:
  tunnel-server:
    image: uniroute-tunnel-server
    ports:
      - "8080:8080"
    environment:
      - TUNNEL_BASE_DOMAIN=uniroute.dev
      - DATABASE_URL=postgres://...
      - REDIS_URL=redis://...
```

Coolify will:
- Automatically obtain SSL certificates
- Configure HTTPS
- Route `*.uniroute.dev` to tunnel server

## Current Limitations

1. **No ACME Client**: SSL handled by Coolify (intentional)
2. **Basic DNS Validation**: Simple checks (can be enhanced)
3. **No Wildcard DNS Management**: Requires manual DNS setup
4. **No Custom Domain API**: Manual domain addition (can add API later)

## Next Steps (Future Phases)

- [ ] Custom domain API endpoints
- [ ] Domain ownership verification workflow
- [ ] Wildcard DNS automation
- [ ] Domain analytics
- [ ] Subdomain reservation system

## Test Results

Run tests to verify:
```bash
go test ./internal/tunnel -v
```

Expected: All Phase 5 tests passing ✅

## Production Readiness

Phase 5 is **production-ready** with:
- ✅ Domain and subdomain management
- ✅ DNS validation
- ✅ Coolify-compatible (SSL handled externally)
- ✅ Thread-safe subdomain allocation
- ✅ Public URL generation

**Note**: For production, ensure:
1. DNS is configured to point `*.uniroute.dev` to Coolify
2. Coolify is configured to handle SSL certificates
3. Tunnel server runs behind Coolify's reverse proxy

Ready for deployment with Coolify! 🚀

