#!/usr/bin/env bash


set -e
BASE_URL="${OLLAMA_BASE_URL:-http://localhost:11434}"
echo "Testing Ollama at $BASE_URL"
echo ""

# 1. Health / list models
echo "1. GET $BASE_URL/api/tags"
if response=$(curl -s -m 10 "$BASE_URL/api/tags" 2>&1); then
  echo "$response" | head -c 800
  echo ""
  if echo "$response" | grep -q '"models"'; then
    echo "   OK: Ollama is reachable and returned models."
  else
    echo "   WARN: Got response but no 'models' key (check output above)."
  fi
else
  echo "   FAIL: Could not reach Ollama. Is it running? (e.g. ollama serve)"
  echo "   $response"
  exit 1
fi

echo ""

# 2. Simple chat (non-streaming) - use a model that exists
MODEL="${OLLAMA_TEST_MODEL:-llama3.2:latest}"
echo "2. POST $BASE_URL/api/chat (model=$MODEL, stream=false)"
body=$(cat <<EOF
{"model":"$MODEL","messages":[{"role":"user","content":"Reply with exactly: OK"}],"stream":false}
EOF
)
if response=$(curl -s -m 60 -X POST "$BASE_URL/api/chat" \
  -H "Content-Type: application/json" \
  -d "$body" 2>&1); then
  if echo "$response" | grep -q '"message"'; then
    echo "   OK: Chat request succeeded."
    echo "$response" | head -c 600
    echo ""
  else
    echo "   Response (first 600 chars):"
    echo "$response" | head -c 600
    echo ""
    if echo "$response" | grep -q 'model not found'; then
      echo "   TIP: Pull the model first: ollama pull $MODEL"
    fi
  fi
else
  echo "   FAIL: Chat request failed or timed out."
  echo "   $response"
  exit 1
fi

echo ""
echo "Done. If both steps passed, OLLAMA_BASE_URL=$BASE_URL is valid for UniRoute."
echo "To test via UniRoute backend, start the gateway with OLLAMA_BASE_URL set and call POST /auth/chat or /auth/chat/stream with model=$MODEL."
