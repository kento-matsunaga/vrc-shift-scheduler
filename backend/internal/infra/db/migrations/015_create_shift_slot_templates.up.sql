-- シフトテンプレートのメタ情報テーブル
CREATE TABLE shift_slot_templates (
    template_id CHAR(26) PRIMARY KEY,
    tenant_id CHAR(26) NOT NULL,
    event_id CHAR(26) NOT NULL,
    template_name VARCHAR(100) NOT NULL,
    description VARCHAR(500),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP,
    CONSTRAINT fk_shift_slot_templates_tenant
        FOREIGN KEY (tenant_id)
        REFERENCES tenants(tenant_id)
        ON DELETE CASCADE,
    CONSTRAINT fk_shift_slot_templates_event
        FOREIGN KEY (event_id)
        REFERENCES events(event_id)
        ON DELETE CASCADE
);

-- テンプレート名のユニークインデックス（同じイベント内で重複不可、削除済みは除外）
CREATE UNIQUE INDEX uq_template_name_per_event
ON shift_slot_templates(event_id, template_name)
WHERE deleted_at IS NULL;

-- テンプレート内の個別シフト枠
CREATE TABLE shift_slot_template_items (
    item_id CHAR(26) PRIMARY KEY,
    template_id CHAR(26) NOT NULL,
    position_id CHAR(26) NOT NULL,
    slot_name VARCHAR(100) NOT NULL,
    instance_name VARCHAR(100) NOT NULL,
    start_time TIME NOT NULL,
    end_time TIME NOT NULL,
    required_count INTEGER NOT NULL DEFAULT 1,
    priority INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_shift_slot_template_items_template
        FOREIGN KEY (template_id)
        REFERENCES shift_slot_templates(template_id)
        ON DELETE CASCADE,
    CONSTRAINT fk_shift_slot_template_items_position
        FOREIGN KEY (position_id)
        REFERENCES positions(position_id)
        ON DELETE RESTRICT,
    CONSTRAINT chk_required_count_positive
        CHECK (required_count > 0)
);

-- インデックス
CREATE INDEX idx_shift_slot_templates_event ON shift_slot_templates(event_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_shift_slot_template_items_template ON shift_slot_template_items(template_id);
