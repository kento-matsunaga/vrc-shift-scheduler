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

// RoleGroupRepository defines the interface for RoleGroup persistence
type RoleGroupRepository interface {
	// Save saves a group (insert or update)
	Save(ctx context.Context, group *RoleGroup) error

	// FindByID finds a group by ID within a tenant
	FindByID(ctx context.Context, tenantID common.TenantID, groupID common.RoleGroupID) (*RoleGroup, error)

	// FindByTenantID finds all groups within a tenant (deleted_at IS NULL)
	FindByTenantID(ctx context.Context, tenantID common.TenantID) ([]*RoleGroup, error)

	// Delete deletes a group (soft delete)
	Delete(ctx context.Context, tenantID common.TenantID, groupID common.RoleGroupID) error

	// AssignRole assigns a role to a group
	AssignRole(ctx context.Context, groupID common.RoleGroupID, roleID common.RoleID) error

	// RemoveRole removes a role from a group
	RemoveRole(ctx context.Context, groupID common.RoleGroupID, roleID common.RoleID) error

	// FindRoleIDsByGroupID finds all role IDs in a group
	FindRoleIDsByGroupID(ctx context.Context, groupID common.RoleGroupID) ([]common.RoleID, error)

	// FindGroupIDsByRoleID finds all group IDs for a role
	FindGroupIDsByRoleID(ctx context.Context, roleID common.RoleID) ([]common.RoleGroupID, error)

	// SetGroupRoles replaces all roles in a group
	SetGroupRoles(ctx context.Context, groupID common.RoleGroupID, roleIDs []common.RoleID) error
}
