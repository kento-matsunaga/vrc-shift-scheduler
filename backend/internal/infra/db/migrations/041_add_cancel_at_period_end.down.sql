-- Migration: 041_add_cancel_at_period_end (down)
-- Description: サブスクリプションのキャンセル予約フラグを削除

ALTER TABLE subscriptions DROP COLUMN IF EXISTS cancel_at;
ALTER TABLE subscriptions DROP COLUMN IF EXISTS cancel_at_period_end;
