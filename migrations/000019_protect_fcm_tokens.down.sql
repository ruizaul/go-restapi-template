-- Rollback Migration 000019: Remove FCM token protection

-- Drop cleanup function
DROP FUNCTION IF EXISTS cleanup_invalid_fcm_tokens();

-- Drop indexes
DROP INDEX IF EXISTS idx_fcm_tokens_invalid_cleanup;
DROP INDEX IF EXISTS idx_fcm_tokens_created_at;
DROP INDEX IF EXISTS idx_fcm_tokens_invalid;

-- Remove columns
ALTER TABLE fcm_tokens
DROP COLUMN IF EXISTS invalid,
DROP COLUMN IF EXISTS token_encrypted;
