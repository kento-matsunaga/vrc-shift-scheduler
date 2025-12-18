package role

import (
	"context"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/role"
)

type ListRolesUsecase struct {
	repo role.RoleRepository
}

func NewListRolesUsecase(repo role.RoleRepository) *ListRolesUsecase {
	return &ListRolesUsecase{repo: repo}
}

func (u *ListRolesUsecase) Execute(ctx context.Context, input ListRolesInput) (*ListRolesOutput, error) {
	tenantID, err := common.ParseTenantID(input.TenantID)
	if err != nil {
		return nil, err
	}

	// Find all roles
	roles, err := u.repo.FindByTenantID(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	// Build output
	roleDTOs := make([]RoleDTO, 0, len(roles))
	for _, roleEntity := range roles {
		roleDTOs = append(roleDTOs, RoleDTO{
			RoleID:       roleEntity.RoleID().String(),
			TenantID:     roleEntity.TenantID().String(),
			Name:         roleEntity.Name(),
			Description:  roleEntity.Description(),
			Color:        roleEntity.Color(),
			DisplayOrder: roleEntity.DisplayOrder(),
			CreatedAt:    roleEntity.CreatedAt(),
			UpdatedAt:    roleEntity.UpdatedAt(),
		})
	}

	return &ListRolesOutput{
		Roles: roleDTOs,
	}, nil
}
