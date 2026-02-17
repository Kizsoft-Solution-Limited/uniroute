# Tunnels

UniRoute tunnels allow you to expose local services to the internet, similar to ngrok.

## Overview

Tunnels create a secure connection between your local machine and the UniRoute edge network. When someone visits your public URL, the request is forwarded through the tunnel to your local server.

## Quick Start

```bash
# Expose a local web server (shortcut - recommended)
uniroute http 8080

# You'll get a public URL
# http://abc123.localhost:8055 -> http://localhost:8080
```

## Protocols

UniRoute supports multiple protocols with convenient shortcuts:

### HTTP Tunnels

```bash
# HTTP tunnel (shortcut - recommended)
uniroute http 8080
```

### TCP Tunnels

```bash
# TCP tunnel for databases (shortcut - recommended)
uniroute tcp 3306
```

### TLS Tunnels

```bash
# TLS tunnel for secure connections (shortcut - recommended)
uniroute tls 5432
```

### UDP Tunnels

```bash
# UDP tunnel for DNS, gaming, etc. (shortcut - recommended)
uniroute udp 53
```

## Features

- **Persistent Tunnels** - Tunnels survive CLI restarts
- **Custom Subdomains** - Request your own subdomain
- **Custom Domains** - Use your own domain
- **Multiple Tunnels** - Run multiple tunnels simultaneously
- **Auto-Reconnection** - Automatic reconnection on disconnect

## Next Steps

- [Opening a Tunnel](/docs/tunnels/opening) - Detailed tunnel creation guide
- [Dev & Run](/docs/tunnels/dev-run) - Start your dev server and tunnel (Laravel, Vue, Rails, etc.)
- [Protocols](/docs/tunnels/protocols) - Learn about each protocol
- [Custom Domains](/docs/tunnels/custom-domains) - Use your own domain
