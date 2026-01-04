# Deploying UniRoute with Coolify

[Coolify](https://coolify.io) is a self-hosted alternative to Heroku, Vercel, and Netlify. This guide shows you how to deploy UniRoute on Coolify.

## 🚀 Quick Deployment

### Prerequisites

1. **Coolify Instance** - Self-hosted or use [Coolify Cloud](https://coolify.io)
2. **PostgreSQL Database** - Can be provisioned via Coolify
3. **Redis Instance** - Can be provisioned via Coolify

## 📋 Step-by-Step Deployment

### Step 1: Create New Resource in Coolify

1. Log into your Coolify instance
2. Click **"New Resource"**
3. Select **"Docker Compose"** or **"Dockerfile"**

### Step 2: Connect Your Repository

1. **Option A: GitHub/GitLab**
   - Connect your repository: `https://github.com/Kizsoft-Solution-Limited/uniroute`
   - Select branch: `main` (or your preferred branch)

2. **Option B: Dockerfile**
   - Coolify will detect the Dockerfile automatically
   - Or use the provided Dockerfile

### Step 3: Configure Environment Variables

In Coolify's environment variables section, add:

#### Required Variables

```bash
PORT=8084
ENV=production
OLLAMA_BASE_URL=http://ollama:11434
API_KEY_SECRET=<generate-with-openssl-rand-hex-32>
JWT_SECRET=<generate-with-openssl-rand-hex-32>
```

#### Phase 2 Variables (Recommended)

```bash
DATABASE_URL=postgres://user:password@postgres:5432/uniroute?sslmode=disable
REDIS_URL=redis://redis:6379
```

**Note:** If using Coolify's database/Redis services, use the service names as hosts (e.g., `postgres`, `redis`).

### Step 4: Set Up Database (Phase 2)

1. In Coolify, create a **PostgreSQL** service
2. Note the connection details
3. Update `DATABASE_URL` in environment variables:
   ```
   DATABASE_URL=postgres://postgres:password@postgres:5432/uniroute?sslmode=disable
   ```

### Step 5: Set Up Redis (Phase 2)

1. In Coolify, create a **Redis** service
2. Note the connection details
3. Update `REDIS_URL` in environment variables:
   ```
   REDIS_URL=redis://redis:6379
   ```

### Step 6: Configure Port

1. Set **Port** to `8084` (or your preferred port)
2. Coolify will handle reverse proxy automatically

### Step 7: Deploy

1. Click **"Deploy"**
2. Wait for build to complete
3. Check logs for any issues

## 🐳 Dockerfile for Coolify

Create a `Dockerfile` in the project root:

```dockerfile
# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build binary
RUN CGO_ENABLED=0 GOOS=linux go build -o bin/uniroute ./cmd/gateway

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Copy binary from builder
COPY --from=builder /app/bin/uniroute .

# Expose port
EXPOSE 8084

# Run the binary
CMD ["./uniroute"]
```

## 📝 Coolify Configuration File

Create `coolify.yml` in project root (optional):

```yaml
version: '3.8'

services:
  uniroute:
    build:
      context: .
      dockerfile: Dockerfile
    ports:
      - "8084:8084"
    environment:
      - PORT=8084
      - ENV=production
      - OLLAMA_BASE_URL=${OLLAMA_BASE_URL}
      - API_KEY_SECRET=${API_KEY_SECRET}
      - JWT_SECRET=${JWT_SECRET}
      - DATABASE_URL=${DATABASE_URL}
      - REDIS_URL=${REDIS_URL}
    depends_on:
      - postgres
      - redis
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "wget", "--quiet", "--tries=1", "--spider", "http://localhost:8084/health"]
      interval: 30s
      timeout: 10s
      retries: 3

  postgres:
    image: postgres:15-alpine
    environment:
      - POSTGRES_DB=uniroute
      - POSTGRES_USER=postgres
      - POSTGRES_PASSWORD=${POSTGRES_PASSWORD}
    volumes:
      - postgres_data:/var/lib/postgresql/data
    restart: unless-stopped

  redis:
    image: redis:7-alpine
    restart: unless-stopped
    volumes:
      - redis_data:/data

volumes:
  postgres_data:
  redis_data:
```

## 🔐 Generating Secrets in Coolify

### Method 1: Using Coolify's Secret Generator

1. In Coolify, go to **Environment Variables**
2. Click **"Generate Secret"** for:
   - `API_KEY_SECRET`
   - `JWT_SECRET`

### Method 2: Using Terminal

```bash
# Generate secrets
openssl rand -hex 32  # For API_KEY_SECRET
openssl rand -hex 32  # For JWT_SECRET
```

Then paste into Coolify's environment variables.

## 🌐 Domain Configuration

1. In Coolify, go to **Domains**
2. Add your domain (e.g., `uniroute.example.com`)
3. Coolify will automatically:
   - Set up SSL certificate (Let's Encrypt)
   - Configure reverse proxy
   - Handle HTTPS

## 🔄 Database Migrations

After first deployment, run migrations:

### Option 1: Via Coolify Terminal

1. Open terminal in Coolify for your UniRoute service
2. Run:
   ```bash
   psql $DATABASE_URL < migrations/001_initial_schema.sql
   ```

### Option 2: Via External Tool

1. Connect to your PostgreSQL instance
2. Run the migration file

## 📊 Health Checks

Coolify will automatically check:
- `http://localhost:8084/health`

Make sure this endpoint returns:
```json
{"status": "ok"}
```

## 🔧 Environment Variables Reference

### Required

| Variable | Example | Description |
|----------|---------|-------------|
| `PORT` | `8084` | Server port |
| `ENV` | `production` | Environment |
| `OLLAMA_BASE_URL` | `http://ollama:11434` | Ollama URL |
| `API_KEY_SECRET` | `(64-char hex)` | API key secret |
| `JWT_SECRET` | `(64-char hex)` | JWT secret |

### Phase 2 (Optional)

| Variable | Example | Description |
|----------|---------|-------------|
| `DATABASE_URL` | `postgres://...` | PostgreSQL URL |
| `REDIS_URL` | `redis://redis:6379` | Redis URL |
| `IP_WHITELIST` | `192.168.1.100` | Allowed IPs |

## 🚨 Troubleshooting

### Build Fails

1. **Check Dockerfile exists**
   - Ensure `Dockerfile` is in project root
   - Or use Coolify's Dockerfile detection

2. **Check Go version**
   - Ensure Dockerfile uses Go 1.24+
   - Update `FROM golang:1.24-alpine`

3. **Check dependencies**
   - Verify `go.mod` is correct
   - Check build logs for errors

### Database Connection Issues

1. **Check service names**
   - Use service names (e.g., `postgres`, `redis`) not `localhost`
   - Coolify services are accessible by name

2. **Check DATABASE_URL format**
   ```
   postgres://user:password@postgres:5432/uniroute?sslmode=disable
   ```

3. **Verify database is running**
   - Check Coolify dashboard
   - Verify PostgreSQL service is healthy

### Redis Connection Issues

1. **Check REDIS_URL format**
   ```
   redis://redis:6379
   ```

2. **Verify Redis is running**
   - Check Coolify dashboard
   - Verify Redis service is healthy

### Port Conflicts

1. **Check port configuration**
   - Ensure `PORT=8084` in environment variables
   - Verify no other service uses port 8084

2. **Check Coolify port mapping**
   - Verify port is correctly mapped
   - Check reverse proxy configuration

## 📈 Scaling

### Horizontal Scaling

1. In Coolify, go to **Scaling**
2. Set number of replicas
3. Coolify will handle load balancing

### Resource Limits

Set in Coolify:
- **CPU**: 0.5-2 cores
- **Memory**: 512MB-2GB
- **Storage**: As needed

## 🔄 Updates

### Automatic Updates

1. Enable **Auto Deploy** in Coolify
2. Coolify will redeploy on git push

### Manual Updates

1. Click **"Redeploy"** in Coolify
2. Or trigger via webhook

## 📚 Additional Resources

- [Coolify Documentation](https://coolify.io/docs)
- [Coolify GitHub](https://github.com/coollabsio/coolify)
- [UniRoute Documentation](./README.md)

## 🎯 Quick Checklist

- [ ] Coolify instance set up
- [ ] Repository connected
- [ ] Environment variables configured
- [ ] PostgreSQL service created (Phase 2)
- [ ] Redis service created (Phase 2)
- [ ] Database migrations run
- [ ] Domain configured (optional)
- [ ] SSL certificate issued (automatic)
- [ ] Health check passing
- [ ] First API key generated

## 💡 Pro Tips

1. **Use Coolify's Database Service**
   - Easier to manage
   - Automatic backups
   - Built-in monitoring

2. **Enable Auto Deploy**
   - Automatic updates on git push
   - Saves time on deployments

3. **Set Up Monitoring**
   - Use Coolify's built-in monitoring
   - Set up alerts for downtime

4. **Use Environment Presets**
   - Create presets for dev/staging/prod
   - Easy switching between environments

5. **Backup Strategy**
   - Regular database backups
   - Export environment variables
   - Document configuration

---

**Ready to deploy?** Follow the steps above and you'll have UniRoute running on Coolify in minutes! 🚀

