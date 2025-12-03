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

