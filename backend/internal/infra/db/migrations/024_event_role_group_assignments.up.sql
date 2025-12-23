-- イベントとロールグループの関連付けテーブル
-- イベントに参加可能なロールグループを指定するための多対多テーブル
CREATE TABLE event_role_group_assignments (
    event_id VARCHAR(26) NOT NULL REFERENCES events(event_id) ON DELETE CASCADE,
    role_group_id VARCHAR(26) NOT NULL REFERENCES role_groups(group_id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    PRIMARY KEY (event_id, role_group_id)
);

-- インデックス
CREATE INDEX idx_event_role_group_event ON event_role_group_assignments(event_id);
CREATE INDEX idx_event_role_group_group ON event_role_group_assignments(role_group_id);
