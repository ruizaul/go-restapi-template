-- Rollback Migration 000012: Remove secure OTP columns

-- Drop indexes
DROP INDEX IF EXISTS idx_users_otp_locked;
DROP INDEX IF EXISTS idx_users_otp_hash;

-- Remove new columns
ALTER TABLE users
DROP COLUMN IF EXISTS otp_locked_until,
DROP COLUMN IF EXISTS otp_attempts,
DROP COLUMN IF EXISTS otp_hash;

-- Restore comment on otp_code
COMMENT ON COLUMN users.otp_code IS 'Código OTP de 6 dígitos';
