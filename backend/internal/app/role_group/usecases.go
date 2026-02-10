package role_group

import (
	"context"
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/role"
)

// CreateGroupUsecase handles creating a new group
type CreateGroupUsecase struct {
	groupRepo role.RoleGroupRepository
}

func NewCreateGroupUsecase(groupRepo role.RoleGroupRepository) *CreateGroupUsecase {
	return &CreateGroupUsecase{groupRepo: groupRepo}
}

func (u *CreateGroupUsecase) Execute(ctx context.Context, input CreateGroupInput) (*CreateGroupOutput, error) {
	tenantID := common.TenantID(input.TenantID)

	now := time.Now()
	group, err := role.NewRoleGroup(now, tenantID, input.Name, input.Description, input.Color, input.DisplayOrder)
	if err != nil {
		return nil, err
	}

	if err := u.groupRepo.Save(ctx, group); err != nil {
		return nil, err
	}

	return &CreateGroupOutput{
		GroupID:      group.GroupID().String(),
		TenantID:     group.TenantID().String(),
		Name:         group.Name(),
		Description:  group.Description(),
		Color:        group.Color(),
		DisplayOrder: group.DisplayOrder(),
		CreatedAt:    group.CreatedAt(),
		UpdatedAt:    group.UpdatedAt(),
	}, nil
}

// UpdateGroupUsecase handles updating a group
type UpdateGroupUsecase struct {
	groupRepo role.RoleGroupRepository
}

func NewUpdateGroupUsecase(groupRepo role.RoleGroupRepository) *UpdateGroupUsecase {
	return &UpdateGroupUsecase{groupRepo: groupRepo}
}

func (u *UpdateGroupUsecase) Execute(ctx context.Context, input UpdateGroupInput) (*UpdateGroupOutput, error) {
	tenantID := common.TenantID(input.TenantID)
	groupID := common.RoleGroupID(input.GroupID)

	group, err := u.groupRepo.FindByID(ctx, tenantID, groupID)
	if err != nil {
		return nil, err
	}
	if group == nil {
		return nil, common.NewNotFoundError("RoleGroup", input.GroupID)
	}

	now := time.Now()
	if err := group.UpdateDetails(now, input.Name, input.Description, input.Color, input.DisplayOrder); err != nil {
		return nil, err
	}

	if err := u.groupRepo.Save(ctx, group); err != nil {
		return nil, err
	}

	return &UpdateGroupOutput{
		GroupID:      group.GroupID().String(),
		TenantID:     group.TenantID().String(),
		Name:         group.Name(),
		Description:  group.Description(),
		Color:        group.Color(),
		DisplayOrder: group.DisplayOrder(),
		CreatedAt:    group.CreatedAt(),
		UpdatedAt:    group.UpdatedAt(),
	}, nil
}

// GetGroupUsecase handles getting a group
type GetGroupUsecase struct {
	groupRepo role.RoleGroupRepository
}

func NewGetGroupUsecase(groupRepo role.RoleGroupRepository) *GetGroupUsecase {
	return &GetGroupUsecase{groupRepo: groupRepo}
}

func (u *GetGroupUsecase) Execute(ctx context.Context, input GetGroupInput) (*GetGroupOutput, error) {
	tenantID := common.TenantID(input.TenantID)
	groupID := common.RoleGroupID(input.GroupID)

	group, err := u.groupRepo.FindByID(ctx, tenantID, groupID)
	if err != nil {
		return nil, err
	}
	if group == nil {
		return nil, common.NewNotFoundError("RoleGroup", input.GroupID)
	}

	roleIDs := make([]string, len(group.RoleIDs()))
	for i, id := range group.RoleIDs() {
		roleIDs[i] = id.String()
	}

	return &GetGroupOutput{
		Group: GroupDTO{
			GroupID:      group.GroupID().String(),
			TenantID:     group.TenantID().String(),
			Name:         group.Name(),
			Description:  group.Description(),
			Color:        group.Color(),
			DisplayOrder: group.DisplayOrder(),
			RoleIDs:      roleIDs,
			CreatedAt:    group.CreatedAt(),
			UpdatedAt:    group.UpdatedAt(),
		},
	}, nil
}

// ListGroupsUsecase handles listing groups
type ListGroupsUsecase struct {
	groupRepo role.RoleGroupRepository
}

func NewListGroupsUsecase(groupRepo role.RoleGroupRepository) *ListGroupsUsecase {
	return &ListGroupsUsecase{groupRepo: groupRepo}
}

func (u *ListGroupsUsecase) Execute(ctx context.Context, input ListGroupsInput) (*ListGroupsOutput, error) {
	tenantID := common.TenantID(input.TenantID)

	groups, err := u.groupRepo.FindByTenantID(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	dtos := make([]GroupDTO, len(groups))
	for i, g := range groups {
		roleIDs := make([]string, len(g.RoleIDs()))
		for j, id := range g.RoleIDs() {
			roleIDs[j] = id.String()
		}

		dtos[i] = GroupDTO{
			GroupID:      g.GroupID().String(),
			TenantID:     g.TenantID().String(),
			Name:         g.Name(),
			Description:  g.Description(),
			Color:        g.Color(),
			DisplayOrder: g.DisplayOrder(),
			RoleIDs:      roleIDs,
			CreatedAt:    g.CreatedAt(),
			UpdatedAt:    g.UpdatedAt(),
		}
	}

	return &ListGroupsOutput{Groups: dtos}, nil
}

// DeleteGroupUsecase handles deleting a group
type DeleteGroupUsecase struct {
	groupRepo role.RoleGroupRepository
}

func NewDeleteGroupUsecase(groupRepo role.RoleGroupRepository) *DeleteGroupUsecase {
	return &DeleteGroupUsecase{groupRepo: groupRepo}
}

func (u *DeleteGroupUsecase) Execute(ctx context.Context, input DeleteGroupInput) (*DeleteGroupOutput, error) {
	tenantID := common.TenantID(input.TenantID)
	groupID := common.RoleGroupID(input.GroupID)

	if err := u.groupRepo.Delete(ctx, tenantID, groupID); err != nil {
		return nil, err
	}

	return &DeleteGroupOutput{
		GroupID:   input.GroupID,
		DeletedAt: time.Now(),
	}, nil
}

// AssignRolesUsecase handles assigning roles to a group
type AssignRolesUsecase struct {
	groupRepo role.RoleGroupRepository
}

func NewAssignRolesUsecase(groupRepo role.RoleGroupRepository) *AssignRolesUsecase {
	return &AssignRolesUsecase{groupRepo: groupRepo}
}

func (u *AssignRolesUsecase) Execute(ctx context.Context, input AssignRolesInput) (*AssignRolesOutput, error) {
	groupID := common.RoleGroupID(input.GroupID)

	roleIDs := make([]common.RoleID, len(input.RoleIDs))
	for i, id := range input.RoleIDs {
		roleIDs[i] = common.RoleID(id)
	}

	if err := u.groupRepo.SetGroupRoles(ctx, groupID, roleIDs); err != nil {
		return nil, err
	}

	return &AssignRolesOutput{
		GroupID: input.GroupID,
		RoleIDs: input.RoleIDs,
	}, nil
}
