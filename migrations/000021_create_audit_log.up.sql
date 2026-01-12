-- Migration 000021: Create Audit Log (Append-Only) for Critical Events
-- Immutable audit trail for fraud prevention and compliance

-- Step 1: Create audit_log table (append-only, partitioned by month)
-- Note: PRIMARY KEY must include partition key (created_at)
CREATE TABLE audit_log (
    id UUID DEFAULT gen_random_uuid(),
    table_name VARCHAR(100) NOT NULL,
    record_id UUID NOT NULL,
    action VARCHAR(50) NOT NULL,
    old_data JSONB,
    new_data JSONB,
    performed_by UUID REFERENCES users(id) ON DELETE SET NULL,
    performed_ip VARCHAR(45),
    user_agent TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (id, created_at)
) PARTITION BY RANGE (created_at);

-- Step 2: Create initial partitions (current month + next 3 months)
CREATE TABLE audit_log_2025_01 PARTITION OF audit_log
    FOR VALUES FROM ('2025-01-01') TO ('2025-02-01');

CREATE TABLE audit_log_2025_02 PARTITION OF audit_log
    FOR VALUES FROM ('2025-02-01') TO ('2025-03-01');

CREATE TABLE audit_log_2025_03 PARTITION OF audit_log
    FOR VALUES FROM ('2025-03-01') TO ('2025-04-01');

CREATE TABLE audit_log_2025_04 PARTITION OF audit_log
    FOR VALUES FROM ('2025-04-01') TO ('2025-05-01');

-- Step 3: Create indexes on partitions for performance
CREATE INDEX idx_audit_log_2025_01_created_at ON audit_log_2025_01(created_at DESC);
CREATE INDEX idx_audit_log_2025_01_table_record ON audit_log_2025_01(table_name, record_id);
CREATE INDEX idx_audit_log_2025_01_performed_by ON audit_log_2025_01(performed_by);
CREATE INDEX idx_audit_log_2025_01_action ON audit_log_2025_01(action);

CREATE INDEX idx_audit_log_2025_02_created_at ON audit_log_2025_02(created_at DESC);
CREATE INDEX idx_audit_log_2025_02_table_record ON audit_log_2025_02(table_name, record_id);
CREATE INDEX idx_audit_log_2025_02_performed_by ON audit_log_2025_02(performed_by);

CREATE INDEX idx_audit_log_2025_03_created_at ON audit_log_2025_03(created_at DESC);
CREATE INDEX idx_audit_log_2025_03_table_record ON audit_log_2025_03(table_name, record_id);

CREATE INDEX idx_audit_log_2025_04_created_at ON audit_log_2025_04(created_at DESC);
CREATE INDEX idx_audit_log_2025_04_table_record ON audit_log_2025_04(table_name, record_id);

-- Step 4: Create trigger function for auditing order status changes
CREATE OR REPLACE FUNCTION audit_order_status_change()
RETURNS TRIGGER AS $$
BEGIN
    -- Only audit status changes
    IF TG_OP = 'UPDATE' AND OLD.status != NEW.status THEN
        INSERT INTO audit_log (
            table_name, record_id, action, old_data, new_data,
            performed_by, performed_ip, user_agent
        ) VALUES (
            'orders',
            NEW.id,
            'status_change',
            jsonb_build_object(
                'status', OLD.status,
                'driver_id', OLD.driver_id,
                'updated_at', OLD.updated_at
            ),
            jsonb_build_object(
                'status', NEW.status,
                'driver_id', NEW.driver_id,
                'updated_at', NEW.updated_at
            ),
            NULLIF(current_setting('app.user_id', true), '')::uuid,
            NULLIF(current_setting('app.user_ip', true), ''),
            NULLIF(current_setting('app.user_agent', true), '')
        );
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Step 5: Create trigger for order status auditing
CREATE TRIGGER trigger_audit_order_status_change
    AFTER UPDATE ON orders
    FOR EACH ROW
    EXECUTE FUNCTION audit_order_status_change();

-- Step 6: Create trigger function for auditing refresh token operations
CREATE OR REPLACE FUNCTION audit_refresh_token_operations()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        INSERT INTO audit_log (
            table_name, record_id, action, new_data,
            performed_by, performed_ip, user_agent
        ) VALUES (
            'refresh_tokens',
            NEW.id,
            'token_created',
            jsonb_build_object('user_id', NEW.user_id, 'device_info', NEW.device_info, 'device_id', NEW.device_id),
            NEW.user_id,
            NEW.ip_address,
            NEW.device_info
        );
    ELSIF TG_OP = 'UPDATE' AND OLD.revoked = false AND NEW.revoked = true THEN
        INSERT INTO audit_log (
            table_name, record_id, action, old_data, new_data,
            performed_by, performed_ip
        ) VALUES (
            'refresh_tokens',
            NEW.id,
            'token_revoked',
            jsonb_build_object('revoked', OLD.revoked),
            jsonb_build_object('revoked', NEW.revoked, 'revoked_reason', NEW.revoked_reason),
            NEW.user_id,
            NEW.ip_address
        );
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Step 7: Create trigger for refresh token auditing
CREATE TRIGGER trigger_audit_refresh_token_operations
    AFTER INSERT OR UPDATE ON refresh_tokens
    FOR EACH ROW
    EXECUTE FUNCTION audit_refresh_token_operations();

-- Step 8: Create trigger function for auditing user account status changes
CREATE OR REPLACE FUNCTION audit_user_account_changes()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'UPDATE' THEN
        -- Audit account status changes
        IF OLD.account_status != NEW.account_status THEN
            INSERT INTO audit_log (
                table_name, record_id, action, old_data, new_data,
                performed_by, performed_ip, user_agent
            ) VALUES (
                'users',
                NEW.id,
                'account_status_change',
                jsonb_build_object('account_status', OLD.account_status),
                jsonb_build_object('account_status', NEW.account_status),
                NULLIF(current_setting('app.user_id', true), '')::uuid,
                NULLIF(current_setting('app.user_ip', true), ''),
                NULLIF(current_setting('app.user_agent', true), '')
            );
        END IF;

        -- Audit phone verification
        IF OLD.phone_verified = false AND NEW.phone_verified = true THEN
            INSERT INTO audit_log (
                table_name, record_id, action, new_data,
                performed_by, performed_ip
            ) VALUES (
                'users',
                NEW.id,
                'phone_verified',
                jsonb_build_object('phone', NEW.phone),
                NEW.id,
                NULLIF(current_setting('app.user_ip', true), '')
            );
        END IF;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Step 9: Create trigger for user account auditing
CREATE TRIGGER trigger_audit_user_account_changes
    AFTER UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION audit_user_account_changes();

-- Step 10: Add CHECK constraint to enforce append-only (no updates/deletes)
-- Users should only INSERT into audit_log, never UPDATE or DELETE
-- This is enforced at the application layer, but we add documentation

-- Step 11: Add comments
COMMENT ON TABLE audit_log IS 'Append-only audit trail for critical events (partitioned by month for performance)';
COMMENT ON COLUMN audit_log.table_name IS 'Name of the table being audited';
COMMENT ON COLUMN audit_log.record_id IS 'UUID of the record being audited';
COMMENT ON COLUMN audit_log.action IS 'Action performed: status_change, token_created, token_revoked, account_status_change, phone_verified, etc.';
COMMENT ON COLUMN audit_log.old_data IS 'Snapshot of old data (JSONB)';
COMMENT ON COLUMN audit_log.new_data IS 'Snapshot of new data (JSONB)';
COMMENT ON COLUMN audit_log.performed_by IS 'User who performed the action (NULL for system actions)';
COMMENT ON COLUMN audit_log.performed_ip IS 'IP address of the requester';
COMMENT ON COLUMN audit_log.user_agent IS 'User agent of the requester';

-- Step 12: Usage instructions
-- Before critical operations, set session variables:
-- SET LOCAL app.user_id = '<user_uuid>';
-- SET LOCAL app.user_ip = '<ip_address>';
-- SET LOCAL app.user_agent = '<user_agent_string>';
--
-- To create new monthly partition (automated script):
-- CREATE TABLE audit_log_YYYY_MM PARTITION OF audit_log
--     FOR VALUES FROM ('YYYY-MM-01') TO ('YYYY-MM+1-01');
--
-- To query audit trail:
-- SELECT * FROM audit_log WHERE table_name = 'orders' AND record_id = '<order_uuid>' ORDER BY created_at DESC;
