# Environment Configuration Setup

UniRoute uses environment variables for configuration. You can set them via:
1. `.env` file (recommended for local development)
2. System environment variables (for production)
3. Command line (temporary)

## Quick Start

### 1. Copy Example File

```bash
cp .env.example .env
```

### 2. Edit `.env` File

Open `.env` and update the values:

```bash
# Basic configuration
PORT=8084
ENV=development
OLLAMA_BASE_URL=http://localhost:11434

# Security (CHANGE THESE!)
API_KEY_SECRET=your-secret-key-here-min-32-chars
JWT_SECRET=your-jwt-secret-here-min-32-chars

# Phase 2 (Optional - uncomment to enable)
# DATABASE_URL=postgres://user:password@localhost/uniroute?sslmode=disable
# REDIS_URL=redis://localhost:6379
```

### 3. Run Server

```bash
make dev
# or
./bin/uniroute
```

The server will automatically load `.env` file!

## Environment Variables

### Required (Phase 1)

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8084` | Server port |
| `ENV` | `development` | Environment (development/production) |
| `OLLAMA_BASE_URL` | `http://localhost:11434` | Ollama server URL |
| `API_KEY_SECRET` | `change-me-in-production` | Secret for API key hashing |

### Optional (Phase 1)

| Variable | Default | Description |
|----------|---------|-------------|
| `JWT_SECRET` | `change-me-in-production-jwt-secret-min-32-chars` | JWT signing secret |

### Phase 2 (Optional)

| Variable | Default | Description |
|----------|---------|-------------|
| `DATABASE_URL` | (empty) | PostgreSQL connection string |
| `REDIS_URL` | (empty) | Redis connection string |
| `IP_WHITELIST` | (empty) | Comma-separated IP addresses |

**Note:** If `DATABASE_URL` or `REDIS_URL` are empty, UniRoute runs in Phase 1 mode (in-memory API keys, no rate limiting).

## Configuration Modes

### Phase 1 Mode (Simple)

```bash
# .env
PORT=8084
ENV=development
OLLAMA_BASE_URL=http://localhost:11434
API_KEY_SECRET=your-secret-here
```

**Features:**
- ✅ In-memory API keys
- ✅ Basic authentication
- ✅ Local LLM support
- ❌ No rate limiting
- ❌ No database persistence

### Phase 2 Mode (Full Features)

```bash
# .env
PORT=8084
ENV=development
OLLAMA_BASE_URL=http://localhost:11434
API_KEY_SECRET=your-secret-here
JWT_SECRET=your-jwt-secret-here-min-32-chars

# Phase 2 features
DATABASE_URL=postgres://user:password@localhost/uniroute?sslmode=disable
REDIS_URL=redis://localhost:6379
IP_WHITELIST=192.168.1.100,10.0.0.1
```

**Features:**
- ✅ Database-backed API keys
- ✅ Redis rate limiting
- ✅ JWT authentication
- ✅ IP whitelisting
- ✅ All Phase 1 features

## Production Setup

### ✅ Option 1: Environment Variables (RECOMMENDED)

**This is the recommended approach for production deployments.**

```bash
# Generate secure secrets
export API_KEY_SECRET=$(openssl rand -hex 32)
export JWT_SECRET=$(openssl rand -hex 32)

# Set required variables
export PORT=8084
export ENV=production
export OLLAMA_BASE_URL=http://ollama:11434
export DATABASE_URL=postgres://user:password@db.example.com/uniroute?sslmode=require
export REDIS_URL=redis://redis.example.com:6379

# Run server
./bin/uniroute
```

**Or use the helper script:**
```bash
# Set DATABASE_URL and REDIS_URL first, then:
source scripts/setup_env_production.sh
./bin/uniroute
```

**See `PRODUCTION_SETUP.md` for complete production deployment guide.**

### Option 2: .env File (Not Recommended for Production)

```bash
# .env (DO NOT commit to git!)
PORT=8084
ENV=production
API_KEY_SECRET=your-production-secret
# ... etc
```

**Security Note:** Never commit `.env` files to git! They're already in `.gitignore`.

## Docker/Container Setup

For Docker, use environment variables:

```yaml
# docker-compose.yml
services:
  uniroute:
    environment:
      - PORT=8084
      - ENV=production
      - DATABASE_URL=postgres://...
      - REDIS_URL=redis://...
    env_file:
      - .env.production
```

## Development Setup

### Local Development

1. Copy `.env.example` to `.env`
2. Update values for your local setup
3. Run `make dev`

### Testing

For integration tests, you can use:

```bash
# .env.local (git-ignored)
TEST_DATABASE_URL=postgres://postgres:postgres@localhost/uniroute_test?sslmode=disable
TEST_REDIS_URL=redis://localhost:6379/15
```

## Environment Priority

Configuration is loaded in this order (later overrides earlier):

1. `.env` file
2. `.env.local` file (if exists)
3. System environment variables

## Security Best Practices

1. **Never commit `.env` files**
   - Already in `.gitignore`
   - Use `.env.example` as template

2. **Use strong secrets**
   ```bash
   # Generate secure secrets
   openssl rand -hex 32  # For API_KEY_SECRET
   openssl rand -hex 32  # For JWT_SECRET
   ```

3. **Different secrets per environment**
   - Development: `.env`
   - Staging: `.env.staging`
   - Production: Environment variables only

4. **Rotate secrets regularly**
   - Especially in production
   - Update all API keys after rotation

## Troubleshooting

### .env file not loading?

1. Check file exists: `ls -la .env`
2. Check file location (should be in project root)
3. Check file permissions: `chmod 600 .env`
4. Verify format (no spaces around `=`)

### Wrong configuration values?

1. Check for `.env.local` override
2. Check system environment variables: `env | grep UNIROUTE`
3. Restart server after changing `.env`

### Database connection issues?

1. Verify `DATABASE_URL` format
2. Test connection: `psql $DATABASE_URL`
3. Check PostgreSQL is running

### Redis connection issues?

1. Verify `REDIS_URL` format
2. Test connection: `redis-cli -u $REDIS_URL ping`
3. Check Redis is running

## Example Configurations

### Minimal (Phase 1)

```bash
PORT=8084
OLLAMA_BASE_URL=http://localhost:11434
API_KEY_SECRET=dev-secret-key-12345
```

### Full (Phase 2)

```bash
PORT=8084
ENV=development
OLLAMA_BASE_URL=http://localhost:11434
API_KEY_SECRET=dev-secret-key-12345
JWT_SECRET=dev-jwt-secret-key-12345
DATABASE_URL=postgres://postgres:postgres@localhost/uniroute?sslmode=disable
REDIS_URL=redis://localhost:6379
IP_WHITELIST=127.0.0.1,::1
```

### Production

```bash
PORT=8084
ENV=production
OLLAMA_BASE_URL=http://ollama:11434
API_KEY_SECRET=<generated-secure-secret>
JWT_SECRET=<generated-secure-secret>
DATABASE_URL=postgres://user:pass@db:5432/uniroute?sslmode=require
REDIS_URL=redis://redis:6379
```

## Next Steps

1. Copy `.env.example` to `.env`
2. Update values for your environment
3. Run `make dev` to start server
4. Check server logs for configuration confirmation

