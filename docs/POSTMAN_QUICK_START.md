# Postman Quick Start Guide

## 🚀 Quick Setup (5 minutes)

### Step 1: Start UniRoute Server

```bash
# Terminal 1: Start the server
cd /Users/adikekizito/GoDev/uniroute
make dev

# Look for this line in the output:
# "Generated default API key (save this!)"
# Copy the API key that appears
```

### Step 2: Import Postman Collection

1. Open Postman
2. Click **Import** button
3. Select `UniRoute.postman_collection.json`
4. Collection imported! ✅

### Step 3: Set Variables

1. Click on **UniRoute Phase 2** collection
2. Go to **Variables** tab
3. Set:
   - `base_url`: `http://localhost:8084`
   - `api_key`: `[paste the API key from server logs]`

### Step 4: Test!

1. **Health Check** - Click "Send"
   - ✅ Should return `{"status": "ok"}`

2. **Chat Completion** - Click "Send"
   - ✅ Should return chat response
   - Check response headers for rate limit info

## 📋 Common Requests

### Basic Chat Request

**URL:** `POST http://localhost:8084/v1/chat`

**Headers:**
```
Authorization: Bearer YOUR_API_KEY_HERE
Content-Type: application/json
```

**Body:**
```json
{
  "model": "llama2",
  "messages": [
    {
      "role": "user",
      "content": "What is 2+2?"
    }
  ]
}
```

### Test Rate Limiting

1. Use the "Test Rate Limiting" request
2. Click "Send" multiple times quickly (5-10 times)
3. Watch the response:
   - First few: `200 OK`
   - After limit: `429 Too Many Requests`
4. Check headers:
   - `X-RateLimit-Remaining-PerMinute` decreases
   - `X-RateLimit-Limit-PerMinute` shows your limit

## 🔍 What to Check

### Response Headers (All Requests)

Look for these security headers:
- ✅ `X-Frame-Options: DENY`
- ✅ `X-Content-Type-Options: nosniff`
- ✅ `X-XSS-Protection: 1; mode=block`
- ✅ `Content-Security-Policy: default-src 'self'`

### Rate Limit Headers (Chat Requests)

- ✅ `X-RateLimit-Limit-PerMinute`
- ✅ `X-RateLimit-Remaining-PerMinute`
- ✅ `X-RateLimit-Limit-PerDay`
- ✅ `X-RateLimit-Remaining-PerDay`

### Response Body (Chat)

```json
{
  "id": "chat-...",
  "model": "llama2",
  "provider": "local",
  "choices": [...],
  "usage": {
    "prompt_tokens": 10,
    "completion_tokens": 15,
    "total_tokens": 25
  },
  "cost": 0,
  "latency_ms": 150
}
```

## 🐛 Troubleshooting

### "Connection Refused"
- ✅ Server running? Check terminal
- ✅ Port 8084? Check server logs
- ✅ Correct `base_url`? Should be `http://localhost:8084`

### "401 Unauthorized"
- ✅ API key correct? Check collection variables
- ✅ Header format? Should be `Bearer YOUR_KEY`
- ✅ Key active? (Phase 2 - check database)

### "429 Too Many Requests"
- ✅ This is expected! Rate limit working
- ✅ Wait 1 minute and try again
- ✅ Or create new API key with higher limits

### "500 Internal Server Error"
- ✅ Check server logs
- ✅ Ollama running? (for local LLM)
- ✅ Database connected? (Phase 2)

## 🎯 Testing Checklist

- [ ] Health check returns `200 OK`
- [ ] Chat endpoint works with valid API key
- [ ] Invalid API key returns `401`
- [ ] Security headers present
- [ ] Rate limit headers present
- [ ] Rate limiting works (429 after limit)
- [ ] Response format is correct

## 💡 Pro Tips

1. **Save API Key Automatically**
   - The "Create API Key" request saves the key automatically
   - Check collection variables after creating

2. **Test Scripts**
   - All requests have test scripts
   - Check "Test Results" tab after sending

3. **Environment Variables**
   - Create different environments for dev/staging/prod
   - Switch easily between them

4. **Collection Runner**
   - Run all requests in sequence
   - Great for regression testing

## 📚 Next Steps

- Read `POSTMAN_TESTING_GUIDE.md` for detailed scenarios
- Test Phase 2 features (JWT, database-backed keys)
- Set up different environments
- Create custom test scenarios

