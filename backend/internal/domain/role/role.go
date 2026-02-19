package role

import (
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
)

// Role represents a role entity (aggregate root)
// ロールはメンバーに付与される役割・属性を表す
type Role struct {
	roleID       common.RoleID
	tenantID     common.TenantID
	name         string
	description  string
	color        string // UI表示用の色コード（例: #FF5733）
	displayOrder int    // 表示順序
	createdAt    time.Time
	updatedAt    time.Time
	deletedAt    *time.Time
}

// NewRole creates a new Role entity
func NewRole(
	now time.Time,
	tenantID common.TenantID,
	name string,
	description string,
	color string,
	displayOrder int,
) (*Role, error) {
	role := &Role{
		roleID:       common.NewRoleID(),
		tenantID:     tenantID,
		name:         name,
		description:  description,
		color:        color,
		displayOrder: displayOrder,
		createdAt:    now,
		updatedAt:    now,
	}

	if err := role.validate(); err != nil {
		return nil, err
	}

	return role, nil
}

// ReconstructRole reconstructs a Role entity from persistence
func ReconstructRole(
	roleID common.RoleID,
	tenantID common.TenantID,
	name string,
	description string,
	color string,
	displayOrder int,
	createdAt time.Time,
	updatedAt time.Time,
	deletedAt *time.Time,
) (*Role, error) {
	role := &Role{
		roleID:       roleID,
		tenantID:     tenantID,
		name:         name,
		description:  description,
		color:        color,
		displayOrder: displayOrder,
		createdAt:    createdAt,
		updatedAt:    updatedAt,
		deletedAt:    deletedAt,
	}

	if err := role.validate(); err != nil {
		return nil, err
	}

	return role, nil
}

func (r *Role) validate() error {
	// TenantID の必須性チェック
	if err := r.tenantID.Validate(); err != nil {
		return common.NewValidationError("tenant_id is required", err)
	}

	// Name の必須性チェック
	if r.name == "" {
		return common.NewValidationError("name is required", nil)
	}

	if len(r.name) > 100 {
		return common.NewValidationError("name must be less than 100 characters", nil)
	}

	// Description の長さチェック
	if len(r.description) > 500 {
		return common.NewValidationError("description must be less than 500 characters", nil)
	}

	// Color の長さチェック（オプショナルだが、設定されている場合）
	if r.color != "" && len(r.color) > 20 {
		return common.NewValidationError("color must be less than 20 characters", nil)
	}

	return nil
}

// Getters
func (r *Role) RoleID() common.RoleID {
	return r.roleID
}

func (r *Role) TenantID() common.TenantID {
	return r.tenantID
}

func (r *Role) Name() string {
	return r.name
}

func (r *Role) Description() string {
	return r.description
}

func (r *Role) Color() string {
	return r.color
}

func (r *Role) DisplayOrder() int {
	return r.displayOrder
}

func (r *Role) CreatedAt() time.Time {
	return r.createdAt
}

func (r *Role) UpdatedAt() time.Time {
	return r.updatedAt
}

func (r *Role) DeletedAt() *time.Time {
	return r.deletedAt
}

// UpdateDetails updates role details
func (r *Role) UpdateDetails(now time.Time, name, description, color string, displayOrder int) error {
	// Validate before mutating using a temporary copy
	tmp := *r
	tmp.name = name
	tmp.description = description
	tmp.color = color
	tmp.displayOrder = displayOrder
	tmp.updatedAt = now
	if err := tmp.validate(); err != nil {
		return err
	}

	// Apply validated changes
	r.name = name
	r.description = description
	r.color = color
	r.displayOrder = displayOrder
	r.updatedAt = now
	return nil
}

// Delete soft deletes the role
func (r *Role) Delete(now time.Time) {
	r.deletedAt = &now
	r.updatedAt = now
}
