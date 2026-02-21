#!/bin/bash

# UniRoute One-Line Installation Script
# Usage: curl -fsSL https://raw.githubusercontent.com/Kizsoft-Solution-Limited/uniroute/main/scripts/install.sh | bash
# Or: bash <(curl -fsSL https://raw.githubusercontent.com/Kizsoft-Solution-Limited/uniroute/main/scripts/install.sh)

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}UniRoute CLI Installation${NC}"
echo ""

# Detect OS and Architecture
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"

# Map architecture
case "$ARCH" in
  x86_64)
    ARCH="amd64"
    ;;
  arm64|aarch64)
    ARCH="arm64"
    ;;
  *)
    echo -e "${RED}Error: Unsupported architecture: $ARCH${NC}"
    exit 1
    ;;
esac

# Map OS
case "$OS" in
  linux)
    OS="linux"
    ;;
  darwin)
    OS="darwin"
    ;;
  *)
    echo -e "${RED}Error: Unsupported OS: $OS${NC}"
    echo "Please download manually from: https://github.com/Kizsoft-Solution-Limited/uniroute/releases"
    exit 1
    ;;
esac

BINARY_NAME="uniroute-${OS}-${ARCH}"
RELEASE_URL="https://github.com/Kizsoft-Solution-Limited/uniroute/releases/latest/download/${BINARY_NAME}"

echo -e "${BLUE}Platform: ${OS}/${ARCH}${NC}"
echo -e "${BLUE}Downloading from: ${RELEASE_URL}${NC}"
echo ""

# Download
TMP_DIR=$(mktemp -d)
trap "rm -rf $TMP_DIR" EXIT

if command -v curl &> /dev/null; then
  curl -fsSL -o "${TMP_DIR}/uniroute" "${RELEASE_URL}"
elif command -v wget &> /dev/null; then
  wget -q -O "${TMP_DIR}/uniroute" "${RELEASE_URL}"
else
  echo -e "${RED}Error: Neither curl nor wget is installed${NC}"
  exit 1
fi

# Make executable
chmod +x "${TMP_DIR}/uniroute"

# Install to /usr/local/bin (requires sudo)
INSTALL_PATH="/usr/local/bin/uniroute"

if [ -w "$(dirname $INSTALL_PATH)" ]; then
  # No sudo needed
  mv "${TMP_DIR}/uniroute" "$INSTALL_PATH"
  echo -e "${GREEN}Installed to: ${INSTALL_PATH}${NC}"
else
  # Need sudo
  echo -e "${YELLOW}Requires sudo to install to ${INSTALL_PATH}${NC}"
  sudo mv "${TMP_DIR}/uniroute" "$INSTALL_PATH"
  echo -e "${GREEN}Installed to: ${INSTALL_PATH}${NC}"
fi

echo ""
echo -e "${GREEN}Installation complete${NC}"
echo ""
echo -e "${BLUE}Next steps:${NC}"
echo "  1. Verify installation:"
echo "     ${GREEN}uniroute --version${NC}"
echo ""
echo "  2. Login (uses hosted UniRoute by default):"
echo "     ${GREEN}uniroute auth login${NC}"
echo "     For a local server instead: ${GREEN}uniroute auth login --server http://localhost:8084${NC}"
echo ""
echo "  3. Start a tunnel:"
echo "     ${GREEN}uniroute http 8080${NC}"
echo ""
echo -e "${BLUE}Documentation:${NC}"
echo "  - Docs: https://uniroute.co/docs"
echo "  - CLI: https://uniroute.co/docs/cli"
echo ""
