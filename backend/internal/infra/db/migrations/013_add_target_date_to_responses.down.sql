-- Rollback: 013_add_target_date_to_responses

-- 1. インデックスを削除
DROP INDEX IF EXISTS idx_attendance_responses_target_date;

-- 2. 新しいUNIQUE制約を削除
ALTER TABLE attendance_responses
DROP CONSTRAINT IF EXISTS uq_attendance_response_member_target_date;

-- 3. 外部キー制約を削除
ALTER TABLE attendance_responses
DROP CONSTRAINT IF EXISTS fk_attendance_responses_target_date;

-- 4. target_date_id カラムを削除
ALTER TABLE attendance_responses
DROP COLUMN IF EXISTS target_date_id;

-- 5. 元のUNIQUE制約を復元
ALTER TABLE attendance_responses
ADD CONSTRAINT uq_attendance_response_member
UNIQUE(collection_id, member_id);
