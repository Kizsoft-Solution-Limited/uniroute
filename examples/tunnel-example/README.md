# Tunnel Example

This example demonstrates how to use UniRoute CLI to tunnel a local service.

## Prerequisites

1. **UniRoute CLI installed**
   ```bash
   # One-line install
   curl -fsSL https://raw.githubusercontent.com/Kizsoft-Solution-Limited/uniroute/main/scripts/install.sh | bash
   ```

2. **Authenticated with UniRoute**
   ```bash
   uniroute auth login
   ```

## Example 1: Tunnel a Local Web Server

### Step 1: Start a local web server

```bash
# Using Python
python3 -m http.server 8080

# Or using Node.js
npx http-server -p 8080

# Or using Go
go run -m http.server 8080
```

### Step 2: Tunnel the service

```bash
# In another terminal
uniroute http 8080
```

### Step 3: Access your service

You'll get a public URL like:
```
🌍 Public URL:   http://abc123.localhost:8080
```

Open this URL in your browser to access your local server from anywhere!

## Example 2: Tunnel a Database (MySQL)

### Step 1: Start MySQL locally

```bash
# Make sure MySQL is running on port 3306
mysql.server start  # macOS
# or
sudo systemctl start mysql  # Linux
```

### Step 2: Tunnel MySQL

```bash
uniroute tcp 3306
```

### Step 3: Connect from remote

```bash
# Use the public URL shown in the tunnel output
mysql -h abc123.localhost -P 20001 -u root -p
```

## Example 3: Tunnel PostgreSQL with TLS

### Step 1: Start PostgreSQL locally

```bash
# Make sure PostgreSQL is running on port 5432
pg_ctl start  # macOS
# or
sudo systemctl start postgresql  # Linux
```

### Step 2: Tunnel PostgreSQL with TLS

```bash
uniroute tls 5432
```

### Step 3: Connect from remote

```bash
# Use the public URL shown in the tunnel output
psql "postgresql://user:password@xyz789.localhost:20002/dbname?sslmode=require"
```

## Example 4: Multiple Tunnels

You can run multiple tunnels simultaneously:

```bash
# Terminal 1: Web server
uniroute http 8080

# Terminal 2: API server
uniroute http 3000

# Terminal 3: Database
uniroute tcp 3306
```

Each tunnel gets its own unique subdomain and port.

## Example 5: Using Configuration File

Create `~/.uniroute/tunnels.json`:

```json
{
  "version": "1.0",
  "tunnels": [
    {
      "name": "Web Server",
      "protocol": "http",
      "local_addr": "localhost:8080",
      "enabled": true
    },
    {
      "name": "API Server",
      "protocol": "http",
      "local_addr": "localhost:3000",
      "enabled": true
    },
    {
      "name": "MySQL",
      "protocol": "tcp",
      "local_addr": "localhost:3306",
      "enabled": true
    }
  ]
}
```

Then start all tunnels:

```bash
uniroute tunnel --all
```

## Tips

1. **Resume Tunnels**: If you disconnect, tunnels auto-resume with the same URL
2. **View Status**: `uniroute status` shows all active tunnels
3. **View Logs**: `uniroute logs` shows real-time request logs
4. **List Tunnels**: `uniroute list` shows all tunnels (local + server)

## Troubleshooting

### Connection Refused

**Problem**: Can't connect to public URL

**Solution**:
- Check tunnel is running: `uniroute status`
- Verify local service is running on the specified port
- Check firewall allows connections

### Port Already in Use

**Problem**: Local port is already in use

**Solution**:
- Stop the service using that port
- Or use a different port

### Authentication Required

**Problem**: "authentication required for public tunnel server"

**Solution**:
```bash
uniroute auth login
```

## More Information

- **CLI Usage**: See `docs/CLI_USAGE.md`
- **Tunnel Config**: See `docs/TUNNEL_CONFIG.md`
- **TCP/TLS Tunnels**: See `docs/TCP_TLS_TUNNELS.md`
