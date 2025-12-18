-- Migration: 016_allow_null_plan_id_in_assignments (rollback)
-- Description: shift_assignmentsのplan_idをNOT NULLに戻す

-- 外部キー制約を削除
ALTER TABLE shift_assignments
    DROP CONSTRAINT fk_shift_assignments_plan;

-- plan_idがNULLのレコードを削除（ロールバック時のデータ整合性のため）
DELETE FROM shift_assignments WHERE plan_id IS NULL;

-- plan_idカラムをNOT NULLに戻す
ALTER TABLE shift_assignments
    ALTER COLUMN plan_id SET NOT NULL;

-- 外部キー制約を再作成（元の仕様に戻す）
ALTER TABLE shift_assignments
    ADD CONSTRAINT fk_shift_assignments_plan
    FOREIGN KEY (plan_id)
    REFERENCES shift_plans(plan_id)
    ON DELETE CASCADE;
