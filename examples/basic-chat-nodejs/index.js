/**
 * UniRoute Basic Chat Example (Node.js)
 * 
 * This example demonstrates how to send a chat request to UniRoute API.
 * 
 * Prerequisites:
 * 1. Install dependencies: npm install
 * 2. Set environment variable: export UNIROUTE_API_KEY='ur_your_key_here'
 * 3. Optional: export UNIROUTE_API_URL='http://localhost:8084'
 */

const https = require('https');
const http = require('http');

// Get API key from environment
const apiKey = process.env.UNIROUTE_API_KEY;
if (!apiKey) {
  console.error('❌ Error: UNIROUTE_API_KEY environment variable is required');
  console.error('');
  console.error('Get your API key:');
  console.error('  1. Run: uniroute keys create');
  console.error('  2. Export: export UNIROUTE_API_KEY=\'ur_your_key_here\'');
  process.exit(1);
}

// Get API URL from environment (default: localhost)
const apiUrl = process.env.UNIROUTE_API_URL || 'http://localhost:8084';
const url = new URL(`${apiUrl}/v1/chat`);

// Create chat request
const requestData = {
  model: 'gpt-4',
  messages: [
    {
      role: 'user',
      content: 'Hello! Explain what UniRoute is in one sentence.'
    }
  ],
  temperature: 0.7,
  max_tokens: 100
};

// Determine protocol
const isHttps = url.protocol === 'https:';
const httpModule = isHttps ? https : http;

// Request options
const options = {
  hostname: url.hostname,
  port: url.port || (isHttps ? 443 : 80),
  path: url.pathname,
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${apiKey}`
  }
};

// Send request
const req = httpModule.request(options, (res) => {
  let data = '';

  res.on('data', (chunk) => {
    data += chunk;
  });

  res.on('end', () => {
    if (res.statusCode !== 200) {
      console.error(`❌ Error: Server returned status ${res.statusCode}`);
      console.error(`Response: ${data}`);
      process.exit(1);
    }

    try {
      const chatResponse = JSON.parse(data);

      // Print response
      console.log('✅ Chat Response:');
      console.log('');
      if (chatResponse.choices && chatResponse.choices.length > 0) {
        const message = chatResponse.choices[0].message;
        console.log(`💬 ${message.content}`);
      }
      console.log('');

      if (chatResponse.usage) {
        const usage = chatResponse.usage;
        console.log(`📊 Tokens: ${usage.prompt_tokens} prompt + ${usage.completion_tokens} completion = ${usage.total_tokens} total`);
      }
    } catch (error) {
      console.error(`❌ Error parsing response: ${error.message}`);
      console.error(`Response: ${data}`);
      process.exit(1);
    }
  });
});

req.on('error', (error) => {
  console.error(`❌ Error sending request: ${error.message}`);
  if (error.code === 'ECONNREFUSED') {
    console.error('');
    console.error('💡 Make sure UniRoute server is running:');
    console.error('  make dev');
  }
  process.exit(1);
});

// Send request body
req.write(JSON.stringify(requestData));
req.end();
