-- Rollback: 027_relax_recurring_pattern_constraint

-- 新しい制約を削除
ALTER TABLE event_business_days
DROP CONSTRAINT IF EXISTS event_business_days_pattern_consistency_check;

-- 元の制約を復元
ALTER TABLE event_business_days
ADD CONSTRAINT event_business_days_pattern_consistency_check CHECK (
    (occurrence_type = 'recurring' AND recurring_pattern_id IS NOT NULL) OR
    (occurrence_type = 'special' AND recurring_pattern_id IS NULL)
);
