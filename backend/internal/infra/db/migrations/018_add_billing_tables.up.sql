-- Migration: 018_add_billing_tables
-- Description: マネタイズ基盤用テーブルの作成
-- BOOTH買い切り・Stripeサブスク対応、テナント状態管理

-- ============================================================
-- 1. tenants テーブルへのカラム追加
-- ============================================================
ALTER TABLE tenants ADD COLUMN IF NOT EXISTS status VARCHAR(20) NOT NULL DEFAULT 'active';
ALTER TABLE tenants ADD COLUMN IF NOT EXISTS grace_until TIMESTAMPTZ NULL;

-- status カラムの CHECK 制約を追加
ALTER TABLE tenants ADD CONSTRAINT tenants_status_check
    CHECK (status IN ('active', 'grace', 'suspended'));

CREATE INDEX IF NOT EXISTS idx_tenants_status ON tenants(status) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_tenants_grace_until ON tenants(grace_until) WHERE grace_until IS NOT NULL;

COMMENT ON COLUMN tenants.status IS 'テナント状態: active（通常）、grace（猶予期間）、suspended（停止）';
COMMENT ON COLUMN tenants.grace_until IS 'grace期間の終了日時（NULL=grace期間外）';

-- ============================================================
-- 2. plans テーブル（プラン定義マスタ）
-- ============================================================
CREATE TABLE IF NOT EXISTS plans (
    plan_code VARCHAR(50) PRIMARY KEY,
    plan_type VARCHAR(20) NOT NULL,
    display_name VARCHAR(100) NOT NULL,
    price_jpy INTEGER NULL,
    features_json JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT plans_type_check CHECK (plan_type IN ('lifetime', 'subscription'))
);

COMMENT ON TABLE plans IS 'プラン定義: 買い切り(lifetime)またはサブスク(subscription)';
COMMENT ON COLUMN plans.plan_code IS 'プランコード: LIFETIME, SUB_980 など';
COMMENT ON COLUMN plans.plan_type IS 'プラン種別: lifetime（買い切り）、subscription（月額）';
COMMENT ON COLUMN plans.price_jpy IS '価格（円）、買い切りはNULL';
COMMENT ON COLUMN plans.features_json IS '機能フラグ（将来拡張用）';

-- 初期データ挿入
INSERT INTO plans (plan_code, plan_type, display_name, price_jpy, features_json) VALUES
    ('LIFETIME', 'lifetime', '買い切りプラン', NULL, '{}'),
    ('SUB_980', 'subscription', '月額プラン', 980, '{}')
ON CONFLICT (plan_code) DO NOTHING;

-- ============================================================
-- 3. entitlements テーブル（権利付与）
-- ============================================================
CREATE TABLE IF NOT EXISTS entitlements (
    entitlement_id CHAR(26) PRIMARY KEY,
    tenant_id CHAR(26) NOT NULL,
    plan_code VARCHAR(50) NOT NULL,
    source VARCHAR(50) NOT NULL,
    starts_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    ends_at TIMESTAMPTZ NULL,
    revoked_at TIMESTAMPTZ NULL,
    revoked_reason TEXT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_entitlements_tenant FOREIGN KEY (tenant_id)
        REFERENCES tenants(tenant_id) ON DELETE CASCADE,
    CONSTRAINT fk_entitlements_plan FOREIGN KEY (plan_code)
        REFERENCES plans(plan_code) ON DELETE RESTRICT,
    CONSTRAINT entitlements_source_check CHECK (source IN ('booth', 'stripe', 'manual'))
);

CREATE INDEX idx_entitlements_tenant_id ON entitlements(tenant_id);
CREATE INDEX idx_entitlements_plan_code ON entitlements(plan_code);
CREATE INDEX idx_entitlements_revoked ON entitlements(revoked_at) WHERE revoked_at IS NOT NULL;

COMMENT ON TABLE entitlements IS '権利付与: テナントに付与されたプラン権利';
COMMENT ON COLUMN entitlements.source IS '権利付与元: booth（BOOTH購入）、stripe（Stripe）、manual（手動付与）';
COMMENT ON COLUMN entitlements.ends_at IS '権利終了日時（NULL=永久）';
COMMENT ON COLUMN entitlements.revoked_at IS '取り消し日時（NULL=有効）';

-- ============================================================
-- 4. subscriptions テーブル（Stripeサブスク管理）
-- ============================================================
CREATE TABLE IF NOT EXISTS subscriptions (
    subscription_id CHAR(26) PRIMARY KEY,
    tenant_id CHAR(26) NOT NULL UNIQUE,
    stripe_customer_id VARCHAR(100) NOT NULL,
    stripe_subscription_id VARCHAR(100) NOT NULL UNIQUE,
    status VARCHAR(50) NOT NULL,
    current_period_end TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_subscriptions_tenant FOREIGN KEY (tenant_id)
        REFERENCES tenants(tenant_id) ON DELETE CASCADE,
    CONSTRAINT subscriptions_status_check CHECK (
        status IN ('active', 'past_due', 'canceled', 'unpaid', 'incomplete', 'trialing')
    )
);

CREATE INDEX idx_subscriptions_stripe_subscription_id ON subscriptions(stripe_subscription_id);
CREATE INDEX idx_subscriptions_status ON subscriptions(status);

COMMENT ON TABLE subscriptions IS 'Stripeサブスク: テナントとStripeサブスクリプションの紐付け';
COMMENT ON COLUMN subscriptions.status IS 'Stripeサブスクステータス';

-- ============================================================
-- 5. webhook_events テーブル（Webhook冪等性管理）
-- ============================================================
CREATE TABLE IF NOT EXISTS webhook_events (
    id SERIAL PRIMARY KEY,
    provider VARCHAR(50) NOT NULL,
    event_id VARCHAR(100) NOT NULL,
    received_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    payload_json JSONB NULL,

    CONSTRAINT webhook_events_provider_event_unique UNIQUE (provider, event_id)
);

CREATE INDEX idx_webhook_events_received_at ON webhook_events(received_at);

COMMENT ON TABLE webhook_events IS 'Webhookイベント: 外部サービスからのWebhook冪等性管理';
COMMENT ON COLUMN webhook_events.provider IS 'プロバイダー: stripe など';
COMMENT ON COLUMN webhook_events.event_id IS 'イベントID（プロバイダー提供）';

-- ============================================================
-- 6. license_keys テーブル（BOOTHライセンスキー）
-- ============================================================
CREATE TABLE IF NOT EXISTS license_keys (
    key_id CHAR(26) PRIMARY KEY,
    key_hash VARCHAR(64) NOT NULL UNIQUE,
    status VARCHAR(20) NOT NULL DEFAULT 'unused',
    issued_batch_id CHAR(26) NULL,
    used_at TIMESTAMPTZ NULL,
    used_tenant_id CHAR(26) NULL,
    revoked_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_license_keys_tenant FOREIGN KEY (used_tenant_id)
        REFERENCES tenants(tenant_id) ON DELETE SET NULL,
    CONSTRAINT license_keys_status_check CHECK (status IN ('unused', 'used', 'revoked'))
);

CREATE INDEX idx_license_keys_status ON license_keys(status);
CREATE INDEX idx_license_keys_batch ON license_keys(issued_batch_id);
CREATE INDEX idx_license_keys_hash ON license_keys(key_hash);

COMMENT ON TABLE license_keys IS 'ライセンスキー: BOOTH購入キー管理（平文は保存しない）';
COMMENT ON COLUMN license_keys.key_hash IS 'キーのSHA-256ハッシュ';
COMMENT ON COLUMN license_keys.status IS 'ステータス: unused（未使用）、used（使用済）、revoked（失効）';
COMMENT ON COLUMN license_keys.issued_batch_id IS '発行バッチID（一括発行時のグループ識別）';

-- ============================================================
-- 7. billing_audit_logs テーブル（課金監査ログ）
-- ============================================================
CREATE TABLE IF NOT EXISTS billing_audit_logs (
    log_id CHAR(26) PRIMARY KEY,
    actor_type VARCHAR(20) NOT NULL,
    actor_id CHAR(26) NULL,
    action VARCHAR(100) NOT NULL,
    target_type VARCHAR(50) NULL,
    target_id CHAR(26) NULL,
    before_json JSONB NULL,
    after_json JSONB NULL,
    ip_address VARCHAR(45) NULL,
    user_agent TEXT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT billing_audit_logs_actor_type_check CHECK (
        actor_type IN ('admin', 'system', 'stripe', 'user')
    )
);

CREATE INDEX idx_billing_audit_logs_created_at ON billing_audit_logs(created_at DESC);
CREATE INDEX idx_billing_audit_logs_actor ON billing_audit_logs(actor_type, actor_id);
CREATE INDEX idx_billing_audit_logs_target ON billing_audit_logs(target_type, target_id);
CREATE INDEX idx_billing_audit_logs_action ON billing_audit_logs(action);

COMMENT ON TABLE billing_audit_logs IS '課金監査ログ: マネタイズ関連操作の記録';
COMMENT ON COLUMN billing_audit_logs.actor_type IS '操作者種別: admin（管理者）、system（システム）、stripe（Stripe）、user（ユーザー）';
COMMENT ON COLUMN billing_audit_logs.action IS '操作内容: license_claim, subscription_created, status_changed など';
