-- Migration: 040_add_pending_payment_support (down)
-- Description: Stripe Checkout用のpending_payment状態サポートを削除

-- SUB_200 プランを削除
DELETE FROM plans WHERE plan_code = 'SUB_200';

-- stripe_price_id カラムを削除
ALTER TABLE plans DROP COLUMN IF EXISTS stripe_price_id;

-- pending関連カラムとインデックスを削除
DROP INDEX IF EXISTS idx_tenants_pending_expires_at;
ALTER TABLE tenants DROP COLUMN IF EXISTS pending_stripe_session_id;
ALTER TABLE tenants DROP COLUMN IF EXISTS pending_expires_at;

-- pending_payment状態のテナントをsuspendedに変更（CHECK制約変更前に必要）
UPDATE tenants SET status = 'suspended' WHERE status = 'pending_payment';

-- CHECK制約を元に戻す
ALTER TABLE tenants DROP CONSTRAINT IF EXISTS tenants_status_check;

ALTER TABLE tenants ADD CONSTRAINT tenants_status_check
    CHECK (status IN ('active', 'grace', 'suspended'));
