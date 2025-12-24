-- マネージャー権限設定テーブル
-- テナントごとにマネージャーが実行可能な操作を設定

CREATE TABLE manager_permissions (
    tenant_id VARCHAR(26) PRIMARY KEY REFERENCES tenants(tenant_id) ON DELETE CASCADE,

    -- メンバー管理
    can_add_member BOOLEAN NOT NULL DEFAULT true,
    can_edit_member BOOLEAN NOT NULL DEFAULT true,
    can_delete_member BOOLEAN NOT NULL DEFAULT false,

    -- イベント管理
    can_create_event BOOLEAN NOT NULL DEFAULT true,
    can_edit_event BOOLEAN NOT NULL DEFAULT true,
    can_delete_event BOOLEAN NOT NULL DEFAULT false,

    -- シフト管理
    can_assign_shift BOOLEAN NOT NULL DEFAULT true,
    can_edit_shift BOOLEAN NOT NULL DEFAULT true,

    -- 出欠確認・日程調整
    can_create_attendance BOOLEAN NOT NULL DEFAULT true,
    can_create_schedule BOOLEAN NOT NULL DEFAULT true,

    -- 設定管理
    can_manage_roles BOOLEAN NOT NULL DEFAULT false,
    can_manage_positions BOOLEAN NOT NULL DEFAULT false,
    can_manage_groups BOOLEAN NOT NULL DEFAULT false,

    -- 管理者招待
    can_invite_manager BOOLEAN NOT NULL DEFAULT false,

    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- インデックス
CREATE INDEX idx_manager_permissions_tenant ON manager_permissions(tenant_id);

-- コメント
COMMENT ON TABLE manager_permissions IS 'テナントごとのマネージャー権限設定';
COMMENT ON COLUMN manager_permissions.can_add_member IS 'メンバー追加権限';
COMMENT ON COLUMN manager_permissions.can_edit_member IS 'メンバー編集権限';
COMMENT ON COLUMN manager_permissions.can_delete_member IS 'メンバー削除権限';
COMMENT ON COLUMN manager_permissions.can_create_event IS 'イベント作成権限';
COMMENT ON COLUMN manager_permissions.can_edit_event IS 'イベント編集権限';
COMMENT ON COLUMN manager_permissions.can_delete_event IS 'イベント削除権限';
COMMENT ON COLUMN manager_permissions.can_assign_shift IS 'シフト割り当て権限';
COMMENT ON COLUMN manager_permissions.can_edit_shift IS 'シフト編集権限';
COMMENT ON COLUMN manager_permissions.can_create_attendance IS '出欠確認作成権限';
COMMENT ON COLUMN manager_permissions.can_create_schedule IS '日程調整作成権限';
COMMENT ON COLUMN manager_permissions.can_manage_roles IS 'ロール管理権限';
COMMENT ON COLUMN manager_permissions.can_manage_positions IS 'ポジション管理権限';
COMMENT ON COLUMN manager_permissions.can_manage_groups IS 'グループ管理権限';
COMMENT ON COLUMN manager_permissions.can_invite_manager IS 'マネージャー招待権限';
