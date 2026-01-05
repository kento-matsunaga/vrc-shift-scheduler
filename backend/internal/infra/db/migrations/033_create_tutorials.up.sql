-- チュートリアルテーブル
CREATE TABLE IF NOT EXISTS tutorials (
    id VARCHAR(26) PRIMARY KEY,
    category VARCHAR(50) NOT NULL,
    title VARCHAR(200) NOT NULL,
    body TEXT NOT NULL,
    display_order INT NOT NULL DEFAULT 0,
    is_published BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_tutorials_category ON tutorials(category);
CREATE INDEX IF NOT EXISTS idx_tutorials_published ON tutorials(is_published);
CREATE INDEX IF NOT EXISTS idx_tutorials_order ON tutorials(display_order);
CREATE INDEX IF NOT EXISTS idx_tutorials_deleted ON tutorials(deleted_at);

COMMENT ON TABLE tutorials IS 'チュートリアル';
COMMENT ON COLUMN tutorials.category IS 'カテゴリ';
COMMENT ON COLUMN tutorials.title IS 'タイトル';
COMMENT ON COLUMN tutorials.body IS '本文';
COMMENT ON COLUMN tutorials.display_order IS '表示順';
COMMENT ON COLUMN tutorials.is_published IS '公開状態';
