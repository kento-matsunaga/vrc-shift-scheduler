package common

import (
	"crypto/rand"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/oklog/ulid/v2"
)

// NewULIDWithTime generates a new ULID using the provided time.
// This is the preferred function for testability.
func NewULIDWithTime(t time.Time) string {
	entropy := ulid.Monotonic(rand.Reader, 0)
	return ulid.MustNew(ulid.Timestamp(t), entropy).String()
}

// NewULID generates a new ULID using the current time.
// Deprecated: Use NewULIDWithTime for better testability. This function will be removed in a future version.
func NewULID() string {
	return NewULIDWithTime(time.Now())
}

// ValidateULID validates if a string is a valid ULID
func ValidateULID(id string) error {
	if len(id) != 26 {
		return NewValidationError(fmt.Sprintf("invalid ULID length: expected 26, got %d", len(id)), nil)
	}
	_, err := ulid.Parse(id)
	if err != nil {
		return NewValidationError("invalid ULID format", err)
	}
	return nil
}

// TenantID represents a tenant identifier
type TenantID string

// NewTenantIDWithTime creates a new TenantID using the provided time.
func NewTenantIDWithTime(t time.Time) TenantID {
	return TenantID(NewULIDWithTime(t))
}

// NewTenantID creates a new TenantID using the current time.
// Deprecated: Use NewTenantIDWithTime for better testability.
func NewTenantID() TenantID {
	return TenantID(NewULID())
}

func (id TenantID) String() string {
	return string(id)
}

func (id TenantID) Validate() error {
	if id == "" {
		return NewValidationError("tenant_id is required", nil)
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

// NewEventIDWithTime creates a new EventID using the provided time.
func NewEventIDWithTime(t time.Time) EventID {
	return EventID(NewULIDWithTime(t))
}

// NewEventID creates a new EventID using the current time.
// Deprecated: Use NewEventIDWithTime for better testability.
func NewEventID() EventID {
	return EventID(NewULID())
}

func (id EventID) String() string {
	return string(id)
}

func (id EventID) Validate() error {
	if id == "" {
		return NewValidationError("event_id is required", nil)
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

// NewMemberIDWithTime creates a new MemberID using the provided time.
func NewMemberIDWithTime(t time.Time) MemberID {
	return MemberID(NewULIDWithTime(t))
}

// NewMemberID creates a new MemberID using the current time.
// Deprecated: Use NewMemberIDWithTime for better testability.
func NewMemberID() MemberID {
	return MemberID(NewULID())
}

func (id MemberID) String() string {
	return string(id)
}

func (id MemberID) Validate() error {
	if id == "" {
		return NewValidationError("member_id is required", nil)
	}
	return ValidateULID(string(id))
}

func ParseMemberID(s string) (MemberID, error) {
	if err := ValidateULID(s); err != nil {
		return "", err
	}
	return MemberID(s), nil
}

// BusinessDayID represents a business day identifier
type BusinessDayID string

// NewBusinessDayIDWithTime creates a new BusinessDayID using the provided time.
func NewBusinessDayIDWithTime(t time.Time) BusinessDayID {
	return BusinessDayID(NewULIDWithTime(t))
}

// NewBusinessDayID creates a new BusinessDayID using the current time.
// Deprecated: Use NewBusinessDayIDWithTime for better testability.
func NewBusinessDayID() BusinessDayID {
	return BusinessDayID(NewULID())
}

func (id BusinessDayID) String() string {
	return string(id)
}

func (id BusinessDayID) Validate() error {
	if id == "" {
		return NewValidationError("business_day_id is required", nil)
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

// NewAssignmentIDWithTime creates a new AssignmentID using the provided time.
func NewAssignmentIDWithTime(t time.Time) AssignmentID {
	return AssignmentID(NewULIDWithTime(t))
}

// NewAssignmentID creates a new AssignmentID using the current time.
// Deprecated: Use NewAssignmentIDWithTime for better testability.
func NewAssignmentID() AssignmentID {
	return AssignmentID(NewULID())
}

func (id AssignmentID) String() string {
	return string(id)
}

func (id AssignmentID) Validate() error {
	if id == "" {
		return NewValidationError("assignment_id is required", nil)
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
		return NewValidationError("public_token is required", nil)
	}
	return ValidatePublicToken(string(t))
}

// ValidatePublicToken validates if a string is a valid UUID v4 format
func ValidatePublicToken(token string) error {
	if token == "" {
		return NewValidationError("public_token is required", nil)
	}
	_, err := uuid.Parse(token)
	if err != nil {
		return NewValidationError(fmt.Sprintf("invalid public_token format: must be UUID v4, got: %s", token), err)
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

// NewAdminIDWithTime creates a new AdminID using the provided time.
func NewAdminIDWithTime(t time.Time) AdminID {
	return AdminID(NewULIDWithTime(t))
}

// NewAdminID creates a new AdminID using the current time.
// Deprecated: Use NewAdminIDWithTime for better testability.
func NewAdminID() AdminID {
	return AdminID(NewULID())
}

func (id AdminID) String() string {
	return string(id)
}

func (id AdminID) Validate() error {
	if id == "" {
		return NewValidationError("admin_id is required", nil)
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

// NewCollectionIDWithTime creates a new CollectionID using the provided time.
func NewCollectionIDWithTime(t time.Time) CollectionID {
	return CollectionID(NewULIDWithTime(t))
}

// NewCollectionID creates a new CollectionID using the current time.
// Deprecated: Use NewCollectionIDWithTime for better testability.
func NewCollectionID() CollectionID {
	return CollectionID(NewULID())
}

func (id CollectionID) String() string {
	return string(id)
}

func (id CollectionID) Validate() error {
	if id == "" {
		return NewValidationError("collection_id is required", nil)
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

// NewResponseIDWithTime creates a new ResponseID using the provided time.
func NewResponseIDWithTime(t time.Time) ResponseID {
	return ResponseID(NewULIDWithTime(t))
}

// NewResponseID creates a new ResponseID using the current time.
// Deprecated: Use NewResponseIDWithTime for better testability.
func NewResponseID() ResponseID {
	return ResponseID(NewULID())
}

func (id ResponseID) String() string {
	return string(id)
}

func (id ResponseID) Validate() error {
	if id == "" {
		return NewValidationError("response_id is required", nil)
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

// NewScheduleIDWithTime creates a new ScheduleID using the provided time.
func NewScheduleIDWithTime(t time.Time) ScheduleID {
	return ScheduleID(NewULIDWithTime(t))
}

// NewScheduleID creates a new ScheduleID using the current time.
// Deprecated: Use NewScheduleIDWithTime for better testability.
func NewScheduleID() ScheduleID {
	return ScheduleID(NewULID())
}

func (id ScheduleID) String() string {
	return string(id)
}

func (id ScheduleID) Validate() error {
	if id == "" {
		return NewValidationError("schedule_id is required", nil)
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

// NewCandidateIDWithTime creates a new CandidateID using the provided time.
func NewCandidateIDWithTime(t time.Time) CandidateID {
	return CandidateID(NewULIDWithTime(t))
}

// NewCandidateID creates a new CandidateID using the current time.
// Deprecated: Use NewCandidateIDWithTime for better testability.
func NewCandidateID() CandidateID {
	return CandidateID(NewULID())
}

func (id CandidateID) String() string {
	return string(id)
}

func (id CandidateID) Validate() error {
	if id == "" {
		return NewValidationError("candidate_id is required", nil)
	}
	return ValidateULID(string(id))
}

func ParseCandidateID(s string) (CandidateID, error) {
	if err := ValidateULID(s); err != nil {
		return "", err
	}
	return CandidateID(s), nil
}

// TargetDateID represents an attendance target date identifier
type TargetDateID string

// NewTargetDateIDWithTime creates a new TargetDateID using the provided time.
func NewTargetDateIDWithTime(t time.Time) TargetDateID {
	return TargetDateID(NewULIDWithTime(t))
}

// NewTargetDateID creates a new TargetDateID using the current time.
// Deprecated: Use NewTargetDateIDWithTime for better testability.
func NewTargetDateID() TargetDateID {
	return TargetDateID(NewULID())
}

func (id TargetDateID) String() string {
	return string(id)
}

func (id TargetDateID) Validate() error {
	return ValidateULID(string(id))
}

func ParseTargetDateID(s string) (TargetDateID, error) {
	if err := ValidateULID(s); err != nil {
		return "", err
	}
	return TargetDateID(s), nil
}

// RoleID represents a role identifier
type RoleID string

// NewRoleIDWithTime creates a new RoleID using the provided time.
func NewRoleIDWithTime(t time.Time) RoleID {
	return RoleID(NewULIDWithTime(t))
}

// NewRoleID creates a new RoleID using the current time.
// Deprecated: Use NewRoleIDWithTime for better testability.
func NewRoleID() RoleID {
	return RoleID(NewULID())
}

func (id RoleID) String() string {
	return string(id)
}

func (id RoleID) Validate() error {
	return ValidateULID(string(id))
}

func ParseRoleID(s string) (RoleID, error) {
	if err := ValidateULID(s); err != nil {
		return "", err
	}
	return RoleID(s), nil
}

// ShiftSlotTemplateID represents a shift slot template identifier
type ShiftSlotTemplateID string

// NewShiftSlotTemplateIDWithTime creates a new ShiftSlotTemplateID using the provided time.
func NewShiftSlotTemplateIDWithTime(t time.Time) ShiftSlotTemplateID {
	return ShiftSlotTemplateID(NewULIDWithTime(t))
}

// NewShiftSlotTemplateID creates a new ShiftSlotTemplateID using the current time.
// Deprecated: Use NewShiftSlotTemplateIDWithTime for better testability.
func NewShiftSlotTemplateID() ShiftSlotTemplateID {
	return ShiftSlotTemplateID(NewULID())
}

func (id ShiftSlotTemplateID) String() string {
	return string(id)
}

func (id ShiftSlotTemplateID) Validate() error {
	if id == "" {
		return NewValidationError("template_id is required", nil)
	}
	return ValidateULID(string(id))
}

func ParseShiftSlotTemplateID(s string) (ShiftSlotTemplateID, error) {
	if err := ValidateULID(s); err != nil {
		return "", err
	}
	return ShiftSlotTemplateID(s), nil
}

// ShiftSlotTemplateItemID represents a shift slot template item identifier
type ShiftSlotTemplateItemID string

// NewShiftSlotTemplateItemIDWithTime creates a new ShiftSlotTemplateItemID using the provided time.
func NewShiftSlotTemplateItemIDWithTime(t time.Time) ShiftSlotTemplateItemID {
	return ShiftSlotTemplateItemID(NewULIDWithTime(t))
}

// NewShiftSlotTemplateItemID creates a new ShiftSlotTemplateItemID using the current time.
// Deprecated: Use NewShiftSlotTemplateItemIDWithTime for better testability.
func NewShiftSlotTemplateItemID() ShiftSlotTemplateItemID {
	return ShiftSlotTemplateItemID(NewULID())
}

func (id ShiftSlotTemplateItemID) String() string {
	return string(id)
}

func (id ShiftSlotTemplateItemID) Validate() error {
	if id == "" {
		return NewValidationError("item_id is required", nil)
	}
	return ValidateULID(string(id))
}

func ParseShiftSlotTemplateItemID(s string) (ShiftSlotTemplateItemID, error) {
	if err := ValidateULID(s); err != nil {
		return "", err
	}
	return ShiftSlotTemplateItemID(s), nil
}

// MemberGroupID represents a member group identifier
type MemberGroupID string

// NewMemberGroupIDWithTime creates a new MemberGroupID using the provided time.
func NewMemberGroupIDWithTime(t time.Time) MemberGroupID {
	return MemberGroupID(NewULIDWithTime(t))
}

// NewMemberGroupID creates a new MemberGroupID using the current time.
// Deprecated: Use NewMemberGroupIDWithTime for better testability.
func NewMemberGroupID() MemberGroupID {
	return MemberGroupID(NewULID())
}

func (id MemberGroupID) String() string {
	return string(id)
}

func (id MemberGroupID) Validate() error {
	if id == "" {
		return NewValidationError("group_id is required", nil)
	}
	return ValidateULID(string(id))
}

func ParseMemberGroupID(s string) (MemberGroupID, error) {
	if err := ValidateULID(s); err != nil {
		return "", err
	}
	return MemberGroupID(s), nil
}

// RoleGroupID represents a role group identifier
type RoleGroupID string

// NewRoleGroupIDWithTime creates a new RoleGroupID using the provided time.
func NewRoleGroupIDWithTime(t time.Time) RoleGroupID {
	return RoleGroupID(NewULIDWithTime(t))
}

// NewRoleGroupID creates a new RoleGroupID using the current time.
// Deprecated: Use NewRoleGroupIDWithTime for better testability.
func NewRoleGroupID() RoleGroupID {
	return RoleGroupID(NewULID())
}

func (id RoleGroupID) String() string {
	return string(id)
}

func (id RoleGroupID) Validate() error {
	if id == "" {
		return NewValidationError("group_id is required", nil)
	}
	return ValidateULID(string(id))
}

func ParseRoleGroupID(s string) (RoleGroupID, error) {
	if err := ValidateULID(s); err != nil {
		return "", err
	}
	return RoleGroupID(s), nil
}

// ImportJobID represents an import job identifier
type ImportJobID string

// NewImportJobIDWithTime creates a new ImportJobID using the provided time.
func NewImportJobIDWithTime(t time.Time) ImportJobID {
	return ImportJobID(NewULIDWithTime(t))
}

// NewImportJobID creates a new ImportJobID using the current time.
// Deprecated: Use NewImportJobIDWithTime for better testability.
func NewImportJobID() ImportJobID {
	return ImportJobID(NewULID())
}

func (id ImportJobID) String() string {
	return string(id)
}

func (id ImportJobID) Validate() error {
	if id == "" {
		return NewValidationError("import_job_id is required", nil)
	}
	return ValidateULID(string(id))
}

func ParseImportJobID(s string) (ImportJobID, error) {
	if err := ValidateULID(s); err != nil {
		return "", err
	}
	return ImportJobID(s), nil
}

// ImportLogID represents an import log identifier
type ImportLogID string

// NewImportLogIDWithTime creates a new ImportLogID using the provided time.
func NewImportLogIDWithTime(t time.Time) ImportLogID {
	return ImportLogID(NewULIDWithTime(t))
}

// NewImportLogID creates a new ImportLogID using the current time.
// Deprecated: Use NewImportLogIDWithTime for better testability.
func NewImportLogID() ImportLogID {
	return ImportLogID(NewULID())
}

func (id ImportLogID) String() string {
	return string(id)
}

func (id ImportLogID) Validate() error {
	if id == "" {
		return NewValidationError("log_id is required", nil)
	}
	return ValidateULID(string(id))
}

func ParseImportLogID(s string) (ImportLogID, error) {
	if err := ValidateULID(s); err != nil {
		return "", err
	}
	return ImportLogID(s), nil
}
