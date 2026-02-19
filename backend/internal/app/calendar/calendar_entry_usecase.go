package calendar

import (
	"context"
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/calendar"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/services"
)

// === Input/Output DTOs for CalendarEntry ===

// CreateCalendarEntryInput represents the input for creating a calendar entry
type CreateCalendarEntryInput struct {
	TenantID   string
	CalendarID string
	Title      string
	Date       string  // YYYY-MM-DD
	StartTime  *string // HH:MM (nullable)
	EndTime    *string // HH:MM (nullable)
	Note       string
}

// UpdateCalendarEntryInput represents the input for updating a calendar entry
type UpdateCalendarEntryInput struct {
	TenantID  string
	EntryID   string
	Title     string
	Date      string  // YYYY-MM-DD
	StartTime *string // HH:MM (nullable)
	EndTime   *string // HH:MM (nullable)
	Note      string
}

// DeleteCalendarEntryInput represents the input for deleting a calendar entry
type DeleteCalendarEntryInput struct {
	TenantID string
	EntryID  string
}

// ListCalendarEntriesInput represents the input for listing calendar entries
type ListCalendarEntriesInput struct {
	TenantID   string
	CalendarID string
}

// === CreateCalendarEntryUsecase ===

// CreateCalendarEntryUsecase handles creating a calendar entry
type CreateCalendarEntryUsecase struct {
	calendarRepo      calendar.Repository
	calendarEntryRepo calendar.CalendarEntryRepository
	clock             services.Clock
}

// NewCreateCalendarEntryUsecase creates a new CreateCalendarEntryUsecase
func NewCreateCalendarEntryUsecase(
	calendarRepo calendar.Repository,
	calendarEntryRepo calendar.CalendarEntryRepository,
	clock services.Clock,
) *CreateCalendarEntryUsecase {
	return &CreateCalendarEntryUsecase{
		calendarRepo:      calendarRepo,
		calendarEntryRepo: calendarEntryRepo,
		clock:             clock,
	}
}

// Execute creates a new calendar entry
func (u *CreateCalendarEntryUsecase) Execute(ctx context.Context, input CreateCalendarEntryInput) (*CalendarEntryDTO, error) {
	// Parse TenantID
	tenantID, err := common.ParseTenantID(input.TenantID)
	if err != nil {
		return nil, err
	}

	// Parse CalendarID
	calendarID, err := common.ParseCalendarID(input.CalendarID)
	if err != nil {
		return nil, err
	}

	// Verify calendar exists
	_, err = u.calendarRepo.FindByID(ctx, tenantID, calendarID)
	if err != nil {
		return nil, err
	}

	// Parse date
	date, err := time.Parse("2006-01-02", input.Date)
	if err != nil {
		return nil, common.NewValidationError("invalid date format, expected YYYY-MM-DD", err)
	}

	// Parse optional start/end times
	var startTime, endTime *time.Time
	if input.StartTime != nil {
		t, err := time.Parse("15:04", *input.StartTime)
		if err != nil {
			return nil, common.NewValidationError("invalid start_time format, expected HH:MM", err)
		}
		startTime = &t
	}
	if input.EndTime != nil {
		t, err := time.Parse("15:04", *input.EndTime)
		if err != nil {
			return nil, common.NewValidationError("invalid end_time format, expected HH:MM", err)
		}
		endTime = &t
	}

	// Create entry entity
	now := u.clock.Now()
	entry, err := calendar.NewCalendarEntry(now, calendarID, tenantID, input.Title, date, startTime, endTime, input.Note)
	if err != nil {
		return nil, err
	}

	// Save to repository
	if err := u.calendarEntryRepo.Save(ctx, entry); err != nil {
		return nil, err
	}

	return NewCalendarEntryDTO(entry), nil
}

// === UpdateCalendarEntryUsecase ===

// UpdateCalendarEntryUsecase handles updating a calendar entry
type UpdateCalendarEntryUsecase struct {
	calendarEntryRepo calendar.CalendarEntryRepository
	clock             services.Clock
}

// NewUpdateCalendarEntryUsecase creates a new UpdateCalendarEntryUsecase
func NewUpdateCalendarEntryUsecase(
	calendarEntryRepo calendar.CalendarEntryRepository,
	clock services.Clock,
) *UpdateCalendarEntryUsecase {
	return &UpdateCalendarEntryUsecase{
		calendarEntryRepo: calendarEntryRepo,
		clock:             clock,
	}
}

// Execute updates a calendar entry
func (u *UpdateCalendarEntryUsecase) Execute(ctx context.Context, input UpdateCalendarEntryInput) (*CalendarEntryDTO, error) {
	// Parse TenantID
	tenantID, err := common.ParseTenantID(input.TenantID)
	if err != nil {
		return nil, err
	}

	// Parse EntryID
	entryID, err := common.ParseCalendarEntryID(input.EntryID)
	if err != nil {
		return nil, err
	}

	// Find existing entry
	entry, err := u.calendarEntryRepo.FindByID(ctx, tenantID, entryID)
	if err != nil {
		return nil, err
	}

	// Parse date
	date, err := time.Parse("2006-01-02", input.Date)
	if err != nil {
		return nil, common.NewValidationError("invalid date format, expected YYYY-MM-DD", err)
	}

	// Parse optional start/end times
	var startTime, endTime *time.Time
	if input.StartTime != nil {
		t, err := time.Parse("15:04", *input.StartTime)
		if err != nil {
			return nil, common.NewValidationError("invalid start_time format, expected HH:MM", err)
		}
		startTime = &t
	}
	if input.EndTime != nil {
		t, err := time.Parse("15:04", *input.EndTime)
		if err != nil {
			return nil, common.NewValidationError("invalid end_time format, expected HH:MM", err)
		}
		endTime = &t
	}

	// Update entry
	now := u.clock.Now()
	if err := entry.Update(now, input.Title, date, startTime, endTime, input.Note); err != nil {
		return nil, err
	}

	// Save to repository
	if err := u.calendarEntryRepo.Save(ctx, entry); err != nil {
		return nil, err
	}

	return NewCalendarEntryDTO(entry), nil
}

// === DeleteCalendarEntryUsecase ===

// DeleteCalendarEntryUsecase handles deleting a calendar entry
type DeleteCalendarEntryUsecase struct {
	calendarEntryRepo calendar.CalendarEntryRepository
	clock             services.Clock
}

// NewDeleteCalendarEntryUsecase creates a new DeleteCalendarEntryUsecase
func NewDeleteCalendarEntryUsecase(
	calendarEntryRepo calendar.CalendarEntryRepository,
	clock services.Clock,
) *DeleteCalendarEntryUsecase {
	return &DeleteCalendarEntryUsecase{
		calendarEntryRepo: calendarEntryRepo,
		clock:             clock,
	}
}

// Execute soft-deletes a calendar entry
func (u *DeleteCalendarEntryUsecase) Execute(ctx context.Context, input DeleteCalendarEntryInput) error {
	// Parse TenantID
	tenantID, err := common.ParseTenantID(input.TenantID)
	if err != nil {
		return err
	}

	// Parse EntryID
	entryID, err := common.ParseCalendarEntryID(input.EntryID)
	if err != nil {
		return err
	}

	// Find entry
	entry, err := u.calendarEntryRepo.FindByID(ctx, tenantID, entryID)
	if err != nil {
		return err
	}

	// Soft delete via domain method
	entry.Delete(u.clock.Now())

	// Save via repository
	return u.calendarEntryRepo.Save(ctx, entry)
}

// === ListCalendarEntriesUsecase ===

// ListCalendarEntriesUsecase handles listing calendar entries
type ListCalendarEntriesUsecase struct {
	calendarEntryRepo calendar.CalendarEntryRepository
}

// NewListCalendarEntriesUsecase creates a new ListCalendarEntriesUsecase
func NewListCalendarEntriesUsecase(
	calendarEntryRepo calendar.CalendarEntryRepository,
) *ListCalendarEntriesUsecase {
	return &ListCalendarEntriesUsecase{
		calendarEntryRepo: calendarEntryRepo,
	}
}

// Execute lists all entries for a calendar
func (u *ListCalendarEntriesUsecase) Execute(ctx context.Context, input ListCalendarEntriesInput) ([]CalendarEntryDTO, error) {
	// Parse TenantID
	tenantID, err := common.ParseTenantID(input.TenantID)
	if err != nil {
		return nil, err
	}

	// Parse CalendarID
	calendarID, err := common.ParseCalendarID(input.CalendarID)
	if err != nil {
		return nil, err
	}

	// Find entries
	entries, err := u.calendarEntryRepo.FindByCalendarID(ctx, tenantID, calendarID)
	if err != nil {
		return nil, err
	}

	return NewCalendarEntryDTOList(entries), nil
}
