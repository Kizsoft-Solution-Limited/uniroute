#!/usr/bin/env python3
"""
UniRoute Basic Chat Example (Python)

This example demonstrates how to send a chat request to UniRoute API.

Prerequisites:
1. Install requests: pip install requests
2. Set environment variable: export UNIROUTE_API_KEY='ur_your_key_here'
3. Optional: export UNIROUTE_API_URL='http://localhost:8084'
"""

import os
import sys
import json
import requests

def main():
    # Get API key from environment
    api_key = os.getenv("UNIROUTE_API_KEY")
    if not api_key:
        print("❌ Error: UNIROUTE_API_KEY environment variable is required")
        print("")
        print("Get your API key:")
        print("  1. Run: uniroute keys create")
        print("  2. Export: export UNIROUTE_API_KEY='ur_your_key_here'")
        sys.exit(1)

    # Get API URL from environment (default: localhost)
    api_url = os.getenv("UNIROUTE_API_URL", "http://localhost:8084")

    # Create chat request
    request_data = {
        "model": "gpt-4",
        "messages": [
            {
                "role": "user",
                "content": "Hello! Explain what UniRoute is in one sentence."
            }
        ],
        "temperature": 0.7,
        "max_tokens": 100
    }

    # Send request
    try:
        response = requests.post(
            f"{api_url}/v1/chat",
            json=request_data,
            headers={
                "Content-Type": "application/json",
                "Authorization": f"Bearer {api_key}"
            },
            timeout=30
        )
        response.raise_for_status()
    except requests.exceptions.RequestException as e:
        print(f"❌ Error sending request: {e}")
        if "Connection" in str(e):
            print("")
            print("💡 Make sure UniRoute server is running:")
            print("  make dev")
        sys.exit(1)

    # Parse response
    try:
        chat_response = response.json()
    except json.JSONDecodeError as e:
        print(f"❌ Error parsing response: {e}")
        print(f"Response: {response.text}")
        sys.exit(1)

    # Print response
    print("✅ Chat Response:")
    print("")
    if "choices" in chat_response and len(chat_response["choices"]) > 0:
        message = chat_response["choices"][0].get("message", {})
        content = message.get("content", "")
        print(f"💬 {content}")
    print("")
    
    if "usage" in chat_response:
        usage = chat_response["usage"]
        prompt_tokens = usage.get("prompt_tokens", 0)
        completion_tokens = usage.get("completion_tokens", 0)
        total_tokens = usage.get("total_tokens", 0)
        print(f"📊 Tokens: {prompt_tokens} prompt + {completion_tokens} completion = {total_tokens} total")

if __name__ == "__main__":
    main()
