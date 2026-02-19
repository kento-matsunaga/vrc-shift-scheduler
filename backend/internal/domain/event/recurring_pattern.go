package event

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
)

// PatternType represents the type of recurring pattern
type PatternType string

const (
	PatternTypeWeekly      PatternType = "weekly"       // 曜日指定
	PatternTypeMonthlyDate PatternType = "monthly_date" // 月内日付指定
	PatternTypeCustom      PatternType = "custom"       // カスタム
)

func (t PatternType) Validate() error {
	switch t {
	case PatternTypeWeekly, PatternTypeMonthlyDate, PatternTypeCustom:
		return nil
	default:
		return common.NewValidationError(fmt.Sprintf("invalid pattern type: %s", t), nil)
	}
}

// DayOfWeek represents a day of the week
type DayOfWeek string

const (
	Monday    DayOfWeek = "MON"
	Tuesday   DayOfWeek = "TUE"
	Wednesday DayOfWeek = "WED"
	Thursday  DayOfWeek = "THU"
	Friday    DayOfWeek = "FRI"
	Saturday  DayOfWeek = "SAT"
	Sunday    DayOfWeek = "SUN"
)

func (d DayOfWeek) Validate() error {
	switch d {
	case Monday, Tuesday, Wednesday, Thursday, Friday, Saturday, Sunday:
		return nil
	default:
		return common.NewValidationError(fmt.Sprintf("invalid day of week: %s", d), nil)
	}
}

// RecurringPatternConfig represents the configuration for a recurring pattern
type RecurringPatternConfig interface {
	Validate() error
	ToJSON() ([]byte, error)
}

// WeeklyPatternConfig represents a weekly pattern configuration
type WeeklyPatternConfig struct {
	DayOfWeeks []DayOfWeek `json:"day_of_weeks"`
	StartTime  string      `json:"start_time"` // HH:MM format
	EndTime    string      `json:"end_time"`   // HH:MM format
}

func (c *WeeklyPatternConfig) Validate() error {
	if len(c.DayOfWeeks) == 0 {
		return common.NewValidationError("day_of_weeks is required for weekly pattern", nil)
	}
	if len(c.DayOfWeeks) > 7 {
		return common.NewValidationError("day_of_weeks must be 7 or less", nil)
	}

	for _, dow := range c.DayOfWeeks {
		if err := dow.Validate(); err != nil {
			return err
		}
	}

	if c.StartTime == "" || c.EndTime == "" {
		return common.NewValidationError("start_time and end_time are required", nil)
	}

	return nil
}

func (c *WeeklyPatternConfig) ToJSON() ([]byte, error) {
	return json.Marshal(c)
}

// MonthlyDatePatternConfig represents a monthly date pattern configuration
type MonthlyDatePatternConfig struct {
	Dates     []int  `json:"dates"`      // 1-31
	StartTime string `json:"start_time"` // HH:MM format
	EndTime   string `json:"end_time"`   // HH:MM format
}

func (c *MonthlyDatePatternConfig) Validate() error {
	if len(c.Dates) == 0 {
		return common.NewValidationError("dates is required for monthly_date pattern", nil)
	}

	for _, date := range c.Dates {
		if date < 1 || date > 31 {
			return common.NewValidationError(fmt.Sprintf("date must be between 1 and 31, got %d", date), nil)
		}
	}

	if c.StartTime == "" || c.EndTime == "" {
		return common.NewValidationError("start_time and end_time are required", nil)
	}

	return nil
}

func (c *MonthlyDatePatternConfig) ToJSON() ([]byte, error) {
	return json.Marshal(c)
}

// CustomPatternConfig represents a custom pattern configuration
type CustomPatternConfig map[string]interface{}

func (c CustomPatternConfig) Validate() error {
	// カスタムパターンは柔軟性のため、特にバリデーションしない
	return nil
}

func (c CustomPatternConfig) ToJSON() ([]byte, error) {
	return json.Marshal(c)
}

// RecurringPattern represents a recurring pattern entity
// 1 Event につき 1 RecurringPattern を基本とする
type RecurringPattern struct {
	patternID   common.EventID // ULID
	tenantID    common.TenantID
	eventID     common.EventID
	patternType PatternType
	config      RecurringPatternConfig
	createdAt   time.Time
	updatedAt   time.Time
	deletedAt   *time.Time
}

// NewRecurringPattern creates a new RecurringPattern entity
func NewRecurringPattern(
	now time.Time,
	tenantID common.TenantID,
	eventID common.EventID,
	patternType PatternType,
	config RecurringPatternConfig,
) (*RecurringPattern, error) {
	pattern := &RecurringPattern{
		patternID:   common.NewEventID(), // PatternID として ULID を生成
		tenantID:    tenantID,
		eventID:     eventID,
		patternType: patternType,
		config:      config,
		createdAt:   now,
		updatedAt:   now,
	}

	if err := pattern.validate(); err != nil {
		return nil, err
	}

	return pattern, nil
}

// ReconstructRecurringPattern reconstructs a RecurringPattern from persistence
func ReconstructRecurringPattern(
	patternID common.EventID,
	tenantID common.TenantID,
	eventID common.EventID,
	patternType PatternType,
	configJSON []byte,
	createdAt time.Time,
	updatedAt time.Time,
	deletedAt *time.Time,
) (*RecurringPattern, error) {
	var config RecurringPatternConfig
	var err error

	switch patternType {
	case PatternTypeWeekly:
		var weeklyConfig WeeklyPatternConfig
		if err := json.Unmarshal(configJSON, &weeklyConfig); err != nil {
			return nil, common.NewValidationError("failed to unmarshal weekly config", err)
		}
		config = &weeklyConfig
	case PatternTypeMonthlyDate:
		var monthlyConfig MonthlyDatePatternConfig
		if err := json.Unmarshal(configJSON, &monthlyConfig); err != nil {
			return nil, common.NewValidationError("failed to unmarshal monthly_date config", err)
		}
		config = &monthlyConfig
	case PatternTypeCustom:
		var customConfig CustomPatternConfig
		if err := json.Unmarshal(configJSON, &customConfig); err != nil {
			return nil, common.NewValidationError("failed to unmarshal custom config", err)
		}
		config = customConfig
	default:
		return nil, common.NewValidationError(fmt.Sprintf("unknown pattern type: %s", patternType), nil)
	}

	pattern := &RecurringPattern{
		patternID:   patternID,
		tenantID:    tenantID,
		eventID:     eventID,
		patternType: patternType,
		config:      config,
		createdAt:   createdAt,
		updatedAt:   updatedAt,
		deletedAt:   deletedAt,
	}

	if err = pattern.validate(); err != nil {
		return nil, err
	}

	return pattern, nil
}

func (p *RecurringPattern) validate() error {
	if err := p.tenantID.Validate(); err != nil {
		return common.NewValidationError("tenant_id is required", err)
	}

	if err := p.eventID.Validate(); err != nil {
		return common.NewValidationError("event_id is required", err)
	}

	if err := p.patternType.Validate(); err != nil {
		return common.NewValidationError("invalid pattern_type", err)
	}

	if p.config == nil {
		return common.NewValidationError("config is required", nil)
	}

	if err := p.config.Validate(); err != nil {
		return common.NewValidationError("invalid config", err)
	}

	return nil
}

// Getters

func (p *RecurringPattern) PatternID() common.EventID {
	return p.patternID
}

func (p *RecurringPattern) TenantID() common.TenantID {
	return p.tenantID
}

func (p *RecurringPattern) EventID() common.EventID {
	return p.eventID
}

func (p *RecurringPattern) PatternType() PatternType {
	return p.patternType
}

func (p *RecurringPattern) Config() RecurringPatternConfig {
	return p.config
}

func (p *RecurringPattern) ConfigJSON() ([]byte, error) {
	return p.config.ToJSON()
}

func (p *RecurringPattern) CreatedAt() time.Time {
	return p.createdAt
}

func (p *RecurringPattern) UpdatedAt() time.Time {
	return p.updatedAt
}

func (p *RecurringPattern) DeletedAt() *time.Time {
	return p.deletedAt
}

func (p *RecurringPattern) IsDeleted() bool {
	return p.deletedAt != nil
}

// UpdateConfig updates the pattern configuration
func (p *RecurringPattern) UpdateConfig(now time.Time, config RecurringPatternConfig) error {
	if config == nil {
		return common.NewValidationError("config is required", nil)
	}

	if err := config.Validate(); err != nil {
		return common.NewValidationError("invalid config", err)
	}

	p.config = config
	p.updatedAt = now
	return nil
}

// Delete marks the pattern as deleted (soft delete)
func (p *RecurringPattern) Delete(now time.Time) {
	p.deletedAt = &now
	p.updatedAt = now
}
