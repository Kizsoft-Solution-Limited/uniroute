-- Webhook Testing Feature - Extend tunnel_requests table

-- Add columns for full request/response storage
ALTER TABLE tunnel_requests 
ADD COLUMN IF NOT EXISTS query_string TEXT,
ADD COLUMN IF NOT EXISTS request_headers JSONB,
ADD COLUMN IF NOT EXISTS request_body BYTEA,
ADD COLUMN IF NOT EXISTS response_headers JSONB,
ADD COLUMN IF NOT EXISTS response_body BYTEA,
ADD COLUMN IF NOT EXISTS remote_addr VARCHAR(255),
ADD COLUMN IF NOT EXISTS user_agent TEXT;

-- Create index for faster querying
CREATE INDEX IF NOT EXISTS idx_requests_method ON tunnel_requests(method);
CREATE INDEX IF NOT EXISTS idx_requests_path_pattern ON tunnel_requests(path text_pattern_ops);
CREATE INDEX IF NOT EXISTS idx_requests_status_code ON tunnel_requests(status_code);

-- Add comment
COMMENT ON COLUMN tunnel_requests.request_body IS 'Full request body for webhook inspection and replay';
COMMENT ON COLUMN tunnel_requests.response_body IS 'Full response body for webhook inspection';
COMMENT ON COLUMN tunnel_requests.request_headers IS 'Request headers as JSON for easy querying';
COMMENT ON COLUMN tunnel_requests.response_headers IS 'Response headers as JSON';

