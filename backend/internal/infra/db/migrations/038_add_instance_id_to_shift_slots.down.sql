-- Migration: 038_add_instance_id_to_shift_slots (down)
-- Description: shift_slotsテーブルからinstance_idカラムを削除

-- インデックスを削除
DROP INDEX IF EXISTS idx_shift_slots_instance;

-- 外部キー制約を削除
ALTER TABLE shift_slots DROP CONSTRAINT IF EXISTS fk_shift_slots_instance;

-- カラムを削除
ALTER TABLE shift_slots DROP COLUMN IF EXISTS instance_id;
