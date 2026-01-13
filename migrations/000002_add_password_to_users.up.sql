-- 000002_add_password_to_users.up.sql
-- Adds password_hash column to users table for authentication

-- Add password_hash column (nullable initially for existing users)
ALTER TABLE users ADD COLUMN IF NOT EXISTS password_hash VARCHAR(255);

-- Create index for faster lookups during authentication
CREATE INDEX IF NOT EXISTS idx_users_email_password ON users(email, password_hash) WHERE deleted_at IS NULL;
