# Production Setup Guide

This guide shows you how to set up UniRoute for production using **Environment Variables** (Option 1 - Recommended).

## 🚀 Quick Production Setup

### Step 1: Generate Secure Secrets

```bash
# Generate secrets
export API_KEY_SECRET=$(openssl rand -hex 32)
export JWT_SECRET=$(openssl rand -hex 32)
```

### Step 2: Set Required Variables

```bash
export PORT=8084
export ENV=production
export OLLAMA_BASE_URL=http://ollama:11434
export DATABASE_URL=postgres://user:password@db.example.com/uniroute?sslmode=require
export REDIS_URL=redis://redis.example.com:6379
```

### Step 3: Run Server

```bash
./bin/uniroute
```

## 📋 Complete Production Configuration

### Required Variables

```bash
# Server
export PORT=8084
export ENV=production

# Ollama
export OLLAMA_BASE_URL=http://ollama:11434

# Security (generate with openssl rand -hex 32)
export API_KEY_SECRET=$(openssl rand -hex 32)
export JWT_SECRET=$(openssl rand -hex 32)
```

### Phase 2 Variables (Recommended)

```bash
# Database
export DATABASE_URL=postgres://user:password@db.example.com/uniroute?sslmode=require

# Redis
export REDIS_URL=redis://redis.example.com:6379

# Optional: IP Whitelist
export IP_WHITELIST=192.168.1.100,10.0.0.1
```

## 🔧 Using the Setup Script

We provide a helper script to automate this:

```bash
# Basic setup (generates secrets)
source scripts/setup_env.sh

# Production setup (with validation)
source scripts/setup_env_production.sh
```

## 🐳 Docker/Container Setup

### Coolify Deployment

**Coolify is a self-hosted alternative to Heroku/Vercel. See `COOLIFY_DEPLOYMENT.md` for complete guide.**

Quick setup:
1. Connect your repository in Coolify
2. Set environment variables
3. Deploy!

### Docker Compose

```yaml
version: '3.8'

services:
  uniroute:
    image: uniroute:latest
    ports:
      - "8084:8084"
    environment:
      - PORT=8084
      - ENV=production
      - OLLAMA_BASE_URL=http://ollama:11434
      - API_KEY_SECRET=${API_KEY_SECRET}
      - JWT_SECRET=${JWT_SECRET}
      - DATABASE_URL=${DATABASE_URL}
      - REDIS_URL=${REDIS_URL}
    restart: unless-stopped
```

### Kubernetes

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: uniroute-secrets
type: Opaque
stringData:
  API_KEY_SECRET: <generated-secret>
  JWT_SECRET: <generated-secret>
  DATABASE_URL: postgres://...
  REDIS_URL: redis://...
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: uniroute
spec:
  template:
    spec:
      containers:
      - name: uniroute
        image: uniroute:latest
        env:
        - name: PORT
          value: "8084"
        - name: ENV
          value: "production"
        - name: API_KEY_SECRET
          valueFrom:
            secretKeyRef:
              name: uniroute-secrets
              key: API_KEY_SECRET
        # ... other env vars
```

## ☁️ Cloud Platform Setup

### AWS (ECS/Fargate)

```bash
# Set secrets in AWS Secrets Manager
aws secretsmanager create-secret \
  --name uniroute/production \
  --secret-string '{"API_KEY_SECRET":"...","JWT_SECRET":"..."}'

# Reference in task definition
{
  "environment": [
    {"name": "PORT", "value": "8084"},
    {"name": "ENV", "value": "production"}
  ],
  "secrets": [
    {"name": "API_KEY_SECRET", "valueFrom": "arn:aws:secretsmanager:..."},
    {"name": "JWT_SECRET", "valueFrom": "arn:aws:secretsmanager:..."}
  ]
}
```

### Google Cloud (Cloud Run)

```bash
# Deploy with environment variables
gcloud run deploy uniroute \
  --set-env-vars PORT=8084,ENV=production \
  --set-secrets API_KEY_SECRET=api-key-secret:latest,JWT_SECRET=jwt-secret:latest
```

### Azure (Container Instances)

```bash
az container create \
  --name uniroute \
  --environment-variables PORT=8084 ENV=production \
  --secrets API_KEY_SECRET=<secret> JWT_SECRET=<secret>
```

## 🔒 Security Best Practices

### 1. Secret Management

**✅ DO:**
- Use environment variables for secrets
- Store secrets in secret management services (AWS Secrets Manager, HashiCorp Vault, etc.)
- Rotate secrets regularly
- Use different secrets per environment

**❌ DON'T:**
- Commit secrets to git
- Hardcode secrets in code
- Share secrets via email/chat
- Use same secrets in dev and production

### 2. Secret Generation

```bash
# Generate secure 32-byte secrets
openssl rand -hex 32  # For API_KEY_SECRET
openssl rand -hex 32  # For JWT_SECRET

# Or use the setup script
source scripts/setup_env_production.sh
```

### 3. Secret Storage

**Option A: Environment Variables (Recommended)**
```bash
export API_KEY_SECRET=$(openssl rand -hex 32)
export JWT_SECRET=$(openssl rand -hex 32)
```

**Option B: Secret Management Service**
- AWS Secrets Manager
- HashiCorp Vault
- Google Secret Manager
- Azure Key Vault

**Option C: Process Manager**
- systemd (EnvironmentFile)
- Docker secrets
- Kubernetes secrets

## 📝 Environment Variable Reference

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `PORT` | No | `8084` | Server port |
| `ENV` | No | `development` | Environment (development/production) |
| `OLLAMA_BASE_URL` | No | `http://localhost:11434` | Ollama server URL |
| `API_KEY_SECRET` | **Yes** | - | Secret for API key hashing (32+ chars) |
| `JWT_SECRET` | No* | - | JWT signing secret (32+ chars) |
| `DATABASE_URL` | No* | - | PostgreSQL connection string |
| `REDIS_URL` | No* | - | Redis connection string |
| `IP_WHITELIST` | No | - | Comma-separated IP addresses |

*Required for Phase 2 features

## 🔄 Secret Rotation

### Rotating Secrets

1. **Generate new secrets:**
   ```bash
   export NEW_API_KEY_SECRET=$(openssl rand -hex 32)
   export NEW_JWT_SECRET=$(openssl rand -hex 32)
   ```

2. **Update environment:**
   ```bash
   export API_KEY_SECRET=$NEW_API_KEY_SECRET
   export JWT_SECRET=$NEW_JWT_SECRET
   ```

3. **Restart service:**
   ```bash
   systemctl restart uniroute
   # or
   docker-compose restart uniroute
   ```

4. **Note:** Existing API keys will need to be recreated after rotation.

## 🧪 Testing Production Setup

### Verify Configuration

```bash
# Check all variables are set
env | grep -E "PORT|ENV|API_KEY_SECRET|JWT_SECRET|DATABASE_URL|REDIS_URL"

# Test database connection
psql $DATABASE_URL -c "SELECT 1"

# Test Redis connection
redis-cli -u $REDIS_URL ping
```

### Health Check

```bash
# After starting server
curl http://localhost:8084/health
# Should return: {"status":"ok"}
```

## 📚 Additional Resources

- **Development Setup:** See `ENV_SETUP.md`
- **Quick Start:** See `QUICK_ENV_SETUP.md`
- **Docker Setup:** See `docker-compose.yml` (if available)
- **Kubernetes:** See `k8s/` directory (if available)

## 🆘 Troubleshooting

### Secrets Not Working

1. Verify secrets are exported: `echo $API_KEY_SECRET`
2. Check secret length (should be 64 hex chars for 32 bytes)
3. Restart service after setting variables

### Database Connection Issues

1. Verify `DATABASE_URL` format
2. Test connection: `psql $DATABASE_URL`
3. Check firewall/network rules
4. Verify SSL mode matches database config

### Redis Connection Issues

1. Verify `REDIS_URL` format
2. Test connection: `redis-cli -u $REDIS_URL ping`
3. Check firewall/network rules
4. Verify Redis is accessible from server

---

**Remember:** Environment variables are the recommended approach for production. Never use `.env` files in production!

