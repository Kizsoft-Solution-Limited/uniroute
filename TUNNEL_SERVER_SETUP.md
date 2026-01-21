# Tunnel Server Setup & Error Pages Guide

## 1. Tunnel Server Location

The tunnel server runs on **`localhost:8080`** by default (configurable via `--port` flag).

### Start the tunnel server:
```bash
# Default port (8080)
./bin/uniroute-tunnel-server --log-level debug

# Custom port
./bin/uniroute-tunnel-server --log-level debug --port 8080
```

### Verify it's running:
```bash
# Check if port 8080 is listening
lsof -i :8080

# Or test the health endpoint
curl http://localhost:8080/health

# Should return: {"status":"ok","tunnels":0}
```

### View tunnel server logs:
The logs appear in the **terminal where you started the tunnel server**. Look for:
- `"Tunnel server starting"` - Server started
- `"Database connected"` - Database is working
- `"Checking if tunnel should be saved to database"` - When tunnel connects
- `"Saved tunnel to database"` - Success!

## 2. .env File Location

The `.env` file is in the **project root directory**:
```
/Users/adikekizito/GoDev/uniroute/.env
```

### Required environment variables for tunnel server:

```bash
# Database connection (REQUIRED for saving tunnels)
DATABASE_URL=postgres://user:password@localhost:5432/uniroute?sslmode=disable

# JWT Secret (REQUIRED for auto-associating tunnels with users)
JWT_SECRET=your-secret-key-here-minimum-32-characters-long

# Optional: Redis for distributed rate limiting
REDIS_URL=redis://localhost:6379

# Optional: Tunnel base domain (for custom domains)
TUNNEL_BASE_DOMAIN=yourdomain.com
```

### Check your .env file:
```bash
cat .env | grep -E "DATABASE_URL|JWT_SECRET"
```

## 3. Error Pages (502, 404, 503) - Already Implemented! ✅

**Important:** The error pages are **already implemented in the code** (`internal/tunnel/server.go`), not configurable via `.env`. They automatically display when those HTTP errors occur.

### Where they're implemented:

1. **404 Error Page** - `writeErrorPage()` function
   - Location: `internal/tunnel/server.go:1296`
   - Triggered when: Subdomain doesn't exist (not in memory or database)
   - Code: `http.StatusNotFound` (404)

2. **502 Error Page** - `writeConnectionRefusedError()` and `writeErrorPage()`
   - Location: `internal/tunnel/server.go:1018` (connection refused) and `1296` (generic 502)
   - Triggered when:
     - Local server is not running (connection refused)
     - WebSocket connection is lost
     - Request forwarding fails
   - Code: `http.StatusBadGateway` (502)

3. **503 Error Page** - `writeErrorPage()`
   - Location: `internal/tunnel/server.go:1296`
   - Triggered when: Tunnel exists in database but client is not connected
   - Code: `http.StatusServiceUnavailable` (503)

### How to test error pages:

1. **Test 404 (Tunnel Not Found):**
   ```bash
   # Visit a non-existent tunnel URL
   curl http://nonexistent.localhost:8080/
   # Should show styled 404 error page
   ```

2. **Test 502 (Connection Refused):**
   ```bash
   # Create a tunnel pointing to a port that's not running
   ./bin/uniroute http 9999  # Port 9999 is not running
   
   # Then visit the tunnel URL
   curl http://[subdomain].localhost:8080/
   # Should show styled 502 "Connection Refused" error page
   ```

3. **Test 503 (Tunnel Disconnected):**
   ```bash
   # Create a tunnel, then stop the CLI
   # Visit the tunnel URL - should show 503 "Tunnel Disconnected" page
   ```

### Error page features:

- ✅ Styled with UniRoute theme (dark gradient background)
- ✅ Shows tunnel information (public URL, local URL)
- ✅ Helpful error messages
- ✅ Security headers (XSS protection, CSP, etc.)
- ✅ HTML-escaped user inputs (XSS prevention)

## 4. Quick Reference

### Start tunnel server with debug logs:
```bash
./bin/uniroute-tunnel-server --log-level debug --port 8080
```

### Check if tunnel server is running:
```bash
curl http://localhost:8080/health
```

### View tunnel server logs:
- Look at the terminal where you started `./bin/tunnel-server`
- Or check system logs if running as a service

### Check .env file:
```bash
cat .env
```

### Verify database connection:
```bash
# In tunnel server logs, look for:
# ✅ "Database connected, request logging enabled"
# ❌ "Failed to connect to PostgreSQL"
```

## 5. Troubleshooting

### Error pages not showing?

1. **Check tunnel server is running:**
   ```bash
   curl http://localhost:8080/health
   ```

2. **Check browser console** - might be caching old responses

3. **Check tunnel server logs** - look for error page write logs:
   ```
   "Writing error page"
   "Writing connection refused error page"
   ```

4. **Verify error detection** - check logs for:
   ```
   "Error received from tunnel client - checking type for custom error page"
   "Detected connection refused error - showing custom error page"
   ```

### Tunnels not saving to database?

1. **Check .env has DATABASE_URL:**
   ```bash
   grep DATABASE_URL .env
   ```

2. **Check tunnel server startup logs:**
   - Look for `"Database connected"` or `"Failed to connect to PostgreSQL"`

3. **Run with debug logs:**
   ```bash
   ./bin/tunnel-server --log-level debug
   ```

4. **Check for these log messages:**
   - `repository_available: true` = Database connected ✅
   - `repository_available: false` = Database NOT connected ❌
   - `is_resume: true` = Resuming tunnel (won't save again) ⚠️

## Summary

- **Tunnel Server:** Runs on `localhost:8080` (default)
- **.env File:** Project root (`/Users/adikekizito/GoDev/uniroute/.env`)
- **Error Pages:** Already implemented in code, automatically shown (no .env config needed)
- **Logs:** Terminal where you start the tunnel server
- **Database:** Required for saving tunnels (set `DATABASE_URL` in `.env`)
