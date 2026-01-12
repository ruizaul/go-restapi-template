-- Rollback Migration 000016: Restore original CASCADE DELETE behavior

-- Remove soft delete columns
DROP INDEX IF EXISTS idx_users_deleted_at;
ALTER TABLE users DROP COLUMN IF EXISTS deleted_at;

DROP INDEX IF EXISTS idx_refresh_tokens_deleted_at;
ALTER TABLE refresh_tokens DROP COLUMN IF EXISTS deleted_at;

-- Restore driver_locations CASCADE
ALTER TABLE driver_locations
DROP CONSTRAINT IF EXISTS driver_locations_driver_id_fkey,
ADD CONSTRAINT driver_locations_driver_id_fkey
    FOREIGN KEY (driver_id) REFERENCES users(id) ON DELETE CASCADE;

-- Restore user_documents CASCADE
ALTER TABLE user_documents
DROP CONSTRAINT IF EXISTS user_documents_user_id_fkey,
ADD CONSTRAINT user_documents_user_id_fkey
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
