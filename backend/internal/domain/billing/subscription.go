package billing

import (
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
)

// SubscriptionID represents the unique identifier for a subscription
type SubscriptionID string

// NewSubscriptionIDWithTime generates a new SubscriptionID using the provided time.
func NewSubscriptionIDWithTime(t time.Time) SubscriptionID {
	return SubscriptionID(common.NewULIDWithTime(t))
}

// NewSubscriptionID generates a new SubscriptionID using the current time.
// Deprecated: Use NewSubscriptionIDWithTime for better testability.
func NewSubscriptionID() SubscriptionID {
	return SubscriptionID(common.NewULID())
}

// ParseSubscriptionID parses a string into SubscriptionID
func ParseSubscriptionID(id string) (SubscriptionID, error) {
	if err := common.ValidateULID(id); err != nil {
		return "", err
	}
	return SubscriptionID(id), nil
}

// String returns the string representation
func (id SubscriptionID) String() string {
	return string(id)
}

// SubscriptionStatus represents the status of a Stripe subscription
type SubscriptionStatus string

const (
	SubscriptionStatusActive     SubscriptionStatus = "active"
	SubscriptionStatusPastDue    SubscriptionStatus = "past_due"
	SubscriptionStatusCanceled   SubscriptionStatus = "canceled"
	SubscriptionStatusUnpaid     SubscriptionStatus = "unpaid"
	SubscriptionStatusIncomplete SubscriptionStatus = "incomplete"
	SubscriptionStatusTrialing   SubscriptionStatus = "trialing"
)

// String returns the string representation
func (s SubscriptionStatus) String() string {
	return string(s)
}

// IsValid checks if the status is valid
func (s SubscriptionStatus) IsValid() bool {
	switch s {
	case SubscriptionStatusActive, SubscriptionStatusPastDue,
		SubscriptionStatusCanceled, SubscriptionStatusUnpaid,
		SubscriptionStatusIncomplete, SubscriptionStatusTrialing:
		return true
	}
	return false
}

// Subscription represents a Stripe subscription.
//
// # Aggregate Relationship with Tenant
//
// Subscription is an independent aggregate from Tenant, but they are closely related:
//
//   - Subscription holds TenantID as a foreign key reference
//   - Both are aggregate roots with their own identity (SubscriptionID, TenantID)
//   - Subscription lifecycle events affect Tenant state:
//     - checkout.session.completed: Tenant -> active
//     - invoice.payment_failed: Tenant -> grace (after retries exhausted)
//     - customer.subscription.deleted: Tenant -> grace -> suspended
//
// # Design Rationale
//
// This design (separate aggregates with coordinated updates via Webhook handler)
// is chosen over embedding Subscription in Tenant because:
//
//  1. Stripe is the source of truth for subscription data
//  2. Webhook events are the primary mechanism for state synchronization
//  3. Keeping them separate allows cleaner domain boundaries
//  4. Transaction boundaries are clear: Webhook handler coordinates both updates
//
// The tradeoff is that consistency between Subscription and Tenant state
// depends on correct Webhook processing. See stripe_webhook_usecase.go for details.
type Subscription struct {
	subscriptionID       SubscriptionID
	tenantID             common.TenantID
	stripeCustomerID     string
	stripeSubscriptionID string
	status               SubscriptionStatus
	currentPeriodEnd     *time.Time
	cancelAtPeriodEnd    bool       // キャンセル予約中かどうか
	cancelAt             *time.Time // キャンセル予定日時
	createdAt            time.Time
	updatedAt            time.Time
}

// NewSubscription creates a new Subscription entity
func NewSubscription(
	now time.Time,
	tenantID common.TenantID,
	stripeCustomerID string,
	stripeSubscriptionID string,
	status SubscriptionStatus,
	currentPeriodEnd *time.Time,
) (*Subscription, error) {
	sub := &Subscription{
		subscriptionID:       NewSubscriptionIDWithTime(now),
		tenantID:             tenantID,
		stripeCustomerID:     stripeCustomerID,
		stripeSubscriptionID: stripeSubscriptionID,
		status:               status,
		currentPeriodEnd:     currentPeriodEnd,
		createdAt:            now,
		updatedAt:            now,
	}

	if err := sub.validate(); err != nil {
		return nil, err
	}

	return sub, nil
}

// ReconstructSubscription reconstructs a Subscription entity from persistence
func ReconstructSubscription(
	subscriptionID SubscriptionID,
	tenantID common.TenantID,
	stripeCustomerID string,
	stripeSubscriptionID string,
	status SubscriptionStatus,
	currentPeriodEnd *time.Time,
	cancelAtPeriodEnd bool,
	cancelAt *time.Time,
	createdAt time.Time,
	updatedAt time.Time,
) (*Subscription, error) {
	sub := &Subscription{
		subscriptionID:       subscriptionID,
		tenantID:             tenantID,
		stripeCustomerID:     stripeCustomerID,
		stripeSubscriptionID: stripeSubscriptionID,
		status:               status,
		currentPeriodEnd:     currentPeriodEnd,
		cancelAtPeriodEnd:    cancelAtPeriodEnd,
		cancelAt:             cancelAt,
		createdAt:            createdAt,
		updatedAt:            updatedAt,
	}

	if err := sub.validate(); err != nil {
		return nil, err
	}

	return sub, nil
}

func (s *Subscription) validate() error {
	if s.stripeCustomerID == "" {
		return common.NewValidationError("stripe_customer_id is required", nil)
	}
	if s.stripeSubscriptionID == "" {
		return common.NewValidationError("stripe_subscription_id is required", nil)
	}
	if !s.status.IsValid() {
		return common.NewValidationError("invalid subscription status", nil)
	}
	return nil
}

// Getters

func (s *Subscription) SubscriptionID() SubscriptionID {
	return s.subscriptionID
}

func (s *Subscription) TenantID() common.TenantID {
	return s.tenantID
}

func (s *Subscription) StripeCustomerID() string {
	return s.stripeCustomerID
}

func (s *Subscription) StripeSubscriptionID() string {
	return s.stripeSubscriptionID
}

func (s *Subscription) Status() SubscriptionStatus {
	return s.status
}

func (s *Subscription) CurrentPeriodEnd() *time.Time {
	return s.currentPeriodEnd
}

func (s *Subscription) CreatedAt() time.Time {
	return s.createdAt
}

func (s *Subscription) UpdatedAt() time.Time {
	return s.updatedAt
}

// IsActive checks if the subscription is active
func (s *Subscription) IsActive() bool {
	return s.status == SubscriptionStatusActive || s.status == SubscriptionStatusTrialing
}

// UpdateStatus updates the subscription status
func (s *Subscription) UpdateStatus(now time.Time, status SubscriptionStatus, currentPeriodEnd *time.Time) {
	s.status = status
	s.currentPeriodEnd = currentPeriodEnd
	s.updatedAt = now
}

// CancelAtPeriodEnd returns whether the subscription is scheduled to cancel at period end
func (s *Subscription) CancelAtPeriodEnd() bool {
	return s.cancelAtPeriodEnd
}

// CancelAt returns the scheduled cancellation time
func (s *Subscription) CancelAt() *time.Time {
	return s.cancelAt
}

// SetCancelAtPeriodEnd updates the cancellation schedule
func (s *Subscription) SetCancelAtPeriodEnd(now time.Time, cancelAtPeriodEnd bool, cancelAt *time.Time) {
	s.cancelAtPeriodEnd = cancelAtPeriodEnd
	s.cancelAt = cancelAt
	s.updatedAt = now
}
