-- Migration: 011_routing_strategy_persistence.sql
-- Description: Adds persistence for routing strategy configuration (Hybrid: Global + Per-User)

-- Create system_settings table for admin-controlled defaults
CREATE TABLE IF NOT EXISTS system_settings (
    id SERIAL PRIMARY KEY,
    key VARCHAR(255) UNIQUE NOT NULL,
    value TEXT NOT NULL,
    description TEXT,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_by UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Insert default routing strategy
INSERT INTO system_settings (key, value, description)
VALUES 
    ('default_routing_strategy', 'model', 'Default routing strategy for all users: model, cost, latency, balanced, or custom'),
    ('routing_strategy_locked', 'false', 'If true, users cannot override the default routing strategy')
ON CONFLICT (key) DO NOTHING;

-- Create index for faster lookups
CREATE INDEX IF NOT EXISTS idx_system_settings_key ON system_settings(key);

-- Add routing_strategy column to users table (nullable - NULL means use default)
ALTER TABLE users ADD COLUMN IF NOT EXISTS routing_strategy VARCHAR(50) NULL;

-- Add check constraint for valid strategy values (PostgreSQL does not support ADD CONSTRAINT IF NOT EXISTS)
ALTER TABLE users DROP CONSTRAINT IF EXISTS check_user_routing_strategy;
ALTER TABLE users ADD CONSTRAINT check_user_routing_strategy
    CHECK (routing_strategy IS NULL OR routing_strategy IN ('model', 'cost', 'latency', 'balanced', 'custom'));

-- Add comments
COMMENT ON TABLE system_settings IS 'System-wide configuration settings';
COMMENT ON COLUMN system_settings.key IS 'Setting key (e.g., default_routing_strategy)';
COMMENT ON COLUMN system_settings.value IS 'Setting value (JSON or plain text)';
COMMENT ON COLUMN system_settings.updated_by IS 'User ID who last updated this setting (NULL for system)';
COMMENT ON COLUMN users.routing_strategy IS 'User-specific routing strategy override (NULL = use default). Valid values: model, cost, latency, balanced, custom';

