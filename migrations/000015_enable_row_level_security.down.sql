-- Rollback Migration 000015: Disable Row-Level Security

-- Drop all RLS policies
DROP POLICY IF EXISTS driver_locations_update_policy ON driver_locations;
DROP POLICY IF EXISTS driver_locations_insert_policy ON driver_locations;
DROP POLICY IF EXISTS driver_locations_select_policy ON driver_locations;

DROP POLICY IF EXISTS user_documents_update_policy ON user_documents;
DROP POLICY IF EXISTS user_documents_insert_policy ON user_documents;
DROP POLICY IF EXISTS user_documents_select_policy ON user_documents;

DROP POLICY IF EXISTS refresh_tokens_update_policy ON refresh_tokens;
DROP POLICY IF EXISTS refresh_tokens_insert_policy ON refresh_tokens;
DROP POLICY IF EXISTS refresh_tokens_select_policy ON refresh_tokens;

DROP POLICY IF EXISTS fcm_tokens_delete_policy ON fcm_tokens;
DROP POLICY IF EXISTS fcm_tokens_insert_policy ON fcm_tokens;
DROP POLICY IF EXISTS fcm_tokens_select_policy ON fcm_tokens;

DROP POLICY IF EXISTS notifications_update_policy ON notifications;
DROP POLICY IF EXISTS notifications_insert_policy ON notifications;
DROP POLICY IF EXISTS notifications_select_policy ON notifications;

DROP POLICY IF EXISTS orders_update_policy ON orders;
DROP POLICY IF EXISTS orders_insert_policy ON orders;
DROP POLICY IF EXISTS orders_select_policy ON orders;

-- Disable RLS on tables
ALTER TABLE driver_locations DISABLE ROW LEVEL SECURITY;
ALTER TABLE user_documents DISABLE ROW LEVEL SECURITY;
ALTER TABLE refresh_tokens DISABLE ROW LEVEL SECURITY;
ALTER TABLE fcm_tokens DISABLE ROW LEVEL SECURITY;
ALTER TABLE notifications DISABLE ROW LEVEL SECURITY;
ALTER TABLE orders DISABLE ROW LEVEL SECURITY;
