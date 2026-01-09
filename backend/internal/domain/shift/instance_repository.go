package shift

import (
	"context"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
)

// InstanceRepository defines the interface for Instance persistence
type InstanceRepository interface {
	// Save saves an instance (insert or update)
	Save(ctx context.Context, instance *Instance) error

	// FindByID finds an instance by ID within a tenant
	FindByID(ctx context.Context, tenantID common.TenantID, instanceID InstanceID) (*Instance, error)

	// FindByEventID finds all instances for an event, ordered by display_order
	FindByEventID(ctx context.Context, tenantID common.TenantID, eventID common.EventID) ([]*Instance, error)

	// FindByEventIDAndName finds an instance by event ID and name
	// Used for matching template instance_name to existing instances
	// Returns nil, nil if not found (not an error)
	FindByEventIDAndName(ctx context.Context, tenantID common.TenantID, eventID common.EventID, name string) (*Instance, error)

	// Delete deletes an instance (physical delete)
	// Note: This may fail if there are shift slots referencing this instance
	Delete(ctx context.Context, tenantID common.TenantID, instanceID InstanceID) error
}
