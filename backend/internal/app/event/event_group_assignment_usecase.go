package event

import (
	"context"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/event"
)

// GetEventGroupAssignmentsInput represents the input for getting event group assignments
type GetEventGroupAssignmentsInput struct {
	TenantID common.TenantID
	EventID  common.EventID
}

// GetEventGroupAssignmentsOutput represents the output of getting event group assignments
type GetEventGroupAssignmentsOutput struct {
	MemberGroupIDs []string
	RoleGroupIDs   []string
}

// GetEventGroupAssignmentsUsecase handles getting group assignments for an event
type GetEventGroupAssignmentsUsecase struct {
	eventRepo       event.EventRepository
	groupAssignRepo event.EventGroupAssignmentRepository
}

// NewGetEventGroupAssignmentsUsecase creates a new GetEventGroupAssignmentsUsecase
func NewGetEventGroupAssignmentsUsecase(
	eventRepo event.EventRepository,
	groupAssignRepo event.EventGroupAssignmentRepository,
) *GetEventGroupAssignmentsUsecase {
	return &GetEventGroupAssignmentsUsecase{
		eventRepo:       eventRepo,
		groupAssignRepo: groupAssignRepo,
	}
}

// Execute retrieves group assignments for an event
func (uc *GetEventGroupAssignmentsUsecase) Execute(ctx context.Context, input GetEventGroupAssignmentsInput) (*GetEventGroupAssignmentsOutput, error) {
	// Verify event exists and belongs to tenant
	_, err := uc.eventRepo.FindByID(ctx, input.TenantID, input.EventID)
	if err != nil {
		return nil, err
	}

	// Get member group assignments
	memberGroupAssignments, err := uc.groupAssignRepo.FindGroupAssignmentsByEventID(ctx, input.EventID)
	if err != nil {
		return nil, err
	}

	// Get role group assignments
	roleGroupAssignments, err := uc.groupAssignRepo.FindRoleGroupAssignmentsByEventID(ctx, input.EventID)
	if err != nil {
		return nil, err
	}

	// Convert to string arrays
	memberGroupIDs := make([]string, len(memberGroupAssignments))
	for i, a := range memberGroupAssignments {
		memberGroupIDs[i] = a.GroupID().String()
	}

	roleGroupIDs := make([]string, len(roleGroupAssignments))
	for i, a := range roleGroupAssignments {
		roleGroupIDs[i] = a.RoleGroupID().String()
	}

	return &GetEventGroupAssignmentsOutput{
		MemberGroupIDs: memberGroupIDs,
		RoleGroupIDs:   roleGroupIDs,
	}, nil
}

// UpdateEventGroupAssignmentsInput represents the input for updating event group assignments
type UpdateEventGroupAssignmentsInput struct {
	TenantID       common.TenantID
	EventID        common.EventID
	MemberGroupIDs []string
	RoleGroupIDs   []string
}

// UpdateEventGroupAssignmentsUsecase handles updating group assignments for an event
type UpdateEventGroupAssignmentsUsecase struct {
	eventRepo       event.EventRepository
	groupAssignRepo event.EventGroupAssignmentRepository
}

// NewUpdateEventGroupAssignmentsUsecase creates a new UpdateEventGroupAssignmentsUsecase
func NewUpdateEventGroupAssignmentsUsecase(
	eventRepo event.EventRepository,
	groupAssignRepo event.EventGroupAssignmentRepository,
) *UpdateEventGroupAssignmentsUsecase {
	return &UpdateEventGroupAssignmentsUsecase{
		eventRepo:       eventRepo,
		groupAssignRepo: groupAssignRepo,
	}
}

// Execute updates group assignments for an event
func (uc *UpdateEventGroupAssignmentsUsecase) Execute(ctx context.Context, input UpdateEventGroupAssignmentsInput) error {
	// Verify event exists and belongs to tenant
	_, err := uc.eventRepo.FindByID(ctx, input.TenantID, input.EventID)
	if err != nil {
		return err
	}

	// Parse member group IDs
	memberGroupIDs := make([]common.MemberGroupID, 0, len(input.MemberGroupIDs))
	for _, id := range input.MemberGroupIDs {
		parsed, err := common.ParseMemberGroupID(id)
		if err != nil {
			return err
		}
		memberGroupIDs = append(memberGroupIDs, parsed)
	}

	// Parse role group IDs
	roleGroupIDs := make([]common.RoleGroupID, 0, len(input.RoleGroupIDs))
	for _, id := range input.RoleGroupIDs {
		parsed, err := common.ParseRoleGroupID(id)
		if err != nil {
			return err
		}
		roleGroupIDs = append(roleGroupIDs, parsed)
	}

	// Save member group assignments
	if err := uc.groupAssignRepo.SaveGroupAssignments(ctx, input.EventID, memberGroupIDs); err != nil {
		return err
	}

	// Save role group assignments
	if err := uc.groupAssignRepo.SaveRoleGroupAssignments(ctx, input.EventID, roleGroupIDs); err != nil {
		return err
	}

	return nil
}
