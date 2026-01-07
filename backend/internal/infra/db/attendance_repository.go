package db

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/attendance"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// AttendanceRepository implements attendance.AttendanceCollectionRepository for PostgreSQL
type AttendanceRepository struct {
	pool *pgxpool.Pool
}

// NewAttendanceRepository creates a new AttendanceRepository
func NewAttendanceRepository(pool *pgxpool.Pool) *AttendanceRepository {
	return &AttendanceRepository{pool: pool}
}

// Save saves a collection (insert or update)
func (r *AttendanceRepository) Save(ctx context.Context, c *attendance.AttendanceCollection) error {
	query := `
		INSERT INTO attendance_collections (
			collection_id, tenant_id, title, description, target_type, target_id,
			public_token, status, deadline, created_at, updated_at, deleted_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		ON CONFLICT (collection_id) DO UPDATE SET
			title = EXCLUDED.title,
			description = EXCLUDED.description,
			target_type = EXCLUDED.target_type,
			target_id = EXCLUDED.target_id,
			status = EXCLUDED.status,
			deadline = EXCLUDED.deadline,
			updated_at = EXCLUDED.updated_at,
			deleted_at = EXCLUDED.deleted_at
	`

	executor := GetTx(ctx, r.pool)

	_, err := executor.Exec(ctx, query,
		c.CollectionID().String(),
		c.TenantID().String(),
		c.Title(),
		c.Description(),
		c.TargetType().String(),
		c.TargetID(),
		c.PublicToken().String(),
		c.Status().String(),
		c.Deadline(),
		c.CreatedAt(),
		c.UpdatedAt(),
		c.DeletedAt(),
	)

	if err != nil {
		return fmt.Errorf("failed to save attendance collection: %w", err)
	}

	return nil
}

// FindByID finds a collection by ID within a tenant
func (r *AttendanceRepository) FindByID(ctx context.Context, tenantID common.TenantID, id common.CollectionID) (*attendance.AttendanceCollection, error) {
	query := `
		SELECT
			collection_id, tenant_id, title, description, target_type, target_id,
			public_token, status, deadline, created_at, updated_at, deleted_at
		FROM attendance_collections
		WHERE tenant_id = $1 AND collection_id = $2 AND deleted_at IS NULL
	`

	executor := GetTx(ctx, r.pool)

	var (
		collectionIDStr string
		tenantIDStr     string
		title           string
		description     string
		targetTypeStr   string
		targetID        string
		publicTokenStr  string
		statusStr       string
		deadline        sql.NullTime
		createdAt       time.Time
		updatedAt       time.Time
		deletedAt       sql.NullTime
	)

	err := executor.QueryRow(ctx, query, tenantID.String(), id.String()).Scan(
		&collectionIDStr,
		&tenantIDStr,
		&title,
		&description,
		&targetTypeStr,
		&targetID,
		&publicTokenStr,
		&statusStr,
		&deadline,
		&createdAt,
		&updatedAt,
		&deletedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, common.NewNotFoundError("AttendanceCollection", id.String())
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find attendance collection: %w", err)
	}

	return r.scanCollection(
		collectionIDStr, tenantIDStr, title, description, targetTypeStr, targetID,
		publicTokenStr, statusStr, deadline, createdAt, updatedAt, deletedAt,
	)
}

// FindByToken finds a collection by public token
func (r *AttendanceRepository) FindByToken(ctx context.Context, token common.PublicToken) (*attendance.AttendanceCollection, error) {
	query := `
		SELECT
			collection_id, tenant_id, title, description, target_type, target_id,
			public_token, status, deadline, created_at, updated_at, deleted_at
		FROM attendance_collections
		WHERE public_token = $1 AND deleted_at IS NULL
	`

	executor := GetTx(ctx, r.pool)

	var (
		collectionIDStr string
		tenantIDStr     string
		title           string
		description     string
		targetTypeStr   string
		targetID        string
		publicTokenStr  string
		statusStr       string
		deadline        sql.NullTime
		createdAt       time.Time
		updatedAt       time.Time
		deletedAt       sql.NullTime
	)

	err := executor.QueryRow(ctx, query, token.String()).Scan(
		&collectionIDStr,
		&tenantIDStr,
		&title,
		&description,
		&targetTypeStr,
		&targetID,
		&publicTokenStr,
		&statusStr,
		&deadline,
		&createdAt,
		&updatedAt,
		&deletedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, common.NewNotFoundError("AttendanceCollection", token.String())
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find attendance collection by token: %w", err)
	}

	return r.scanCollection(
		collectionIDStr, tenantIDStr, title, description, targetTypeStr, targetID,
		publicTokenStr, statusStr, deadline, createdAt, updatedAt, deletedAt,
	)
}

// FindByTenantID finds all collections within a tenant
func (r *AttendanceRepository) FindByTenantID(ctx context.Context, tenantID common.TenantID) ([]*attendance.AttendanceCollection, error) {
	query := `
		SELECT
			collection_id, tenant_id, title, description, target_type, target_id,
			public_token, status, deadline, created_at, updated_at, deleted_at
		FROM attendance_collections
		WHERE tenant_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
	`

	executor := GetTx(ctx, r.pool)

	rows, err := executor.Query(ctx, query, tenantID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to find collections by tenant: %w", err)
	}
	defer rows.Close()

	var collections []*attendance.AttendanceCollection
	for rows.Next() {
		var (
			collectionIDStr string
			tenantIDStr     string
			title           string
			description     string
			targetTypeStr   string
			targetID        string
			publicTokenStr  string
			statusStr       string
			deadline        sql.NullTime
			createdAt       time.Time
			updatedAt       time.Time
			deletedAt       sql.NullTime
		)

		err := rows.Scan(
			&collectionIDStr,
			&tenantIDStr,
			&title,
			&description,
			&targetTypeStr,
			&targetID,
			&publicTokenStr,
			&statusStr,
			&deadline,
			&createdAt,
			&updatedAt,
			&deletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan collection: %w", err)
		}

		collection, err := r.scanCollection(
			collectionIDStr, tenantIDStr, title, description, targetTypeStr, targetID,
			publicTokenStr, statusStr, deadline, createdAt, updatedAt, deletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to reconstruct collection: %w", err)
		}

		collections = append(collections, collection)
	}

	return collections, nil
}

// UpsertResponse は回答を登録/更新する（ON CONFLICT DO UPDATE）
func (r *AttendanceRepository) UpsertResponse(ctx context.Context, response *attendance.AttendanceResponse) error {
	query := `
		INSERT INTO attendance_responses (
			response_id, tenant_id, collection_id, member_id, target_date_id, response, note,
			available_from, available_to, responded_at, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		ON CONFLICT (collection_id, member_id, target_date_id) DO UPDATE SET
			response = EXCLUDED.response,
			note = EXCLUDED.note,
			available_from = EXCLUDED.available_from,
			available_to = EXCLUDED.available_to,
			responded_at = EXCLUDED.responded_at,
			updated_at = EXCLUDED.updated_at
	`

	executor := GetTx(ctx, r.pool)

	_, err := executor.Exec(ctx, query,
		response.ResponseID().String(),
		response.TenantID().String(),
		response.CollectionID().String(),
		response.MemberID().String(),
		response.TargetDateID().String(),
		response.Response().String(),
		response.Note(),
		response.AvailableFrom(),
		response.AvailableTo(),
		response.RespondedAt(),
		response.CreatedAt(),
		response.UpdatedAt(),
	)

	if err != nil {
		return fmt.Errorf("failed to upsert attendance response: %w", err)
	}

	return nil
}

// FindResponsesByCollectionID は collection の回答一覧を取得する
func (r *AttendanceRepository) FindResponsesByCollectionID(ctx context.Context, collectionID common.CollectionID) ([]*attendance.AttendanceResponse, error) {
	query := `
		SELECT
			response_id, tenant_id, collection_id, member_id, target_date_id, response, note,
			to_char(available_from, 'HH24:MI'), to_char(available_to, 'HH24:MI'), responded_at, created_at, updated_at
		FROM attendance_responses
		WHERE collection_id = $1
		ORDER BY responded_at DESC
	`

	executor := GetTx(ctx, r.pool)

	rows, err := executor.Query(ctx, query, collectionID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to find responses by collection: %w", err)
	}
	defer rows.Close()

	var responses []*attendance.AttendanceResponse
	for rows.Next() {
		var (
			responseIDStr   string
			tenantIDStr     string
			collectionIDStr string
			memberIDStr     string
			targetDateIDStr string
			responseStr     string
			note            string
			availableFrom   sql.NullString
			availableTo     sql.NullString
			respondedAt     time.Time
			createdAt       time.Time
			updatedAt       time.Time
		)

		err := rows.Scan(
			&responseIDStr,
			&tenantIDStr,
			&collectionIDStr,
			&memberIDStr,
			&targetDateIDStr,
			&responseStr,
			&note,
			&availableFrom,
			&availableTo,
			&respondedAt,
			&createdAt,
			&updatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan response: %w", err)
		}

		responseID, err := common.ParseResponseID(responseIDStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse response_id: %w", err)
		}

		tenantID, err := common.ParseTenantID(tenantIDStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse tenant_id: %w", err)
		}

		colID, err := common.ParseCollectionID(collectionIDStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse collection_id: %w", err)
		}

		memberID, err := common.ParseMemberID(memberIDStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse member_id: %w", err)
		}

		targetDateID, err := common.ParseTargetDateID(targetDateIDStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse target_date_id: %w", err)
		}

		responseType, err := attendance.NewResponseType(responseStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse response type: %w", err)
		}

		var availableFromPtr, availableToPtr *string
		if availableFrom.Valid {
			availableFromPtr = &availableFrom.String
		}
		if availableTo.Valid {
			availableToPtr = &availableTo.String
		}

		resp, err := attendance.ReconstructAttendanceResponse(
			responseID,
			tenantID,
			colID,
			memberID,
			targetDateID,
			responseType,
			note,
			availableFromPtr,
			availableToPtr,
			respondedAt,
			createdAt,
			updatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to reconstruct response: %w", err)
		}

		responses = append(responses, resp)
	}

	return responses, nil
}

// FindResponsesByMemberID は member の回答一覧を取得する
func (r *AttendanceRepository) FindResponsesByMemberID(ctx context.Context, tenantID common.TenantID, memberID common.MemberID) ([]*attendance.AttendanceResponse, error) {
	query := `
		SELECT
			response_id, tenant_id, collection_id, member_id, target_date_id, response, note,
			to_char(available_from, 'HH24:MI'), to_char(available_to, 'HH24:MI'), responded_at, created_at, updated_at
		FROM attendance_responses
		WHERE tenant_id = $1 AND member_id = $2
		ORDER BY responded_at DESC
	`

	executor := GetTx(ctx, r.pool)

	rows, err := executor.Query(ctx, query, tenantID.String(), memberID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to find responses by member: %w", err)
	}
	defer rows.Close()

	var responses []*attendance.AttendanceResponse
	for rows.Next() {
		var (
			responseIDStr   string
			tenantIDStr     string
			collectionIDStr string
			memberIDStr     string
			targetDateIDStr string
			responseStr     string
			note            string
			availableFrom   sql.NullString
			availableTo     sql.NullString
			respondedAt     time.Time
			createdAt       time.Time
			updatedAt       time.Time
		)

		err := rows.Scan(
			&responseIDStr,
			&tenantIDStr,
			&collectionIDStr,
			&memberIDStr,
			&targetDateIDStr,
			&responseStr,
			&note,
			&availableFrom,
			&availableTo,
			&respondedAt,
			&createdAt,
			&updatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan response: %w", err)
		}

		responseID, err := common.ParseResponseID(responseIDStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse response_id: %w", err)
		}

		tid, err := common.ParseTenantID(tenantIDStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse tenant_id: %w", err)
		}

		colID, err := common.ParseCollectionID(collectionIDStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse collection_id: %w", err)
		}

		mid, err := common.ParseMemberID(memberIDStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse member_id: %w", err)
		}

		targetDateID, err := common.ParseTargetDateID(targetDateIDStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse target_date_id: %w", err)
		}

		responseType, err := attendance.NewResponseType(responseStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse response type: %w", err)
		}

		var availableFromPtr, availableToPtr *string
		if availableFrom.Valid {
			availableFromPtr = &availableFrom.String
		}
		if availableTo.Valid {
			availableToPtr = &availableTo.String
		}

		resp, err := attendance.ReconstructAttendanceResponse(
			responseID,
			tid,
			colID,
			mid,
			targetDateID,
			responseType,
			note,
			availableFromPtr,
			availableToPtr,
			respondedAt,
			createdAt,
			updatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to reconstruct response: %w", err)
		}

		responses = append(responses, resp)
	}

	return responses, nil
}

// FindResponsesByCollectionIDAndMemberID は collection 内の特定 member の回答一覧を取得する
// tenant_id でスコープすることでクロステナントアクセスを防止
func (r *AttendanceRepository) FindResponsesByCollectionIDAndMemberID(ctx context.Context, tenantID common.TenantID, collectionID common.CollectionID, memberID common.MemberID) ([]*attendance.AttendanceResponse, error) {
	query := `
		SELECT
			response_id, tenant_id, collection_id, member_id, target_date_id, response, note,
			to_char(available_from, 'HH24:MI'), to_char(available_to, 'HH24:MI'), responded_at, created_at, updated_at
		FROM attendance_responses
		WHERE tenant_id = $1 AND collection_id = $2 AND member_id = $3
		ORDER BY responded_at DESC
	`

	executor := GetTx(ctx, r.pool)

	rows, err := executor.Query(ctx, query, tenantID.String(), collectionID.String(), memberID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to find responses by collection and member: %w", err)
	}
	defer rows.Close()

	var responses []*attendance.AttendanceResponse
	for rows.Next() {
		var (
			responseIDStr   string
			tenantIDStr     string
			collectionIDStr string
			memberIDStr     string
			targetDateIDStr string
			responseStr     string
			note            string
			availableFrom   sql.NullString
			availableTo     sql.NullString
			respondedAt     time.Time
			createdAt       time.Time
			updatedAt       time.Time
		)

		err := rows.Scan(
			&responseIDStr,
			&tenantIDStr,
			&collectionIDStr,
			&memberIDStr,
			&targetDateIDStr,
			&responseStr,
			&note,
			&availableFrom,
			&availableTo,
			&respondedAt,
			&createdAt,
			&updatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan response: %w", err)
		}

		responseID, err := common.ParseResponseID(responseIDStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse response_id: %w", err)
		}

		tid, err := common.ParseTenantID(tenantIDStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse tenant_id: %w", err)
		}

		colID, err := common.ParseCollectionID(collectionIDStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse collection_id: %w", err)
		}

		mid, err := common.ParseMemberID(memberIDStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse member_id: %w", err)
		}

		targetDateID, err := common.ParseTargetDateID(targetDateIDStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse target_date_id: %w", err)
		}

		responseType, err := attendance.NewResponseType(responseStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse response type: %w", err)
		}

		var availableFromPtr, availableToPtr *string
		if availableFrom.Valid {
			availableFromPtr = &availableFrom.String
		}
		if availableTo.Valid {
			availableToPtr = &availableTo.String
		}

		resp, err := attendance.ReconstructAttendanceResponse(
			responseID,
			tid,
			colID,
			mid,
			targetDateID,
			responseType,
			note,
			availableFromPtr,
			availableToPtr,
			respondedAt,
			createdAt,
			updatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to reconstruct response: %w", err)
		}

		responses = append(responses, resp)
	}

	return responses, nil
}

// scanCollection is a helper to reconstruct an AttendanceCollection from DB row
func (r *AttendanceRepository) scanCollection(
	collectionIDStr, tenantIDStr, title, description, targetTypeStr, targetID,
	publicTokenStr, statusStr string,
	deadline sql.NullTime,
	createdAt, updatedAt time.Time,
	deletedAt sql.NullTime,
) (*attendance.AttendanceCollection, error) {
	collectionID, err := common.ParseCollectionID(collectionIDStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse collection_id: %w", err)
	}

	tenantID, err := common.ParseTenantID(tenantIDStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse tenant_id: %w", err)
	}

	targetType, err := attendance.NewTargetType(targetTypeStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse target_type: %w", err)
	}

	publicToken, err := common.ParsePublicToken(publicTokenStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public_token: %w", err)
	}

	status, err := attendance.NewStatus(statusStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse status: %w", err)
	}

	var deadlinePtr *time.Time
	if deadline.Valid {
		deadlinePtr = &deadline.Time
	}

	var deletedAtPtr *time.Time
	if deletedAt.Valid {
		deletedAtPtr = &deletedAt.Time
	}

	return attendance.ReconstructAttendanceCollection(
		collectionID,
		tenantID,
		title,
		description,
		targetType,
		targetID,
		publicToken,
		status,
		deadlinePtr,
		createdAt,
		updatedAt,
		deletedAtPtr,
	)
}

// SaveTargetDates saves target dates for a collection
func (r *AttendanceRepository) SaveTargetDates(ctx context.Context, collectionID common.CollectionID, targetDates []*attendance.TargetDate) error {
	executor := GetTx(ctx, r.pool)

	// 既存の対象日を削除
	deleteQuery := `DELETE FROM attendance_target_dates WHERE collection_id = $1`
	_, err := executor.Exec(ctx, deleteQuery, collectionID.String())
	if err != nil {
		return fmt.Errorf("failed to delete old target dates: %w", err)
	}

	// 新しい対象日を挿入
	if len(targetDates) == 0 {
		return nil
	}

	insertQuery := `
		INSERT INTO attendance_target_dates (
			target_date_id, collection_id, target_date, start_time, end_time, display_order, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	for _, td := range targetDates {
		_, err := executor.Exec(ctx, insertQuery,
			td.TargetDateID().String(),
			td.CollectionID().String(),
			td.TargetDateValue(),
			td.StartTime(),
			td.EndTime(),
			td.DisplayOrder(),
			td.CreatedAt(),
		)
		if err != nil {
			return fmt.Errorf("failed to save target date: %w", err)
		}
	}

	return nil
}

// FindTargetDatesByCollectionID finds all target dates for a collection
func (r *AttendanceRepository) FindTargetDatesByCollectionID(ctx context.Context, collectionID common.CollectionID) ([]*attendance.TargetDate, error) {
	query := `
		SELECT target_date_id, collection_id, target_date,
		       to_char(start_time, 'HH24:MI'), to_char(end_time, 'HH24:MI'),
		       display_order, created_at
		FROM attendance_target_dates
		WHERE collection_id = $1
		ORDER BY display_order, target_date
	`

	executor := GetTx(ctx, r.pool)
	rows, err := executor.Query(ctx, query, collectionID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to query target dates: %w", err)
	}
	defer rows.Close()

	var targetDates []*attendance.TargetDate
	for rows.Next() {
		var targetDateIDStr, collectionIDStr string
		var targetDate time.Time
		var startTime, endTime sql.NullString
		var displayOrder int
		var createdAt time.Time

		err := rows.Scan(&targetDateIDStr, &collectionIDStr, &targetDate, &startTime, &endTime, &displayOrder, &createdAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan target date: %w", err)
		}

		// Convert sql.NullString to *string
		var startTimePtr, endTimePtr *string
		if startTime.Valid {
			startTimePtr = &startTime.String
		}
		if endTime.Valid {
			endTimePtr = &endTime.String
		}

		targetDateID, err := common.ParseTargetDateID(targetDateIDStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse target_date_id: %w", err)
		}

		parsedCollectionID, err := common.ParseCollectionID(collectionIDStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse collection_id: %w", err)
		}

		td, err := attendance.ReconstructTargetDate(targetDateID, parsedCollectionID, targetDate, startTimePtr, endTimePtr, displayOrder, createdAt)
		if err != nil {
			return nil, fmt.Errorf("failed to reconstruct target date: %w", err)
		}

		targetDates = append(targetDates, td)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return targetDates, nil
}

// SaveGroupAssignments saves group assignments for a collection (deletes existing ones first)
func (r *AttendanceRepository) SaveGroupAssignments(ctx context.Context, collectionID common.CollectionID, assignments []*attendance.CollectionGroupAssignment) error {
	executor := GetTx(ctx, r.pool)

	// 既存のグループ割り当てを削除
	deleteQuery := `DELETE FROM attendance_collection_group_assignments WHERE collection_id = $1`
	_, err := executor.Exec(ctx, deleteQuery, collectionID.String())
	if err != nil {
		return fmt.Errorf("failed to delete old group assignments: %w", err)
	}

	// 新しいグループ割り当てを挿入
	if len(assignments) == 0 {
		return nil
	}

	insertQuery := `
		INSERT INTO attendance_collection_group_assignments (
			collection_id, group_id, created_at
		) VALUES ($1, $2, $3)
	`

	for _, a := range assignments {
		_, err := executor.Exec(ctx, insertQuery,
			a.CollectionID().String(),
			a.GroupID().String(),
			a.CreatedAt(),
		)
		if err != nil {
			return fmt.Errorf("failed to save group assignment: %w", err)
		}
	}

	return nil
}

// FindGroupAssignmentsByCollectionID finds all group assignments for a collection
func (r *AttendanceRepository) FindGroupAssignmentsByCollectionID(ctx context.Context, collectionID common.CollectionID) ([]*attendance.CollectionGroupAssignment, error) {
	query := `
		SELECT collection_id, group_id, created_at
		FROM attendance_collection_group_assignments
		WHERE collection_id = $1
		ORDER BY created_at
	`

	executor := GetTx(ctx, r.pool)
	rows, err := executor.Query(ctx, query, collectionID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to query group assignments: %w", err)
	}
	defer rows.Close()

	var assignments []*attendance.CollectionGroupAssignment
	for rows.Next() {
		var collectionIDStr, groupIDStr string
		var createdAt time.Time

		err := rows.Scan(&collectionIDStr, &groupIDStr, &createdAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan group assignment: %w", err)
		}

		parsedCollectionID, err := common.ParseCollectionID(collectionIDStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse collection_id: %w", err)
		}

		groupID, err := common.ParseMemberGroupID(groupIDStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse group_id: %w", err)
		}

		assignment, err := attendance.ReconstructCollectionGroupAssignment(parsedCollectionID, groupID, createdAt)
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

// SaveRoleAssignments saves role assignments for a collection (deletes existing ones first)
// Uses multi-row INSERT for atomicity and performance
func (r *AttendanceRepository) SaveRoleAssignments(ctx context.Context, collectionID common.CollectionID, assignments []*attendance.CollectionRoleAssignment) error {
	executor := GetTx(ctx, r.pool)

	// 既存のロール割り当てを削除
	deleteQuery := `DELETE FROM attendance_collection_role_assignments WHERE collection_id = $1`
	_, err := executor.Exec(ctx, deleteQuery, collectionID.String())
	if err != nil {
		return fmt.Errorf("failed to delete old role assignments: %w", err)
	}

	// 新しいロール割り当てを挿入
	if len(assignments) == 0 {
		return nil
	}

	// マルチロー INSERT で一括挿入（部分的な失敗を防ぐ）
	// VALUES ($1, $2, $3), ($4, $5, $6), ... の形式で構築
	valueStrings := make([]string, 0, len(assignments))
	args := make([]interface{}, 0, len(assignments)*3)

	for i, a := range assignments {
		valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d, $%d)", i*3+1, i*3+2, i*3+3))
		args = append(args, a.CollectionID().String(), a.RoleID().String(), a.CreatedAt())
	}

	insertQuery := fmt.Sprintf(`
		INSERT INTO attendance_collection_role_assignments (
			collection_id, role_id, created_at
		) VALUES %s
	`, strings.Join(valueStrings, ", "))

	_, err = executor.Exec(ctx, insertQuery, args...)
	if err != nil {
		return fmt.Errorf("failed to save role assignments: %w", err)
	}

	return nil
}

// FindRoleAssignmentsByCollectionID finds all role assignments for a collection
func (r *AttendanceRepository) FindRoleAssignmentsByCollectionID(ctx context.Context, collectionID common.CollectionID) ([]*attendance.CollectionRoleAssignment, error) {
	query := `
		SELECT collection_id, role_id, created_at
		FROM attendance_collection_role_assignments
		WHERE collection_id = $1
		ORDER BY created_at
	`

	executor := GetTx(ctx, r.pool)
	rows, err := executor.Query(ctx, query, collectionID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to query role assignments: %w", err)
	}
	defer rows.Close()

	var assignments []*attendance.CollectionRoleAssignment
	for rows.Next() {
		var collectionIDStr, roleIDStr string
		var createdAt time.Time

		err := rows.Scan(&collectionIDStr, &roleIDStr, &createdAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan role assignment: %w", err)
		}

		parsedCollectionID, err := common.ParseCollectionID(collectionIDStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse collection_id: %w", err)
		}

		roleID, err := common.ParseRoleID(roleIDStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse role_id: %w", err)
		}

		assignment, err := attendance.ReconstructCollectionRoleAssignment(parsedCollectionID, roleID, createdAt)
		if err != nil {
			return nil, fmt.Errorf("failed to reconstruct role assignment: %w", err)
		}

		assignments = append(assignments, assignment)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return assignments, nil
}
