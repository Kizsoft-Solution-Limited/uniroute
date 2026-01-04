# BYOK (Bring Your Own Keys) Implementation Status

## ❌ Current Status: **NOT FULLY IMPLEMENTED**

BYOK (Bring Your Own Keys) is **partially implemented** but needs additional work for full per-user support.

---

## ✅ What's Currently Implemented

### Backend (Server-Level)

**Current Implementation:**
- ✅ Provider API keys configured via **environment variables**
  - `OPENAI_API_KEY`
  - `ANTHROPIC_API_KEY`
  - `GOOGLE_API_KEY`
- ✅ Providers auto-register if keys are present
- ✅ All users share the same provider keys (server-level)

**Location:**
- `internal/config/config.go` - Reads from environment
- `cmd/gateway/main.go` - Registers providers with server-level keys

**How it works now:**
```bash
# Server administrator sets keys in .env
OPENAI_API_KEY=sk-...
ANTHROPIC_API_KEY=sk-ant-...
GOOGLE_API_KEY=AIza...

# All users of this server share these keys
# Requests use the same provider keys for everyone
```

---

## ❌ What's Missing for True BYOK

### Backend (Per-User)

**Missing Implementation:**

1. **Database Schema** - No table for user provider keys
   ```sql
   -- NEEDED: user_provider_keys table
   CREATE TABLE user_provider_keys (
       id UUID PRIMARY KEY,
       user_id UUID REFERENCES users(id),
       provider VARCHAR(50),  -- 'openai', 'anthropic', 'google'
       api_key_encrypted TEXT,  -- Encrypted provider API key
       is_active BOOLEAN DEFAULT true,
       created_at TIMESTAMP,
       updated_at TIMESTAMP
   );
   ```

2. **API Endpoints** - No endpoints to manage provider keys
   ```bash
   # NEEDED Endpoints:
   POST   /admin/provider-keys          # Add provider key
   GET    /admin/provider-keys          # List user's provider keys
   PUT    /admin/provider-keys/:id      # Update provider key
   DELETE /admin/provider-keys/:id      # Remove provider key
   ```

3. **Service Layer** - No service to manage provider keys
   - Need: `ProviderKeyService` to encrypt/decrypt keys
   - Need: Repository to store/retrieve keys
   - Need: Integration with router to use user's keys

4. **Router Integration** - Router doesn't use user-specific keys
   - Currently: Router uses server-level keys from config
   - Needed: Router should use user's keys from database when available

### Frontend

**Missing Implementation:**
- ❌ No UI to add/manage provider keys
- ❌ No settings page for provider configuration
- ❌ No form to enter OpenAI, Anthropic, Google keys
- ❌ No encryption/security indicators

---

## 🎯 What Needs to Be Implemented

### Phase 1: Backend - Database & Storage

**1. Database Migration**
```sql
-- migrations/004_user_provider_keys.sql
CREATE TABLE IF NOT EXISTS user_provider_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    provider VARCHAR(50) NOT NULL,  -- 'openai', 'anthropic', 'google'
    api_key_encrypted TEXT NOT NULL,  -- Encrypted with user-specific key
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(user_id, provider)  -- One key per provider per user
);

CREATE INDEX IF NOT EXISTS idx_user_provider_keys_user_id ON user_provider_keys(user_id);
CREATE INDEX IF NOT EXISTS idx_user_provider_keys_provider ON user_provider_keys(provider);
```

**2. Storage Layer**
- `internal/storage/provider_key_repository.go` - CRUD operations
- `internal/storage/models.go` - Add `UserProviderKey` model

**3. Service Layer**
- `internal/security/provider_key_service.go` - Encryption/decryption
- Methods:
  - `AddProviderKey(userID, provider, apiKey)`
  - `GetProviderKey(userID, provider)`
  - `ListProviderKeys(userID)`
  - `UpdateProviderKey(userID, provider, apiKey)`
  - `DeleteProviderKey(userID, provider)`

**4. Encryption**
- Encrypt provider keys at rest
- Use user-specific encryption key (derived from user password or master key)
- Never store plaintext provider keys

### Phase 2: Backend - API & Integration

**1. API Handlers**
- `internal/api/handlers/provider_keys.go`
- Endpoints:
  - `POST /admin/provider-keys` - Add provider key
  - `GET /admin/provider-keys` - List user's keys
  - `PUT /admin/provider-keys/:provider` - Update key
  - `DELETE /admin/provider-keys/:provider` - Remove key

**2. Router Integration**
- Modify `internal/gateway/router.go` to:
  - Accept `userID` in routing context
  - Look up user's provider keys
  - Use user's keys instead of server-level keys
  - Fall back to server-level keys if user has none

**3. Provider Factory**
- Create providers dynamically using user's keys
- Cache providers per user
- Handle key rotation/updates

### Phase 3: Frontend

**1. Settings Page**
- `views/settings/ProviderKeys.vue`
- Form to add/update provider keys
- List of configured providers
- Security indicators (encrypted, last updated)

**2. Provider Key Management**
- Add key form (with encryption indicator)
- Update key form
- Delete key confirmation
- Test connection button

**3. Security UI**
- Show encryption status
- Mask keys in UI (show only last 4 chars)
- Warning about key security

---

## 📋 Implementation Checklist

### Backend

- [ ] **Database Migration**
  - [ ] Create `user_provider_keys` table
  - [ ] Add indexes
  - [ ] Add foreign key constraints

- [ ] **Storage Layer**
  - [ ] `ProviderKeyRepository` interface
  - [ ] `PostgresProviderKeyRepository` implementation
  - [ ] `UserProviderKey` model

- [ ] **Service Layer**
  - [ ] `ProviderKeyService` with encryption
  - [ ] Key encryption/decryption methods
  - [ ] Key validation methods

- [ ] **API Endpoints**
  - [ ] `POST /admin/provider-keys`
  - [ ] `GET /admin/provider-keys`
  - [ ] `PUT /admin/provider-keys/:provider`
  - [ ] `DELETE /admin/provider-keys/:provider`
  - [ ] `POST /admin/provider-keys/:provider/test` - Test connection

- [ ] **Router Integration**
  - [ ] Modify router to accept user context
  - [ ] Lookup user's provider keys
  - [ ] Create providers with user's keys
  - [ ] Fallback to server-level keys

- [ ] **Security**
  - [ ] Encrypt keys at rest
  - [ ] Never log keys
  - [ ] Secure key rotation

### Frontend

- [ ] **Settings Page**
  - [ ] Provider keys management UI
  - [ ] Add/Edit/Delete forms
  - [ ] Security indicators

- [ ] **API Integration**
  - [ ] `useProviderKeys` composable
  - [ ] API service methods
  - [ ] Error handling

- [ ] **Security UI**
  - [ ] Key masking
  - [ ] Encryption status
  - [ ] Warnings

---

## 🔄 Current vs. Future Architecture

### Current (Server-Level Keys)

```
User Request
  ↓
UniRoute Gateway
  ↓
Router (uses server-level keys from config)
  ↓
Provider (OpenAI, Anthropic, Google)
```

**Limitation:** All users share the same provider keys.

### Future (Per-User Keys - BYOK)

```
User Request (with user_id from API key)
  ↓
UniRoute Gateway
  ↓
Router (looks up user's provider keys)
  ↓
Provider (uses user's own keys)
  ↓
User billed directly by provider
```

**Benefit:** Each user uses their own provider keys, billed directly.

---

## 🚀 Implementation Priority

### High Priority (Core BYOK)
1. Database schema for user provider keys
2. Encryption service for keys
3. API endpoints to manage keys
4. Router integration to use user keys

### Medium Priority (UX)
5. Frontend UI for key management
6. Key validation/testing
7. Security indicators

### Low Priority (Advanced)
8. Key rotation
9. Key sharing (teams)
10. Key usage analytics

---

## 📝 Summary

**Current State:**
- ✅ Server-level provider keys work
- ❌ Per-user provider keys NOT implemented
- ❌ No API endpoints for managing provider keys
- ❌ No frontend UI for provider keys

**What's Needed:**
1. Database table for user provider keys
2. Encryption service
3. API endpoints
4. Router integration
5. Frontend UI

**Estimated Work:**
- Backend: 2-3 days
- Frontend: 1-2 days
- Testing: 1 day
- **Total: ~1 week**

---

**BYOK is a key feature for the hosted service and needs to be implemented before launch!** 🎯

