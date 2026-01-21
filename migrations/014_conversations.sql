-- Migration: 014_conversations.sql
-- Description: Adds conversations and messages tables for chat persistence

-- Conversations table
CREATE TABLE IF NOT EXISTS conversations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title TEXT, -- Auto-generated or user-defined title
    model TEXT, -- Default model for this conversation
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create index for faster user conversation lookups
CREATE INDEX IF NOT EXISTS idx_conversations_user_id ON conversations(user_id);
CREATE INDEX IF NOT EXISTS idx_conversations_user_id_updated_at ON conversations(user_id, updated_at DESC);

-- Messages table
CREATE TABLE IF NOT EXISTS messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    conversation_id UUID NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    role TEXT NOT NULL CHECK (role IN ('system', 'user', 'assistant')),
    content JSONB NOT NULL, -- Can be string or array of ContentPart (multimodal)
    metadata JSONB, -- Store tokens, cost, provider, latency, etc.
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create index for faster message lookups
CREATE INDEX IF NOT EXISTS idx_messages_conversation_id ON messages(conversation_id);
CREATE INDEX IF NOT EXISTS idx_messages_conversation_id_created_at ON messages(conversation_id, created_at ASC);

-- Add comments
COMMENT ON TABLE conversations IS 'Stores user chat conversations';
COMMENT ON TABLE messages IS 'Stores individual messages within conversations';
COMMENT ON COLUMN messages.content IS 'JSONB: string for text-only, array of ContentPart for multimodal';
COMMENT ON COLUMN messages.metadata IS 'JSONB: tokens, cost, provider, latency, etc.';
