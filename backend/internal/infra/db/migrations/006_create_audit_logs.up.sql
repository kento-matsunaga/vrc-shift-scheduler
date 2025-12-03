-- Migration: 006_create_audit_logs
-- Description: 監査ログテーブルの作成
-- 真のMVP: 重要操作（ShiftAssignment CREATE）のみ記録

-- ============================================================
-- Audit Logs テーブル
-- ============================================================
CREATE TABLE IF NOT EXISTS audit_logs (
    log_id CHAR(26) PRIMARY KEY,  -- ULID形式
    tenant_id CHAR(26) NOT NULL,
    entity_type VARCHAR(50) NOT NULL,
    entity_id CHAR(26) NOT NULL,
    action VARCHAR(20) NOT NULL,
    actor_id CHAR(26) NOT NULL,  -- 操作者（メンバーID）
    changed_data_before JSONB NULL,  -- 変更前データ（UPDATE/DELETEの場合）
    changed_data_after JSONB NULL,   -- 変更後データ（CREATE/UPDATEの場合）
    timestamp TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT fk_audit_logs_tenant FOREIGN KEY (tenant_id) 
        REFERENCES tenants(tenant_id) ON DELETE CASCADE,
    CONSTRAINT fk_audit_logs_actor FOREIGN KEY (actor_id) 
        REFERENCES members(member_id) ON DELETE RESTRICT,
    
    CONSTRAINT audit_logs_action_check CHECK (
        action IN ('CREATE', 'UPDATE', 'DELETE')
    ),
    
    CONSTRAINT audit_logs_entity_type_check CHECK (
        entity_type IN (
            'events',
            'recurring_patterns',
            'event_business_days',
            'shift_slots',
            'shift_plans',
            'shift_assignments',
            'members',
            'positions',
            'availabilities'
        )
    )
);

-- エンティティごとの変更履歴検索用
CREATE INDEX idx_audit_logs_tenant_entity 
    ON audit_logs(tenant_id, entity_type, entity_id, timestamp DESC);

-- 操作者ごとの履歴検索用
CREATE INDEX idx_audit_logs_tenant_actor 
    ON audit_logs(tenant_id, actor_id, timestamp DESC);

-- 時系列検索用
CREATE INDEX idx_audit_logs_timestamp 
    ON audit_logs(timestamp DESC);

-- 特定のアクションのみを検索する場合
CREATE INDEX idx_audit_logs_tenant_action 
    ON audit_logs(tenant_id, action, timestamp DESC);

COMMENT ON TABLE audit_logs IS '監査ログ: エンティティの変更履歴（真のMVP: 重要操作のみ記録）';
COMMENT ON COLUMN audit_logs.entity_type IS 'エンティティ種別（events, shift_assignments など）';
COMMENT ON COLUMN audit_logs.entity_id IS '対象エンティティのID';
COMMENT ON COLUMN audit_logs.action IS '操作種別（CREATE/UPDATE/DELETE）';
COMMENT ON COLUMN audit_logs.actor_id IS '操作者のメンバーID';
COMMENT ON COLUMN audit_logs.changed_data_before IS '変更前のデータ（JSONB、UPDATE/DELETEの場合）';
COMMENT ON COLUMN audit_logs.changed_data_after IS '変更後のデータ（JSONB、CREATE/UPDATEの場合）';

-- changed_data_before/after の例:
-- CREATE: before=null, after={"slot_id": "...", "member_id": "...", "status": "confirmed"}
-- UPDATE: before={"status": "confirmed"}, after={"status": "cancelled"}
-- DELETE: before={"slot_id": "...", "member_id": "..."}, after=null

