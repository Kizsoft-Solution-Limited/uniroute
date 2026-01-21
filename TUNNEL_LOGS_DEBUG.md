# Tunnel Server Logs - Debugging Guide

## How to View Tunnel Server Logs

### 1. Start Tunnel Server with Debug Logging

```bash
# Run with debug logging (shows all log levels)
./bin/uniroute-tunnel-server --log-level debug --port 8080

# Or if running from source:
go run cmd/tunnel-server/main.go --log-level debug --port 8080
```

### 2. Logs are Written to STDOUT

All logs are written to **standard output (stdout)**, so you'll see them in your terminal where you started the tunnel server.

## Key Log Messages to Check

### ✅ Database Connection Status

**Look for this when tunnel server starts:**
```
{"level":"info","message":"Database connected, request logging enabled"}
```

**If you see this instead, database is NOT connected:**
```
{"level":"warn","error":"...","message":"Failed to connect to PostgreSQL, continuing without database"}
```

**⚠️ If database is not connected, tunnels will NOT be saved!**

### 🔍 Tunnel Creation Logs

When a tunnel is created, look for these messages in order:

#### 1. Check if tunnel should be saved (DEBUG level)
```
{"level":"debug","repository_available":true,"is_resume":false,"tunnel_id":"...","subdomain":"...","message":"Checking if tunnel should be saved to database"}
```

**Important fields:**
- `repository_available: true` = Database is connected ✅
- `repository_available: false` = Database NOT connected ❌ (tunnel won't save)
- `is_resume: false` = New tunnel (will be saved) ✅
- `is_resume: true` = Resuming existing tunnel (won't save again) ⚠️

#### 2. Attempting to save (INFO level)
```
{"level":"info","tunnel_id":"...","subdomain":"...","user_id":"...","local_url":"...","public_url":"...","message":"Attempting to save tunnel to database"}
```

#### 3. Success or Failure

**✅ Success (with user):**
```
{"level":"info","tunnel_id":"...","subdomain":"...","user_id":"...","message":"Saved tunnel to database with user association"}
```

**✅ Success (without user):**
```
{"level":"info","tunnel_id":"...","subdomain":"...","message":"Saved tunnel to database (unassociated - user can associate later)"}
```

**❌ Failure:**
```
{"level":"error","error":"...","tunnel_id":"...","tunnel_id_length":"...","subdomain":"...","user_id":"...","local_url":"...","public_url":"...","message":"Failed to save tunnel to database - check database connection and tunnel ID format"}
```

## Common Issues and Solutions

### Issue 1: `repository_available: false`

**Problem:** Database is not connected.

**Solution:**
1. Check your `.env` file has `DATABASE_URL` set:
   ```bash
   DATABASE_URL=postgres://user:password@localhost:5432/uniroute?sslmode=disable
   ```

2. Verify database is running:
   ```bash
   psql $DATABASE_URL -c "SELECT 1;"
   ```

3. Check tunnel server startup logs for database connection errors.

### Issue 2: `is_resume: true` (tunnel thinks it's resuming)

**Problem:** Tunnel client is sending a saved tunnel ID/subdomain, so server thinks it's resuming an existing tunnel and doesn't save it again.

**Solution:**
- This is expected behavior - resumed tunnels are not saved again
- To create a NEW tunnel, delete the saved tunnel state:
  ```bash
  rm ~/.uniroute/tunnel_state.json
  ```
- Or use `--new` flag if your CLI supports it

### Issue 3: Database save fails with error

**Check the error message:**
- **"invalid input syntax for type uuid"** = Tunnel ID is not a valid UUID format
  - This should be auto-fixed now, but check logs for "Tunnel ID was updated to valid UUID format"
- **"duplicate key value violates unique constraint"** = Tunnel with this ID already exists
- **"connection refused"** = Database connection issue
- **"relation \"tunnels\" does not exist"** = Database migrations not run

**Solution:**
1. Run database migrations:
   ```bash
   ./scripts/run_migrations.sh
   ```

2. Check database schema:
   ```sql
   \d tunnels
   ```

### Issue 4: JWT Token Issues

**Look for:**
```
{"level":"warn","tunnel_id":"...","message":"Failed to extract user ID from token (tunnel will be unassociated) - JWT validator may not be configured or token invalid"}
```

**Solution:**
- Tunnel will still be saved, just without user association
- To enable user association, ensure `JWT_SECRET` is set in tunnel server environment
- JWT_SECRET must match the gateway's JWT_SECRET

## Quick Diagnostic Commands

### Check if tunnel server is running with database:
```bash
# Look for "Database connected" in logs
ps aux | grep tunnel-server
```

### Check database for tunnels:
```sql
SELECT id, subdomain, user_id, status, created_at 
FROM tunnels 
ORDER BY created_at DESC 
LIMIT 10;
```

### Check recent tunnel server logs:
```bash
# If running in terminal, scroll up to see startup logs
# Or if using systemd/docker, check service logs:
journalctl -u tunnel-server -f
# or
docker logs tunnel-server -f
```

## Example: Full Log Flow for Successful Tunnel Creation

```
{"level":"info","port":8080,"environment":"development","message":"Starting UniRoute Tunnel Server"}
{"level":"info","message":"Database connected, request logging enabled"}
{"level":"info","message":"JWT validator configured - tunnels will be auto-associated with authenticated users"}
{"level":"info","port":8080,"message":"Tunnel server starting"}

... (when tunnel connects) ...

{"level":"debug","repository_available":true,"is_resume":false,"tunnel_id":"abc123...","subdomain":"xyz789","message":"Checking if tunnel should be saved to database"}
{"level":"info","tunnel_id":"abc123...","user_id":"user-uuid","message":"Extracted user ID from auth token - tunnel will be associated with user"}
{"level":"info","tunnel_id":"abc123...","subdomain":"xyz789","user_id":"user-uuid","local_url":"http://localhost:8890","public_url":"http://xyz789.localhost:8080","message":"Attempting to save tunnel to database"}
{"level":"info","tunnel_id":"abc123...","subdomain":"xyz789","user_id":"user-uuid","message":"Saved tunnel to database with user association"}
```

## Need More Help?

If tunnels still aren't saving after checking these logs:
1. Share the full tunnel server startup logs
2. Share the tunnel creation logs (with `--log-level debug`)
3. Check database connection and migrations
4. Verify `DATABASE_URL` environment variable is set correctly
