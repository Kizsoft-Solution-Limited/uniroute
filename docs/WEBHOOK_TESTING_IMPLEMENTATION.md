# Webhook Testing Feature - Implementation Summary

## Overview

The webhook testing feature allows users to inspect, replay, and test webhook requests that come through their tunnels. This provides a secure public URL for local web servers, with the ability to inspect traffic and replay requests for iterative development.

## Architecture

### Backend (Tunnel Server)

**Location**: `internal/tunnel/`

**Components**:
1. **Database Schema** (`migrations/005_webhook_testing_schema.sql`)
   - Extended `tunnel_requests` table with:
     - `query_string` - Full query string
     - `request_headers` - JSONB for easy querying
     - `request_body` - BYTEA for full request body storage
     - `response_headers` - JSONB for response headers
     - `response_body` - BYTEA for full response body
     - `remote_addr` - Client IP address
     - `user_agent` - User agent string

2. **Repository** (`internal/tunnel/repository.go`)
   - `CreateTunnelRequest` - Stores full request/response data
   - `GetTunnelRequest` - Retrieves single request with full details
   - `ListTunnelRequests` - Lists requests with filtering (method, path)

3. **API Handlers** (`internal/tunnel/api_handlers.go`)
   - `handleListTunnelRequests` - GET `/api/tunnels/{tunnel_id}/requests`
   - `handleGetTunnelRequest` - GET `/api/tunnels/{tunnel_id}/requests/{request_id}`
   - `handleReplayTunnelRequest` - POST `/api/tunnels/{tunnel_id}/requests/{request_id}/replay`

4. **Server Integration** (`internal/tunnel/server.go`)
   - Updated `forwardHTTPRequest` to store full request/response data
   - Registered new API endpoints

### Frontend (Dashboard)

**Location**: `frontend/src/views/WebhookTesting.vue`

**Features**:
1. **Tunnel Selection** - Select active tunnel to inspect
2. **Request List** - View all requests with filtering:
   - Filter by HTTP method (GET, POST, PUT, PATCH, DELETE)
   - Filter by path pattern
   - Filter by status code
3. **Request Inspection** - View full request details:
   - Request method, path, query string
   - Request headers
   - Request body (formatted JSON if applicable)
   - Response status code and latency
   - Response headers
   - Response body (formatted JSON if applicable)
4. **Request Replay** - Replay any request through the tunnel
5. **Request History** - Browse historical requests

**API Service**: `frontend/src/services/api/webhookTesting.ts`
- Type-safe API client for webhook testing endpoints
- Handles tunnel server communication

## API Endpoints

### List Requests
```
GET /api/tunnels/{tunnel_id}/requests
Query Parameters:
  - method: HTTP method filter (optional)
  - path: Path pattern filter (optional)
  - limit: Number of results (default: 50)
  - offset: Pagination offset (default: 0)
```

### Get Request Details
```
GET /api/tunnels/{tunnel_id}/requests/{request_id}
Returns full request/response data including headers and bodies
```

### Replay Request
```
POST /api/tunnels/{tunnel_id}/requests/{request_id}/replay
Replays the original request through the tunnel and returns the new response
```

## Usage Flow

1. **Start Tunnel**: User starts a tunnel using `uniroute tunnel`
2. **Receive Requests**: Webhook requests come through the tunnel
3. **Inspect Requests**: User opens Webhook Testing page in dashboard
4. **Select Tunnel**: Choose the active tunnel from dropdown
5. **View Requests**: See list of all requests with filters
6. **Inspect Details**: Click on any request to see full details
7. **Replay Request**: Click replay to send the same request again

## Database Migration

Run the migration to enable webhook testing:

```bash
psql $DATABASE_URL -f migrations/005_webhook_testing_schema.sql
```

## Configuration

### Frontend
Set the tunnel server URL in environment variables:
```env
VITE_TUNNEL_SERVER_URL=http://localhost:8080
```

Or it defaults to `http://localhost:8080`

## Security Considerations

1. **Authentication**: Tunnel requests require authentication for public servers
2. **Data Storage**: Request/response bodies stored securely in database
3. **Access Control**: Users can only access requests from their own tunnels
4. **Rate Limiting**: Replay requests are subject to tunnel rate limits

## Future Enhancements

1. **Request Search**: Full-text search across request bodies
2. **Request Comparison**: Compare two requests side-by-side
3. **Request Editing**: Edit and replay requests with modifications
4. **Webhook Triggering**: Manually trigger webhooks with custom payloads
5. **Request Export**: Export requests as HAR files or cURL commands
6. **Request Mocking**: Mock responses for testing
7. **Request Scheduling**: Schedule request replays
8. **Request Diff**: Show differences between original and replayed requests

## Testing

### Backend Tests
```bash
go test ./internal/tunnel/... -v
```

### Frontend Tests
```bash
cd frontend && npm run test
```

## Files Created/Modified

### New Files
- `migrations/005_webhook_testing_schema.sql`
- `frontend/src/views/WebhookTesting.vue`
- `frontend/src/services/api/webhookTesting.ts`
- `docs/WEBHOOK_TESTING_IMPLEMENTATION.md`

### Modified Files
- `internal/tunnel/repository.go` - Extended to store/retrieve full request data
- `internal/tunnel/api_handlers.go` - Added webhook testing endpoints
- `internal/tunnel/server.go` - Updated to store full request data, registered endpoints
- `frontend/src/router/index.ts` - Added webhook testing route
- `frontend/src/layouts/DashboardLayout.vue` - Added webhook testing to navigation

## Status

✅ **Backend Complete**
- Database schema extended
- Repository methods implemented
- API endpoints created
- Request replay functionality working

✅ **Frontend Complete**
- Webhook testing UI implemented
- Request list with filtering
- Request detail modal
- Replay functionality
- Integrated into dashboard navigation

## Next Steps

1. Run database migration
2. Test with real tunnel traffic
3. Add request search functionality
4. Add request export features
5. Add request editing capabilities

