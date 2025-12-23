package member_group

import (
	"context"
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/member"
)

// CreateGroupUsecase handles creating a member group
type CreateGroupUsecase struct {
	repo member.MemberGroupRepository
}

func NewCreateGroupUsecase(repo member.MemberGroupRepository) *CreateGroupUsecase {
	return &CreateGroupUsecase{repo: repo}
}

func (u *CreateGroupUsecase) Execute(ctx context.Context, input CreateGroupInput) (*CreateGroupOutput, error) {
	tenantID, err := common.ParseTenantID(input.TenantID)
	if err != nil {
		return nil, err
	}

	group, err := member.NewMemberGroup(
		time.Now(),
		tenantID,
		input.Name,
		input.Description,
		input.Color,
		input.DisplayOrder,
	)
	if err != nil {
		return nil, err
	}

	if err := u.repo.Save(ctx, group); err != nil {
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

// UpdateGroupUsecase handles updating a member group
type UpdateGroupUsecase struct {
	repo member.MemberGroupRepository
}

func NewUpdateGroupUsecase(repo member.MemberGroupRepository) *UpdateGroupUsecase {
	return &UpdateGroupUsecase{repo: repo}
}

func (u *UpdateGroupUsecase) Execute(ctx context.Context, input UpdateGroupInput) (*UpdateGroupOutput, error) {
	tenantID, err := common.ParseTenantID(input.TenantID)
	if err != nil {
		return nil, err
	}

	groupID, err := common.ParseMemberGroupID(input.GroupID)
	if err != nil {
		return nil, err
	}

	group, err := u.repo.FindByID(ctx, tenantID, groupID)
	if err != nil {
		return nil, err
	}

	if err := group.UpdateDetails(input.Name, input.Description, input.Color, input.DisplayOrder); err != nil {
		return nil, err
	}

	if err := u.repo.Save(ctx, group); err != nil {
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

// GetGroupUsecase handles getting a member group
type GetGroupUsecase struct {
	repo member.MemberGroupRepository
}

func NewGetGroupUsecase(repo member.MemberGroupRepository) *GetGroupUsecase {
	return &GetGroupUsecase{repo: repo}
}

func (u *GetGroupUsecase) Execute(ctx context.Context, input GetGroupInput) (*GetGroupOutput, error) {
	tenantID, err := common.ParseTenantID(input.TenantID)
	if err != nil {
		return nil, err
	}

	groupID, err := common.ParseMemberGroupID(input.GroupID)
	if err != nil {
		return nil, err
	}

	group, err := u.repo.FindByID(ctx, tenantID, groupID)
	if err != nil {
		return nil, err
	}

	// Get members in this group
	memberIDs, err := u.repo.FindMemberIDsByGroupID(ctx, groupID)
	if err != nil {
		memberIDs = []common.MemberID{}
	}

	memberIDStrings := make([]string, len(memberIDs))
	for i, mid := range memberIDs {
		memberIDStrings[i] = mid.String()
	}

	return &GetGroupOutput{
		Group: GroupDTO{
			GroupID:      group.GroupID().String(),
			TenantID:     group.TenantID().String(),
			Name:         group.Name(),
			Description:  group.Description(),
			Color:        group.Color(),
			DisplayOrder: group.DisplayOrder(),
			MemberIDs:    memberIDStrings,
			CreatedAt:    group.CreatedAt(),
			UpdatedAt:    group.UpdatedAt(),
		},
	}, nil
}

// ListGroupsUsecase handles listing member groups
type ListGroupsUsecase struct {
	repo member.MemberGroupRepository
}

func NewListGroupsUsecase(repo member.MemberGroupRepository) *ListGroupsUsecase {
	return &ListGroupsUsecase{repo: repo}
}

func (u *ListGroupsUsecase) Execute(ctx context.Context, input ListGroupsInput) (*ListGroupsOutput, error) {
	tenantID, err := common.ParseTenantID(input.TenantID)
	if err != nil {
		return nil, err
	}

	groups, err := u.repo.FindByTenantID(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	dtos := make([]GroupDTO, 0, len(groups))
	for _, g := range groups {
		// Get members for each group
		memberIDs, err := u.repo.FindMemberIDsByGroupID(ctx, g.GroupID())
		if err != nil {
			memberIDs = []common.MemberID{}
		}

		memberIDStrings := make([]string, len(memberIDs))
		for i, mid := range memberIDs {
			memberIDStrings[i] = mid.String()
		}

		dtos = append(dtos, GroupDTO{
			GroupID:      g.GroupID().String(),
			TenantID:     g.TenantID().String(),
			Name:         g.Name(),
			Description:  g.Description(),
			Color:        g.Color(),
			DisplayOrder: g.DisplayOrder(),
			MemberIDs:    memberIDStrings,
			CreatedAt:    g.CreatedAt(),
			UpdatedAt:    g.UpdatedAt(),
		})
	}

	return &ListGroupsOutput{Groups: dtos}, nil
}

// DeleteGroupUsecase handles deleting a member group
type DeleteGroupUsecase struct {
	repo member.MemberGroupRepository
}

func NewDeleteGroupUsecase(repo member.MemberGroupRepository) *DeleteGroupUsecase {
	return &DeleteGroupUsecase{repo: repo}
}

func (u *DeleteGroupUsecase) Execute(ctx context.Context, input DeleteGroupInput) (*DeleteGroupOutput, error) {
	tenantID, err := common.ParseTenantID(input.TenantID)
	if err != nil {
		return nil, err
	}

	groupID, err := common.ParseMemberGroupID(input.GroupID)
	if err != nil {
		return nil, err
	}

	if err := u.repo.Delete(ctx, tenantID, groupID); err != nil {
		return nil, err
	}

	return &DeleteGroupOutput{
		GroupID:   input.GroupID,
		DeletedAt: time.Now(),
	}, nil
}

// AssignMembersUsecase handles assigning members to a group
type AssignMembersUsecase struct {
	repo member.MemberGroupRepository
}

func NewAssignMembersUsecase(repo member.MemberGroupRepository) *AssignMembersUsecase {
	return &AssignMembersUsecase{repo: repo}
}

func (u *AssignMembersUsecase) Execute(ctx context.Context, input AssignMembersInput) (*AssignMembersOutput, error) {
	groupID, err := common.ParseMemberGroupID(input.GroupID)
	if err != nil {
		return nil, err
	}

	// Get current members
	currentMemberIDs, err := u.repo.FindMemberIDsByGroupID(ctx, groupID)
	if err != nil {
		currentMemberIDs = []common.MemberID{}
	}

	// Create a map for fast lookup
	currentMap := make(map[string]bool)
	for _, mid := range currentMemberIDs {
		currentMap[mid.String()] = true
	}

	// Create a map of new member IDs
	newMap := make(map[string]bool)
	for _, mid := range input.MemberIDs {
		newMap[mid] = true
	}

	// Remove members not in new list
	for _, mid := range currentMemberIDs {
		if !newMap[mid.String()] {
			if err := u.repo.RemoveMember(ctx, groupID, mid); err != nil {
				return nil, err
			}
		}
	}

	// Add new members
	for _, midStr := range input.MemberIDs {
		if !currentMap[midStr] {
			memberID := common.MemberID(midStr)
			if err := u.repo.AssignMember(ctx, groupID, memberID); err != nil {
				return nil, err
			}
		}
	}

	return &AssignMembersOutput{
		GroupID:   input.GroupID,
		MemberIDs: input.MemberIDs,
	}, nil
}
