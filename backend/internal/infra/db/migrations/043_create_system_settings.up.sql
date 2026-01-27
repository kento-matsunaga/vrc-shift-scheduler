-- システム設定テーブル
-- リリース状態などのシステム全体の設定を管理
CREATE TABLE system_settings (
    key VARCHAR(100) PRIMARY KEY,
    value JSONB NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- 初期データ: リリース状態（デフォルトはリリース前）
INSERT INTO system_settings (key, value) VALUES
    ('release_status', '{"released": false}');

-- インデックス
CREATE INDEX idx_system_settings_updated_at ON system_settings(updated_at);

COMMENT ON TABLE system_settings IS 'システム全体の設定を管理するテーブル';
COMMENT ON COLUMN system_settings.key IS '設定キー';
COMMENT ON COLUMN system_settings.value IS '設定値（JSON形式）';
COMMENT ON COLUMN system_settings.updated_at IS '更新日時';
