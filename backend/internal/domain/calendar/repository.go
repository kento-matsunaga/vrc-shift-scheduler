package calendar

import (
	"context"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
)

// Repository defines the interface for Calendar persistence
// Multi-Tenant前提: 全メソッドで tenant_id を引数に取る
type Repository interface {
	// Create saves a new calendar
	Create(ctx context.Context, calendar *Calendar) error

	// FindByID finds a calendar by ID within a tenant
	// tenant_id を引数に取ることで、テナント境界を越えたアクセスを防ぐ
	FindByID(ctx context.Context, tenantID common.TenantID, calendarID common.CalendarID) (*Calendar, error)

	// FindByTenantID finds all calendars within a tenant
	FindByTenantID(ctx context.Context, tenantID common.TenantID) ([]*Calendar, error)

	// FindByPublicToken finds a calendar by its public token
	// 公開カレンダーへのアクセス用（tenant_id不要）
	FindByPublicToken(ctx context.Context, token common.PublicToken) (*Calendar, error)

	// Update updates an existing calendar
	Update(ctx context.Context, calendar *Calendar) error

	// Delete deletes a calendar
	Delete(ctx context.Context, tenantID common.TenantID, calendarID common.CalendarID) error
}
