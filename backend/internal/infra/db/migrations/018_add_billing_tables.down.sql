-- Migration: 018_add_billing_tables (down)
-- Description: マネタイズ基盤用テーブルの削除

-- ============================================================
-- 7. billing_audit_logs テーブル削除
-- ============================================================
DROP TABLE IF EXISTS billing_audit_logs;

-- ============================================================
-- 6. license_keys テーブル削除
-- ============================================================
DROP TABLE IF EXISTS license_keys;

-- ============================================================
-- 5. webhook_events テーブル削除
-- ============================================================
DROP TABLE IF EXISTS webhook_events;

-- ============================================================
-- 4. subscriptions テーブル削除
-- ============================================================
DROP TABLE IF EXISTS subscriptions;

-- ============================================================
-- 3. entitlements テーブル削除
-- ============================================================
DROP TABLE IF EXISTS entitlements;

-- ============================================================
-- 2. plans テーブル削除
-- ============================================================
DROP TABLE IF EXISTS plans;

-- ============================================================
-- 1. tenants テーブルからカラム削除
-- ============================================================
DROP INDEX IF EXISTS idx_tenants_grace_until;
DROP INDEX IF EXISTS idx_tenants_status;

ALTER TABLE tenants DROP CONSTRAINT IF EXISTS tenants_status_check;
ALTER TABLE tenants DROP COLUMN IF EXISTS grace_until;
ALTER TABLE tenants DROP COLUMN IF EXISTS status;
