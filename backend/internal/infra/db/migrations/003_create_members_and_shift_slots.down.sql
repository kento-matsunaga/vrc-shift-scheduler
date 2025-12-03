-- Migration: 003_create_members_and_shift_slots (DOWN)
-- Description: メンバー、役職、シフト枠テーブルの削除

DROP TABLE IF EXISTS shift_slots CASCADE;
DROP TABLE IF EXISTS positions CASCADE;
DROP TABLE IF EXISTS members CASCADE;

