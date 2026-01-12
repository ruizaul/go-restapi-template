-- Rollback Migration 000020: Remove order state machine enforcement

-- Drop triggers
DROP TRIGGER IF EXISTS trigger_update_order_status_timestamps ON orders;
DROP TRIGGER IF EXISTS trigger_validate_driver_assignment ON orders;
DROP TRIGGER IF EXISTS trigger_validate_order_state_transition ON orders;

-- Drop functions
DROP FUNCTION IF EXISTS update_order_status_timestamps();
DROP FUNCTION IF EXISTS validate_driver_assignment();
DROP FUNCTION IF EXISTS validate_order_state_transition();
