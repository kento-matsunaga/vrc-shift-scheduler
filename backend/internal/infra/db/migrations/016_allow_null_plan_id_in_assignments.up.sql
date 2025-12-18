-- Migration: 016_allow_null_plan_id_in_assignments
-- Description: shift_assignmentsのplan_idをNULL許可に変更（手動割り当て対応）

-- plan_idカラムをNULL許可に変更
ALTER TABLE shift_assignments
    ALTER COLUMN plan_id DROP NOT NULL;

-- 外部キー制約を削除して再作成（ON DELETE CASCADEからON DELETE SET NULLに変更）
ALTER TABLE shift_assignments
    DROP CONSTRAINT fk_shift_assignments_plan;

ALTER TABLE shift_assignments
    ADD CONSTRAINT fk_shift_assignments_plan
    FOREIGN KEY (plan_id)
    REFERENCES shift_plans(plan_id)
    ON DELETE SET NULL;

COMMENT ON COLUMN shift_assignments.plan_id IS 'シフト計画ID（手動割り当ての場合はNULL）';
