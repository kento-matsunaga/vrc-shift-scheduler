package common

import (
	"crypto/rand"
	"fmt"
	"time"

	"github.com/google/uuid"
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

// PublicToken represents a public access token for attendance collections and date schedules.
// It uses UUID v4 format (RFC 4122) for security and standardization.
type PublicToken string

// NewPublicToken generates a new UUID v4 public token
func NewPublicToken() PublicToken {
	return PublicToken(uuid.New().String())
}

func (t PublicToken) String() string {
	return string(t)
}

// Validate checks if the token is a valid UUID v4 format
func (t PublicToken) Validate() error {
	if t == "" {
		return fmt.Errorf("public_token is required")
	}
	return ValidatePublicToken(string(t))
}

// ValidatePublicToken validates if a string is a valid UUID v4 format
func ValidatePublicToken(token string) error {
	if token == "" {
		return fmt.Errorf("public_token is required")
	}
	_, err := uuid.Parse(token)
	if err != nil {
		return fmt.Errorf("invalid public_token format: must be UUID v4, got: %s", token)
	}
	return nil
}

// ParsePublicToken parses a string into a PublicToken after validation
func ParsePublicToken(s string) (PublicToken, error) {
	if err := ValidatePublicToken(s); err != nil {
		return "", err
	}
	return PublicToken(s), nil
}

// AdminID represents an admin identifier
type AdminID string

func NewAdminID() AdminID {
	return AdminID(NewULID())
}

func (id AdminID) String() string {
	return string(id)
}

func (id AdminID) Validate() error {
	if id == "" {
		return fmt.Errorf("admin_id is required")
	}
	return ValidateULID(string(id))
}

func ParseAdminID(s string) (AdminID, error) {
	if err := ValidateULID(s); err != nil {
		return "", err
	}
	return AdminID(s), nil
}

// CollectionID represents an attendance collection identifier
type CollectionID string

func NewCollectionID() CollectionID {
	return CollectionID(NewULID())
}

func (id CollectionID) String() string {
	return string(id)
}

func (id CollectionID) Validate() error {
	if id == "" {
		return fmt.Errorf("collection_id is required")
	}
	return ValidateULID(string(id))
}

func ParseCollectionID(s string) (CollectionID, error) {
	if err := ValidateULID(s); err != nil {
		return "", err
	}
	return CollectionID(s), nil
}

// ResponseID represents an attendance response identifier
type ResponseID string

func NewResponseID() ResponseID {
	return ResponseID(NewULID())
}

func (id ResponseID) String() string {
	return string(id)
}

func (id ResponseID) Validate() error {
	if id == "" {
		return fmt.Errorf("response_id is required")
	}
	return ValidateULID(string(id))
}

func ParseResponseID(s string) (ResponseID, error) {
	if err := ValidateULID(s); err != nil {
		return "", err
	}
	return ResponseID(s), nil
}

// ScheduleID represents a date schedule identifier
type ScheduleID string

func NewScheduleID() ScheduleID {
	return ScheduleID(NewULID())
}

func (id ScheduleID) String() string {
	return string(id)
}

func (id ScheduleID) Validate() error {
	if id == "" {
		return fmt.Errorf("schedule_id is required")
	}
	return ValidateULID(string(id))
}

func ParseScheduleID(s string) (ScheduleID, error) {
	if err := ValidateULID(s); err != nil {
		return "", err
	}
	return ScheduleID(s), nil
}

// CandidateID represents a schedule candidate identifier
type CandidateID string

func NewCandidateID() CandidateID {
	return CandidateID(NewULID())
}

func (id CandidateID) String() string {
	return string(id)
}

func (id CandidateID) Validate() error {
	if id == "" {
		return fmt.Errorf("candidate_id is required")
	}
	return ValidateULID(string(id))
}

func ParseCandidateID(s string) (CandidateID, error) {
	if err := ValidateULID(s); err != nil {
		return "", err
	}
	return CandidateID(s), nil
}
