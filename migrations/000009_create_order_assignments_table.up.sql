-- Create order_assignments table to track assignment attempts
CREATE TABLE order_assignments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    driver_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    
    -- Assignment details
    attempt_number INT NOT NULL, -- 1st, 2nd, 3rd attempt, etc.
    search_radius_km DECIMAL(6, 2) NOT NULL, -- Radius used for this attempt
    distance_to_pickup_km DECIMAL(6, 2) NOT NULL, -- Driver's distance to pickup
    estimated_arrival_minutes INT, -- Estimated time to reach pickup
    
    -- Status tracking
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN (
        'pending',   -- Notification sent, waiting for response
        'accepted',  -- Driver accepted the order
        'rejected',  -- Driver explicitly rejected
        'timeout',   -- Driver didn't respond in time
        'expired'    -- Assignment expired (order assigned to another driver)
    )),
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    responded_at TIMESTAMP WITH TIME ZONE,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL, -- When this assignment expires (10 seconds)
    
    -- Response details
    rejection_reason TEXT
);

-- Create indexes for order_assignments
CREATE INDEX idx_order_assignments_order_id ON order_assignments(order_id);
CREATE INDEX idx_order_assignments_driver_id ON order_assignments(driver_id);
CREATE INDEX idx_order_assignments_status ON order_assignments(status);
CREATE INDEX idx_order_assignments_created_at ON order_assignments(created_at DESC);
CREATE INDEX idx_order_assignments_expires_at ON order_assignments(expires_at);

-- Composite index for finding pending assignments
CREATE INDEX idx_order_assignments_pending ON order_assignments(order_id, status, expires_at) WHERE status = 'pending';

-- Add comments
COMMENT ON TABLE order_assignments IS 'History of order assignment attempts to drivers';
COMMENT ON COLUMN order_assignments.attempt_number IS 'Sequential attempt number for this order (1, 2, 3...)';
COMMENT ON COLUMN order_assignments.search_radius_km IS 'The search radius used when this driver was found';
COMMENT ON COLUMN order_assignments.expires_at IS 'When this assignment offer expires (typically 10 seconds after creation)';
