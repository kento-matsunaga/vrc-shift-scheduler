-- Migration: 003_create_members_and_shift_slots
-- Description: メンバー、役職、シフト枠テーブルの作成

-- ============================================================
-- 1. Members テーブル（真のMVP版: 最小限のフィールド）
-- ============================================================
CREATE TABLE IF NOT EXISTS members (
    member_id CHAR(26) PRIMARY KEY,  -- ULID形式
    tenant_id CHAR(26) NOT NULL,
    display_name VARCHAR(255) NOT NULL,
    discord_user_id VARCHAR(100) NULL,
    email VARCHAR(255) NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ NULL,
    
    CONSTRAINT fk_members_tenant FOREIGN KEY (tenant_id) 
        REFERENCES tenants(tenant_id) ON DELETE CASCADE,
    CONSTRAINT members_display_name_check CHECK (LENGTH(display_name) >= 1)
);

-- テナント内でDiscord User IDを一意にする（NULLは除く）
CREATE UNIQUE INDEX idx_members_tenant_discord_unique 
    ON members(tenant_id, discord_user_id) 
    WHERE discord_user_id IS NOT NULL AND deleted_at IS NULL;

-- テナント内でメールアドレスを一意にする（NULLは除く）
CREATE UNIQUE INDEX idx_members_tenant_email_unique 
    ON members(tenant_id, email) 
    WHERE email IS NOT NULL AND deleted_at IS NULL;

-- テナント内のアクティブなメンバー検索用
CREATE INDEX idx_members_tenant_is_active 
    ON members(tenant_id, is_active) 
    WHERE deleted_at IS NULL;

COMMENT ON TABLE members IS 'メンバー: シフトに参加するメンバー情報（MVP: 最小限の実装）';
COMMENT ON COLUMN members.discord_user_id IS 'Discord ユーザーID（オプショナル）';
COMMENT ON COLUMN members.email IS 'メールアドレス（オプショナル）';

-- ============================================================
-- 2. Positions テーブル（シフト枠の役職）
-- ============================================================
CREATE TABLE IF NOT EXISTS positions (
    position_id CHAR(26) PRIMARY KEY,  -- ULID形式
    tenant_id CHAR(26) NOT NULL,
    position_name VARCHAR(255) NOT NULL,
    description TEXT,
    display_order INT NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ NULL,
    
    CONSTRAINT fk_positions_tenant FOREIGN KEY (tenant_id) 
        REFERENCES tenants(tenant_id) ON DELETE CASCADE,
    CONSTRAINT positions_name_check CHECK (LENGTH(position_name) >= 1)
);

-- テナント内で役職名を一意にする
CREATE UNIQUE INDEX idx_positions_tenant_name_unique 
    ON positions(tenant_id, position_name) 
    WHERE deleted_at IS NULL;

-- 表示順ソート用
CREATE INDEX idx_positions_tenant_order 
    ON positions(tenant_id, display_order, position_name) 
    WHERE deleted_at IS NULL;

COMMENT ON TABLE positions IS '役職: シフト枠に必要な役職（例: スタッフ、警備、受付）';
COMMENT ON COLUMN positions.display_order IS '表示順（小さいほど上位）';

-- ============================================================
-- 3. Shift Slots テーブル
-- ============================================================
CREATE TABLE IF NOT EXISTS shift_slots (
    slot_id CHAR(26) PRIMARY KEY,  -- ULID形式
    tenant_id CHAR(26) NOT NULL,
    business_day_id CHAR(26) NOT NULL,
    position_id CHAR(26) NOT NULL,
    slot_name VARCHAR(255) NOT NULL,
    instance_name VARCHAR(255) NULL,
    start_time TIME NOT NULL,
    end_time TIME NOT NULL,
    required_count INT NOT NULL DEFAULT 1,
    priority INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ NULL,
    
    CONSTRAINT fk_shift_slots_tenant FOREIGN KEY (tenant_id) 
        REFERENCES tenants(tenant_id) ON DELETE CASCADE,
    CONSTRAINT fk_shift_slots_business_day FOREIGN KEY (business_day_id) 
        REFERENCES event_business_days(business_day_id) ON DELETE CASCADE,
    CONSTRAINT fk_shift_slots_position FOREIGN KEY (position_id) 
        REFERENCES positions(position_id) ON DELETE RESTRICT,
    
    -- 深夜営業対応: 終了時刻が開始時刻より前の場合、日付をまたぐシフト
    CONSTRAINT shift_slots_time_check CHECK (
        start_time < end_time OR end_time < start_time
    ),
    
    -- 必要人数は1以上
    CONSTRAINT shift_slots_required_count_check CHECK (required_count >= 1),
    
    CONSTRAINT shift_slots_name_check CHECK (LENGTH(slot_name) >= 1)
);

-- テナント内の営業日単位でシフト枠を検索
CREATE INDEX idx_shift_slots_tenant_business_day 
    ON shift_slots(tenant_id, business_day_id) 
    WHERE deleted_at IS NULL;

-- 営業日内での時刻順ソート用
CREATE INDEX idx_shift_slots_business_day_time 
    ON shift_slots(business_day_id, start_time, priority) 
    WHERE deleted_at IS NULL;

-- 役職ごとのシフト枠検索用
CREATE INDEX idx_shift_slots_position 
    ON shift_slots(position_id) 
    WHERE deleted_at IS NULL;

COMMENT ON TABLE shift_slots IS 'シフト枠: 営業日内の各シフト枠（独立したエンティティ）';
COMMENT ON COLUMN shift_slots.slot_name IS 'シフト枠名（例: 早番スタッフ、夜間警備）';
COMMENT ON COLUMN shift_slots.instance_name IS 'インスタンス名（同じslot_nameで複数枠がある場合の識別用）';
COMMENT ON COLUMN shift_slots.required_count IS '必要人数';
COMMENT ON COLUMN shift_slots.priority IS '優先度（小さいほど優先、1が最上位）';

