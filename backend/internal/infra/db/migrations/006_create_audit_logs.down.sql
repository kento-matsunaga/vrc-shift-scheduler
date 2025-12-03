-- Migration: 006_create_audit_logs (DOWN)
-- Description: 監査ログテーブルの削除

DROP TABLE IF EXISTS audit_logs CASCADE;

