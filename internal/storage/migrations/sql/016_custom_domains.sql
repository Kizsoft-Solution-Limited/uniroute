-- Migration: 016_custom_domains.sql
-- Description: Creates a separate domains table for managing custom domains independently from tunnels

-- Create custom_domains table
CREATE TABLE IF NOT EXISTS custom_domains (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE NOT NULL,
    domain VARCHAR(255) NOT NULL,
    verified BOOLEAN DEFAULT false,
    verification_token VARCHAR(255),
    dns_configured BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(user_id, domain) -- One domain per user (but user can have multiple domains)
);

-- Create indexes for faster lookups
CREATE INDEX IF NOT EXISTS idx_custom_domains_user_id ON custom_domains(user_id);
CREATE INDEX IF NOT EXISTS idx_custom_domains_domain ON custom_domains(domain);
CREATE INDEX IF NOT EXISTS idx_custom_domains_verified ON custom_domains(verified);
CREATE INDEX IF NOT EXISTS idx_custom_domains_dns_configured ON custom_domains(dns_configured);
CREATE INDEX IF NOT EXISTS idx_custom_domains_user_domain ON custom_domains(user_id, domain);

-- Add comment
COMMENT ON TABLE custom_domains IS 'User-managed custom domains that can be assigned to tunnels';
COMMENT ON COLUMN custom_domains.verified IS 'Whether the domain has been verified';
COMMENT ON COLUMN custom_domains.dns_configured IS 'Whether DNS has been properly configured';
