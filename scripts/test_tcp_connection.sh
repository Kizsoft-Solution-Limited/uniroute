#!/bin/bash

# Test TCP tunnel connection
# Usage: ./test_tcp_connection.sh <port>
# Example: ./test_tcp_connection.sh 20000

if [ -z "$1" ]; then
    echo "Usage: $0 <allocated-port>"
    echo "Example: $0 20000"
    echo ""
    echo "Get the allocated port from the tunnel CLI output"
    exit 1
fi

PORT=$1

echo "Testing TCP connection to localhost:$PORT"
echo "Type something and press Enter (Ctrl+C to exit)"
echo ""

nc localhost $PORT
