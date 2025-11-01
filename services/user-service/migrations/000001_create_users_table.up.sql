-- ============================================================================
-- Migration: 001_create_users_table
-- Description: Creates users table with UUID primary key, email validation,
--              automatic updated_at trigger, and proper indexes
-- Author: User Service Team
-- Date: 2025-10-31
-- ============================================================================

-- Enable UUID extension if not already enabled
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ============================================================================
-- Function: update_updated_at_column
-- Description: Generic function to automatically update the updated_at column
--              whenever a row is modified
-- ============================================================================
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Add comment to function
COMMENT ON FUNCTION update_updated_at_column() IS 
'Trigger function that automatically updates the updated_at column to current timestamp on row update';

-- ============================================================================
-- Table: users
-- Description: Stores user account information including authentication data
-- ============================================================================
CREATE TABLE IF NOT EXISTS users (
    -- Primary key: Auto-generated UUID
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- Authentication fields
    email VARCHAR(255) NOT NULL UNIQUE 
        CHECK (email = LOWER(email) AND email ~* '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$'),
    password VARCHAR(255) NOT NULL 
        CHECK (LENGTH(password) >= 60 AND password ~ '^\$2[aby]\$'),
    
    -- User information
    name VARCHAR(255) NOT NULL 
        CHECK (LENGTH(TRIM(name)) >= 2),
    phone VARCHAR(20) 
        CHECK (phone IS NULL OR (LENGTH(phone) >= 10 AND LENGTH(phone) <= 15 AND phone ~ '^[0-9]+$')),
    
    -- Audit fields
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- ============================================================================
-- Indexes
-- ============================================================================

-- Index on email for fast user lookup during login
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);

-- Index on created_at for sorting users by registration date (descending)
CREATE INDEX IF NOT EXISTS idx_users_created_at ON users(created_at DESC);

-- Partial index for active sessions (if needed in the future)
-- CREATE INDEX IF NOT EXISTS idx_users_active ON users(id) WHERE deleted_at IS NULL;

-- ============================================================================
-- Trigger: update_users_updated_at
-- Description: Automatically updates updated_at when a user row is modified
-- ============================================================================
CREATE TRIGGER update_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- ============================================================================
-- Comments (Documentation)
-- ============================================================================

-- Table comment
COMMENT ON TABLE users IS 
'Stores user account information for the Pedidos Online system. Includes authentication credentials, personal information, and audit timestamps.';

-- Column comments
COMMENT ON COLUMN users.id IS 
'Unique identifier for the user (UUID v4). Auto-generated using gen_random_uuid().';

COMMENT ON COLUMN users.email IS 
'User email address (unique, lowercase). Used for authentication and communication. Must be a valid email format.';

COMMENT ON COLUMN users.password IS 
'Bcrypt hashed password (60+ characters). Never store plaintext passwords. Uses bcrypt with cost factor 10.';

COMMENT ON COLUMN users.name IS 
'User full name (2-255 characters). Required for personalization and identification.';

COMMENT ON COLUMN users.phone IS 
'User phone number (10-15 digits, numeric only). Optional field for contact and 2FA. NULL allowed.';

COMMENT ON COLUMN users.created_at IS 
'Timestamp when the user account was created (UTC with timezone). Immutable after creation.';

COMMENT ON COLUMN users.updated_at IS 
'Timestamp when the user account was last updated (UTC with timezone). Automatically updated by trigger on any modification.';

-- Index comments
COMMENT ON INDEX idx_users_email IS 
'B-tree index on email column for fast user lookup during authentication (O(log n) complexity).';

COMMENT ON INDEX idx_users_created_at IS 
'B-tree index on created_at (descending) for efficient sorting of users by registration date.';

-- ============================================================================
-- Initial Data (Optional)
-- ============================================================================

-- Uncomment to create a default admin user (change password in production!)
-- INSERT INTO users (email, password, name, phone, created_at, updated_at)
-- VALUES (
--     'admin@pedidosonline.com',
--     '$2a$10$examplehashedpassword1234567890123456789012345678901234', -- Change this!
--     'System Administrator',
--     '11999999999',
--     CURRENT_TIMESTAMP,
--     CURRENT_TIMESTAMP
-- )
-- ON CONFLICT (email) DO NOTHING;

-- ============================================================================
-- Verification Queries (for testing)
-- ============================================================================

-- Verify table structure
-- SELECT column_name, data_type, character_maximum_length, is_nullable, column_default
-- FROM information_schema.columns
-- WHERE table_name = 'users'
-- ORDER BY ordinal_position;

-- Verify indexes
-- SELECT indexname, indexdef
-- FROM pg_indexes
-- WHERE tablename = 'users';

-- Verify triggers
-- SELECT trigger_name, event_manipulation, event_object_table, action_statement
-- FROM information_schema.triggers
-- WHERE event_object_table = 'users';
