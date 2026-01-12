-- Migration 000016: Fix Dangerous CASCADE DELETE Relationships
-- Changes risky cascades to RESTRICT or implements soft delete

-- Step 1: Change user_documents FK to RESTRICT (don't delete users with pending docs)
ALTER TABLE user_documents
DROP CONSTRAINT IF EXISTS user_documents_user_id_fkey,
ADD CONSTRAINT user_documents_user_id_fkey
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE RESTRICT;

-- Step 2: Add soft delete to refresh_tokens (audit trail)
ALTER TABLE refresh_tokens
ADD COLUMN deleted_at TIMESTAMPTZ;

CREATE INDEX idx_refresh_tokens_deleted_at ON refresh_tokens(deleted_at) WHERE deleted_at IS NULL;

COMMENT ON COLUMN refresh_tokens.deleted_at IS 'Soft delete timestamp (NULL = active, NOT NULL = deleted)';

-- Step 3: Change driver_locations to RESTRICT (don't auto-delete location history)
-- Keep the location for audit/investigation even if driver account is deleted
ALTER TABLE driver_locations
DROP CONSTRAINT IF EXISTS driver_locations_driver_id_fkey,
ADD CONSTRAINT driver_locations_driver_id_fkey
    FOREIGN KEY (driver_id) REFERENCES users(id) ON DELETE RESTRICT;

-- Step 4: Keep orders.driver_id as SET NULL (correct behavior)
-- When driver is deleted, orders should remain but driver_id becomes NULL
-- This is already correct in the schema

-- Step 5: Keep orders.cancelled_by as SET NULL (correct behavior)
-- If user who cancelled is deleted, we still keep the order record
-- This is already correct in the schema

-- Step 6: Add deleted_at to users for soft delete (future migration)
ALTER TABLE users
ADD COLUMN deleted_at TIMESTAMPTZ;

CREATE INDEX idx_users_deleted_at ON users(deleted_at) WHERE deleted_at IS NULL;

COMMENT ON COLUMN users.deleted_at IS 'Soft delete timestamp for GDPR compliance and audit (NULL = active)';

-- Step 7: Change fcm_tokens to CASCADE (correct - tokens are worthless without user)
-- This is already correct - tokens should be deleted with user

-- Step 8: Change notifications to CASCADE (correct - notifications are user-specific)
-- This is already correct - notifications should be deleted with user

-- Step 9: Summary of changes
-- RESTRICT: user_documents, driver_locations (preserve for audit/compliance)
-- SET NULL: orders.driver_id, orders.cancelled_by (preserve order history)
-- CASCADE: fcm_tokens, notifications, refresh_tokens, order_assignments (ephemeral data)
-- SOFT DELETE: users, refresh_tokens (with deleted_at timestamp)
