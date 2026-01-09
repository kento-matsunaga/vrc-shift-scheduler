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

// InstanceRepository implements shift.InstanceRepository for PostgreSQL
type InstanceRepository struct {
	db *pgxpool.Pool
}

// NewInstanceRepository creates a new InstanceRepository
func NewInstanceRepository(db *pgxpool.Pool) *InstanceRepository {
	return &InstanceRepository{db: db}
}

// Save saves an instance (insert or update)
func (r *InstanceRepository) Save(ctx context.Context, instance *shift.Instance) error {
	query := `
		INSERT INTO instances (
			instance_id, tenant_id, event_id,
			name, display_order, max_members,
			created_at, updated_at, deleted_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (instance_id) DO UPDATE SET
			name = EXCLUDED.name,
			display_order = EXCLUDED.display_order,
			max_members = EXCLUDED.max_members,
			updated_at = EXCLUDED.updated_at,
			deleted_at = EXCLUDED.deleted_at
	`

	_, err := r.db.Exec(ctx, query,
		instance.InstanceID().String(),
		instance.TenantID().String(),
		instance.EventID().String(),
		instance.Name(),
		instance.DisplayOrder(),
		instance.MaxMembers(),
		instance.CreatedAt(),
		instance.UpdatedAt(),
		instance.DeletedAt(),
	)

	if err != nil {
		return fmt.Errorf("failed to save instance: %w", err)
	}

	return nil
}

// FindByID finds an instance by ID within a tenant
func (r *InstanceRepository) FindByID(ctx context.Context, tenantID common.TenantID, instanceID shift.InstanceID) (*shift.Instance, error) {
	query := `
		SELECT
			instance_id, tenant_id, event_id,
			name, display_order, max_members,
			created_at, updated_at, deleted_at
		FROM instances
		WHERE tenant_id = $1 AND instance_id = $2 AND deleted_at IS NULL
	`

	var (
		instanceIDStr string
		tenantIDStr   string
		eventIDStr    string
		name          string
		displayOrder  int
		maxMembers    sql.NullInt32
		createdAt     time.Time
		updatedAt     time.Time
		deletedAt     sql.NullTime
	)

	err := r.db.QueryRow(ctx, query, tenantID.String(), instanceID.String()).Scan(
		&instanceIDStr,
		&tenantIDStr,
		&eventIDStr,
		&name,
		&displayOrder,
		&maxMembers,
		&createdAt,
		&updatedAt,
		&deletedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, common.NewNotFoundError("Instance", instanceID.String())
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find instance: %w", err)
	}

	return r.scanToInstance(
		instanceIDStr, tenantIDStr, eventIDStr,
		name, displayOrder, maxMembers,
		createdAt, updatedAt, deletedAt,
	)
}

// FindByEventID finds all instances for an event, ordered by display_order
func (r *InstanceRepository) FindByEventID(ctx context.Context, tenantID common.TenantID, eventID common.EventID) ([]*shift.Instance, error) {
	query := `
		SELECT
			instance_id, tenant_id, event_id,
			name, display_order, max_members,
			created_at, updated_at, deleted_at
		FROM instances
		WHERE tenant_id = $1 AND event_id = $2 AND deleted_at IS NULL
		ORDER BY display_order ASC, name ASC
	`

	return r.queryInstances(ctx, query, tenantID.String(), eventID.String())
}

// FindByEventIDAndName finds an instance by event ID and name
func (r *InstanceRepository) FindByEventIDAndName(ctx context.Context, tenantID common.TenantID, eventID common.EventID, name string) (*shift.Instance, error) {
	query := `
		SELECT
			instance_id, tenant_id, event_id,
			name, display_order, max_members,
			created_at, updated_at, deleted_at
		FROM instances
		WHERE tenant_id = $1 AND event_id = $2 AND name = $3 AND deleted_at IS NULL
	`

	var (
		instanceIDStr string
		tenantIDStr   string
		eventIDStr    string
		instanceName  string
		displayOrder  int
		maxMembers    sql.NullInt32
		createdAt     time.Time
		updatedAt     time.Time
		deletedAt     sql.NullTime
	)

	err := r.db.QueryRow(ctx, query, tenantID.String(), eventID.String(), name).Scan(
		&instanceIDStr,
		&tenantIDStr,
		&eventIDStr,
		&instanceName,
		&displayOrder,
		&maxMembers,
		&createdAt,
		&updatedAt,
		&deletedAt,
	)

	if err == pgx.ErrNoRows {
		// Not found is not an error - return nil
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find instance by name: %w", err)
	}

	return r.scanToInstance(
		instanceIDStr, tenantIDStr, eventIDStr,
		instanceName, displayOrder, maxMembers,
		createdAt, updatedAt, deletedAt,
	)
}

// Delete deletes an instance (physical delete)
func (r *InstanceRepository) Delete(ctx context.Context, tenantID common.TenantID, instanceID shift.InstanceID) error {
	query := `
		DELETE FROM instances
		WHERE tenant_id = $1 AND instance_id = $2
	`

	result, err := r.db.Exec(ctx, query, tenantID.String(), instanceID.String())
	if err != nil {
		return fmt.Errorf("failed to delete instance: %w", err)
	}

	if result.RowsAffected() == 0 {
		return common.NewNotFoundError("Instance", instanceID.String())
	}

	return nil
}

// queryInstances executes a query and returns a list of instances
func (r *InstanceRepository) queryInstances(ctx context.Context, query string, args ...interface{}) ([]*shift.Instance, error) {
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query instances: %w", err)
	}
	defer rows.Close()

	var instances []*shift.Instance
	for rows.Next() {
		var (
			instanceIDStr string
			tenantIDStr   string
			eventIDStr    string
			name          string
			displayOrder  int
			maxMembers    sql.NullInt32
			createdAt     time.Time
			updatedAt     time.Time
			deletedAt     sql.NullTime
		)

		err := rows.Scan(
			&instanceIDStr,
			&tenantIDStr,
			&eventIDStr,
			&name,
			&displayOrder,
			&maxMembers,
			&createdAt,
			&updatedAt,
			&deletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan instance row: %w", err)
		}

		instance, err := r.scanToInstance(
			instanceIDStr, tenantIDStr, eventIDStr,
			name, displayOrder, maxMembers,
			createdAt, updatedAt, deletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to reconstruct instance: %w", err)
		}

		instances = append(instances, instance)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating instance rows: %w", err)
	}

	return instances, nil
}

// scanToInstance converts scanned row data to Instance entity
func (r *InstanceRepository) scanToInstance(
	instanceIDStr, tenantIDStr, eventIDStr string,
	name string,
	displayOrder int,
	maxMembers sql.NullInt32,
	createdAt, updatedAt time.Time,
	deletedAt sql.NullTime,
) (*shift.Instance, error) {
	var deletedAtPtr *time.Time
	if deletedAt.Valid {
		deletedAtPtr = &deletedAt.Time
	}

	var maxMembersPtr *int
	if maxMembers.Valid {
		val := int(maxMembers.Int32)
		maxMembersPtr = &val
	}

	return shift.ReconstructInstance(
		shift.InstanceID(instanceIDStr),
		common.TenantID(tenantIDStr),
		common.EventID(eventIDStr),
		name,
		displayOrder,
		maxMembersPtr,
		createdAt,
		updatedAt,
		deletedAtPtr,
	)
}
