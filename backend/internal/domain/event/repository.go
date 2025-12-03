package event

import (
	"context"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
)

// EventRepository defines the interface for Event persistence
// Multi-Tenant前提: 全メソッドで tenant_id を引数に取る
type EventRepository interface {
	// Save saves an event (insert or update)
	Save(ctx context.Context, event *Event) error

	// FindByID finds an event by ID within a tenant
	// tenant_id を引数に取ることで、テナント境界を越えたアクセスを防ぐ
	FindByID(ctx context.Context, tenantID common.TenantID, eventID common.EventID) (*Event, error)

	// FindByTenantID finds all events within a tenant
	// deleted_at IS NULL のレコードのみ返す
	FindByTenantID(ctx context.Context, tenantID common.TenantID) ([]*Event, error)

	// FindActiveByTenantID finds all active events within a tenant
	FindActiveByTenantID(ctx context.Context, tenantID common.TenantID) ([]*Event, error)

	// Delete deletes an event (physical delete)
	// 通常は Event.Delete() で論理削除を使用するため、このメソッドは稀に使用
	Delete(ctx context.Context, tenantID common.TenantID, eventID common.EventID) error

	// ExistsByName checks if an event with the given name exists within a tenant
	// 同一テナント内でのイベント名重複チェックに使用
	ExistsByName(ctx context.Context, tenantID common.TenantID, eventName string) (bool, error)
}

