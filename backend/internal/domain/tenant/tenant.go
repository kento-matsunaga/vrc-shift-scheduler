package tenant

import (
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
)

// TenantStatus represents the tenant's billing status
type TenantStatus string

const (
	TenantStatusActive    TenantStatus = "active"
	TenantStatusGrace     TenantStatus = "grace"
	TenantStatusSuspended TenantStatus = "suspended"
)

// ValidTenantStatuses returns all valid tenant statuses
func ValidTenantStatuses() []TenantStatus {
	return []TenantStatus{TenantStatusActive, TenantStatusGrace, TenantStatusSuspended}
}

// IsValid checks if the status is valid
func (s TenantStatus) IsValid() bool {
	for _, valid := range ValidTenantStatuses() {
		if s == valid {
			return true
		}
	}
	return false
}

// String returns the string representation
func (s TenantStatus) String() string {
	return string(s)
}

// Tenant represents a tenant entity (aggregate root)
// テナントは組織単位で、メンバー・イベント・シフトなどを管理する
type Tenant struct {
	tenantID   common.TenantID
	tenantName string
	timezone   string
	isActive   bool
	status     TenantStatus
	graceUntil *time.Time
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
		status:     TenantStatusActive,
		graceUntil: nil,
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
	status TenantStatus,
	graceUntil *time.Time,
	createdAt time.Time,
	updatedAt time.Time,
	deletedAt *time.Time,
) (*Tenant, error) {
	tenant := &Tenant{
		tenantID:   tenantID,
		tenantName: tenantName,
		timezone:   timezone,
		isActive:   isActive,
		status:     status,
		graceUntil: graceUntil,
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

func (t *Tenant) Status() TenantStatus {
	return t.status
}

func (t *Tenant) GraceUntil() *time.Time {
	return t.graceUntil
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

// SetStatusActive sets the tenant status to active
func (t *Tenant) SetStatusActive(now time.Time) {
	t.status = TenantStatusActive
	t.graceUntil = nil
	t.isActive = true
	t.updatedAt = now
}

// SetStatusGrace sets the tenant status to grace with a grace period end time
func (t *Tenant) SetStatusGrace(now time.Time, graceUntil time.Time) {
	t.status = TenantStatusGrace
	t.graceUntil = &graceUntil
	t.isActive = false
	t.updatedAt = now
}

// SetStatusSuspended sets the tenant status to suspended
func (t *Tenant) SetStatusSuspended(now time.Time) {
	t.status = TenantStatusSuspended
	t.graceUntil = nil
	t.isActive = false
	t.updatedAt = now
}

// CanWrite checks if write operations are allowed for this tenant
func (t *Tenant) CanWrite() bool {
	return t.status == TenantStatusActive && !t.IsDeleted()
}

// CanRead checks if read operations are allowed for this tenant
func (t *Tenant) CanRead() bool {
	return (t.status == TenantStatusActive || t.status == TenantStatusGrace || t.status == TenantStatusSuspended) && !t.IsDeleted()
}

// IsInGracePeriod checks if the tenant is currently in grace period
func (t *Tenant) IsInGracePeriod() bool {
	return t.status == TenantStatusGrace
}

// IsSuspended checks if the tenant is suspended
func (t *Tenant) IsSuspended() bool {
	return t.status == TenantStatusSuspended
}
