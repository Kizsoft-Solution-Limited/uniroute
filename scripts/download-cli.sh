#!/bin/bash

# UniRoute CLI Download Script
# Downloads the latest pre-built binary for your platform

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

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

echo -e "${GREEN}Downloading UniRoute CLI...${NC}"
echo "Platform: ${OS}/${ARCH}"
echo "URL: ${RELEASE_URL}"
echo ""

# Download
if command -v curl &> /dev/null; then
  curl -L -o uniroute "${RELEASE_URL}"
elif command -v wget &> /dev/null; then
  wget -O uniroute "${RELEASE_URL}"
else
  echo -e "${RED}Error: Neither curl nor wget is installed${NC}"
  exit 1
fi

# Make executable
chmod +x uniroute

echo -e "${GREEN}✓ Downloaded successfully!${NC}"
echo ""
echo "Binary location: $(pwd)/uniroute"
echo ""
echo -e "${YELLOW}To install globally, run:${NC}"
echo "  sudo mv uniroute /usr/local/bin/"
echo ""
echo -e "${YELLOW}Or add to your PATH:${NC}"
echo "  export PATH=\"\$(pwd):\$PATH\""
echo ""
echo "Verify installation:"
echo "  ./uniroute --version"

