-- Migration: 019_fix_billing_tables (down)
-- Description: 課金テーブル修正のロールバック

-- stripe_webhook_logs テーブルを削除
DROP TABLE IF EXISTS stripe_webhook_logs;

-- license_keys テーブルのカラムを削除
ALTER TABLE license_keys DROP COLUMN IF EXISTS expires_at;
ALTER TABLE license_keys DROP COLUMN IF EXISTS memo;
