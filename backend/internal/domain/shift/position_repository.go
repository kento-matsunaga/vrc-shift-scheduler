package shift

import (
	"context"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
)

// PositionRepository defines the interface for Position persistence
type PositionRepository interface {
	// Save saves a position (insert or update)
	Save(ctx context.Context, position *Position) error

	// FindByID finds a position by ID within a tenant
	FindByID(ctx context.Context, tenantID common.TenantID, positionID PositionID) (*Position, error)

	// FindByTenantID finds all positions within a tenant
	FindByTenantID(ctx context.Context, tenantID common.TenantID) ([]*Position, error)

	// FindActiveByTenantID finds all active positions within a tenant
	FindActiveByTenantID(ctx context.Context, tenantID common.TenantID) ([]*Position, error)

	// Delete deletes a position (physical delete)
	Delete(ctx context.Context, tenantID common.TenantID, positionID PositionID) error
}

