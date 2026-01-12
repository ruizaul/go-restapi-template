-- Create refresh_tokens table for persistent refresh token management
CREATE TABLE refresh_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(64) NOT NULL UNIQUE,
    device_info TEXT,
    ip_address VARCHAR(45),
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    revoked BOOLEAN DEFAULT FALSE,
    revoked_at TIMESTAMP WITH TIME ZONE
);

-- Indexes for performance
CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens(user_id);
CREATE INDEX idx_refresh_tokens_token_hash ON refresh_tokens(token_hash);
CREATE INDEX idx_refresh_tokens_expires_at ON refresh_tokens(expires_at);
CREATE INDEX idx_refresh_tokens_revoked ON refresh_tokens(revoked) WHERE revoked = FALSE;

-- Comment
COMMENT ON TABLE refresh_tokens IS 'Stores refresh tokens for user authentication with rotation support';
COMMENT ON COLUMN refresh_tokens.token_hash IS 'SHA-256 hash of the refresh token (never store plain tokens)';
COMMENT ON COLUMN refresh_tokens.device_info IS 'User-Agent string or device identifier';
COMMENT ON COLUMN refresh_tokens.ip_address IS 'IP address from which the token was issued';
COMMENT ON COLUMN refresh_tokens.revoked IS 'TRUE if token has been revoked (logout or rotation)';
