package role

import (
	"context"
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/role"
)

type DeleteRoleUsecase struct {
	repo role.RoleRepository
}

func NewDeleteRoleUsecase(repo role.RoleRepository) *DeleteRoleUsecase {
	return &DeleteRoleUsecase{repo: repo}
}

func (u *DeleteRoleUsecase) Execute(ctx context.Context, input DeleteRoleInput) (*DeleteRoleOutput, error) {
	tenantID, err := common.ParseTenantID(input.TenantID)
	if err != nil {
		return nil, err
	}

	roleID, err := common.ParseRoleID(input.RoleID)
	if err != nil {
		return nil, err
	}

	// Delete role (soft delete)
	if err := u.repo.Delete(ctx, tenantID, roleID); err != nil {
		return nil, err
	}

	// Build output
	return &DeleteRoleOutput{
		RoleID:    roleID.String(),
		DeletedAt: time.Now(),
	}, nil
}
