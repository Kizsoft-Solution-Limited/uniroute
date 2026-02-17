# API Reference

UniRoute provides a unified API for all AI providers.

## Base URL

```
https://app.uniroute.co/v1
```

For self-hosted instances:
```
http://localhost:8084/v1
```

## Authentication

All API requests require authentication via API key:

```bash
curl -X POST https://app.uniroute.co/v1/chat \
  -H "Authorization: Bearer ur_your-api-key" \
  -H "Content-Type: application/json" \
  -d '{...}'
```

## Chat API

### Standard Request

```bash
curl -X POST https://app.uniroute.co/v1/chat \
  -H "Authorization: Bearer ur_your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4o",
    "messages": [
      {
        "role": "user",
        "content": "Hello!"
      }
    ]
  }'
```

### Streaming (SSE)

```bash
curl -X POST https://app.uniroute.co/v1/chat/stream \
  -H "Authorization: Bearer ur_your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4o",
    "messages": [{"role": "user", "content": "Hello!"}]
  }'
```

### Streaming (WebSocket)

```javascript
const ws = new WebSocket('wss://app.uniroute.co/v1/chat/ws?token=ur_your-api-key');

ws.onopen = () => {
  ws.send(JSON.stringify({
    model: 'gpt-4o',
    messages: [{ role: 'user', content: 'Hello!' }]
  }));
};

ws.onmessage = (event) => {
  const chunk = JSON.parse(event.data);
  console.log(chunk.content);
};
```

## Supported Models

- **OpenAI**: `gpt-4o`, `gpt-4`, `gpt-3.5-turbo`
- **Anthropic**: `claude-3-opus`, `claude-3-sonnet`, `claude-3-haiku`
- **Google**: `gemini-pro`, `gemini-ultra`
- **Local**: `llama2`, `mistral`, `codellama` (when using Ollama)

## Multimodal Support

UniRoute supports images and audio in chat requests:

```json
{
  "model": "gpt-4o",
  "messages": [
    {
      "role": "user",
      "content": [
        {
          "type": "text",
          "text": "What's in this image?"
        },
        {
          "type": "image_url",
          "image_url": {
            "url": "data:image/jpeg;base64,..."
          }
        }
      ]
    }
  ]
}
```

## Error Handling

All errors follow a consistent format:

```json
{
  "error": "error_type",
  "message": "Human-readable error message"
}
```

Common error codes:
- `401` - Unauthorized (invalid API key)
- `400` - Bad Request (invalid parameters)
- `429` - Rate Limited
- `500` - Server Error

## API Explorer

Try out the API interactively using our Swagger UI:

**[Open API Explorer](http://localhost:8084/swagger)** (opens in new tab)

The API Explorer provides:
- Interactive API testing
- Request/response examples
- Try it out functionality
- Complete endpoint documentation
- Authentication testing

## Next Steps

- [Getting Started](/docs/getting-started) - Make your first request
- [Routing](/docs/routing) - Configure routing strategies
- [CLI Reference](/docs/cli) - Command-line interface
