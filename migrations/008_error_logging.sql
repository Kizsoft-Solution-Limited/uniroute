-- Migration 008: Frontend Error Logging
-- This migration adds support for logging frontend errors in the database

CREATE TABLE IF NOT EXISTS error_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    error_type VARCHAR(50) NOT NULL, -- 'exception', 'message', 'network', 'server'
    message TEXT NOT NULL,
    stack_trace TEXT,
    url VARCHAR(500),
    user_agent TEXT,
    ip_address VARCHAR(45), -- IPv6 compatible
    context JSONB, -- Additional context data
    severity VARCHAR(20) DEFAULT 'error', -- 'error', 'warning', 'info'
    resolved BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Create indexes for efficient querying
CREATE INDEX IF NOT EXISTS idx_error_logs_created_at ON error_logs(created_at);
CREATE INDEX IF NOT EXISTS idx_error_logs_user_id ON error_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_error_logs_error_type ON error_logs(error_type);
CREATE INDEX IF NOT EXISTS idx_error_logs_severity ON error_logs(severity);
CREATE INDEX IF NOT EXISTS idx_error_logs_resolved ON error_logs(resolved);
CREATE INDEX IF NOT EXISTS idx_error_logs_created_at_type ON error_logs(created_at, error_type);

-- Index for querying unresolved errors
CREATE INDEX IF NOT EXISTS idx_error_logs_unresolved ON error_logs(resolved, created_at) WHERE resolved = false;

-- Add comments
COMMENT ON TABLE error_logs IS 'Stores frontend error logs for debugging and monitoring';
COMMENT ON COLUMN error_logs.context IS 'Additional context data (JSON) like request details, user actions, etc.';
COMMENT ON COLUMN error_logs.severity IS 'Error severity: error, warning, info';


