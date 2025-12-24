-- Migration: 027_relax_recurring_pattern_constraint
-- Description: recurring タイプでも recurring_pattern_id が NULL を許可
-- イベント自体の定期設定から生成される営業日に対応

-- 既存のCHECK制約を削除
ALTER TABLE event_business_days
DROP CONSTRAINT IF EXISTS event_business_days_pattern_consistency_check;

-- 新しいCHECK制約を追加（recurringでもNULLを許可）
ALTER TABLE event_business_days
ADD CONSTRAINT event_business_days_pattern_consistency_check CHECK (
    occurrence_type IN ('recurring', 'special')
);

COMMENT ON COLUMN event_business_days.recurring_pattern_id IS '生成元の定期パターン（recurring_patternsテーブル使用時のみ、イベント自体の定期設定から生成時はNULL）';
