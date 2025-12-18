package usecase

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/shift"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Note: ShiftSlotRepository is defined in business_day_usecase.go
// Note: MemberRepository is defined in member_usecase.go
// Note: ShiftAssignmentRepository is extended in shift_slot_usecase.go

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
	dbPool         *pgxpool.Pool
	slotRepo       ShiftSlotRepository
	assignmentRepo ShiftAssignmentRepository
	memberRepo     MemberRepository
}

// NewConfirmManualAssignmentUsecase creates a new ConfirmManualAssignmentUsecase
func NewConfirmManualAssignmentUsecase(
	dbPool *pgxpool.Pool,
	slotRepo ShiftSlotRepository,
	assignmentRepo ShiftAssignmentRepository,
	memberRepo MemberRepository,
) *ConfirmManualAssignmentUsecase {
	return &ConfirmManualAssignmentUsecase{
		dbPool:         dbPool,
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
//  3. Begin transaction
//  4. Count existing confirmed assignments for the slot
//  5. Return ErrSlotFull if count >= required_count
//  6. Create ShiftAssignment
//  7. Save assignment
//  8. Commit transaction
//  9. Log notification stub
// 10. Log audit log stub
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

	// 3. Begin transaction
	tx, err := uc.dbPool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// 4. Count existing confirmed assignments
	query := `
		SELECT COUNT(*)
		FROM shift_assignments
		WHERE tenant_id = $1
		  AND slot_id = $2
		  AND assignment_status = 'confirmed'
		  AND deleted_at IS NULL
	`
	var currentCount int
	err = tx.QueryRow(ctx, query, input.TenantID.String(), input.SlotID.String()).Scan(&currentCount)
	if err != nil {
		return nil, fmt.Errorf("failed to count assignments: %w", err)
	}

	// 5. Return ErrSlotFull if full
	if currentCount >= slot.RequiredCount() {
		return nil, common.NewDomainError(
			common.ErrConflict,
			fmt.Sprintf("slot is full: %d/%d", currentCount, slot.RequiredCount()),
		)
	}

	// 6. Create ShiftAssignment
	var nilPlanID shift.PlanID // Zero value (treated as NULL)
	assignment, err := shift.NewShiftAssignment(
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

	// 7. Save assignment (within transaction)
	err = uc.saveAssignmentInTx(ctx, tx, assignment)
	if err != nil {
		return nil, fmt.Errorf("failed to save shift assignment: %w", err)
	}

	// 8. Commit transaction
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// 9. Notification stub (log output)
	log.Printf("[Notification Stub] シフト確定通知: member=%s, slot=%s, assigned_at=%s",
		memberEntity.DisplayName(),
		slot.SlotName(),
		assignment.AssignedAt().Format("2006-01-02 15:04:05"),
	)

	// 10. AuditLog stub (log output)
	log.Printf("[AuditLog Stub] CREATE ShiftAssignment: actor_id=%s, assignment_id=%s, member_id=%s, slot_id=%s",
		input.ActorID.String(),
		assignment.AssignmentID().String(),
		input.MemberID.String(),
		input.SlotID.String(),
	)

	return assignment, nil
}

// saveAssignmentInTx saves a ShiftAssignment within a transaction
func (uc *ConfirmManualAssignmentUsecase) saveAssignmentInTx(
	ctx context.Context,
	tx pgx.Tx,
	assignment *shift.ShiftAssignment,
) error {
	query := `
		INSERT INTO shift_assignments (
			assignment_id,
			tenant_id,
			plan_id,
			slot_id,
			member_id,
			assignment_status,
			assignment_method,
			is_outside_preference,
			assigned_at,
			created_at,
			updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	_, err := tx.Exec(ctx, query,
		assignment.AssignmentID().String(),
		assignment.TenantID().String(),
		nil, // plan_id is NULL (simplified implementation)
		assignment.SlotID().String(),
		assignment.MemberID().String(),
		"confirmed", // assignment_status
		"manual",    // assignment_method
		assignment.IsOutsidePreference(),
		assignment.AssignedAt(),
		assignment.CreatedAt(),
		assignment.UpdatedAt(),
	)

	return err
}

// GetAssignmentsInput represents the input for getting assignments
type GetAssignmentsInput struct {
	TenantID       common.TenantID
	MemberID       *common.MemberID
	SlotID         *shift.SlotID
	Status         string
	StartDate      *time.Time
	EndDate        *time.Time
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
	dbPool         *pgxpool.Pool
	assignmentRepo ShiftAssignmentRepository
}

// NewGetAssignmentsUsecase creates a new GetAssignmentsUsecase
func NewGetAssignmentsUsecase(
	dbPool *pgxpool.Pool,
	assignmentRepo ShiftAssignmentRepository,
) *GetAssignmentsUsecase {
	return &GetAssignmentsUsecase{
		dbPool:         dbPool,
		assignmentRepo: assignmentRepo,
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
	} else {
		return nil, common.NewValidationError("member_id or slot_id is required", nil)
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

// getAssignmentDetails retrieves JOIN data for an assignment
func (uc *GetAssignmentsUsecase) getAssignmentDetails(
	ctx context.Context,
	assignment *shift.ShiftAssignment,
) (*AssignmentWithDetails, error) {
	query := `
		SELECT
			m.display_name,
			ss.slot_name,
			ebd.target_date,
			ss.start_time,
			ss.end_time
		FROM shift_assignments sa
		INNER JOIN members m ON sa.member_id = m.member_id AND m.deleted_at IS NULL
		INNER JOIN shift_slots ss ON sa.slot_id = ss.slot_id AND ss.deleted_at IS NULL
		INNER JOIN event_business_days ebd ON ss.business_day_id = ebd.business_day_id AND ebd.deleted_at IS NULL
		WHERE sa.assignment_id = $1 AND sa.tenant_id = $2 AND sa.deleted_at IS NULL
	`

	var (
		memberDisplayName string
		slotName          string
		targetDate        time.Time
		startTime         time.Time
		endTime           time.Time
	)

	err := uc.dbPool.QueryRow(ctx, query, assignment.AssignmentID().String(), assignment.TenantID().String()).Scan(
		&memberDisplayName,
		&slotName,
		&targetDate,
		&startTime,
		&endTime,
	)
	if err != nil {
		return nil, err
	}

	return &AssignmentWithDetails{
		Assignment:        assignment,
		MemberDisplayName: memberDisplayName,
		SlotName:          slotName,
		TargetDate:        targetDate,
		StartTime:         startTime,
		EndTime:           endTime,
	}, nil
}

// GetAssignmentDetailInput represents the input for getting an assignment detail
type GetAssignmentDetailInput struct {
	TenantID     common.TenantID
	AssignmentID shift.AssignmentID
}

// GetAssignmentDetailUsecase handles getting a shift assignment detail
type GetAssignmentDetailUsecase struct {
	dbPool         *pgxpool.Pool
	assignmentRepo ShiftAssignmentRepository
}

// NewGetAssignmentDetailUsecase creates a new GetAssignmentDetailUsecase
func NewGetAssignmentDetailUsecase(
	dbPool *pgxpool.Pool,
	assignmentRepo ShiftAssignmentRepository,
) *GetAssignmentDetailUsecase {
	return &GetAssignmentDetailUsecase{
		dbPool:         dbPool,
		assignmentRepo: assignmentRepo,
	}
}

// Execute retrieves a shift assignment detail with JOIN data
func (uc *GetAssignmentDetailUsecase) Execute(
	ctx context.Context,
	input GetAssignmentDetailInput,
) (*AssignmentWithDetails, error) {
	// Get assignment
	assignment, err := uc.assignmentRepo.FindByID(ctx, input.TenantID, input.AssignmentID)
	if err != nil {
		return nil, err
	}

	// Get JOIN data
	query := `
		SELECT
			m.display_name,
			ss.slot_name,
			ebd.target_date,
			ss.start_time,
			ss.end_time
		FROM shift_assignments sa
		INNER JOIN members m ON sa.member_id = m.member_id AND m.deleted_at IS NULL
		INNER JOIN shift_slots ss ON sa.slot_id = ss.slot_id AND ss.deleted_at IS NULL
		INNER JOIN event_business_days ebd ON ss.business_day_id = ebd.business_day_id AND ebd.deleted_at IS NULL
		WHERE sa.assignment_id = $1 AND sa.tenant_id = $2 AND sa.deleted_at IS NULL
	`

	var (
		memberDisplayName string
		slotName          string
		targetDate        time.Time
		startTime         time.Time
		endTime           time.Time
	)

	err = uc.dbPool.QueryRow(ctx, query, input.AssignmentID.String(), input.TenantID.String()).Scan(
		&memberDisplayName,
		&slotName,
		&targetDate,
		&startTime,
		&endTime,
	)
	if err != nil {
		return nil, err
	}

	return &AssignmentWithDetails{
		Assignment:        assignment,
		MemberDisplayName: memberDisplayName,
		SlotName:          slotName,
		TargetDate:        targetDate,
		StartTime:         startTime,
		EndTime:           endTime,
	}, nil
}

// CancelAssignmentInput represents the input for canceling an assignment
type CancelAssignmentInput struct {
	TenantID     common.TenantID
	AssignmentID shift.AssignmentID
}

// CancelAssignmentUsecase handles canceling a shift assignment
type CancelAssignmentUsecase struct {
	assignmentRepo ShiftAssignmentRepository
}

// NewCancelAssignmentUsecase creates a new CancelAssignmentUsecase
func NewCancelAssignmentUsecase(assignmentRepo ShiftAssignmentRepository) *CancelAssignmentUsecase {
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
