# Postman Testing Guide for Phase 2

This guide shows you how to test UniRoute Phase 2 features using Postman.

## Prerequisites

1. **UniRoute server running**
   ```bash
   make dev
   # or
   ./bin/uniroute
   ```

2. **PostgreSQL and Redis** (for Phase 2 features)
   - Or use Phase 1 mode (in-memory API keys)

3. **Postman** installed

## Server Setup

### Phase 1 Mode (Simple - No Database Required)

```bash
# Just run the server
./bin/uniroute

# Server will generate a default API key on startup
# Look for: "Generated default API key (save this!)"
```

### Phase 2 Mode (Full Features - Requires Database)

```bash
# Set environment variables
export DATABASE_URL="postgres://user:password@localhost/uniroute?sslmode=disable"
export REDIS_URL="redis://localhost:6379"
export JWT_SECRET="your-secret-key-min-32-chars"

# Run server
./bin/uniroute
```

## Postman Collection Setup

### 1. Create a New Collection

1. Open Postman
2. Click "New" → "Collection"
3. Name it "UniRoute Phase 2"

### 2. Set Collection Variables

Go to Collection → Variables and add:

| Variable | Initial Value | Current Value |
|----------|---------------|---------------|
| `base_url` | `http://localhost:8084` | `http://localhost:8084` |
| `api_key` | `your-api-key-here` | `your-api-key-here` |
| `jwt_token` | `your-jwt-token-here` | `your-jwt-token-here` |

## API Endpoints

### 1. Health Check (No Auth Required)

**Request:**
```
GET {{base_url}}/health
```

**Expected Response:**
```json
{
  "status": "ok"
}
```

**Postman Setup:**
- Method: `GET`
- URL: `{{base_url}}/health`
- Headers: None required

---

### 2. Create API Key (Phase 2 - Requires JWT)

**Request:**
```
POST {{base_url}}/admin/api-keys
Authorization: Bearer {{jwt_token}}
Content-Type: application/json

{
  "name": "My Test Key",
  "rate_limit_per_minute": 60,
  "rate_limit_per_day": 10000
}
```

**Expected Response:**
```json
{
  "id": "uuid-here",
  "key": "ur_abc123...",
  "name": "My Test Key",
  "created_at": "2024-01-04T15:30:00Z",
  "expires_at": null,
  "message": "Save this key - it will not be shown again"
}
```

**Postman Setup:**
- Method: `POST`
- URL: `{{base_url}}/admin/api-keys`
- Headers:
  - `Authorization: Bearer {{jwt_token}}`
  - `Content-Type: application/json`
- Body (raw JSON):
  ```json
  {
    "name": "My Test Key",
    "rate_limit_per_minute": 60,
    "rate_limit_per_day": 10000
  }
  ```

**Note:** Save the returned `key` value to your `api_key` variable!

---

### 3. List API Keys (Phase 2 - Requires JWT)

**Request:**
```
GET {{base_url}}/admin/api-keys
Authorization: Bearer {{jwt_token}}
```

**Expected Response:**
```json
{
  "keys": [],
  "message": "API key listing will be implemented"
}
```

**Postman Setup:**
- Method: `GET`
- URL: `{{base_url}}/admin/api-keys`
- Headers:
  - `Authorization: Bearer {{jwt_token}}`

---

### 4. Chat Completion (Main Endpoint)

**Request:**
```
POST {{base_url}}/v1/chat
Authorization: Bearer {{api_key}}
Content-Type: application/json

{
  "model": "llama2",
  "messages": [
    {
      "role": "user",
      "content": "Hello, how are you?"
    }
  ]
}
```

**Expected Response:**
```json
{
  "id": "chat-1234567890",
  "model": "llama2",
  "provider": "local",
  "choices": [
    {
      "message": {
        "role": "assistant",
        "content": "Hello! I'm doing well, thank you for asking..."
      }
    }
  ],
  "usage": {
    "prompt_tokens": 10,
    "completion_tokens": 15,
    "total_tokens": 25
  },
  "cost": 0,
  "latency_ms": 150
}
```

**Postman Setup:**
- Method: `POST`
- URL: `{{base_url}}/v1/chat`
- Headers:
  - `Authorization: Bearer {{api_key}}`
  - `Content-Type: application/json`
- Body (raw JSON):
  ```json
  {
    "model": "llama2",
    "messages": [
      {
        "role": "user",
        "content": "Hello, how are you?"
      }
    ]
  }
  ```

---

## Testing Scenarios

### Scenario 1: Test API Key Authentication

**Test Valid API Key:**
1. Use the chat endpoint with a valid API key
2. Should return `200 OK` with chat response

**Test Invalid API Key:**
1. Change `{{api_key}}` to `invalid_key`
2. Should return `401 Unauthorized`
3. Response:
   ```json
   {
     "error": "invalid API key"
   }
   ```

**Test Missing API Key:**
1. Remove `Authorization` header
2. Should return `401 Unauthorized`

---

### Scenario 2: Test Rate Limiting (Phase 2)

**Setup:**
1. Create an API key with low limits:
   ```json
   {
     "name": "Rate Limit Test",
     "rate_limit_per_minute": 3,
     "rate_limit_per_day": 10
   }
   ```

**Test:**
1. Make 3 requests to `/v1/chat` - all should succeed
2. Check response headers:
   - `X-RateLimit-Limit-PerMinute: 3`
   - `X-RateLimit-Remaining-PerMinute: 2, 1, 0`
3. Make 4th request - should return `429 Too Many Requests`
4. Response:
   ```json
   {
     "error": "rate limit exceeded"
   }
   ```

**Postman Test Script:**
Add this to the "Tests" tab:
```javascript
// Check rate limit headers
pm.test("Rate limit headers present", function () {
    pm.response.to.have.header("X-RateLimit-Limit-PerMinute");
    pm.response.to.have.header("X-RateLimit-Remaining-PerMinute");
});

// Check if rate limited
pm.test("Rate limit check", function () {
    if (pm.response.code === 429) {
        pm.expect(pm.response.json().error).to.include("rate limit");
    }
});
```

---

### Scenario 3: Test Security Headers

**Request:**
```
GET {{base_url}}/health
```

**Check Response Headers:**
- `X-Frame-Options: DENY`
- `X-Content-Type-Options: nosniff`
- `X-XSS-Protection: 1; mode=block`
- `Content-Security-Policy: default-src 'self'`
- `Referrer-Policy: strict-origin-when-cross-origin`

**Postman Test Script:**
```javascript
pm.test("Security headers present", function () {
    pm.response.to.have.header("X-Frame-Options");
    pm.response.to.have.header("X-Content-Type-Options");
    pm.response.to.have.header("X-XSS-Protection");
    pm.response.to.have.header("Content-Security-Policy");
    pm.response.to.have.header("Referrer-Policy");
});

pm.test("X-Frame-Options is DENY", function () {
    pm.response.to.have.header("X-Frame-Options", "DENY");
});
```

---

### Scenario 4: Test JWT Authentication (Phase 2)

**Note:** You'll need to generate a JWT token first. For now, you can:
1. Use a JWT generation tool (jwt.io)
2. Or implement a login endpoint (future phase)

**Test Valid JWT:**
```
POST {{base_url}}/admin/api-keys
Authorization: Bearer {{jwt_token}}
```

**Test Invalid JWT:**
1. Use an invalid token
2. Should return `401 Unauthorized`

---

### Scenario 5: Test IP Whitelisting (Phase 2)

**Setup:**
```bash
export IP_WHITELIST="192.168.1.100,10.0.0.1"
./bin/uniroute
```

**Test:**
1. Request from whitelisted IP - should succeed
2. Request from non-whitelisted IP - should return `403 Forbidden`

---

## Postman Collection JSON

Save this as `UniRoute.postman_collection.json`:

```json
{
  "info": {
    "name": "UniRoute Phase 2",
    "schema": "https://schema.getpostman.com/json/collection/v2.1.0/collection.json"
  },
  "variable": [
    {
      "key": "base_url",
      "value": "http://localhost:8084"
    },
    {
      "key": "api_key",
      "value": ""
    },
    {
      "key": "jwt_token",
      "value": ""
    }
  ],
  "item": [
    {
      "name": "Health Check",
      "request": {
        "method": "GET",
        "header": [],
        "url": {
          "raw": "{{base_url}}/health",
          "host": ["{{base_url}}"],
          "path": ["health"]
        }
      }
    },
    {
      "name": "Chat Completion",
      "request": {
        "method": "POST",
        "header": [
          {
            "key": "Authorization",
            "value": "Bearer {{api_key}}"
          },
          {
            "key": "Content-Type",
            "value": "application/json"
          }
        ],
        "body": {
          "mode": "raw",
          "raw": "{\n  \"model\": \"llama2\",\n  \"messages\": [\n    {\n      \"role\": \"user\",\n      \"content\": \"Hello, how are you?\"\n    }\n  ]\n}"
        },
        "url": {
          "raw": "{{base_url}}/v1/chat",
          "host": ["{{base_url}}"],
          "path": ["v1", "chat"]
        }
      }
    },
    {
      "name": "Create API Key",
      "request": {
        "method": "POST",
        "header": [
          {
            "key": "Authorization",
            "value": "Bearer {{jwt_token}}"
          },
          {
            "key": "Content-Type",
            "value": "application/json"
          }
        ],
        "body": {
          "mode": "raw",
          "raw": "{\n  \"name\": \"My Test Key\",\n  \"rate_limit_per_minute\": 60,\n  \"rate_limit_per_day\": 10000\n}"
        },
        "url": {
          "raw": "{{base_url}}/admin/api-keys",
          "host": ["{{base_url}}"],
          "path": ["admin", "api-keys"]
        }
      }
    }
  ]
}
```

## Quick Start Checklist

- [ ] Start UniRoute server (`make dev`)
- [ ] Get API key (from server logs or create via admin endpoint)
- [ ] Set `base_url` variable in Postman
- [ ] Set `api_key` variable in Postman
- [ ] Test health endpoint
- [ ] Test chat endpoint
- [ ] Test rate limiting (make multiple requests)
- [ ] Test invalid API key
- [ ] Check security headers

## Troubleshooting

### 401 Unauthorized
- Check API key is correct
- Check `Authorization` header format: `Bearer <key>`
- Verify API key is active in database (Phase 2)

### 429 Too Many Requests
- Rate limit exceeded
- Wait for rate limit window to reset
- Check `X-RateLimit-Remaining-*` headers

### Connection Refused
- Verify server is running on port 8084
- Check firewall settings
- Verify `base_url` is correct

### 500 Internal Server Error
- Check server logs
- Verify PostgreSQL/Redis connections (Phase 2)
- Check Ollama is running (for local LLM)

## Advanced: Automated Testing

### Postman Pre-request Script

Extract API key from create response:
```javascript
// In "Create API Key" request, Tests tab:
if (pm.response.code === 201) {
    const response = pm.response.json();
    pm.collectionVariables.set("api_key", response.key);
    console.log("API key saved:", response.key);
}
```

### Postman Test Scripts

Rate limit test:
```javascript
pm.test("Status code is 200 or 429", function () {
    pm.expect([200, 429]).to.include(pm.response.code);
});

if (pm.response.code === 429) {
    pm.test("Rate limit error message", function () {
        pm.expect(pm.response.json().error).to.include("rate limit");
    });
}
```

---

## Example Workflow

1. **Start Server**
   ```bash
   ./bin/uniroute
   # Note the API key from logs
   ```

2. **Set API Key in Postman**
   - Copy API key from server logs
   - Set in collection variables

3. **Test Health**
   - GET `/health` - Should return `{"status": "ok"}`

4. **Test Chat**
   - POST `/v1/chat` with API key
   - Should return chat response

5. **Test Rate Limiting** (Phase 2)
   - Make multiple requests quickly
   - Watch `X-RateLimit-Remaining-*` headers
   - Should get 429 after limit

6. **Test Security Headers**
   - Check response headers
   - All security headers should be present

