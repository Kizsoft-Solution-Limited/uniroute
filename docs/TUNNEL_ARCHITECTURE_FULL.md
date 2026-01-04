# UniRoute Custom Tunnel - Full Architecture Design

## Table of Contents
1. [System Overview](#system-overview)
2. [Architecture Components](#architecture-components)
3. [Data Flow](#data-flow)
4. [Protocol Design](#protocol-design)
5. [Database Schema](#database-schema)
6. [API Design](#api-design)
7. [Security Architecture](#security-architecture)
8. [Infrastructure Requirements](#infrastructure-requirements)
9. [Scalability Design](#scalability-design)
10. [Implementation Phases](#implementation-phases)

---

## System Overview

### High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                        Internet Users                            │
└────────────────────────────┬────────────────────────────────────┘
                             │ HTTPS
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│                    Load Balancer / CDN                          │
│                  (Cloudflare / AWS ALB)                         │
└────────────────────────────┬────────────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│                    Tunnel Server Cluster                        │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐         │
│  │ Server 1     │  │ Server 2     │  │ Server N     │         │
│  │ :8080        │  │ :8080        │  │ :8080        │         │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘         │
│         │                 │                 │                  │
│  ┌──────▼─────────────────▼─────────────────▼──────┐         │
│  │         Shared State (Redis Cluster)             │         │
│  └──────────────────────────────────────────────────┘         │
└────────────────────────────┬────────────────────────────────────┘
                             │ WebSocket
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│                    Tunnel Clients (Agents)                      │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐         │
│  │ Client 1     │  │ Client 2     │  │ Client N     │         │
│  │ Local:8084   │  │ Local:8080   │  │ Local:3000   │         │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘         │
│         │                 │                 │                  │
│         └─────────────────┼─────────────────┘                 │
│                           │ HTTP                                │
│                           ▼                                     │
│              ┌─────────────────────────┐                       │
│              │   Local UniRoute         │                       │
│              │   Gateway Servers       │                       │
│              └─────────────────────────┘                       │
└─────────────────────────────────────────────────────────────────┘
```

---

## Architecture Components

### 1. Tunnel Server (Control Plane)

**Responsibilities:**
- Accept WebSocket connections from tunnel clients
- Route incoming HTTP requests to appropriate tunnels
- Manage tunnel lifecycle (create, update, delete)
- Handle authentication and authorization
- Provide metrics and monitoring
- Manage subdomain assignment

**Components:**

```
Tunnel Server
├── WebSocket Handler
│   ├── Connection Manager
│   ├── Message Router
│   └── Heartbeat Monitor
├── HTTP Handler
│   ├── Request Router
│   ├── Subdomain Resolver
│   └── Request Forwarder
├── Tunnel Manager
│   ├── Subdomain Allocator
│   ├── Session Manager
│   └── Tunnel Registry
├── Authentication Service
│   ├── Token Validator
│   ├── Rate Limiter
│   └── Access Controller
└── Metrics Collector
    ├── Request Counter
    ├── Latency Tracker
    └── Connection Monitor
```

**Key Features:**
- Multi-tenant support
- Automatic failover
- Load balancing
- Request queuing
- Connection pooling

### 2. Tunnel Client (Agent)

**Responsibilities:**
- Establish WebSocket connection to tunnel server
- Forward incoming requests to local server
- Handle reconnection logic
- Manage local server health
- Report metrics

**Components:**

```
Tunnel Client
├── WebSocket Client
│   ├── Connection Manager
│   ├── Reconnection Logic
│   └── Heartbeat Sender
├── HTTP Proxy
│   ├── Request Forwarder
│   ├── Response Handler
│   └── Error Handler
├── Local Server Monitor
│   ├── Health Checker
│   └── Availability Monitor
└── Configuration Manager
    ├── Settings Loader
    └── Dynamic Config
```

**Key Features:**
- Automatic reconnection
- Request queuing during disconnection
- Local server health monitoring
- Configurable timeouts
- Request/response logging

### 3. Domain & SSL Management

**Responsibilities:**
- Subdomain assignment
- DNS record management
- SSL certificate generation
- Certificate renewal
- Domain validation

**Components:**

```
Domain Manager
├── Subdomain Allocator
│   ├── Random Generator
│   ├── Custom Domain Support
│   └── Collision Detection
├── DNS Manager
│   ├── Record Creator
│   ├── Record Updater
│   └── Propagation Monitor
└── SSL Manager
    ├── Certificate Generator (Let's Encrypt)
    ├── Certificate Renewer
    └── Certificate Distributor
```

**Key Features:**
- Wildcard DNS support
- Automatic SSL via Let's Encrypt
- Certificate auto-renewal
- Custom domain support
- DNS propagation monitoring

### 4. Web Interface

**Responsibilities:**
- Real-time tunnel monitoring
- Request inspection
- Connection statistics
- Tunnel management
- User dashboard

**Components:**

```
Web Interface
├── Frontend (React/Vue)
│   ├── Dashboard
│   ├── Request Inspector
│   ├── Tunnel Manager
│   └── Analytics
├── WebSocket API
│   ├── Real-time Updates
│   └── Event Streaming
└── REST API
    ├── Tunnel CRUD
    ├── Statistics
    └── Configuration
```

**Key Features:**
- Real-time updates
- Request replay
- Connection graphs
- Performance metrics
- Tunnel configuration

### 5. Database Layer

**Responsibilities:**
- Tunnel metadata storage
- User/token management
- Request logging
- Analytics data
- Configuration storage

**Components:**

```
Database Layer
├── PostgreSQL
│   ├── Tunnel Metadata
│   ├── User Accounts
│   ├── Request Logs
│   └── Analytics
└── Redis
    ├── Session Cache
    ├── Subdomain Mapping
    ├── Rate Limiting
    └── Real-time Stats
```

---

## Data Flow

### 1. Tunnel Creation Flow

```
Client                    Server                  Database
  │                         │                        │
  │─── Connect (WS) ────────>│                        │
  │                         │                        │
  │<─── Auth Challenge ──────│                        │
  │                         │                        │
  │─── Auth Token ──────────>│                        │
  │                         │                        │
  │                         │─── Validate Token ────>│
  │                         │<─── User Info ─────────│
  │                         │                        │
  │                         │─── Allocate Subdomain ─>│
  │                         │<─── Subdomain ─────────│
  │                         │                        │
  │                         │─── Create Tunnel ──────>│
  │                         │                        │
  │<─── Tunnel Info ────────│                        │
  │                         │                        │
  │─── Ready ───────────────>│                        │
```

### 2. Request Forwarding Flow

```
Internet User              Server                  Client              Local Server
     │                       │                       │                    │
     │─── HTTP Request ──────>│                       │                    │
     │                       │                       │                    │
     │                       │─── Resolve Subdomain ─>│                    │
     │                       │<─── Tunnel Info ──────│                    │
     │                       │                       │                    │
     │                       │─── Forward Request ───>│                    │
     │                       │    (via WebSocket)     │                    │
     │                       │                       │                    │
     │                       │                       │─── HTTP Request ───>│
     │                       │                       │                    │
     │                       │                       │<─── HTTP Response ──│
     │                       │                       │                    │
     │                       │<─── Response ────────│                    │
     │                       │    (via WebSocket)    │                    │
     │                       │                       │                    │
     │<─── HTTP Response ─────│                       │                    │
```

### 3. Reconnection Flow

```
Client                    Server                  Database
  │                         │                        │
  │─── Connection Lost       │                        │
  │                         │                        │
  │─── Attempt Reconnect ───>│                        │
  │                         │                        │
  │<─── Connection Refused ──│                        │
  │                         │                        │
  │─── Wait (Exponential)    │                        │
  │                         │                        │
  │─── Reconnect ───────────>│                        │
  │                         │                        │
  │                         │─── Validate Session ──>│
  │                         │<─── Session Valid ──────│
  │                         │                        │
  │<─── Reconnect Success ───│                        │
  │                         │                        │
  │─── Restore State ───────>│                        │
```

---

## Protocol Design

### WebSocket Message Types

#### 1. Control Messages

```json
// Client -> Server: Initialize Tunnel
{
  "type": "init",
  "version": "1.0",
  "local_url": "http://localhost:8084",
  "token": "auth_token_here",
  "metadata": {
    "name": "My Tunnel",
    "region": "us"
  }
}

// Server -> Client: Tunnel Created
{
  "type": "tunnel_created",
  "tunnel_id": "abc123",
  "subdomain": "abc123",
  "public_url": "https://abc123.uniroute.dev",
  "status": "active"
}

// Client/Server: Heartbeat
{
  "type": "ping",
  "timestamp": 1234567890
}

{
  "type": "pong",
  "timestamp": 1234567890
}
```

#### 2. Request/Response Messages

```json
// Server -> Client: Forward HTTP Request
{
  "type": "http_request",
  "request_id": "req_123",
  "method": "POST",
  "path": "/v1/chat",
  "headers": {
    "Content-Type": "application/json",
    "Authorization": "Bearer ..."
  },
  "body": "base64_encoded_body",
  "query": "param=value"
}

// Client -> Server: HTTP Response
{
  "type": "http_response",
  "request_id": "req_123",
  "status": 200,
  "headers": {
    "Content-Type": "application/json"
  },
  "body": "base64_encoded_body"
}

// Client -> Server: Request Error
{
  "type": "http_error",
  "request_id": "req_123",
  "error": "connection_refused",
  "message": "Local server not responding"
}
```

#### 3. Management Messages

```json
// Client -> Server: Update Tunnel
{
  "type": "update_tunnel",
  "tunnel_id": "abc123",
  "local_url": "http://localhost:8085"
}

// Server -> Client: Tunnel Status
{
  "type": "tunnel_status",
  "tunnel_id": "abc123",
  "status": "active",
  "stats": {
    "requests": 1000,
    "latency_ms": 45,
    "uptime_seconds": 3600
  }
}
```

### Message Flow Protocol

```
┌─────────────────────────────────────────────────────────┐
│                    Connection Phase                     │
├─────────────────────────────────────────────────────────┤
│ 1. Client connects via WebSocket                        │
│ 2. Server sends auth challenge                          │
│ 3. Client sends auth token                              │
│ 4. Server validates and creates tunnel                  │
│ 5. Server sends tunnel info                             │
│ 6. Client confirms ready                                │
└─────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────┐
│                    Request Phase                        │
├─────────────────────────────────────────────────────────┤
│ 1. Internet user sends HTTP request                     │
│ 2. Server resolves subdomain to tunnel                  │
│ 3. Server forwards request via WebSocket                │
│ 4. Client receives and forwards to local server        │
│ 5. Client receives response from local server           │
│ 6. Client sends response via WebSocket                   │
│ 7. Server sends response to internet user               │
└─────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────┐
│                  Maintenance Phase                      │
├─────────────────────────────────────────────────────────┤
│ 1. Periodic heartbeat (every 30s)                       │
│ 2. Connection health monitoring                          │
│ 3. Automatic reconnection on failure                     │
│ 4. State synchronization                                │
└─────────────────────────────────────────────────────────┘
```

---

## Database Schema

### PostgreSQL Tables

#### 1. Users Table
```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    is_active BOOLEAN DEFAULT true
);
```

#### 2. Tunnels Table
```sql
CREATE TABLE tunnels (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    subdomain VARCHAR(63) UNIQUE NOT NULL,
    custom_domain VARCHAR(255),
    local_url VARCHAR(255) NOT NULL,
    public_url VARCHAR(255) NOT NULL,
    status VARCHAR(20) DEFAULT 'active', -- active, paused, deleted
    region VARCHAR(10) DEFAULT 'us',
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    last_active_at TIMESTAMP,
    request_count BIGINT DEFAULT 0,
    metadata JSONB
);

CREATE INDEX idx_tunnels_user_id ON tunnels(user_id);
CREATE INDEX idx_tunnels_subdomain ON tunnels(subdomain);
CREATE INDEX idx_tunnels_status ON tunnels(status);
```

#### 3. Tunnel Sessions Table
```sql
CREATE TABLE tunnel_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tunnel_id UUID REFERENCES tunnels(id) ON DELETE CASCADE,
    client_id VARCHAR(255) NOT NULL,
    server_id VARCHAR(255) NOT NULL,
    connected_at TIMESTAMP DEFAULT NOW(),
    disconnected_at TIMESTAMP,
    last_heartbeat TIMESTAMP,
    status VARCHAR(20) DEFAULT 'connected'
);

CREATE INDEX idx_sessions_tunnel_id ON tunnel_sessions(tunnel_id);
CREATE INDEX idx_sessions_status ON tunnel_sessions(status);
```

#### 4. Tunnel Requests Table
```sql
CREATE TABLE tunnel_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tunnel_id UUID REFERENCES tunnels(id) ON DELETE CASCADE,
    request_id VARCHAR(255) NOT NULL,
    method VARCHAR(10) NOT NULL,
    path VARCHAR(500) NOT NULL,
    status_code INTEGER,
    latency_ms INTEGER,
    request_size INTEGER,
    response_size INTEGER,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_requests_tunnel_id ON tunnel_requests(tunnel_id);
CREATE INDEX idx_requests_created_at ON tunnel_requests(created_at);
```

#### 5. Tunnel Tokens Table
```sql
CREATE TABLE tunnel_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255),
    expires_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    last_used_at TIMESTAMP,
    is_active BOOLEAN DEFAULT true
);

CREATE INDEX idx_tokens_user_id ON tunnel_tokens(user_id);
CREATE INDEX idx_tokens_hash ON tunnel_tokens(token_hash);
```

### Redis Keys

```
# Subdomain to Tunnel ID mapping
tunnel:subdomain:{subdomain} -> tunnel_id

# Active tunnel sessions
tunnel:session:{tunnel_id} -> {session_data}

# Tunnel statistics
tunnel:stats:{tunnel_id} -> {stats_json}

# Rate limiting
tunnel:ratelimit:{user_id} -> {count}

# Request queue (during disconnection)
tunnel:queue:{tunnel_id} -> [request_ids]
```

---

## API Design

### REST API Endpoints

#### Tunnel Management

```
POST   /api/v1/tunnels                    # Create tunnel
GET    /api/v1/tunnels                    # List tunnels
GET    /api/v1/tunnels/:id                # Get tunnel
PATCH  /api/v1/tunnels/:id                # Update tunnel
DELETE /api/v1/tunnels/:id                # Delete tunnel
POST   /api/v1/tunnels/:id/pause          # Pause tunnel
POST   /api/v1/tunnels/:id/resume         # Resume tunnel
```

#### Statistics

```
GET    /api/v1/tunnels/:id/stats          # Get tunnel statistics
GET    /api/v1/tunnels/:id/requests       # Get request history
GET    /api/v1/tunnels/:id/metrics       # Get metrics
```

#### Web Interface

```
GET    /web                                # Web dashboard
GET    /web/tunnels/:id                    # Tunnel detail page
GET    /web/tunnels/:id/inspector          # Request inspector
WS     /ws/tunnels/:id                     # Real-time updates
```

### WebSocket Endpoints

```
WS     /tunnel                             # Tunnel connection
WS     /ws/tunnels/:id                     # Tunnel-specific updates
WS     /ws/user                            # User-wide updates
```

---

## Security Architecture

### Authentication & Authorization

1. **Token-Based Auth**
   - JWT tokens for API access
   - Long-lived tokens for tunnel connections
   - Token rotation support

2. **Tunnel Authentication**
   - Pre-shared tokens
   - User-based authentication
   - IP whitelisting (optional)

3. **Rate Limiting**
   - Per-user limits
   - Per-tunnel limits
   - Global rate limits

### Security Features

1. **TLS/SSL**
   - End-to-end encryption
   - Certificate pinning
   - TLS 1.3 support

2. **Request Validation**
   - Header validation
   - Size limits
   - Timeout limits

3. **Abuse Prevention**
   - DDoS protection
   - Request throttling
   - Suspicious activity detection

---

## Infrastructure Requirements

### Server Requirements

**Minimum (Development):**
- 2 CPU cores
- 4GB RAM
- 20GB storage
- 100Mbps network

**Recommended (Production):**
- 4+ CPU cores
- 8GB+ RAM
- 100GB+ storage
- 1Gbps+ network

### Services Required

1. **PostgreSQL**
   - Version 15+
   - Replication for HA
   - Automated backups

2. **Redis**
   - Version 7+
   - Cluster mode for scale
   - Persistence enabled

3. **Load Balancer**
   - Cloudflare / AWS ALB
   - SSL termination
   - Health checks

4. **DNS Provider**
   - Cloudflare / Route53
   - API access for automation
   - Wildcard support

5. **SSL Certificate**
   - Let's Encrypt
   - Auto-renewal
   - Wildcard certificates

### Deployment Architecture

```
┌─────────────────────────────────────────┐
│         Cloudflare / CDN                │
│      (DNS + SSL + DDoS Protection)      │
└───────────────┬─────────────────────────┘
                │
                ▼
┌─────────────────────────────────────────┐
│         Load Balancer                   │
│      (Health Checks + SSL)              │
└───────────────┬─────────────────────────┘
                │
        ┌───────┴───────┐
        │               │
        ▼               ▼
┌──────────┐      ┌──────────┐
│ Server 1 │      │ Server 2 │
│ :8080    │      │ :8080    │
└────┬─────┘      └────┬─────┘
     │                 │
     └────────┬────────┘
              │
        ┌─────┴─────┐
        │           │
        ▼           ▼
┌──────────┐  ┌──────────┐
│PostgreSQL│  │  Redis   │
│ Cluster  │  │ Cluster  │
└──────────┘  └──────────┘
```

---

## Scalability Design

### Horizontal Scaling

1. **Stateless Servers**
   - All tunnel state in Redis
   - Session affinity via subdomain
   - Load balancer distribution

2. **Database Scaling**
   - Read replicas
   - Connection pooling
   - Query optimization

3. **Redis Scaling**
   - Redis Cluster
   - Sharding by tunnel ID
   - Replication

### Performance Optimization

1. **Connection Pooling**
   - WebSocket connection reuse
   - HTTP client pooling
   - Database connection pooling

2. **Caching**
   - Subdomain resolution cache
   - Tunnel metadata cache
   - Statistics cache

3. **Async Processing**
   - Request queuing
   - Background jobs
   - Event-driven architecture

---

## Implementation Phases

### Phase 1: Core Infrastructure (2-3 weeks)
- [ ] Basic WebSocket server
- [ ] Tunnel client
- [ ] Request forwarding
- [ ] Database schema
- [ ] Basic authentication

### Phase 2: Domain & SSL (1-2 weeks)
- [ ] Subdomain allocation
- [ ] DNS management
- [ ] SSL certificate generation
- [ ] Certificate renewal

### Phase 3: Production Features (2-3 weeks)
- [ ] Web interface
- [ ] Request inspection
- [ ] Statistics and metrics
- [ ] Reconnection logic
- [ ] Error handling

### Phase 4: Scale & Polish (2-4 weeks)
- [ ] Load balancing
- [ ] High availability
- [ ] Performance optimization
- [ ] Security hardening
- [ ] Monitoring and alerting

### Phase 5: Advanced Features (2-3 weeks)
- [ ] Custom domains
- [ ] Request replay
- [ ] Analytics dashboard
- [ ] API improvements
- [ ] Documentation

**Total Estimated Time: 9-15 weeks**

---

## Technology Stack

### Backend
- **Go** - Server and client
- **PostgreSQL** - Primary database
- **Redis** - Caching and sessions
- **Gorilla WebSocket** - WebSocket library
- **Let's Encrypt** - SSL certificates

### Frontend (Web Interface)
- **React** or **Vue.js** - UI framework
- **WebSocket Client** - Real-time updates
- **Chart.js** - Statistics visualization

### Infrastructure
- **Docker** - Containerization
- **Kubernetes** (optional) - Orchestration
- **Cloudflare** - DNS and CDN
- **Let's Encrypt** - SSL certificates

---

## Next Steps

1. **Review and Approve Architecture**
2. **Set Up Development Environment**
3. **Implement Phase 1 (Core Infrastructure)**
4. **Iterate Based on Feedback**
5. **Proceed to Next Phases**

This architecture provides a solid foundation for building a production-ready custom tunnel solution for UniRoute.

