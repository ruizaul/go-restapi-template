-- Revert email and password_hash nullable changes
-- This down migration restores NOT NULL constraints

-- First, ensure all users have email and password_hash populated
-- Set temporary values for any users without them
UPDATE users SET email = 'pending_' || id::text || '@tacoshare.temp' WHERE email IS NULL;
UPDATE users SET password_hash = '' WHERE password_hash IS NULL;

-- Drop the check constraint
ALTER TABLE users DROP CONSTRAINT IF EXISTS check_active_user_credentials;

-- Restore NOT NULL constraint on email
ALTER TABLE users ALTER COLUMN email SET NOT NULL;

-- Restore NOT NULL constraint on password_hash
ALTER TABLE users ALTER COLUMN password_hash SET NOT NULL;

-- Make phone nullable again (original state)
ALTER TABLE users ALTER COLUMN phone DROP NOT NULL;

-- Success marker
SELECT 'Migration 000003 reverted: email and password_hash are NOT NULL again' as status;
