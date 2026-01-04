# 🔐 UniRoute Security Overview

## Executive Summary

UniRoute implements multiple layers of security to protect against common attack vectors including authentication bypass, injection attacks, DDoS, and unauthorized access. This document outlines all security measures implemented across the gateway, tunnel, and CLI components.

---

## 🛡️ Security Layers

### 1. Authentication & Authorization

#### API Key Authentication
- **Location**: `internal/security/apikey_v2.go`
- **Implementation**:
  - ✅ API keys generated with cryptographically secure random bytes (32 bytes)
  - ✅ Keys hashed with **bcrypt** (not plaintext storage)
  - ✅ SHA256 lookup hash for fast database queries
  - ✅ Bcrypt verification hash for secure validation
  - ✅ Support for key expiration
  - ✅ Soft delete (is_active flag) for key revocation
- **Protection**: Prevents unauthorized API access

#### JWT Authentication
- **Location**: `internal/security/jwt.go`
- **Implementation**:
  - ✅ Tokens signed with strong secret (minimum 32 characters)
  - ✅ Configurable expiration
  - ✅ Claims-based user identification
  - ✅ Token validation with proper error handling
- **Protection**: Secure admin endpoint access

#### Tunnel Token Authentication
- **Location**: `internal/tunnel/auth.go`
- **Implementation**:
  - ✅ Cryptographically secure token generation (32 bytes)
  - ✅ SHA256 hashing for fast lookups
  - ✅ Bcrypt hashing for secure storage
  - ✅ Token expiration support
  - ✅ Active/inactive status checking
- **Protection**: Prevents unauthorized tunnel connections

#### CLI Authentication
- **Location**: `cmd/cli/commands/auth.go`
- **Implementation**:
  - ✅ Token-based authentication with public UniRoute server
  - ✅ Secure token storage in `~/.uniroute/auth.json`
  - ✅ Token validation before sensitive operations
  - ✅ Automatic logout on token expiration
- **Protection**: Prevents unauthorized CLI access to user resources

---

### 2. Rate Limiting

#### Gateway Rate Limiting
- **Location**: `internal/security/ratelimit.go`, `internal/api/middleware/ratelimit.go`
- **Implementation**:
  - ✅ **Redis-based** distributed rate limiting
  - ✅ Per-API-key limits (per-minute and per-day)
  - ✅ Per-IP fallback limits
  - ✅ Rate limit headers in responses (`X-RateLimit-*`)
  - ✅ Graceful degradation if Redis unavailable
- **Protection**: Prevents DDoS attacks and API abuse

#### Tunnel Rate Limiting
- **Location**: `internal/tunnel/ratelimit_redis.go`
- **Implementation**:
  - ✅ **Redis-based** rate limiting per tunnel
  - ✅ Per-minute limits
  - ✅ Per-hour limits
  - ✅ Per-day limits
  - ✅ Configurable limits per tunnel
- **Protection**: Prevents tunnel abuse and resource exhaustion

---

### 3. Input Validation & Sanitization

#### Request Validation
- **Location**: `internal/tunnel/security.go`
- **Implementation**:
  - ✅ HTTP method validation (only allowed methods)
  - ✅ Path length limits (max 2048 characters)
  - ✅ Header size limits (max 8KB)
  - ✅ Path traversal prevention (`..` removal)
  - ✅ Null byte removal
  - ✅ Double slash removal (`//` → `/`)
- **Protection**: Prevents path traversal, injection attacks, and resource exhaustion

#### API Input Validation
- **Location**: `internal/api/handlers/`
- **Implementation**:
  - ✅ JSON binding validation (Gin framework)
  - ✅ Required field validation
  - ✅ Type checking
  - ✅ Error responses with validation details
- **Protection**: Prevents malformed requests and injection attacks

#### SQL Injection Prevention
- **Location**: `internal/storage/`
- **Implementation**:
  - ✅ **Parameterized queries only** (pgx library)
  - ✅ No string concatenation in SQL
  - ✅ Prepared statements
- **Protection**: Prevents SQL injection attacks

---

### 4. Security Headers

#### Gateway Security Headers
- **Location**: `internal/api/middleware/security_headers.go`
- **Headers**:
  - ✅ `X-Frame-Options: DENY` - Prevents clickjacking
  - ✅ `X-Content-Type-Options: nosniff` - Prevents MIME type sniffing
  - ✅ `X-XSS-Protection: 1; mode=block` - Enables XSS filter
  - ✅ `Content-Security-Policy: default-src 'self'` - Restricts resource loading
  - ✅ `Strict-Transport-Security` - Forces HTTPS (when TLS enabled)
  - ✅ `Referrer-Policy: strict-origin-when-cross-origin` - Controls referrer info
- **Protection**: Prevents XSS, clickjacking, and MIME sniffing attacks

#### Tunnel Security Headers
- **Location**: `internal/tunnel/security.go`
- **Headers**:
  - ✅ CORS headers (configurable origins)
  - ✅ `X-Content-Type-Options: nosniff`
  - ✅ `X-Frame-Options: DENY`
  - ✅ `X-XSS-Protection: 1; mode=block`
  - ✅ `Referrer-Policy: strict-origin-when-cross-origin`
- **Protection**: Prevents XSS and clickjacking in tunneled applications

---

### 5. Network Security

#### IP Whitelisting
- **Location**: `internal/api/middleware/ip_whitelist.go`
- **Implementation**:
  - ✅ Configurable IP allowlist via `IP_WHITELIST` environment variable
  - ✅ Comma-separated IP addresses
  - ✅ Applied globally if configured
  - ✅ Fast lookup using map structure
- **Protection**: Restricts access to known IP addresses

#### TLS/HTTPS Support
- **Location**: Configuration and deployment
- **Implementation**:
  - ✅ HTTPS enforced in production (via reverse proxy/Coolify)
  - ✅ HSTS header enabled
  - ✅ SSL/TLS termination at reverse proxy level
- **Protection**: Encrypts data in transit

---

### 6. Error Handling & Information Disclosure

#### Secure Error Messages
- **Location**: Throughout codebase
- **Implementation**:
  - ✅ Generic error messages to clients
  - ✅ Detailed errors logged server-side only
  - ✅ No sensitive data in error responses
  - ✅ No stack traces in production
- **Protection**: Prevents information disclosure to attackers

---

### 7. Secrets Management

#### Environment Variables
- **Location**: `internal/config/config.go`
- **Implementation**:
  - ✅ All secrets via environment variables
  - ✅ `.env` files for development (not in git)
  - ✅ `.env` in `.gitignore`
  - ✅ No hardcoded secrets
- **Protection**: Prevents secret leakage in code

#### Token Storage
- **Location**: `cmd/cli/commands/auth.go`
- **Implementation**:
  - ✅ CLI tokens stored in `~/.uniroute/auth.json`
  - ✅ File permissions: 0600 (read/write owner only)
  - ✅ Tunnel state stored securely
- **Protection**: Prevents unauthorized access to stored tokens

---

### 8. Database Security

#### Connection Security
- **Location**: `internal/storage/postgres.go`
- **Implementation**:
  - ✅ Connection pooling
  - ✅ SSL mode configurable (`sslmode` parameter)
  - ✅ Credentials from environment variables
  - ✅ Health checks
- **Protection**: Secures database connections

#### Query Security
- **Location**: `internal/storage/`
- **Implementation**:
  - ✅ Parameterized queries only
  - ✅ No SQL string concatenation
  - ✅ Prepared statements
  - ✅ Input validation before queries
- **Protection**: Prevents SQL injection

---

## 🔒 Security by Component

### Gateway Server (`uniroute-gateway`)

**Security Measures:**
1. ✅ API key authentication (bcrypt hashed)
2. ✅ JWT authentication for admin endpoints
3. ✅ Redis-based rate limiting
4. ✅ Security headers middleware
5. ✅ IP whitelisting (optional)
6. ✅ Input validation
7. ✅ SQL injection prevention
8. ✅ Secure error handling

**Attack Vectors Protected:**
- ✅ Unauthorized API access
- ✅ DDoS attacks
- ✅ SQL injection
- ✅ XSS attacks
- ✅ Clickjacking
- ✅ Information disclosure

---

### Tunnel Server (`uniroute-tunnel-server`)

**Security Measures:**
1. ✅ Token-based authentication
2. ✅ Redis-based rate limiting per tunnel
3. ✅ Request validation (method, path, headers)
4. ✅ Path sanitization (traversal prevention)
5. ✅ Security headers
6. ✅ CORS support (configurable)
7. ✅ Request size limits

**Attack Vectors Protected:**
- ✅ Unauthorized tunnel connections
- ✅ Path traversal attacks
- ✅ DDoS attacks
- ✅ XSS attacks
- ✅ Clickjacking
- ✅ Resource exhaustion

---

### Tunnel Client (CLI)

**Security Measures:**
1. ✅ Authentication with public server
2. ✅ Secure token storage (file permissions)
3. ✅ Reconnection with authentication
4. ✅ Subdomain persistence (resume capability)

**Attack Vectors Protected:**
- ✅ Unauthorized tunnel creation
- ✅ Token theft (file permissions)
- ✅ Man-in-the-middle (TLS required)

---

### CLI Tool (`uniroute`)

**Security Measures:**
1. ✅ Authentication required for public server
2. ✅ Secure token storage
3. ✅ Localhost bypass for development
4. ✅ Input validation for commands

**Attack Vectors Protected:**
- ✅ Unauthorized access to user resources
- ✅ Token theft
- ✅ Command injection (input validation)

---

## ⚠️ Security Recommendations

### High Priority

1. **Request Body Size Limits**
   - ⚠️ **Missing**: Maximum request body size limits
   - **Recommendation**: Add `MaxRequestBodySize` middleware
   - **Impact**: Prevents memory exhaustion attacks

2. **Request Timeout**
   - ⚠️ **Missing**: Request timeout configuration
   - **Recommendation**: Add timeout middleware
   - **Impact**: Prevents slowloris attacks

3. **CORS Configuration**
   - ⚠️ **Partial**: CORS exists but needs stricter configuration
   - **Recommendation**: Whitelist specific origins only
   - **Impact**: Prevents unauthorized cross-origin requests

4. **Logging & Monitoring**
   - ⚠️ **Partial**: Basic logging exists
   - **Recommendation**: Add security event logging (failed auth, rate limits)
   - **Impact**: Enables threat detection

5. **WebSocket Security**
   - ⚠️ **Partial**: WebSocket connections need origin validation
   - **Recommendation**: Validate WebSocket origin headers
   - **Impact**: Prevents unauthorized WebSocket connections

### Medium Priority

6. **API Key Rotation**
   - ⚠️ **Missing**: Automatic API key rotation
   - **Recommendation**: Add key rotation policy
   - **Impact**: Reduces impact of key compromise

7. **IP Reputation Checking**
   - ⚠️ **Missing**: IP reputation/blacklist checking
   - **Recommendation**: Integrate with threat intelligence feeds
   - **Impact**: Blocks known malicious IPs

8. **Request Signing**
   - ⚠️ **Missing**: Request signature validation
   - **Recommendation**: Add HMAC signature support
   - **Impact**: Prevents request tampering

9. **Audit Logging**
   - ⚠️ **Missing**: Comprehensive audit logs
   - **Recommendation**: Log all security events
   - **Impact**: Enables security forensics

10. **2FA/MFA Support**
    - ⚠️ **Missing**: Two-factor authentication
    - **Recommendation**: Add TOTP/WebAuthn support
    - **Impact**: Enhances authentication security

### Low Priority

11. **Geolocation Filtering**
    - ⚠️ **Missing**: Geographic IP filtering
    - **Recommendation**: Add country-based filtering
    - **Impact**: Restricts access by geography

12. **CAPTCHA Integration**
    - ⚠️ **Missing**: CAPTCHA for suspicious activity
    - **Recommendation**: Add CAPTCHA on rate limit
    - **Impact**: Prevents automated attacks

13. **Request Fingerprinting**
    - ⚠️ **Missing**: Request fingerprinting for bot detection
    - **Recommendation**: Add fingerprinting middleware
    - **Impact**: Detects automated attacks

---

## 🧪 Security Testing

### Recommended Tests

1. **Penetration Testing**
   - SQL injection attempts
   - XSS payload testing
   - Path traversal attempts
   - Authentication bypass attempts
   - Rate limit bypass attempts

2. **Load Testing**
   - DDoS simulation
   - Rate limit effectiveness
   - Resource exhaustion tests

3. **Security Scanning**
   - Dependency vulnerability scanning
   - Static code analysis
   - Dynamic application security testing (DAST)

---

## 📋 Security Checklist

### Deployment Checklist

- [ ] All secrets in environment variables (not in code)
- [ ] HTTPS/TLS enabled in production
- [ ] Security headers configured
- [ ] Rate limiting enabled
- [ ] IP whitelist configured (if needed)
- [ ] Database SSL enabled
- [ ] Redis authentication enabled
- [ ] Error messages sanitized
- [ ] Logging configured (no sensitive data)
- [ ] Firewall rules configured
- [ ] Regular security updates scheduled

---

## 📚 Additional Resources

- **OWASP Top 10**: https://owasp.org/www-project-top-ten/
- **Go Security Best Practices**: https://go.dev/doc/security/best-practices
- **CWE Top 25**: https://cwe.mitre.org/top25/

---

## 🔄 Security Updates

This document should be reviewed and updated:
- After each major release
- When new security features are added
- When vulnerabilities are discovered
- Quarterly security audits

---

**Last Updated**: 2024
**Version**: 1.0

