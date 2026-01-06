-- Migration: 035_remove_positions
-- Description: positions機能の廃止（Issue #69）
-- position_idカラムとpositionsテーブルを削除

-- 1. shift_slot_template_itemsからposition_idを削除
-- まず外部キー制約を削除
ALTER TABLE shift_slot_template_items DROP CONSTRAINT IF EXISTS fk_shift_slot_template_items_position;

-- position_idカラムを削除
ALTER TABLE shift_slot_template_items DROP COLUMN IF EXISTS position_id;

-- 2. shift_slotsからposition_idを削除
-- まず外部キー制約を削除
ALTER TABLE shift_slots DROP CONSTRAINT IF EXISTS fk_shift_slots_position;

-- インデックスを削除
DROP INDEX IF EXISTS idx_shift_slots_position;

-- position_idカラムを削除
ALTER TABLE shift_slots DROP COLUMN IF EXISTS position_id;

-- 3. positionsテーブルを削除
DROP TABLE IF EXISTS positions;
