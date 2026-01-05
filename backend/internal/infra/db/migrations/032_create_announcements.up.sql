-- お知らせテーブル
CREATE TABLE IF NOT EXISTS announcements (
    id VARCHAR(26) PRIMARY KEY,
    tenant_id VARCHAR(26), -- NULL = 全テナント向け
    title VARCHAR(200) NOT NULL,
    body TEXT NOT NULL,
    published_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    FOREIGN KEY (tenant_id) REFERENCES tenants(id)
);

CREATE INDEX IF NOT EXISTS idx_announcements_tenant ON announcements(tenant_id);
CREATE INDEX IF NOT EXISTS idx_announcements_published ON announcements(published_at);
CREATE INDEX IF NOT EXISTS idx_announcements_deleted ON announcements(deleted_at);

COMMENT ON TABLE announcements IS 'お知らせ';
COMMENT ON COLUMN announcements.tenant_id IS 'テナントID（NULLは全テナント向け）';
COMMENT ON COLUMN announcements.title IS 'タイトル';
COMMENT ON COLUMN announcements.body IS '本文';
COMMENT ON COLUMN announcements.published_at IS '公開日時';

-- お知らせ既読テーブル
CREATE TABLE IF NOT EXISTS announcement_reads (
    id VARCHAR(26) PRIMARY KEY,
    announcement_id VARCHAR(26) NOT NULL,
    admin_id VARCHAR(26) NOT NULL,
    read_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (announcement_id) REFERENCES announcements(id),
    FOREIGN KEY (admin_id) REFERENCES admins(id),
    UNIQUE (announcement_id, admin_id)
);

CREATE INDEX IF NOT EXISTS idx_announcement_reads_admin ON announcement_reads(admin_id);

COMMENT ON TABLE announcement_reads IS 'お知らせ既読状態';
COMMENT ON COLUMN announcement_reads.announcement_id IS 'お知らせID';
COMMENT ON COLUMN announcement_reads.admin_id IS '既読した管理者ID';
COMMENT ON COLUMN announcement_reads.read_at IS '既読日時';
