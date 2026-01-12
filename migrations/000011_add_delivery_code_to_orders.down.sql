-- Remove delivery_code column and related constraints/indexes
DROP INDEX IF EXISTS idx_orders_delivery_code;
ALTER TABLE orders DROP CONSTRAINT IF EXISTS check_delivery_code_format;
ALTER TABLE orders DROP COLUMN IF EXISTS delivery_code;
