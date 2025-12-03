-- Migration: 004_create_shift_plans_and_assignments (DOWN)
-- Description: シフト計画とシフト割り当てテーブルの削除

DROP TABLE IF EXISTS shift_assignments CASCADE;
DROP TABLE IF EXISTS shift_plans CASCADE;

