-- Create enum for fiscal regime
CREATE TYPE fiscal_regime_type AS ENUM (
    'general',                          -- Régimen General de Ley Personas Morales
    'simplificado_confianza',          -- Régimen Simplificado de Confianza (RESICO)
    'actividad_empresarial',           -- Actividad Empresarial y Profesional
    'arrendamiento',                   -- Arrendamiento
    'salarios',                        -- Sueldos y Salarios
    'incorporacion_fiscal'             -- Régimen de Incorporación Fiscal
);

-- Create user_documents table
CREATE TABLE IF NOT EXISTS user_documents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    
    -- Vehicle information
    vehicle_brand VARCHAR(100),
    vehicle_model VARCHAR(100),
    license_plate VARCHAR(20),
    
    -- Document URLs (images)
    circulation_card_url TEXT,
    ine_front_url TEXT,
    ine_back_url TEXT,
    driver_license_front_url TEXT,
    driver_license_back_url TEXT,
    profile_photo_url TEXT,
    
    -- Fiscal information
    fiscal_name VARCHAR(255),
    fiscal_rfc VARCHAR(13),
    fiscal_zip_code VARCHAR(5),
    fiscal_regime fiscal_regime_type,
    fiscal_street VARCHAR(255),
    fiscal_ext_number VARCHAR(20),
    fiscal_int_number VARCHAR(20),
    fiscal_neighborhood VARCHAR(100),
    fiscal_city VARCHAR(100),
    fiscal_state VARCHAR(100),
    fiscal_certificate_url TEXT,
    
    -- Review status
    reviewed BOOLEAN NOT NULL DEFAULT false,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create index on user_id for faster lookups
CREATE INDEX idx_user_documents_user_id ON user_documents(user_id);

-- Create index on reviewed status for admin queries
CREATE INDEX idx_user_documents_reviewed ON user_documents(reviewed);
