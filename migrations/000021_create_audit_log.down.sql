-- Rollback Migration 000021: Remove audit log

-- Drop triggers
DROP TRIGGER IF EXISTS trigger_audit_user_account_changes ON users;
DROP TRIGGER IF EXISTS trigger_audit_refresh_token_operations ON refresh_tokens;
DROP TRIGGER IF EXISTS trigger_audit_order_status_change ON orders;

-- Drop trigger functions
DROP FUNCTION IF EXISTS audit_user_account_changes();
DROP FUNCTION IF EXISTS audit_refresh_token_operations();
DROP FUNCTION IF EXISTS audit_order_status_change();

-- Drop partitioned table (cascades to partitions)
DROP TABLE IF EXISTS audit_log CASCADE;
