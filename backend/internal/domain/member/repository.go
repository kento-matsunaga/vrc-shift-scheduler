package member

import (
	"context"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
)

// MemberRepository defines the interface for Member persistence
// Multi-Tenant前提: 全メソッドで tenant_id を引数に取る
type MemberRepository interface {
	// Save saves a member (insert or update)
	Save(ctx context.Context, member *Member) error

	// FindByID finds a member by ID within a tenant
	FindByID(ctx context.Context, tenantID common.TenantID, memberID common.MemberID) (*Member, error)

	// FindByTenantID finds all members within a tenant
	// deleted_at IS NULL のレコードのみ返す
	FindByTenantID(ctx context.Context, tenantID common.TenantID) ([]*Member, error)

	// FindActiveByTenantID finds all active members within a tenant
	FindActiveByTenantID(ctx context.Context, tenantID common.TenantID) ([]*Member, error)

	// FindByDiscordUserID finds a member by Discord user ID within a tenant
	FindByDiscordUserID(ctx context.Context, tenantID common.TenantID, discordUserID string) (*Member, error)

	// FindByEmail finds a member by email within a tenant
	FindByEmail(ctx context.Context, tenantID common.TenantID, email string) (*Member, error)

	// Delete deletes a member (physical delete)
	// 通常は Member.Delete() で論理削除を使用するため、このメソッドは稀に使用
	Delete(ctx context.Context, tenantID common.TenantID, memberID common.MemberID) error

	// ExistsByDiscordUserID checks if a member with the given Discord user ID exists within a tenant
	ExistsByDiscordUserID(ctx context.Context, tenantID common.TenantID, discordUserID string) (bool, error)

	// ExistsByEmail checks if a member with the given email exists within a tenant
	ExistsByEmail(ctx context.Context, tenantID common.TenantID, email string) (bool, error)
}

// MemberRoleRepository defines the interface for member-role association persistence
type MemberRoleRepository interface {
	// AssignRole assigns a role to a member
	AssignRole(ctx context.Context, memberID common.MemberID, roleID common.RoleID) error

	// RemoveRole removes a role from a member
	RemoveRole(ctx context.Context, memberID common.MemberID, roleID common.RoleID) error

	// FindRolesByMemberID finds all roles assigned to a member
	FindRolesByMemberID(ctx context.Context, memberID common.MemberID) ([]common.RoleID, error)

	// FindMemberIDsByRoleID finds all members with a specific role
	FindMemberIDsByRoleID(ctx context.Context, roleID common.RoleID) ([]common.MemberID, error)

	// SetMemberRoles sets all roles for a member (replaces existing roles)
	SetMemberRoles(ctx context.Context, memberID common.MemberID, roleIDs []common.RoleID) error
}

// MemberGroupRepository defines the interface for MemberGroup persistence
type MemberGroupRepository interface {
	// Save saves a member group (insert or update)
	Save(ctx context.Context, group *MemberGroup) error

	// FindByID finds a member group by ID within a tenant
	FindByID(ctx context.Context, tenantID common.TenantID, groupID common.MemberGroupID) (*MemberGroup, error)

	// FindByTenantID finds all member groups within a tenant
	FindByTenantID(ctx context.Context, tenantID common.TenantID) ([]*MemberGroup, error)

	// Delete deletes a member group (physical delete)
	Delete(ctx context.Context, tenantID common.TenantID, groupID common.MemberGroupID) error

	// AssignMember assigns a member to a group
	AssignMember(ctx context.Context, groupID common.MemberGroupID, memberID common.MemberID) error

	// RemoveMember removes a member from a group
	RemoveMember(ctx context.Context, groupID common.MemberGroupID, memberID common.MemberID) error

	// FindMemberIDsByGroupID finds all members in a group
	FindMemberIDsByGroupID(ctx context.Context, groupID common.MemberGroupID) ([]common.MemberID, error)

	// FindGroupIDsByMemberID finds all groups a member belongs to
	FindGroupIDsByMemberID(ctx context.Context, memberID common.MemberID) ([]common.MemberGroupID, error)

	// SetMemberGroups sets all groups for a member (replaces existing groups)
	SetMemberGroups(ctx context.Context, memberID common.MemberID, groupIDs []common.MemberGroupID) error
}

