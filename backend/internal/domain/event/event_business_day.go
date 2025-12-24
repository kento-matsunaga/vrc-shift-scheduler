package event

import (
	"fmt"
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
)

// OccurrenceType represents the occurrence type of a business day
type OccurrenceType string

const (
	OccurrenceTypeRecurring OccurrenceType = "recurring" // 定期営業
	OccurrenceTypeSpecial   OccurrenceType = "special"   // 特別営業
)

func (t OccurrenceType) Validate() error {
	switch t {
	case OccurrenceTypeRecurring, OccurrenceTypeSpecial:
		return nil
	default:
		return common.NewValidationError(fmt.Sprintf("invalid occurrence type: %s", t), nil)
	}
}

// BusinessDayID represents a business day identifier
type BusinessDayID string

// NewBusinessDayIDWithTime creates a new BusinessDayID using the provided time.
func NewBusinessDayIDWithTime(t time.Time) BusinessDayID {
	return BusinessDayID(common.NewULIDWithTime(t))
}

// NewBusinessDayID creates a new BusinessDayID using the current time.
// Deprecated: Use NewBusinessDayIDWithTime for better testability.
func NewBusinessDayID() BusinessDayID {
	return BusinessDayID(common.NewULID())
}

func (id BusinessDayID) String() string {
	return string(id)
}

func (id BusinessDayID) Validate() error {
	if id == "" {
		return common.NewValidationError("business_day_id is required", nil)
	}
	return common.ValidateULID(string(id))
}

func ParseBusinessDayID(s string) (BusinessDayID, error) {
	if err := common.ValidateULID(s); err != nil {
		return "", err
	}
	return BusinessDayID(s), nil
}

// EventBusinessDay represents an event business day entity
// Event とは独立したエンティティ（Event集約には含まれない）
// Event は「営業の定義」、EventBusinessDay は「生成されたインスタンス」
type EventBusinessDay struct {
	businessDayID       BusinessDayID
	tenantID            common.TenantID
	eventID             common.EventID
	targetDate          time.Time // DATE型として扱う
	startTime           time.Time // TIME型として扱う（HH:MM:SS）
	endTime             time.Time // TIME型として扱う（HH:MM:SS）
	occurrenceType      OccurrenceType
	recurringPatternID  *common.EventID // recurring の場合のみ
	isActive            bool
	validFrom           *time.Time // DATE型として扱う
	validTo             *time.Time // DATE型として扱う
	createdAt           time.Time
	updatedAt           time.Time
	deletedAt           *time.Time
}

// NewEventBusinessDay creates a new EventBusinessDay entity
func NewEventBusinessDay(
	now time.Time,
	tenantID common.TenantID,
	eventID common.EventID,
	targetDate time.Time,
	startTime time.Time,
	endTime time.Time,
	occurrenceType OccurrenceType,
	recurringPatternID *common.EventID,
) (*EventBusinessDay, error) {
	businessDay := &EventBusinessDay{
		businessDayID:      NewBusinessDayIDWithTime(now),
		tenantID:           tenantID,
		eventID:            eventID,
		targetDate:         truncateToDate(targetDate),
		startTime:          truncateToTime(startTime),
		endTime:            truncateToTime(endTime),
		occurrenceType:     occurrenceType,
		recurringPatternID: recurringPatternID,
		isActive:           true,
		createdAt:          now,
		updatedAt:          now,
	}

	if err := businessDay.validate(); err != nil {
		return nil, err
	}

	return businessDay, nil
}

// ReconstructEventBusinessDay reconstructs an EventBusinessDay from persistence
func ReconstructEventBusinessDay(
	businessDayID BusinessDayID,
	tenantID common.TenantID,
	eventID common.EventID,
	targetDate time.Time,
	startTime time.Time,
	endTime time.Time,
	occurrenceType OccurrenceType,
	recurringPatternID *common.EventID,
	isActive bool,
	validFrom *time.Time,
	validTo *time.Time,
	createdAt time.Time,
	updatedAt time.Time,
	deletedAt *time.Time,
) (*EventBusinessDay, error) {
	businessDay := &EventBusinessDay{
		businessDayID:      businessDayID,
		tenantID:           tenantID,
		eventID:            eventID,
		targetDate:         truncateToDate(targetDate),
		startTime:          truncateToTime(startTime),
		endTime:            truncateToTime(endTime),
		occurrenceType:     occurrenceType,
		recurringPatternID: recurringPatternID,
		isActive:           isActive,
		validFrom:          validFrom,
		validTo:            validTo,
		createdAt:          createdAt,
		updatedAt:          updatedAt,
		deletedAt:          deletedAt,
	}

	if err := businessDay.validate(); err != nil {
		return nil, err
	}

	return businessDay, nil
}

func (b *EventBusinessDay) validate() error {
	// TenantID の必須性チェック
	if err := b.tenantID.Validate(); err != nil {
		return common.NewValidationError("tenant_id is required", err)
	}

	// EventID の必須性チェック
	if err := b.eventID.Validate(); err != nil {
		return common.NewValidationError("event_id is required", err)
	}

	// OccurrenceType のバリデーション
	if err := b.occurrenceType.Validate(); err != nil {
		return common.NewValidationError("invalid occurrence_type", err)
	}

	// special の場合は recurringPatternID は NULL である必要がある
	// recurring の場合は recurringPatternID は任意（イベント自体が定期情報を持っている場合はnilでも可）
	if b.occurrenceType == OccurrenceTypeSpecial {
		if b.recurringPatternID != nil {
			return common.NewValidationError("recurring_pattern_id must be null for special occurrence", nil)
		}
	}

	// valid_from と valid_to の整合性チェック
	if b.validFrom != nil && b.validTo != nil {
		if b.validFrom.After(*b.validTo) {
			return common.NewValidationError("valid_from must be before valid_to", nil)
		}
	} else if (b.validFrom != nil && b.validTo == nil) || (b.validFrom == nil && b.validTo != nil) {
		return common.NewValidationError("valid_from and valid_to must be both set or both null", nil)
	}

	// 時刻の前後関係チェック（深夜営業対応）
	// start_time < end_time OR end_time < start_time のどちらかを満たす必要がある
	// （深夜営業の場合、end_time が start_time より前になる）
	// このチェックは常に true になるため、省略可能だが、明示的に記述

	return nil
}

// truncateToDate truncates a time to date only (YYYY-MM-DD 00:00:00)
func truncateToDate(t time.Time) time.Time {
	year, month, day := t.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, t.Location())
}

// truncateToTime truncates a time to time only (HH:MM:SS)
// For TIME WITHOUT TIME ZONE, we use a fixed date (2000-01-01)
func truncateToTime(t time.Time) time.Time {
	hour, min, sec := t.Clock()
	return time.Date(2000, 1, 1, hour, min, sec, 0, time.UTC)
}

// Getters

func (b *EventBusinessDay) BusinessDayID() BusinessDayID {
	return b.businessDayID
}

func (b *EventBusinessDay) TenantID() common.TenantID {
	return b.tenantID
}

func (b *EventBusinessDay) EventID() common.EventID {
	return b.eventID
}

func (b *EventBusinessDay) TargetDate() time.Time {
	return b.targetDate
}

func (b *EventBusinessDay) StartTime() time.Time {
	return b.startTime
}

func (b *EventBusinessDay) EndTime() time.Time {
	return b.endTime
}

func (b *EventBusinessDay) OccurrenceType() OccurrenceType {
	return b.occurrenceType
}

func (b *EventBusinessDay) RecurringPatternID() *common.EventID {
	return b.recurringPatternID
}

func (b *EventBusinessDay) IsActive() bool {
	return b.isActive
}

func (b *EventBusinessDay) ValidFrom() *time.Time {
	return b.validFrom
}

func (b *EventBusinessDay) ValidTo() *time.Time {
	return b.validTo
}

func (b *EventBusinessDay) CreatedAt() time.Time {
	return b.createdAt
}

func (b *EventBusinessDay) UpdatedAt() time.Time {
	return b.updatedAt
}

func (b *EventBusinessDay) DeletedAt() *time.Time {
	return b.deletedAt
}

func (b *EventBusinessDay) IsDeleted() bool {
	return b.deletedAt != nil
}

// Activate activates the business day
func (b *EventBusinessDay) Activate() {
	b.isActive = true
	b.updatedAt = time.Now()
}

// Deactivate deactivates the business day
func (b *EventBusinessDay) Deactivate() {
	b.isActive = false
	b.updatedAt = time.Now()
}

// SetValidPeriod sets the valid period for the business day
func (b *EventBusinessDay) SetValidPeriod(validFrom, validTo *time.Time) error {
	if validFrom != nil && validTo != nil {
		if validFrom.After(*validTo) {
			return common.NewValidationError("valid_from must be before valid_to", nil)
		}
	} else if (validFrom != nil && validTo == nil) || (validFrom == nil && validTo != nil) {
		return common.NewValidationError("valid_from and valid_to must be both set or both null", nil)
	}

	b.validFrom = validFrom
	b.validTo = validTo
	b.updatedAt = time.Now()
	return nil
}

// Delete marks the business day as deleted (soft delete)
func (b *EventBusinessDay) Delete() {
	now := time.Now()
	b.deletedAt = &now
	b.updatedAt = now
}

// IsValidOn checks if the business day is valid on the given date
func (b *EventBusinessDay) IsValidOn(date time.Time) bool {
	if !b.isActive {
		return false
	}

	if b.validFrom != nil && b.validTo != nil {
		dateOnly := truncateToDate(date)
		return !dateOnly.Before(*b.validFrom) && !dateOnly.After(*b.validTo)
	}

	return true
}

// DayOfWeek returns the day of week of the target date
func (b *EventBusinessDay) DayOfWeek() time.Weekday {
	return b.targetDate.Weekday()
}

// DayOfWeekString returns the day of week as a string (MON, TUE, etc.)
func (b *EventBusinessDay) DayOfWeekString() DayOfWeek {
	switch b.targetDate.Weekday() {
	case time.Monday:
		return Monday
	case time.Tuesday:
		return Tuesday
	case time.Wednesday:
		return Wednesday
	case time.Thursday:
		return Thursday
	case time.Friday:
		return Friday
	case time.Saturday:
		return Saturday
	case time.Sunday:
		return Sunday
	default:
		return ""
	}
}

