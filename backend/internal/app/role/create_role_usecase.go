package role

import (
	"context"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/role"
)

type CreateRoleUsecase struct {
	repo role.RoleRepository
}

func NewCreateRoleUsecase(repo role.RoleRepository) *CreateRoleUsecase {
	return &CreateRoleUsecase{repo: repo}
}

func (u *CreateRoleUsecase) Execute(ctx context.Context, input CreateRoleInput) (*CreateRoleOutput, error) {
	tenantID, err := common.ParseTenantID(input.TenantID)
	if err != nil {
		return nil, err
	}

	// Create role entity
	roleEntity, err := role.NewRole(
		tenantID,
		input.Name,
		input.Description,
		input.Color,
		input.DisplayOrder,
	)
	if err != nil {
		return nil, err
	}

	// Save
	if err := u.repo.Save(ctx, roleEntity); err != nil {
		return nil, err
	}

	// Build output
	return &CreateRoleOutput{
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
