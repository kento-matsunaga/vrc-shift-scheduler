-- Migration: 008_create_attendance_tables (Rollback)
-- Description: 出欠確認テーブルの削除

DROP TABLE IF EXISTS attendance_responses;
DROP TABLE IF EXISTS attendance_collections;
