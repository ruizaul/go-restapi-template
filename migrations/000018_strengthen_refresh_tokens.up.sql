-- Migration 000018: Strengthen Refresh Tokens with Device Binding and Theft Detection
-- Adds device_id for stronger binding and last_used_at for reuse detection

-- Step 1: Add new security columns to refresh_tokens
ALTER TABLE refresh_tokens
ADD COLUMN device_id VARCHAR(255),                -- Strong device identifier (not just User-Agent)
ADD COLUMN last_used_at TIMESTAMPTZ,              -- Track when token was last used for reuse detection
ADD COLUMN revoked_reason TEXT;                   -- Why token was revoked (audit trail)

-- Step 2: Create indexes for performance and security queries
CREATE INDEX idx_refresh_tokens_device_id ON refresh_tokens(device_id) WHERE revoked = false;
CREATE INDEX idx_refresh_tokens_last_used_at ON refresh_tokens(last_used_at DESC);
CREATE INDEX idx_refresh_tokens_user_device ON refresh_tokens(user_id, device_id) WHERE revoked = false;

-- Step 3: Add constraint to enforce device_id presence for new tokens
-- (Allow NULL temporarily for existing tokens during migration)

-- Step 4: Add comments for documentation
COMMENT ON COLUMN refresh_tokens.device_id IS 'Strong device identifier (e.g., device UUID, not just User-Agent) for token binding';
COMMENT ON COLUMN refresh_tokens.last_used_at IS 'Last time token was used (for detecting stolen token reuse)';
COMMENT ON COLUMN refresh_tokens.revoked_reason IS 'Reason for revocation: "logout", "token_theft_detected", "password_changed", "admin_revoke"';

-- Step 5: Security notes
-- Token Theft Detection:
-- If a revoked token is used again, it indicates potential theft.
-- Action: Revoke ALL tokens for that user and force re-login.
--
-- Device Binding:
-- New tokens MUST include device_id (enforce in application code).
-- When refreshing, verify device_id matches the stored value.
-- If mismatch: reject refresh and mark as suspicious.
--
-- Rotation Security:
-- On each refresh:
--   1. Update last_used_at
--   2. Revoke old token
--   3. Issue new token with same device_id
--   4. If old token used again after revocation: ALERT (theft detected)
