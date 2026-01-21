# Quick Start: Running Tunnel Server with Debug Logs

## Step 1: Start Tunnel Server (in a separate terminal)

The tunnel server is a **separate process** that must be running before you use the CLI.

### Build the tunnel server:
```bash
make build
# or
go build -o bin/uniroute-tunnel-server cmd/tunnel-server/main.go
```

### Start tunnel server with debug logs:

```bash
# Terminal 1: Start tunnel server with debug logging
./bin/uniroute-tunnel-server --log-level debug --port 8080
```

**Look for these startup messages:**
- ✅ `"Database connected, request logging enabled"` = Database is working
- ❌ `"Failed to connect to PostgreSQL"` = Database not connected (tunnels won't save!)

## Step 2: Use CLI (in another terminal)

```bash
# Terminal 2: Create tunnel using CLI
./bin/uniroute http 8890
```

## Step 3: Check Logs

**In Terminal 1 (tunnel server), you'll see logs like:**

```
{"level":"info","message":"Database connected, request logging enabled"}
{"level":"debug","repository_available":true,"is_resume":false,"tunnel_id":"...","message":"Checking if tunnel should be saved to database"}
{"level":"info","message":"Attempting to save tunnel to database"}
{"level":"info","message":"Saved tunnel to database"}
```

## Quick Check: Is Database Connected?

When tunnel server starts, look for:
- ✅ `"Database connected"` = Good!
- ❌ `"Failed to connect to PostgreSQL"` = Check your `DATABASE_URL` in `.env`

## Environment Variables Needed

Make sure your `.env` file has:
```bash
DATABASE_URL=postgres://user:password@localhost:5432/uniroute?sslmode=disable
JWT_SECRET=your-secret-key-here-min-32-chars
```

## Troubleshooting

### Tunnel not saving to database?

1. **Check tunnel server logs** (Terminal 1) for:
   - `repository_available: false` = Database not connected
   - `is_resume: true` = Tunnel thinks it's resuming (won't save again)
   - Error messages about database connection

2. **Verify database is running:**
   ```bash
   psql $DATABASE_URL -c "SELECT 1;"
   ```

3. **Check database has migrations:**
   ```bash
   ./scripts/run_migrations.sh
   ```

### See full debugging guide:
See `TUNNEL_LOGS_DEBUG.md` for detailed log message explanations.
