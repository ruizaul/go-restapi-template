-- Make email and password_hash nullable to support pending user registration
-- Users in 'pending' state only have phone + OTP, no email/password yet

-- Drop NOT NULL constraint from email
ALTER TABLE users ALTER COLUMN email DROP NOT NULL;

-- Drop NOT NULL constraint from password_hash
ALTER TABLE users ALTER COLUMN password_hash DROP NOT NULL;

-- Add check constraint: if account_status is 'active', email and password_hash must be present
ALTER TABLE users ADD CONSTRAINT check_active_user_credentials
    CHECK (
        account_status != 'active' OR
        (email IS NOT NULL AND password_hash IS NOT NULL)
    );

-- Update existing users with empty strings to NULL (if any)
UPDATE users SET email = NULL WHERE email = '';
UPDATE users SET password_hash = NULL WHERE password_hash = '';

-- Ensure phone is unique and not null for pending users
ALTER TABLE users ALTER COLUMN phone SET NOT NULL;

-- Success marker
SELECT 'Migration 000003: email and password_hash are now nullable for pending users' as status;
