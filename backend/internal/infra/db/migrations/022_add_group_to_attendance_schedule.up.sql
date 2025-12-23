-- 出欠確認にグループ指定を追加
ALTER TABLE attendance_collections
ADD COLUMN target_group_id VARCHAR(26) REFERENCES member_groups(group_id) ON DELETE SET NULL;

-- 日程調整にグループ指定を追加
ALTER TABLE date_schedules
ADD COLUMN target_group_id VARCHAR(26) REFERENCES member_groups(group_id) ON DELETE SET NULL;

-- インデックス
CREATE INDEX idx_attendance_collections_group ON attendance_collections(target_group_id) WHERE target_group_id IS NOT NULL;
CREATE INDEX idx_date_schedules_group ON date_schedules(target_group_id) WHERE target_group_id IS NOT NULL;

COMMENT ON COLUMN attendance_collections.target_group_id IS '対象グループID（NULL=全メンバー対象）';
COMMENT ON COLUMN date_schedules.target_group_id IS '対象グループID（NULL=全メンバー対象）';
