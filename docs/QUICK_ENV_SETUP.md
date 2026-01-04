# Quick Environment Setup

## 🚀 3-Step Setup

### Step 1: Copy Example File
```bash
cp .env.example .env
```

### Step 2: Edit `.env` (Optional)
```bash
# Open .env and update if needed
# Defaults work for Phase 1 (local development)
```

### Step 3: Run Server
```bash
make dev
```

**That's it!** The server automatically loads `.env` file.

## 📝 What's in `.env`?

### Phase 1 (Default - Works Out of the Box)
```bash
PORT=8084
ENV=development
OLLAMA_BASE_URL=http://localhost:11434
API_KEY_SECRET=change-me-in-production-min-32-chars
JWT_SECRET=change-me-in-production-jwt-secret-min-32-chars
```

### Phase 2 (Optional - Uncomment to Enable)
```bash
# DATABASE_URL=postgres://user:password@localhost/uniroute?sslmode=disable
# REDIS_URL=redis://localhost:6379
# IP_WHITELIST=192.168.1.100,10.0.0.1
```

## ✅ Verification

After starting the server, you should see:
```
Starting UniRoute Gateway...
Generated default API key (save this!): ur_abc123...
Server starting on :8084
```

## 🔒 Security Note

- `.env` is git-ignored (never committed)
- Use `.env.example` as template
- Change secrets in production!

## 📚 More Details

See `ENV_SETUP.md` for complete documentation.

