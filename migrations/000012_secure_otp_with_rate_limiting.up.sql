-- Migration 000012: Secure OTP with Hash + Rate Limiting
-- Replaces plaintext OTP storage with SHA-256 hash and implements rate limiting

-- Step 1: Add new secure OTP columns
ALTER TABLE users
ADD COLUMN otp_hash VARCHAR(64),           -- SHA-256 hash of OTP (64 hex chars)
ADD COLUMN otp_attempts INTEGER DEFAULT 0,  -- Failed verification attempts counter
ADD COLUMN otp_locked_until TIMESTAMPTZ;    -- Temporary lockout timestamp

-- Step 2: Create index for OTP hash lookup (performance)
CREATE INDEX idx_users_otp_hash ON users(phone, otp_hash) WHERE otp_hash IS NOT NULL;

-- Step 3: Create index for locked accounts (to check lockouts quickly)
-- Note: Cannot use NOW() in partial index predicate (not IMMUTABLE)
-- Instead, we create a simple index on otp_locked_until for filtering in queries
CREATE INDEX idx_users_otp_locked ON users(otp_locked_until) WHERE otp_locked_until IS NOT NULL;

-- Step 4: Add comment for documentation
COMMENT ON COLUMN users.otp_hash IS 'SHA-256 hash of OTP code + server-side pepper (never store plaintext OTP)';
COMMENT ON COLUMN users.otp_attempts IS 'Number of failed OTP verification attempts (reset on success)';
COMMENT ON COLUMN users.otp_locked_until IS 'Temporary lockout until this timestamp (NULL if not locked)';

-- Step 5: Keep old otp_code column temporarily for gradual migration
-- Will be dropped in a future migration after confirming new system works
-- DO NOT USE otp_code for new OTP generation - it's deprecated

COMMENT ON COLUMN users.otp_code IS 'DEPRECATED: Use otp_hash instead. Will be removed in future migration.';
