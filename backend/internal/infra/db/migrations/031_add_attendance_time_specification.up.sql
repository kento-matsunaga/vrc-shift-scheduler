-- Migration: 030_add_attendance_time_specification
-- Description: 出欠回答に参加可能時間帯を追加

-- 参加可能な開始時間と終了時間を追加
ALTER TABLE attendance_responses
ADD COLUMN available_from TIME NULL,
ADD COLUMN available_to TIME NULL;

COMMENT ON COLUMN attendance_responses.available_from IS '参加可能な開始時間（例: 18:00）';
COMMENT ON COLUMN attendance_responses.available_to IS '参加可能な終了時間（例: 22:00）';
