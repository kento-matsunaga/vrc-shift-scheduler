-- Migration: 029_add_password_reset_allowed
-- Description: 管理者パスワードリセット許可カラムの追加

-- admins テーブルにカラム追加
ALTER TABLE admins ADD COLUMN IF NOT EXISTS password_reset_allowed_at TIMESTAMPTZ NULL;
ALTER TABLE admins ADD COLUMN IF NOT EXISTS password_reset_allowed_by CHAR(26) NULL;

-- インデックス: PWリセット許可中の管理者を効率的に検索
CREATE INDEX IF NOT EXISTS idx_admins_password_reset_allowed
    ON admins(password_reset_allowed_at)
    WHERE password_reset_allowed_at IS NOT NULL AND deleted_at IS NULL;

-- 外部キー: 許可した管理者への参照（任意、削除時はNULLに）
ALTER TABLE admins ADD CONSTRAINT fk_admins_password_reset_allowed_by
    FOREIGN KEY (password_reset_allowed_by) REFERENCES admins(admin_id)
    ON DELETE SET NULL;

COMMENT ON COLUMN admins.password_reset_allowed_at IS 'PWリセット許可日時（NULL=未許可、24時間有効）';
COMMENT ON COLUMN admins.password_reset_allowed_by IS 'PWリセットを許可した管理者ID';
