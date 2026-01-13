-- 000002_add_password_to_users.down.sql
-- Rollback migration: Removes password_hash column from users table

ALTER TABLE users DROP COLUMN IF EXISTS password_hash;
