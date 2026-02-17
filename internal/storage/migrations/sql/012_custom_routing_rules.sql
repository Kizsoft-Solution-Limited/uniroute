-- Migration: 012_custom_routing_rules.sql
-- Description: Adds table for storing custom routing rules

-- Create custom_routing_rules table
CREATE TABLE IF NOT EXISTS custom_routing_rules (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    condition_type VARCHAR(50) NOT NULL, -- 'model', 'cost_threshold', 'latency_threshold', 'custom'
    condition_value JSONB, -- Flexible condition data
    provider_name VARCHAR(100) NOT NULL, -- Provider to route to
    priority INTEGER NOT NULL DEFAULT 0, -- Higher priority = checked first
    enabled BOOLEAN NOT NULL DEFAULT true,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    updated_by UUID REFERENCES users(id) ON DELETE SET NULL
);

-- Create index for faster lookups
CREATE INDEX IF NOT EXISTS idx_custom_routing_rules_enabled ON custom_routing_rules(enabled, priority DESC);
CREATE INDEX IF NOT EXISTS idx_custom_routing_rules_provider ON custom_routing_rules(provider_name);

-- Add comments
COMMENT ON TABLE custom_routing_rules IS 'Custom routing rules for CustomStrategy';
COMMENT ON COLUMN custom_routing_rules.condition_type IS 'Type of condition: model, cost_threshold, latency_threshold, custom';
COMMENT ON COLUMN custom_routing_rules.condition_value IS 'JSON condition data (e.g., {"model": "gpt-4"} or {"max_cost": 0.01})';
COMMENT ON COLUMN custom_routing_rules.priority IS 'Higher priority rules are checked first';



