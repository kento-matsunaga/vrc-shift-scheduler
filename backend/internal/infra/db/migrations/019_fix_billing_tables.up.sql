-- Migration: 019_fix_billing_tables
-- Description: 課金テーブルの修正（expires_at, memo追加、stripe_webhook_logs作成）

-- ============================================================
-- 1. license_keys テーブルへのカラム追加
-- ============================================================
ALTER TABLE license_keys ADD COLUMN IF NOT EXISTS expires_at TIMESTAMPTZ NULL;
ALTER TABLE license_keys ADD COLUMN IF NOT EXISTS memo TEXT NULL;

COMMENT ON COLUMN license_keys.expires_at IS '有効期限（NULL=無期限）';
COMMENT ON COLUMN license_keys.memo IS '管理メモ（発行バッチ識別など）';

-- ============================================================
-- 2. stripe_webhook_logs テーブル作成（バッチ処理用）
-- ============================================================
CREATE TABLE IF NOT EXISTS stripe_webhook_logs (
    id SERIAL PRIMARY KEY,
    event_id VARCHAR(100) NOT NULL UNIQUE,
    event_type VARCHAR(100) NOT NULL,
    payload_json JSONB NULL,
    processed_at TIMESTAMPTZ NULL,
    error_message TEXT NULL,
    received_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_stripe_webhook_logs_received_at ON stripe_webhook_logs(received_at);
CREATE INDEX IF NOT EXISTS idx_stripe_webhook_logs_event_type ON stripe_webhook_logs(event_type);

COMMENT ON TABLE stripe_webhook_logs IS 'Stripe Webhookログ: 受信履歴と処理状態';
COMMENT ON COLUMN stripe_webhook_logs.event_id IS 'Stripeイベント ID';
COMMENT ON COLUMN stripe_webhook_logs.event_type IS 'イベントタイプ（customer.subscription.created 等）';
COMMENT ON COLUMN stripe_webhook_logs.processed_at IS '処理完了日時';
COMMENT ON COLUMN stripe_webhook_logs.error_message IS 'エラーメッセージ（失敗時）';
