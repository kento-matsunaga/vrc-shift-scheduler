package tenant

import (
	"context"
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/tenant"
)

// ManagerPermissionsRepository defines the interface for manager permissions persistence
type ManagerPermissionsRepository interface {
	FindByTenantID(ctx context.Context, tenantID common.TenantID) (*tenant.ManagerPermissions, error)
	Save(ctx context.Context, permissions *tenant.ManagerPermissions) error
}

// GetManagerPermissionsInput represents the input for getting manager permissions
type GetManagerPermissionsInput struct {
	TenantID common.TenantID
}

// GetManagerPermissionsOutput represents the output for getting manager permissions
type GetManagerPermissionsOutput struct {
	CanAddMember        bool `json:"can_add_member"`
	CanEditMember       bool `json:"can_edit_member"`
	CanDeleteMember     bool `json:"can_delete_member"`
	CanCreateEvent      bool `json:"can_create_event"`
	CanEditEvent        bool `json:"can_edit_event"`
	CanDeleteEvent      bool `json:"can_delete_event"`
	CanAssignShift      bool `json:"can_assign_shift"`
	CanEditShift        bool `json:"can_edit_shift"`
	CanCreateAttendance bool `json:"can_create_attendance"`
	CanCreateSchedule   bool `json:"can_create_schedule"`
	CanManageRoles      bool `json:"can_manage_roles"`
	CanManagePositions  bool `json:"can_manage_positions"`
	CanManageGroups     bool `json:"can_manage_groups"`
	CanInviteManager    bool `json:"can_invite_manager"`
}

// GetManagerPermissionsUsecase handles the manager permissions retrieval use case
type GetManagerPermissionsUsecase struct {
	permissionsRepo ManagerPermissionsRepository
}

// NewGetManagerPermissionsUsecase creates a new GetManagerPermissionsUsecase
func NewGetManagerPermissionsUsecase(permissionsRepo ManagerPermissionsRepository) *GetManagerPermissionsUsecase {
	return &GetManagerPermissionsUsecase{
		permissionsRepo: permissionsRepo,
	}
}

// Execute retrieves manager permissions by tenant ID
func (uc *GetManagerPermissionsUsecase) Execute(ctx context.Context, input GetManagerPermissionsInput) (*GetManagerPermissionsOutput, error) {
	permissions, err := uc.permissionsRepo.FindByTenantID(ctx, input.TenantID)
	if err != nil {
		return nil, err
	}

	// 設定が存在しない場合はデフォルト値を返す
	if permissions == nil {
		return &GetManagerPermissionsOutput{
			CanAddMember:        true,
			CanEditMember:       true,
			CanDeleteMember:     false,
			CanCreateEvent:      true,
			CanEditEvent:        true,
			CanDeleteEvent:      false,
			CanAssignShift:      true,
			CanEditShift:        true,
			CanCreateAttendance: true,
			CanCreateSchedule:   true,
			CanManageRoles:      false,
			CanManagePositions:  false,
			CanManageGroups:     false,
			CanInviteManager:    false,
		}, nil
	}

	return &GetManagerPermissionsOutput{
		CanAddMember:        permissions.CanAddMember(),
		CanEditMember:       permissions.CanEditMember(),
		CanDeleteMember:     permissions.CanDeleteMember(),
		CanCreateEvent:      permissions.CanCreateEvent(),
		CanEditEvent:        permissions.CanEditEvent(),
		CanDeleteEvent:      permissions.CanDeleteEvent(),
		CanAssignShift:      permissions.CanAssignShift(),
		CanEditShift:        permissions.CanEditShift(),
		CanCreateAttendance: permissions.CanCreateAttendance(),
		CanCreateSchedule:   permissions.CanCreateSchedule(),
		CanManageRoles:      permissions.CanManageRoles(),
		CanManagePositions:  permissions.CanManagePositions(),
		CanManageGroups:     permissions.CanManageGroups(),
		CanInviteManager:    permissions.CanInviteManager(),
	}, nil
}

// UpdateManagerPermissionsInput represents the input for updating manager permissions
type UpdateManagerPermissionsInput struct {
	TenantID            common.TenantID
	CanAddMember        bool
	CanEditMember       bool
	CanDeleteMember     bool
	CanCreateEvent      bool
	CanEditEvent        bool
	CanDeleteEvent      bool
	CanAssignShift      bool
	CanEditShift        bool
	CanCreateAttendance bool
	CanCreateSchedule   bool
	CanManageRoles      bool
	CanManagePositions  bool
	CanManageGroups     bool
	CanInviteManager    bool
}

// UpdateManagerPermissionsUsecase handles the manager permissions update use case
type UpdateManagerPermissionsUsecase struct {
	permissionsRepo ManagerPermissionsRepository
}

// NewUpdateManagerPermissionsUsecase creates a new UpdateManagerPermissionsUsecase
func NewUpdateManagerPermissionsUsecase(permissionsRepo ManagerPermissionsRepository) *UpdateManagerPermissionsUsecase {
	return &UpdateManagerPermissionsUsecase{
		permissionsRepo: permissionsRepo,
	}
}

// Execute updates manager permissions
func (uc *UpdateManagerPermissionsUsecase) Execute(ctx context.Context, input UpdateManagerPermissionsInput) (*GetManagerPermissionsOutput, error) {
	now := time.Now()

	// 既存の権限設定を取得
	permissions, err := uc.permissionsRepo.FindByTenantID(ctx, input.TenantID)
	if err != nil {
		return nil, err
	}

	// 存在しない場合は新規作成
	if permissions == nil {
		permissions, err = tenant.NewManagerPermissions(input.TenantID, now)
		if err != nil {
			return nil, err
		}
	}

	// 権限を更新
	permissions.Update(
		now,
		input.CanAddMember,
		input.CanEditMember,
		input.CanDeleteMember,
		input.CanCreateEvent,
		input.CanEditEvent,
		input.CanDeleteEvent,
		input.CanAssignShift,
		input.CanEditShift,
		input.CanCreateAttendance,
		input.CanCreateSchedule,
		input.CanManageRoles,
		input.CanManagePositions,
		input.CanManageGroups,
		input.CanInviteManager,
	)

	// 保存
	if err := uc.permissionsRepo.Save(ctx, permissions); err != nil {
		return nil, err
	}

	return &GetManagerPermissionsOutput{
		CanAddMember:        permissions.CanAddMember(),
		CanEditMember:       permissions.CanEditMember(),
		CanDeleteMember:     permissions.CanDeleteMember(),
		CanCreateEvent:      permissions.CanCreateEvent(),
		CanEditEvent:        permissions.CanEditEvent(),
		CanDeleteEvent:      permissions.CanDeleteEvent(),
		CanAssignShift:      permissions.CanAssignShift(),
		CanEditShift:        permissions.CanEditShift(),
		CanCreateAttendance: permissions.CanCreateAttendance(),
		CanCreateSchedule:   permissions.CanCreateSchedule(),
		CanManageRoles:      permissions.CanManageRoles(),
		CanManagePositions:  permissions.CanManagePositions(),
		CanManageGroups:     permissions.CanManageGroups(),
		CanInviteManager:    permissions.CanInviteManager(),
	}, nil
}

// CheckManagerPermissionInput represents the input for checking a manager's permission
type CheckManagerPermissionInput struct {
	TenantID       common.TenantID
	PermissionType tenant.PermissionType
}

// CheckManagerPermissionUsecase checks if a manager has a specific permission
type CheckManagerPermissionUsecase struct {
	permissionsRepo ManagerPermissionsRepository
}

// NewCheckManagerPermissionUsecase creates a new CheckManagerPermissionUsecase
func NewCheckManagerPermissionUsecase(permissionsRepo ManagerPermissionsRepository) *CheckManagerPermissionUsecase {
	return &CheckManagerPermissionUsecase{
		permissionsRepo: permissionsRepo,
	}
}

// Execute checks if a manager has the specified permission
func (uc *CheckManagerPermissionUsecase) Execute(ctx context.Context, input CheckManagerPermissionInput) (bool, error) {
	permissions, err := uc.permissionsRepo.FindByTenantID(ctx, input.TenantID)
	if err != nil {
		return false, err
	}

	// 設定が存在しない場合はデフォルト値で判定
	if permissions == nil {
		defaultPerms, err := tenant.NewManagerPermissions(input.TenantID, time.Now())
		if err != nil {
			return false, err
		}
		return defaultPerms.HasPermission(input.PermissionType), nil
	}

	return permissions.HasPermission(input.PermissionType), nil
}
