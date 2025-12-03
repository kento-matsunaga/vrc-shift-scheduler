-- Migration: 001_create_tenants_events_and_patterns
-- Description: テナント、イベント、定期パターンテーブルの作成
-- Multi-tenant設計の基盤となるテーブル群

-- ============================================================
-- 1. Tenants テーブル
-- ============================================================
CREATE TABLE IF NOT EXISTS tenants (
    tenant_id CHAR(26) PRIMARY KEY,  -- ULID形式
    tenant_name VARCHAR(255) NOT NULL,
    timezone VARCHAR(50) NOT NULL DEFAULT 'Asia/Tokyo',
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ NULL,
    
    CONSTRAINT tenants_tenant_name_check CHECK (LENGTH(tenant_name) >= 1)
);

CREATE INDEX idx_tenants_is_active ON tenants(is_active) WHERE deleted_at IS NULL;
CREATE INDEX idx_tenants_deleted_at ON tenants(deleted_at) WHERE deleted_at IS NOT NULL;

COMMENT ON TABLE tenants IS 'テナント: 組織単位での完全なデータ分離を実現';
COMMENT ON COLUMN tenants.tenant_id IS 'テナントID (ULID)';
COMMENT ON COLUMN tenants.timezone IS 'テナントのタイムゾーン（デフォルト: Asia/Tokyo）';

-- ============================================================
-- 2. Events テーブル
-- ============================================================
CREATE TABLE IF NOT EXISTS events (
    event_id CHAR(26) PRIMARY KEY,  -- ULID形式
    tenant_id CHAR(26) NOT NULL,
    event_name VARCHAR(255) NOT NULL,
    event_type VARCHAR(20) NOT NULL DEFAULT 'normal',
    description TEXT,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ NULL,
    
    CONSTRAINT fk_events_tenant FOREIGN KEY (tenant_id) 
        REFERENCES tenants(tenant_id) ON DELETE CASCADE,
    CONSTRAINT events_event_type_check CHECK (event_type IN ('normal', 'special')),
    CONSTRAINT events_event_name_check CHECK (LENGTH(event_name) >= 1)
);

-- テナント内でイベント名を一意にする（論理削除されていないもののみ）
CREATE UNIQUE INDEX idx_events_tenant_event_name_unique 
    ON events(tenant_id, event_name) 
    WHERE deleted_at IS NULL;

-- テナント内のアクティブなイベント検索用
CREATE INDEX idx_events_tenant_is_active 
    ON events(tenant_id, is_active) 
    WHERE deleted_at IS NULL;

COMMENT ON TABLE events IS 'イベント: VRChatイベントの定義（集約ルート）';
COMMENT ON COLUMN events.event_type IS 'イベント種別: normal（通常営業）、special（特別イベント）';

-- ============================================================
-- 3. Recurring Patterns テーブル
-- ============================================================
CREATE TABLE IF NOT EXISTS recurring_patterns (
    pattern_id CHAR(26) PRIMARY KEY,  -- ULID形式
    tenant_id CHAR(26) NOT NULL,
    event_id CHAR(26) NOT NULL,
    pattern_type VARCHAR(20) NOT NULL,
    config JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ NULL,
    
    CONSTRAINT fk_recurring_patterns_tenant FOREIGN KEY (tenant_id) 
        REFERENCES tenants(tenant_id) ON DELETE CASCADE,
    CONSTRAINT fk_recurring_patterns_event FOREIGN KEY (event_id) 
        REFERENCES events(event_id) ON DELETE CASCADE,
    CONSTRAINT recurring_patterns_type_check CHECK (pattern_type IN ('weekly', 'monthly_date', 'custom'))
);

-- 1 Event につき 1 RecurringPattern（論理削除されていないもののみ）
CREATE UNIQUE INDEX idx_recurring_patterns_tenant_event_unique 
    ON recurring_patterns(tenant_id, event_id) 
    WHERE deleted_at IS NULL;

-- イベントからパターンを引く際のインデックス
CREATE INDEX idx_recurring_patterns_event 
    ON recurring_patterns(event_id) 
    WHERE deleted_at IS NULL;

COMMENT ON TABLE recurring_patterns IS '定期パターン: イベントの営業日生成ルール';
COMMENT ON COLUMN recurring_patterns.pattern_type IS 'パターン種別: weekly（曜日指定）、monthly_date（月内日付）、custom（カスタム）';
COMMENT ON COLUMN recurring_patterns.config IS 'パターン設定（JSONB）: day_of_weeks, dates, start_time, end_time など';

-- config JSONBの例:
-- Weekly: {"day_of_weeks": ["MON", "FRI"], "start_time": "21:30", "end_time": "23:00"}
-- MonthlyDate: {"dates": [1, 15], "start_time": "21:30", "end_time": "23:00"}

