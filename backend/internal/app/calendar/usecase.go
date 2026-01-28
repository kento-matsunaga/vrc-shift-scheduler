package calendar

import (
	"context"
	"log/slog"
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/calendar"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/event"
)

// CreateCalendarUsecase handles creating a calendar
type CreateCalendarUsecase struct {
	calendarRepo calendar.Repository
	eventRepo    event.EventRepository
}

// NewCreateCalendarUsecase creates a new CreateCalendarUsecase
func NewCreateCalendarUsecase(
	calendarRepo calendar.Repository,
	eventRepo event.EventRepository,
) *CreateCalendarUsecase {
	return &CreateCalendarUsecase{
		calendarRepo: calendarRepo,
		eventRepo:    eventRepo,
	}
}

// Execute creates a new calendar
func (u *CreateCalendarUsecase) Execute(ctx context.Context, input CreateCalendarInput) (*CalendarOutput, error) {
	// Parse TenantID
	tenantID, err := common.ParseTenantID(input.TenantID)
	if err != nil {
		return nil, err
	}

	// Parse EventIDs and validate existence
	eventIDs := make([]common.EventID, 0, len(input.EventIDs))
	for _, eidStr := range input.EventIDs {
		eid, err := common.ParseEventID(eidStr)
		if err != nil {
			return nil, err
		}
		// Validate event exists
		_, err = u.eventRepo.FindByID(ctx, tenantID, eid)
		if err != nil {
			return nil, common.NewNotFoundError("event", eidStr)
		}
		eventIDs = append(eventIDs, eid)
	}

	// Create calendar entity
	now := time.Now()
	cal, err := calendar.NewCalendar(now, tenantID, input.Title, input.Description, eventIDs)
	if err != nil {
		return nil, err
	}

	// Save to repository
	if err := u.calendarRepo.Create(ctx, cal); err != nil {
		return nil, err
	}

	return toCalendarOutput(cal), nil
}

// GetCalendarUsecase handles getting a calendar by ID
type GetCalendarUsecase struct {
	calendarRepo    calendar.Repository
	eventRepo       event.EventRepository
	businessDayRepo event.EventBusinessDayRepository
}

// NewGetCalendarUsecase creates a new GetCalendarUsecase
func NewGetCalendarUsecase(
	calendarRepo calendar.Repository,
	eventRepo event.EventRepository,
	businessDayRepo event.EventBusinessDayRepository,
) *GetCalendarUsecase {
	return &GetCalendarUsecase{
		calendarRepo:    calendarRepo,
		eventRepo:       eventRepo,
		businessDayRepo: businessDayRepo,
	}
}

// Execute gets a calendar by ID with events and business days
func (u *GetCalendarUsecase) Execute(ctx context.Context, input GetCalendarInput) (*CalendarWithEventsOutput, error) {
	// Parse IDs
	tenantID, err := common.ParseTenantID(input.TenantID)
	if err != nil {
		return nil, err
	}
	calendarID, err := common.ParseCalendarID(input.CalendarID)
	if err != nil {
		return nil, err
	}

	// Find calendar
	cal, err := u.calendarRepo.FindByID(ctx, tenantID, calendarID)
	if err != nil {
		return nil, err
	}

	// Get events with business days
	events, err := u.getEventsWithBusinessDays(ctx, tenantID, cal.EventIDs())
	if err != nil {
		return nil, err
	}

	return &CalendarWithEventsOutput{
		Calendar: *toCalendarOutput(cal),
		Events:   events,
	}, nil
}

// getEventsWithBusinessDays fetches events and their business days
func (u *GetCalendarUsecase) getEventsWithBusinessDays(ctx context.Context, tenantID common.TenantID, eventIDs []common.EventID) ([]EventOutput, error) {
	var outputs []EventOutput

	for _, eventID := range eventIDs {
		evt, err := u.eventRepo.FindByID(ctx, tenantID, eventID)
		if err != nil {
			slog.Warn("event not found, skipping", "event_id", eventID.String())
			continue // Skip if event not found
		}

		// Get business days for this event
		businessDays, err := u.businessDayRepo.FindByEventID(ctx, tenantID, eventID)
		if err != nil {
			return nil, err
		}

		var bdOutputs []BusinessDayOutput
		for _, bd := range businessDays {
			bdOutputs = append(bdOutputs, BusinessDayOutput{
				ID:        bd.BusinessDayID().String(),
				Date:      bd.TargetDate(),
				StartTime: bd.StartTime().Format("15:04"),
				EndTime:   bd.EndTime().Format("15:04"),
			})
		}

		outputs = append(outputs, EventOutput{
			ID:           evt.EventID().String(),
			Title:        evt.EventName(),
			Description:  evt.Description(),
			BusinessDays: bdOutputs,
		})
	}

	return outputs, nil
}

// ListCalendarsUsecase handles listing calendars for a tenant
type ListCalendarsUsecase struct {
	calendarRepo calendar.Repository
}

// NewListCalendarsUsecase creates a new ListCalendarsUsecase
func NewListCalendarsUsecase(calendarRepo calendar.Repository) *ListCalendarsUsecase {
	return &ListCalendarsUsecase{
		calendarRepo: calendarRepo,
	}
}

// Execute lists all calendars for a tenant
func (u *ListCalendarsUsecase) Execute(ctx context.Context, input ListCalendarsInput) ([]*CalendarOutput, error) {
	// Parse TenantID
	tenantID, err := common.ParseTenantID(input.TenantID)
	if err != nil {
		return nil, err
	}

	// Find calendars
	calendars, err := u.calendarRepo.FindByTenantID(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	// Convert to output
	outputs := make([]*CalendarOutput, 0, len(calendars))
	for _, cal := range calendars {
		outputs = append(outputs, toCalendarOutput(cal))
	}

	return outputs, nil
}

// GetCalendarByTokenUsecase handles getting a calendar by public token
type GetCalendarByTokenUsecase struct {
	calendarRepo    calendar.Repository
	eventRepo       event.EventRepository
	businessDayRepo event.EventBusinessDayRepository
}

// NewGetCalendarByTokenUsecase creates a new GetCalendarByTokenUsecase
func NewGetCalendarByTokenUsecase(
	calendarRepo calendar.Repository,
	eventRepo event.EventRepository,
	businessDayRepo event.EventBusinessDayRepository,
) *GetCalendarByTokenUsecase {
	return &GetCalendarByTokenUsecase{
		calendarRepo:    calendarRepo,
		eventRepo:       eventRepo,
		businessDayRepo: businessDayRepo,
	}
}

// Execute gets a calendar by public token
func (u *GetCalendarByTokenUsecase) Execute(ctx context.Context, input GetCalendarByTokenInput) (*CalendarWithEventsOutput, error) {
	// Parse token
	token, err := common.ParsePublicToken(input.Token)
	if err != nil {
		return nil, common.NewNotFoundError("calendar", input.Token)
	}

	// Find calendar by token
	cal, err := u.calendarRepo.FindByPublicToken(ctx, token)
	if err != nil {
		return nil, common.NewNotFoundError("calendar", input.Token)
	}

	// Check if calendar is public
	if !cal.IsPublic() {
		return nil, common.NewNotFoundError("calendar", input.Token)
	}

	// Get events with business days
	events, err := u.getEventsWithBusinessDays(ctx, cal.TenantID(), cal.EventIDs())
	if err != nil {
		return nil, err
	}

	return &CalendarWithEventsOutput{
		Calendar: *toCalendarOutput(cal),
		Events:   events,
	}, nil
}

// getEventsWithBusinessDays fetches events and their business days
func (u *GetCalendarByTokenUsecase) getEventsWithBusinessDays(ctx context.Context, tenantID common.TenantID, eventIDs []common.EventID) ([]EventOutput, error) {
	var outputs []EventOutput

	for _, eventID := range eventIDs {
		evt, err := u.eventRepo.FindByID(ctx, tenantID, eventID)
		if err != nil {
			slog.Warn("event not found, skipping", "event_id", eventID.String())
			continue // Skip if event not found
		}

		// Get business days for this event
		businessDays, err := u.businessDayRepo.FindByEventID(ctx, tenantID, eventID)
		if err != nil {
			return nil, err
		}

		var bdOutputs []BusinessDayOutput
		for _, bd := range businessDays {
			bdOutputs = append(bdOutputs, BusinessDayOutput{
				ID:        bd.BusinessDayID().String(),
				Date:      bd.TargetDate(),
				StartTime: bd.StartTime().Format("15:04"),
				EndTime:   bd.EndTime().Format("15:04"),
			})
		}

		outputs = append(outputs, EventOutput{
			ID:           evt.EventID().String(),
			Title:        evt.EventName(),
			Description:  evt.Description(),
			BusinessDays: bdOutputs,
		})
	}

	return outputs, nil
}

// UpdateCalendarUsecase handles updating a calendar
type UpdateCalendarUsecase struct {
	calendarRepo calendar.Repository
	eventRepo    event.EventRepository
}

// NewUpdateCalendarUsecase creates a new UpdateCalendarUsecase
func NewUpdateCalendarUsecase(
	calendarRepo calendar.Repository,
	eventRepo event.EventRepository,
) *UpdateCalendarUsecase {
	return &UpdateCalendarUsecase{
		calendarRepo: calendarRepo,
		eventRepo:    eventRepo,
	}
}

// Execute updates a calendar
func (u *UpdateCalendarUsecase) Execute(ctx context.Context, input UpdateCalendarInput) (*CalendarOutput, error) {
	// Parse IDs
	tenantID, err := common.ParseTenantID(input.TenantID)
	if err != nil {
		return nil, err
	}
	calendarID, err := common.ParseCalendarID(input.CalendarID)
	if err != nil {
		return nil, err
	}

	// Find existing calendar
	cal, err := u.calendarRepo.FindByID(ctx, tenantID, calendarID)
	if err != nil {
		return nil, err
	}

	// Parse and validate EventIDs
	eventIDs := make([]common.EventID, 0, len(input.EventIDs))
	for _, eidStr := range input.EventIDs {
		eid, err := common.ParseEventID(eidStr)
		if err != nil {
			return nil, err
		}
		// Validate event exists
		_, err = u.eventRepo.FindByID(ctx, tenantID, eid)
		if err != nil {
			return nil, common.NewNotFoundError("event", eidStr)
		}
		eventIDs = append(eventIDs, eid)
	}

	// Update calendar
	now := time.Now()
	if err := cal.Update(input.Title, input.Description, eventIDs, now); err != nil {
		return nil, err
	}

	// Handle public/private toggle
	if input.IsPublic && !cal.IsPublic() {
		cal.MakePublic(now)
	} else if !input.IsPublic && cal.IsPublic() {
		cal.MakePrivate(now)
	}

	// Save to repository
	if err := u.calendarRepo.Update(ctx, cal); err != nil {
		return nil, err
	}

	return toCalendarOutput(cal), nil
}

// DeleteCalendarUsecase handles deleting a calendar
type DeleteCalendarUsecase struct {
	calendarRepo calendar.Repository
}

// NewDeleteCalendarUsecase creates a new DeleteCalendarUsecase
func NewDeleteCalendarUsecase(calendarRepo calendar.Repository) *DeleteCalendarUsecase {
	return &DeleteCalendarUsecase{
		calendarRepo: calendarRepo,
	}
}

// Execute deletes a calendar
func (u *DeleteCalendarUsecase) Execute(ctx context.Context, input DeleteCalendarInput) error {
	// Parse IDs
	tenantID, err := common.ParseTenantID(input.TenantID)
	if err != nil {
		return err
	}
	calendarID, err := common.ParseCalendarID(input.CalendarID)
	if err != nil {
		return err
	}

	// Delete from repository
	return u.calendarRepo.Delete(ctx, tenantID, calendarID)
}

// toCalendarOutput converts a Calendar entity to CalendarOutput DTO
func toCalendarOutput(cal *calendar.Calendar) *CalendarOutput {
	// Convert EventIDs to string slice
	eventIDStrs := make([]string, 0, len(cal.EventIDs()))
	for _, eid := range cal.EventIDs() {
		eventIDStrs = append(eventIDStrs, eid.String())
	}

	// Convert PublicToken
	var publicToken *string
	if cal.PublicToken() != nil {
		tokenStr := cal.PublicToken().String()
		publicToken = &tokenStr
	}

	return &CalendarOutput{
		ID:          cal.CalendarID().String(),
		TenantID:    cal.TenantID().String(),
		Title:       cal.Title(),
		Description: cal.Description(),
		IsPublic:    cal.IsPublic(),
		PublicToken: publicToken,
		EventIDs:    eventIDStrs,
		CreatedAt:   cal.CreatedAt(),
		UpdatedAt:   cal.UpdatedAt(),
	}
}
