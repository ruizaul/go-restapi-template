-- Rollback Migration 000013: Remove secure delivery code features

-- Drop audit table
DROP TABLE IF EXISTS delivery_code_audit;

-- Remove attempt counter
ALTER TABLE orders
DROP COLUMN IF EXISTS delivery_code_attempts;

-- Restore default value (not recommended, but for rollback compatibility)
ALTER TABLE orders
ALTER COLUMN delivery_code SET DEFAULT '0000';

-- Restore comment
COMMENT ON COLUMN orders.delivery_code IS 'Código de 4 dígitos para confirmar entrega';
