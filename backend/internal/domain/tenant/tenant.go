package tenant

import (
	"fmt"
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
)

// TenantStatus represents the tenant's billing status
type TenantStatus string

const (
	TenantStatusActive         TenantStatus = "active"
	TenantStatusGrace          TenantStatus = "grace"
	TenantStatusSuspended      TenantStatus = "suspended"
	TenantStatusPendingPayment TenantStatus = "pending_payment"
)

// DefaultGracePeriodDays is the default number of days for grace period after subscription ends.
// This is a business rule: users have 14 days to re-subscribe after their subscription ends
// before their tenant becomes suspended.
const DefaultGracePeriodDays = 14

// ValidTenantStatuses returns all valid tenant statuses
func ValidTenantStatuses() []TenantStatus {
	return []TenantStatus{TenantStatusActive, TenantStatusGrace, TenantStatusSuspended, TenantStatusPendingPayment}
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

// validTransitions defines the allowed state transitions for a tenant.
// This encapsulates the business rules for tenant lifecycle:
//
//	pending_payment → active (payment completed)
//	pending_payment → suspended (payment failed/expired)
//	active → grace (subscription ended)
//	active → suspended (forced suspension by admin)
//	grace → active (re-subscribed)
//	grace → suspended (grace period expired)
//	suspended → pending_payment (re-subscription attempt)
//	suspended → active (admin manual activation)
var validTransitions = map[TenantStatus][]TenantStatus{
	TenantStatusPendingPayment: {TenantStatusActive, TenantStatusSuspended},
	TenantStatusActive:         {TenantStatusGrace, TenantStatusSuspended},
	TenantStatusGrace:          {TenantStatusActive, TenantStatusSuspended},
	TenantStatusSuspended:      {TenantStatusPendingPayment, TenantStatusActive},
}

// CanTransitionTo checks if the status can transition to the new status.
// Returns true if the transition is valid according to the business rules.
// Note: Same status transitions (e.g., active -> active) are allowed as they
// represent valid operations like payment renewals where the status doesn't change.
func (s TenantStatus) CanTransitionTo(newStatus TenantStatus) bool {
	// Same status is always allowed (no-op but valid for renewals, etc.)
	if s == newStatus {
		return true
	}
	allowed, ok := validTransitions[s]
	if !ok {
		return false
	}
	for _, status := range allowed {
		if status == newStatus {
			return true
		}
	}
	return false
}

// Tenant represents a tenant entity (aggregate root)
// テナントは組織単位で、メンバー・イベント・シフトなどを管理する
type Tenant struct {
	tenantID               common.TenantID
	tenantName             string
	timezone               string
	isActive               bool
	status                 TenantStatus
	graceUntil             *time.Time
	pendingExpiresAt       *time.Time
	pendingStripeSessionID *string
	createdAt              time.Time
	updatedAt              time.Time
	deletedAt              *time.Time
}

// NewTenant creates a new Tenant entity
func NewTenant(
	now time.Time,
	tenantName string,
	timezone string,
) (*Tenant, error) {
	tenant := &Tenant{
		tenantID:               common.NewTenantID(),
		tenantName:             tenantName,
		timezone:               timezone,
		isActive:               true,
		status:                 TenantStatusActive,
		graceUntil:             nil,
		pendingExpiresAt:       nil,
		pendingStripeSessionID: nil,
		createdAt:              now,
		updatedAt:              now,
	}

	if err := tenant.validate(); err != nil {
		return nil, err
	}

	return tenant, nil
}

// NewTenantPendingPayment creates a new Tenant entity in pending_payment status
func NewTenantPendingPayment(
	now time.Time,
	tenantName string,
	timezone string,
	stripeSessionID string,
	pendingExpiresAt time.Time,
) (*Tenant, error) {
	tenant := &Tenant{
		tenantID:               common.NewTenantID(),
		tenantName:             tenantName,
		timezone:               timezone,
		isActive:               false,
		status:                 TenantStatusPendingPayment,
		graceUntil:             nil,
		pendingExpiresAt:       &pendingExpiresAt,
		pendingStripeSessionID: &stripeSessionID,
		createdAt:              now,
		updatedAt:              now,
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
	pendingExpiresAt *time.Time,
	pendingStripeSessionID *string,
	createdAt time.Time,
	updatedAt time.Time,
	deletedAt *time.Time,
) (*Tenant, error) {
	tenant := &Tenant{
		tenantID:               tenantID,
		tenantName:             tenantName,
		timezone:               timezone,
		isActive:               isActive,
		status:                 status,
		graceUntil:             graceUntil,
		pendingExpiresAt:       pendingExpiresAt,
		pendingStripeSessionID: pendingStripeSessionID,
		createdAt:              createdAt,
		updatedAt:              updatedAt,
		deletedAt:              deletedAt,
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
	// IANA タイムゾーン形式の検証
	if _, err := time.LoadLocation(t.timezone); err != nil {
		return common.NewValidationError("invalid timezone format", err)
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

func (t *Tenant) PendingExpiresAt() *time.Time {
	return t.pendingExpiresAt
}

func (t *Tenant) PendingStripeSessionID() *string {
	return t.pendingStripeSessionID
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

// SetStatusActive sets the tenant status to active.
// Returns an error if the transition is not allowed from the current status.
func (t *Tenant) SetStatusActive(now time.Time) error {
	if !t.status.CanTransitionTo(TenantStatusActive) {
		return common.NewValidationError(
			fmt.Sprintf("invalid status transition from %s to active", t.status), nil)
	}
	t.status = TenantStatusActive
	t.graceUntil = nil
	t.pendingExpiresAt = nil
	t.pendingStripeSessionID = nil
	t.isActive = true
	t.updatedAt = now
	return nil
}

// SetStatusGrace sets the tenant status to grace with a grace period end time.
// Returns an error if the transition is not allowed from the current status.
func (t *Tenant) SetStatusGrace(now time.Time, graceUntil time.Time) error {
	if !t.status.CanTransitionTo(TenantStatusGrace) {
		return common.NewValidationError(
			fmt.Sprintf("invalid status transition from %s to grace", t.status), nil)
	}
	t.status = TenantStatusGrace
	t.graceUntil = &graceUntil
	t.isActive = false
	t.updatedAt = now
	return nil
}

// CalculateGraceUntil calculates the grace period end date based on the subscription period end.
// This encapsulates the business rule: grace_until = periodEnd + DefaultGracePeriodDays (14 days).
func CalculateGraceUntil(periodEnd time.Time) time.Time {
	return periodEnd.AddDate(0, 0, DefaultGracePeriodDays)
}

// TransitionToGraceAfterSubscriptionEnd transitions the tenant to grace status
// after their subscription ends. This is called when:
// - User cancels subscription (cancel_at_period_end) and period ends
// - Payment fails and all retries are exhausted
//
// The grace period gives users 14 days to re-subscribe before suspension.
// Returns an error if the transition is not allowed from the current status.
func (t *Tenant) TransitionToGraceAfterSubscriptionEnd(now time.Time, periodEnd time.Time) error {
	graceUntil := CalculateGraceUntil(periodEnd)
	return t.SetStatusGrace(now, graceUntil)
}

// SetStatusSuspended sets the tenant status to suspended.
// Returns an error if the transition is not allowed from the current status.
func (t *Tenant) SetStatusSuspended(now time.Time) error {
	if !t.status.CanTransitionTo(TenantStatusSuspended) {
		return common.NewValidationError(
			fmt.Sprintf("invalid status transition from %s to suspended", t.status), nil)
	}
	t.status = TenantStatusSuspended
	t.graceUntil = nil
	t.pendingExpiresAt = nil
	t.pendingStripeSessionID = nil
	t.isActive = false
	t.updatedAt = now
	return nil
}

// SetStatusPendingPayment sets the tenant status to pending_payment.
// Returns an error if the transition is not allowed from the current status.
func (t *Tenant) SetStatusPendingPayment(now time.Time, stripeSessionID string, expiresAt time.Time) error {
	if !t.status.CanTransitionTo(TenantStatusPendingPayment) {
		return common.NewValidationError(
			fmt.Sprintf("invalid status transition from %s to pending_payment", t.status), nil)
	}
	t.status = TenantStatusPendingPayment
	t.graceUntil = nil
	t.pendingExpiresAt = &expiresAt
	t.pendingStripeSessionID = &stripeSessionID
	t.isActive = false
	t.updatedAt = now
	return nil
}

// IsPendingPayment checks if the tenant is in pending payment status
func (t *Tenant) IsPendingPayment() bool {
	return t.status == TenantStatusPendingPayment
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
