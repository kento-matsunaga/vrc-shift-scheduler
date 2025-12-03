-- Migration: 005_create_notification_tables (DOWN)
-- Description: 通知ログと通知テンプレートテーブルの削除

DROP TABLE IF EXISTS notification_templates CASCADE;
DROP TABLE IF EXISTS notification_logs CASCADE;

