# UniRoute CLI: `keys` Command Usage

## Overview

The `uniroute keys` command allows you to manage API keys for the UniRoute gateway. API keys are used to authenticate requests to the UniRoute API.

## Location

**Implementation**: `cmd/cli/commands/keys.go`  
**Backend Handler**: `internal/api/handlers/apikeys.go`  
**Service**: `internal/security/apikey_v2.go`  
**Repository**: `internal/storage/apikey_repository.go`

## Available Commands

### 1. `uniroute keys create` - Create a new API key

Creates a new API key for accessing the UniRoute gateway.

**Usage:**
```bash
uniroute keys create
uniroute keys create --name "My API Key"
uniroute keys create --url http://localhost:8084 --name "My API Key"
uniroute keys create --jwt-token YOUR_JWT_TOKEN
```

**Flags:**
- `--name, -n`: Name for the API key (optional)
- `--url, -u`: Gateway server URL (default: `https://api.uniroute.dev`)
- `--jwt-token, -t`: JWT token for authentication (if not logged in)

**Example Output:**
```
✅ API Key created successfully!

API Key: ur_abc123def456...

⚠️  Save this key securely - it won't be shown again!
ID: 123e4567-e89b-12d3-a456-426614174000
```

**Backend Endpoint:**
- `POST /admin/api-keys` (requires JWT authentication)

**How it works:**
1. Authenticates using JWT token (from `uniroute auth login` or `--jwt-token` flag)
2. Sends POST request to `/admin/api-keys` with optional name
3. Backend generates a secure API key with prefix `ur_`
4. Key is hashed and stored in PostgreSQL
5. Returns the raw key (only shown once!)

---

### 2. `uniroute keys list` - List all API keys

Lists all API keys for your account.

**Usage:**
```bash
uniroute keys list
uniroute keys list --url http://localhost:8084
uniroute keys list --jwt-token YOUR_JWT_TOKEN
```

**Flags:**
- `--url, -u`: Gateway server URL (default: `https://api.uniroute.dev`)
- `--jwt-token, -t`: JWT token for authentication (if not logged in)

**Example Output:**
```
Your API Keys:
--------------------------------------------------------------------------------

1. Production Key
   ID: 123e4567-e89b-12d3-a456-426614174000
   Created: 2024-01-15T10:30:00Z
   Expires: 2024-12-31T23:59:59Z
   Status: Active
   Rate Limit: 60/min, 10000/day

2. Development Key
   ID: 987e6543-e21b-43d2-b654-321987654321
   Created: 2024-01-10T08:15:00Z
   Status: Active
   Rate Limit: 30/min, 1000/day
```

**Backend Endpoint:**
- `GET /admin/api-keys` (requires JWT authentication)

**How it works:**
1. Authenticates using JWT token
2. Sends GET request to `/admin/api-keys`
3. Backend queries database for all API keys belonging to the authenticated user
4. Returns list of keys (without exposing the actual key values for security)

---

### 3. `uniroute keys revoke` - Revoke an API key

Revokes (deletes) an API key by ID.

**Usage:**
```bash
uniroute keys revoke <key-id>
uniroute keys revoke 123e4567-e89b-12d3-a456-426614174000
uniroute keys revoke <key-id> --url http://localhost:8084
```

**Arguments:**
- `<key-id>`: The UUID of the API key to revoke (required)

**Flags:**
- `--url, -u`: Gateway server URL (default: `https://api.uniroute.dev`)
- `--jwt-token, -t`: JWT token for authentication (if not logged in)

**Example Output:**
```
✅ API key 123e4567-e89b-12d3-a456-426614174000 revoked successfully
```

**Backend Endpoint:**
- `DELETE /admin/api-keys/:id` (requires JWT authentication)

**How it works:**
1. Authenticates using JWT token
2. Verifies the API key belongs to the authenticated user
3. Sends DELETE request to `/admin/api-keys/:id`
4. Backend soft-deletes the key (sets `is_active = false`)
5. Key can no longer be used for authentication

---

## Authentication

All `keys` commands require authentication:

1. **Login first** (recommended):
   ```bash
   uniroute auth login
   uniroute keys create
   ```

2. **Or use JWT token directly**:
   ```bash
   uniroute keys create --jwt-token YOUR_JWT_TOKEN
   ```

3. **For self-hosted servers**:
   ```bash
   uniroute keys create --url http://localhost:8084 --jwt-token YOUR_JWT_TOKEN
   ```

---

## Backend Implementation Details

### API Key Storage

- **Format**: `ur_` + 64 hex characters (e.g., `ur_abc123...`)
- **Lookup Hash**: SHA256 hash for fast database queries
- **Verification Hash**: bcrypt hash for secure verification
- **Database**: PostgreSQL `api_keys` table

### Security Features

- ✅ Keys are hashed with bcrypt before storage
- ✅ Only the key owner can list/revoke their keys
- ✅ Keys are soft-deleted (can be restored if needed)
- ✅ Rate limiting per key (configurable)
- ✅ Expiration dates supported

### Database Schema

```sql
CREATE TABLE api_keys (
    id UUID PRIMARY KEY,
    user_id UUID REFERENCES users(id),
    lookup_hash VARCHAR(255) UNIQUE NOT NULL,  -- SHA256 for fast lookup
    verification_hash TEXT NOT NULL,            -- bcrypt for verification
    name VARCHAR(255),
    rate_limit_per_minute INTEGER DEFAULT 60,
    rate_limit_per_day INTEGER DEFAULT 10000,
    created_at TIMESTAMP DEFAULT NOW(),
    expires_at TIMESTAMP,
    is_active BOOLEAN DEFAULT true
);
```

---

## Usage Examples

### Create and use an API key

```bash
# 1. Login
uniroute auth login

# 2. Create API key
uniroute keys create --name "My App Key"

# 3. Use the key in your application
curl -X POST https://api.uniroute.dev/v1/chat \
  -H "Authorization: Bearer ur_abc123..." \
  -H "Content-Type: application/json" \
  -d '{"model": "gpt-4", "messages": [...]}'
```

### List all keys

```bash
uniroute keys list
```

### Revoke a key

```bash
# First, list keys to get the ID
uniroute keys list

# Then revoke by ID
uniroute keys revoke 123e4567-e89b-12d3-a456-426614174000
```

### Self-hosted usage

```bash
# Create key on self-hosted server
uniroute keys create \
  --url http://localhost:8084 \
  --name "Local Dev Key" \
  --jwt-token YOUR_JWT_TOKEN
```

---

## Troubleshooting

### "not authenticated" error

**Solution**: Login first or provide JWT token
```bash
uniroute auth login
# OR
uniroute keys create --jwt-token YOUR_JWT_TOKEN
```

### "API key not found" when revoking

**Solution**: Make sure you're using the correct key ID and it belongs to your account
```bash
# List keys first to see all IDs
uniroute keys list
```

### Connection errors

**Solution**: Check server URL and network connectivity
```bash
# Use --url flag for custom servers
uniroute keys create --url http://your-server:8084
```

---

## Related Commands

- `uniroute auth login` - Authenticate with UniRoute
- `uniroute auth status` - Check authentication status
- `uniroute status` - Check gateway health and provider status

---

## See Also

- [CLI Installation Guide](./CLI_INSTALLATION.md)
- [Self-Hosted Usage](./CLI_SELF_HOSTED_USAGE.md)
- [API Keys Explained](./API_KEYS_EXPLAINED.md)

