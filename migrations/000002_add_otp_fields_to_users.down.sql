-- Remove indexes
DROP INDEX IF EXISTS idx_users_account_status;
DROP INDEX IF EXISTS idx_users_phone_otp;

-- Remove additional registration fields
ALTER TABLE users
DROP COLUMN IF EXISTS birth_date,
DROP COLUMN IF EXISTS mother_last_name,
DROP COLUMN IF EXISTS last_name,
DROP COLUMN IF EXISTS first_name;

-- Remove OTP verification fields
ALTER TABLE users
DROP COLUMN IF EXISTS account_status,
DROP COLUMN IF EXISTS phone_verified,
DROP COLUMN IF EXISTS otp_expires_at,
DROP COLUMN IF EXISTS otp_code;
