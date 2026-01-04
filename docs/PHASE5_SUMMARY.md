# Phase 5: Monitoring & Analytics - Implementation Summary

## Overview

Phase 5 adds comprehensive monitoring and analytics capabilities to UniRoute, including Prometheus metrics, usage tracking, cost calculation, and analytics API endpoints.

## ✅ Completed Features

### 1. Prometheus Metrics
- **Location**: `internal/monitoring/metrics.go`
- **Metrics**:
  - `uniroute_requests_total` - Total requests by provider, model, status
  - `uniroute_request_duration_seconds` - Request duration histogram
  - `uniroute_tokens_total` - Token usage by provider, model, type
  - `uniroute_cost_total` - Total cost by provider, model
  - `uniroute_provider_health` - Provider health status
  - `uniroute_rate_limit_hits_total` - Rate limit hits
- **Endpoint**: `GET /metrics` (no auth required)

### 2. Usage Tracking
- **Location**: `internal/storage/request_repository.go`
- **Database Table**: `requests` (migration: `migrations/002_analytics_schema.sql`)
- **Tracks**:
  - API key ID and user ID
  - Provider and model used
  - Token usage (input, output, total)
  - Cost per request
  - Latency
  - Status code and errors
  - Timestamp

### 3. Cost Calculation Tracking
- **Location**: `internal/api/handlers/chat.go`
- **Features**:
  - Automatic cost calculation per request
  - Cost stored in database
  - Cost included in Prometheus metrics
  - Cost aggregation in analytics API

### 4. Analytics API
- **Location**: `internal/api/handlers/analytics.go`
- **Endpoints**:
  - `GET /v1/analytics/usage` - Get usage statistics
    - Query params: `start_time`, `end_time` (RFC3339 format)
    - Returns: Total requests, tokens, cost, latency, breakdowns
  - `GET /v1/analytics/requests` - Get paginated request history
    - Query params: `limit` (default: 50, max: 100), `offset` (default: 0)
    - Returns: List of requests with details

### 5. Automatic Request Tracking
- **Location**: `internal/api/handlers/chat.go`
- **Features**:
  - All chat requests automatically tracked
  - Async tracking (doesn't block response)
  - Tracks both successful and failed requests
  - Includes Prometheus metrics

## Architecture

### Database Schema

```sql
CREATE TABLE requests (
    id UUID PRIMARY KEY,
    api_key_id UUID REFERENCES api_keys(id),
    user_id UUID REFERENCES users(id),
    provider VARCHAR(50) NOT NULL,
    model VARCHAR(100) NOT NULL,
    request_type VARCHAR(50) DEFAULT 'chat',
    input_tokens INTEGER DEFAULT 0,
    output_tokens INTEGER DEFAULT 0,
    total_tokens INTEGER DEFAULT 0,
    cost DECIMAL(10, 6) DEFAULT 0,
    latency_ms INTEGER DEFAULT 0,
    status_code INTEGER DEFAULT 200,
    error_message TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);
```

### Metrics Flow

1. **Request arrives** → Chat handler
2. **Request processed** → Router routes to provider
3. **Response received** → Metrics recorded:
   - Prometheus metrics (synchronous)
   - Database tracking (asynchronous)
4. **Analytics available** → Via API endpoints

## Usage Examples

### Get Usage Statistics

```bash
# Last 30 days (default)
curl http://localhost:8084/v1/analytics/usage \
  -H "Authorization: Bearer API_KEY"

# Custom time range
curl "http://localhost:8084/v1/analytics/usage?start_time=2024-01-01T00:00:00Z&end_time=2024-01-31T23:59:59Z" \
  -H "Authorization: Bearer API_KEY"
```

**Response:**
```json
{
  "period": {
    "start": "2024-01-01T00:00:00Z",
    "end": "2024-01-31T23:59:59Z"
  },
  "total_requests": 1000,
  "total_tokens": 500000,
  "total_cost": 12.50,
  "average_latency_ms": 150.5,
  "requests_by_provider": {
    "openai": 600,
    "anthropic": 300,
    "local": 100
  },
  "requests_by_model": {
    "gpt-4": 400,
    "claude-3": 300,
    "llama2": 100
  },
  "cost_by_provider": {
    "openai": 10.00,
    "anthropic": 2.50,
    "local": 0.00
  }
}
```

### Get Request History

```bash
# First page (50 requests)
curl http://localhost:8084/v1/analytics/requests \
  -H "Authorization: Bearer API_KEY"

# With pagination
curl "http://localhost:8084/v1/analytics/requests?limit=20&offset=40" \
  -H "Authorization: Bearer API_KEY"
```

**Response:**
```json
{
  "requests": [
    {
      "id": "uuid",
      "provider": "openai",
      "model": "gpt-4",
      "input_tokens": 100,
      "output_tokens": 50,
      "total_tokens": 150,
      "cost": 0.005,
      "latency_ms": 200,
      "status_code": 200,
      "created_at": "2024-01-15T10:30:00Z"
    }
  ],
  "limit": 50,
  "offset": 0,
  "count": 50
}
```

### Prometheus Metrics

```bash
# Scrape metrics
curl http://localhost:8084/metrics
```

**Example metrics:**
```
# HELP uniroute_requests_total Total number of requests
# TYPE uniroute_requests_total counter
uniroute_requests_total{model="gpt-4",provider="openai",status="success"} 100

# HELP uniroute_request_duration_seconds Request duration in seconds
# TYPE uniroute_request_duration_seconds histogram
uniroute_request_duration_seconds_bucket{model="gpt-4",provider="openai",le="0.1"} 50
uniroute_request_duration_seconds_bucket{model="gpt-4",provider="openai",le="0.5"} 100

# HELP uniroute_tokens_total Total number of tokens processed
# TYPE uniroute_tokens_total counter
uniroute_tokens_total{model="gpt-4",provider="openai",type="input"} 10000
uniroute_tokens_total{model="gpt-4",provider="openai",type="output"} 5000

# HELP uniroute_cost_total Total cost in USD
# TYPE uniroute_cost_total counter
uniroute_cost_total{model="gpt-4",provider="openai"} 0.5
```

## Files Created

- `migrations/002_analytics_schema.sql` - Database schema for analytics
- `internal/monitoring/metrics.go` - Prometheus metrics definitions
- `internal/storage/request_repository.go` - Request tracking repository
- `internal/api/handlers/analytics.go` - Analytics API handlers
- `internal/api/handlers/metrics.go` - Prometheus metrics endpoint
- `internal/monitoring/metrics_test.go` - Metrics tests

## Files Modified

- `internal/api/handlers/chat.go` - Added request tracking and metrics
- `internal/api/router.go` - Added analytics endpoints and metrics endpoint
- `cmd/gateway/main.go` - Initialize request repository

## Testing

### Unit Tests
- ✅ Metrics tests (5 tests) - All passing

**Total: 5+ new tests - All passing ✅**

## Configuration

### Database Required

Phase 5 requires PostgreSQL to be configured:
- Set `DATABASE_URL` environment variable
- Run migration: `migrations/002_analytics_schema.sql`

### Optional

- Prometheus scraping: Configure Prometheus to scrape `/metrics`
- Analytics access: Requires API key authentication

## Next Steps

- [ ] Add Grafana dashboards
- [ ] Add cost alerts
- [ ] Add usage quotas
- [ ] Add real-time analytics
- [ ] Add export functionality (CSV, JSON)

## Notes

- Request tracking is asynchronous (doesn't block responses)
- Metrics are recorded synchronously (for accuracy)
- Analytics API requires database connection
- Prometheus metrics available without database
- All tracking respects user/API key isolation

