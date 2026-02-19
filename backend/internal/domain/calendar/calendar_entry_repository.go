package calendar

import (
	"context"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
)

// CalendarEntryRepository defines the interface for calendar entry persistence
// Multi-Tenant前提: 全メソッドで tenant_id を引数に取る
type CalendarEntryRepository interface {
	// Save saves a calendar entry (insert or update)
	Save(ctx context.Context, entry *CalendarEntry) error

	// FindByID finds a calendar entry by ID within a tenant
	FindByID(ctx context.Context, tenantID common.TenantID, entryID common.CalendarEntryID) (*CalendarEntry, error)

	// FindByCalendarID finds all entries for a calendar (ordered by date)
	FindByCalendarID(ctx context.Context, tenantID common.TenantID, calendarID common.CalendarID) ([]*CalendarEntry, error)

	// Delete deletes a calendar entry
	Delete(ctx context.Context, tenantID common.TenantID, entryID common.CalendarEntryID) error
}
