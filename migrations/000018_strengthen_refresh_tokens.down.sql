-- Rollback Migration 000018: Remove refresh token security enhancements

-- Drop indexes
DROP INDEX IF EXISTS idx_refresh_tokens_user_device;
DROP INDEX IF EXISTS idx_refresh_tokens_last_used_at;
DROP INDEX IF EXISTS idx_refresh_tokens_device_id;

-- Remove columns
ALTER TABLE refresh_tokens
DROP COLUMN IF EXISTS revoked_reason,
DROP COLUMN IF EXISTS last_used_at,
DROP COLUMN IF EXISTS device_id;
