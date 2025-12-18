package shift

import (
	"context"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
)

// ShiftSlotTemplateRepository defines the interface for ShiftSlotTemplate persistence
type ShiftSlotTemplateRepository interface {
	// Save saves a shift slot template (insert or update)
	Save(ctx context.Context, template *ShiftSlotTemplate) error

	// FindByID finds a shift slot template by ID within a tenant
	FindByID(ctx context.Context, tenantID common.TenantID, templateID common.ShiftSlotTemplateID) (*ShiftSlotTemplate, error)

	// FindByEventID finds all shift slot templates for an event (excluding soft-deleted)
	FindByEventID(ctx context.Context, tenantID common.TenantID, eventID common.EventID) ([]*ShiftSlotTemplate, error)

	// Delete deletes a shift slot template (physical delete)
	// Usually uses ShiftSlotTemplate.Delete() for soft delete instead
	Delete(ctx context.Context, tenantID common.TenantID, templateID common.ShiftSlotTemplateID) error
}
