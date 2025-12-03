-- Migration: 005_create_notification_tables
-- Description: 通知ログと通知テンプレートテーブルの作成
-- 真のMVP: テーブルは作成するが、実際の利用は最小限（ログ出力stub）

-- ============================================================
-- 1. Notification Logs テーブル（通知送信履歴）
-- ============================================================
CREATE TABLE IF NOT EXISTS notification_logs (
    log_id CHAR(26) PRIMARY KEY,  -- ULID形式
    tenant_id CHAR(26) NOT NULL,
    business_day_id CHAR(26) NULL,  -- 営業日関連通知の場合のみ
    recipient_id CHAR(26) NOT NULL,  -- 送信先メンバー
    notification_type VARCHAR(50) NOT NULL,
    message_content TEXT NOT NULL,
    delivery_channel VARCHAR(20) NOT NULL,
    delivery_status VARCHAR(20) NOT NULL DEFAULT 'pending',
    error_message TEXT NULL,
    sent_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT fk_notification_logs_tenant FOREIGN KEY (tenant_id) 
        REFERENCES tenants(tenant_id) ON DELETE CASCADE,
    CONSTRAINT fk_notification_logs_business_day FOREIGN KEY (business_day_id) 
        REFERENCES event_business_days(business_day_id) ON DELETE SET NULL,
    CONSTRAINT fk_notification_logs_recipient FOREIGN KEY (recipient_id) 
        REFERENCES members(member_id) ON DELETE CASCADE,
    
    CONSTRAINT notification_logs_type_check CHECK (
        notification_type IN (
            'shift_recruitment',      -- シフト募集
            'deadline_reminder',       -- 締切リマインダー
            'shift_confirmed',         -- シフト確定通知
            'attendance_reminder',     -- 出勤リマインダー
            'urgent_help'              -- 緊急ヘルプ要請
        )
    ),
    
    CONSTRAINT notification_logs_channel_check CHECK (
        delivery_channel IN ('Discord', 'Email', 'WebPush')
    ),
    
    CONSTRAINT notification_logs_status_check CHECK (
        delivery_status IN ('success', 'failed', 'pending')
    ),
    
    -- 送信成功/失敗の場合は sent_at が必須
    CONSTRAINT notification_logs_sent_consistency CHECK (
        (delivery_status IN ('success', 'failed') AND sent_at IS NOT NULL) OR
        (delivery_status = 'pending' AND sent_at IS NULL)
    )
);

-- FrequencyControl 用の必須インデックス
-- 「過去N分以内に同一recipientへの通知が何件あるか」を高速に数える
CREATE INDEX idx_notification_logs_recipient_sent_at 
    ON notification_logs(recipient_id, sent_at DESC) 
    WHERE sent_at IS NOT NULL;

-- 営業日ごとの通知履歴検索用
CREATE INDEX idx_notification_logs_tenant_business_day_type 
    ON notification_logs(tenant_id, business_day_id, notification_type) 
    WHERE business_day_id IS NOT NULL;

-- 通知種別ごとの履歴検索用
CREATE INDEX idx_notification_logs_tenant_type_sent 
    ON notification_logs(tenant_id, notification_type, sent_at DESC);

COMMENT ON TABLE notification_logs IS '通知ログ: 通知送信履歴（真のMVP: stub実装、v1.1で本格利用）';
COMMENT ON COLUMN notification_logs.notification_type IS '通知種別（シフト募集、締切、確定通知など）';
COMMENT ON COLUMN notification_logs.delivery_channel IS '配信チャネル（Discord/Email/WebPush）';
COMMENT ON COLUMN notification_logs.delivery_status IS '配信ステータス（success/failed/pending）';

-- ============================================================
-- 2. Notification Templates テーブル（通知テンプレート）
-- ============================================================
CREATE TABLE IF NOT EXISTS notification_templates (
    template_id CHAR(26) PRIMARY KEY,  -- ULID形式
    tenant_id CHAR(26) NOT NULL,
    template_type VARCHAR(50) NOT NULL,
    template_name VARCHAR(255) NOT NULL,
    message_template TEXT NOT NULL,
    variable_definitions JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ NULL,
    
    CONSTRAINT fk_notification_templates_tenant FOREIGN KEY (tenant_id) 
        REFERENCES tenants(tenant_id) ON DELETE CASCADE,
    
    CONSTRAINT notification_templates_type_check CHECK (
        template_type IN (
            'shift_recruitment',
            'deadline_reminder',
            'shift_confirmed',
            'attendance_reminder',
            'urgent_help'
        )
    ),
    
    CONSTRAINT notification_templates_name_check CHECK (LENGTH(template_name) >= 1)
);

-- 同一テナント内で種別ごとに一意
CREATE UNIQUE INDEX idx_notification_templates_tenant_type_unique 
    ON notification_templates(tenant_id, template_type) 
    WHERE deleted_at IS NULL;

COMMENT ON TABLE notification_templates IS '通知テンプレート: 通知メッセージのテンプレート（v1.1以降で本格利用）';
COMMENT ON COLUMN notification_templates.message_template IS 'メッセージテンプレート（変数埋め込み可能）';
COMMENT ON COLUMN notification_templates.variable_definitions IS 'テンプレート変数の定義（JSONB）';

-- variable_definitions の例:
-- {
--   "member_name": {"type": "string", "description": "メンバー名"},
--   "shift_date": {"type": "date", "description": "シフト日"},
--   "shift_time": {"type": "string", "description": "シフト時刻"}
-- }

