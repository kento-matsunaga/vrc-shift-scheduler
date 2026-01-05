-- Migration: 030_add_attendance_time_specification (down)
-- Description: 出欠回答の参加可能時間帯を削除

ALTER TABLE attendance_responses
DROP COLUMN available_from,
DROP COLUMN available_to;
