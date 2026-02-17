-- Migration: 015_performance_indexes.sql
-- Description: Add performance indexes for common query patterns

-- Composite index for user-specific time-based queries (requests table)
-- This optimizes queries like: WHERE user_id = X AND created_at >= Y AND created_at <= Z
CREATE INDEX IF NOT EXISTS idx_requests_user_id_created_at ON requests(user_id, created_at DESC);

-- Composite index for provider + time queries (already exists but ensure it's optimal)
-- idx_requests_created_at_provider already exists from 002_analytics_schema.sql

-- Composite index for custom routing rules (user_id + enabled)
-- Optimizes: WHERE enabled = true AND (user_id = X OR user_id IS NULL)
CREATE INDEX IF NOT EXISTS idx_custom_routing_rules_user_enabled ON custom_routing_rules(user_id, enabled) WHERE enabled = true;

-- Index for tunnels by user_id and status (common query pattern)
CREATE INDEX IF NOT EXISTS idx_tunnels_user_id_status ON tunnels(user_id, status);

-- Index for error logs by user_id and created_at (for time-based filtering)
CREATE INDEX IF NOT EXISTS idx_error_logs_user_id_created_at ON error_logs(user_id, created_at DESC);

-- Index for conversations by user_id and created_at (for pagination)
CREATE INDEX IF NOT EXISTS idx_conversations_user_id_created_at ON conversations(user_id, created_at DESC);

-- Index for messages by conversation_id and created_at (for message retrieval)
CREATE INDEX IF NOT EXISTS idx_messages_conversation_id_created_at ON messages(conversation_id, created_at ASC);

-- Partial index for active API keys (only index active keys for faster lookups)
CREATE INDEX IF NOT EXISTS idx_api_keys_active_lookup ON api_keys(lookup_hash) WHERE is_active = true;

-- Partial index for active provider keys
CREATE INDEX IF NOT EXISTS idx_provider_keys_user_active ON provider_keys(user_id, provider) WHERE is_active = true;
