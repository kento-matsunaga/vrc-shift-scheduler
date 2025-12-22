package db

import (
	"context"
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/role"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// roleGroupRepository implements role.RoleGroupRepository
type roleGroupRepository struct {
	pool *pgxpool.Pool
}

// NewRoleGroupRepository creates a new RoleGroupRepository
func NewRoleGroupRepository(pool *pgxpool.Pool) role.RoleGroupRepository {
	return &roleGroupRepository{pool: pool}
}

// Save saves a group (insert or update)
func (r *roleGroupRepository) Save(ctx context.Context, group *role.RoleGroup) error {
	query := `
		INSERT INTO role_groups (group_id, tenant_id, name, description, color, display_order, created_at, updated_at, deleted_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (group_id) DO UPDATE SET
			name = EXCLUDED.name,
			description = EXCLUDED.description,
			color = EXCLUDED.color,
			display_order = EXCLUDED.display_order,
			updated_at = EXCLUDED.updated_at,
			deleted_at = EXCLUDED.deleted_at
	`
	_, err := r.pool.Exec(ctx, query,
		group.GroupID().String(),
		group.TenantID().String(),
		group.Name(),
		group.Description(),
		group.Color(),
		group.DisplayOrder(),
		group.CreatedAt(),
		group.UpdatedAt(),
		group.DeletedAt(),
	)
	return err
}

// FindByID finds a group by ID within a tenant
func (r *roleGroupRepository) FindByID(ctx context.Context, tenantID common.TenantID, groupID common.RoleGroupID) (*role.RoleGroup, error) {
	query := `
		SELECT group_id, tenant_id, name, description, color, display_order, created_at, updated_at, deleted_at
		FROM role_groups
		WHERE group_id = $1 AND tenant_id = $2 AND deleted_at IS NULL
	`
	row := r.pool.QueryRow(ctx, query, groupID.String(), tenantID.String())

	var (
		id           string
		tid          string
		name         string
		description  string
		color        string
		displayOrder int
		createdAt    time.Time
		updatedAt    time.Time
		deletedAt    *time.Time
	)

	err := row.Scan(&id, &tid, &name, &description, &color, &displayOrder, &createdAt, &updatedAt, &deletedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	// Load role IDs
	roleIDs, err := r.FindRoleIDsByGroupID(ctx, groupID)
	if err != nil {
		return nil, err
	}

	return role.ReconstructRoleGroup(
		common.RoleGroupID(id),
		common.TenantID(tid),
		name,
		description,
		color,
		displayOrder,
		createdAt,
		updatedAt,
		deletedAt,
		roleIDs,
	), nil
}

// FindByTenantID finds all groups within a tenant
func (r *roleGroupRepository) FindByTenantID(ctx context.Context, tenantID common.TenantID) ([]*role.RoleGroup, error) {
	query := `
		SELECT group_id, tenant_id, name, description, color, display_order, created_at, updated_at, deleted_at
		FROM role_groups
		WHERE tenant_id = $1 AND deleted_at IS NULL
		ORDER BY display_order, name
	`
	rows, err := r.pool.Query(ctx, query, tenantID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var groups []*role.RoleGroup
	for rows.Next() {
		var (
			id           string
			tid          string
			name         string
			description  string
			color        string
			displayOrder int
			createdAt    time.Time
			updatedAt    time.Time
			deletedAt    *time.Time
		)

		err := rows.Scan(&id, &tid, &name, &description, &color, &displayOrder, &createdAt, &updatedAt, &deletedAt)
		if err != nil {
			return nil, err
		}

		// Load role IDs for this group
		roleIDs, err := r.FindRoleIDsByGroupID(ctx, common.RoleGroupID(id))
		if err != nil {
			return nil, err
		}

		group := role.ReconstructRoleGroup(
			common.RoleGroupID(id),
			common.TenantID(tid),
			name,
			description,
			color,
			displayOrder,
			createdAt,
			updatedAt,
			deletedAt,
			roleIDs,
		)
		groups = append(groups, group)
	}

	return groups, rows.Err()
}

// Delete soft deletes a group
func (r *roleGroupRepository) Delete(ctx context.Context, tenantID common.TenantID, groupID common.RoleGroupID) error {
	query := `
		UPDATE role_groups
		SET deleted_at = NOW(), updated_at = NOW()
		WHERE group_id = $1 AND tenant_id = $2 AND deleted_at IS NULL
	`
	_, err := r.pool.Exec(ctx, query, groupID.String(), tenantID.String())
	return err
}

// AssignRole assigns a role to a group
func (r *roleGroupRepository) AssignRole(ctx context.Context, groupID common.RoleGroupID, roleID common.RoleID) error {
	query := `
		INSERT INTO role_group_assignments (assignment_id, role_id, group_id, created_at)
		VALUES ($1, $2, $3, NOW())
		ON CONFLICT (role_id, group_id) DO NOTHING
	`
	_, err := r.pool.Exec(ctx, query, common.NewULID(), roleID.String(), groupID.String())
	return err
}

// RemoveRole removes a role from a group
func (r *roleGroupRepository) RemoveRole(ctx context.Context, groupID common.RoleGroupID, roleID common.RoleID) error {
	query := `DELETE FROM role_group_assignments WHERE group_id = $1 AND role_id = $2`
	_, err := r.pool.Exec(ctx, query, groupID.String(), roleID.String())
	return err
}

// FindRoleIDsByGroupID finds all role IDs in a group
func (r *roleGroupRepository) FindRoleIDsByGroupID(ctx context.Context, groupID common.RoleGroupID) ([]common.RoleID, error) {
	query := `
		SELECT role_id FROM role_group_assignments WHERE group_id = $1
	`
	rows, err := r.pool.Query(ctx, query, groupID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roleIDs []common.RoleID
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		roleIDs = append(roleIDs, common.RoleID(id))
	}

	return roleIDs, rows.Err()
}

// FindGroupIDsByRoleID finds all group IDs for a role
func (r *roleGroupRepository) FindGroupIDsByRoleID(ctx context.Context, roleID common.RoleID) ([]common.RoleGroupID, error) {
	query := `
		SELECT group_id FROM role_group_assignments WHERE role_id = $1
	`
	rows, err := r.pool.Query(ctx, query, roleID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var groupIDs []common.RoleGroupID
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		groupIDs = append(groupIDs, common.RoleGroupID(id))
	}

	return groupIDs, rows.Err()
}

// SetGroupRoles replaces all roles in a group
func (r *roleGroupRepository) SetGroupRoles(ctx context.Context, groupID common.RoleGroupID, roleIDs []common.RoleID) error {
	// Start transaction
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Delete existing assignments
	_, err = tx.Exec(ctx, `DELETE FROM role_group_assignments WHERE group_id = $1`, groupID.String())
	if err != nil {
		return err
	}

	// Insert new assignments
	for _, roleID := range roleIDs {
		_, err = tx.Exec(ctx,
			`INSERT INTO role_group_assignments (assignment_id, role_id, group_id, created_at) VALUES ($1, $2, $3, NOW())`,
			common.NewULID(), roleID.String(), groupID.String(),
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}
