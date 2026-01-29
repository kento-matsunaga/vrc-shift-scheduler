package calendar

import (
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
)

// CalendarEntry represents a free-form entry in a calendar
type CalendarEntry struct {
	entryID    common.CalendarEntryID
	calendarID common.CalendarID
	tenantID   common.TenantID
	title      string
	date       time.Time
	startTime  *time.Time
	endTime    *time.Time
	note       string
	createdAt  time.Time
	updatedAt  time.Time
}

// NewCalendarEntry creates a new CalendarEntry
func NewCalendarEntry(
	now time.Time,
	calendarID common.CalendarID,
	tenantID common.TenantID,
	title string,
	date time.Time,
	startTime *time.Time,
	endTime *time.Time,
	note string,
) (*CalendarEntry, error) {
	entry := &CalendarEntry{
		entryID:    common.NewCalendarEntryIDWithTime(now),
		calendarID: calendarID,
		tenantID:   tenantID,
		title:      title,
		date:       date,
		startTime:  startTime,
		endTime:    endTime,
		note:       note,
		createdAt:  now,
		updatedAt:  now,
	}

	if err := entry.validate(); err != nil {
		return nil, err
	}

	return entry, nil
}

// ReconstructCalendarEntry recreates a CalendarEntry from stored data
func ReconstructCalendarEntry(
	entryID common.CalendarEntryID,
	calendarID common.CalendarID,
	tenantID common.TenantID,
	title string,
	date time.Time,
	startTime *time.Time,
	endTime *time.Time,
	note string,
	createdAt time.Time,
	updatedAt time.Time,
) (*CalendarEntry, error) {
	entry := &CalendarEntry{
		entryID:    entryID,
		calendarID: calendarID,
		tenantID:   tenantID,
		title:      title,
		date:       date,
		startTime:  startTime,
		endTime:    endTime,
		note:       note,
		createdAt:  createdAt,
		updatedAt:  updatedAt,
	}

	if err := entry.validate(); err != nil {
		return nil, err
	}

	return entry, nil
}

// validate checks invariants
func (e *CalendarEntry) validate() error {
	if e.title == "" {
		return common.NewValidationError("title is required", nil)
	}
	if len(e.title) > 255 {
		return common.NewValidationError("title must be 255 characters or less", nil)
	}
	return nil
}

// Getters

func (e *CalendarEntry) EntryID() common.CalendarEntryID {
	return e.entryID
}

func (e *CalendarEntry) CalendarID() common.CalendarID {
	return e.calendarID
}

func (e *CalendarEntry) TenantID() common.TenantID {
	return e.tenantID
}

func (e *CalendarEntry) Title() string {
	return e.title
}

func (e *CalendarEntry) Date() time.Time {
	return e.date
}

func (e *CalendarEntry) StartTime() *time.Time {
	return e.startTime
}

func (e *CalendarEntry) EndTime() *time.Time {
	return e.endTime
}

func (e *CalendarEntry) Note() string {
	return e.note
}

func (e *CalendarEntry) CreatedAt() time.Time {
	return e.createdAt
}

func (e *CalendarEntry) UpdatedAt() time.Time {
	return e.updatedAt
}

// Update updates the entry fields
func (e *CalendarEntry) Update(now time.Time, title string, date time.Time, startTime *time.Time, endTime *time.Time, note string) error {
	if title == "" {
		return common.NewValidationError("title is required", nil)
	}
	if len(title) > 255 {
		return common.NewValidationError("title must be 255 characters or less", nil)
	}

	e.title = title
	e.date = date
	e.startTime = startTime
	e.endTime = endTime
	e.note = note
	e.updatedAt = now
	return nil
}
