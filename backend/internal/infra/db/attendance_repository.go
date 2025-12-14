package db

import (
	"context"
	"database/sql"
	"fmt"
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
			response_id, tenant_id, collection_id, member_id, response, note,
			responded_at, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (collection_id, member_id) DO UPDATE SET
			response = EXCLUDED.response,
			note = EXCLUDED.note,
			responded_at = EXCLUDED.responded_at,
			updated_at = EXCLUDED.updated_at
	`

	executor := GetTx(ctx, r.pool)

	_, err := executor.Exec(ctx, query,
		response.ResponseID().String(),
		response.TenantID().String(),
		response.CollectionID().String(),
		response.MemberID().String(),
		response.Response().String(),
		response.Note(),
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
			response_id, tenant_id, collection_id, member_id, response, note,
			responded_at, created_at, updated_at
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
			responseStr     string
			note            string
			respondedAt     time.Time
			createdAt       time.Time
			updatedAt       time.Time
		)

		err := rows.Scan(
			&responseIDStr,
			&tenantIDStr,
			&collectionIDStr,
			&memberIDStr,
			&responseStr,
			&note,
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

		responseType, err := attendance.NewResponseType(responseStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse response type: %w", err)
		}

		resp, err := attendance.ReconstructAttendanceResponse(
			responseID,
			tenantID,
			colID,
			memberID,
			responseType,
			note,
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
