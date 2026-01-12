-- Rollback Migration 000014: Remove PII encryption

-- Drop encrypted columns
ALTER TABLE user_documents
DROP COLUMN IF EXISTS profile_photo_url_encrypted,
DROP COLUMN IF EXISTS fiscal_certificate_url_encrypted,
DROP COLUMN IF EXISTS circulation_card_url_encrypted,
DROP COLUMN IF EXISTS driver_license_back_url_encrypted,
DROP COLUMN IF EXISTS driver_license_front_url_encrypted,
DROP COLUMN IF EXISTS ine_back_url_encrypted,
DROP COLUMN IF EXISTS ine_front_url_encrypted,
DROP COLUMN IF EXISTS fiscal_name_encrypted,
DROP COLUMN IF EXISTS fiscal_rfc_encrypted;

ALTER TABLE orders
DROP COLUMN IF EXISTS delivery_address_encrypted,
DROP COLUMN IF EXISTS pickup_address_encrypted,
DROP COLUMN IF EXISTS customer_phone_encrypted;

ALTER TABLE users
DROP COLUMN IF EXISTS phone_encrypted;

-- Drop helper functions
DROP FUNCTION IF EXISTS decrypt_text(BYTEA, TEXT);
DROP FUNCTION IF EXISTS encrypt_text(TEXT, TEXT);

-- Note: We don't drop pgcrypto extension in case other features use it
-- DROP EXTENSION IF EXISTS pgcrypto;
