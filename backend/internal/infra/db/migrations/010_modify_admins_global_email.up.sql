-- Migration: 010_modify_admins_global_email
-- Description: メールアドレスをグローバル一意に変更（ログインID化）
-- 理由: email + password のみでログインできるようにするため

-- 1. 既存の制約を削除
DROP INDEX IF EXISTS idx_admins_email;
ALTER TABLE admins DROP CONSTRAINT IF EXISTS uq_admins_tenant_email;

-- 2. メールアドレスをグローバル一意にする（deleted_at IS NULL のみ）
CREATE UNIQUE INDEX uq_admins_email_global
    ON admins(email)
    WHERE deleted_at IS NULL;

-- 3. メールアドレスでの高速検索用インデックス
CREATE INDEX idx_admins_email_lookup
    ON admins(email)
    WHERE deleted_at IS NULL AND is_active = true;

COMMENT ON INDEX uq_admins_email_global IS 'メールアドレスはシステム全体で一意（ログインID）';
COMMENT ON COLUMN admins.email IS 'メールアドレス（ログインID、システム全体で一意）';
