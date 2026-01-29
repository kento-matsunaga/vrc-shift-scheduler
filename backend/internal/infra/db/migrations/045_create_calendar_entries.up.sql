CREATE TABLE calendar_entries (
    entry_id VARCHAR(26) PRIMARY KEY,
    calendar_id VARCHAR(26) NOT NULL,
    tenant_id VARCHAR(26) NOT NULL,
    title VARCHAR(255) NOT NULL,
    entry_date DATE NOT NULL,
    start_time TIME,
    end_time TIME,
    note TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_calendar_entries_calendar FOREIGN KEY (calendar_id) REFERENCES calendars(calendar_id) ON DELETE CASCADE,
    CONSTRAINT fk_calendar_entries_tenant FOREIGN KEY (tenant_id) REFERENCES tenants(tenant_id) ON DELETE CASCADE
);

CREATE INDEX idx_calendar_entries_calendar_date ON calendar_entries(calendar_id, entry_date);
CREATE INDEX idx_calendar_entries_tenant_id ON calendar_entries(tenant_id);
