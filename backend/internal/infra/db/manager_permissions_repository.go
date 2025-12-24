package db

import (
	"context"
	"fmt"
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/tenant"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ManagerPermissionsRepository implements tenant.ManagerPermissionsRepository for PostgreSQL
type ManagerPermissionsRepository struct {
	db *pgxpool.Pool
}

// NewManagerPermissionsRepository creates a new ManagerPermissionsRepository
func NewManagerPermissionsRepository(db *pgxpool.Pool) *ManagerPermissionsRepository {
	return &ManagerPermissionsRepository{db: db}
}

// FindByTenantID finds manager permissions by tenant ID
func (r *ManagerPermissionsRepository) FindByTenantID(ctx context.Context, tenantID common.TenantID) (*tenant.ManagerPermissions, error) {
	query := `
		SELECT
			tenant_id,
			can_add_member, can_edit_member, can_delete_member,
			can_create_event, can_edit_event, can_delete_event,
			can_assign_shift, can_edit_shift,
			can_create_attendance, can_create_schedule,
			can_manage_roles, can_manage_positions, can_manage_groups,
			can_invite_manager,
			created_at, updated_at
		FROM manager_permissions
		WHERE tenant_id = $1
	`

	var (
		tenantIDStr         string
		canAddMember        bool
		canEditMember       bool
		canDeleteMember     bool
		canCreateEvent      bool
		canEditEvent        bool
		canDeleteEvent      bool
		canAssignShift      bool
		canEditShift        bool
		canCreateAttendance bool
		canCreateSchedule   bool
		canManageRoles      bool
		canManagePositions  bool
		canManageGroups     bool
		canInviteManager    bool
		createdAt           time.Time
		updatedAt           time.Time
	)

	err := r.db.QueryRow(ctx, query, tenantID.String()).Scan(
		&tenantIDStr,
		&canAddMember,
		&canEditMember,
		&canDeleteMember,
		&canCreateEvent,
		&canEditEvent,
		&canDeleteEvent,
		&canAssignShift,
		&canEditShift,
		&canCreateAttendance,
		&canCreateSchedule,
		&canManageRoles,
		&canManagePositions,
		&canManageGroups,
		&canInviteManager,
		&createdAt,
		&updatedAt,
	)

	if err == pgx.ErrNoRows {
		// 設定が存在しない場合はnilを返す（デフォルト値を使用）
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find manager permissions: %w", err)
	}

	parsedTenantID, err := common.ParseTenantID(tenantIDStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse tenant_id: %w", err)
	}

	return tenant.ReconstructManagerPermissions(
		parsedTenantID,
		canAddMember,
		canEditMember,
		canDeleteMember,
		canCreateEvent,
		canEditEvent,
		canDeleteEvent,
		canAssignShift,
		canEditShift,
		canCreateAttendance,
		canCreateSchedule,
		canManageRoles,
		canManagePositions,
		canManageGroups,
		canInviteManager,
		createdAt,
		updatedAt,
	), nil
}

// Save saves manager permissions (insert or update)
func (r *ManagerPermissionsRepository) Save(ctx context.Context, p *tenant.ManagerPermissions) error {
	query := `
		INSERT INTO manager_permissions (
			tenant_id,
			can_add_member, can_edit_member, can_delete_member,
			can_create_event, can_edit_event, can_delete_event,
			can_assign_shift, can_edit_shift,
			can_create_attendance, can_create_schedule,
			can_manage_roles, can_manage_positions, can_manage_groups,
			can_invite_manager,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
		ON CONFLICT (tenant_id) DO UPDATE SET
			can_add_member = EXCLUDED.can_add_member,
			can_edit_member = EXCLUDED.can_edit_member,
			can_delete_member = EXCLUDED.can_delete_member,
			can_create_event = EXCLUDED.can_create_event,
			can_edit_event = EXCLUDED.can_edit_event,
			can_delete_event = EXCLUDED.can_delete_event,
			can_assign_shift = EXCLUDED.can_assign_shift,
			can_edit_shift = EXCLUDED.can_edit_shift,
			can_create_attendance = EXCLUDED.can_create_attendance,
			can_create_schedule = EXCLUDED.can_create_schedule,
			can_manage_roles = EXCLUDED.can_manage_roles,
			can_manage_positions = EXCLUDED.can_manage_positions,
			can_manage_groups = EXCLUDED.can_manage_groups,
			can_invite_manager = EXCLUDED.can_invite_manager,
			updated_at = EXCLUDED.updated_at
	`

	_, err := r.db.Exec(ctx, query,
		p.TenantID().String(),
		p.CanAddMember(),
		p.CanEditMember(),
		p.CanDeleteMember(),
		p.CanCreateEvent(),
		p.CanEditEvent(),
		p.CanDeleteEvent(),
		p.CanAssignShift(),
		p.CanEditShift(),
		p.CanCreateAttendance(),
		p.CanCreateSchedule(),
		p.CanManageRoles(),
		p.CanManagePositions(),
		p.CanManageGroups(),
		p.CanInviteManager(),
		p.CreatedAt(),
		p.UpdatedAt(),
	)

	if err != nil {
		return fmt.Errorf("failed to save manager permissions: %w", err)
	}

	return nil
}
