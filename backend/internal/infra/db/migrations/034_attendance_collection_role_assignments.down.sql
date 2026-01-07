-- Migration: 034_attendance_collection_role_assignments (DOWN)
-- Description: 出欠確認の対象ロール割り当てテーブルを削除

DROP TABLE IF EXISTS attendance_collection_role_assignments;
