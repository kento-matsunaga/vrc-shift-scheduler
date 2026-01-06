-- Migration: 035_remove_positions (rollback)
-- Description: positions機能の復元

-- 1. positionsテーブルを再作成
CREATE TABLE IF NOT EXISTS positions (
    position_id CHAR(26) PRIMARY KEY,
    tenant_id CHAR(26) NOT NULL,
    position_name VARCHAR(255) NOT NULL,
    description TEXT,
    display_order INT NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ NULL,

    CONSTRAINT fk_positions_tenant FOREIGN KEY (tenant_id)
        REFERENCES tenants(tenant_id) ON DELETE CASCADE,
    CONSTRAINT positions_name_check CHECK (LENGTH(position_name) >= 1)
);

CREATE UNIQUE INDEX idx_positions_tenant_name_unique
    ON positions(tenant_id, position_name)
    WHERE deleted_at IS NULL;

CREATE INDEX idx_positions_tenant_order
    ON positions(tenant_id, display_order, position_name)
    WHERE deleted_at IS NULL;

-- 2. shift_slotsにposition_idカラムを追加
-- 注意: rollback時にはダミーのposition_idを作成する必要がある
ALTER TABLE shift_slots ADD COLUMN IF NOT EXISTS position_id CHAR(26);

-- 3. shift_slot_template_itemsにposition_idカラムを追加
ALTER TABLE shift_slot_template_items ADD COLUMN IF NOT EXISTS position_id CHAR(26);

-- 注意: rollback後にposition_idにデータを設定し、外部キー制約を追加する必要がある
-- このマイグレーションのrollbackは完全ではありません（データが失われるため）
