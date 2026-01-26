-- Migration: 040_add_pending_payment_support
-- Description: Stripe Checkout用のpending_payment状態サポート

-- ============================================================
-- 1. tenants.status に pending_payment を追加
-- ============================================================
-- 既存のCHECK制約を削除して再作成
ALTER TABLE tenants DROP CONSTRAINT IF EXISTS tenants_status_check;

ALTER TABLE tenants ADD CONSTRAINT tenants_status_check
    CHECK (status IN ('active', 'grace', 'suspended', 'pending_payment'));

-- pending_expires_at カラム追加（Checkout Session有効期限）
ALTER TABLE tenants ADD COLUMN IF NOT EXISTS pending_expires_at TIMESTAMPTZ NULL;

-- pending_stripe_session_id カラム追加（Checkout Session ID）
ALTER TABLE tenants ADD COLUMN IF NOT EXISTS pending_stripe_session_id VARCHAR(100) NULL;

CREATE INDEX IF NOT EXISTS idx_tenants_pending_expires_at
    ON tenants(pending_expires_at)
    WHERE pending_expires_at IS NOT NULL;

COMMENT ON COLUMN tenants.pending_expires_at IS 'pending_payment状態の有効期限（Checkout Session期限）';
COMMENT ON COLUMN tenants.pending_stripe_session_id IS 'Stripe Checkout Session ID（pending中のみ）';

-- ============================================================
-- 2. plans テーブルに stripe_price_id を追加
-- ============================================================
ALTER TABLE plans ADD COLUMN IF NOT EXISTS stripe_price_id VARCHAR(100) NULL;

COMMENT ON COLUMN plans.stripe_price_id IS 'StripeのPrice ID（price_で始まる）';

-- ============================================================
-- 3. SUB_200 プランの追加・更新
-- ============================================================
-- 初期キャンペーン用の200円プランを追加（既存のSUB_980は残す）
INSERT INTO plans (plan_code, plan_type, display_name, price_jpy, stripe_price_id, features_json)
VALUES ('SUB_200', 'subscription', '月額プラン（初期キャンペーン）', 200, 'price_1SsMpa9hOZneOpjX3rPwqJoO', '{}')
ON CONFLICT (plan_code) DO UPDATE SET
    price_jpy = EXCLUDED.price_jpy,
    stripe_price_id = EXCLUDED.stripe_price_id,
    display_name = EXCLUDED.display_name,
    updated_at = CURRENT_TIMESTAMP;
