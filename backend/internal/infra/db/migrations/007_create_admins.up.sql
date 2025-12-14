-- Migration: 007_create_admins
-- Description: 管理者（店長/副店長）認証テーブルの作成
-- MVP: email + password でログイン、テナント境界を守る

-- ============================================================
-- Admins テーブル
-- ============================================================
CREATE TABLE IF NOT EXISTS admins (
    admin_id CHAR(26) PRIMARY KEY,  -- ULID形式
    tenant_id CHAR(26) NOT NULL,
    email VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    display_name VARCHAR(255) NOT NULL,
    role VARCHAR(20) NOT NULL DEFAULT 'manager',  -- 'owner' | 'manager'
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ NULL,

    CONSTRAINT fk_admins_tenant FOREIGN KEY (tenant_id)
        REFERENCES tenants(tenant_id) ON DELETE CASCADE,

    CONSTRAINT uq_admins_tenant_email UNIQUE(tenant_id, email),

    CONSTRAINT admins_role_check CHECK (
        role IN ('owner', 'manager')
    )
);

-- テナントごとの管理者一覧検索用（アクティブのみ）
CREATE INDEX idx_admins_tenant
    ON admins(tenant_id) WHERE deleted_at IS NULL;

-- メールアドレスでの検索用
CREATE INDEX idx_admins_email
    ON admins(tenant_id, email) WHERE deleted_at IS NULL;

COMMENT ON TABLE admins IS '管理者（店長/副店長）: テナント内の管理操作を行う権限を持つユーザー';
COMMENT ON COLUMN admins.admin_id IS '管理者ID（ULID）';
COMMENT ON COLUMN admins.tenant_id IS 'テナントID';
COMMENT ON COLUMN admins.email IS 'メールアドレス（ログインID、テナント内一意）';
COMMENT ON COLUMN admins.password_hash IS 'bcryptハッシュ化されたパスワード';
COMMENT ON COLUMN admins.display_name IS '表示名（例: 店長 田中）';
COMMENT ON COLUMN admins.role IS 'ロール（owner=店長, manager=副店長）';
COMMENT ON COLUMN admins.is_active IS 'アクティブフラグ（false: ログイン不可）';
COMMENT ON COLUMN admins.deleted_at IS '削除日時（論理削除）';
