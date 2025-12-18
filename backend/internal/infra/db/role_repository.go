package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/role"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// RoleRepository implements role.RoleRepository for PostgreSQL
type RoleRepository struct {
	db *pgxpool.Pool
}

// NewRoleRepository creates a new RoleRepository
func NewRoleRepository(db *pgxpool.Pool) *RoleRepository {
	return &RoleRepository{db: db}
}

// Save saves a role (insert or update)
func (r *RoleRepository) Save(ctx context.Context, roleEntity *role.Role) error {
	query := `
		INSERT INTO roles (
			role_id, tenant_id, name, description, color,
			display_order, created_at, updated_at, deleted_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (role_id) DO UPDATE SET
			name = EXCLUDED.name,
			description = EXCLUDED.description,
			color = EXCLUDED.color,
			display_order = EXCLUDED.display_order,
			updated_at = EXCLUDED.updated_at,
			deleted_at = EXCLUDED.deleted_at
	`

	_, err := r.db.Exec(ctx, query,
		roleEntity.RoleID().String(),
		roleEntity.TenantID().String(),
		roleEntity.Name(),
		nullString(roleEntity.Description()),
		nullString(roleEntity.Color()),
		roleEntity.DisplayOrder(),
		roleEntity.CreatedAt(),
		roleEntity.UpdatedAt(),
		roleEntity.DeletedAt(),
	)

	if err != nil {
		return fmt.Errorf("failed to save role: %w", err)
	}

	return nil
}

// FindByID finds a role by ID within a tenant
func (r *RoleRepository) FindByID(ctx context.Context, tenantID common.TenantID, roleID common.RoleID) (*role.Role, error) {
	query := `
		SELECT
			role_id, tenant_id, name, description, color,
			display_order, created_at, updated_at, deleted_at
		FROM roles
		WHERE tenant_id = $1 AND role_id = $2 AND deleted_at IS NULL
	`

	var (
		roleIDStr    string
		tenantIDStr  string
		name         string
		description  sql.NullString
		color        sql.NullString
		displayOrder int
		createdAt    time.Time
		updatedAt    time.Time
		deletedAt    sql.NullTime
	)

	err := r.db.QueryRow(ctx, query, tenantID.String(), roleID.String()).Scan(
		&roleIDStr,
		&tenantIDStr,
		&name,
		&description,
		&color,
		&displayOrder,
		&createdAt,
		&updatedAt,
		&deletedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, common.NewNotFoundError("Role", roleID.String())
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find role: %w", err)
	}

	return r.scanToRole(
		roleIDStr, tenantIDStr, name, description, color,
		displayOrder, createdAt, updatedAt, deletedAt,
	)
}

// FindByTenantID finds all roles within a tenant (deleted_at IS NULL)
func (r *RoleRepository) FindByTenantID(ctx context.Context, tenantID common.TenantID) ([]*role.Role, error) {
	query := `
		SELECT
			role_id, tenant_id, name, description, color,
			display_order, created_at, updated_at, deleted_at
		FROM roles
		WHERE tenant_id = $1 AND deleted_at IS NULL
		ORDER BY display_order ASC, created_at DESC
	`

	return r.queryRoles(ctx, query, tenantID.String())
}

// Delete deletes a role (soft delete)
func (r *RoleRepository) Delete(ctx context.Context, tenantID common.TenantID, roleID common.RoleID) error {
	// ロールエンティティを取得
	roleEntity, err := r.FindByID(ctx, tenantID, roleID)
	if err != nil {
		return err
	}

	// ソフトデリート実行
	roleEntity.Delete()

	// 保存
	return r.Save(ctx, roleEntity)
}

// queryRoles executes a query and returns a list of roles
func (r *RoleRepository) queryRoles(ctx context.Context, query string, args ...interface{}) ([]*role.Role, error) {
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query roles: %w", err)
	}
	defer rows.Close()

	var roles []*role.Role
	for rows.Next() {
		var (
			roleIDStr    string
			tenantIDStr  string
			name         string
			description  sql.NullString
			color        sql.NullString
			displayOrder int
			createdAt    time.Time
			updatedAt    time.Time
			deletedAt    sql.NullTime
		)

		err := rows.Scan(
			&roleIDStr,
			&tenantIDStr,
			&name,
			&description,
			&color,
			&displayOrder,
			&createdAt,
			&updatedAt,
			&deletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan role row: %w", err)
		}

		roleEntity, err := r.scanToRole(
			roleIDStr, tenantIDStr, name, description, color,
			displayOrder, createdAt, updatedAt, deletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to reconstruct role: %w", err)
		}

		roles = append(roles, roleEntity)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating role rows: %w", err)
	}

	return roles, nil
}

// scanToRole converts scanned row data to Role entity
func (r *RoleRepository) scanToRole(
	roleIDStr, tenantIDStr, name string,
	description, color sql.NullString,
	displayOrder int,
	createdAt, updatedAt time.Time,
	deletedAt sql.NullTime,
) (*role.Role, error) {
	var deletedAtPtr *time.Time
	if deletedAt.Valid {
		deletedAtPtr = &deletedAt.Time
	}

	return role.ReconstructRole(
		common.RoleID(roleIDStr),
		common.TenantID(tenantIDStr),
		name,
		stringValue(description),
		stringValue(color),
		displayOrder,
		createdAt,
		updatedAt,
		deletedAtPtr,
	)
}
