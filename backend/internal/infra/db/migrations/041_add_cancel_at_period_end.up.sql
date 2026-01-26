-- Migration: 041_add_cancel_at_period_end
-- Description: サブスクリプションのキャンセル予約フラグを追加

-- cancel_at_period_end: キャンセル予約中かどうか
ALTER TABLE subscriptions ADD COLUMN IF NOT EXISTS cancel_at_period_end BOOLEAN NOT NULL DEFAULT FALSE;

-- cancel_at: キャンセル予定日時（Stripeから取得）
ALTER TABLE subscriptions ADD COLUMN IF NOT EXISTS cancel_at TIMESTAMP WITH TIME ZONE;
