-- イベント↔グループ割り当てテーブル
-- イベントに対して対象となるメンバーグループを複数割り当て可能
CREATE TABLE event_group_assignments (
    event_id VARCHAR(26) NOT NULL REFERENCES events(event_id) ON DELETE CASCADE,
    group_id VARCHAR(26) NOT NULL REFERENCES member_groups(group_id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    PRIMARY KEY (event_id, group_id)
);

-- インデックス
CREATE INDEX idx_event_group_event ON event_group_assignments(event_id);
CREATE INDEX idx_event_group_group ON event_group_assignments(group_id);

COMMENT ON TABLE event_group_assignments IS 'イベントへのグループ割り当て（対象メンバーを限定）';
COMMENT ON COLUMN event_group_assignments.event_id IS 'イベントID';
COMMENT ON COLUMN event_group_assignments.group_id IS 'グループID';
