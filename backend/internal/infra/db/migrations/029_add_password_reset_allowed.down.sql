-- Migration: 029_add_password_reset_allowed (rollback)

ALTER TABLE admins DROP CONSTRAINT IF EXISTS fk_admins_password_reset_allowed_by;
DROP INDEX IF EXISTS idx_admins_password_reset_allowed;
ALTER TABLE admins DROP COLUMN IF EXISTS password_reset_allowed_by;
ALTER TABLE admins DROP COLUMN IF EXISTS password_reset_allowed_at;
