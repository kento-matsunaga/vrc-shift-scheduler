-- Migration: 013_add_target_date_to_responses
-- Description: 出欠回答を対象日ごとに分ける
-- 各メンバーが各対象日に対して個別に出欠を回答できるようにする

-- 1. 既存のUNIQUE制約を削除
ALTER TABLE attendance_responses
DROP CONSTRAINT IF EXISTS uq_attendance_response_member;

-- 2. target_date_id カラムを追加（NULL許可、後でNOT NULLに変更）
ALTER TABLE attendance_responses
ADD COLUMN target_date_id CHAR(26);

-- 3. 既存データがある場合のための一時的な処理
-- 既存の回答には最初の対象日を割り当てる（対象日が存在する場合のみ）
UPDATE attendance_responses ar
SET target_date_id = (
    SELECT atd.target_date_id
    FROM attendance_target_dates atd
    WHERE atd.collection_id = ar.collection_id
    ORDER BY atd.display_order, atd.target_date
    LIMIT 1
)
WHERE target_date_id IS NULL
  AND EXISTS (
    SELECT 1 FROM attendance_target_dates atd
    WHERE atd.collection_id = ar.collection_id
);

-- 対応する target_date が存在しない古い回答レコードは削除
DELETE FROM attendance_responses
WHERE target_date_id IS NULL;

-- 4. target_date_id を NOT NULL に変更
ALTER TABLE attendance_responses
ALTER COLUMN target_date_id SET NOT NULL;

-- 5. 外部キー制約を追加
ALTER TABLE attendance_responses
ADD CONSTRAINT fk_attendance_responses_target_date
FOREIGN KEY (target_date_id)
REFERENCES attendance_target_dates(target_date_id)
ON DELETE CASCADE;

-- 6. 新しいUNIQUE制約を追加（同一コレクション×メンバー×対象日は1回答のみ）
ALTER TABLE attendance_responses
ADD CONSTRAINT uq_attendance_response_member_target_date
UNIQUE(collection_id, member_id, target_date_id);

-- 7. インデックスを追加（パフォーマンス向上）
CREATE INDEX idx_attendance_responses_target_date
ON attendance_responses(target_date_id);

COMMENT ON COLUMN attendance_responses.target_date_id IS '対象日ID: 各対象日ごとに個別の回答を持つ';
