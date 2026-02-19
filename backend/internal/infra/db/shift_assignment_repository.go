package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/event"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/shift"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ShiftAssignmentRepository implements shift.ShiftAssignmentRepository for PostgreSQL
type ShiftAssignmentRepository struct {
	db *pgxpool.Pool
}

// NewShiftAssignmentRepository creates a new ShiftAssignmentRepository
func NewShiftAssignmentRepository(db *pgxpool.Pool) *ShiftAssignmentRepository {
	return &ShiftAssignmentRepository{db: db}
}

// Save saves a shift assignment (insert or update)
func (r *ShiftAssignmentRepository) Save(ctx context.Context, assignment *shift.ShiftAssignment) error {
	query := `
		INSERT INTO shift_assignments (
			assignment_id, tenant_id, plan_id, slot_id, member_id,
			assignment_status, assignment_method, is_outside_preference,
			assigned_at, cancelled_at, created_at, updated_at, deleted_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		ON CONFLICT (assignment_id) DO UPDATE SET
			assignment_status = EXCLUDED.assignment_status,
			assignment_method = EXCLUDED.assignment_method,
			is_outside_preference = EXCLUDED.is_outside_preference,
			cancelled_at = EXCLUDED.cancelled_at,
			updated_at = EXCLUDED.updated_at,
			deleted_at = EXCLUDED.deleted_at
	`

	// plan_id が空文字列の場合は NULL を渡す
	var planIDValue interface{}
	if assignment.PlanID().String() == "" {
		planIDValue = nil
	} else {
		planIDValue = assignment.PlanID().String()
	}

	_, err := r.db.Exec(ctx, query,
		assignment.AssignmentID().String(),
		assignment.TenantID().String(),
		planIDValue,
		assignment.SlotID().String(),
		assignment.MemberID().String(),
		string(assignment.AssignmentStatus()),
		string(assignment.AssignmentMethod()),
		assignment.IsOutsidePreference(),
		assignment.AssignedAt(),
		assignment.CancelledAt(),
		assignment.CreatedAt(),
		assignment.UpdatedAt(),
		assignment.DeletedAt(),
	)

	if err != nil {
		return fmt.Errorf("failed to save shift assignment: %w", err)
	}

	return nil
}

// FindByID finds a shift assignment by ID within a tenant
func (r *ShiftAssignmentRepository) FindByID(ctx context.Context, tenantID common.TenantID, assignmentID shift.AssignmentID) (*shift.ShiftAssignment, error) {
	query := `
		SELECT
			assignment_id, tenant_id, plan_id, slot_id, member_id,
			assignment_status, assignment_method, is_outside_preference,
			assigned_at, cancelled_at, created_at, updated_at, deleted_at
		FROM shift_assignments
		WHERE tenant_id = $1 AND assignment_id = $2 AND deleted_at IS NULL
	`

	var (
		assignmentIDStr     string
		tenantIDStr         string
		planIDStr           sql.NullString
		slotIDStr           string
		memberIDStr         string
		assignmentStatusStr string
		assignmentMethodStr string
		isOutsidePreference bool
		assignedAt          time.Time
		cancelledAt         sql.NullTime
		createdAt           time.Time
		updatedAt           time.Time
		deletedAt           sql.NullTime
	)

	err := r.db.QueryRow(ctx, query, tenantID.String(), assignmentID.String()).Scan(
		&assignmentIDStr,
		&tenantIDStr,
		&planIDStr,
		&slotIDStr,
		&memberIDStr,
		&assignmentStatusStr,
		&assignmentMethodStr,
		&isOutsidePreference,
		&assignedAt,
		&cancelledAt,
		&createdAt,
		&updatedAt,
		&deletedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, common.NewNotFoundError("ShiftAssignment", assignmentID.String())
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find shift assignment: %w", err)
	}

	return r.scanToShiftAssignment(
		assignmentIDStr, tenantIDStr, stringValue(planIDStr), slotIDStr, memberIDStr,
		assignmentStatusStr, assignmentMethodStr, isOutsidePreference,
		assignedAt, cancelledAt, createdAt, updatedAt, deletedAt,
	)
}

// FindBySlotID finds all shift assignments for a slot
func (r *ShiftAssignmentRepository) FindBySlotID(ctx context.Context, tenantID common.TenantID, slotID shift.SlotID) ([]*shift.ShiftAssignment, error) {
	query := `
		SELECT
			assignment_id, tenant_id, plan_id, slot_id, member_id,
			assignment_status, assignment_method, is_outside_preference,
			assigned_at, cancelled_at, created_at, updated_at, deleted_at
		FROM shift_assignments
		WHERE tenant_id = $1 AND slot_id = $2 AND deleted_at IS NULL
		ORDER BY assigned_at ASC
	`

	return r.queryShiftAssignments(ctx, query, tenantID.String(), slotID.String())
}

// FindConfirmedBySlotID finds all confirmed shift assignments for a slot
func (r *ShiftAssignmentRepository) FindConfirmedBySlotID(ctx context.Context, tenantID common.TenantID, slotID shift.SlotID) ([]*shift.ShiftAssignment, error) {
	query := `
		SELECT
			assignment_id, tenant_id, plan_id, slot_id, member_id,
			assignment_status, assignment_method, is_outside_preference,
			assigned_at, cancelled_at, created_at, updated_at, deleted_at
		FROM shift_assignments
		WHERE tenant_id = $1 AND slot_id = $2 AND assignment_status = 'confirmed' AND deleted_at IS NULL
		ORDER BY assigned_at ASC
	`

	return r.queryShiftAssignments(ctx, query, tenantID.String(), slotID.String())
}

// FindByMemberID finds all shift assignments for a member
func (r *ShiftAssignmentRepository) FindByMemberID(ctx context.Context, tenantID common.TenantID, memberID common.MemberID) ([]*shift.ShiftAssignment, error) {
	query := `
		SELECT
			assignment_id, tenant_id, plan_id, slot_id, member_id,
			assignment_status, assignment_method, is_outside_preference,
			assigned_at, cancelled_at, created_at, updated_at, deleted_at
		FROM shift_assignments
		WHERE tenant_id = $1 AND member_id = $2 AND deleted_at IS NULL
		ORDER BY assigned_at DESC
	`

	return r.queryShiftAssignments(ctx, query, tenantID.String(), memberID.String())
}

// FindConfirmedByMemberID finds all confirmed shift assignments for a member
func (r *ShiftAssignmentRepository) FindConfirmedByMemberID(ctx context.Context, tenantID common.TenantID, memberID common.MemberID) ([]*shift.ShiftAssignment, error) {
	query := `
		SELECT
			assignment_id, tenant_id, plan_id, slot_id, member_id,
			assignment_status, assignment_method, is_outside_preference,
			assigned_at, cancelled_at, created_at, updated_at, deleted_at
		FROM shift_assignments
		WHERE tenant_id = $1 AND member_id = $2 AND assignment_status = 'confirmed' AND deleted_at IS NULL
		ORDER BY assigned_at DESC
	`

	return r.queryShiftAssignments(ctx, query, tenantID.String(), memberID.String())
}

// FindByPlanID finds all shift assignments for a plan
func (r *ShiftAssignmentRepository) FindByPlanID(ctx context.Context, tenantID common.TenantID, planID shift.PlanID) ([]*shift.ShiftAssignment, error) {
	query := `
		SELECT
			assignment_id, tenant_id, plan_id, slot_id, member_id,
			assignment_status, assignment_method, is_outside_preference,
			assigned_at, cancelled_at, created_at, updated_at, deleted_at
		FROM shift_assignments
		WHERE tenant_id = $1 AND plan_id = $2 AND deleted_at IS NULL
		ORDER BY assigned_at ASC
	`

	return r.queryShiftAssignments(ctx, query, tenantID.String(), planID.String())
}

// CountConfirmedBySlotID counts confirmed assignments for a slot
func (r *ShiftAssignmentRepository) CountConfirmedBySlotID(ctx context.Context, tenantID common.TenantID, slotID shift.SlotID) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM shift_assignments
		WHERE tenant_id = $1 AND slot_id = $2 AND assignment_status = 'confirmed' AND deleted_at IS NULL
	`

	var count int
	err := r.db.QueryRow(ctx, query, tenantID.String(), slotID.String()).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count confirmed assignments: %w", err)
	}

	return count, nil
}

// Delete deletes a shift assignment (physical delete)
func (r *ShiftAssignmentRepository) Delete(ctx context.Context, tenantID common.TenantID, assignmentID shift.AssignmentID) error {
	query := `
		DELETE FROM shift_assignments
		WHERE tenant_id = $1 AND assignment_id = $2
	`

	result, err := r.db.Exec(ctx, query, tenantID.String(), assignmentID.String())
	if err != nil {
		return fmt.Errorf("failed to delete shift assignment: %w", err)
	}

	if result.RowsAffected() == 0 {
		return common.NewNotFoundError("ShiftAssignment", assignmentID.String())
	}

	return nil
}

// ExistsBySlotIDAndMemberID checks if a confirmed assignment exists for the given slot and member
func (r *ShiftAssignmentRepository) ExistsBySlotIDAndMemberID(ctx context.Context, tenantID common.TenantID, slotID shift.SlotID, memberID common.MemberID) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM shift_assignments
			WHERE tenant_id = $1 AND slot_id = $2 AND member_id = $3 AND assignment_status = 'confirmed' AND deleted_at IS NULL
		)
	`

	var exists bool
	err := r.db.QueryRow(ctx, query, tenantID.String(), slotID.String(), memberID.String()).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check shift assignment existence: %w", err)
	}

	return exists, nil
}

// HasConfirmedByMemberAndBusinessDayID checks if a confirmed assignment exists for the given member and business day
func (r *ShiftAssignmentRepository) HasConfirmedByMemberAndBusinessDayID(ctx context.Context, tenantID common.TenantID, memberID common.MemberID, businessDayID event.BusinessDayID) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM shift_assignments sa
			INNER JOIN shift_slots ss ON sa.slot_id = ss.slot_id AND ss.deleted_at IS NULL
			WHERE sa.tenant_id = $1
			  AND sa.member_id = $2
			  AND ss.business_day_id = $3
			  AND sa.assignment_status = 'confirmed'
			  AND sa.deleted_at IS NULL
		)
	`

	var exists bool
	err := r.db.QueryRow(ctx, query, tenantID.String(), memberID.String(), string(businessDayID)).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check member attendance: %w", err)
	}

	return exists, nil
}

// FindByBusinessDayID finds all shift assignments for a business day
func (r *ShiftAssignmentRepository) FindByBusinessDayID(ctx context.Context, tenantID common.TenantID, businessDayID event.BusinessDayID) ([]*shift.ShiftAssignment, error) {
	query := `
		SELECT
			sa.assignment_id, sa.tenant_id, sa.plan_id, sa.slot_id, sa.member_id,
			sa.assignment_status, sa.assignment_method, sa.is_outside_preference,
			sa.assigned_at, sa.cancelled_at, sa.created_at, sa.updated_at, sa.deleted_at
		FROM shift_assignments sa
		INNER JOIN shift_slots ss ON sa.slot_id = ss.slot_id AND ss.deleted_at IS NULL
		WHERE sa.tenant_id = $1 AND ss.business_day_id = $2 AND sa.deleted_at IS NULL
		ORDER BY ss.start_time ASC, sa.assigned_at ASC
	`

	return r.queryShiftAssignments(ctx, query, tenantID.String(), string(businessDayID))
}

// queryShiftAssignments executes a query and returns a list of shift assignments
func (r *ShiftAssignmentRepository) queryShiftAssignments(ctx context.Context, query string, args ...interface{}) ([]*shift.ShiftAssignment, error) {
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query shift assignments: %w", err)
	}
	defer rows.Close()

	var assignments []*shift.ShiftAssignment
	for rows.Next() {
		var (
			assignmentIDStr     string
			tenantIDStr         string
			planIDStr           sql.NullString
			slotIDStr           string
			memberIDStr         string
			assignmentStatusStr string
			assignmentMethodStr string
			isOutsidePreference bool
			assignedAt          time.Time
			cancelledAt         sql.NullTime
			createdAt           time.Time
			updatedAt           time.Time
			deletedAt           sql.NullTime
		)

		err := rows.Scan(
			&assignmentIDStr,
			&tenantIDStr,
			&planIDStr,
			&slotIDStr,
			&memberIDStr,
			&assignmentStatusStr,
			&assignmentMethodStr,
			&isOutsidePreference,
			&assignedAt,
			&cancelledAt,
			&createdAt,
			&updatedAt,
			&deletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan shift assignment row: %w", err)
		}

		assignment, err := r.scanToShiftAssignment(
			assignmentIDStr, tenantIDStr, stringValue(planIDStr), slotIDStr, memberIDStr,
			assignmentStatusStr, assignmentMethodStr, isOutsidePreference,
			assignedAt, cancelledAt, createdAt, updatedAt, deletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to reconstruct shift assignment: %w", err)
		}

		assignments = append(assignments, assignment)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating shift assignment rows: %w", err)
	}

	return assignments, nil
}

// scanToShiftAssignment converts scanned row data to ShiftAssignment entity
func (r *ShiftAssignmentRepository) scanToShiftAssignment(
	assignmentIDStr, tenantIDStr, planIDStr, slotIDStr, memberIDStr string,
	assignmentStatusStr, assignmentMethodStr string,
	isOutsidePreference bool,
	assignedAt time.Time,
	cancelledAt sql.NullTime,
	createdAt, updatedAt time.Time,
	deletedAt sql.NullTime,
) (*shift.ShiftAssignment, error) {
	var cancelledAtPtr *time.Time
	if cancelledAt.Valid {
		cancelledAtPtr = &cancelledAt.Time
	}

	var deletedAtPtr *time.Time
	if deletedAt.Valid {
		deletedAtPtr = &deletedAt.Time
	}

	return shift.ReconstructShiftAssignment(
		shift.AssignmentID(assignmentIDStr),
		common.TenantID(tenantIDStr),
		shift.PlanID(planIDStr),
		shift.SlotID(slotIDStr),
		common.MemberID(memberIDStr),
		shift.AssignmentStatus(assignmentStatusStr),
		shift.AssignmentMethod(assignmentMethodStr),
		isOutsidePreference,
		assignedAt,
		cancelledAtPtr,
		createdAt,
		updatedAt,
		deletedAtPtr,
	)
}
