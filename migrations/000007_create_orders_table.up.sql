-- Create orders table
CREATE TABLE orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    external_order_id VARCHAR(255), -- ID from external backend
    merchant_id UUID NOT NULL REFERENCES merchants(id) ON DELETE RESTRICT,
    driver_id UUID REFERENCES users(id) ON DELETE SET NULL,
    customer_name VARCHAR(255) NOT NULL,
    customer_phone VARCHAR(20) NOT NULL,
    
    -- Pickup information
    pickup_address TEXT NOT NULL,
    pickup_latitude DECIMAL(10, 8) NOT NULL,
    pickup_longitude DECIMAL(11, 8) NOT NULL,
    pickup_instructions TEXT,
    
    -- Delivery information
    delivery_address TEXT NOT NULL,
    delivery_latitude DECIMAL(10, 8) NOT NULL,
    delivery_longitude DECIMAL(11, 8) NOT NULL,
    delivery_instructions TEXT,
    
    -- Order details
    items JSONB NOT NULL, -- Array of order items
    total_amount DECIMAL(10, 2) NOT NULL CHECK (total_amount > 0),
    delivery_fee DECIMAL(10, 2) DEFAULT 0.00,
    
    -- Status tracking
    status VARCHAR(50) NOT NULL DEFAULT 'searching_driver' CHECK (status IN (
        'searching_driver', -- Looking for available driver
        'assigned',         -- Driver assigned, waiting for acceptance
        'accepted',         -- Driver accepted the order
        'picked_up',        -- Driver picked up the order
        'in_transit',       -- Driver is delivering
        'delivered',        -- Order delivered successfully
        'cancelled',        -- Order cancelled
        'no_driver_available' -- No driver found
    )),
    
    -- Distance and time estimates
    distance_km DECIMAL(6, 2), -- Distance from pickup to delivery
    estimated_duration_minutes INT, -- Estimated delivery time
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    assigned_at TIMESTAMP WITH TIME ZONE,
    accepted_at TIMESTAMP WITH TIME ZONE,
    picked_up_at TIMESTAMP WITH TIME ZONE,
    delivered_at TIMESTAMP WITH TIME ZONE,
    cancelled_at TIMESTAMP WITH TIME ZONE,
    
    -- Cancellation info
    cancellation_reason TEXT,
    cancelled_by UUID REFERENCES users(id) ON DELETE SET NULL
);

-- Create indexes for orders
CREATE INDEX idx_orders_merchant_id ON orders(merchant_id);
CREATE INDEX idx_orders_driver_id ON orders(driver_id);
CREATE INDEX idx_orders_status ON orders(status);
CREATE INDEX idx_orders_external_id ON orders(external_order_id);
CREATE INDEX idx_orders_created_at ON orders(created_at DESC);
CREATE INDEX idx_orders_pickup_location ON orders(pickup_latitude, pickup_longitude);
CREATE INDEX idx_orders_delivery_location ON orders(delivery_latitude, delivery_longitude);

-- Composite index for active orders by driver
CREATE INDEX idx_orders_driver_status ON orders(driver_id, status) WHERE status IN ('assigned', 'accepted', 'picked_up', 'in_transit');

-- Add comments
COMMENT ON TABLE orders IS 'Delivery orders from merchants to customers';
COMMENT ON COLUMN orders.external_order_id IS 'Original order ID from external backend';
COMMENT ON COLUMN orders.items IS 'JSON array of order items with name, quantity, price';
COMMENT ON COLUMN orders.status IS 'Current status of the order in the delivery workflow';
