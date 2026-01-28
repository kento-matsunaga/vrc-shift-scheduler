package calendar

import (
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
)

// Calendar represents a calendar entity (aggregate root)
// カレンダーは複数のイベントをまとめて公開するための集約ルート
type Calendar struct {
	calendarID  common.CalendarID
	tenantID    common.TenantID
	title       string
	description string
	isPublic    bool
	publicToken *common.PublicToken
	eventIDs    []common.EventID
	createdAt   time.Time
	updatedAt   time.Time
}

// NewCalendar creates a new Calendar entity
func NewCalendar(
	now time.Time,
	tenantID common.TenantID,
	title string,
	description string,
	eventIDs []common.EventID,
) (*Calendar, error) {
	calendar := &Calendar{
		calendarID:  common.NewCalendarIDWithTime(now),
		tenantID:    tenantID,
		title:       title,
		description: description,
		isPublic:    false,
		publicToken: nil,
		eventIDs:    eventIDs,
		createdAt:   now,
		updatedAt:   now,
	}

	if err := calendar.validate(); err != nil {
		return nil, err
	}

	return calendar, nil
}

// ReconstructCalendar reconstructs a Calendar entity from persistence
func ReconstructCalendar(
	calendarID common.CalendarID,
	tenantID common.TenantID,
	title string,
	description string,
	isPublic bool,
	publicToken *common.PublicToken,
	eventIDs []common.EventID,
	createdAt time.Time,
	updatedAt time.Time,
) (*Calendar, error) {
	calendar := &Calendar{
		calendarID:  calendarID,
		tenantID:    tenantID,
		title:       title,
		description: description,
		isPublic:    isPublic,
		publicToken: publicToken,
		eventIDs:    eventIDs,
		createdAt:   createdAt,
		updatedAt:   updatedAt,
	}

	if err := calendar.validate(); err != nil {
		return nil, err
	}

	return calendar, nil
}

// validate checks invariants
func (c *Calendar) validate() error {
	if err := c.tenantID.Validate(); err != nil {
		return common.NewValidationError("tenant_id is required", err)
	}

	if c.title == "" {
		return common.NewValidationError("title is required", nil)
	}

	if len(c.title) > 255 {
		return common.NewValidationError("title must be less than 255 characters", nil)
	}

	return nil
}

// Getters

func (c *Calendar) CalendarID() common.CalendarID {
	return c.calendarID
}

func (c *Calendar) TenantID() common.TenantID {
	return c.tenantID
}

func (c *Calendar) Title() string {
	return c.title
}

func (c *Calendar) Description() string {
	return c.description
}

func (c *Calendar) IsPublic() bool {
	return c.isPublic
}

func (c *Calendar) PublicToken() *common.PublicToken {
	return c.publicToken
}

func (c *Calendar) EventIDs() []common.EventID {
	return c.eventIDs
}

func (c *Calendar) CreatedAt() time.Time {
	return c.createdAt
}

func (c *Calendar) UpdatedAt() time.Time {
	return c.updatedAt
}

// MakePublic makes the calendar publicly accessible
func (c *Calendar) MakePublic() {
	if c.publicToken == nil {
		token := common.NewPublicToken()
		c.publicToken = &token
	}
	c.isPublic = true
	c.updatedAt = time.Now()
}

// MakePrivate makes the calendar private
func (c *Calendar) MakePrivate() {
	c.isPublic = false
	c.updatedAt = time.Now()
}

// Update updates the calendar's title, description, and event IDs
func (c *Calendar) Update(title, description string, eventIDs []common.EventID) error {
	if title == "" {
		return common.NewValidationError("title is required", nil)
	}
	if len(title) > 255 {
		return common.NewValidationError("title must be less than 255 characters", nil)
	}

	c.title = title
	c.description = description
	c.eventIDs = eventIDs
	c.updatedAt = time.Now()
	return nil
}
