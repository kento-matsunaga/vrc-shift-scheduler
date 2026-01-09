-- Migration: 037_create_instances (down)
-- Description: インスタンステーブルを削除

-- インデックスを削除
DROP INDEX IF EXISTS idx_instances_event_name_unique;
DROP INDEX IF EXISTS idx_instances_event_order;

-- テーブルを削除
DROP TABLE IF EXISTS instances;
