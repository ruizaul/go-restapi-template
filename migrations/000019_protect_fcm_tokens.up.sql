-- Migration 000019: Protect FCM Tokens with Encryption and Cleanup
-- Encrypts FCM tokens and adds invalid flag for cleanup

-- Step 1: Add encrypted token column alongside plaintext (gradual migration)
ALTER TABLE fcm_tokens
ADD COLUMN token_encrypted BYTEA,
ADD COLUMN invalid BOOLEAN DEFAULT false NOT NULL;

-- Step 2: Create indexes for cleanup and queries
CREATE INDEX idx_fcm_tokens_invalid ON fcm_tokens(invalid) WHERE invalid = true;
CREATE INDEX idx_fcm_tokens_created_at ON fcm_tokens(created_at DESC);
CREATE INDEX idx_fcm_tokens_invalid_cleanup ON fcm_tokens(invalid, created_at) WHERE invalid = true;

-- Step 3: Add comments
COMMENT ON COLUMN fcm_tokens.token_encrypted IS 'Encrypted FCM token (AES-256, key in KMS). Use decrypt_text() to read.';
COMMENT ON COLUMN fcm_tokens.invalid IS 'Mark as invalid when FCM returns InvalidRegistration or NotRegistered';

-- Step 4: Create cleanup function for invalid/old tokens
CREATE OR REPLACE FUNCTION cleanup_invalid_fcm_tokens()
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER := 0;
    temp_count INTEGER;
BEGIN
    -- Delete tokens marked as invalid for more than 7 days
    DELETE FROM fcm_tokens
    WHERE invalid = true
    AND updated_at < NOW() - INTERVAL '7 days';

    GET DIAGNOSTICS temp_count = ROW_COUNT;
    deleted_count := deleted_count + temp_count;

    -- Also delete very old tokens (>90 days) regardless of status
    DELETE FROM fcm_tokens
    WHERE created_at < NOW() - INTERVAL '90 days';

    GET DIAGNOSTICS temp_count = ROW_COUNT;
    deleted_count := deleted_count + temp_count;

    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION cleanup_invalid_fcm_tokens IS 'Cleanup function to delete invalid FCM tokens (run daily via cron)';

-- Step 5: Usage instructions
-- To populate encrypted tokens after migration:
-- UPDATE fcm_tokens SET token_encrypted = encrypt_text(token, :encryption_key) WHERE token IS NOT NULL;
--
-- To mark token as invalid (when FCM rejects it):
-- UPDATE fcm_tokens SET invalid = true WHERE token = :rejected_token;
--
-- Setup daily cleanup job (cron):
-- SELECT cleanup_invalid_fcm_tokens(); -- Run daily at 3am
