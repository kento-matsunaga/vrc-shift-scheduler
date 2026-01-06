-- Add start_time and end_time columns to attendance_target_dates table
ALTER TABLE attendance_target_dates
ADD COLUMN start_time TIME NULL,
ADD COLUMN end_time TIME NULL;

COMMENT ON COLUMN attendance_target_dates.start_time IS 'Optional start time for the target date (HH:MM:SS format)';
COMMENT ON COLUMN attendance_target_dates.end_time IS 'Optional end time for the target date (HH:MM:SS format)';
