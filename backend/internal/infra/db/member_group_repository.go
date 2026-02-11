package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/member"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// MemberGroupRepository implements member.MemberGroupRepository for PostgreSQL
type MemberGroupRepository struct {
	db *pgxpool.Pool
}

// NewMemberGroupRepository creates a new MemberGroupRepository
func NewMemberGroupRepository(db *pgxpool.Pool) *MemberGroupRepository {
	return &MemberGroupRepository{db: db}
}

// Save saves a member group (insert or update)
func (r *MemberGroupRepository) Save(ctx context.Context, group *member.MemberGroup) error {
	query := `
		INSERT INTO member_groups (
			group_id, tenant_id, name, description, color,
			display_order, created_at, updated_at, deleted_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (group_id) DO UPDATE SET
			name = EXCLUDED.name,
			description = EXCLUDED.description,
			color = EXCLUDED.color,
			display_order = EXCLUDED.display_order,
			updated_at = EXCLUDED.updated_at,
			deleted_at = EXCLUDED.deleted_at
	`

	_, err := r.db.Exec(ctx, query,
		group.GroupID().String(),
		group.TenantID().String(),
		group.Name(),
		nullString(group.Description()),
		nullString(group.Color()),
		group.DisplayOrder(),
		group.CreatedAt(),
		group.UpdatedAt(),
		group.DeletedAt(),
	)

	if err != nil {
		return fmt.Errorf("failed to save member group: %w", err)
	}

	return nil
}

// FindByID finds a member group by ID within a tenant
func (r *MemberGroupRepository) FindByID(ctx context.Context, tenantID common.TenantID, groupID common.MemberGroupID) (*member.MemberGroup, error) {
	query := `
		SELECT
			group_id, tenant_id, name, description, color,
			display_order, created_at, updated_at, deleted_at
		FROM member_groups
		WHERE tenant_id = $1 AND group_id = $2 AND deleted_at IS NULL
	`

	var (
		groupIDStr   string
		tenantIDStr  string
		name         string
		description  sql.NullString
		color        sql.NullString
		displayOrder int
		createdAt    time.Time
		updatedAt    time.Time
		deletedAt    sql.NullTime
	)

	err := r.db.QueryRow(ctx, query, tenantID.String(), groupID.String()).Scan(
		&groupIDStr,
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
		return nil, common.NewNotFoundError("MemberGroup", groupID.String())
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find member group: %w", err)
	}

	return r.scanToMemberGroup(
		groupIDStr, tenantIDStr, name, description, color,
		displayOrder, createdAt, updatedAt, deletedAt,
	)
}

// FindByTenantID finds all member groups within a tenant (deleted_at IS NULL)
func (r *MemberGroupRepository) FindByTenantID(ctx context.Context, tenantID common.TenantID) ([]*member.MemberGroup, error) {
	query := `
		SELECT
			group_id, tenant_id, name, description, color,
			display_order, created_at, updated_at, deleted_at
		FROM member_groups
		WHERE tenant_id = $1 AND deleted_at IS NULL
		ORDER BY display_order ASC, created_at DESC
	`

	return r.queryMemberGroups(ctx, query, tenantID.String())
}

// Delete deletes a member group (soft delete)
func (r *MemberGroupRepository) Delete(ctx context.Context, tenantID common.TenantID, groupID common.MemberGroupID) error {
	// グループエンティティを取得
	group, err := r.FindByID(ctx, tenantID, groupID)
	if err != nil {
		return err
	}

	// ソフトデリート実行
	now := time.Now()
	group.Delete(now)

	// 保存
	return r.Save(ctx, group)
}

// AssignMember assigns a member to a group
func (r *MemberGroupRepository) AssignMember(ctx context.Context, groupID common.MemberGroupID, memberID common.MemberID) error {
	query := `
		INSERT INTO member_group_assignments (assignment_id, member_id, group_id, created_at)
		SELECT $1::VARCHAR, $2::VARCHAR, $3::VARCHAR, $4::TIMESTAMPTZ
		WHERE EXISTS (
			SELECT 1 FROM members m
			JOIN member_groups mg ON m.tenant_id = mg.tenant_id
			WHERE m.member_id = $2
			  AND mg.group_id = $3
			  AND m.deleted_at IS NULL
			  AND mg.deleted_at IS NULL
		)
		ON CONFLICT (member_id, group_id) DO NOTHING
	`

	_, err := r.db.Exec(ctx, query,
		common.NewULID(),
		memberID.String(),
		groupID.String(),
		time.Now(),
	)

	if err != nil {
		return fmt.Errorf("failed to assign member to group: %w", err)
	}

	return nil
}

// RemoveMember removes a member from a group
func (r *MemberGroupRepository) RemoveMember(ctx context.Context, groupID common.MemberGroupID, memberID common.MemberID) error {
	query := `
		DELETE FROM member_group_assignments mga
		USING member_groups mg
		WHERE mga.group_id = mg.group_id
		  AND mga.group_id = $1
		  AND mga.member_id = $2
		  AND mg.tenant_id = (SELECT m.tenant_id FROM members m WHERE m.member_id = $2)
	`

	_, err := r.db.Exec(ctx, query, groupID.String(), memberID.String())
	if err != nil {
		return fmt.Errorf("failed to remove member from group: %w", err)
	}

	return nil
}

// FindMemberIDsByGroupID finds all members in a group
func (r *MemberGroupRepository) FindMemberIDsByGroupID(ctx context.Context, groupID common.MemberGroupID) ([]common.MemberID, error) {
	query := `
		SELECT mga.member_id
		FROM member_group_assignments mga
		JOIN member_groups mg ON mga.group_id = mg.group_id
		JOIN members m ON mga.member_id = m.member_id AND m.tenant_id = mg.tenant_id
		WHERE mga.group_id = $1
		  AND mg.deleted_at IS NULL
		  AND m.deleted_at IS NULL
		ORDER BY mga.created_at ASC
	`

	rows, err := r.db.Query(ctx, query, groupID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to find members by group ID: %w", err)
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
		return nil, fmt.Errorf("error iterating member IDs: %w", err)
	}

	return memberIDs, nil
}

// FindGroupIDsByMemberID finds all groups a member belongs to
func (r *MemberGroupRepository) FindGroupIDsByMemberID(ctx context.Context, memberID common.MemberID) ([]common.MemberGroupID, error) {
	query := `
		SELECT mga.group_id
		FROM member_group_assignments mga
		JOIN members m ON mga.member_id = m.member_id
		JOIN member_groups mg ON mga.group_id = mg.group_id AND mg.tenant_id = m.tenant_id
		WHERE mga.member_id = $1
		  AND m.deleted_at IS NULL
		  AND mg.deleted_at IS NULL
		ORDER BY mga.created_at ASC
	`

	rows, err := r.db.Query(ctx, query, memberID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to find groups by member ID: %w", err)
	}
	defer rows.Close()

	var groupIDs []common.MemberGroupID
	for rows.Next() {
		var groupIDStr string
		if err := rows.Scan(&groupIDStr); err != nil {
			return nil, fmt.Errorf("failed to scan group ID: %w", err)
		}
		groupIDs = append(groupIDs, common.MemberGroupID(groupIDStr))
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating group IDs: %w", err)
	}

	return groupIDs, nil
}

// SetMemberGroups sets all groups for a member (replaces existing groups)
func (r *MemberGroupRepository) SetMemberGroups(ctx context.Context, memberID common.MemberID, groupIDs []common.MemberGroupID) error {
	// 既存の関連を削除（メンバーのテナントに属するグループとの関連のみ）
	deleteQuery := `
		DELETE FROM member_group_assignments mga
		USING members m
		WHERE mga.member_id = m.member_id
		  AND mga.member_id = $1
	`
	_, err := r.db.Exec(ctx, deleteQuery, memberID.String())
	if err != nil {
		return fmt.Errorf("failed to delete existing group assignments: %w", err)
	}

	// 新しい関連を追加
	for _, groupID := range groupIDs {
		if err := r.AssignMember(ctx, groupID, memberID); err != nil {
			return err
		}
	}

	return nil
}

// queryMemberGroups executes a query and returns a list of member groups
func (r *MemberGroupRepository) queryMemberGroups(ctx context.Context, query string, args ...interface{}) ([]*member.MemberGroup, error) {
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query member groups: %w", err)
	}
	defer rows.Close()

	var groups []*member.MemberGroup
	for rows.Next() {
		var (
			groupIDStr   string
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
			&groupIDStr,
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
			return nil, fmt.Errorf("failed to scan member group row: %w", err)
		}

		group, err := r.scanToMemberGroup(
			groupIDStr, tenantIDStr, name, description, color,
			displayOrder, createdAt, updatedAt, deletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to reconstruct member group: %w", err)
		}

		groups = append(groups, group)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating member group rows: %w", err)
	}

	return groups, nil
}

// scanToMemberGroup converts scanned row data to MemberGroup entity
func (r *MemberGroupRepository) scanToMemberGroup(
	groupIDStr, tenantIDStr, name string,
	description, color sql.NullString,
	displayOrder int,
	createdAt, updatedAt time.Time,
	deletedAt sql.NullTime,
) (*member.MemberGroup, error) {
	var deletedAtPtr *time.Time
	if deletedAt.Valid {
		deletedAtPtr = &deletedAt.Time
	}

	return member.ReconstructMemberGroup(
		common.MemberGroupID(groupIDStr),
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
