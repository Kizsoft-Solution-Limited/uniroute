#!/bin/bash

# Local Tunnel Testing Script
# This script helps test TCP, TLS, and UDP tunnels locally

set -e

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${BLUE}=== UniRoute Local Tunnel Testing ===${NC}\n"

# Check if uniroute CLI is available
if ! command -v uniroute &> /dev/null; then
    echo -e "${RED}Error: uniroute CLI not found in PATH${NC}"
    echo "Please build the CLI first: go build -o uniroute ./cmd/cli"
    exit 1
fi

# Check if user is authenticated
if ! uniroute auth status &> /dev/null; then
    echo -e "${YELLOW}Warning: Not authenticated. Run 'uniroute auth login' first${NC}"
    exit 1
fi

echo -e "${GREEN}Starting local tunnel tests...${NC}\n"

# Function to test TCP tunnel
test_tcp() {
    echo -e "${BLUE}--- Testing TCP Tunnel ---${NC}"
    echo "1. Starting TCP tunnel on port 3306..."
    echo "   Command: uniroute tcp 3306 testtcp"
    echo ""
    echo "2. In another terminal, start a TCP server:"
    echo "   nc -l 3306"
    echo ""
    echo "3. In a third terminal, connect to the tunnel:"
    echo "   nc localhost <allocated-port>"
    echo "   (Check the CLI output for the allocated port)"
    echo ""
    echo "4. Type something in terminal 3, it should appear in terminal 2"
    echo ""
    read -p "Press Enter when ready to test TCP tunnel..."
}

# Function to test TLS tunnel
test_tls() {
    echo -e "${BLUE}--- Testing TLS Tunnel ---${NC}"
    echo "1. Starting TLS tunnel on port 5432..."
    echo "   Command: uniroute tls 5432 testtls"
    echo ""
    echo "2. In another terminal, start a TLS server (requires cert):"
    echo "   openssl s_server -accept 5432 -cert server.crt -key server.key"
    echo "   Or use a PostgreSQL instance with TLS enabled"
    echo ""
    echo "3. In a third terminal, connect with TLS client:"
    echo "   openssl s_client -connect localhost:<allocated-port>"
    echo ""
    read -p "Press Enter when ready to test TLS tunnel..."
}

# Function to test UDP tunnel
test_udp() {
    echo -e "${BLUE}--- Testing UDP Tunnel ---${NC}"
    echo "1. Starting UDP tunnel on port 53..."
    echo "   Command: uniroute udp 53 testudp"
    echo ""
    echo "2. In another terminal, start a UDP server:"
    echo "   nc -u -l 53"
    echo ""
    echo "3. In a third terminal, send UDP packets:"
    echo "   echo 'test' | nc -u localhost <allocated-port>"
    echo ""
    read -p "Press Enter when ready to test UDP tunnel..."
}

# Function to create simple test servers
create_test_servers() {
    echo -e "${BLUE}--- Creating Test Servers ---${NC}"
    
    # Create a simple TCP test server script
    cat > /tmp/test_tcp_server.sh << 'EOF'
#!/bin/bash
echo "TCP Test Server listening on port 3306"
echo "Connect to this server to test TCP tunnel"
echo "Press Ctrl+C to stop"
nc -l 3306
EOF
    chmod +x /tmp/test_tcp_server.sh
    
    # Create a simple UDP test server script
    cat > /tmp/test_udp_server.sh << 'EOF'
#!/bin/bash
echo "UDP Test Server listening on port 53"
echo "Send UDP packets to this server to test UDP tunnel"
echo "Press Ctrl+C to stop"
nc -u -l 53
EOF
    chmod +x /tmp/test_udp_server.sh
    
    echo -e "${GREEN}Test server scripts created:${NC}"
    echo "  TCP: /tmp/test_tcp_server.sh"
    echo "  UDP: /tmp/test_udp_server.sh"
    echo ""
}

# Main menu
echo "Select test to run:"
echo "1) Test TCP tunnel"
echo "2) Test TLS tunnel"
echo "3) Test UDP tunnel"
echo "4) Create test server scripts"
echo "5) Test all (interactive)"
echo "6) Exit"
echo ""
read -p "Enter choice [1-6]: " choice

case $choice in
    1)
        test_tcp
        echo -e "${GREEN}Starting TCP tunnel...${NC}"
        uniroute tcp 3306 testtcp
        ;;
    2)
        test_tls
        echo -e "${GREEN}Starting TLS tunnel...${NC}"
        uniroute tls 5432 testtls
        ;;
    3)
        test_udp
        echo -e "${GREEN}Starting UDP tunnel...${NC}"
        uniroute udp 53 testudp
        ;;
    4)
        create_test_servers
        ;;
    5)
        create_test_servers
        echo ""
        test_tcp
        test_tls
        test_udp
        echo -e "${GREEN}All tests ready!${NC}"
        ;;
    6)
        exit 0
        ;;
    *)
        echo -e "${RED}Invalid choice${NC}"
        exit 1
        ;;
esac
