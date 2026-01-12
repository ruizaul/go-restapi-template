-- Rollback Migration 000017: Remove database roles

-- Revoke all privileges
REVOKE ALL PRIVILEGES ON ALL TABLES IN SCHEMA public FROM app_admin;
REVOKE ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public FROM app_admin;
REVOKE ALL PRIVILEGES ON SCHEMA public FROM app_admin;

REVOKE ALL ON ALL TABLES IN SCHEMA public FROM app_write;
REVOKE ALL ON ALL SEQUENCES IN SCHEMA public FROM app_write;

REVOKE ALL ON ALL TABLES IN SCHEMA public FROM app_read;

REVOKE USAGE ON SCHEMA public FROM app_read, app_write, app_admin;
REVOKE CONNECT ON DATABASE tacoshare_delivery FROM app_read, app_write, app_admin;

-- Drop roles
DROP ROLE IF EXISTS app_admin;
DROP ROLE IF EXISTS app_write;
DROP ROLE IF EXISTS app_read;

-- Restore PUBLIC access
GRANT USAGE ON SCHEMA public TO PUBLIC;
