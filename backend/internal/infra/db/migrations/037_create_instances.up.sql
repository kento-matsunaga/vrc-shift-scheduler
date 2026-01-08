-- Migration: 037_create_instances
-- Description: インスタンスエンティティテーブルの作成（Issue #140）
-- インスタンス = イベント内で並行稼働するサーバー単位

CREATE TABLE IF NOT EXISTS instances (
    instance_id CHAR(26) PRIMARY KEY,  -- ULID形式
    tenant_id CHAR(26) NOT NULL,
    event_id CHAR(26) NOT NULL,
    name VARCHAR(255) NOT NULL,
    display_order INT NOT NULL DEFAULT 0,
    max_members INT NULL,  -- インスタンス最大人数（NULLは無制限）
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ NULL,

    CONSTRAINT fk_instances_tenant FOREIGN KEY (tenant_id)
        REFERENCES tenants(tenant_id) ON DELETE CASCADE,
    CONSTRAINT fk_instances_event FOREIGN KEY (event_id)
        REFERENCES events(event_id) ON DELETE CASCADE,

    -- インスタンス名は1文字以上
    CONSTRAINT instances_name_check CHECK (LENGTH(name) >= 1),
    -- max_members は1以上（設定されている場合）
    CONSTRAINT instances_max_members_check CHECK (max_members IS NULL OR max_members >= 1)
);

-- イベント内でインスタンス名を一意にする
CREATE UNIQUE INDEX idx_instances_event_name_unique
    ON instances(tenant_id, event_id, name)
    WHERE deleted_at IS NULL;

-- イベント内のインスタンス一覧取得用（表示順でソート）
CREATE INDEX idx_instances_event_order
    ON instances(tenant_id, event_id, display_order, name)
    WHERE deleted_at IS NULL;

COMMENT ON TABLE instances IS 'インスタンス: イベント内で並行稼働するサーバー単位（例: JP1, JP2など）';
COMMENT ON COLUMN instances.name IS 'インスタンス名（例: JP1, VRChat Japan 1）';
COMMENT ON COLUMN instances.display_order IS '表示順（小さいほど上位）';
COMMENT ON COLUMN instances.max_members IS 'インスタンス最大人数（NULLは無制限）';
