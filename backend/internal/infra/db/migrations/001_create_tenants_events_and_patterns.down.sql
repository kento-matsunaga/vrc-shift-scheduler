-- Migration: 001_create_tenants_events_and_patterns (DOWN)
-- Description: テナント、イベント、定期パターンテーブルの削除

DROP TABLE IF EXISTS recurring_patterns CASCADE;
DROP TABLE IF EXISTS events CASCADE;
DROP TABLE IF EXISTS tenants CASCADE;

