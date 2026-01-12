-- Add OTP verification fields
ALTER TABLE users
ADD COLUMN otp_code VARCHAR(6),
ADD COLUMN otp_expires_at TIMESTAMPTZ,
ADD COLUMN phone_verified BOOLEAN DEFAULT FALSE,
ADD COLUMN account_status VARCHAR(20) DEFAULT 'pending' CHECK (account_status IN ('pending', 'active', 'suspended'));

-- Add additional registration fields
ALTER TABLE users
ADD COLUMN first_name VARCHAR(100),
ADD COLUMN last_name VARCHAR(100),
ADD COLUMN mother_last_name VARCHAR(100),
ADD COLUMN birth_date DATE;

-- Create index for OTP lookup
CREATE INDEX idx_users_phone_otp ON users(phone, otp_code) WHERE otp_code IS NOT NULL;

-- Create index for account status
CREATE INDEX idx_users_account_status ON users(account_status);

-- Update existing users to have active status (migration safety)
UPDATE users SET account_status = 'active', phone_verified = TRUE WHERE account_status = 'pending';
