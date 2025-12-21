package billing

import (
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
)

// EntitlementID represents the unique identifier for an entitlement
type EntitlementID string

// NewEntitlementID generates a new EntitlementID
func NewEntitlementID() EntitlementID {
	return EntitlementID(common.NewULID())
}

// ParseEntitlementID parses a string into EntitlementID
func ParseEntitlementID(id string) (EntitlementID, error) {
	if err := common.ValidateULID(id); err != nil {
		return "", err
	}
	return EntitlementID(id), nil
}

// String returns the string representation
func (id EntitlementID) String() string {
	return string(id)
}

// EntitlementSource represents the source of the entitlement
type EntitlementSource string

const (
	EntitlementSourceBooth  EntitlementSource = "booth"
	EntitlementSourceStripe EntitlementSource = "stripe"
	EntitlementSourceManual EntitlementSource = "manual"
)

// String returns the string representation
func (es EntitlementSource) String() string {
	return string(es)
}

// IsValid checks if the source is valid
func (es EntitlementSource) IsValid() bool {
	return es == EntitlementSourceBooth || es == EntitlementSourceStripe || es == EntitlementSourceManual
}

// Entitlement represents a tenant's plan entitlement
type Entitlement struct {
	entitlementID EntitlementID
	tenantID      common.TenantID
	planCode      string
	source        EntitlementSource
	startsAt      time.Time
	endsAt        *time.Time
	revokedAt     *time.Time
	revokedReason *string
	createdAt     time.Time
	updatedAt     time.Time
}

// NewEntitlement creates a new Entitlement entity
func NewEntitlement(
	now time.Time,
	tenantID common.TenantID,
	planCode string,
	source EntitlementSource,
	endsAt *time.Time,
) (*Entitlement, error) {
	entitlement := &Entitlement{
		entitlementID: NewEntitlementID(),
		tenantID:      tenantID,
		planCode:      planCode,
		source:        source,
		startsAt:      now,
		endsAt:        endsAt,
		revokedAt:     nil,
		revokedReason: nil,
		createdAt:     now,
		updatedAt:     now,
	}

	if err := entitlement.validate(); err != nil {
		return nil, err
	}

	return entitlement, nil
}

// ReconstructEntitlement reconstructs an Entitlement entity from persistence
func ReconstructEntitlement(
	entitlementID EntitlementID,
	tenantID common.TenantID,
	planCode string,
	source EntitlementSource,
	startsAt time.Time,
	endsAt *time.Time,
	revokedAt *time.Time,
	revokedReason *string,
	createdAt time.Time,
	updatedAt time.Time,
) (*Entitlement, error) {
	entitlement := &Entitlement{
		entitlementID: entitlementID,
		tenantID:      tenantID,
		planCode:      planCode,
		source:        source,
		startsAt:      startsAt,
		endsAt:        endsAt,
		revokedAt:     revokedAt,
		revokedReason: revokedReason,
		createdAt:     createdAt,
		updatedAt:     updatedAt,
	}

	if err := entitlement.validate(); err != nil {
		return nil, err
	}

	return entitlement, nil
}

func (e *Entitlement) validate() error {
	if e.planCode == "" {
		return common.NewValidationError("plan_code is required", nil)
	}
	if !e.source.IsValid() {
		return common.NewValidationError("invalid entitlement source", nil)
	}
	return nil
}

// Getters

func (e *Entitlement) EntitlementID() EntitlementID {
	return e.entitlementID
}

func (e *Entitlement) TenantID() common.TenantID {
	return e.tenantID
}

func (e *Entitlement) PlanCode() string {
	return e.planCode
}

func (e *Entitlement) Source() EntitlementSource {
	return e.source
}

func (e *Entitlement) StartsAt() time.Time {
	return e.startsAt
}

func (e *Entitlement) EndsAt() *time.Time {
	return e.endsAt
}

func (e *Entitlement) RevokedAt() *time.Time {
	return e.revokedAt
}

func (e *Entitlement) RevokedReason() *string {
	return e.revokedReason
}

func (e *Entitlement) CreatedAt() time.Time {
	return e.createdAt
}

func (e *Entitlement) UpdatedAt() time.Time {
	return e.updatedAt
}

// IsRevoked checks if the entitlement has been revoked
func (e *Entitlement) IsRevoked() bool {
	return e.revokedAt != nil
}

// IsLifetime checks if this is a lifetime entitlement
func (e *Entitlement) IsLifetime() bool {
	return e.endsAt == nil
}

// IsActive checks if the entitlement is currently active
func (e *Entitlement) IsActive(now time.Time) bool {
	if e.IsRevoked() {
		return false
	}
	if e.startsAt.After(now) {
		return false
	}
	if e.endsAt != nil && e.endsAt.Before(now) {
		return false
	}
	return true
}

// Revoke revokes the entitlement
func (e *Entitlement) Revoke(now time.Time, reason string) {
	e.revokedAt = &now
	e.revokedReason = &reason
	e.updatedAt = now
}

// ExtendEndsAt extends the entitlement end date
func (e *Entitlement) ExtendEndsAt(now time.Time, newEndsAt time.Time) {
	e.endsAt = &newEndsAt
	e.updatedAt = now
}
