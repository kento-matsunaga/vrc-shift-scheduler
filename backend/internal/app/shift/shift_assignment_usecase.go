package shift

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/event"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/member"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/shift"
)

// ConfirmManualAssignmentInput represents the input for confirming a manual shift assignment
type ConfirmManualAssignmentInput struct {
	TenantID common.TenantID
	SlotID   shift.SlotID
	MemberID common.MemberID
	ActorID  common.MemberID
	Note     string
}

// ConfirmManualAssignmentUsecase handles manual shift assignment confirmation
type ConfirmManualAssignmentUsecase struct {
	slotRepo       shift.ShiftSlotRepository
	assignmentRepo shift.ShiftAssignmentRepository
	memberRepo     member.MemberRepository
}

// NewConfirmManualAssignmentUsecase creates a new ConfirmManualAssignmentUsecase
func NewConfirmManualAssignmentUsecase(
	slotRepo shift.ShiftSlotRepository,
	assignmentRepo shift.ShiftAssignmentRepository,
	memberRepo member.MemberRepository,
) *ConfirmManualAssignmentUsecase {
	return &ConfirmManualAssignmentUsecase{
		slotRepo:       slotRepo,
		assignmentRepo: assignmentRepo,
		memberRepo:     memberRepo,
	}
}

// Execute confirms a manual shift assignment
//
// Logic:
//  1. Get ShiftSlot (with tenant_id check)
//  2. Get Member (with tenant_id check)
//  3. Count existing confirmed assignments for the slot
//  4. Return ErrSlotFull if count >= required_count
//  5. Create ShiftAssignment
//  6. Save assignment
//  7. Log notification stub
//  8. Log audit log stub
func (uc *ConfirmManualAssignmentUsecase) Execute(
	ctx context.Context,
	input ConfirmManualAssignmentInput,
) (*shift.ShiftAssignment, error) {
	// 1. Get ShiftSlot (with tenant_id check)
	slot, err := uc.slotRepo.FindByID(ctx, input.TenantID, input.SlotID)
	if err != nil {
		return nil, fmt.Errorf("failed to find shift slot: %w", err)
	}

	// 2. Get Member (with tenant_id check)
	memberEntity, err := uc.memberRepo.FindByID(ctx, input.TenantID, input.MemberID)
	if err != nil {
		return nil, fmt.Errorf("failed to find member: %w", err)
	}

	// 3. Count existing confirmed assignments
	currentCount, err := uc.assignmentRepo.CountConfirmedBySlotID(ctx, input.TenantID, input.SlotID)
	if err != nil {
		return nil, fmt.Errorf("failed to count assignments: %w", err)
	}

	// 4. Return ErrSlotFull if full
	if currentCount >= slot.RequiredCount() {
		return nil, common.NewDomainError(
			common.ErrConflict,
			fmt.Sprintf("slot is full: %d/%d", currentCount, slot.RequiredCount()),
		)
	}

	// 5. Create ShiftAssignment
	now := time.Now()
	var nilPlanID shift.PlanID // Zero value (treated as NULL)
	assignment, err := shift.NewShiftAssignment(
		now,
		input.TenantID,
		nilPlanID,
		input.SlotID,
		input.MemberID,
		shift.AssignmentMethodManual,
		false, // is_outside_preference
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create shift assignment: %w", err)
	}

	// 6. Save assignment
	if err := uc.assignmentRepo.Save(ctx, assignment); err != nil {
		return nil, fmt.Errorf("failed to save shift assignment: %w", err)
	}

	// 7. Notification stub (log output)
	log.Printf("[Notification Stub] シフト確定通知: member=%s, slot=%s, assigned_at=%s",
		memberEntity.DisplayName(),
		slot.SlotName(),
		assignment.AssignedAt().Format("2006-01-02 15:04:05"),
	)

	// 8. AuditLog stub (log output)
	log.Printf("[AuditLog Stub] CREATE ShiftAssignment: actor_id=%s, assignment_id=%s, member_id=%s, slot_id=%s",
		input.ActorID.String(),
		assignment.AssignmentID().String(),
		input.MemberID.String(),
		input.SlotID.String(),
	)

	return assignment, nil
}

// GetAssignmentsInput represents the input for getting assignments
type GetAssignmentsInput struct {
	TenantID      common.TenantID
	MemberID      *common.MemberID
	SlotID        *shift.SlotID
	BusinessDayID *string
	Status        string
	StartDate     *time.Time
	EndDate       *time.Time
}

// AssignmentWithDetails represents an assignment with JOIN data
type AssignmentWithDetails struct {
	Assignment        *shift.ShiftAssignment
	MemberDisplayName string
	SlotName          string
	TargetDate        time.Time
	StartTime         time.Time
	EndTime           time.Time
}

// GetAssignmentsUsecase handles getting shift assignments
type GetAssignmentsUsecase struct {
	assignmentRepo  shift.ShiftAssignmentRepository
	memberRepo      member.MemberRepository
	slotRepo        shift.ShiftSlotRepository
	businessDayRepo event.EventBusinessDayRepository
}

// NewGetAssignmentsUsecase creates a new GetAssignmentsUsecase
func NewGetAssignmentsUsecase(
	assignmentRepo shift.ShiftAssignmentRepository,
	memberRepo member.MemberRepository,
	slotRepo shift.ShiftSlotRepository,
	businessDayRepo event.EventBusinessDayRepository,
) *GetAssignmentsUsecase {
	return &GetAssignmentsUsecase{
		assignmentRepo:  assignmentRepo,
		memberRepo:      memberRepo,
		slotRepo:        slotRepo,
		businessDayRepo: businessDayRepo,
	}
}

// Execute retrieves shift assignments with filtering
func (uc *GetAssignmentsUsecase) Execute(
	ctx context.Context,
	input GetAssignmentsInput,
) ([]*AssignmentWithDetails, error) {
	var assignments []*shift.ShiftAssignment
	var err error

	// Get assignments based on filter
	if input.MemberID != nil {
		assignments, err = uc.assignmentRepo.FindByMemberID(ctx, input.TenantID, *input.MemberID)
	} else if input.SlotID != nil {
		assignments, err = uc.assignmentRepo.FindBySlotID(ctx, input.TenantID, *input.SlotID)
	} else if input.BusinessDayID != nil {
		assignments, err = uc.assignmentRepo.FindByBusinessDayID(ctx, input.TenantID, event.BusinessDayID(*input.BusinessDayID))
	} else {
		return nil, common.NewValidationError("member_id, slot_id, or business_day_id is required", nil)
	}

	if err != nil {
		return nil, err
	}

	// Build response with JOIN data
	var result []*AssignmentWithDetails
	for _, assignment := range assignments {
		// Filter by status
		if input.Status != "" {
			if (input.Status == "confirmed" && assignment.IsCancelled()) ||
				(input.Status == "cancelled" && !assignment.IsCancelled()) {
				continue
			}
		}

		// Get JOIN data
		details, err := uc.getAssignmentDetails(ctx, assignment)
		if err != nil {
			continue // Skip on error
		}

		// Filter by date range
		if input.StartDate != nil && details.TargetDate.Before(*input.StartDate) {
			continue
		}
		if input.EndDate != nil && details.TargetDate.After(*input.EndDate) {
			continue
		}

		result = append(result, details)
	}

	return result, nil
}

// getAssignmentDetails retrieves details for an assignment using repositories
func (uc *GetAssignmentsUsecase) getAssignmentDetails(
	ctx context.Context,
	assignment *shift.ShiftAssignment,
) (*AssignmentWithDetails, error) {
	// Get member
	memberEntity, err := uc.memberRepo.FindByID(ctx, assignment.TenantID(), assignment.MemberID())
	if err != nil {
		return nil, err
	}

	// Get shift slot
	slotEntity, err := uc.slotRepo.FindByID(ctx, assignment.TenantID(), assignment.SlotID())
	if err != nil {
		return nil, err
	}

	// Get business day
	businessDay, err := uc.businessDayRepo.FindByID(ctx, assignment.TenantID(), slotEntity.BusinessDayID())
	if err != nil {
		return nil, err
	}

	return &AssignmentWithDetails{
		Assignment:        assignment,
		MemberDisplayName: memberEntity.DisplayName(),
		SlotName:          slotEntity.SlotName(),
		TargetDate:        businessDay.TargetDate(),
		StartTime:         slotEntity.StartTime(),
		EndTime:           slotEntity.EndTime(),
	}, nil
}

// GetAssignmentDetailInput represents the input for getting an assignment detail
type GetAssignmentDetailInput struct {
	TenantID     common.TenantID
	AssignmentID shift.AssignmentID
}

// GetAssignmentDetailUsecase handles getting a shift assignment detail
type GetAssignmentDetailUsecase struct {
	assignmentRepo  shift.ShiftAssignmentRepository
	memberRepo      member.MemberRepository
	slotRepo        shift.ShiftSlotRepository
	businessDayRepo event.EventBusinessDayRepository
}

// NewGetAssignmentDetailUsecase creates a new GetAssignmentDetailUsecase
func NewGetAssignmentDetailUsecase(
	assignmentRepo shift.ShiftAssignmentRepository,
	memberRepo member.MemberRepository,
	slotRepo shift.ShiftSlotRepository,
	businessDayRepo event.EventBusinessDayRepository,
) *GetAssignmentDetailUsecase {
	return &GetAssignmentDetailUsecase{
		assignmentRepo:  assignmentRepo,
		memberRepo:      memberRepo,
		slotRepo:        slotRepo,
		businessDayRepo: businessDayRepo,
	}
}

// Execute retrieves a shift assignment detail using repositories
func (uc *GetAssignmentDetailUsecase) Execute(
	ctx context.Context,
	input GetAssignmentDetailInput,
) (*AssignmentWithDetails, error) {
	// Get assignment
	assignment, err := uc.assignmentRepo.FindByID(ctx, input.TenantID, input.AssignmentID)
	if err != nil {
		return nil, err
	}

	// Get member
	memberEntity, err := uc.memberRepo.FindByID(ctx, input.TenantID, assignment.MemberID())
	if err != nil {
		return nil, err
	}

	// Get shift slot
	slotEntity, err := uc.slotRepo.FindByID(ctx, input.TenantID, assignment.SlotID())
	if err != nil {
		return nil, err
	}

	// Get business day
	businessDay, err := uc.businessDayRepo.FindByID(ctx, input.TenantID, slotEntity.BusinessDayID())
	if err != nil {
		return nil, err
	}

	return &AssignmentWithDetails{
		Assignment:        assignment,
		MemberDisplayName: memberEntity.DisplayName(),
		SlotName:          slotEntity.SlotName(),
		TargetDate:        businessDay.TargetDate(),
		StartTime:         slotEntity.StartTime(),
		EndTime:           slotEntity.EndTime(),
	}, nil
}

// CancelAssignmentInput represents the input for canceling an assignment
type CancelAssignmentInput struct {
	TenantID     common.TenantID
	AssignmentID shift.AssignmentID
}

// CancelAssignmentUsecase handles canceling a shift assignment
type CancelAssignmentUsecase struct {
	assignmentRepo shift.ShiftAssignmentRepository
}

// NewCancelAssignmentUsecase creates a new CancelAssignmentUsecase
func NewCancelAssignmentUsecase(assignmentRepo shift.ShiftAssignmentRepository) *CancelAssignmentUsecase {
	return &CancelAssignmentUsecase{
		assignmentRepo: assignmentRepo,
	}
}

// Execute cancels a shift assignment (logical delete)
func (uc *CancelAssignmentUsecase) Execute(
	ctx context.Context,
	input CancelAssignmentInput,
) error {
	return uc.assignmentRepo.Delete(ctx, input.TenantID, input.AssignmentID)
}
