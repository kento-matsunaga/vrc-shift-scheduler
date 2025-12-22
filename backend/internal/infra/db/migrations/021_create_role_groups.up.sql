-- ロールグループテーブル
CREATE TABLE role_groups (
    group_id VARCHAR(26) PRIMARY KEY,
    tenant_id VARCHAR(26) NOT NULL REFERENCES tenants(tenant_id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    color VARCHAR(7),
    display_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- ロールとグループの関連テーブル（多対多）
CREATE TABLE role_group_assignments (
    assignment_id VARCHAR(26) PRIMARY KEY,
    role_id VARCHAR(26) NOT NULL REFERENCES roles(role_id) ON DELETE CASCADE,
    group_id VARCHAR(26) NOT NULL REFERENCES role_groups(group_id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE(role_id, group_id)
);

-- インデックス
CREATE INDEX idx_role_groups_tenant_id ON role_groups(tenant_id);
CREATE INDEX idx_role_groups_deleted_at ON role_groups(deleted_at);
CREATE INDEX idx_role_group_assignments_role_id ON role_group_assignments(role_id);
CREATE INDEX idx_role_group_assignments_group_id ON role_group_assignments(group_id);
