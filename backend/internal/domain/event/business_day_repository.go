package event

import (
	"context"
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
)

// EventBusinessDayRepository defines the interface for EventBusinessDay persistence
type EventBusinessDayRepository interface {
	// Save saves an event business day (insert or update)
	Save(ctx context.Context, businessDay *EventBusinessDay) error

	// FindByID finds a business day by ID within a tenant
	FindByID(ctx context.Context, tenantID common.TenantID, businessDayID BusinessDayID) (*EventBusinessDay, error)

	// FindByEventID finds all business days for an event
	FindByEventID(ctx context.Context, tenantID common.TenantID, eventID common.EventID) ([]*EventBusinessDay, error)

	// FindByEventIDAndDateRange finds business days within a date range for an event
	FindByEventIDAndDateRange(ctx context.Context, tenantID common.TenantID, eventID common.EventID, startDate, endDate time.Time) ([]*EventBusinessDay, error)

	// FindActiveByEventID finds all active business days for an event
	FindActiveByEventID(ctx context.Context, tenantID common.TenantID, eventID common.EventID) ([]*EventBusinessDay, error)

	// FindByTenantIDAndDate finds all business days on a specific date within a tenant
	FindByTenantIDAndDate(ctx context.Context, tenantID common.TenantID, date time.Time) ([]*EventBusinessDay, error)

	// Delete deletes a business day (physical delete)
	// 通常は EventBusinessDay.Delete() で論理削除を使用するため、このメソッドは稀に使用
	Delete(ctx context.Context, tenantID common.TenantID, businessDayID BusinessDayID) error

	// ExistsByEventIDAndDate checks if a business day exists for the given event and date
	ExistsByEventIDAndDate(ctx context.Context, tenantID common.TenantID, eventID common.EventID, date time.Time, startTime time.Time) (bool, error)

	// FindRecentByTenantID finds recent N business days within a tenant (past only, oldest first)
	// Used for actual attendance calculation
	FindRecentByTenantID(ctx context.Context, tenantID common.TenantID, limit int) ([]*EventBusinessDay, error)
}

