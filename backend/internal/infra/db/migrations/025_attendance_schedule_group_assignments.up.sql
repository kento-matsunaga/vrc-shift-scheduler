-- 出欠確認↔グループ割り当てテーブル
-- 出欠確認に対して対象となるメンバーグループを複数割り当て可能
CREATE TABLE attendance_collection_group_assignments (
    collection_id VARCHAR(26) NOT NULL REFERENCES attendance_collections(collection_id) ON DELETE CASCADE,
    group_id VARCHAR(26) NOT NULL REFERENCES member_groups(group_id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    PRIMARY KEY (collection_id, group_id)
);

-- インデックス
CREATE INDEX idx_acga_collection ON attendance_collection_group_assignments(collection_id);
CREATE INDEX idx_acga_group ON attendance_collection_group_assignments(group_id);

COMMENT ON TABLE attendance_collection_group_assignments IS '出欠確認へのグループ割り当て（対象メンバーを限定）';
COMMENT ON COLUMN attendance_collection_group_assignments.collection_id IS '出欠確認ID';
COMMENT ON COLUMN attendance_collection_group_assignments.group_id IS 'グループID';

-- 日程調整↔グループ割り当てテーブル
-- 日程調整に対して対象となるメンバーグループを複数割り当て可能
CREATE TABLE date_schedule_group_assignments (
    schedule_id VARCHAR(26) NOT NULL REFERENCES date_schedules(schedule_id) ON DELETE CASCADE,
    group_id VARCHAR(26) NOT NULL REFERENCES member_groups(group_id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    PRIMARY KEY (schedule_id, group_id)
);

-- インデックス
CREATE INDEX idx_dsga_schedule ON date_schedule_group_assignments(schedule_id);
CREATE INDEX idx_dsga_group ON date_schedule_group_assignments(group_id);

COMMENT ON TABLE date_schedule_group_assignments IS '日程調整へのグループ割り当て（対象メンバーを限定）';
COMMENT ON COLUMN date_schedule_group_assignments.schedule_id IS '日程調整ID';
COMMENT ON COLUMN date_schedule_group_assignments.group_id IS 'グループID';

-- 既存の target_group_id データを新テーブルに移行
INSERT INTO attendance_collection_group_assignments (collection_id, group_id, created_at)
SELECT collection_id, target_group_id, NOW()
FROM attendance_collections
WHERE target_group_id IS NOT NULL;

INSERT INTO date_schedule_group_assignments (schedule_id, group_id, created_at)
SELECT schedule_id, target_group_id, NOW()
FROM date_schedules
WHERE target_group_id IS NOT NULL;

-- 古いカラムを削除
ALTER TABLE attendance_collections DROP COLUMN target_group_id;
ALTER TABLE date_schedules DROP COLUMN target_group_id;
