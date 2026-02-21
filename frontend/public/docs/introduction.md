# Introduction

Welcome to the UniRoute documentation!

UniRoute is a fast, secure, and open-source unified gateway platform that allows you to route, secure, and manage traffic to any LLM (cloud or local) with one unified API. It includes built-in tunneling to expose local services to the internet, designed for developers who value performance, control, and flexibility.

## Why UniRoute?

- **Open Source**: The core of UniRoute is open source, giving you the freedom to inspect, modify, and self-host the entire stack.
- **Fast & Secure**: Built with performance in mind, UniRoute ensures your tunnels and API requests are stable and secure.
- **Developer Friendly**: A powerful CLI and intuitive dashboard make managing your tunnels and API routing a breeze.
- **Multi-Provider Support**: Seamlessly switch between OpenAI, Anthropic, Google, Local LLMs, and more with a single API.
- **Intelligent Routing**: Automatic load balancing, failover, and cost-based routing to optimize your AI infrastructure.
- **Self-Hostable**: Deploy anywhere with full control, or use our managed service.

## How it Works

UniRoute creates a secure gateway between your applications and AI providers. When you make a request to UniRoute:

1. **Request Reception**: Your application sends a request to UniRoute's unified API endpoint
2. **Intelligent Routing**: UniRoute analyzes the request and routes it to the best available provider based on your routing strategy
3. **Provider Communication**: The request is forwarded to the selected AI provider (OpenAI, Anthropic, Google, Local LLM, etc.)
4. **Response Handling**: The response is streamed back through UniRoute to your application
5. **Analytics & Monitoring**: All requests are logged for analytics, cost tracking, and performance monitoring

## Key Features

### Unified API
One endpoint for all AI providers. Switch providers without changing your code.

### Tunneling
Expose your local services to the internet with built-in tunneling.

### Security
Enterprise-grade security with API keys, JWT authentication, rate limiting, and IP whitelisting.

### Monitoring
Track usage, costs, and performance metrics across all your AI requests.

### Custom Routing
Define custom routing rules and strategies to optimize cost, latency, and availability.

## Authentication

UniRoute supports multiple authentication methods for the CLI:

### Email/Password Login
Standard login with session expiration:
```bash
uniroute auth login
# Or with email flag
uniroute auth login --email user@example.com
```

### API Key Login (Recommended for Automation)
API keys provide longer sessions without expiration, making them ideal for automation and CI/CD:
```bash
# Login with API key
uniroute auth login --api-key ur_xxxxxxxxxxxxx
# or using short flag
uniroute auth login -k ur_xxxxxxxxxxxxx
# use --live for hosted, --local for local server (default when unspecified: env → saved → hosted)
```

**Benefits of API Key Login:**
- ✅ No expiration (unlike JWT tokens)
- ✅ Perfect for automation and scripts
- ✅ Longer sessions for CI/CD pipelines
- ✅ Same authentication as email/password

## Support the project

If UniRoute is useful to you, consider [**Donate**](https://buy.polar.sh/polar_cl_h5uF0bHhXXF6EO8Mx3tVP1Ry1G4wNWn4V8phg3rStVs) to help keep development going.

## Next Steps

- [Installation](/docs/installation) - Get UniRoute up and running
- [Getting Started](/docs/getting-started) - Create your first tunnel or API request
- [Authentication](/docs/authentication) - Detailed authentication guide
- [Tunnels](/docs/tunnels) - Learn about tunneling features
