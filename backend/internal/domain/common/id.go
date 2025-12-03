package common

import (
	"crypto/rand"
	"fmt"
	"time"

	"github.com/oklog/ulid/v2"
)

// ULID generates a new ULID (Universally Unique Lexicographically Sortable Identifier)
func NewULID() string {
	entropy := ulid.Monotonic(rand.Reader, 0)
	return ulid.MustNew(ulid.Timestamp(time.Now()), entropy).String()
}

// ValidateULID validates if a string is a valid ULID
func ValidateULID(id string) error {
	if len(id) != 26 {
		return fmt.Errorf("invalid ULID length: expected 26, got %d", len(id))
	}
	_, err := ulid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid ULID format: %w", err)
	}
	return nil
}

// TenantID represents a tenant identifier
type TenantID string

func NewTenantID() TenantID {
	return TenantID(NewULID())
}

func (id TenantID) String() string {
	return string(id)
}

func (id TenantID) Validate() error {
	if id == "" {
		return fmt.Errorf("tenant_id is required")
	}
	return ValidateULID(string(id))
}

func ParseTenantID(s string) (TenantID, error) {
	if err := ValidateULID(s); err != nil {
		return "", err
	}
	return TenantID(s), nil
}

// EventID represents an event identifier
type EventID string

func NewEventID() EventID {
	return EventID(NewULID())
}

func (id EventID) String() string {
	return string(id)
}

func (id EventID) Validate() error {
	if id == "" {
		return fmt.Errorf("event_id is required")
	}
	return ValidateULID(string(id))
}

func ParseEventID(s string) (EventID, error) {
	if err := ValidateULID(s); err != nil {
		return "", err
	}
	return EventID(s), nil
}

// MemberID represents a member identifier
type MemberID string

func NewMemberID() MemberID {
	return MemberID(NewULID())
}

func (id MemberID) String() string {
	return string(id)
}

func (id MemberID) Validate() error {
	if id == "" {
		return fmt.Errorf("member_id is required")
	}
	return ValidateULID(string(id))
}

func ParseMemberID(s string) (MemberID, error) {
	if err := ValidateULID(s); err != nil {
		return "", err
	}
	return MemberID(s), nil
}

// PositionID represents a position identifier
type PositionID string

func NewPositionID() PositionID {
	return PositionID(NewULID())
}

func (id PositionID) String() string {
	return string(id)
}

func (id PositionID) Validate() error {
	if id == "" {
		return fmt.Errorf("position_id is required")
	}
	return ValidateULID(string(id))
}

func ParsePositionID(s string) (PositionID, error) {
	if err := ValidateULID(s); err != nil {
		return "", err
	}
	return PositionID(s), nil
}

// BusinessDayID represents a business day identifier
type BusinessDayID string

func NewBusinessDayID() BusinessDayID {
	return BusinessDayID(NewULID())
}

func (id BusinessDayID) String() string {
	return string(id)
}

func (id BusinessDayID) Validate() error {
	if id == "" {
		return fmt.Errorf("business_day_id is required")
	}
	return ValidateULID(string(id))
}

func ParseBusinessDayID(s string) (BusinessDayID, error) {
	if err := ValidateULID(s); err != nil {
		return "", err
	}
	return BusinessDayID(s), nil
}

// AssignmentID represents a shift assignment identifier
type AssignmentID string

func NewAssignmentID() AssignmentID {
	return AssignmentID(NewULID())
}

func (id AssignmentID) String() string {
	return string(id)
}

func (id AssignmentID) Validate() error {
	if id == "" {
		return fmt.Errorf("assignment_id is required")
	}
	return ValidateULID(string(id))
}

func ParseAssignmentID(s string) (AssignmentID, error) {
	if err := ValidateULID(s); err != nil {
		return "", err
	}
	return AssignmentID(s), nil
}

