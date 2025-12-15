-- 009_create_schedule_tables.up.sql

-- 日程調整テーブル
CREATE TABLE IF NOT EXISTS date_schedules (
    schedule_id CHAR(26) PRIMARY KEY,
    tenant_id CHAR(26) NOT NULL REFERENCES tenants(tenant_id),
    title VARCHAR(255) NOT NULL,
    description TEXT,
    event_id CHAR(26) REFERENCES events(event_id),
    public_token UUID NOT NULL UNIQUE,
    status VARCHAR(20) NOT NULL DEFAULT 'open',  -- 'open' | 'closed' | 'decided'
    deadline TIMESTAMPTZ,
    decided_candidate_id CHAR(26),  -- 確定した候補日ID（NULL許可）
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ NULL
);

CREATE INDEX idx_date_schedules_tenant ON date_schedules(tenant_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_date_schedules_token ON date_schedules(public_token);

-- 候補日テーブル
CREATE TABLE IF NOT EXISTS schedule_candidates (
    candidate_id CHAR(26) PRIMARY KEY,
    schedule_id CHAR(26) NOT NULL REFERENCES date_schedules(schedule_id) ON DELETE CASCADE,
    candidate_date DATE NOT NULL,
    start_time TIME,
    end_time TIME,
    display_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_schedule_candidates_schedule ON schedule_candidates(schedule_id);

-- 日程調整回答テーブル
CREATE TABLE IF NOT EXISTS schedule_responses (
    response_id CHAR(26) PRIMARY KEY,
    tenant_id CHAR(26) NOT NULL REFERENCES tenants(tenant_id),
    schedule_id CHAR(26) NOT NULL REFERENCES date_schedules(schedule_id) ON DELETE CASCADE,
    member_id CHAR(26) NOT NULL REFERENCES members(member_id),
    candidate_id CHAR(26) NOT NULL REFERENCES schedule_candidates(candidate_id) ON DELETE CASCADE,
    availability VARCHAR(20) NOT NULL,  -- 'available' | 'unavailable' | 'maybe'
    note TEXT,
    responded_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

    -- UNIQUE制約: 同一スケジュール×メンバー×候補日は1回答のみ（UPSERTで上書き）
    CONSTRAINT uq_schedule_response_member_candidate UNIQUE(schedule_id, member_id, candidate_id)
);

CREATE INDEX idx_schedule_responses_schedule ON schedule_responses(schedule_id);
CREATE INDEX idx_schedule_responses_member ON schedule_responses(member_id);
