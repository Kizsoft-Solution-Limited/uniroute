# Opening a Tunnel

Learn how to create and manage tunnels in UniRoute.

## Basic Usage

### Create a Tunnel

```bash
# Simple HTTP tunnel (shortcut - recommended)
uniroute http 8080

# Force new tunnel (don't resume)
uniroute http 8080 --new
```

### Request a Specific Subdomain

```bash
# Request a custom subdomain (shortcut syntax - recommended)
uniroute http 8080 myapp
uniroute http 8080 myapp --new
uniroute tcp 3306 mydb
uniroute tcp 3306 mydb --new

# Request a custom subdomain (flag syntax - also works)
uniroute http 8080 --host myapp
uniroute http 8080 --host myapp --new
```

### Resume a Tunnel

```bash
# Resume last tunnel (shortcut - recommended)
uniroute resume

# Resume specific subdomain
uniroute resume abc123
```

## Advanced Usage

### Multiple Tunnels

Create a configuration file at `~/.uniroute/tunnels.json`:

```json
{
  "version": "1.0",
  "tunnels": [
    {
      "name": "web",
      "protocol": "http",
      "local_addr": "localhost:8080",
      "enabled": true
    },
    {
      "name": "api",
      "protocol": "http",
      "local_addr": "localhost:3000",
      "enabled": true
    }
  ]
}
```

Start all enabled tunnels:

```bash
uniroute tunnel --all
```

### Custom Domain

```bash
# Set domain for existing tunnel (shortcut - recommended)
uniroute domain example.com

# Set domain for specific tunnel by subdomain (shortcut syntax - recommended)
uniroute domain example.com abc123

# Set domain for specific tunnel by subdomain (flag syntax - also works)
uniroute domain example.com --subdomain abc123
```

## Tunnel Management

### List Tunnels

```bash
# List all tunnels (shortcut - recommended)
uniroute list
```

### Disconnect Tunnel

Press `Ctrl+C` in the terminal where the tunnel is running, or use the API:

```bash
curl -X POST https://app.uniroute.co/v1/tunnels/{tunnel-id}/disconnect \
  -H "Authorization: Bearer ur_your-api-key"
```

## Next Steps

- [Dev & Run](/docs/tunnels/dev-run) - Start dev server and tunnel in one go (all languages)
- [Protocols](/docs/tunnels/protocols) - Learn about different protocols
- [Custom Domains](/docs/tunnels/custom-domains) - Configure your own domain
