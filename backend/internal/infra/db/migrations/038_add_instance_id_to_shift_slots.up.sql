-- Migration: 038_add_instance_id_to_shift_slots
-- Description: shift_slotsテーブルにinstance_idカラムを追加（Issue #140）
-- 移行期間中はNULLableとし、既存データは後のマイグレーションで紐付ける

-- instance_id カラムを追加（NULLable: 既存データとの互換性のため）
ALTER TABLE shift_slots ADD COLUMN IF NOT EXISTS instance_id CHAR(26) NULL;

-- 外部キー制約を追加
ALTER TABLE shift_slots ADD CONSTRAINT fk_shift_slots_instance
    FOREIGN KEY (instance_id) REFERENCES instances(instance_id) ON DELETE RESTRICT;

-- インスタンス別のシフト枠検索用インデックス
CREATE INDEX idx_shift_slots_instance
    ON shift_slots(instance_id)
    WHERE deleted_at IS NULL;

COMMENT ON COLUMN shift_slots.instance_id IS 'インスタンスID（instancesテーブルへの参照）';
