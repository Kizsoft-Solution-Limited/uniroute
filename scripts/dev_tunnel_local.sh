#!/usr/bin/env bash
# Local tunnel loop: tunnel server on :8055 + optional CLI usage (no Postgres/JWT required).
#
# Terminal 1:
#   ./scripts/dev_tunnel_local.sh
#
# Terminal 2 (pick a free subdomain label; serve something on the port):
#   cd "$(git rev-parse --show-toplevel 2>/dev/null || pwd)"
#   export UNIROUTE_TUNNEL_URL=localhost:8055   # if needed (CLI auto-probes :8055 when healthy)
#   ./bin/uniroute tunnel --port 3000 --new --host mydev
#
# Then open: http://mydev.localhost:8055/
#
# Env:
#   TUNNEL_DEV_SKIP_AUTH=1   Required on server (set by this script). Never use on production.
#   PORT / -port             Tunnel server listen port (default 8055).

set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

mkdir -p bin
export CGO_ENABLED=0
echo "Building CLI and tunnel server..."
go build -o bin/uniroute cmd/cli/main.go
go build -o bin/uniroute-tunnel-server cmd/tunnel-server/main.go

PORT="${PORT:-8055}"
export ENV="${ENV:-development}"
export TUNNEL_DEV_SKIP_AUTH=1
export PORT
export TUNNEL_LOCALHOST_DOMAIN=localhost
export TUNNEL_BASE_DOMAIN="localhost:${PORT}"

echo ""
echo "Starting tunnel server on http://0.0.0.0:${PORT}/ (WebSocket: ws://localhost:${PORT}/tunnel)"
echo "Public URLs look like: http://<subdomain>.localhost:${PORT}/"
echo ""
echo "Example (terminal 2):"
echo "  export UNIROUTE_TUNNEL_URL=localhost:${PORT}"
echo "  ./bin/uniroute tunnel --port 3000 --new --host mydev"
echo "Then: curl -sS -o /dev/null -w '%{http_code}\\n' http://mydev.localhost:${PORT}/"
echo ""

exec ./bin/uniroute-tunnel-server -port "${PORT}"
