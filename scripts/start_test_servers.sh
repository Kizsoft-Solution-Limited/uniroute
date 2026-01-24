#!/bin/bash

# Start test servers for tunnel testing
# Run this in a separate terminal before testing tunnels

echo "Starting test servers for tunnel testing..."
echo "Press Ctrl+C to stop all servers"
echo ""

# Function to cleanup on exit
cleanup() {
    echo ""
    echo "Stopping all test servers..."
    kill $(jobs -p) 2>/dev/null
    exit 0
}

trap cleanup SIGINT SIGTERM

# Start TCP server on port 3306
echo "Starting TCP server on port 3306..."
nc -l 3306 &
TCP_PID=$!

# Start UDP server on port 53
echo "Starting UDP server on port 53..."
nc -u -l 53 &
UDP_PID=$!

# Start HTTP server on port 3000
echo "Starting HTTP server on port 3000..."
python3 -m http.server 3000 > /dev/null 2>&1 &
HTTP_PID=$!

echo ""
echo "✅ Test servers running:"
echo "   TCP:  localhost:3306 (PID: $TCP_PID)"
echo "   UDP:  localhost:53 (PID: $UDP_PID)"
echo "   HTTP: localhost:3000 (PID: $HTTP_PID)"
echo ""
echo "Now you can test tunnels in another terminal:"
echo "   uniroute tcp 3306 testtcp"
echo "   uniroute udp 53 testudp"
echo "   uniroute http 3000 testhttp"
echo ""
echo "Press Ctrl+C to stop all servers..."

# Wait for all background jobs
wait
