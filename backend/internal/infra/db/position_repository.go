package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/shift"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PositionRepository implements shift.PositionRepository for PostgreSQL
type PositionRepository struct {
	db *pgxpool.Pool
}

// NewPositionRepository creates a new PositionRepository
func NewPositionRepository(db *pgxpool.Pool) *PositionRepository {
	return &PositionRepository{db: db}
}

// Save saves a position (insert or update)
func (r *PositionRepository) Save(ctx context.Context, position *shift.Position) error {
	query := `
		INSERT INTO positions (
			position_id, tenant_id, position_name, description,
			display_order, is_active, created_at, updated_at, deleted_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (position_id) DO UPDATE SET
			position_name = EXCLUDED.position_name,
			description = EXCLUDED.description,
			display_order = EXCLUDED.display_order,
			is_active = EXCLUDED.is_active,
			updated_at = EXCLUDED.updated_at,
			deleted_at = EXCLUDED.deleted_at
	`

	_, err := r.db.Exec(ctx, query,
		position.PositionID().String(),
		position.TenantID().String(),
		position.PositionName(),
		position.Description(),
		position.DisplayOrder(),
		position.IsActive(),
		position.CreatedAt(),
		position.UpdatedAt(),
		position.DeletedAt(),
	)

	if err != nil {
		return fmt.Errorf("failed to save position: %w", err)
	}

	return nil
}

// FindByID finds a position by ID within a tenant
func (r *PositionRepository) FindByID(ctx context.Context, tenantID common.TenantID, positionID shift.PositionID) (*shift.Position, error) {
	query := `
		SELECT
			position_id, tenant_id, position_name, description,
			display_order, is_active, created_at, updated_at, deleted_at
		FROM positions
		WHERE tenant_id = $1 AND position_id = $2 AND deleted_at IS NULL
	`

	var (
		positionIDStr string
		tenantIDStr   string
		positionName  string
		description   string
		displayOrder  int
		isActive      bool
		createdAt     time.Time
		updatedAt     time.Time
		deletedAt     sql.NullTime
	)

	err := r.db.QueryRow(ctx, query, tenantID.String(), positionID.String()).Scan(
		&positionIDStr,
		&tenantIDStr,
		&positionName,
		&description,
		&displayOrder,
		&isActive,
		&createdAt,
		&updatedAt,
		&deletedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, common.NewNotFoundError("Position", positionID.String())
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find position: %w", err)
	}

	var deletedAtPtr *time.Time
	if deletedAt.Valid {
		deletedAtPtr = &deletedAt.Time
	}

	return shift.ReconstructPosition(
		shift.PositionID(positionIDStr),
		common.TenantID(tenantIDStr),
		positionName,
		description,
		displayOrder,
		isActive,
		createdAt,
		updatedAt,
		deletedAtPtr,
	)
}

// FindByTenantID finds all positions within a tenant
func (r *PositionRepository) FindByTenantID(ctx context.Context, tenantID common.TenantID) ([]*shift.Position, error) {
	query := `
		SELECT
			position_id, tenant_id, position_name, description,
			display_order, is_active, created_at, updated_at, deleted_at
		FROM positions
		WHERE tenant_id = $1 AND deleted_at IS NULL
		ORDER BY display_order ASC, position_name ASC
	`

	return r.queryPositions(ctx, query, tenantID.String())
}

// FindActiveByTenantID finds all active positions within a tenant
func (r *PositionRepository) FindActiveByTenantID(ctx context.Context, tenantID common.TenantID) ([]*shift.Position, error) {
	query := `
		SELECT
			position_id, tenant_id, position_name, description,
			display_order, is_active, created_at, updated_at, deleted_at
		FROM positions
		WHERE tenant_id = $1 AND is_active = true AND deleted_at IS NULL
		ORDER BY display_order ASC, position_name ASC
	`

	return r.queryPositions(ctx, query, tenantID.String())
}

// Delete deletes a position (physical delete)
func (r *PositionRepository) Delete(ctx context.Context, tenantID common.TenantID, positionID shift.PositionID) error {
	query := `
		DELETE FROM positions
		WHERE tenant_id = $1 AND position_id = $2
	`

	result, err := r.db.Exec(ctx, query, tenantID.String(), positionID.String())
	if err != nil {
		return fmt.Errorf("failed to delete position: %w", err)
	}

	if result.RowsAffected() == 0 {
		return common.NewNotFoundError("Position", positionID.String())
	}

	return nil
}

// queryPositions executes a query and returns a list of positions
func (r *PositionRepository) queryPositions(ctx context.Context, query string, args ...interface{}) ([]*shift.Position, error) {
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query positions: %w", err)
	}
	defer rows.Close()

	var positions []*shift.Position
	for rows.Next() {
		var (
			positionIDStr string
			tenantIDStr   string
			positionName  string
			description   string
			displayOrder  int
			isActive      bool
			createdAt     time.Time
			updatedAt     time.Time
			deletedAt     sql.NullTime
		)

		err := rows.Scan(
			&positionIDStr,
			&tenantIDStr,
			&positionName,
			&description,
			&displayOrder,
			&isActive,
			&createdAt,
			&updatedAt,
			&deletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan position row: %w", err)
		}

		var deletedAtPtr *time.Time
		if deletedAt.Valid {
			deletedAtPtr = &deletedAt.Time
		}

		position, err := shift.ReconstructPosition(
			shift.PositionID(positionIDStr),
			common.TenantID(tenantIDStr),
			positionName,
			description,
			displayOrder,
			isActive,
			createdAt,
			updatedAt,
			deletedAtPtr,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to reconstruct position: %w", err)
		}

		positions = append(positions, position)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating position rows: %w", err)
	}

	return positions, nil
}

