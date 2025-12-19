package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/tenant"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// TenantRepository implements tenant.TenantRepository for PostgreSQL
type TenantRepository struct {
	db *pgxpool.Pool
}

// NewTenantRepository creates a new TenantRepository
func NewTenantRepository(db *pgxpool.Pool) *TenantRepository {
	return &TenantRepository{db: db}
}

// FindByID finds a tenant by ID
func (r *TenantRepository) FindByID(ctx context.Context, tenantID common.TenantID) (*tenant.Tenant, error) {
	query := `
		SELECT
			tenant_id, tenant_name, timezone, is_active,
			created_at, updated_at, deleted_at
		FROM tenants
		WHERE tenant_id = $1 AND deleted_at IS NULL
	`

	var (
		tenantIDStr string
		tenantName  string
		timezone    string
		isActive    bool
		createdAt   time.Time
		updatedAt   time.Time
		deletedAt   sql.NullTime
	)

	err := r.db.QueryRow(ctx, query, tenantID.String()).Scan(
		&tenantIDStr,
		&tenantName,
		&timezone,
		&isActive,
		&createdAt,
		&updatedAt,
		&deletedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, common.NewNotFoundError("Tenant", tenantID.String())
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find tenant: %w", err)
	}

	parsedTenantID, err := common.ParseTenantID(tenantIDStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse tenant_id: %w", err)
	}

	var deletedAtPtr *time.Time
	if deletedAt.Valid {
		deletedAtPtr = &deletedAt.Time
	}

	return tenant.ReconstructTenant(
		parsedTenantID,
		tenantName,
		timezone,
		isActive,
		createdAt,
		updatedAt,
		deletedAtPtr,
	)
}

// Save saves a tenant (insert or update)
func (r *TenantRepository) Save(ctx context.Context, t *tenant.Tenant) error {
	query := `
		INSERT INTO tenants (
			tenant_id, tenant_name, timezone, is_active,
			created_at, updated_at, deleted_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (tenant_id) DO UPDATE SET
			tenant_name = EXCLUDED.tenant_name,
			timezone = EXCLUDED.timezone,
			is_active = EXCLUDED.is_active,
			updated_at = EXCLUDED.updated_at,
			deleted_at = EXCLUDED.deleted_at
	`

	_, err := r.db.Exec(ctx, query,
		t.TenantID().String(),
		t.TenantName(),
		t.Timezone(),
		t.IsActive(),
		t.CreatedAt(),
		t.UpdatedAt(),
		t.DeletedAt(),
	)

	if err != nil {
		return fmt.Errorf("failed to save tenant: %w", err)
	}

	return nil
}
