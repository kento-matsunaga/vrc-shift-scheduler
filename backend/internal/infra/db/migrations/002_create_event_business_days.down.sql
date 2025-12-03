-- Migration: 002_create_event_business_days (DOWN)
-- Description: イベント営業日テーブルの削除

DROP TABLE IF EXISTS event_business_days CASCADE;

