-- パスワードリセットトークンテーブルの作成
CREATE TABLE password_reset_tokens (
    token_id CHAR(26) PRIMARY KEY,
    admin_id CHAR(26) NOT NULL REFERENCES admins(admin_id) ON DELETE CASCADE,
    token VARCHAR(64) NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    used_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 未使用トークンに対するユニークインデックス（有効なトークンは一意）
CREATE UNIQUE INDEX idx_password_reset_tokens_token_unused
    ON password_reset_tokens(token) WHERE used_at IS NULL;

-- 管理者IDによる検索用インデックス
CREATE INDEX idx_password_reset_tokens_admin_id
    ON password_reset_tokens(admin_id);

-- 有効期限による検索用インデックス（期限切れトークンのクリーンアップ用）
CREATE INDEX idx_password_reset_tokens_expires_at
    ON password_reset_tokens(expires_at);

COMMENT ON TABLE password_reset_tokens IS 'パスワードリセットトークン';
COMMENT ON COLUMN password_reset_tokens.token_id IS 'トークンID（ULID）';
COMMENT ON COLUMN password_reset_tokens.admin_id IS '管理者ID';
COMMENT ON COLUMN password_reset_tokens.token IS 'リセットトークン（64文字のhex）';
COMMENT ON COLUMN password_reset_tokens.expires_at IS '有効期限';
COMMENT ON COLUMN password_reset_tokens.used_at IS '使用日時（NULLは未使用）';
COMMENT ON COLUMN password_reset_tokens.created_at IS '作成日時';
