package role

import (
	"context"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/role"
)

type UpdateRoleUsecase struct {
	repo role.RoleRepository
}

func NewUpdateRoleUsecase(repo role.RoleRepository) *UpdateRoleUsecase {
	return &UpdateRoleUsecase{repo: repo}
}

func (u *UpdateRoleUsecase) Execute(ctx context.Context, input UpdateRoleInput) (*UpdateRoleOutput, error) {
	tenantID, err := common.ParseTenantID(input.TenantID)
	if err != nil {
		return nil, err
	}

	roleID, err := common.ParseRoleID(input.RoleID)
	if err != nil {
		return nil, err
	}

	// Find existing role
	roleEntity, err := u.repo.FindByID(ctx, tenantID, roleID)
	if err != nil {
		return nil, err
	}

	// Update role details
	if err := roleEntity.UpdateDetails(input.Name, input.Description, input.Color, input.DisplayOrder); err != nil {
		return nil, err
	}

	// Save
	if err := u.repo.Save(ctx, roleEntity); err != nil {
		return nil, err
	}

	// Build output
	return &UpdateRoleOutput{
		RoleID:       roleEntity.RoleID().String(),
		TenantID:     roleEntity.TenantID().String(),
		Name:         roleEntity.Name(),
		Description:  roleEntity.Description(),
		Color:        roleEntity.Color(),
		DisplayOrder: roleEntity.DisplayOrder(),
		CreatedAt:    roleEntity.CreatedAt(),
		UpdatedAt:    roleEntity.UpdatedAt(),
	}, nil
}
