-- Migration 000013: Secure Delivery Code with Cryptographic Generation + Audit Trail
-- Removes default '0000', adds attempt counter, and creates audit table

-- Step 1: Remove dangerous default value from delivery_code
ALTER TABLE orders
ALTER COLUMN delivery_code DROP DEFAULT;

-- Step 2: Add delivery code attempt counter
ALTER TABLE orders
ADD COLUMN delivery_code_attempts INTEGER DEFAULT 0 NOT NULL;

-- Step 3: Create delivery_code_audit table for tracking attempts
CREATE TABLE delivery_code_audit (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    attempted_code VARCHAR(4) NOT NULL,
    success BOOLEAN NOT NULL,
    attempted_by UUID REFERENCES users(id) ON DELETE SET NULL,
    ip_address VARCHAR(45),
    user_agent TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Step 4: Create indexes for audit queries
CREATE INDEX idx_delivery_code_audit_order_id ON delivery_code_audit(order_id);
CREATE INDEX idx_delivery_code_audit_created_at ON delivery_code_audit(created_at DESC);
CREATE INDEX idx_delivery_code_audit_success ON delivery_code_audit(success);
CREATE INDEX idx_delivery_code_audit_order_success ON delivery_code_audit(order_id, success);

-- Step 5: Add comments for documentation
COMMENT ON COLUMN orders.delivery_code IS 'Cryptographically random 4-digit code (no default, generated via crypto/rand)';
COMMENT ON COLUMN orders.delivery_code_attempts IS 'Number of failed delivery code verification attempts (max 3)';
COMMENT ON TABLE delivery_code_audit IS 'Audit trail for all delivery code verification attempts (fraud prevention)';
COMMENT ON COLUMN delivery_code_audit.attempted_code IS 'The code that was attempted (for fraud pattern analysis)';
COMMENT ON COLUMN delivery_code_audit.success IS 'Whether the attempt was successful';
COMMENT ON COLUMN delivery_code_audit.attempted_by IS 'User who attempted verification (usually driver)';

-- Step 6: Update existing orders with random codes (migration safety)
-- Generate cryptographically secure random 4-digit codes for existing orders
UPDATE orders
SET delivery_code = LPAD(FLOOR(RANDOM() * 10000)::TEXT, 4, '0')
WHERE delivery_code = '0000' OR delivery_code IS NULL;
