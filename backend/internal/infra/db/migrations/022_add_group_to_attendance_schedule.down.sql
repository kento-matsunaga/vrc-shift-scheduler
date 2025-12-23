DROP INDEX IF EXISTS idx_date_schedules_group;
DROP INDEX IF EXISTS idx_attendance_collections_group;
ALTER TABLE date_schedules DROP COLUMN IF EXISTS target_group_id;
ALTER TABLE attendance_collections DROP COLUMN IF EXISTS target_group_id;
