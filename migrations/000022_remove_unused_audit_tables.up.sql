-- Migration 000022: Remove Unused Audit Tables
-- Eliminates audit_log and delivery_code_audit tables that are not used by the application
-- Keeps delivery_code_attempts column in orders (essential for rate limiting)

-- Step 1: Drop triggers for audit_log (must drop before functions)
DROP TRIGGER IF EXISTS trigger_audit_order_status_change ON orders;
DROP TRIGGER IF EXISTS trigger_audit_refresh_token_operations ON refresh_tokens;
DROP TRIGGER IF EXISTS trigger_audit_user_account_changes ON users;

-- Step 2: Drop trigger functions for audit_log
DROP FUNCTION IF EXISTS audit_order_status_change();
DROP FUNCTION IF EXISTS audit_refresh_token_operations();
DROP FUNCTION IF EXISTS audit_user_account_changes();

-- Step 3: Drop partitioned audit_log table (cascades to all partitions)
DROP TABLE IF EXISTS audit_log CASCADE;

-- Step 4: Drop delivery_code_audit table
DROP TABLE IF EXISTS delivery_code_audit;

-- Step 5: Add comment explaining why delivery_code_attempts is kept
COMMENT ON COLUMN orders.delivery_code_attempts IS 'Number of failed delivery code verification attempts (max 3) - used for rate limiting in application layer';

-- Step 6: Migration notes
-- The following have been removed:
-- - audit_log (partitioned table with monthly partitions)
-- - audit_log_2025_01, audit_log_2025_02, audit_log_2025_03, audit_log_2025_04 (partitions)
-- - audit_order_status_change() trigger function
-- - audit_refresh_token_operations() trigger function
-- - audit_user_account_changes() trigger function
-- - delivery_code_audit table
--
-- Rationale:
-- - No application code queries these audit tables
-- - Triggers add overhead to every INSERT/UPDATE on orders, refresh_tokens, users
-- - Manual partition management required for audit_log (maintenance burden)
-- - No retention policy = unbounded growth (eventual database bloat)
-- - delivery_code_attempts column in orders is sufficient for rate limiting
--
-- Alternative approach for production auditing:
-- - Use structured application logging (CloudWatch, Loki, etc.)
-- - Enable Postgres WAL archiving for point-in-time recovery
-- - Implement targeted audit logging only for critical compliance events
