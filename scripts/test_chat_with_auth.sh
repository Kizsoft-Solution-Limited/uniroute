#!/usr/bin/env bash
# Test chat stream endpoint using login token.
# Run with: TEST_EMAIL=your@email.com TEST_PASSWORD=yourpass BASE_URL=https://app.uniroute.co ./scripts/test_chat_with_auth.sh
# Do NOT commit real credentials. Use env vars only.

set -e
BASE_URL="${BASE_URL:-https://app.uniroute.co}"
if [[ -z "$TEST_EMAIL" || -z "$TEST_PASSWORD" ]]; then
  echo "Usage: TEST_EMAIL=... TEST_PASSWORD=... [BASE_URL=https://app.uniroute.co] $0"
  echo "Credentials are read from environment only."
  exit 1
fi

echo "=== 1. Login ==="
LOGIN_RESP=$(curl -s -X POST "$BASE_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"$TEST_EMAIL\",\"password\":\"$TEST_PASSWORD\",\"remember_me\":false}")
TOKEN=$(echo "$LOGIN_RESP" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
if [[ -z "$TOKEN" ]]; then
  echo "Login failed. Response: $LOGIN_RESP"
  exit 1
fi
echo "Login OK (token received)."

echo ""
echo "=== 2. Chat stream (POST /auth/chat/stream) ==="
STREAM_RESP=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/auth/chat/stream" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"model":"gemini-pro","messages":[{"role":"user","content":"Say hello in one word."}]}')
HTTP_CODE=$(echo "$STREAM_RESP" | tail -n 1)
HTTP_BODY=$(echo "$STREAM_RESP" | sed '$d')
echo "HTTP status: $HTTP_CODE"
if [[ "$HTTP_CODE" != "200" ]]; then
  echo "Response body: $HTTP_BODY"
  exit 1
fi
# Show first few data lines
echo "Stream (first 5 data lines):"
echo "$HTTP_BODY" | grep -o 'data: [^[:space:]]*' | head -5
if echo "$HTTP_BODY" | grep -q '"error"'; then
  echo ""
  echo "Error in stream:"
  echo "$HTTP_BODY" | grep 'data: ' | head -1 | sed 's/^data: //' | grep -o '"error":"[^"]*"' || true
fi
echo ""
echo "=== Done ==="
