-- Phase 1: Tunnel Database Schema

-- Create tunnels table
CREATE TABLE IF NOT EXISTS tunnels (
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

-- Create tunnel sessions table
CREATE TABLE IF NOT EXISTS tunnel_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tunnel_id UUID REFERENCES tunnels(id) ON DELETE CASCADE,
    client_id VARCHAR(255) NOT NULL,
    server_id VARCHAR(255) NOT NULL,
    connected_at TIMESTAMP DEFAULT NOW(),
    disconnected_at TIMESTAMP,
    last_heartbeat TIMESTAMP,
    status VARCHAR(20) DEFAULT 'connected' -- connected, disconnected
);

-- Create tunnel requests table (for logging)
CREATE TABLE IF NOT EXISTS tunnel_requests (
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

-- Create tunnel tokens table
CREATE TABLE IF NOT EXISTS tunnel_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255),
    expires_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    last_used_at TIMESTAMP,
    is_active BOOLEAN DEFAULT true
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_tunnels_user_id ON tunnels(user_id);
CREATE INDEX IF NOT EXISTS idx_tunnels_subdomain ON tunnels(subdomain);
CREATE INDEX IF NOT EXISTS idx_tunnels_status ON tunnels(status);
CREATE INDEX IF NOT EXISTS idx_tunnels_created_at ON tunnels(created_at);

CREATE INDEX IF NOT EXISTS idx_sessions_tunnel_id ON tunnel_sessions(tunnel_id);
CREATE INDEX IF NOT EXISTS idx_sessions_status ON tunnel_sessions(status);
CREATE INDEX IF NOT EXISTS idx_sessions_connected_at ON tunnel_sessions(connected_at);

CREATE INDEX IF NOT EXISTS idx_requests_tunnel_id ON tunnel_requests(tunnel_id);
CREATE INDEX IF NOT EXISTS idx_requests_created_at ON tunnel_requests(created_at);
CREATE INDEX IF NOT EXISTS idx_requests_request_id ON tunnel_requests(request_id);

CREATE INDEX IF NOT EXISTS idx_tokens_user_id ON tunnel_tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_tokens_hash ON tunnel_tokens(token_hash);
CREATE INDEX IF NOT EXISTS idx_tokens_is_active ON tunnel_tokens(is_active);

