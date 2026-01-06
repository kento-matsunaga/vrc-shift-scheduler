package shift

import (
	"context"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/event"
)

// ShiftSlotRepository defines the interface for ShiftSlot persistence
type ShiftSlotRepository interface {
	// Save saves a shift slot (insert or update)
	Save(ctx context.Context, slot *ShiftSlot) error

	// FindByID finds a shift slot by ID within a tenant
	FindByID(ctx context.Context, tenantID common.TenantID, slotID SlotID) (*ShiftSlot, error)

	// FindByBusinessDayID finds all shift slots for a business day
	FindByBusinessDayID(ctx context.Context, tenantID common.TenantID, businessDayID event.BusinessDayID) ([]*ShiftSlot, error)

	// Delete deletes a shift slot (physical delete)
	// 通常は ShiftSlot.Delete() で論理削除を使用するため、このメソッドは稀に使用
	Delete(ctx context.Context, tenantID common.TenantID, slotID SlotID) error
}

