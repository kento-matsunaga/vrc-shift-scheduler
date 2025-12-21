package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/billing"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// EntitlementRepository implements billing.EntitlementRepository for PostgreSQL
type EntitlementRepository struct {
	db *pgxpool.Pool
}

// NewEntitlementRepository creates a new EntitlementRepository
func NewEntitlementRepository(db *pgxpool.Pool) *EntitlementRepository {
	return &EntitlementRepository{db: db}
}

// Save saves an entitlement
func (r *EntitlementRepository) Save(ctx context.Context, e *billing.Entitlement) error {
	query := `
		INSERT INTO entitlements (
			entitlement_id, tenant_id, plan_code, source, starts_at, ends_at,
			revoked_at, revoked_reason, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (entitlement_id) DO UPDATE SET
			revoked_at = EXCLUDED.revoked_at,
			revoked_reason = EXCLUDED.revoked_reason,
			ends_at = EXCLUDED.ends_at,
			updated_at = EXCLUDED.updated_at
	`

	_, err := r.db.Exec(ctx, query,
		e.EntitlementID().String(),
		e.TenantID().String(),
		e.PlanCode(),
		e.Source().String(),
		e.StartsAt(),
		e.EndsAt(),
		e.RevokedAt(),
		e.RevokedReason(),
		e.CreatedAt(),
		e.UpdatedAt(),
	)

	if err != nil {
		return fmt.Errorf("failed to save entitlement: %w", err)
	}

	return nil
}

// FindByID finds an entitlement by ID
func (r *EntitlementRepository) FindByID(ctx context.Context, entitlementID billing.EntitlementID) (*billing.Entitlement, error) {
	query := `
		SELECT
			entitlement_id, tenant_id, plan_code, source, starts_at, ends_at,
			revoked_at, revoked_reason, created_at, updated_at
		FROM entitlements
		WHERE entitlement_id = $1
	`

	return r.scanEntitlement(ctx, query, entitlementID.String())
}

// FindByTenantID finds all entitlements for a tenant
func (r *EntitlementRepository) FindByTenantID(ctx context.Context, tenantID common.TenantID) ([]*billing.Entitlement, error) {
	query := `
		SELECT
			entitlement_id, tenant_id, plan_code, source, starts_at, ends_at,
			revoked_at, revoked_reason, created_at, updated_at
		FROM entitlements
		WHERE tenant_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query, tenantID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to query entitlements: %w", err)
	}
	defer rows.Close()

	return r.scanEntitlements(rows)
}

// FindActiveByTenantID finds the active entitlement for a tenant (prioritized)
// Priority: 1. revoked -> skip, 2. lifetime (ends_at IS NULL), 3. latest ends_at
func (r *EntitlementRepository) FindActiveByTenantID(ctx context.Context, tenantID common.TenantID) (*billing.Entitlement, error) {
	query := `
		SELECT
			entitlement_id, tenant_id, plan_code, source, starts_at, ends_at,
			revoked_at, revoked_reason, created_at, updated_at
		FROM entitlements
		WHERE tenant_id = $1
			AND revoked_at IS NULL
			AND (ends_at IS NULL OR ends_at > NOW())
			AND starts_at <= NOW()
		ORDER BY
			CASE WHEN ends_at IS NULL THEN 0 ELSE 1 END,
			ends_at DESC NULLS FIRST
		LIMIT 1
	`

	return r.scanEntitlement(ctx, query, tenantID.String())
}

// HasRevokedByTenantID checks if any entitlement for the tenant is revoked
func (r *EntitlementRepository) HasRevokedByTenantID(ctx context.Context, tenantID common.TenantID) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM entitlements
			WHERE tenant_id = $1 AND revoked_at IS NOT NULL
		)
	`

	var hasRevoked bool
	err := r.db.QueryRow(ctx, query, tenantID.String()).Scan(&hasRevoked)
	if err != nil {
		return false, fmt.Errorf("failed to check revoked entitlements: %w", err)
	}

	return hasRevoked, nil
}

func (r *EntitlementRepository) scanEntitlement(ctx context.Context, query string, args ...interface{}) (*billing.Entitlement, error) {
	var (
		entitlementIDStr string
		tenantIDStr      string
		planCode         string
		source           string
		startsAt         time.Time
		endsAt           sql.NullTime
		revokedAt        sql.NullTime
		revokedReason    sql.NullString
		createdAt        time.Time
		updatedAt        time.Time
	)

	err := r.db.QueryRow(ctx, query, args...).Scan(
		&entitlementIDStr,
		&tenantIDStr,
		&planCode,
		&source,
		&startsAt,
		&endsAt,
		&revokedAt,
		&revokedReason,
		&createdAt,
		&updatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find entitlement: %w", err)
	}

	entitlementID, err := billing.ParseEntitlementID(entitlementIDStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse entitlement_id: %w", err)
	}

	tenantID, err := common.ParseTenantID(tenantIDStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse tenant_id: %w", err)
	}

	var endsAtPtr *time.Time
	if endsAt.Valid {
		endsAtPtr = &endsAt.Time
	}

	var revokedAtPtr *time.Time
	if revokedAt.Valid {
		revokedAtPtr = &revokedAt.Time
	}

	var revokedReasonPtr *string
	if revokedReason.Valid {
		revokedReasonPtr = &revokedReason.String
	}

	return billing.ReconstructEntitlement(
		entitlementID,
		tenantID,
		planCode,
		billing.EntitlementSource(source),
		startsAt,
		endsAtPtr,
		revokedAtPtr,
		revokedReasonPtr,
		createdAt,
		updatedAt,
	)
}

func (r *EntitlementRepository) scanEntitlements(rows pgx.Rows) ([]*billing.Entitlement, error) {
	var entitlements []*billing.Entitlement

	for rows.Next() {
		var (
			entitlementIDStr string
			tenantIDStr      string
			planCode         string
			source           string
			startsAt         time.Time
			endsAt           sql.NullTime
			revokedAt        sql.NullTime
			revokedReason    sql.NullString
			createdAt        time.Time
			updatedAt        time.Time
		)

		if err := rows.Scan(
			&entitlementIDStr,
			&tenantIDStr,
			&planCode,
			&source,
			&startsAt,
			&endsAt,
			&revokedAt,
			&revokedReason,
			&createdAt,
			&updatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan entitlement: %w", err)
		}

		entitlementID, err := billing.ParseEntitlementID(entitlementIDStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse entitlement_id: %w", err)
		}

		tenantID, err := common.ParseTenantID(tenantIDStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse tenant_id: %w", err)
		}

		var endsAtPtr *time.Time
		if endsAt.Valid {
			endsAtPtr = &endsAt.Time
		}

		var revokedAtPtr *time.Time
		if revokedAt.Valid {
			revokedAtPtr = &revokedAt.Time
		}

		var revokedReasonPtr *string
		if revokedReason.Valid {
			revokedReasonPtr = &revokedReason.String
		}

		e, err := billing.ReconstructEntitlement(
			entitlementID,
			tenantID,
			planCode,
			billing.EntitlementSource(source),
			startsAt,
			endsAtPtr,
			revokedAtPtr,
			revokedReasonPtr,
			createdAt,
			updatedAt,
		)
		if err != nil {
			return nil, err
		}
		entitlements = append(entitlements, e)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating entitlements: %w", err)
	}

	return entitlements, nil
}
