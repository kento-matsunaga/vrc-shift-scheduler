-- 古いカラムを復元
ALTER TABLE attendance_collections
ADD COLUMN target_group_id VARCHAR(26) REFERENCES member_groups(group_id) ON DELETE SET NULL;

ALTER TABLE date_schedules
ADD COLUMN target_group_id VARCHAR(26) REFERENCES member_groups(group_id) ON DELETE SET NULL;

-- 新テーブルから最初のグループだけ復元（複数ある場合は1つのみ）
UPDATE attendance_collections ac
SET target_group_id = (
    SELECT group_id FROM attendance_collection_group_assignments acga
    WHERE acga.collection_id = ac.collection_id
    LIMIT 1
);

UPDATE date_schedules ds
SET target_group_id = (
    SELECT group_id FROM date_schedule_group_assignments dsga
    WHERE dsga.schedule_id = ds.schedule_id
    LIMIT 1
);

-- インデックス
CREATE INDEX idx_attendance_collections_group ON attendance_collections(target_group_id) WHERE target_group_id IS NOT NULL;
CREATE INDEX idx_date_schedules_group ON date_schedules(target_group_id) WHERE target_group_id IS NOT NULL;

-- 新テーブルを削除
DROP TABLE IF EXISTS date_schedule_group_assignments;
DROP TABLE IF EXISTS attendance_collection_group_assignments;
