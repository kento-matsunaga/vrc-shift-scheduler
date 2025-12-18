package role

import (
	"context"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/role"
)

type GetRoleUsecase struct {
	repo role.RoleRepository
}

func NewGetRoleUsecase(repo role.RoleRepository) *GetRoleUsecase {
	return &GetRoleUsecase{repo: repo}
}

func (u *GetRoleUsecase) Execute(ctx context.Context, input GetRoleInput) (*GetRoleOutput, error) {
	tenantID, err := common.ParseTenantID(input.TenantID)
	if err != nil {
		return nil, err
	}

	roleID, err := common.ParseRoleID(input.RoleID)
	if err != nil {
		return nil, err
	}

	// Find role
	roleEntity, err := u.repo.FindByID(ctx, tenantID, roleID)
	if err != nil {
		return nil, err
	}

	// Build output
	return &GetRoleOutput{
		Role: RoleDTO{
			RoleID:       roleEntity.RoleID().String(),
			TenantID:     roleEntity.TenantID().String(),
			Name:         roleEntity.Name(),
			Description:  roleEntity.Description(),
			Color:        roleEntity.Color(),
			DisplayOrder: roleEntity.DisplayOrder(),
			CreatedAt:    roleEntity.CreatedAt(),
			UpdatedAt:    roleEntity.UpdatedAt(),
		},
	}, nil
}
