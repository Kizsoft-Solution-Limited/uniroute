-- Migration: Change role column to roles array to support multiple roles
-- Description: Allows users to have multiple roles (e.g., both 'user' and 'admin')
-- Date: 2026-01-05

-- Drop the check constraint first
ALTER TABLE users DROP CONSTRAINT IF EXISTS check_user_role;

-- Add new roles column as array
ALTER TABLE users ADD COLUMN IF NOT EXISTS roles TEXT[] DEFAULT ARRAY['user']::TEXT[];

-- Migrate existing role data to roles array
UPDATE users SET roles = ARRAY[role]::TEXT[] WHERE roles IS NULL OR array_length(roles, 1) IS NULL;

-- Ensure all users have at least 'user' role
UPDATE users SET roles = ARRAY['user']::TEXT[] WHERE array_length(roles, 1) IS NULL;

-- Add check constraint: at least one role, and only 'user' or 'admin' (no subqueries allowed in CHECK)
ALTER TABLE users ADD CONSTRAINT check_user_roles CHECK (
  array_length(roles, 1) > 0 AND
  roles <@ ARRAY['user', 'admin']::TEXT[]
);

-- Create index on roles array for faster queries
CREATE INDEX IF NOT EXISTS idx_users_roles ON users USING GIN(roles);

-- Drop old role column (after migration)
ALTER TABLE users DROP COLUMN IF EXISTS role;

-- Drop old index if it exists
DROP INDEX IF EXISTS idx_users_role;

-- Comment on the column
COMMENT ON COLUMN users.roles IS 'Array of user roles: user (default), admin, or both';

