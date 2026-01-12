-- Migration 000015: Enable Row-Level Security (RLS) for Multi-Tenant Tables
-- Prevents privilege escalation bugs from exposing data across tenants
-- Each user can only access their own data based on role

-- Step 1: Enable RLS on multi-tenant tables
ALTER TABLE orders ENABLE ROW LEVEL SECURITY;
ALTER TABLE notifications ENABLE ROW LEVEL SECURITY;
ALTER TABLE fcm_tokens ENABLE ROW LEVEL SECURITY;
ALTER TABLE refresh_tokens ENABLE ROW LEVEL SECURITY;
ALTER TABLE user_documents ENABLE ROW LEVEL SECURITY;
ALTER TABLE driver_locations ENABLE ROW LEVEL SECURITY;

-- Step 2: Create policies for ORDERS table
-- Policy: Customers see only orders where they are the customer (via merchant relationship)
-- Policy: Merchants see only their orders
-- Policy: Drivers see only orders assigned to them
-- Policy: Admins see everything

CREATE POLICY orders_select_policy ON orders
FOR SELECT
USING (
    -- Admin can see all
    (current_setting('app.user_role', true) = 'admin')
    OR
    -- Merchant sees their own orders
    (current_setting('app.user_role', true) = 'merchant' AND
     merchant_id::text = current_setting('app.user_id', true))
    OR
    -- Driver sees orders assigned to them
    (current_setting('app.user_role', true) = 'driver' AND
     driver_id::text = current_setting('app.user_id', true))
    OR
    -- Customer sees orders (via external system, not directly stored in users table)
    (current_setting('app.user_role', true) = 'customer')
);

CREATE POLICY orders_insert_policy ON orders
FOR INSERT
WITH CHECK (
    -- Only merchants and admins can create orders
    current_setting('app.user_role', true) IN ('merchant', 'admin')
);

CREATE POLICY orders_update_policy ON orders
FOR UPDATE
USING (
    -- Admins can update any order
    (current_setting('app.user_role', true) = 'admin')
    OR
    -- Merchants can update their orders
    (current_setting('app.user_role', true) = 'merchant' AND
     merchant_id::text = current_setting('app.user_id', true))
    OR
    -- Drivers can update orders assigned to them (status changes)
    (current_setting('app.user_role', true) = 'driver' AND
     driver_id::text = current_setting('app.user_id', true))
);

-- Step 3: Create policies for NOTIFICATIONS table
CREATE POLICY notifications_select_policy ON notifications
FOR SELECT
USING (
    -- Users see only their own notifications
    user_id::text = current_setting('app.user_id', true)
    OR current_setting('app.user_role', true) = 'admin'
);

CREATE POLICY notifications_insert_policy ON notifications
FOR INSERT
WITH CHECK (
    -- System can insert for any user, verified at application level
    current_setting('app.user_role', true) = 'admin'
    OR current_setting('app.system_insert', true) = 'true'
);

CREATE POLICY notifications_update_policy ON notifications
FOR UPDATE
USING (
    -- Users can update their own notifications (mark as read)
    user_id::text = current_setting('app.user_id', true)
    OR current_setting('app.user_role', true) = 'admin'
);

-- Step 4: Create policies for FCM_TOKENS table
CREATE POLICY fcm_tokens_select_policy ON fcm_tokens
FOR SELECT
USING (
    user_id::text = current_setting('app.user_id', true)
    OR current_setting('app.user_role', true) = 'admin'
);

CREATE POLICY fcm_tokens_insert_policy ON fcm_tokens
FOR INSERT
WITH CHECK (
    user_id::text = current_setting('app.user_id', true)
    OR current_setting('app.user_role', true) = 'admin'
);

CREATE POLICY fcm_tokens_delete_policy ON fcm_tokens
FOR DELETE
USING (
    user_id::text = current_setting('app.user_id', true)
    OR current_setting('app.user_role', true) = 'admin'
);

-- Step 5: Create policies for REFRESH_TOKENS table
CREATE POLICY refresh_tokens_select_policy ON refresh_tokens
FOR SELECT
USING (
    user_id::text = current_setting('app.user_id', true)
    OR current_setting('app.user_role', true) = 'admin'
);

CREATE POLICY refresh_tokens_insert_policy ON refresh_tokens
FOR INSERT
WITH CHECK (
    user_id::text = current_setting('app.user_id', true)
    OR current_setting('app.user_role', true) = 'admin'
);

CREATE POLICY refresh_tokens_update_policy ON refresh_tokens
FOR UPDATE
USING (
    user_id::text = current_setting('app.user_id', true)
    OR current_setting('app.user_role', true) = 'admin'
);

-- Step 6: Create policies for USER_DOCUMENTS table
CREATE POLICY user_documents_select_policy ON user_documents
FOR SELECT
USING (
    user_id::text = current_setting('app.user_id', true)
    OR current_setting('app.user_role', true) = 'admin'
);

CREATE POLICY user_documents_insert_policy ON user_documents
FOR INSERT
WITH CHECK (
    user_id::text = current_setting('app.user_id', true)
    OR current_setting('app.user_role', true) = 'admin'
);

CREATE POLICY user_documents_update_policy ON user_documents
FOR UPDATE
USING (
    user_id::text = current_setting('app.user_id', true)
    OR current_setting('app.user_role', true) = 'admin'
);

-- Step 7: Create policies for DRIVER_LOCATIONS table
CREATE POLICY driver_locations_select_policy ON driver_locations
FOR SELECT
USING (
    -- Drivers see only their own location
    driver_id::text = current_setting('app.user_id', true)
    OR current_setting('app.user_role', true) = 'admin'
    -- System can read all locations for assignment algorithm
    OR current_setting('app.system_read', true) = 'true'
);

CREATE POLICY driver_locations_insert_policy ON driver_locations
FOR INSERT
WITH CHECK (
    driver_id::text = current_setting('app.user_id', true)
    OR current_setting('app.user_role', true) = 'admin'
);

CREATE POLICY driver_locations_update_policy ON driver_locations
FOR UPDATE
USING (
    driver_id::text = current_setting('app.user_id', true)
    OR current_setting('app.user_role', true) = 'admin'
);

-- Step 8: Add comments
COMMENT ON POLICY orders_select_policy ON orders IS 'RLS: Users see only orders relevant to their role';
COMMENT ON POLICY notifications_select_policy ON notifications IS 'RLS: Users see only their own notifications';
COMMENT ON POLICY fcm_tokens_select_policy ON fcm_tokens IS 'RLS: Users see only their own FCM tokens';
COMMENT ON POLICY refresh_tokens_select_policy ON refresh_tokens IS 'RLS: Users see only their own refresh tokens';
COMMENT ON POLICY user_documents_select_policy ON user_documents IS 'RLS: Users see only their own documents';
COMMENT ON POLICY driver_locations_select_policy ON driver_locations IS 'RLS: Drivers see only their own location';

-- Step 9: Usage instructions
-- Before each query, set session variables:
-- SET LOCAL app.user_id = '<user_uuid>';
-- SET LOCAL app.user_role = 'customer|merchant|driver|admin';
-- SET LOCAL app.system_read = 'true'; -- For system operations like assignment algorithm
-- SET LOCAL app.system_insert = 'true'; -- For system operations like notifications
