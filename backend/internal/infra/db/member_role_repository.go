package db

import (
	"context"
	"fmt"
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/jackc/pgx/v5/pgxpool"
)

// MemberRoleRepository manages member-role associations
type MemberRoleRepository struct {
	db *pgxpool.Pool
}

// NewMemberRoleRepository creates a new MemberRoleRepository
func NewMemberRoleRepository(db *pgxpool.Pool) *MemberRoleRepository {
	return &MemberRoleRepository{db: db}
}

// AssignRole assigns a role to a member
func (r *MemberRoleRepository) AssignRole(ctx context.Context, memberID common.MemberID, roleID common.RoleID) error {
	query := `
		INSERT INTO member_roles (member_id, role_id, assigned_at)
		SELECT $1, $2, $3
		WHERE EXISTS (
			SELECT 1 FROM members m
			JOIN roles ro ON m.tenant_id = ro.tenant_id
			WHERE m.member_id = $1
			  AND ro.role_id = $2
			  AND m.deleted_at IS NULL
			  AND ro.deleted_at IS NULL
		)
		ON CONFLICT (member_id, role_id) DO NOTHING
	`

	_, err := r.db.Exec(ctx, query, memberID.String(), roleID.String(), time.Now())
	if err != nil {
		return fmt.Errorf("failed to assign role to member: %w", err)
	}

	return nil
}

// RemoveRole removes a role from a member
func (r *MemberRoleRepository) RemoveRole(ctx context.Context, memberID common.MemberID, roleID common.RoleID) error {
	query := `
		DELETE FROM member_roles mr
		USING members m
		WHERE mr.member_id = m.member_id
		  AND mr.member_id = $1
		  AND mr.role_id = $2
		  AND m.tenant_id = (SELECT ro.tenant_id FROM roles ro WHERE ro.role_id = $2)
	`

	result, err := r.db.Exec(ctx, query, memberID.String(), roleID.String())
	if err != nil {
		return fmt.Errorf("failed to remove role from member: %w", err)
	}

	if result.RowsAffected() == 0 {
		return common.NewNotFoundError("MemberRole", fmt.Sprintf("%s-%s", memberID.String(), roleID.String()))
	}

	return nil
}

// FindRolesByMemberID finds all roles assigned to a member
func (r *MemberRoleRepository) FindRolesByMemberID(ctx context.Context, memberID common.MemberID) ([]common.RoleID, error) {
	query := `
		SELECT mr.role_id
		FROM member_roles mr
		JOIN members m ON mr.member_id = m.member_id
		JOIN roles ro ON mr.role_id = ro.role_id AND ro.tenant_id = m.tenant_id
		WHERE mr.member_id = $1
		  AND m.deleted_at IS NULL
		  AND ro.deleted_at IS NULL
		ORDER BY mr.assigned_at
	`

	rows, err := r.db.Query(ctx, query, memberID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to query member roles: %w", err)
	}
	defer rows.Close()

	var roleIDs []common.RoleID
	for rows.Next() {
		var roleIDStr string
		if err := rows.Scan(&roleIDStr); err != nil {
			return nil, fmt.Errorf("failed to scan role ID: %w", err)
		}
		roleIDs = append(roleIDs, common.RoleID(roleIDStr))
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating role rows: %w", err)
	}

	return roleIDs, nil
}

// FindMemberIDsByRoleID finds all members with a specific role
func (r *MemberRoleRepository) FindMemberIDsByRoleID(ctx context.Context, roleID common.RoleID) ([]common.MemberID, error) {
	query := `
		SELECT mr.member_id
		FROM member_roles mr
		JOIN roles ro ON mr.role_id = ro.role_id
		JOIN members m ON mr.member_id = m.member_id AND m.tenant_id = ro.tenant_id
		WHERE mr.role_id = $1
		  AND ro.deleted_at IS NULL
		  AND m.deleted_at IS NULL
		ORDER BY mr.assigned_at
	`

	rows, err := r.db.Query(ctx, query, roleID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to query role members: %w", err)
	}
	defer rows.Close()

	var memberIDs []common.MemberID
	for rows.Next() {
		var memberIDStr string
		if err := rows.Scan(&memberIDStr); err != nil {
			return nil, fmt.Errorf("failed to scan member ID: %w", err)
		}
		memberIDs = append(memberIDs, common.MemberID(memberIDStr))
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating member rows: %w", err)
	}

	return memberIDs, nil
}

// SetMemberRoles sets all roles for a member (removes existing and adds new ones)
func (r *MemberRoleRepository) SetMemberRoles(ctx context.Context, memberID common.MemberID, roleIDs []common.RoleID) error {
	// Start a transaction
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// Delete existing roles (scoped to member's tenant via JOIN)
	_, err = tx.Exec(ctx, `
		DELETE FROM member_roles mr
		USING members m
		WHERE mr.member_id = m.member_id
		  AND mr.member_id = $1
	`, memberID.String())
	if err != nil {
		return fmt.Errorf("failed to delete existing member roles: %w", err)
	}

	// Insert new roles (only if role belongs to the same tenant as the member)
	if len(roleIDs) > 0 {
		for _, roleID := range roleIDs {
			_, err = tx.Exec(ctx, `
				INSERT INTO member_roles (member_id, role_id, assigned_at)
				SELECT $1, $2, $3
				WHERE EXISTS (
					SELECT 1 FROM members m
					JOIN roles ro ON m.tenant_id = ro.tenant_id
					WHERE m.member_id = $1
					  AND ro.role_id = $2
					  AND m.deleted_at IS NULL
					  AND ro.deleted_at IS NULL
				)
				ON CONFLICT (member_id, role_id) DO NOTHING
			`, memberID.String(), roleID.String(), time.Now())
			if err != nil {
				return fmt.Errorf("failed to insert member role: %w", err)
			}
		}
	}

	// Commit transaction
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
