-- ============================================================================
-- Migration Rollback: 001_create_users_table
-- Description: Reverts all changes made by the up migration
-- Author: User Service Team
-- Date: 2025-10-31
-- ============================================================================

-- ============================================================================
-- Drop Trigger
-- ============================================================================
DROP TRIGGER IF EXISTS update_users_updated_at ON users;

-- ============================================================================
-- Drop Indexes
-- ============================================================================
DROP INDEX IF EXISTS idx_users_created_at;
DROP INDEX IF EXISTS idx_users_email;

-- ============================================================================
-- Drop Table
-- ============================================================================
DROP TABLE IF EXISTS users;

-- ============================================================================
-- Drop Function
-- ============================================================================
DROP FUNCTION IF EXISTS update_updated_at_column();

-- ============================================================================
-- Drop Extension (Optional - only if not used by other tables)
-- ============================================================================
-- DROP EXTENSION IF EXISTS "pgcrypto";
-- Note: Commented out to prevent breaking other migrations that might use UUIDs

-- ============================================================================
-- Verification (for testing)
-- ============================================================================
-- Verify table was dropped
-- SELECT EXISTS (
--     SELECT FROM information_schema.tables 
--     WHERE table_schema = 'public' 
--     AND table_name = 'users'
-- );
