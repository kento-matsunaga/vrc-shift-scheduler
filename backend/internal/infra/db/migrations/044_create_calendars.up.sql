-- カレンダーテーブル
CREATE TABLE calendars (
    calendar_id VARCHAR(26) PRIMARY KEY,
    tenant_id VARCHAR(26) NOT NULL REFERENCES tenants(tenant_id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    is_public BOOLEAN NOT NULL DEFAULT FALSE,
    public_token VARCHAR(255) UNIQUE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_calendars_tenant_id ON calendars(tenant_id);
CREATE INDEX idx_calendars_public_token ON calendars(public_token) WHERE public_token IS NOT NULL;

-- カレンダーとイベントの中間テーブル
CREATE TABLE calendar_events (
    calendar_id VARCHAR(26) NOT NULL REFERENCES calendars(calendar_id) ON DELETE CASCADE,
    event_id VARCHAR(26) NOT NULL REFERENCES events(event_id) ON DELETE CASCADE,
    PRIMARY KEY (calendar_id, event_id)
);

CREATE INDEX idx_calendar_events_event_id ON calendar_events(event_id);
