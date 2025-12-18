-- Migration: 017_add_recurrence_to_events
-- Description: イベントテーブルに定期設定フィールドを追加

-- eventsテーブルに定期設定カラムを追加
ALTER TABLE events
    ADD COLUMN recurrence_type VARCHAR(20) NOT NULL DEFAULT 'none',
    ADD COLUMN recurrence_start_date DATE,
    ADD COLUMN recurrence_day_of_week INT,
    ADD COLUMN default_start_time TIME,
    ADD COLUMN default_end_time TIME;

-- recurrence_typeのチェック制約
ALTER TABLE events
    ADD CONSTRAINT events_recurrence_type_check
    CHECK (recurrence_type IN ('none', 'weekly', 'biweekly'));

-- recurrence_day_of_weekのチェック制約（0-6: 日曜日=0, 土曜日=6）
ALTER TABLE events
    ADD CONSTRAINT events_recurrence_day_of_week_check
    CHECK (recurrence_day_of_week IS NULL OR (recurrence_day_of_week >= 0 AND recurrence_day_of_week <= 6));

-- 定期設定は任意: recurrence_typeが'none'の場合は他のフィールドはNULLであるべき
-- 'weekly'または'biweekly'の場合は、関連フィールドを設定することを推奨（強制はしない）

COMMENT ON COLUMN events.recurrence_type IS '定期タイプ: none（定期なし）、weekly（毎週）、biweekly（隔週）';
COMMENT ON COLUMN events.recurrence_start_date IS '定期開始日（定期設定がある場合）';
COMMENT ON COLUMN events.recurrence_day_of_week IS '定期曜日（0-6: 日曜日=0, 土曜日=6）';
COMMENT ON COLUMN events.default_start_time IS 'デフォルト開始時刻';
COMMENT ON COLUMN events.default_end_time IS 'デフォルト終了時刻';
