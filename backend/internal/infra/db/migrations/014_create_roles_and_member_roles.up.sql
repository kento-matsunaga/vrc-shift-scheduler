-- Migration: 014_create_roles_and_member_roles
-- Description: ロール（役割/属性）システムの導入
-- メンバーに複数のロールを付与できるようにする

-- 1. ロールテーブルを作成
CREATE TABLE roles (
    role_id CHAR(26) PRIMARY KEY,
    tenant_id CHAR(26) NOT NULL,
    name VARCHAR(100) NOT NULL,
    description VARCHAR(500),
    color VARCHAR(20),
    display_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP,

    -- 外部キー制約
    CONSTRAINT fk_roles_tenant
        FOREIGN KEY (tenant_id)
        REFERENCES tenants(tenant_id)
        ON DELETE CASCADE
);

-- 2. メンバーロール関連テーブル（Many-to-Many）を作成
CREATE TABLE member_roles (
    member_id CHAR(26) NOT NULL,
    role_id CHAR(26) NOT NULL,
    assigned_at TIMESTAMP NOT NULL DEFAULT NOW(),

    -- 複合主キー
    PRIMARY KEY (member_id, role_id),

    -- 外部キー制約
    CONSTRAINT fk_member_roles_member
        FOREIGN KEY (member_id)
        REFERENCES members(member_id)
        ON DELETE CASCADE,

    CONSTRAINT fk_member_roles_role
        FOREIGN KEY (role_id)
        REFERENCES roles(role_id)
        ON DELETE CASCADE
);

-- 3. インデックスを追加（パフォーマンス向上）
CREATE INDEX idx_roles_tenant_id ON roles(tenant_id);
CREATE INDEX idx_roles_display_order ON roles(tenant_id, display_order);
CREATE INDEX idx_roles_deleted_at ON roles(deleted_at);
CREATE INDEX idx_member_roles_role_id ON member_roles(role_id);

-- 4. テナント内でロール名は一意（削除されていないもののみ）
CREATE UNIQUE INDEX uq_role_name_per_tenant ON roles(tenant_id, name) WHERE deleted_at IS NULL;

-- 5. コメントを追加
COMMENT ON TABLE roles IS 'ロール（役割/属性）: メンバーに付与する役割や属性を管理';
COMMENT ON COLUMN roles.role_id IS 'ロールID: ULID形式の一意識別子';
COMMENT ON COLUMN roles.tenant_id IS 'テナントID: このロールが属するテナント';
COMMENT ON COLUMN roles.name IS 'ロール名: 表示用の名前（例: リーダー、サブリーダー、新人）';
COMMENT ON COLUMN roles.description IS 'ロール説明: このロールの詳細説明';
COMMENT ON COLUMN roles.color IS 'カラーコード: UI表示用の色（例: #FF5733）';
COMMENT ON COLUMN roles.display_order IS '表示順序: 小さい順に表示される';
COMMENT ON COLUMN roles.deleted_at IS '削除日時: ソフトデリート用';

COMMENT ON TABLE member_roles IS 'メンバーロール関連: メンバーとロールの多対多関連を管理';
COMMENT ON COLUMN member_roles.member_id IS 'メンバーID: ロールが付与されるメンバー';
COMMENT ON COLUMN member_roles.role_id IS 'ロールID: 付与されるロール';
COMMENT ON COLUMN member_roles.assigned_at IS '付与日時: このロールが付与された日時';
