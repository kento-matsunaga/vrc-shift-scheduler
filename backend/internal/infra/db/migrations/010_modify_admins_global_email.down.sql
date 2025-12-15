-- Migration: 010_modify_admins_global_email (rollback)

-- グローバル一意制約を削除
DROP INDEX IF EXISTS uq_admins_email_global;
DROP INDEX IF EXISTS idx_admins_email_lookup;

-- 元のテナント内一意制約に戻す
ALTER TABLE admins ADD CONSTRAINT uq_admins_tenant_email UNIQUE(tenant_id, email);
CREATE INDEX idx_admins_email ON admins(tenant_id, email) WHERE deleted_at IS NULL;

COMMENT ON COLUMN admins.email IS 'メールアドレス（ログインID、テナント内一意）';
