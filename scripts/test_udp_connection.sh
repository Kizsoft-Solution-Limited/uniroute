#!/bin/bash

# Test UDP tunnel connection
# Usage: ./test_udp_connection.sh <port>
# Example: ./test_udp_connection.sh 20001

if [ -z "$1" ]; then
    echo "Usage: $0 <allocated-port>"
    echo "Example: $0 20001"
    echo ""
    echo "Get the allocated port from the tunnel CLI output"
    exit 1
fi

PORT=$1

echo "Testing UDP connection to localhost:$PORT"
echo "Sending test packet..."
echo ""

echo "Hello UDP from test script" | nc -u localhost $PORT

echo ""
echo "✅ UDP packet sent! Check the UDP server terminal to see if it was received."
