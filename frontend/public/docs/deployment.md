# Deployment

Deploy UniRoute to production environments.

## Self-Hosted Deployment

### Prerequisites

- PostgreSQL 15+
- Redis 7+
- SMTP server (for emails)
- Domain name (optional, for custom domains)

### Docker Compose

```yaml
version: '3.8'
services:
  gateway:
    image: uniroute/gateway:latest
    ports:
      - "8084:8084"
    environment:
      - DATABASE_URL=postgresql://user:pass@postgres:5432/uniroute
      - REDIS_URL=redis://redis:6379
    depends_on:
      - postgres
      - redis

  postgres:
    image: postgres:15
    environment:
      - POSTGRES_DB=uniroute
      - POSTGRES_USER=uniroute
      - POSTGRES_PASSWORD=password
    volumes:
      - postgres_data:/var/lib/postgresql/data

  redis:
    image: redis:7
    volumes:
      - redis_data:/data
```

### Environment Variables

Required environment variables:

```bash
DATABASE_URL=postgresql://user:pass@host:5432/uniroute
REDIS_URL=redis://host:6379
JWT_SECRET=your-secret-key
SMTP_HOST=smtp.example.com
SMTP_USER=user@example.com
SMTP_PASSWORD=password
```

### Coolify (one repo, three apps)

Deploy frontend, backend, and tunnel as three separate Coolify applications from the same Git repository.

| App | Base Directory | Dockerfile Location | Port | Domain(s) |
|-----|----------------|--------------------|------|-----------|
| Frontend | *(empty)* | `frontend/Dockerfile` | 80 | uniroute.co, www.uniroute.co |
| Backend | *(empty or .)* | `Dockerfile` | 8084 | app.uniroute.co |
| Tunnel | *(empty or .)* | `Dockerfile.tunnel` | 8080 + **20000-20100** (or 20000-30000) | \*.uniroute.co |

- **Frontend:** Leave Base Directory empty so the build context is the repo root (the Dockerfile uses `COPY frontend/...`). Do not set Base Directory to `frontend` or the build may fail with a path error.
- **Backend:** Root Dockerfile; set env vars (e.g. `DATABASE_URL`, `JWT_SECRET`, `API_KEY_SECRET`, `TUNNEL_BASE_DOMAIN`, `WEBSITE_URL`).
- **Tunnel:** Use `Dockerfile.tunnel`; set `TUNNEL_BASE_DOMAIN`, `WEBSITE_URL`, `DATABASE_URL`, `JWT_SECRET`, `API_KEY_SECRET`. For wildcard SSL use DNS challenge (e.g. Cloudflare) in Coolify/Traefik.
  - **TCP/TLS/UDP tunnels:** Publish TCP and UDP port ranges in addition to the main port (8080 or 8081). Recommended: **20000-20100** (101 ports) with `TUNNEL_TCP_PORT_RANGE=101`. In Coolify, add **Port** mappings: `20000-20100:20000-20100` (TCP) and `20000-20100:20000-20100/udp` (UDP). Docker publishes TCP only by default, so UDP tunnels need an explicit UDP mapping. Without TCP mapping, `uniroute tcp` public URLs fail with "Connection refused"; without UDP mapping, `uniroute udp` packets never reach the server. Open the range on the host firewall for both: `ufw allow 20000:20100/tcp` and `ufw allow 20000:20100/udp`.
  - **If Coolify’s Port Mappings don’t support UDP:** Coolify’s UI may only expose TCP when you enter a port range. To get UDP as well, deploy the tunnel using the **Docker Compose** build pack and the repo’s `docker-compose.tunnel.yml`, which already includes both TCP and UDP port entries. In the Compose deploy, the file is the source of truth, so both mappings will be applied.
  - **Liveness vs readiness:** `GET /live` always returns `200` with `{"status":"ok"}` when the process is up (no locks, no tunnel state)—use **`/live`** as the Coolify health check URL so the server is treated as up and running and is not restarted after long uptime. `GET /health` returns `200` with `{"status":"ok","tunnels":N,"connected":M}` when healthy, or **503** if there are tunnel entries but no active client (`tunnels` > 0 and `connected` == 0)—use `/health` only if you want Coolify to restart the container when tunnels are stale.

### WebSocket forwarding (Coolify + Caddy)

If the tunnel server sits behind **Caddy** (e.g. in Coolify with Caddy as the reverse proxy), the proxy must forward WebSocket upgrades so that browser WebSockets (e.g. Vite HMR, live reload) work through the tunnel. Otherwise the Go tunnel server receives a normal HTTP request instead of an upgrade and the WebSocket connection fails.

With **Caddy v2**, `reverse_proxy` supports WebSocket by default: it forwards the `Upgrade` and `Connection` headers. Ensure your tunnel server block does not override or strip those. Example for `*.tunnel.example.com`:

```caddy
*.tunnel.example.com {
	reverse_proxy tunnel_backend:8080
}
```

If you use a custom config that sets headers, avoid removing `Upgrade` or `Connection`. For long-lived HMR connections you can increase timeouts:

```caddy
*.tunnel.example.com {
	reverse_proxy tunnel_backend:8080 {
		transport http {
			read_timeout 86400s
			write_timeout 86400s
		}
	}
}
```

Replace `tunnel_backend:8080` with your tunnel service name and port (e.g. the Coolify-generated host/port for the tunnel app).

### Using the repo Caddyfile with Coolify

This repo includes Caddy configs in **`scripts/caddy/`**:

- **`Caddyfile.example`** – template with placeholders (`BASE_DOMAIN`, `TUNNEL_PORT`, etc.); substitute or use env vars.
- **`Caddyfile.server`** – example for a server where Caddy runs and proxies to services on the host (e.g. Coolify apps bound to host ports).

In `Caddyfile.server`, `tunnel.uniroute.co` and the `:443` catch-all point at **`host.docker.internal:TUNNEL_PORT`** (e.g. `8081`). For **tunnel.uniroute.co** to keep working after the tunnel container restarts (not only after a full redeploy), the tunnel app in Coolify must use a **fixed host port** (e.g. **8081**) so that Caddy always forwards to the same port. If Coolify assigns a random port per deploy, Caddy’s config would point at the wrong port after a restart until you redeploy (and regenerate the Caddyfile or update the port). In Coolify, set the tunnel application’s **Port** to a fixed value (e.g. **8081**) and keep that port in your Caddyfile.

If the tunnel works but shows **“no available server”** (or tunnel connection lost) after about a day, Caddy is likely closing the long-lived WebSocket to the tunnel server due to default transport timeouts. The repo’s **`scripts/caddy/Caddyfile.server`** sets `read_timeout 168h` and `write_timeout 168h` on the tunnel `reverse_proxy` so Caddy does not close the connection. Apply that Caddyfile (or add the same `transport http { read_timeout 168h; write_timeout 168h }` to your tunnel block) and reload Caddy.

**Keeping Caddy when Coolify overwrites the proxy:** Coolify may replace your Caddy proxy with Traefik when you revalidate the server, restart the proxy, or redeploy. **Coolify does not provide an option to disable proxy management or to prevent it from overwriting the proxy config** (see [Coolify docs](https://coolify.io/docs/knowledge-base/server/proxies) and [issue #3018](https://github.com/coollabsio/coolify/issues/3018)). So the only way to keep Caddy in place is to **restore it after Coolify overwrites**—either manually or automatically.

**Automatic restore (recommended):** Run from your machine once per server:

```bash
PROXY_SERVER=user@your-server ./scripts/caddy/install_persist_remote.sh
```

This uploads **`Caddyfile.server`** and the Caddy docker-compose to `/data/coolify/proxy/.uniroute-caddy/` on the server and installs a **cron job every 10 minutes**. If the proxy has been reverted to Traefik or the Caddyfile no longer has the tunnel timeouts, the script restores Caddy automatically (site may be unreachable for up to ~10 minutes after Coolify overwrites). Log on the server: `/data/coolify/proxy/.uniroute-caddy/persist.log`.

**Manual restore:** If you don’t use the cron, re-apply Caddy after each overwrite (e.g. SCP `Caddyfile.server` and `docker-compose.static.yml` to the server, replace the proxy compose and Caddyfile, then `docker rm -f coolify-proxy` and `docker compose up -d` in `/data/coolify/proxy`).

### Persisting tunnel state when the tunnel client runs in Coolify

If you run the **tunnel client** (CLI) inside Coolify (e.g. a service that runs `uniroute tunnel` to expose an app), redeploys clear the in-memory connection on the tunnel server. The client cannot “keep” a WebSocket across a restart, but it can **resume the same tunnel** after restart by loading saved state.

To persist tunnel state (subdomain, tunnel ID, auth) across Coolify redeploys/restarts:

1. **Add a persistent volume** in Coolify for the tunnel client app, e.g. mount path `/data`.
2. **Set environment variable:** `UNIROUTE_CONFIG_DIR=/data/.uniroute`
3. Ensure the tunnel client writes into that directory (auth, `tunnel-state.json`, `tunnels.json`). The CLI uses `UNIROUTE_CONFIG_DIR` when set.

After a redeploy, the new container will read the same `tunnel-state.json` and reconnect with the same subdomain, so the tunnel is back in the tunnel server’s memory without manual steps.

## Managed Service

Use UniRoute's managed service for:
- Automatic updates
- Managed infrastructure
- 24/7 monitoring
- Automatic scaling

Sign up at https://uniroute.co

## Next Steps

- [Getting Started](/docs/getting-started) - Start using UniRoute
- [Security](/docs/security) - Secure your deployment
