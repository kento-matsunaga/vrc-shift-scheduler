-- Migration: 012_add_attendance_target_dates
-- Description: 出欠確認の対象日テーブルを追加

-- ============================================================
-- Attendance Target Dates テーブル
-- ============================================================
CREATE TABLE IF NOT EXISTS attendance_target_dates (
    target_date_id CHAR(26) PRIMARY KEY,  -- ULID形式
    collection_id CHAR(26) NOT NULL,
    target_date DATE NOT NULL,
    display_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_attendance_target_dates_collection FOREIGN KEY (collection_id)
        REFERENCES attendance_collections(collection_id) ON DELETE CASCADE
);

-- コレクションごとの対象日検索用
CREATE INDEX idx_attendance_target_dates_collection
    ON attendance_target_dates(collection_id, display_order);

COMMENT ON TABLE attendance_target_dates IS '出欠確認の対象日: 複数日に対応';
COMMENT ON COLUMN attendance_target_dates.target_date IS '対象日（日付のみ）';
COMMENT ON COLUMN attendance_target_dates.display_order IS '表示順序';
