-- Create merchants table
CREATE TABLE merchants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    business_name VARCHAR(255) NOT NULL,
    business_type VARCHAR(100) NOT NULL, -- restaurant, pharmacy, grocery, etc.
    phone VARCHAR(20) NOT NULL,
    email VARCHAR(255),
    address TEXT NOT NULL,
    latitude DECIMAL(10, 8) NOT NULL,
    longitude DECIMAL(11, 8) NOT NULL,
    city VARCHAR(100) NOT NULL,
    state VARCHAR(100) NOT NULL,
    postal_code VARCHAR(10),
    country VARCHAR(2) DEFAULT 'MX',
    status VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'inactive', 'suspended')),
    rating DECIMAL(3, 2) DEFAULT 0.00 CHECK (rating >= 0 AND rating <= 5),
    total_orders INT DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for merchants
CREATE INDEX idx_merchants_user_id ON merchants(user_id);
CREATE INDEX idx_merchants_status ON merchants(status);
CREATE INDEX idx_merchants_location ON merchants(latitude, longitude);
CREATE INDEX idx_merchants_city ON merchants(city);
CREATE INDEX idx_merchants_business_type ON merchants(business_type);

-- Ensure only one merchant per user
CREATE UNIQUE INDEX idx_merchants_user_id_unique ON merchants(user_id);

-- Add comment
COMMENT ON TABLE merchants IS 'Merchants/businesses that receive delivery orders';
COMMENT ON COLUMN merchants.latitude IS 'Merchant pickup location latitude';
COMMENT ON COLUMN merchants.longitude IS 'Merchant pickup location longitude';
