-- Migration: Add role column to users table
-- Description: Adds a role column to support admin and regular user roles
-- Date: 2026-01-05

-- Add role column with default value 'user'
ALTER TABLE users ADD COLUMN IF NOT EXISTS role VARCHAR(20) NOT NULL DEFAULT 'user';

-- Create index on role for faster queries
CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);

-- Add check constraint to ensure role is either 'user' or 'admin'
ALTER TABLE users ADD CONSTRAINT check_user_role CHECK (role IN ('user', 'admin'));

-- Comment on the column
COMMENT ON COLUMN users.role IS 'User role: user (default) or admin';

