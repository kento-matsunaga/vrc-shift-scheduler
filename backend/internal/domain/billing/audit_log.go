package billing

import (
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
)

// BillingAuditLogID represents the unique identifier for a billing audit log
type BillingAuditLogID string

// NewBillingAuditLogID generates a new BillingAuditLogID
func NewBillingAuditLogID() BillingAuditLogID {
	return BillingAuditLogID(common.NewULID())
}

// String returns the string representation
func (id BillingAuditLogID) String() string {
	return string(id)
}

// ActorType represents the type of actor that performed the action
type ActorType string

const (
	ActorTypeAdmin  ActorType = "admin"
	ActorTypeSystem ActorType = "system"
	ActorTypeStripe ActorType = "stripe"
	ActorTypeUser   ActorType = "user"
)

// String returns the string representation
func (at ActorType) String() string {
	return string(at)
}

// IsValid checks if the actor type is valid
func (at ActorType) IsValid() bool {
	switch at {
	case ActorTypeAdmin, ActorTypeSystem, ActorTypeStripe, ActorTypeUser:
		return true
	}
	return false
}

// BillingAuditAction represents common billing audit actions
type BillingAuditAction string

const (
	BillingAuditActionLicenseClaim       BillingAuditAction = "license_claim"
	BillingAuditActionLicenseClaimFailed BillingAuditAction = "license_claim_failed"
	BillingAuditActionLicenseGenerated   BillingAuditAction = "license_generated"
	BillingAuditActionLicenseRevoked     BillingAuditAction = "license_revoked"
	BillingAuditActionSubscriptionCreate BillingAuditAction = "subscription_created"
	BillingAuditActionSubscriptionUpdate BillingAuditAction = "subscription_updated"
	BillingAuditActionTenantStatusChange BillingAuditAction = "tenant_status_changed"
	BillingAuditActionPaymentFailed      BillingAuditAction = "payment_failed"
	BillingAuditActionPaymentSucceeded   BillingAuditAction = "payment_succeeded"
	BillingAuditActionEntitlementRevoked BillingAuditAction = "entitlement_revoked"
)

// String returns the string representation
func (a BillingAuditAction) String() string {
	return string(a)
}

// BillingAuditLog represents a billing-related audit log entry
type BillingAuditLog struct {
	logID      BillingAuditLogID
	actorType  ActorType
	actorID    *string
	action     string
	targetType *string
	targetID   *string
	beforeJSON *string
	afterJSON  *string
	ipAddress  *string
	userAgent  *string
	createdAt  time.Time
}

// NewBillingAuditLog creates a new BillingAuditLog entity
func NewBillingAuditLog(
	now time.Time,
	actorType ActorType,
	actorID *string,
	action string,
	targetType *string,
	targetID *string,
	beforeJSON *string,
	afterJSON *string,
	ipAddress *string,
	userAgent *string,
) (*BillingAuditLog, error) {
	log := &BillingAuditLog{
		logID:      NewBillingAuditLogID(),
		actorType:  actorType,
		actorID:    actorID,
		action:     action,
		targetType: targetType,
		targetID:   targetID,
		beforeJSON: beforeJSON,
		afterJSON:  afterJSON,
		ipAddress:  ipAddress,
		userAgent:  userAgent,
		createdAt:  now,
	}

	if err := log.validate(); err != nil {
		return nil, err
	}

	return log, nil
}

// ReconstructBillingAuditLog reconstructs a BillingAuditLog entity from persistence
func ReconstructBillingAuditLog(
	logID BillingAuditLogID,
	actorType ActorType,
	actorID *string,
	action string,
	targetType *string,
	targetID *string,
	beforeJSON *string,
	afterJSON *string,
	ipAddress *string,
	userAgent *string,
	createdAt time.Time,
) (*BillingAuditLog, error) {
	log := &BillingAuditLog{
		logID:      logID,
		actorType:  actorType,
		actorID:    actorID,
		action:     action,
		targetType: targetType,
		targetID:   targetID,
		beforeJSON: beforeJSON,
		afterJSON:  afterJSON,
		ipAddress:  ipAddress,
		userAgent:  userAgent,
		createdAt:  createdAt,
	}

	if err := log.validate(); err != nil {
		return nil, err
	}

	return log, nil
}

func (l *BillingAuditLog) validate() error {
	if !l.actorType.IsValid() {
		return common.NewValidationError("invalid actor type", nil)
	}
	if l.action == "" {
		return common.NewValidationError("action is required", nil)
	}
	return nil
}

// Getters

func (l *BillingAuditLog) LogID() BillingAuditLogID {
	return l.logID
}

func (l *BillingAuditLog) ActorType() ActorType {
	return l.actorType
}

func (l *BillingAuditLog) ActorID() *string {
	return l.actorID
}

func (l *BillingAuditLog) Action() string {
	return l.action
}

func (l *BillingAuditLog) TargetType() *string {
	return l.targetType
}

func (l *BillingAuditLog) TargetID() *string {
	return l.targetID
}

func (l *BillingAuditLog) BeforeJSON() *string {
	return l.beforeJSON
}

func (l *BillingAuditLog) AfterJSON() *string {
	return l.afterJSON
}

func (l *BillingAuditLog) IPAddress() *string {
	return l.ipAddress
}

func (l *BillingAuditLog) UserAgent() *string {
	return l.userAgent
}

func (l *BillingAuditLog) CreatedAt() time.Time {
	return l.createdAt
}
