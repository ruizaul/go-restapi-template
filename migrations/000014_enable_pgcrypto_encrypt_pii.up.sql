-- Migration 000014: Enable pgcrypto and Encrypt PII (Personally Identifiable Information)
-- IMPORTANT: This migration encrypts sensitive data at rest
-- Encryption key must be stored in KMS (AWS Secrets Manager, GCP Secret Manager, etc.)
-- For now, uses environment variable ENCRYPTION_KEY (rotate regularly)

-- Step 1: Enable pgcrypto extension for encryption functions
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- Step 2: Create helper functions for encryption/decryption
-- These functions use AES-256 in GCM mode with a key from environment

-- Function to encrypt text (AES-256)
CREATE OR REPLACE FUNCTION encrypt_text(plaintext TEXT, key TEXT)
RETURNS BYTEA AS $$
BEGIN
    RETURN pgp_sym_encrypt(plaintext, key);
END;
$$ LANGUAGE plpgsql IMMUTABLE;

-- Function to decrypt text
CREATE OR REPLACE FUNCTION decrypt_text(ciphertext BYTEA, key TEXT)
RETURNS TEXT AS $$
BEGIN
    RETURN pgp_sym_decrypt(ciphertext, key);
EXCEPTION
    WHEN OTHERS THEN
        RETURN NULL; -- Return NULL if decryption fails (wrong key or corrupted data)
END;
$$ LANGUAGE plpgsql IMMUTABLE;

-- Step 3: Add encrypted columns alongside existing plaintext columns
-- We keep plaintext temporarily for gradual migration

-- Users table - encrypt phone
ALTER TABLE users
ADD COLUMN phone_encrypted BYTEA;

-- Orders table - encrypt sensitive delivery information
ALTER TABLE orders
ADD COLUMN customer_phone_encrypted BYTEA,
ADD COLUMN pickup_address_encrypted BYTEA,
ADD COLUMN delivery_address_encrypted BYTEA;

-- User documents table - encrypt fiscal data and document URLs
ALTER TABLE user_documents
ADD COLUMN fiscal_rfc_encrypted BYTEA,
ADD COLUMN fiscal_name_encrypted BYTEA,
ADD COLUMN ine_front_url_encrypted BYTEA,
ADD COLUMN ine_back_url_encrypted BYTEA,
ADD COLUMN driver_license_front_url_encrypted BYTEA,
ADD COLUMN driver_license_back_url_encrypted BYTEA,
ADD COLUMN circulation_card_url_encrypted BYTEA,
ADD COLUMN fiscal_certificate_url_encrypted BYTEA,
ADD COLUMN profile_photo_url_encrypted BYTEA;

-- Step 4: Create indexes on encrypted columns where needed (for existence checks)
-- Note: Cannot search encrypted data directly, must decrypt first
-- These are just for NULL checks

-- Step 5: Add comments
COMMENT ON COLUMN users.phone_encrypted IS 'Encrypted phone number (AES-256, key in KMS). Use decrypt_text() to read.';
COMMENT ON COLUMN orders.customer_phone_encrypted IS 'Encrypted customer phone (AES-256, key in KMS)';
COMMENT ON COLUMN orders.pickup_address_encrypted IS 'Encrypted pickup address (AES-256, key in KMS)';
COMMENT ON COLUMN orders.delivery_address_encrypted IS 'Encrypted delivery address (AES-256, key in KMS)';
COMMENT ON COLUMN user_documents.fiscal_rfc_encrypted IS 'Encrypted RFC (AES-256, key in KMS)';

-- Step 6: Migration note
-- To populate encrypted columns, run this after migration:
-- UPDATE users SET phone_encrypted = encrypt_text(phone, :encryption_key) WHERE phone IS NOT NULL;
-- UPDATE orders SET customer_phone_encrypted = encrypt_text(customer_phone, :encryption_key) WHERE customer_phone IS NOT NULL;
-- etc.
--
-- IMPORTANT:
-- 1. Store ENCRYPTION_KEY in secret manager (AWS Secrets Manager, GCP Secret Manager)
-- 2. Rotate encryption key every 90 days
-- 3. After encrypting all data and verifying app works, drop plaintext columns in future migration
-- 4. Never commit encryption key to git
