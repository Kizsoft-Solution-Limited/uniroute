-- Migration 004: User Provider Keys (BYOK Support)
-- This migration adds support for users to store their own provider API keys

CREATE TABLE IF NOT EXISTS user_provider_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    provider VARCHAR(50) NOT NULL,  -- 'openai', 'anthropic', 'google'
    api_key_encrypted TEXT NOT NULL,  -- Encrypted provider API key
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(user_id, provider)  -- One key per provider per user
);

-- Indexes for fast lookups
CREATE INDEX IF NOT EXISTS idx_user_provider_keys_user_id ON user_provider_keys(user_id);
CREATE INDEX IF NOT EXISTS idx_user_provider_keys_provider ON user_provider_keys(provider);
CREATE INDEX IF NOT EXISTS idx_user_provider_keys_user_provider ON user_provider_keys(user_id, provider);
CREATE INDEX IF NOT EXISTS idx_user_provider_keys_is_active ON user_provider_keys(is_active);

-- Add comment
COMMENT ON TABLE user_provider_keys IS 'Stores encrypted provider API keys per user (BYOK - Bring Your Own Keys)';
COMMENT ON COLUMN user_provider_keys.provider IS 'Provider name: openai, anthropic, google';
COMMENT ON COLUMN user_provider_keys.api_key_encrypted IS 'Encrypted provider API key (never stored in plaintext)';

