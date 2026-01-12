-- Create driver_locations table for real-time location tracking
CREATE TABLE driver_locations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    driver_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    latitude DECIMAL(10, 8) NOT NULL,
    longitude DECIMAL(11, 8) NOT NULL,
    heading DECIMAL(5, 2), -- Direction in degrees (0-360)
    speed_kmh DECIMAL(5, 2), -- Speed in km/h
    accuracy_meters DECIMAL(6, 2), -- GPS accuracy in meters
    is_available BOOLEAN DEFAULT true, -- Driver available for new orders
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for driver_locations
CREATE INDEX idx_driver_locations_driver_id ON driver_locations(driver_id);
CREATE INDEX idx_driver_locations_location ON driver_locations(latitude, longitude);
CREATE INDEX idx_driver_locations_available ON driver_locations(is_available) WHERE is_available = true;
CREATE INDEX idx_driver_locations_updated_at ON driver_locations(updated_at DESC);

-- Composite index for finding available drivers by location
CREATE INDEX idx_driver_locations_available_location ON driver_locations(is_available, latitude, longitude) WHERE is_available = true;

-- Ensure only one location record per driver (upsert pattern)
CREATE UNIQUE INDEX idx_driver_locations_driver_unique ON driver_locations(driver_id);

-- Add comments
COMMENT ON TABLE driver_locations IS 'Real-time location tracking for drivers';
COMMENT ON COLUMN driver_locations.is_available IS 'Whether driver is available to receive new orders';
COMMENT ON COLUMN driver_locations.heading IS 'Direction of movement in degrees (0=North, 90=East, 180=South, 270=West)';
COMMENT ON COLUMN driver_locations.accuracy_meters IS 'GPS accuracy radius in meters';
