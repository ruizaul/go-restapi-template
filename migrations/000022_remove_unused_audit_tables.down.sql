-- Migration 000022 DOWN: Restore Audit Tables (Emergency Rollback Only)
-- WARNING: This will recreate empty tables - historical data is NOT restored

-- Step 1: Recreate audit_log partitioned table
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

-- Step 2: Recreate current month partition only
CREATE TABLE audit_log_2025_01 PARTITION OF audit_log
    FOR VALUES FROM ('2025-01-01') TO ('2025-02-01');

CREATE INDEX idx_audit_log_2025_01_created_at ON audit_log_2025_01(created_at DESC);
CREATE INDEX idx_audit_log_2025_01_table_record ON audit_log_2025_01(table_name, record_id);
CREATE INDEX idx_audit_log_2025_01_performed_by ON audit_log_2025_01(performed_by);
CREATE INDEX idx_audit_log_2025_01_action ON audit_log_2025_01(action);

-- Step 3: Recreate trigger functions
CREATE OR REPLACE FUNCTION audit_order_status_change()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'UPDATE' AND OLD.status != NEW.status THEN
        INSERT INTO audit_log (
            table_name, record_id, action, old_data, new_data,
            performed_by, performed_ip, user_agent
        ) VALUES (
            'orders',
            NEW.id,
            'status_change',
            jsonb_build_object('status', OLD.status, 'driver_id', OLD.driver_id, 'updated_at', OLD.updated_at),
            jsonb_build_object('status', NEW.status, 'driver_id', NEW.driver_id, 'updated_at', NEW.updated_at),
            NULLIF(current_setting('app.user_id', true), '')::uuid,
            NULLIF(current_setting('app.user_ip', true), ''),
            NULLIF(current_setting('app.user_agent', true), '')
        );
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

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

CREATE OR REPLACE FUNCTION audit_user_account_changes()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'UPDATE' THEN
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

-- Step 4: Recreate triggers
CREATE TRIGGER trigger_audit_order_status_change
    AFTER UPDATE ON orders
    FOR EACH ROW
    EXECUTE FUNCTION audit_order_status_change();

CREATE TRIGGER trigger_audit_refresh_token_operations
    AFTER INSERT OR UPDATE ON refresh_tokens
    FOR EACH ROW
    EXECUTE FUNCTION audit_refresh_token_operations();

CREATE TRIGGER trigger_audit_user_account_changes
    AFTER UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION audit_user_account_changes();

-- Step 5: Recreate delivery_code_audit table
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

CREATE INDEX idx_delivery_code_audit_order_id ON delivery_code_audit(order_id);
CREATE INDEX idx_delivery_code_audit_created_at ON delivery_code_audit(created_at DESC);
CREATE INDEX idx_delivery_code_audit_success ON delivery_code_audit(success);
CREATE INDEX idx_delivery_code_audit_order_success ON delivery_code_audit(order_id, success);

-- Step 6: Add comments
COMMENT ON TABLE audit_log IS 'Append-only audit trail for critical events (partitioned by month for performance)';
COMMENT ON TABLE delivery_code_audit IS 'Audit trail for all delivery code verification attempts (fraud prevention)';
COMMENT ON COLUMN orders.delivery_code_attempts IS 'Number of failed delivery code verification attempts (max 3)';
