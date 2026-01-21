-- Migration: 013_user_custom_routing_rules.sql
-- Description: Adds user_id column to custom_routing_rules to support user-specific custom rules

-- Add user_id column (nullable - NULL means global/admin rule)
ALTER TABLE custom_routing_rules 
ADD COLUMN IF NOT EXISTS user_id UUID REFERENCES users(id) ON DELETE CASCADE;

-- Create index for faster user-specific lookups
CREATE INDEX IF NOT EXISTS idx_custom_routing_rules_user_id ON custom_routing_rules(user_id) WHERE user_id IS NOT NULL;

-- Create index for global rules (user_id IS NULL)
CREATE INDEX IF NOT EXISTS idx_custom_routing_rules_global ON custom_routing_rules(enabled, priority DESC) WHERE user_id IS NULL;

-- Add comment
COMMENT ON COLUMN custom_routing_rules.user_id IS 'User ID for user-specific rules. NULL means global/admin rule.';
