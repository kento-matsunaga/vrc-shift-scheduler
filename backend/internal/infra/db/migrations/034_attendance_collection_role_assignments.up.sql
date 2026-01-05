-- Migration: 034_attendance_collection_role_assignments
-- Description: 出欠確認の対象ロールを指定可能にする

-- 出欠確認↔ロール割り当てテーブル
-- 出欠確認に対して対象となるロールを複数割り当て可能
CREATE TABLE attendance_collection_role_assignments (
    collection_id VARCHAR(26) NOT NULL REFERENCES attendance_collections(collection_id) ON DELETE CASCADE,
    role_id VARCHAR(26) NOT NULL REFERENCES roles(role_id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    PRIMARY KEY (collection_id, role_id)
);

-- インデックス
CREATE INDEX idx_acra_collection ON attendance_collection_role_assignments(collection_id);
CREATE INDEX idx_acra_role ON attendance_collection_role_assignments(role_id);

COMMENT ON TABLE attendance_collection_role_assignments IS '出欠確認へのロール割り当て（対象メンバーをロールで限定）';
COMMENT ON COLUMN attendance_collection_role_assignments.collection_id IS '出欠確認ID';
COMMENT ON COLUMN attendance_collection_role_assignments.role_id IS 'ロールID';
