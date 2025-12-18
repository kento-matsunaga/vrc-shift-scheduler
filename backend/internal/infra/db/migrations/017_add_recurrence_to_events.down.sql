-- Migration: 017_add_recurrence_to_events (rollback)
-- Description: イベントテーブルから定期設定フィールドを削除

-- 制約を削除
ALTER TABLE events
    DROP CONSTRAINT IF EXISTS events_recurrence_fields_check,
    DROP CONSTRAINT IF EXISTS events_recurrence_day_of_week_check,
    DROP CONSTRAINT IF EXISTS events_recurrence_type_check;

-- カラムを削除
ALTER TABLE events
    DROP COLUMN IF EXISTS recurrence_type,
    DROP COLUMN IF EXISTS recurrence_start_date,
    DROP COLUMN IF EXISTS recurrence_day_of_week,
    DROP COLUMN IF EXISTS default_start_time,
    DROP COLUMN IF EXISTS default_end_time;
