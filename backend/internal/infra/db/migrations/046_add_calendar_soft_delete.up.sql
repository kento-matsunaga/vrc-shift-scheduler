ALTER TABLE calendars ADD COLUMN deleted_at TIMESTAMPTZ;
ALTER TABLE calendar_entries ADD COLUMN deleted_at TIMESTAMPTZ;
