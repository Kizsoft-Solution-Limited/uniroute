-- Add protocol column to tunnels table
-- Protocol: http, tcp, tls, udp

ALTER TABLE tunnels ADD COLUMN IF NOT EXISTS protocol VARCHAR(10) DEFAULT 'http';

-- Create index for faster protocol-based queries
CREATE INDEX IF NOT EXISTS idx_tunnels_user_id_protocol ON tunnels(user_id, protocol);
CREATE INDEX IF NOT EXISTS idx_tunnels_user_id_status_protocol ON tunnels(user_id, status, protocol);

-- Update existing tunnels to infer protocol from local_url
-- HTTP: starts with http:// or https://
-- TCP/TLS/UDP: format is host:port
UPDATE tunnels 
SET protocol = CASE 
    WHEN local_url LIKE 'http://%' OR local_url LIKE 'https://%' THEN 'http'
    WHEN local_url LIKE '%:%' THEN 'tcp'  -- Default to tcp for host:port format
    ELSE 'http'
END
WHERE protocol IS NULL OR protocol = 'http';
