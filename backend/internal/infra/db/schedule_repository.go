package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/schedule"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// pgtypeTimeToTimePtr converts pgtype.Time to *time.Time (nil if not valid)
func pgtypeTimeToTimePtr(pt pgtype.Time) *time.Time {
	if !pt.Valid {
		return nil
	}
	t := time.Date(0, 1, 1, int(pt.Microseconds/3600000000), int((pt.Microseconds%3600000000)/60000000), int((pt.Microseconds%60000000)/1000000), 0, time.UTC)
	return &t
}

// ScheduleRepository implements schedule.DateScheduleRepository for PostgreSQL
type ScheduleRepository struct {
	pool *pgxpool.Pool
}

// NewScheduleRepository creates a new ScheduleRepository
func NewScheduleRepository(pool *pgxpool.Pool) *ScheduleRepository {
	return &ScheduleRepository{pool: pool}
}

// Save saves a schedule with its candidates
func (r *ScheduleRepository) Save(ctx context.Context, s *schedule.DateSchedule) error {
	executor := GetTx(ctx, r.pool)

	// Save schedule
	query := `
		INSERT INTO date_schedules (
			schedule_id, tenant_id, title, description, event_id,
			public_token, status, deadline, decided_candidate_id,
			created_at, updated_at, deleted_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		ON CONFLICT (schedule_id) DO UPDATE SET
			title = EXCLUDED.title,
			description = EXCLUDED.description,
			status = EXCLUDED.status,
			deadline = EXCLUDED.deadline,
			decided_candidate_id = EXCLUDED.decided_candidate_id,
			updated_at = EXCLUDED.updated_at,
			deleted_at = EXCLUDED.deleted_at
	`

	var eventIDStr *string
	if s.EventID() != nil {
		str := s.EventID().String()
		eventIDStr = &str
	}

	var decidedCandidateIDStr *string
	if s.DecidedCandidateID() != nil {
		str := s.DecidedCandidateID().String()
		decidedCandidateIDStr = &str
	}

	_, err := executor.Exec(ctx, query,
		s.ScheduleID().String(),
		s.TenantID().String(),
		s.Title(),
		s.Description(),
		eventIDStr,
		s.PublicToken().String(),
		s.Status().String(),
		s.Deadline(),
		decidedCandidateIDStr,
		s.CreatedAt(),
		s.UpdatedAt(),
		s.DeletedAt(),
	)
	if err != nil {
		return fmt.Errorf("failed to save schedule: %w", err)
	}

	// Save candidates
	for _, candidate := range s.Candidates() {
		candidateQuery := `
			INSERT INTO schedule_candidates (
				candidate_id, schedule_id, candidate_date, start_time, end_time, display_order, created_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7)
			ON CONFLICT (candidate_id) DO NOTHING
		`
		_, err := executor.Exec(ctx, candidateQuery,
			candidate.CandidateID().String(),
			candidate.ScheduleID().String(),
			candidate.CandidateDateValue(),
			candidate.StartTime(),
			candidate.EndTime(),
			candidate.DisplayOrder(),
			candidate.CreatedAt(),
		)
		if err != nil {
			return fmt.Errorf("failed to save candidate: %w", err)
		}
	}

	return nil
}

// FindByID finds a schedule by ID
func (r *ScheduleRepository) FindByID(ctx context.Context, tenantID common.TenantID, id common.ScheduleID) (*schedule.DateSchedule, error) {
	executor := GetTx(ctx, r.pool)

	query := `
		SELECT schedule_id, tenant_id, title, description, event_id, public_token, status,
			deadline, decided_candidate_id, created_at, updated_at, deleted_at
		FROM date_schedules
		WHERE tenant_id = $1 AND schedule_id = $2 AND deleted_at IS NULL
	`

	var (
		scheduleIDStr, tenantIDStr, title, description, publicTokenStr, statusStr string
		eventIDStr, decidedCandidateIDStr                                         *string
		deadline, deletedAt                                                       sql.NullTime
		createdAt, updatedAt                                                      time.Time
	)

	err := executor.QueryRow(ctx, query, tenantID.String(), id.String()).Scan(
		&scheduleIDStr, &tenantIDStr, &title, &description, &eventIDStr, &publicTokenStr, &statusStr,
		&deadline, &decidedCandidateIDStr, &createdAt, &updatedAt, &deletedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, common.NewNotFoundError("DateSchedule", id.String())
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find schedule: %w", err)
	}

	// Load candidates
	candidates, err := r.FindCandidatesByScheduleID(ctx, id)
	if err != nil {
		return nil, err
	}

	return r.scanSchedule(scheduleIDStr, tenantIDStr, title, description, eventIDStr, publicTokenStr, statusStr,
		deadline, decidedCandidateIDStr, createdAt, updatedAt, deletedAt, candidates)
}

// FindByToken finds a schedule by public token
func (r *ScheduleRepository) FindByToken(ctx context.Context, token common.PublicToken) (*schedule.DateSchedule, error) {
	executor := GetTx(ctx, r.pool)

	query := `
		SELECT schedule_id, tenant_id, title, description, event_id, public_token, status,
			deadline, decided_candidate_id, created_at, updated_at, deleted_at
		FROM date_schedules
		WHERE public_token = $1 AND deleted_at IS NULL
	`

	var (
		scheduleIDStr, tenantIDStr, title, description, publicTokenStr, statusStr string
		eventIDStr, decidedCandidateIDStr                                         *string
		deadline, deletedAt                                                       sql.NullTime
		createdAt, updatedAt                                                      time.Time
	)

	err := executor.QueryRow(ctx, query, token.String()).Scan(
		&scheduleIDStr, &tenantIDStr, &title, &description, &eventIDStr, &publicTokenStr, &statusStr,
		&deadline, &decidedCandidateIDStr, &createdAt, &updatedAt, &deletedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, common.NewNotFoundError("DateSchedule", token.String())
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find schedule by token: %w", err)
	}

	scheduleID, _ := common.ParseScheduleID(scheduleIDStr)
	candidates, err := r.FindCandidatesByScheduleID(ctx, scheduleID)
	if err != nil {
		return nil, err
	}

	return r.scanSchedule(scheduleIDStr, tenantIDStr, title, description, eventIDStr, publicTokenStr, statusStr,
		deadline, decidedCandidateIDStr, createdAt, updatedAt, deletedAt, candidates)
}

// FindByTenantID finds all schedules within a tenant
func (r *ScheduleRepository) FindByTenantID(ctx context.Context, tenantID common.TenantID) ([]*schedule.DateSchedule, error) {
	executor := GetTx(ctx, r.pool)

	query := `
		SELECT schedule_id, tenant_id, title, description, event_id, public_token, status,
			deadline, decided_candidate_id, created_at, updated_at, deleted_at
		FROM date_schedules
		WHERE tenant_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
	`

	rows, err := executor.Query(ctx, query, tenantID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to find schedules by tenant: %w", err)
	}
	defer rows.Close()

	var schedules []*schedule.DateSchedule
	for rows.Next() {
		var (
			scheduleIDStr, tenantIDStr, title, description, publicTokenStr, statusStr string
			eventIDStr, decidedCandidateIDStr                                         *string
			deadline, deletedAt                                                       sql.NullTime
			createdAt, updatedAt                                                      time.Time
		)

		err := rows.Scan(&scheduleIDStr, &tenantIDStr, &title, &description, &eventIDStr, &publicTokenStr, &statusStr,
			&deadline, &decidedCandidateIDStr, &createdAt, &updatedAt, &deletedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan schedule: %w", err)
		}

		scheduleID, _ := common.ParseScheduleID(scheduleIDStr)
		candidates, err := r.FindCandidatesByScheduleID(ctx, scheduleID)
		if err != nil {
			return nil, err
		}

		s, err := r.scanSchedule(scheduleIDStr, tenantIDStr, title, description, eventIDStr, publicTokenStr, statusStr,
			deadline, decidedCandidateIDStr, createdAt, updatedAt, deletedAt, candidates)
		if err != nil {
			return nil, err
		}
		schedules = append(schedules, s)
	}

	return schedules, nil
}

// UpsertResponse upserts a schedule response
func (r *ScheduleRepository) UpsertResponse(ctx context.Context, response *schedule.DateScheduleResponse) error {
	query := `
		INSERT INTO schedule_responses (
			response_id, tenant_id, schedule_id, member_id, candidate_id, availability, note,
			responded_at, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (schedule_id, member_id, candidate_id) DO UPDATE SET
			availability = EXCLUDED.availability,
			note = EXCLUDED.note,
			responded_at = EXCLUDED.responded_at,
			updated_at = EXCLUDED.updated_at
	`

	executor := GetTx(ctx, r.pool)

	_, err := executor.Exec(ctx, query,
		response.ResponseID().String(),
		response.TenantID().String(),
		response.ScheduleID().String(),
		response.MemberID().String(),
		response.CandidateID().String(),
		response.Availability().String(),
		response.Note(),
		response.RespondedAt(),
		response.CreatedAt(),
		response.UpdatedAt(),
	)

	return err
}

// FindResponsesByScheduleID finds all responses for a schedule
func (r *ScheduleRepository) FindResponsesByScheduleID(ctx context.Context, scheduleID common.ScheduleID) ([]*schedule.DateScheduleResponse, error) {
	query := `
		SELECT response_id, tenant_id, schedule_id, member_id, candidate_id, availability, note,
			responded_at, created_at, updated_at
		FROM schedule_responses
		WHERE schedule_id = $1
		ORDER BY responded_at DESC
	`

	executor := GetTx(ctx, r.pool)
	rows, err := executor.Query(ctx, query, scheduleID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var responses []*schedule.DateScheduleResponse
	for rows.Next() {
		var (
			responseIDStr, tenantIDStr, scheduleIDStr, memberIDStr, candidateIDStr, availabilityStr, note string
			respondedAt, createdAt, updatedAt                                                             time.Time
		)

		err := rows.Scan(&responseIDStr, &tenantIDStr, &scheduleIDStr, &memberIDStr, &candidateIDStr, &availabilityStr, &note,
			&respondedAt, &createdAt, &updatedAt)
		if err != nil {
			return nil, err
		}

		responseID, _ := common.ParseResponseID(responseIDStr)
		tenantID, _ := common.ParseTenantID(tenantIDStr)
		schedID, _ := common.ParseScheduleID(scheduleIDStr)
		memberID, _ := common.ParseMemberID(memberIDStr)
		candidateID, _ := common.ParseCandidateID(candidateIDStr)
		availability, _ := schedule.NewAvailability(availabilityStr)

		resp, err := schedule.ReconstructDateScheduleResponse(responseID, tenantID, schedID, memberID, candidateID,
			availability, note, respondedAt, createdAt, updatedAt)
		if err != nil {
			return nil, err
		}
		responses = append(responses, resp)
	}

	return responses, nil
}

// FindCandidatesByScheduleID finds all candidates for a schedule
func (r *ScheduleRepository) FindCandidatesByScheduleID(ctx context.Context, scheduleID common.ScheduleID) ([]*schedule.CandidateDate, error) {
	query := `
		SELECT candidate_id, schedule_id, candidate_date, start_time, end_time, display_order, created_at
		FROM schedule_candidates
		WHERE schedule_id = $1
		ORDER BY display_order
	`

	executor := GetTx(ctx, r.pool)
	rows, err := executor.Query(ctx, query, scheduleID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var candidates []*schedule.CandidateDate
	for rows.Next() {
		var (
			candidateIDStr, scheduleIDStr string
			candidateDate                 time.Time
			startTime, endTime            pgtype.Time
			displayOrder                  int
			createdAt                     time.Time
		)

		err := rows.Scan(&candidateIDStr, &scheduleIDStr, &candidateDate, &startTime, &endTime, &displayOrder, &createdAt)
		if err != nil {
			return nil, err
		}

		candidateID, _ := common.ParseCandidateID(candidateIDStr)
		schedID, _ := common.ParseScheduleID(scheduleIDStr)

		candidate, err := schedule.ReconstructCandidateDate(candidateID, schedID, candidateDate, pgtypeTimeToTimePtr(startTime), pgtypeTimeToTimePtr(endTime), displayOrder, createdAt)
		if err != nil {
			return nil, err
		}
		candidates = append(candidates, candidate)
	}

	return candidates, nil
}

func (r *ScheduleRepository) scanSchedule(
	scheduleIDStr, tenantIDStr, title, description string,
	eventIDStr *string, publicTokenStr, statusStr string,
	deadline sql.NullTime, decidedCandidateIDStr *string,
	createdAt, updatedAt time.Time, deletedAt sql.NullTime,
	candidates []*schedule.CandidateDate,
) (*schedule.DateSchedule, error) {
	scheduleID, _ := common.ParseScheduleID(scheduleIDStr)
	tenantID, _ := common.ParseTenantID(tenantIDStr)
	publicToken, _ := common.ParsePublicToken(publicTokenStr)
	status, _ := schedule.NewStatus(statusStr)

	var eventID *common.EventID
	if eventIDStr != nil {
		eid, _ := common.ParseEventID(*eventIDStr)
		eventID = &eid
	}

	var deadlinePtr *time.Time
	if deadline.Valid {
		deadlinePtr = &deadline.Time
	}

	var decidedCandidateID *common.CandidateID
	if decidedCandidateIDStr != nil {
		cid, _ := common.ParseCandidateID(*decidedCandidateIDStr)
		decidedCandidateID = &cid
	}

	var deletedAtPtr *time.Time
	if deletedAt.Valid {
		deletedAtPtr = &deletedAt.Time
	}

	return schedule.ReconstructDateSchedule(scheduleID, tenantID, title, description, eventID, publicToken, status,
		deadlinePtr, decidedCandidateID, candidates, createdAt, updatedAt, deletedAtPtr)
}

// SaveGroupAssignments saves group assignments for a schedule (deletes existing ones first)
func (r *ScheduleRepository) SaveGroupAssignments(ctx context.Context, scheduleID common.ScheduleID, assignments []*schedule.ScheduleGroupAssignment) error {
	executor := GetTx(ctx, r.pool)

	// 既存のグループ割り当てを削除
	deleteQuery := `DELETE FROM date_schedule_group_assignments WHERE schedule_id = $1`
	_, err := executor.Exec(ctx, deleteQuery, scheduleID.String())
	if err != nil {
		return fmt.Errorf("failed to delete old group assignments: %w", err)
	}

	// 新しいグループ割り当てを挿入
	if len(assignments) == 0 {
		return nil
	}

	insertQuery := `
		INSERT INTO date_schedule_group_assignments (
			schedule_id, group_id, created_at
		) VALUES ($1, $2, $3)
	`

	for _, a := range assignments {
		_, err := executor.Exec(ctx, insertQuery,
			a.ScheduleID().String(),
			a.GroupID().String(),
			a.CreatedAt(),
		)
		if err != nil {
			return fmt.Errorf("failed to save group assignment: %w", err)
		}
	}

	return nil
}

// FindGroupAssignmentsByScheduleID finds all group assignments for a schedule
func (r *ScheduleRepository) FindGroupAssignmentsByScheduleID(ctx context.Context, scheduleID common.ScheduleID) ([]*schedule.ScheduleGroupAssignment, error) {
	query := `
		SELECT schedule_id, group_id, created_at
		FROM date_schedule_group_assignments
		WHERE schedule_id = $1
		ORDER BY created_at
	`

	executor := GetTx(ctx, r.pool)
	rows, err := executor.Query(ctx, query, scheduleID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to query group assignments: %w", err)
	}
	defer rows.Close()

	var assignments []*schedule.ScheduleGroupAssignment
	for rows.Next() {
		var scheduleIDStr, groupIDStr string
		var createdAt time.Time

		err := rows.Scan(&scheduleIDStr, &groupIDStr, &createdAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan group assignment: %w", err)
		}

		parsedScheduleID, err := common.ParseScheduleID(scheduleIDStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse schedule_id: %w", err)
		}

		groupID, err := common.ParseMemberGroupID(groupIDStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse group_id: %w", err)
		}

		assignment, err := schedule.ReconstructScheduleGroupAssignment(parsedScheduleID, groupID, createdAt)
		if err != nil {
			return nil, fmt.Errorf("failed to reconstruct group assignment: %w", err)
		}

		assignments = append(assignments, assignment)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return assignments, nil
}
