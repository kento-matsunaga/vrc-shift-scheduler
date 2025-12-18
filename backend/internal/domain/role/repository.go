package role

import (
	"context"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
)

// RoleRepository defines the interface for Role persistence
type RoleRepository interface {
	// Save saves a role (insert or update)
	Save(ctx context.Context, role *Role) error

	// FindByID finds a role by ID within a tenant
	FindByID(ctx context.Context, tenantID common.TenantID, roleID common.RoleID) (*Role, error)

	// FindByTenantID finds all roles within a tenant (deleted_at IS NULL)
	FindByTenantID(ctx context.Context, tenantID common.TenantID) ([]*Role, error)

	// Delete deletes a role (soft delete)
	Delete(ctx context.Context, tenantID common.TenantID, roleID common.RoleID) error
}
