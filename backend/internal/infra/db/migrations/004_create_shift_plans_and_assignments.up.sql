-- Migration: 004_create_shift_plans_and_assignments
-- Description: シフト計画とシフト割り当てテーブルの作成

-- ============================================================
-- 1. Shift Plans テーブル（真のMVP版: 簡易実装）
-- ============================================================
CREATE TABLE IF NOT EXISTS shift_plans (
    plan_id CHAR(26) PRIMARY KEY,  -- ULID形式
    tenant_id CHAR(26) NOT NULL,
    event_id CHAR(26) NOT NULL,
    plan_name VARCHAR(255) NOT NULL,
    plan_status VARCHAR(20) NOT NULL DEFAULT 'draft',
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ NULL,
    
    CONSTRAINT fk_shift_plans_tenant FOREIGN KEY (tenant_id) 
        REFERENCES tenants(tenant_id) ON DELETE CASCADE,
    CONSTRAINT fk_shift_plans_event FOREIGN KEY (event_id) 
        REFERENCES events(event_id) ON DELETE CASCADE,
    CONSTRAINT shift_plans_status_check CHECK (
        plan_status IN ('draft', 'published', 'finalized')
    ),
    CONSTRAINT shift_plans_name_check CHECK (LENGTH(plan_name) >= 1)
);

-- テナント内のイベント単位でプラン検索
CREATE INDEX idx_shift_plans_tenant_event 
    ON shift_plans(tenant_id, event_id) 
    WHERE deleted_at IS NULL;

COMMENT ON TABLE shift_plans IS 'シフト計画: シフト割り当ての集約ルート（MVP: 簡易実装）';
COMMENT ON COLUMN shift_plans.plan_status IS 'プラン状態: draft（下書き）、published（公開）、finalized（確定）';

-- ============================================================
-- 2. Shift Assignments テーブル
-- ============================================================
CREATE TABLE IF NOT EXISTS shift_assignments (
    assignment_id CHAR(26) PRIMARY KEY,  -- ULID形式
    tenant_id CHAR(26) NOT NULL,
    plan_id CHAR(26) NOT NULL,
    slot_id CHAR(26) NOT NULL,
    member_id CHAR(26) NOT NULL,
    assignment_status VARCHAR(20) NOT NULL DEFAULT 'confirmed',
    assignment_method VARCHAR(20) NOT NULL DEFAULT 'manual',
    is_outside_preference BOOLEAN NOT NULL DEFAULT false,
    assigned_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    cancelled_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ NULL,
    
    CONSTRAINT fk_shift_assignments_tenant FOREIGN KEY (tenant_id) 
        REFERENCES tenants(tenant_id) ON DELETE CASCADE,
    CONSTRAINT fk_shift_assignments_plan FOREIGN KEY (plan_id) 
        REFERENCES shift_plans(plan_id) ON DELETE CASCADE,
    CONSTRAINT fk_shift_assignments_slot FOREIGN KEY (slot_id) 
        REFERENCES shift_slots(slot_id) ON DELETE CASCADE,
    CONSTRAINT fk_shift_assignments_member FOREIGN KEY (member_id) 
        REFERENCES members(member_id) ON DELETE RESTRICT,
    
    CONSTRAINT shift_assignments_status_check CHECK (
        assignment_status IN ('confirmed', 'cancelled')
    ),
    
    CONSTRAINT shift_assignments_method_check CHECK (
        assignment_method IN ('auto', 'manual')
    ),
    
    -- cancelled の場合は cancelled_at が必須
    CONSTRAINT shift_assignments_cancelled_consistency CHECK (
        (assignment_status = 'cancelled' AND cancelled_at IS NOT NULL) OR
        (assignment_status = 'confirmed' AND cancelled_at IS NULL)
    )
);

-- 部分一意インデックス: 同じ枠に同じメンバーを重複確定させない
-- （履歴管理のため、assignment_status = 'confirmed' のもののみ一意制約）
CREATE UNIQUE INDEX idx_shift_assignments_slot_member_confirmed_unique 
    ON shift_assignments(slot_id, member_id, assignment_status) 
    WHERE assignment_status = 'confirmed' AND deleted_at IS NULL;

-- メンバーの確定済みシフト検索用
CREATE INDEX idx_shift_assignments_tenant_member_status 
    ON shift_assignments(tenant_id, member_id, assignment_status) 
    WHERE deleted_at IS NULL;

-- シフト枠の充足状況確認用
CREATE INDEX idx_shift_assignments_slot_status 
    ON shift_assignments(slot_id, assignment_status) 
    WHERE deleted_at IS NULL;

-- プランに紐づく割り当て検索用
CREATE INDEX idx_shift_assignments_plan 
    ON shift_assignments(plan_id) 
    WHERE deleted_at IS NULL;

-- 割り当て日時順ソート用
CREATE INDEX idx_shift_assignments_assigned_at 
    ON shift_assignments(assigned_at DESC) 
    WHERE deleted_at IS NULL;

COMMENT ON TABLE shift_assignments IS 'シフト割り当て: ShiftPlan 集約内のエンティティ';
COMMENT ON COLUMN shift_assignments.assignment_status IS '割り当て状態: confirmed（確定）、cancelled（キャンセル）';
COMMENT ON COLUMN shift_assignments.assignment_method IS '割り当て方法: auto（自動）、manual（手動）';
COMMENT ON COLUMN shift_assignments.is_outside_preference IS '希望外割り当てフラグ（true: メンバーの希望外）';
COMMENT ON COLUMN shift_assignments.cancelled_at IS 'キャンセル日時（assignment_status=cancelled の場合のみ）';

