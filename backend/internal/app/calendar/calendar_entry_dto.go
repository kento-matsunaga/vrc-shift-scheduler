package calendar

import (
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/calendar"
)

// CalendarEntryDTO represents the output for a calendar entry
type CalendarEntryDTO struct {
	EntryID    string    `json:"entry_id"`
	CalendarID string    `json:"calendar_id"`
	Title      string    `json:"title"`
	Date       string    `json:"date"`       // YYYY-MM-DD
	StartTime  *string   `json:"start_time"` // HH:MM (nullable)
	EndTime    *string   `json:"end_time"`   // HH:MM (nullable)
	Note       string    `json:"note"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// NewCalendarEntryDTO creates a CalendarEntryDTO from a CalendarEntry entity
func NewCalendarEntryDTO(entry *calendar.CalendarEntry) *CalendarEntryDTO {
	var startTime, endTime *string
	if entry.StartTime() != nil {
		s := entry.StartTime().Format("15:04")
		startTime = &s
	}
	if entry.EndTime() != nil {
		s := entry.EndTime().Format("15:04")
		endTime = &s
	}

	return &CalendarEntryDTO{
		EntryID:    entry.EntryID().String(),
		CalendarID: entry.CalendarID().String(),
		Title:      entry.Title(),
		Date:       entry.Date().Format("2006-01-02"),
		StartTime:  startTime,
		EndTime:    endTime,
		Note:       entry.Note(),
		CreatedAt:  entry.CreatedAt(),
		UpdatedAt:  entry.UpdatedAt(),
	}
}

// NewCalendarEntryDTOList creates a list of CalendarEntryDTO from CalendarEntry entities
func NewCalendarEntryDTOList(entries []*calendar.CalendarEntry) []CalendarEntryDTO {
	dtos := make([]CalendarEntryDTO, 0, len(entries))
	for _, entry := range entries {
		dtos = append(dtos, *NewCalendarEntryDTO(entry))
	}
	return dtos
}
