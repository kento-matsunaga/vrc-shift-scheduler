-- メンバーグループテーブル
CREATE TABLE member_groups (
    group_id VARCHAR(26) PRIMARY KEY,
    tenant_id VARCHAR(26) NOT NULL REFERENCES tenants(tenant_id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    color VARCHAR(7), -- #RRGGBB形式
    display_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- メンバー・グループ関連テーブル（多対多）
CREATE TABLE member_group_assignments (
    assignment_id VARCHAR(26) PRIMARY KEY,
    member_id VARCHAR(26) NOT NULL REFERENCES members(member_id) ON DELETE CASCADE,
    group_id VARCHAR(26) NOT NULL REFERENCES member_groups(group_id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE(member_id, group_id)
);

-- インデックス
CREATE INDEX idx_member_groups_tenant_id ON member_groups(tenant_id);
CREATE INDEX idx_member_groups_deleted_at ON member_groups(deleted_at);
CREATE INDEX idx_member_group_assignments_member_id ON member_group_assignments(member_id);
CREATE INDEX idx_member_group_assignments_group_id ON member_group_assignments(group_id);
