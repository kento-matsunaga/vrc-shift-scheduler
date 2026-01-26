-- Migration: 040_add_pending_payment_support (down)
-- Description: Stripe Checkout用のpending_payment状態サポートを削除
--
-- IMPORTANT: This migration handles the rollback of pending_payment status.
-- It converts any pending_payment tenants to suspended before removing the status.

BEGIN;

-- Step 1: pending_payment状態のテナントをsuspendedに変更（CHECK制約変更前に必要）
-- This must happen BEFORE altering the CHECK constraint
UPDATE tenants SET status = 'suspended' WHERE status = 'pending_payment';

-- Step 2: CHECK制約を元に戻す（pending_paymentを除外）
ALTER TABLE tenants DROP CONSTRAINT IF EXISTS tenants_status_check;
ALTER TABLE tenants ADD CONSTRAINT tenants_status_check
    CHECK (status IN ('active', 'grace', 'suspended'));

-- Step 3: pending関連カラムとインデックスを削除
DROP INDEX IF EXISTS idx_tenants_pending_expires_at;
ALTER TABLE tenants DROP COLUMN IF EXISTS pending_stripe_session_id;
ALTER TABLE tenants DROP COLUMN IF EXISTS pending_expires_at;

-- Step 4: stripe_price_id カラムを削除
ALTER TABLE plans DROP COLUMN IF EXISTS stripe_price_id;

-- Step 5: SUB_200 プランを削除
DELETE FROM plans WHERE plan_code = 'SUB_200';

COMMIT;
