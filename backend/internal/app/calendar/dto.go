package calendar

import "time"

// === Input DTOs ===

// CreateCalendarInput represents the input for creating a calendar
type CreateCalendarInput struct {
	TenantID    string
	Title       string
	Description string
	EventIDs    []string
}

// UpdateCalendarInput represents the input for updating a calendar
type UpdateCalendarInput struct {
	TenantID    string
	CalendarID  string
	Title       string
	Description string
	EventIDs    []string
	IsPublic    bool
}

// GetCalendarInput represents the input for getting a calendar
type GetCalendarInput struct {
	TenantID   string
	CalendarID string
}

// DeleteCalendarInput represents the input for deleting a calendar
type DeleteCalendarInput struct {
	TenantID   string
	CalendarID string
}

// GetCalendarByTokenInput represents the input for getting a calendar by public token
type GetCalendarByTokenInput struct {
	Token string
}

// ListCalendarsInput represents the input for listing calendars
type ListCalendarsInput struct {
	TenantID string
}

// === Output DTOs ===

// CalendarOutput represents the output for a calendar
type CalendarOutput struct {
	ID          string    `json:"id"`
	TenantID    string    `json:"tenant_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	IsPublic    bool      `json:"is_public"`
	PublicToken *string   `json:"public_token,omitempty"`
	EventIDs    []string  `json:"event_ids"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CalendarWithEventsOutput represents a calendar with its events and entries
type CalendarWithEventsOutput struct {
	Calendar CalendarOutput     `json:"calendar"`
	Events   []EventOutput      `json:"events"`
	Entries  []CalendarEntryDTO `json:"entries"`
}

// EventOutput represents event information for calendar display
type EventOutput struct {
	ID           string              `json:"id"`
	Title        string              `json:"title"`
	Description  string              `json:"description"`
	BusinessDays []BusinessDayOutput `json:"business_days"`
}

// BusinessDayOutput represents a business day for calendar display
type BusinessDayOutput struct {
	ID        string    `json:"id"`
	Date      time.Time `json:"date"`
	StartTime string    `json:"start_time"`
	EndTime   string    `json:"end_time"`
}
