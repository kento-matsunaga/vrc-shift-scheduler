-- Migration: 011_create_invitations
-- Description: 管理者招待テーブルの作成

CREATE TABLE IF NOT EXISTS invitations (
    invitation_id CHAR(26) PRIMARY KEY,  -- ULID形式
    tenant_id CHAR(26) NOT NULL,
    email VARCHAR(255) NOT NULL,
    role VARCHAR(20) NOT NULL, -- 'owner' | 'manager'
    token VARCHAR(64) NOT NULL UNIQUE, -- セキュアランダムトークン
    created_by_admin_id CHAR(26) NOT NULL, -- 招待した管理者（必須）
    expires_at TIMESTAMPTZ NOT NULL,
    accepted_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_invitations_tenant FOREIGN KEY (tenant_id)
        REFERENCES tenants(tenant_id) ON DELETE CASCADE,
    CONSTRAINT fk_invitations_created_by FOREIGN KEY (created_by_admin_id)
        REFERENCES admins(admin_id) ON DELETE CASCADE,
    CONSTRAINT invitations_role_check CHECK (role IN ('owner', 'manager'))
);

-- トークンでの高速検索
CREATE UNIQUE INDEX idx_invitations_token
    ON invitations(token);

-- テナント×メールでの招待状況確認（未受理のみ）
CREATE INDEX idx_invitations_tenant_email
    ON invitations(tenant_id, email)
    WHERE accepted_at IS NULL;

-- テナント内の招待一覧検索用
CREATE INDEX idx_invitations_tenant
    ON invitations(tenant_id, created_at DESC);

COMMENT ON TABLE invitations IS '管理者招待: 招待者のテナントに自動紐付け';
COMMENT ON COLUMN invitations.created_by_admin_id IS '招待した管理者（このAdminのtenant_idに自動紐付け）';
COMMENT ON COLUMN invitations.token IS 'セキュアランダムトークン（64文字hex）';
COMMENT ON COLUMN invitations.expires_at IS '有効期限（デフォルト7日間）';
COMMENT ON COLUMN invitations.accepted_at IS '受理日時（NULL=未受理）';
