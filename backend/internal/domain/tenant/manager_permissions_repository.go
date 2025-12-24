package tenant

import (
	"context"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
)

// ManagerPermissionsRepository defines the interface for manager permissions persistence
type ManagerPermissionsRepository interface {
	// FindByTenantID finds manager permissions by tenant ID
	// Returns nil if not found (not an error - use default permissions)
	FindByTenantID(ctx context.Context, tenantID common.TenantID) (*ManagerPermissions, error)

	// Save saves manager permissions (INSERT or UPDATE)
	Save(ctx context.Context, permissions *ManagerPermissions) error
}
