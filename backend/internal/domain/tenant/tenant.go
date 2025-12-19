package tenant

import (
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
)

// Tenant represents a tenant entity (aggregate root)
// テナントは組織単位で、メンバー・イベント・シフトなどを管理する
type Tenant struct {
	tenantID   common.TenantID
	tenantName string
	timezone   string
	isActive   bool
	createdAt  time.Time
	updatedAt  time.Time
	deletedAt  *time.Time
}

// NewTenant creates a new Tenant entity
func NewTenant(
	now time.Time,
	tenantName string,
	timezone string,
) (*Tenant, error) {
	tenant := &Tenant{
		tenantID:   common.NewTenantID(),
		tenantName: tenantName,
		timezone:   timezone,
		isActive:   true,
		createdAt:  now,
		updatedAt:  now,
	}

	if err := tenant.validate(); err != nil {
		return nil, err
	}

	return tenant, nil
}

// ReconstructTenant reconstructs a Tenant entity from persistence
func ReconstructTenant(
	tenantID common.TenantID,
	tenantName string,
	timezone string,
	isActive bool,
	createdAt time.Time,
	updatedAt time.Time,
	deletedAt *time.Time,
) (*Tenant, error) {
	tenant := &Tenant{
		tenantID:   tenantID,
		tenantName: tenantName,
		timezone:   timezone,
		isActive:   isActive,
		createdAt:  createdAt,
		updatedAt:  updatedAt,
		deletedAt:  deletedAt,
	}

	if err := tenant.validate(); err != nil {
		return nil, err
	}

	return tenant, nil
}

func (t *Tenant) validate() error {
	// TenantID の検証
	if err := t.tenantID.Validate(); err != nil {
		return common.NewValidationError("tenant_id is invalid", err)
	}

	// TenantName の必須性チェック
	if t.tenantName == "" {
		return common.NewValidationError("tenant_name is required", nil)
	}
	if len(t.tenantName) > 255 {
		return common.NewValidationError("tenant_name must be less than 255 characters", nil)
	}

	// Timezone の必須性チェック
	if t.timezone == "" {
		return common.NewValidationError("timezone is required", nil)
	}
	if len(t.timezone) > 50 {
		return common.NewValidationError("timezone must be less than 50 characters", nil)
	}

	return nil
}

// Getters

func (t *Tenant) TenantID() common.TenantID {
	return t.tenantID
}

func (t *Tenant) TenantName() string {
	return t.tenantName
}

func (t *Tenant) Timezone() string {
	return t.timezone
}

func (t *Tenant) IsActive() bool {
	return t.isActive
}

func (t *Tenant) CreatedAt() time.Time {
	return t.createdAt
}

func (t *Tenant) UpdatedAt() time.Time {
	return t.updatedAt
}

func (t *Tenant) DeletedAt() *time.Time {
	return t.deletedAt
}

func (t *Tenant) IsDeleted() bool {
	return t.deletedAt != nil
}

// UpdateTenantName updates the tenant name
func (t *Tenant) UpdateTenantName(now time.Time, tenantName string) error {
	if tenantName == "" {
		return common.NewValidationError("tenant_name is required", nil)
	}
	if len(tenantName) > 255 {
		return common.NewValidationError("tenant_name must be less than 255 characters", nil)
	}

	t.tenantName = tenantName
	t.updatedAt = now
	return nil
}

// UpdateTimezone updates the timezone
func (t *Tenant) UpdateTimezone(now time.Time, timezone string) error {
	if timezone == "" {
		return common.NewValidationError("timezone is required", nil)
	}
	if len(timezone) > 50 {
		return common.NewValidationError("timezone must be less than 50 characters", nil)
	}

	t.timezone = timezone
	t.updatedAt = now
	return nil
}

// Activate activates the tenant
func (t *Tenant) Activate(now time.Time) {
	t.isActive = true
	t.updatedAt = now
}

// Deactivate deactivates the tenant
func (t *Tenant) Deactivate(now time.Time) {
	t.isActive = false
	t.updatedAt = now
}

// Delete marks the tenant as deleted (soft delete)
func (t *Tenant) Delete(now time.Time) {
	t.deletedAt = &now
	t.updatedAt = now
}
