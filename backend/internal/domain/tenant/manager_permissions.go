package tenant

import (
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
)

// ManagerPermissions represents the permission settings for managers in a tenant
type ManagerPermissions struct {
	tenantID common.TenantID

	// メンバー管理
	canAddMember    bool
	canEditMember   bool
	canDeleteMember bool

	// イベント管理
	canCreateEvent bool
	canEditEvent   bool
	canDeleteEvent bool

	// シフト管理
	canAssignShift bool
	canEditShift   bool

	// 出欠確認・日程調整
	canCreateAttendance bool
	canCreateSchedule   bool

	// 設定管理
	canManageRoles     bool
	canManagePositions bool
	canManageGroups    bool

	// 管理者招待
	canInviteManager bool

	createdAt time.Time
	updatedAt time.Time
}

// NewManagerPermissions creates a new ManagerPermissions with default values
func NewManagerPermissions(tenantID common.TenantID, now time.Time) (*ManagerPermissions, error) {
	if err := tenantID.Validate(); err != nil {
		return nil, common.NewValidationError("tenant_id is invalid", err)
	}

	return &ManagerPermissions{
		tenantID: tenantID,
		// デフォルト: 基本操作は許可、管理系は禁止
		canAddMember:        true,
		canEditMember:       true,
		canDeleteMember:     false,
		canCreateEvent:      true,
		canEditEvent:        true,
		canDeleteEvent:      false,
		canAssignShift:      true,
		canEditShift:        true,
		canCreateAttendance: true,
		canCreateSchedule:   true,
		canManageRoles:      false,
		canManagePositions:  false,
		canManageGroups:     false,
		canInviteManager:    false,
		createdAt:           now,
		updatedAt:           now,
	}, nil
}

// ReconstructManagerPermissions reconstructs a ManagerPermissions from persistence
func ReconstructManagerPermissions(
	tenantID common.TenantID,
	canAddMember bool,
	canEditMember bool,
	canDeleteMember bool,
	canCreateEvent bool,
	canEditEvent bool,
	canDeleteEvent bool,
	canAssignShift bool,
	canEditShift bool,
	canCreateAttendance bool,
	canCreateSchedule bool,
	canManageRoles bool,
	canManagePositions bool,
	canManageGroups bool,
	canInviteManager bool,
	createdAt time.Time,
	updatedAt time.Time,
) *ManagerPermissions {
	return &ManagerPermissions{
		tenantID:            tenantID,
		canAddMember:        canAddMember,
		canEditMember:       canEditMember,
		canDeleteMember:     canDeleteMember,
		canCreateEvent:      canCreateEvent,
		canEditEvent:        canEditEvent,
		canDeleteEvent:      canDeleteEvent,
		canAssignShift:      canAssignShift,
		canEditShift:        canEditShift,
		canCreateAttendance: canCreateAttendance,
		canCreateSchedule:   canCreateSchedule,
		canManageRoles:      canManageRoles,
		canManagePositions:  canManagePositions,
		canManageGroups:     canManageGroups,
		canInviteManager:    canInviteManager,
		createdAt:           createdAt,
		updatedAt:           updatedAt,
	}
}

// Getters

func (p *ManagerPermissions) TenantID() common.TenantID {
	return p.tenantID
}

func (p *ManagerPermissions) CanAddMember() bool {
	return p.canAddMember
}

func (p *ManagerPermissions) CanEditMember() bool {
	return p.canEditMember
}

func (p *ManagerPermissions) CanDeleteMember() bool {
	return p.canDeleteMember
}

func (p *ManagerPermissions) CanCreateEvent() bool {
	return p.canCreateEvent
}

func (p *ManagerPermissions) CanEditEvent() bool {
	return p.canEditEvent
}

func (p *ManagerPermissions) CanDeleteEvent() bool {
	return p.canDeleteEvent
}

func (p *ManagerPermissions) CanAssignShift() bool {
	return p.canAssignShift
}

func (p *ManagerPermissions) CanEditShift() bool {
	return p.canEditShift
}

func (p *ManagerPermissions) CanCreateAttendance() bool {
	return p.canCreateAttendance
}

func (p *ManagerPermissions) CanCreateSchedule() bool {
	return p.canCreateSchedule
}

func (p *ManagerPermissions) CanManageRoles() bool {
	return p.canManageRoles
}

func (p *ManagerPermissions) CanManagePositions() bool {
	return p.canManagePositions
}

func (p *ManagerPermissions) CanManageGroups() bool {
	return p.canManageGroups
}

func (p *ManagerPermissions) CanInviteManager() bool {
	return p.canInviteManager
}

func (p *ManagerPermissions) CreatedAt() time.Time {
	return p.createdAt
}

func (p *ManagerPermissions) UpdatedAt() time.Time {
	return p.updatedAt
}

// Update updates all permissions at once
func (p *ManagerPermissions) Update(
	now time.Time,
	canAddMember bool,
	canEditMember bool,
	canDeleteMember bool,
	canCreateEvent bool,
	canEditEvent bool,
	canDeleteEvent bool,
	canAssignShift bool,
	canEditShift bool,
	canCreateAttendance bool,
	canCreateSchedule bool,
	canManageRoles bool,
	canManagePositions bool,
	canManageGroups bool,
	canInviteManager bool,
) {
	p.canAddMember = canAddMember
	p.canEditMember = canEditMember
	p.canDeleteMember = canDeleteMember
	p.canCreateEvent = canCreateEvent
	p.canEditEvent = canEditEvent
	p.canDeleteEvent = canDeleteEvent
	p.canAssignShift = canAssignShift
	p.canEditShift = canEditShift
	p.canCreateAttendance = canCreateAttendance
	p.canCreateSchedule = canCreateSchedule
	p.canManageRoles = canManageRoles
	p.canManagePositions = canManagePositions
	p.canManageGroups = canManageGroups
	p.canInviteManager = canInviteManager
	p.updatedAt = now
}

// PermissionType represents a type of permission
type PermissionType string

const (
	PermissionAddMember        PermissionType = "add_member"
	PermissionEditMember       PermissionType = "edit_member"
	PermissionDeleteMember     PermissionType = "delete_member"
	PermissionCreateEvent      PermissionType = "create_event"
	PermissionEditEvent        PermissionType = "edit_event"
	PermissionDeleteEvent      PermissionType = "delete_event"
	PermissionAssignShift      PermissionType = "assign_shift"
	PermissionEditShift        PermissionType = "edit_shift"
	PermissionCreateAttendance PermissionType = "create_attendance"
	PermissionCreateSchedule   PermissionType = "create_schedule"
	PermissionManageRoles      PermissionType = "manage_roles"
	PermissionManagePositions  PermissionType = "manage_positions"
	PermissionManageGroups     PermissionType = "manage_groups"
	PermissionInviteManager    PermissionType = "invite_manager"
)

// HasPermission checks if the manager has a specific permission
func (p *ManagerPermissions) HasPermission(permType PermissionType) bool {
	switch permType {
	case PermissionAddMember:
		return p.canAddMember
	case PermissionEditMember:
		return p.canEditMember
	case PermissionDeleteMember:
		return p.canDeleteMember
	case PermissionCreateEvent:
		return p.canCreateEvent
	case PermissionEditEvent:
		return p.canEditEvent
	case PermissionDeleteEvent:
		return p.canDeleteEvent
	case PermissionAssignShift:
		return p.canAssignShift
	case PermissionEditShift:
		return p.canEditShift
	case PermissionCreateAttendance:
		return p.canCreateAttendance
	case PermissionCreateSchedule:
		return p.canCreateSchedule
	case PermissionManageRoles:
		return p.canManageRoles
	case PermissionManagePositions:
		return p.canManagePositions
	case PermissionManageGroups:
		return p.canManageGroups
	case PermissionInviteManager:
		return p.canInviteManager
	default:
		return false
	}
}
