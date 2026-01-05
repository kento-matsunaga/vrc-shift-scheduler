-- Add 'undecided' to attendance_responses response check constraint
-- Issue #78: 出欠に「未定」選択肢を追加

-- Drop existing constraint
ALTER TABLE attendance_responses
DROP CONSTRAINT IF EXISTS attendance_responses_response_check;

-- Add updated constraint with 'undecided'
ALTER TABLE attendance_responses
ADD CONSTRAINT attendance_responses_response_check CHECK (
    response IN ('attending', 'absent', 'undecided')
);
