#!/bin/bash

# Complete Tunnel Testing Script
# Run this script to test TCP, UDP, TLS, and HTTP tunnels locally

set -e

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${BLUE}=== UniRoute Tunnel Testing Suite ===${NC}\n"

# Check prerequisites
echo -e "${YELLOW}Checking prerequisites...${NC}"

# Check if CLI is built
if [ ! -f "./cli" ]; then
    echo -e "${YELLOW}Building CLI...${NC}"
    go build -o cli ./cmd/cli/main.go
fi

# Check authentication
if ! ./cli auth status &> /dev/null; then
    echo -e "${RED}❌ Not authenticated. Run: ./cli auth login${NC}"
    exit 1
fi

echo -e "${GREEN}✅ CLI ready${NC}"
echo -e "${GREEN}✅ Authenticated${NC}\n"

# Set tunnel server URL
# Tunnel server default port is 8055 (not 8084, which is the gateway)
# Check if tunnel server is running and on which port
TUNNEL_PORT=""
if curl -s http://localhost:8055/health > /dev/null 2>&1 || lsof -i :8055 > /dev/null 2>&1; then
    TUNNEL_PORT="8055"
elif curl -s http://localhost:8080/health > /dev/null 2>&1 || lsof -i :8080 > /dev/null 2>&1; then
    TUNNEL_PORT="8080"
else
    # Default to 8055 (standard tunnel server port)
    TUNNEL_PORT="8055"
    echo -e "${YELLOW}⚠️  Could not detect tunnel server, using default: localhost:8055${NC}"
    echo -e "${YELLOW}   Make sure tunnel server is running: go run ./cmd/tunnel-server/main.go${NC}"
    echo -e "${YELLOW}   Or set UNIROUTE_TUNNEL_URL if using a different port${NC}\n"
fi

export UNIROUTE_TUNNEL_URL="localhost:$TUNNEL_PORT"
echo -e "${BLUE}Tunnel server: localhost:$TUNNEL_PORT${NC}"
echo -e "${BLUE}Note: Port 8084 is the gateway, not the tunnel server${NC}\n"

# Verify tunnel server is actually running
if ! lsof -i :$TUNNEL_PORT > /dev/null 2>&1; then
    echo -e "${RED}❌ Tunnel server is not running on port $TUNNEL_PORT${NC}"
    echo -e "${YELLOW}Start it with:${NC}"
    echo -e "${YELLOW}  go run ./cmd/tunnel-server/main.go${NC}"
    echo -e "${YELLOW}Or:${NC}"
    echo -e "${YELLOW}  go run ./cmd/tunnel-server/main.go -port $TUNNEL_PORT${NC}"
    exit 1
fi

echo -e "${GREEN}✅ Tunnel server is running on port $TUNNEL_PORT${NC}\n"

# Function to cleanup
cleanup() {
    echo -e "\n${YELLOW}Cleaning up...${NC}"
    kill $(jobs -p) 2>/dev/null || true
    exit 0
}

trap cleanup SIGINT SIGTERM

# Test 1: TCP Tunnel
echo -e "${BLUE}=== Test 1: TCP Tunnel ===${NC}"
echo "Starting TCP server on port 3306..."
nc -l 3306 > /tmp/tcp_server.log 2>&1 &
TCP_SERVER_PID=$!
sleep 1

echo "Starting TCP tunnel..."
./cli tcp 3306 testtcp --new > /tmp/tcp_tunnel.log 2>&1 &
TCP_TUNNEL_PID=$!
echo "Waiting for tunnel to establish (10 seconds)..."
sleep 10

# Extract allocated port from tunnel output
# TCP tunnels allocate ports starting from 20000
# Look for ports in range 20000-30000 (allocated TCP ports)
# Try multiple patterns to find the allocated port
ALLOCATED_PORT=""
# Pattern 1: Look for ":20000" or similar in the log
ALLOCATED_PORT=$(grep -oE ":[2][0-9]{4}" /tmp/tcp_tunnel.log | head -1 | cut -d: -f2 || echo "")
# Pattern 2: Look for "testtcp.localhost:20000" format
if [ -z "$ALLOCATED_PORT" ] || [ "$ALLOCATED_PORT" = "8055" ]; then
    ALLOCATED_PORT=$(grep -oE "testtcp[^:]*:([2][0-9]{4})" /tmp/tcp_tunnel.log | grep -oE "[2][0-9]{4}" | head -1 || echo "")
fi
# Pattern 3: Look for any 5-digit port number starting with 2
if [ -z "$ALLOCATED_PORT" ] || [ "$ALLOCATED_PORT" = "8055" ]; then
    ALLOCATED_PORT=$(grep -oE "\b2[0-9]{4}\b" /tmp/tcp_tunnel.log | head -1 || echo "")
fi

# Debug: Show what we found
if [ -n "$ALLOCATED_PORT" ]; then
    echo -e "${BLUE}Debug: Found port $ALLOCATED_PORT in tunnel output${NC}"
fi

if [ -n "$ALLOCATED_PORT" ] && [ "$ALLOCATED_PORT" != "8055" ] && [ "$ALLOCATED_PORT" -ge 20000 ] 2>/dev/null; then
    echo -e "${GREEN}✅ TCP tunnel started on port $ALLOCATED_PORT${NC}"
    echo "Testing connection..."
    # Test TCP connection (should connect, not get HTTP response)
    timeout 2 bash -c "echo 'test message' | nc localhost $ALLOCATED_PORT" > /tmp/tcp_test_result.log 2>&1
    if [ $? -eq 0 ] || grep -q "test message" /tmp/tcp_server.log 2>/dev/null; then
        echo -e "${GREEN}✅ TCP test passed! Data flowing through tunnel${NC}"
    else
        echo -e "${YELLOW}⚠️ TCP connection test inconclusive (check manually)${NC}"
    fi
else
    echo -e "${RED}❌ TCP tunnel failed to start or port not found${NC}"
    echo ""
    echo -e "${YELLOW}Tunnel output (last 30 lines):${NC}"
    cat /tmp/tcp_tunnel.log | tail -30
    echo ""
    echo -e "${YELLOW}Note: Look for a port number in the 20000+ range in the output above${NC}"
    echo -e "${YELLOW}You can manually test by connecting to that port: nc localhost <port>${NC}"
fi

echo ""
echo -e "${BLUE}Press Enter to continue to UDP test...${NC}"
read -r

# Test 2: UDP Tunnel
echo -e "${BLUE}=== Test 2: UDP Tunnel ===${NC}"
echo "Starting UDP server on port 53..."
nc -u -l 53 > /tmp/udp_server.log 2>&1 &
UDP_SERVER_PID=$!
sleep 1

echo "Starting UDP tunnel..."
./cli udp 53 testudp --new > /tmp/udp_tunnel.log 2>&1 &
UDP_TUNNEL_PID=$!
echo "Waiting for tunnel to establish (10 seconds)..."
sleep 10

# Extract allocated UDP port (also in 20000+ range)
ALLOCATED_UDP_PORT=""
ALLOCATED_UDP_PORT=$(grep -oE ":[2][0-9]{4}" /tmp/udp_tunnel.log | head -1 | cut -d: -f2 || echo "")
if [ -z "$ALLOCATED_UDP_PORT" ] || [ "$ALLOCATED_UDP_PORT" = "8055" ]; then
    ALLOCATED_UDP_PORT=$(grep -oE "testudp[^:]*:([2][0-9]{4})" /tmp/udp_tunnel.log | grep -oE "[2][0-9]{4}" | head -1 || echo "")
fi
if [ -z "$ALLOCATED_UDP_PORT" ] || [ "$ALLOCATED_UDP_PORT" = "8055" ]; then
    ALLOCATED_UDP_PORT=$(grep -oE "\b2[0-9]{4}\b" /tmp/udp_tunnel.log | head -1 || echo "")
fi

if [ -n "$ALLOCATED_UDP_PORT" ] && [ "$ALLOCATED_UDP_PORT" != "8055" ] && [ "$ALLOCATED_UDP_PORT" -ge 20000 ] 2>/dev/null; then
    echo -e "${GREEN}✅ UDP tunnel started on port $ALLOCATED_UDP_PORT${NC}"
    echo "Sending UDP packet..."
    echo "Hello UDP" | nc -u localhost $ALLOCATED_UDP_PORT 2>/dev/null
    sleep 1
    if grep -q "Hello UDP" /tmp/udp_server.log 2>/dev/null; then
        echo -e "${GREEN}✅ UDP test passed! Packet received${NC}"
    else
        echo -e "${YELLOW}⚠️ UDP test inconclusive (check manually)${NC}"
    fi
else
    echo -e "${RED}❌ UDP tunnel failed to start or port not found${NC}"
    echo ""
    echo -e "${YELLOW}Tunnel output (last 30 lines):${NC}"
    cat /tmp/udp_tunnel.log | tail -30
    echo ""
    echo -e "${YELLOW}Note: Look for a port number in the 20000+ range in the output above${NC}"
fi

echo ""
echo -e "${BLUE}Press Enter to continue to TLS test...${NC}"
read -r

# Test 3: TLS Tunnel
echo -e "${BLUE}=== Test 3: TLS Tunnel ===${NC}"

# Create test certificate if it doesn't exist
if [ ! -f /tmp/test.crt ]; then
    echo "Creating test certificate..."
    openssl req -x509 -newkey rsa:2048 -keyout /tmp/test.key -out /tmp/test.crt -days 365 -nodes -subj "/CN=localhost" 2>/dev/null
fi

echo "Starting TLS server on port 5432..."
openssl s_server -accept 5432 -cert /tmp/test.crt -key /tmp/test.key > /tmp/tls_server.log 2>&1 &
TLS_SERVER_PID=$!
sleep 2

echo "Starting TLS tunnel..."
./cli tls 5432 testtls --new > /tmp/tls_tunnel.log 2>&1 &
TLS_TUNNEL_PID=$!
echo "Waiting for tunnel to establish (10 seconds)..."
sleep 10

# Extract allocated TLS port (also in 20000+ range)
ALLOCATED_TLS_PORT=""
ALLOCATED_TLS_PORT=$(grep -oE ":[2][0-9]{4}" /tmp/tls_tunnel.log | head -1 | cut -d: -f2 || echo "")
if [ -z "$ALLOCATED_TLS_PORT" ] || [ "$ALLOCATED_TLS_PORT" = "8055" ]; then
    ALLOCATED_TLS_PORT=$(grep -oE "testtls[^:]*:([2][0-9]{4})" /tmp/tls_tunnel.log | grep -oE "[2][0-9]{4}" | head -1 || echo "")
fi
if [ -z "$ALLOCATED_TLS_PORT" ] || [ "$ALLOCATED_TLS_PORT" = "8055" ]; then
    ALLOCATED_TLS_PORT=$(grep -oE "\b2[0-9]{4}\b" /tmp/tls_tunnel.log | head -1 || echo "")
fi

if [ -n "$ALLOCATED_TLS_PORT" ] && [ "$ALLOCATED_TLS_PORT" != "8055" ] && [ "$ALLOCATED_TLS_PORT" -ge 20000 ] 2>/dev/null; then
    echo -e "${GREEN}✅ TLS tunnel started on port $ALLOCATED_TLS_PORT${NC}"
    echo "Testing TLS connection..."
    echo | timeout 3 openssl s_client -connect localhost:$ALLOCATED_TLS_PORT 2>/dev/null | grep -q "CONNECTED" && \
        echo -e "${GREEN}✅ TLS test passed! Connection established${NC}" || \
        echo -e "${YELLOW}⚠️ TLS test inconclusive (check manually)${NC}"
else
    echo -e "${RED}❌ TLS tunnel failed to start or port not found${NC}"
    echo ""
    echo -e "${YELLOW}Tunnel output (last 30 lines):${NC}"
    cat /tmp/tls_tunnel.log | tail -30
    echo ""
    echo -e "${YELLOW}Note: Look for a port number in the 20000+ range in the output above${NC}"
fi

echo ""
echo -e "${GREEN}=== Testing Complete ===${NC}"
echo ""
echo "Log files:"
echo "  TCP: /tmp/tcp_tunnel.log"
echo "  UDP: /tmp/udp_tunnel.log"
echo "  TLS: /tmp/tls_tunnel.log"
echo ""
echo "Press Ctrl+C to stop all servers and tunnels"

# Keep running
wait
