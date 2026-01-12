-- Drop triggers first
DROP TRIGGER IF EXISTS trigger_update_fcm_tokens_updated_at ON fcm_tokens;
DROP TRIGGER IF EXISTS trigger_update_notifications_updated_at ON notifications;

-- Drop trigger functions
DROP FUNCTION IF EXISTS update_fcm_tokens_updated_at();
DROP FUNCTION IF EXISTS update_notifications_updated_at();

-- Drop tables (CASCADE will drop dependent objects like indexes)
DROP TABLE IF EXISTS fcm_tokens CASCADE;
DROP TABLE IF EXISTS notifications CASCADE;

-- Success marker
SELECT 'Migration 000005 rolled back: notifications and fcm_tokens tables dropped' as status;
