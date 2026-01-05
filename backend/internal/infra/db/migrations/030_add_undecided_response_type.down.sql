-- Revert: Remove 'undecided' from attendance_responses response check constraint
-- Note: This will fail if there are existing 'undecided' responses

-- Drop updated constraint
ALTER TABLE attendance_responses
DROP CONSTRAINT IF EXISTS attendance_responses_response_check;

-- Restore original constraint
ALTER TABLE attendance_responses
ADD CONSTRAINT attendance_responses_response_check CHECK (
    response IN ('attending', 'absent')
);
