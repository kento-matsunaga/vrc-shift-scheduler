package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/auth"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// AdminRepository implements auth.AdminRepository for PostgreSQL
type AdminRepository struct {
	db *pgxpool.Pool
}

// NewAdminRepository creates a new AdminRepository
func NewAdminRepository(db *pgxpool.Pool) *AdminRepository {
	return &AdminRepository{db: db}
}

// Save saves an admin (insert or update)
func (r *AdminRepository) Save(ctx context.Context, a *auth.Admin) error {
	query := `
		INSERT INTO admins (
			admin_id, tenant_id, email, password_hash, display_name, role,
			is_active, created_at, updated_at, deleted_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (admin_id) DO UPDATE SET
			email = EXCLUDED.email,
			password_hash = EXCLUDED.password_hash,
			display_name = EXCLUDED.display_name,
			role = EXCLUDED.role,
			is_active = EXCLUDED.is_active,
			updated_at = EXCLUDED.updated_at,
			deleted_at = EXCLUDED.deleted_at
	`

	_, err := r.db.Exec(ctx, query,
		a.AdminID().String(),
		a.TenantID().String(),
		a.Email(),
		a.PasswordHash(),
		a.DisplayName(),
		a.Role().String(),
		a.IsActive(),
		a.CreatedAt(),
		a.UpdatedAt(),
		a.DeletedAt(),
	)

	if err != nil {
		return fmt.Errorf("failed to save admin: %w", err)
	}

	return nil
}

// FindByIDWithTenant finds an admin by ID within a tenant (backward compatible)
func (r *AdminRepository) FindByIDWithTenant(ctx context.Context, tenantID common.TenantID, adminID common.AdminID) (*auth.Admin, error) {
	query := `
		SELECT
			admin_id, tenant_id, email, password_hash, display_name, role,
			is_active, created_at, updated_at, deleted_at
		FROM admins
		WHERE tenant_id = $1 AND admin_id = $2 AND deleted_at IS NULL
	`

	var (
		adminIDStr    string
		tenantIDStr   string
		email         string
		passwordHash  string
		displayName   string
		roleStr       string
		isActive      bool
		createdAt     time.Time
		updatedAt     time.Time
		deletedAt     sql.NullTime
	)

	err := r.db.QueryRow(ctx, query, tenantID.String(), adminID.String()).Scan(
		&adminIDStr,
		&tenantIDStr,
		&email,
		&passwordHash,
		&displayName,
		&roleStr,
		&isActive,
		&createdAt,
		&updatedAt,
		&deletedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, common.NewNotFoundError("Admin", adminID.String())
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find admin: %w", err)
	}

	parsedAdminID, err := common.ParseAdminID(adminIDStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse admin_id: %w", err)
	}

	parsedTenantID, err := common.ParseTenantID(tenantIDStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse tenant_id: %w", err)
	}

	role, err := auth.NewRole(roleStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse role: %w", err)
	}

	var deletedAtPtr *time.Time
	if deletedAt.Valid {
		deletedAtPtr = &deletedAt.Time
	}

	return auth.ReconstructAdmin(
		parsedAdminID,
		parsedTenantID,
		email,
		passwordHash,
		displayName,
		role,
		isActive,
		createdAt,
		updatedAt,
		deletedAtPtr,
	)
}

// FindByID finds an admin by ID (global search, no tenant filtering)
func (r *AdminRepository) FindByID(ctx context.Context, adminID common.AdminID) (*auth.Admin, error) {
	query := `
		SELECT
			admin_id, tenant_id, email, password_hash, display_name, role,
			is_active, created_at, updated_at, deleted_at
		FROM admins
		WHERE admin_id = $1 AND deleted_at IS NULL
		LIMIT 1
	`

	var (
		adminIDStr    string
		tenantIDStr   string
		email         string
		passwordHash  string
		displayName   string
		roleStr       string
		isActive      bool
		createdAt     time.Time
		updatedAt     time.Time
		deletedAt     sql.NullTime
	)

	err := r.db.QueryRow(ctx, query, adminID.String()).Scan(
		&adminIDStr,
		&tenantIDStr,
		&email,
		&passwordHash,
		&displayName,
		&roleStr,
		&isActive,
		&createdAt,
		&updatedAt,
		&deletedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, common.NewNotFoundError("Admin", adminID.String())
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find admin: %w", err)
	}

	parsedAdminID, err := common.ParseAdminID(adminIDStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse admin_id: %w", err)
	}

	parsedTenantID, err := common.ParseTenantID(tenantIDStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse tenant_id: %w", err)
	}

	role, err := auth.NewRole(roleStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse role: %w", err)
	}

	var deletedAtPtr *time.Time
	if deletedAt.Valid {
		deletedAtPtr = &deletedAt.Time
	}

	return auth.ReconstructAdmin(
		parsedAdminID,
		parsedTenantID,
		email,
		passwordHash,
		displayName,
		role,
		isActive,
		createdAt,
		updatedAt,
		deletedAtPtr,
	)
}

// FindByEmail finds an admin by email within a tenant
func (r *AdminRepository) FindByEmail(ctx context.Context, tenantID common.TenantID, email string) (*auth.Admin, error) {
	query := `
		SELECT
			admin_id, tenant_id, email, password_hash, display_name, role,
			is_active, created_at, updated_at, deleted_at
		FROM admins
		WHERE tenant_id = $1 AND email = $2 AND deleted_at IS NULL
	`

	var (
		adminIDStr    string
		tenantIDStr   string
		emailStr      string
		passwordHash  string
		displayName   string
		roleStr       string
		isActive      bool
		createdAt     time.Time
		updatedAt     time.Time
		deletedAt     sql.NullTime
	)

	err := r.db.QueryRow(ctx, query, tenantID.String(), email).Scan(
		&adminIDStr,
		&tenantIDStr,
		&emailStr,
		&passwordHash,
		&displayName,
		&roleStr,
		&isActive,
		&createdAt,
		&updatedAt,
		&deletedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, common.NewNotFoundError("Admin", email)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find admin by email: %w", err)
	}

	parsedAdminID, err := common.ParseAdminID(adminIDStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse admin_id: %w", err)
	}

	parsedTenantID, err := common.ParseTenantID(tenantIDStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse tenant_id: %w", err)
	}

	role, err := auth.NewRole(roleStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse role: %w", err)
	}

	var deletedAtPtr *time.Time
	if deletedAt.Valid {
		deletedAtPtr = &deletedAt.Time
	}

	return auth.ReconstructAdmin(
		parsedAdminID,
		parsedTenantID,
		emailStr,
		passwordHash,
		displayName,
		role,
		isActive,
		createdAt,
		updatedAt,
		deletedAtPtr,
	)
}

// FindByEmailGlobal finds an admin by email (global search)
func (r *AdminRepository) FindByEmailGlobal(ctx context.Context, email string) (*auth.Admin, error) {
	query := `
		SELECT
			admin_id, tenant_id, email, password_hash, display_name, role,
			is_active, created_at, updated_at, deleted_at
		FROM admins
		WHERE email = $1 AND deleted_at IS NULL
		LIMIT 1
	`

	var (
		adminIDStr    string
		tenantIDStr   string
		emailStr      string
		passwordHash  string
		displayName   string
		roleStr       string
		isActive      bool
		createdAt     time.Time
		updatedAt     time.Time
		deletedAt     sql.NullTime
	)

	err := r.db.QueryRow(ctx, query, email).Scan(
		&adminIDStr,
		&tenantIDStr,
		&emailStr,
		&passwordHash,
		&displayName,
		&roleStr,
		&isActive,
		&createdAt,
		&updatedAt,
		&deletedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, common.NewNotFoundError("Admin", email)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find admin by email: %w", err)
	}

	parsedAdminID, err := common.ParseAdminID(adminIDStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse admin_id: %w", err)
	}

	parsedTenantID, err := common.ParseTenantID(tenantIDStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse tenant_id: %w", err)
	}

	role, err := auth.NewRole(roleStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse role: %w", err)
	}

	var deletedAtPtr *time.Time
	if deletedAt.Valid {
		deletedAtPtr = &deletedAt.Time
	}

	return auth.ReconstructAdmin(
		parsedAdminID,
		parsedTenantID,
		emailStr,
		passwordHash,
		displayName,
		role,
		isActive,
		createdAt,
		updatedAt,
		deletedAtPtr,
	)
}

// FindByTenantID finds all admins within a tenant
func (r *AdminRepository) FindByTenantID(ctx context.Context, tenantID common.TenantID) ([]*auth.Admin, error) {
	query := `
		SELECT
			admin_id, tenant_id, email, password_hash, display_name, role,
			is_active, created_at, updated_at, deleted_at
		FROM admins
		WHERE tenant_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query, tenantID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to find admins by tenant: %w", err)
	}
	defer rows.Close()

	var admins []*auth.Admin
	for rows.Next() {
		var (
			adminIDStr    string
			tenantIDStr   string
			email         string
			passwordHash  string
			displayName   string
			roleStr       string
			isActive      bool
			createdAt     time.Time
			updatedAt     time.Time
			deletedAt     sql.NullTime
		)

		err := rows.Scan(
			&adminIDStr,
			&tenantIDStr,
			&email,
			&passwordHash,
			&displayName,
			&roleStr,
			&isActive,
			&createdAt,
			&updatedAt,
			&deletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan admin: %w", err)
		}

		parsedAdminID, err := common.ParseAdminID(adminIDStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse admin_id: %w", err)
		}

		parsedTenantID, err := common.ParseTenantID(tenantIDStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse tenant_id: %w", err)
		}

		role, err := auth.NewRole(roleStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse role: %w", err)
		}

		var deletedAtPtr *time.Time
		if deletedAt.Valid {
			deletedAtPtr = &deletedAt.Time
		}

		admin, err := auth.ReconstructAdmin(
			parsedAdminID,
			parsedTenantID,
			email,
			passwordHash,
			displayName,
			role,
			isActive,
			createdAt,
			updatedAt,
			deletedAtPtr,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to reconstruct admin: %w", err)
		}

		admins = append(admins, admin)
	}

	return admins, nil
}

// FindActiveByTenantID finds all active admins within a tenant
func (r *AdminRepository) FindActiveByTenantID(ctx context.Context, tenantID common.TenantID) ([]*auth.Admin, error) {
	query := `
		SELECT
			admin_id, tenant_id, email, password_hash, display_name, role,
			is_active, created_at, updated_at, deleted_at
		FROM admins
		WHERE tenant_id = $1 AND is_active = true AND deleted_at IS NULL
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query, tenantID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to find active admins: %w", err)
	}
	defer rows.Close()

	var admins []*auth.Admin
	for rows.Next() {
		var (
			adminIDStr    string
			tenantIDStr   string
			email         string
			passwordHash  string
			displayName   string
			roleStr       string
			isActive      bool
			createdAt     time.Time
			updatedAt     time.Time
			deletedAt     sql.NullTime
		)

		err := rows.Scan(
			&adminIDStr,
			&tenantIDStr,
			&email,
			&passwordHash,
			&displayName,
			&roleStr,
			&isActive,
			&createdAt,
			&updatedAt,
			&deletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan admin: %w", err)
		}

		parsedAdminID, err := common.ParseAdminID(adminIDStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse admin_id: %w", err)
		}

		parsedTenantID, err := common.ParseTenantID(tenantIDStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse tenant_id: %w", err)
		}

		role, err := auth.NewRole(roleStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse role: %w", err)
		}

		var deletedAtPtr *time.Time
		if deletedAt.Valid {
			deletedAtPtr = &deletedAt.Time
		}

		admin, err := auth.ReconstructAdmin(
			parsedAdminID,
			parsedTenantID,
			email,
			passwordHash,
			displayName,
			role,
			isActive,
			createdAt,
			updatedAt,
			deletedAtPtr,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to reconstruct admin: %w", err)
		}

		admins = append(admins, admin)
	}

	return admins, nil
}

// Delete deletes an admin (physical delete)
func (r *AdminRepository) Delete(ctx context.Context, tenantID common.TenantID, adminID common.AdminID) error {
	query := `
		DELETE FROM admins
		WHERE tenant_id = $1 AND admin_id = $2
	`

	result, err := r.db.Exec(ctx, query, tenantID.String(), adminID.String())
	if err != nil {
		return fmt.Errorf("failed to delete admin: %w", err)
	}

	if result.RowsAffected() == 0 {
		return common.NewNotFoundError("Admin", adminID.String())
	}

	return nil
}

// ExistsByEmail checks if an admin with the given email exists within a tenant
func (r *AdminRepository) ExistsByEmail(ctx context.Context, tenantID common.TenantID, email string) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM admins
			WHERE tenant_id = $1 AND email = $2 AND deleted_at IS NULL
		)
	`

	var exists bool
	err := r.db.QueryRow(ctx, query, tenantID.String(), email).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check admin existence: %w", err)
	}

	return exists, nil
}
