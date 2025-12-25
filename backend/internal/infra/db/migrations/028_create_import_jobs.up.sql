-- インポートジョブ管理テーブル
CREATE TABLE import_jobs (
    import_job_id  VARCHAR(26) PRIMARY KEY,
    tenant_id      VARCHAR(26) NOT NULL REFERENCES tenants(tenant_id),
    import_type    VARCHAR(50) NOT NULL,  -- 'members' | 'actual_attendance'
    status         VARCHAR(20) NOT NULL DEFAULT 'pending',
                   -- pending | processing | completed | failed
    file_name      VARCHAR(255),
    total_rows     INTEGER DEFAULT 0,
    processed_rows INTEGER DEFAULT 0,
    success_count  INTEGER DEFAULT 0,
    error_count    INTEGER DEFAULT 0,
    error_details  JSONB DEFAULT '[]',
    options        JSONB DEFAULT '{}',
    started_at     TIMESTAMP WITH TIME ZONE,
    completed_at   TIMESTAMP WITH TIME ZONE,
    created_at     TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_by     VARCHAR(26) REFERENCES admins(admin_id)
);

CREATE INDEX idx_import_jobs_tenant ON import_jobs(tenant_id);
CREATE INDEX idx_import_jobs_status ON import_jobs(status);
CREATE INDEX idx_import_jobs_created_at ON import_jobs(created_at DESC);

COMMENT ON TABLE import_jobs IS 'CSVインポートジョブの管理テーブル';
COMMENT ON COLUMN import_jobs.import_type IS 'インポート種別: members, actual_attendance';
COMMENT ON COLUMN import_jobs.status IS 'ステータス: pending, processing, completed, failed';
COMMENT ON COLUMN import_jobs.options IS 'インポートオプション（JSON）';

-- インポート詳細ログテーブル
CREATE TABLE import_job_logs (
    log_id        VARCHAR(26) PRIMARY KEY,
    import_job_id VARCHAR(26) NOT NULL REFERENCES import_jobs(import_job_id) ON DELETE CASCADE,
    row_number    INTEGER NOT NULL,
    status        VARCHAR(20) NOT NULL,  -- success | error | skipped
    input_data    JSONB,
    error_message TEXT,
    created_at    TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_import_job_logs_job ON import_job_logs(import_job_id);
CREATE INDEX idx_import_job_logs_status ON import_job_logs(import_job_id, status);

COMMENT ON TABLE import_job_logs IS 'インポートジョブの行ごとの処理ログ';
COMMENT ON COLUMN import_job_logs.status IS '処理結果: success, error, skipped';
