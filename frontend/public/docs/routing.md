# Routing & Strategy

Configure how UniRoute selects AI providers for your requests.

## Overview

UniRoute uses intelligent routing to select the best provider for each request based on:
- Cost optimization
- Latency minimization
- Availability and failover
- Custom rules

## Routing Strategies

### Cost-Based Routing

Select the cheapest available provider:

```bash
curl -X PUT https://api.uniroute.co/auth/routing/strategy \
  -H "Authorization: Bearer your-jwt-token" \
  -H "Content-Type: application/json" \
  -d '{
    "strategy": "cost"
  }'
```

### Latency-Based Routing

Select the fastest provider:

```bash
curl -X PUT https://api.uniroute.co/auth/routing/strategy \
  -H "Authorization: Bearer your-jwt-token" \
  -H "Content-Type: application/json" \
  -d '{
    "strategy": "latency"
  }'
```

### Availability-Based Routing

Select based on provider availability:

```bash
curl -X PUT https://api.uniroute.co/auth/routing/strategy \
  -H "Authorization: Bearer your-jwt-token" \
  -H "Content-Type: application/json" \
  -d '{
    "strategy": "availability"
  }'
```

## Custom Routing Rules

Define custom rules for specific models or use cases:

```bash
curl -X POST https://api.uniroute.co/auth/routing/custom-rules \
  -H "Authorization: Bearer your-jwt-token" \
  -H "Content-Type: application/json" \
  -d '{
    "rules": [
      {
        "model": "gpt-4o",
        "provider": "openai",
        "priority": 1
      },
      {
        "model": "claude-3",
        "provider": "anthropic",
        "priority": 2
      }
    ]
  }'
```

## Failover

UniRoute automatically fails over to backup providers if the primary provider is unavailable.

## Next Steps

- [API Reference](/docs/api) - Make requests with routing
- [Security](/docs/security) - Secure your routing configuration
