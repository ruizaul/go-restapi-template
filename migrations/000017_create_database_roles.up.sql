-- Migration 000017: Create PostgreSQL Roles with Minimal Privileges
-- Implements principle of least privilege for database access

-- Step 1: Create roles (if they don't exist)
DO $$
BEGIN
    -- Read-only role for analytics/reporting
    IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'app_read') THEN
        CREATE ROLE app_read WITH LOGIN PASSWORD 'CHANGE_ME_IN_PRODUCTION';
    END IF;

    -- Read/Write role for application (main role)
    IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'app_write') THEN
        CREATE ROLE app_write WITH LOGIN PASSWORD 'CHANGE_ME_IN_PRODUCTION';
    END IF;

    -- Admin role for migrations and schema changes
    IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'app_admin') THEN
        CREATE ROLE app_admin WITH LOGIN PASSWORD 'CHANGE_ME_IN_PRODUCTION';
    END IF;
END
$$;

-- Step 2: Grant CONNECT to database
GRANT CONNECT ON DATABASE tacoshare_delivery TO app_read, app_write, app_admin;

-- Step 3: Grant USAGE on schema
GRANT USAGE ON SCHEMA public TO app_read, app_write, app_admin;

-- Step 4: Configure app_read role (SELECT only)
GRANT SELECT ON ALL TABLES IN SCHEMA public TO app_read;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT ON TABLES TO app_read;

-- Step 5: Configure app_write role (SELECT, INSERT, UPDATE, DELETE)
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO app_write;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO app_write;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO app_write;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT USAGE, SELECT ON SEQUENCES TO app_write;

-- Step 6: Configure app_admin role (full access including DDL)
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO app_admin;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO app_admin;
GRANT ALL PRIVILEGES ON SCHEMA public TO app_admin;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL PRIVILEGES ON TABLES TO app_admin;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL PRIVILEGES ON SEQUENCES TO app_admin;

-- Step 7: Revoke PUBLIC access (security hardening)
REVOKE ALL ON SCHEMA public FROM PUBLIC;
REVOKE ALL ON ALL TABLES IN SCHEMA public FROM PUBLIC;

-- Step 8: Add role descriptions
COMMENT ON ROLE app_read IS 'Read-only access for analytics and reporting';
COMMENT ON ROLE app_write IS 'Application runtime role with read/write access (use this for the app)';
COMMENT ON ROLE app_admin IS 'Admin role for migrations and schema changes only';

-- Step 9: Usage instructions
-- After running this migration:
-- 1. Change passwords immediately:
--    ALTER ROLE app_read WITH PASSWORD '<strong_random_password>';
--    ALTER ROLE app_write WITH PASSWORD '<strong_random_password>';
--    ALTER ROLE app_admin WITH PASSWORD '<strong_random_password>';
--
-- 2. Store passwords in secret manager:
--    AWS: aws secretsmanager create-secret --name tacoshare/db/app_write
--    GCP: gcloud secrets create tacoshare-db-app-write
--
-- 3. Update application DATABASE_URL to use app_write:
--    DATABASE_URL=postgres://app_write:<password>@host:5432/tacoshare_delivery
--
-- 4. Use app_admin ONLY for migrations:
--    migrate -database postgres://app_admin:password@host/db up
--
-- 5. Disable superuser (postgres) for application access:
--    ALTER ROLE postgres WITH NOLOGIN; -- Only after verifying app works with app_write
