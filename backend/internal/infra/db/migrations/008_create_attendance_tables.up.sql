-- Migration: 008_create_attendance_tables
-- Description: 出欠確認テーブルの作成
-- MVP: 公開トークンによる回答収集、同一メンバーは上書き

-- ============================================================
-- Attendance Collections テーブル
-- ============================================================
CREATE TABLE IF NOT EXISTS attendance_collections (
    collection_id CHAR(26) PRIMARY KEY,  -- ULID形式
    tenant_id CHAR(26) NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    target_type VARCHAR(20) NOT NULL,  -- 'event' | 'business_day'
    target_id CHAR(26),                -- event_id または business_day_id（NULL許可）
    public_token UUID NOT NULL UNIQUE, -- UUID v4形式
    status VARCHAR(20) NOT NULL DEFAULT 'open',  -- 'open' | 'closed'
    deadline TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ NULL,

    CONSTRAINT fk_attendance_collections_tenant FOREIGN KEY (tenant_id)
        REFERENCES tenants(tenant_id) ON DELETE CASCADE,

    CONSTRAINT attendance_collections_status_check CHECK (
        status IN ('open', 'closed')
    ),

    CONSTRAINT attendance_collections_target_type_check CHECK (
        target_type IN ('event', 'business_day')
    )
);

-- テナントごとのコレクション一覧検索用
CREATE INDEX idx_attendance_collections_tenant
    ON attendance_collections(tenant_id) WHERE deleted_at IS NULL;

-- 公開トークンでの検索用
CREATE INDEX idx_attendance_collections_token
    ON attendance_collections(public_token);

-- ============================================================
-- Attendance Responses テーブル
-- ============================================================
CREATE TABLE IF NOT EXISTS attendance_responses (
    response_id CHAR(26) PRIMARY KEY,  -- ULID形式
    tenant_id CHAR(26) NOT NULL,
    collection_id CHAR(26) NOT NULL,
    member_id CHAR(26) NOT NULL,
    response VARCHAR(20) NOT NULL,  -- 'attending' | 'absent'
    note TEXT,
    responded_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_attendance_responses_tenant FOREIGN KEY (tenant_id)
        REFERENCES tenants(tenant_id) ON DELETE CASCADE,

    CONSTRAINT fk_attendance_responses_collection FOREIGN KEY (collection_id)
        REFERENCES attendance_collections(collection_id) ON DELETE CASCADE,

    CONSTRAINT fk_attendance_responses_member FOREIGN KEY (member_id)
        REFERENCES members(member_id) ON DELETE RESTRICT,

    -- ★ 重複回答防止: 同一コレクション×メンバーは1回答のみ（UPSERTで上書き）
    CONSTRAINT uq_attendance_response_member UNIQUE(collection_id, member_id),

    CONSTRAINT attendance_responses_response_check CHECK (
        response IN ('attending', 'absent')
    )
);

-- コレクションごとの回答一覧検索用
CREATE INDEX idx_attendance_responses_collection
    ON attendance_responses(collection_id);

-- メンバーごとの回答履歴検索用
CREATE INDEX idx_attendance_responses_member
    ON attendance_responses(tenant_id, member_id, responded_at DESC);

COMMENT ON TABLE attendance_collections IS '出欠確認コレクション: 公開トークンで回答を収集';
COMMENT ON COLUMN attendance_collections.target_type IS '対象種別（event=イベント, business_day=営業日）';
COMMENT ON COLUMN attendance_collections.public_token IS '公開URL用トークン（UUID v4）';
COMMENT ON COLUMN attendance_collections.status IS 'ステータス（open=回答受付中, closed=締切）';

COMMENT ON TABLE attendance_responses IS '出欠確認回答: 同一メンバーは上書き（UNIQUE制約）';
COMMENT ON COLUMN attendance_responses.response IS '回答（attending=出席, absent=欠席）';
