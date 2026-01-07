-- Remove start_time and end_time columns from attendance_target_dates table
ALTER TABLE attendance_target_dates
DROP COLUMN IF EXISTS start_time,
DROP COLUMN IF EXISTS end_time;
