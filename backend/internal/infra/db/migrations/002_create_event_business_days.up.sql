-- Migration: 002_create_event_business_days
-- Description: イベント営業日テーブルの作成
-- RecurringPatternから生成されたインスタンスを格納

-- ============================================================
-- Event Business Days テーブル
-- ============================================================
CREATE TABLE IF NOT EXISTS event_business_days (
    business_day_id CHAR(26) PRIMARY KEY,  -- ULID形式
    tenant_id CHAR(26) NOT NULL,
    event_id CHAR(26) NOT NULL,
    target_date DATE NOT NULL,
    start_time TIME NOT NULL,
    end_time TIME NOT NULL,
    occurrence_type VARCHAR(20) NOT NULL,
    recurring_pattern_id CHAR(26) NULL,  -- 通常営業の場合のみ
    is_active BOOLEAN NOT NULL DEFAULT true,
    valid_from DATE NULL,
    valid_to DATE NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ NULL,
    
    CONSTRAINT fk_event_business_days_tenant FOREIGN KEY (tenant_id) 
        REFERENCES tenants(tenant_id) ON DELETE CASCADE,
    CONSTRAINT fk_event_business_days_event FOREIGN KEY (event_id) 
        REFERENCES events(event_id) ON DELETE CASCADE,
    CONSTRAINT fk_event_business_days_pattern FOREIGN KEY (recurring_pattern_id) 
        REFERENCES recurring_patterns(pattern_id) ON DELETE SET NULL,
    
    -- 深夜営業対応: 終了時刻が開始時刻より前の場合、日付をまたぐ営業として扱う
    CONSTRAINT event_business_days_time_check CHECK (
        start_time < end_time OR end_time < start_time
    ),
    
    CONSTRAINT event_business_days_occurrence_type_check CHECK (
        occurrence_type IN ('recurring', 'special')
    ),
    
    -- recurring の場合は recurring_pattern_id が必須、special の場合は NULL
    CONSTRAINT event_business_days_pattern_consistency_check CHECK (
        (occurrence_type = 'recurring' AND recurring_pattern_id IS NOT NULL) OR
        (occurrence_type = 'special' AND recurring_pattern_id IS NULL)
    ),
    
    -- valid_from と valid_to の整合性チェック
    CONSTRAINT event_business_days_valid_period_check CHECK (
        (valid_from IS NULL AND valid_to IS NULL) OR
        (valid_from IS NOT NULL AND valid_to IS NOT NULL AND valid_from <= valid_to)
    )
);

-- 同一テナント・イベント・日時で一意（論理削除されていないもののみ）
CREATE UNIQUE INDEX idx_event_business_days_tenant_event_date_time_unique 
    ON event_business_days(tenant_id, event_id, target_date, start_time) 
    WHERE deleted_at IS NULL;

-- テナント内の日付検索用（カレンダー表示などで使用）
CREATE INDEX idx_event_business_days_tenant_date 
    ON event_business_days(tenant_id, target_date) 
    WHERE deleted_at IS NULL;

-- イベント内の営業日検索用（イベント詳細画面で使用）
CREATE INDEX idx_event_business_days_event_date 
    ON event_business_days(event_id, target_date) 
    WHERE deleted_at IS NULL;

-- アクティブな営業日のみを検索する際のインデックス
CREATE INDEX idx_event_business_days_event_active 
    ON event_business_days(event_id, is_active, target_date) 
    WHERE deleted_at IS NULL;

COMMENT ON TABLE event_business_days IS 'イベント営業日: RecurringPatternから生成されたインスタンス';
COMMENT ON COLUMN event_business_days.target_date IS '営業日の日付（JSTのローカル日付）';
COMMENT ON COLUMN event_business_days.start_time IS '営業開始時刻（TIME WITHOUT TIME ZONE）';
COMMENT ON COLUMN event_business_days.end_time IS '営業終了時刻（深夜営業の場合、start_timeより前の時刻も許可）';
COMMENT ON COLUMN event_business_days.occurrence_type IS '発生種別: recurring（定期）、special（特別）';
COMMENT ON COLUMN event_business_days.recurring_pattern_id IS '生成元の定期パターン（通常営業の場合のみ）';
COMMENT ON COLUMN event_business_days.valid_from IS '有効期間開始日（例外的な無効化に使用）';
COMMENT ON COLUMN event_business_days.valid_to IS '有効期間終了日（例外的な無効化に使用）';

