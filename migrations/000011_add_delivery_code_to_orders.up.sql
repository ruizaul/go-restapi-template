-- Add delivery_code column to orders table
ALTER TABLE orders ADD COLUMN delivery_code VARCHAR(4) NOT NULL DEFAULT '0000';

-- Add check constraint to ensure delivery_code is exactly 4 digits
ALTER TABLE orders ADD CONSTRAINT check_delivery_code_format
    CHECK (delivery_code ~ '^\d{4}$');

-- Add index for faster lookups when verifying delivery codes
CREATE INDEX idx_orders_delivery_code ON orders(id, delivery_code);

-- Add comment
COMMENT ON COLUMN orders.delivery_code IS '4-digit numeric code required to complete delivery';
