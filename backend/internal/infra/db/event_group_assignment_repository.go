package db

import (
	"context"
	"fmt"
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/event"
	"github.com/jackc/pgx/v5/pgxpool"
)

// EventGroupAssignmentRepository implements event.EventGroupAssignmentRepository for PostgreSQL
type EventGroupAssignmentRepository struct {
	db *pgxpool.Pool
}

// NewEventGroupAssignmentRepository creates a new EventGroupAssignmentRepository
func NewEventGroupAssignmentRepository(db *pgxpool.Pool) *EventGroupAssignmentRepository {
	return &EventGroupAssignmentRepository{db: db}
}

// SaveGroupAssignments saves member group assignments for an event
// Replaces all existing assignments with the new ones
func (r *EventGroupAssignmentRepository) SaveGroupAssignments(ctx context.Context, eventID common.EventID, groupIDs []common.MemberGroupID) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Delete existing assignments
	deleteQuery := `DELETE FROM event_group_assignments WHERE event_id = $1`
	_, err = tx.Exec(ctx, deleteQuery, eventID.String())
	if err != nil {
		return fmt.Errorf("failed to delete existing group assignments: %w", err)
	}

	// Insert new assignments
	if len(groupIDs) > 0 {
		insertQuery := `INSERT INTO event_group_assignments (event_id, group_id, created_at) VALUES ($1, $2, $3)`
		now := time.Now()
		for _, groupID := range groupIDs {
			_, err = tx.Exec(ctx, insertQuery, eventID.String(), groupID.String(), now)
			if err != nil {
				return fmt.Errorf("failed to insert group assignment: %w", err)
			}
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// FindGroupAssignmentsByEventID returns all group assignments for an event
func (r *EventGroupAssignmentRepository) FindGroupAssignmentsByEventID(ctx context.Context, eventID common.EventID) ([]*event.EventGroupAssignment, error) {
	query := `
		SELECT event_id, group_id, created_at
		FROM event_group_assignments
		WHERE event_id = $1
		ORDER BY created_at
	`

	rows, err := r.db.Query(ctx, query, eventID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to find group assignments: %w", err)
	}
	defer rows.Close()

	var assignments []*event.EventGroupAssignment
	for rows.Next() {
		var eventIDStr, groupIDStr string
		var createdAt time.Time

		if err := rows.Scan(&eventIDStr, &groupIDStr, &createdAt); err != nil {
			return nil, fmt.Errorf("failed to scan group assignment row: %w", err)
		}

		assignment, err := event.ReconstructEventGroupAssignment(
			common.EventID(eventIDStr),
			common.MemberGroupID(groupIDStr),
			createdAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to reconstruct group assignment: %w", err)
		}

		assignments = append(assignments, assignment)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating group assignment rows: %w", err)
	}

	return assignments, nil
}

// DeleteGroupAssignments deletes all group assignments for an event
func (r *EventGroupAssignmentRepository) DeleteGroupAssignments(ctx context.Context, eventID common.EventID) error {
	query := `DELETE FROM event_group_assignments WHERE event_id = $1`
	_, err := r.db.Exec(ctx, query, eventID.String())
	if err != nil {
		return fmt.Errorf("failed to delete group assignments: %w", err)
	}
	return nil
}

// SaveRoleGroupAssignments saves role group assignments for an event
// Replaces all existing assignments with the new ones
func (r *EventGroupAssignmentRepository) SaveRoleGroupAssignments(ctx context.Context, eventID common.EventID, roleGroupIDs []common.RoleGroupID) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Delete existing assignments
	deleteQuery := `DELETE FROM event_role_group_assignments WHERE event_id = $1`
	_, err = tx.Exec(ctx, deleteQuery, eventID.String())
	if err != nil {
		return fmt.Errorf("failed to delete existing role group assignments: %w", err)
	}

	// Insert new assignments
	if len(roleGroupIDs) > 0 {
		insertQuery := `INSERT INTO event_role_group_assignments (event_id, role_group_id, created_at) VALUES ($1, $2, $3)`
		now := time.Now()
		for _, roleGroupID := range roleGroupIDs {
			_, err = tx.Exec(ctx, insertQuery, eventID.String(), roleGroupID.String(), now)
			if err != nil {
				return fmt.Errorf("failed to insert role group assignment: %w", err)
			}
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// FindRoleGroupAssignmentsByEventID returns all role group assignments for an event
func (r *EventGroupAssignmentRepository) FindRoleGroupAssignmentsByEventID(ctx context.Context, eventID common.EventID) ([]*event.EventRoleGroupAssignment, error) {
	query := `
		SELECT event_id, role_group_id, created_at
		FROM event_role_group_assignments
		WHERE event_id = $1
		ORDER BY created_at
	`

	rows, err := r.db.Query(ctx, query, eventID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to find role group assignments: %w", err)
	}
	defer rows.Close()

	var assignments []*event.EventRoleGroupAssignment
	for rows.Next() {
		var eventIDStr, roleGroupIDStr string
		var createdAt time.Time

		if err := rows.Scan(&eventIDStr, &roleGroupIDStr, &createdAt); err != nil {
			return nil, fmt.Errorf("failed to scan role group assignment row: %w", err)
		}

		assignment, err := event.ReconstructEventRoleGroupAssignment(
			common.EventID(eventIDStr),
			common.RoleGroupID(roleGroupIDStr),
			createdAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to reconstruct role group assignment: %w", err)
		}

		assignments = append(assignments, assignment)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating role group assignment rows: %w", err)
	}

	return assignments, nil
}

// DeleteRoleGroupAssignments deletes all role group assignments for an event
func (r *EventGroupAssignmentRepository) DeleteRoleGroupAssignments(ctx context.Context, eventID common.EventID) error {
	query := `DELETE FROM event_role_group_assignments WHERE event_id = $1`
	_, err := r.db.Exec(ctx, query, eventID.String())
	if err != nil {
		return fmt.Errorf("failed to delete role group assignments: %w", err)
	}
	return nil
}
