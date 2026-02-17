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
| Tunnel | *(empty or .)* | `Dockerfile.tunnel` | 8080 | \*.uniroute.co |

- **Frontend:** Leave Base Directory empty so the build context is the repo root (the Dockerfile uses `COPY frontend/...`). Do not set Base Directory to `frontend` or the build may fail with a path error.
- **Backend:** Root Dockerfile; set env vars (e.g. `DATABASE_URL`, `JWT_SECRET`, `API_KEY_SECRET`, `TUNNEL_BASE_DOMAIN`, `WEBSITE_URL`).
- **Tunnel:** Use `Dockerfile.tunnel`; set `TUNNEL_BASE_DOMAIN`, `WEBSITE_URL`, `DATABASE_URL`, `JWT_SECRET`, `API_KEY_SECRET`. For wildcard SSL use DNS challenge (e.g. Cloudflare) in Coolify/Traefik.

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
