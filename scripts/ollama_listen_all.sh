#!/usr/bin/env bash
# Run on the server where Ollama is installed.
# Makes Ollama listen on 0.0.0.0 so Coolify/remote backend can reach it.
# Usage: copy to server and run: sudo bash ollama_listen_all.sh

set -e
SERVICE_FILE="${SERVICE_FILE:-/etc/systemd/system/ollama.service}"

if [[ ! -f "$SERVICE_FILE" ]]; then
  echo "Ollama systemd service not found at $SERVICE_FILE"
  exit 1
fi

if grep -q 'OLLAMA_HOST' "$SERVICE_FILE"; then
  echo "OLLAMA_HOST already set in $SERVICE_FILE"
  sudo systemctl daemon-reload
  sudo systemctl restart ollama
  echo "Ollama restarted."
  exit 0
fi

# Add Environment="OLLAMA_HOST=0.0.0.0" after [Service]
sudo sed -i '/^\[Service\]/a Environment="OLLAMA_HOST=0.0.0.0"' "$SERVICE_FILE"
sudo systemctl daemon-reload
sudo systemctl restart ollama
echo "Ollama now listens on 0.0.0.0:11434. Restarted."
ss -tlnp | grep 11434 || true
