package auth

import (
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
)

// Admin represents an admin entity (aggregate root)
// 管理者（店長/副店長）はテナント内の管理操作を行う権限を持つ
type Admin struct {
	adminID      common.AdminID
	tenantID     common.TenantID
	email        string
	passwordHash string // ドメインはハッシュを保持するが、bcrypt処理はしない
	displayName  string
	role         Role
	isActive     bool
	createdAt    time.Time
	updatedAt    time.Time
	deletedAt    *time.Time
}

// NewAdmin creates a new Admin entity
// NOTE: passwordHash は既にハッシュ化された値を受け取る（bcrypt処理はInfra層で行う）
// NOTE: now は App層で clock.Now() を呼んで渡す（Domain層で time.Now() を呼ばない）
func NewAdmin(
	now time.Time,
	tenantID common.TenantID,
	email string,
	passwordHash string,
	displayName string,
	role Role,
) (*Admin, error) {
	admin := &Admin{
		adminID:      common.NewAdminID(),
		tenantID:     tenantID,
		email:        email,
		passwordHash: passwordHash,
		displayName:  displayName,
		role:         role,
		isActive:     true,
		createdAt:    now,
		updatedAt:    now,
	}

	if err := admin.validate(); err != nil {
		return nil, err
	}

	return admin, nil
}

// ReconstructAdmin reconstructs an Admin entity from persistence
func ReconstructAdmin(
	adminID common.AdminID,
	tenantID common.TenantID,
	email string,
	passwordHash string,
	displayName string,
	role Role,
	isActive bool,
	createdAt time.Time,
	updatedAt time.Time,
	deletedAt *time.Time,
) (*Admin, error) {
	admin := &Admin{
		adminID:      adminID,
		tenantID:     tenantID,
		email:        email,
		passwordHash: passwordHash,
		displayName:  displayName,
		role:         role,
		isActive:     isActive,
		createdAt:    createdAt,
		updatedAt:    updatedAt,
		deletedAt:    deletedAt,
	}

	if err := admin.validate(); err != nil {
		return nil, err
	}

	return admin, nil
}

func (a *Admin) validate() error {
	// TenantID の必須性チェック
	if err := a.tenantID.Validate(); err != nil {
		return common.NewValidationError("tenant_id is required", err)
	}

	// Email の必須性チェック
	if a.email == "" {
		return common.NewValidationError("email is required", nil)
	}
	if len(a.email) > 255 {
		return common.NewValidationError("email must be less than 255 characters", nil)
	}

	// PasswordHash の必須性チェック
	if a.passwordHash == "" {
		return common.NewValidationError("password_hash is required", nil)
	}

	// DisplayName の必須性チェック
	if a.displayName == "" {
		return common.NewValidationError("display_name is required", nil)
	}
	if len(a.displayName) > 255 {
		return common.NewValidationError("display_name must be less than 255 characters", nil)
	}

	// Role の検証
	if err := a.role.Validate(); err != nil {
		return err
	}

	return nil
}

// CanLogin は認証可能かを判定（ドメインルール）
func (a *Admin) CanLogin() bool {
	return a.isActive && a.deletedAt == nil
}

// Getters

func (a *Admin) AdminID() common.AdminID {
	return a.adminID
}

func (a *Admin) TenantID() common.TenantID {
	return a.tenantID
}

func (a *Admin) Email() string {
	return a.email
}

// PasswordHash は認証処理用にハッシュを返す（App/Infra層でのみ使用）
func (a *Admin) PasswordHash() string {
	return a.passwordHash
}

func (a *Admin) DisplayName() string {
	return a.displayName
}

func (a *Admin) Role() Role {
	return a.role
}

func (a *Admin) IsActive() bool {
	return a.isActive
}

func (a *Admin) CreatedAt() time.Time {
	return a.createdAt
}

func (a *Admin) UpdatedAt() time.Time {
	return a.updatedAt
}

func (a *Admin) DeletedAt() *time.Time {
	return a.deletedAt
}

func (a *Admin) IsDeleted() bool {
	return a.deletedAt != nil
}

// UpdateEmail updates the email address
func (a *Admin) UpdateEmail(now time.Time, email string) error {
	if email == "" {
		return common.NewValidationError("email is required", nil)
	}
	if len(email) > 255 {
		return common.NewValidationError("email must be less than 255 characters", nil)
	}

	a.email = email
	a.updatedAt = now
	return nil
}

// UpdatePasswordHash updates the password hash
// NOTE: passwordHash は既にハッシュ化された値を受け取る（bcrypt処理はInfra層で行う）
func (a *Admin) UpdatePasswordHash(now time.Time, passwordHash string) error {
	if passwordHash == "" {
		return common.NewValidationError("password_hash is required", nil)
	}

	a.passwordHash = passwordHash
	a.updatedAt = now
	return nil
}

// UpdateDisplayName updates the display name
func (a *Admin) UpdateDisplayName(now time.Time, displayName string) error {
	if displayName == "" {
		return common.NewValidationError("display_name is required", nil)
	}
	if len(displayName) > 255 {
		return common.NewValidationError("display_name must be less than 255 characters", nil)
	}

	a.displayName = displayName
	a.updatedAt = now
	return nil
}

// UpdateRole updates the role
func (a *Admin) UpdateRole(now time.Time, role Role) error {
	if err := role.Validate(); err != nil {
		return err
	}

	a.role = role
	a.updatedAt = now
	return nil
}

// Activate activates the admin
func (a *Admin) Activate(now time.Time) {
	a.isActive = true
	a.updatedAt = now
}

// Deactivate deactivates the admin
func (a *Admin) Deactivate(now time.Time) {
	a.isActive = false
	a.updatedAt = now
}

// Delete marks the admin as deleted (soft delete)
func (a *Admin) Delete(now time.Time) {
	a.deletedAt = &now
	a.updatedAt = now
}
