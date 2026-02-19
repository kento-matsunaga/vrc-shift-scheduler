package calendar_test

import (
	"context"
	"errors"
	"testing"
	"time"

	appcalendar "github.com/erenoa/vrc-shift-scheduler/backend/internal/app/calendar"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/calendar"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/event"
)

// =============================================================================
// Mock Repositories
// =============================================================================

type mockCalendarRepository struct {
	createFunc            func(ctx context.Context, cal *calendar.Calendar) error
	findByIDFunc          func(ctx context.Context, tenantID common.TenantID, calendarID common.CalendarID) (*calendar.Calendar, error)
	findByTenantIDFunc    func(ctx context.Context, tenantID common.TenantID) ([]*calendar.Calendar, error)
	findByPublicTokenFunc func(ctx context.Context, token common.PublicToken) (*calendar.Calendar, error)
	updateFunc            func(ctx context.Context, cal *calendar.Calendar) error
	deleteFunc            func(ctx context.Context, tenantID common.TenantID, calendarID common.CalendarID) error
}

func (m *mockCalendarRepository) Create(ctx context.Context, cal *calendar.Calendar) error {
	if m.createFunc != nil {
		return m.createFunc(ctx, cal)
	}
	return nil
}

func (m *mockCalendarRepository) FindByID(ctx context.Context, tenantID common.TenantID, calendarID common.CalendarID) (*calendar.Calendar, error) {
	if m.findByIDFunc != nil {
		return m.findByIDFunc(ctx, tenantID, calendarID)
	}
	return nil, nil
}

func (m *mockCalendarRepository) FindByTenantID(ctx context.Context, tenantID common.TenantID) ([]*calendar.Calendar, error) {
	if m.findByTenantIDFunc != nil {
		return m.findByTenantIDFunc(ctx, tenantID)
	}
	return nil, nil
}

func (m *mockCalendarRepository) FindByPublicToken(ctx context.Context, token common.PublicToken) (*calendar.Calendar, error) {
	if m.findByPublicTokenFunc != nil {
		return m.findByPublicTokenFunc(ctx, token)
	}
	return nil, nil
}

func (m *mockCalendarRepository) Update(ctx context.Context, cal *calendar.Calendar) error {
	if m.updateFunc != nil {
		return m.updateFunc(ctx, cal)
	}
	return nil
}

func (m *mockCalendarRepository) Delete(ctx context.Context, tenantID common.TenantID, calendarID common.CalendarID) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, tenantID, calendarID)
	}
	return nil
}

type mockEventRepository struct {
	findByIDFunc func(ctx context.Context, tenantID common.TenantID, eventID common.EventID) (*event.Event, error)
}

func (m *mockEventRepository) Save(ctx context.Context, evt *event.Event) error {
	return nil
}

func (m *mockEventRepository) FindByID(ctx context.Context, tenantID common.TenantID, eventID common.EventID) (*event.Event, error) {
	if m.findByIDFunc != nil {
		return m.findByIDFunc(ctx, tenantID, eventID)
	}
	return nil, nil
}

func (m *mockEventRepository) FindByTenantID(ctx context.Context, tenantID common.TenantID) ([]*event.Event, error) {
	return nil, nil
}

func (m *mockEventRepository) FindActiveByTenantID(ctx context.Context, tenantID common.TenantID) ([]*event.Event, error) {
	return nil, nil
}

func (m *mockEventRepository) Delete(ctx context.Context, tenantID common.TenantID, eventID common.EventID) error {
	return nil
}

func (m *mockEventRepository) ExistsByName(ctx context.Context, tenantID common.TenantID, eventName string) (bool, error) {
	return false, nil
}

type mockBusinessDayRepository struct {
	findByEventIDFunc func(ctx context.Context, tenantID common.TenantID, eventID common.EventID) ([]*event.EventBusinessDay, error)
}

func (m *mockBusinessDayRepository) Save(ctx context.Context, businessDay *event.EventBusinessDay) error {
	return nil
}

func (m *mockBusinessDayRepository) FindByID(ctx context.Context, tenantID common.TenantID, businessDayID event.BusinessDayID) (*event.EventBusinessDay, error) {
	return nil, nil
}

func (m *mockBusinessDayRepository) FindByEventID(ctx context.Context, tenantID common.TenantID, eventID common.EventID) ([]*event.EventBusinessDay, error) {
	if m.findByEventIDFunc != nil {
		return m.findByEventIDFunc(ctx, tenantID, eventID)
	}
	return nil, nil
}

func (m *mockBusinessDayRepository) FindByEventIDAndDateRange(ctx context.Context, tenantID common.TenantID, eventID common.EventID, startDate, endDate time.Time) ([]*event.EventBusinessDay, error) {
	return nil, nil
}

func (m *mockBusinessDayRepository) FindActiveByEventID(ctx context.Context, tenantID common.TenantID, eventID common.EventID) ([]*event.EventBusinessDay, error) {
	return nil, nil
}

func (m *mockBusinessDayRepository) FindByTenantIDAndDate(ctx context.Context, tenantID common.TenantID, date time.Time) ([]*event.EventBusinessDay, error) {
	return nil, nil
}

func (m *mockBusinessDayRepository) Delete(ctx context.Context, tenantID common.TenantID, businessDayID event.BusinessDayID) error {
	return nil
}

func (m *mockBusinessDayRepository) ExistsByEventIDAndDate(ctx context.Context, tenantID common.TenantID, eventID common.EventID, date time.Time, startTime time.Time) (bool, error) {
	return false, nil
}

func (m *mockBusinessDayRepository) FindRecentByTenantID(ctx context.Context, tenantID common.TenantID, limit int) ([]*event.EventBusinessDay, error) {
	return nil, nil
}

func (m *mockBusinessDayRepository) FindRecentByEventID(ctx context.Context, tenantID common.TenantID, eventID common.EventID, limit int, includeFuture bool) ([]*event.EventBusinessDay, error) {
	return nil, nil
}

type mockCalendarEntryRepository struct {
	saveFunc           func(ctx context.Context, entry *calendar.CalendarEntry) error
	findByIDFunc       func(ctx context.Context, tenantID common.TenantID, entryID common.CalendarEntryID) (*calendar.CalendarEntry, error)
	findByCalendarIDFunc func(ctx context.Context, tenantID common.TenantID, calendarID common.CalendarID) ([]*calendar.CalendarEntry, error)
	deleteFunc         func(ctx context.Context, tenantID common.TenantID, entryID common.CalendarEntryID) error
}

func (m *mockCalendarEntryRepository) Save(ctx context.Context, entry *calendar.CalendarEntry) error {
	if m.saveFunc != nil {
		return m.saveFunc(ctx, entry)
	}
	return nil
}

func (m *mockCalendarEntryRepository) FindByID(ctx context.Context, tenantID common.TenantID, entryID common.CalendarEntryID) (*calendar.CalendarEntry, error) {
	if m.findByIDFunc != nil {
		return m.findByIDFunc(ctx, tenantID, entryID)
	}
	return nil, nil
}

func (m *mockCalendarEntryRepository) FindByCalendarID(ctx context.Context, tenantID common.TenantID, calendarID common.CalendarID) ([]*calendar.CalendarEntry, error) {
	if m.findByCalendarIDFunc != nil {
		return m.findByCalendarIDFunc(ctx, tenantID, calendarID)
	}
	return nil, nil
}

func (m *mockCalendarEntryRepository) Delete(ctx context.Context, tenantID common.TenantID, entryID common.CalendarEntryID) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, tenantID, entryID)
	}
	return nil
}

type mockClock struct{}

func (m *mockClock) Now() time.Time { return time.Now() }

// =============================================================================
// Test Helpers
// =============================================================================

func createTestTenantID(t *testing.T) common.TenantID {
	t.Helper()
	return common.NewTenantIDWithTime(time.Now())
}

func createTestEventID(t *testing.T) common.EventID {
	t.Helper()
	return common.NewEventIDWithTime(time.Now())
}

func createTestCalendarID(t *testing.T) common.CalendarID {
	t.Helper()
	return common.NewCalendarIDWithTime(time.Now())
}

func createTestEvent(t *testing.T, tenantID common.TenantID) *event.Event {
	t.Helper()
	now := time.Now()
	evt, err := event.NewEvent(
		now,
		tenantID,
		"Test Event",
		event.EventTypeNormal,
		"Test Description",
		event.RecurrenceTypeNone,
		nil,
		nil,
		nil,
		nil,
	)
	if err != nil {
		t.Fatalf("failed to create test event: %v", err)
	}
	return evt
}

func createTestCalendar(t *testing.T, tenantID common.TenantID, eventIDs []common.EventID) *calendar.Calendar {
	t.Helper()
	now := time.Now()
	cal, err := calendar.NewCalendar(now, tenantID, "Test Calendar", "Test Description", eventIDs)
	if err != nil {
		t.Fatalf("failed to create test calendar: %v", err)
	}
	return cal
}

func createTestBusinessDay(t *testing.T, tenantID common.TenantID, eventID common.EventID) *event.EventBusinessDay {
	t.Helper()
	now := time.Now()
	targetDate := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	startTime := time.Date(2000, 1, 1, 20, 0, 0, 0, time.UTC)
	endTime := time.Date(2000, 1, 1, 23, 0, 0, 0, time.UTC)

	bd, err := event.NewEventBusinessDay(
		now,
		tenantID,
		eventID,
		targetDate,
		startTime,
		endTime,
		event.OccurrenceTypeSpecial,
		nil,
	)
	if err != nil {
		t.Fatalf("failed to create test business day: %v", err)
	}
	return bd
}

// =============================================================================
// CreateCalendarUsecase Tests
// =============================================================================

func TestCreateCalendarUsecase_Success(t *testing.T) {
	tenantID := createTestTenantID(t)
	eventID := createTestEventID(t)
	testEvent := createTestEvent(t, tenantID)

	mockCalRepo := &mockCalendarRepository{
		createFunc: func(ctx context.Context, cal *calendar.Calendar) error {
			return nil
		},
	}

	mockEventRepo := &mockEventRepository{
		findByIDFunc: func(ctx context.Context, tid common.TenantID, eid common.EventID) (*event.Event, error) {
			return testEvent, nil
		},
	}

	uc := appcalendar.NewCreateCalendarUsecase(mockCalRepo, mockEventRepo, &mockClock{})

	input := appcalendar.CreateCalendarInput{
		TenantID:    tenantID.String(),
		Title:       "New Calendar",
		Description: "Calendar Description",
		EventIDs:    []string{eventID.String()},
	}

	result, err := uc.Execute(context.Background(), input)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if result == nil {
		t.Fatal("expected result, got nil")
	}
	if result.Title != "New Calendar" {
		t.Errorf("expected title 'New Calendar', got '%s'", result.Title)
	}
	if result.Description != "Calendar Description" {
		t.Errorf("expected description 'Calendar Description', got '%s'", result.Description)
	}
	if result.IsPublic {
		t.Error("expected IsPublic to be false")
	}
}

func TestCreateCalendarUsecase_ErrorWhenEventNotFound(t *testing.T) {
	tenantID := createTestTenantID(t)
	eventID := createTestEventID(t)

	mockCalRepo := &mockCalendarRepository{}
	mockEventRepo := &mockEventRepository{
		findByIDFunc: func(ctx context.Context, tid common.TenantID, eid common.EventID) (*event.Event, error) {
			return nil, common.NewNotFoundError("event", eid.String())
		},
	}

	uc := appcalendar.NewCreateCalendarUsecase(mockCalRepo, mockEventRepo, &mockClock{})

	input := appcalendar.CreateCalendarInput{
		TenantID:    tenantID.String(),
		Title:       "New Calendar",
		Description: "Calendar Description",
		EventIDs:    []string{eventID.String()},
	}

	result, err := uc.Execute(context.Background(), input)

	if err == nil {
		t.Error("expected error, got nil")
	}
	if result != nil {
		t.Errorf("expected nil result, got %v", result)
	}
	if !common.IsNotFoundError(err) {
		t.Errorf("expected not found error, got %v", err)
	}
}

func TestCreateCalendarUsecase_ErrorWhenInvalidTenantID(t *testing.T) {
	mockCalRepo := &mockCalendarRepository{}
	mockEventRepo := &mockEventRepository{}

	uc := appcalendar.NewCreateCalendarUsecase(mockCalRepo, mockEventRepo, &mockClock{})

	input := appcalendar.CreateCalendarInput{
		TenantID:    "invalid-tenant-id",
		Title:       "New Calendar",
		Description: "Calendar Description",
		EventIDs:    []string{},
	}

	result, err := uc.Execute(context.Background(), input)

	if err == nil {
		t.Error("expected error, got nil")
	}
	if result != nil {
		t.Errorf("expected nil result, got %v", result)
	}
}

// =============================================================================
// GetCalendarUsecase Tests
// =============================================================================

func TestGetCalendarUsecase_Success(t *testing.T) {
	tenantID := createTestTenantID(t)
	eventID := createTestEventID(t)
	testCalendar := createTestCalendar(t, tenantID, []common.EventID{eventID})
	testEvent := createTestEvent(t, tenantID)
	testBusinessDay := createTestBusinessDay(t, tenantID, eventID)

	mockCalRepo := &mockCalendarRepository{
		findByIDFunc: func(ctx context.Context, tid common.TenantID, cid common.CalendarID) (*calendar.Calendar, error) {
			return testCalendar, nil
		},
	}

	mockEventRepo := &mockEventRepository{
		findByIDFunc: func(ctx context.Context, tid common.TenantID, eid common.EventID) (*event.Event, error) {
			return testEvent, nil
		},
	}

	mockBdRepo := &mockBusinessDayRepository{
		findByEventIDFunc: func(ctx context.Context, tid common.TenantID, eid common.EventID) ([]*event.EventBusinessDay, error) {
			return []*event.EventBusinessDay{testBusinessDay}, nil
		},
	}

	uc := appcalendar.NewGetCalendarUsecase(mockCalRepo, mockEventRepo, mockBdRepo)

	input := appcalendar.GetCalendarInput{
		TenantID:   tenantID.String(),
		CalendarID: testCalendar.CalendarID().String(),
	}

	result, err := uc.Execute(context.Background(), input)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if result == nil {
		t.Fatal("expected result, got nil")
	}
	if result.Calendar.Title != "Test Calendar" {
		t.Errorf("expected title 'Test Calendar', got '%s'", result.Calendar.Title)
	}
	if len(result.Events) != 1 {
		t.Errorf("expected 1 event, got %d", len(result.Events))
	}
}

func TestGetCalendarUsecase_ErrorWhenCalendarNotFound(t *testing.T) {
	tenantID := createTestTenantID(t)
	calendarID := createTestCalendarID(t)

	mockCalRepo := &mockCalendarRepository{
		findByIDFunc: func(ctx context.Context, tid common.TenantID, cid common.CalendarID) (*calendar.Calendar, error) {
			return nil, common.NewNotFoundError("calendar", cid.String())
		},
	}

	mockEventRepo := &mockEventRepository{}
	mockBdRepo := &mockBusinessDayRepository{}

	uc := appcalendar.NewGetCalendarUsecase(mockCalRepo, mockEventRepo, mockBdRepo)

	input := appcalendar.GetCalendarInput{
		TenantID:   tenantID.String(),
		CalendarID: calendarID.String(),
	}

	result, err := uc.Execute(context.Background(), input)

	if err == nil {
		t.Error("expected error, got nil")
	}
	if result != nil {
		t.Errorf("expected nil result, got %v", result)
	}
	if !common.IsNotFoundError(err) {
		t.Errorf("expected not found error, got %v", err)
	}
}

// =============================================================================
// UpdateCalendarUsecase Tests
// =============================================================================

func TestUpdateCalendarUsecase_Success(t *testing.T) {
	tenantID := createTestTenantID(t)
	eventID := createTestEventID(t)
	testCalendar := createTestCalendar(t, tenantID, []common.EventID{eventID})
	testEvent := createTestEvent(t, tenantID)

	mockCalRepo := &mockCalendarRepository{
		findByIDFunc: func(ctx context.Context, tid common.TenantID, cid common.CalendarID) (*calendar.Calendar, error) {
			return testCalendar, nil
		},
		updateFunc: func(ctx context.Context, cal *calendar.Calendar) error {
			return nil
		},
	}

	mockEventRepo := &mockEventRepository{
		findByIDFunc: func(ctx context.Context, tid common.TenantID, eid common.EventID) (*event.Event, error) {
			return testEvent, nil
		},
	}

	uc := appcalendar.NewUpdateCalendarUsecase(mockCalRepo, mockEventRepo, &mockClock{})

	input := appcalendar.UpdateCalendarInput{
		TenantID:    tenantID.String(),
		CalendarID:  testCalendar.CalendarID().String(),
		Title:       "Updated Calendar",
		Description: "Updated Description",
		EventIDs:    []string{eventID.String()},
		IsPublic:    false,
	}

	result, err := uc.Execute(context.Background(), input)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if result == nil {
		t.Fatal("expected result, got nil")
	}
	if result.Title != "Updated Calendar" {
		t.Errorf("expected title 'Updated Calendar', got '%s'", result.Title)
	}
}

func TestUpdateCalendarUsecase_MakePublicGeneratesToken(t *testing.T) {
	tenantID := createTestTenantID(t)
	eventID := createTestEventID(t)
	testCalendar := createTestCalendar(t, tenantID, []common.EventID{eventID})
	testEvent := createTestEvent(t, tenantID)

	var updatedCalendar *calendar.Calendar

	mockCalRepo := &mockCalendarRepository{
		findByIDFunc: func(ctx context.Context, tid common.TenantID, cid common.CalendarID) (*calendar.Calendar, error) {
			return testCalendar, nil
		},
		updateFunc: func(ctx context.Context, cal *calendar.Calendar) error {
			updatedCalendar = cal
			return nil
		},
	}

	mockEventRepo := &mockEventRepository{
		findByIDFunc: func(ctx context.Context, tid common.TenantID, eid common.EventID) (*event.Event, error) {
			return testEvent, nil
		},
	}

	uc := appcalendar.NewUpdateCalendarUsecase(mockCalRepo, mockEventRepo, &mockClock{})

	input := appcalendar.UpdateCalendarInput{
		TenantID:    tenantID.String(),
		CalendarID:  testCalendar.CalendarID().String(),
		Title:       "Public Calendar",
		Description: "Public Description",
		EventIDs:    []string{eventID.String()},
		IsPublic:    true,
	}

	result, err := uc.Execute(context.Background(), input)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if result == nil {
		t.Fatal("expected result, got nil")
	}
	if !result.IsPublic {
		t.Error("expected IsPublic to be true")
	}
	if result.PublicToken == nil {
		t.Error("expected PublicToken to be generated")
	}
	if updatedCalendar == nil {
		t.Fatal("expected calendar to be updated")
	}
	if !updatedCalendar.IsPublic() {
		t.Error("expected updated calendar to be public")
	}
}

func TestUpdateCalendarUsecase_ErrorWhenCalendarNotFound(t *testing.T) {
	tenantID := createTestTenantID(t)
	calendarID := createTestCalendarID(t)

	mockCalRepo := &mockCalendarRepository{
		findByIDFunc: func(ctx context.Context, tid common.TenantID, cid common.CalendarID) (*calendar.Calendar, error) {
			return nil, common.NewNotFoundError("calendar", cid.String())
		},
	}

	mockEventRepo := &mockEventRepository{}

	uc := appcalendar.NewUpdateCalendarUsecase(mockCalRepo, mockEventRepo, &mockClock{})

	input := appcalendar.UpdateCalendarInput{
		TenantID:    tenantID.String(),
		CalendarID:  calendarID.String(),
		Title:       "Updated Calendar",
		Description: "Updated Description",
		EventIDs:    []string{},
		IsPublic:    false,
	}

	result, err := uc.Execute(context.Background(), input)

	if err == nil {
		t.Error("expected error, got nil")
	}
	if result != nil {
		t.Errorf("expected nil result, got %v", result)
	}
}

// =============================================================================
// DeleteCalendarUsecase Tests
// =============================================================================

func TestDeleteCalendarUsecase_Success(t *testing.T) {
	tenantID := createTestTenantID(t)
	calendarID := createTestCalendarID(t)

	deleteCalled := false
	mockCalRepo := &mockCalendarRepository{
		deleteFunc: func(ctx context.Context, tid common.TenantID, cid common.CalendarID) error {
			deleteCalled = true
			return nil
		},
	}

	uc := appcalendar.NewDeleteCalendarUsecase(mockCalRepo)

	input := appcalendar.DeleteCalendarInput{
		TenantID:   tenantID.String(),
		CalendarID: calendarID.String(),
	}

	err := uc.Execute(context.Background(), input)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if !deleteCalled {
		t.Error("expected delete to be called")
	}
}

func TestDeleteCalendarUsecase_ErrorWhenDeleteFails(t *testing.T) {
	tenantID := createTestTenantID(t)
	calendarID := createTestCalendarID(t)

	mockCalRepo := &mockCalendarRepository{
		deleteFunc: func(ctx context.Context, tid common.TenantID, cid common.CalendarID) error {
			return errors.New("database error")
		},
	}

	uc := appcalendar.NewDeleteCalendarUsecase(mockCalRepo)

	input := appcalendar.DeleteCalendarInput{
		TenantID:   tenantID.String(),
		CalendarID: calendarID.String(),
	}

	err := uc.Execute(context.Background(), input)

	if err == nil {
		t.Error("expected error, got nil")
	}
}

// =============================================================================
// GetCalendarByTokenUsecase Tests
// =============================================================================

func TestGetCalendarByTokenUsecase_Success(t *testing.T) {
	tenantID := createTestTenantID(t)
	eventID := createTestEventID(t)
	testCalendar := createTestCalendar(t, tenantID, []common.EventID{eventID})
	testEvent := createTestEvent(t, tenantID)
	testBusinessDay := createTestBusinessDay(t, tenantID, eventID)

	// Make the calendar public
	now := time.Now()
	testCalendar.MakePublic(now)

	mockCalRepo := &mockCalendarRepository{
		findByPublicTokenFunc: func(ctx context.Context, token common.PublicToken) (*calendar.Calendar, error) {
			return testCalendar, nil
		},
	}

	mockEventRepo := &mockEventRepository{
		findByIDFunc: func(ctx context.Context, tid common.TenantID, eid common.EventID) (*event.Event, error) {
			return testEvent, nil
		},
	}

	mockBdRepo := &mockBusinessDayRepository{
		findByEventIDFunc: func(ctx context.Context, tid common.TenantID, eid common.EventID) ([]*event.EventBusinessDay, error) {
			return []*event.EventBusinessDay{testBusinessDay}, nil
		},
	}

	mockEntryRepo := &mockCalendarEntryRepository{}

	uc := appcalendar.NewGetCalendarByTokenUsecase(mockCalRepo, mockEventRepo, mockBdRepo, mockEntryRepo)

	input := appcalendar.GetCalendarByTokenInput{
		Token: testCalendar.PublicToken().String(),
	}

	result, err := uc.Execute(context.Background(), input)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if result == nil {
		t.Fatal("expected result, got nil")
	}
	if result.Calendar.Title != "Test Calendar" {
		t.Errorf("expected title 'Test Calendar', got '%s'", result.Calendar.Title)
	}
	if !result.Calendar.IsPublic {
		t.Error("expected calendar to be public")
	}
}

func TestGetCalendarByTokenUsecase_ErrorWhenCalendarNotPublic(t *testing.T) {
	tenantID := createTestTenantID(t)
	eventID := createTestEventID(t)
	testCalendar := createTestCalendar(t, tenantID, []common.EventID{eventID})

	// Calendar is NOT public (default)

	mockCalRepo := &mockCalendarRepository{
		findByPublicTokenFunc: func(ctx context.Context, token common.PublicToken) (*calendar.Calendar, error) {
			return testCalendar, nil
		},
	}

	mockEventRepo := &mockEventRepository{}
	mockBdRepo := &mockBusinessDayRepository{}
	mockEntryRepo := &mockCalendarEntryRepository{}

	uc := appcalendar.NewGetCalendarByTokenUsecase(mockCalRepo, mockEventRepo, mockBdRepo, mockEntryRepo)

	// Use a valid UUID token
	validToken := common.NewPublicToken()
	input := appcalendar.GetCalendarByTokenInput{
		Token: validToken.String(),
	}

	result, err := uc.Execute(context.Background(), input)

	if err == nil {
		t.Error("expected error, got nil")
	}
	if result != nil {
		t.Errorf("expected nil result, got %v", result)
	}
	if !common.IsNotFoundError(err) {
		t.Errorf("expected not found error, got %v", err)
	}
}

func TestGetCalendarByTokenUsecase_ErrorWhenTokenNotFound(t *testing.T) {
	mockCalRepo := &mockCalendarRepository{
		findByPublicTokenFunc: func(ctx context.Context, token common.PublicToken) (*calendar.Calendar, error) {
			return nil, common.NewNotFoundError("calendar", token.String())
		},
	}

	mockEventRepo := &mockEventRepository{}
	mockBdRepo := &mockBusinessDayRepository{}
	mockEntryRepo := &mockCalendarEntryRepository{}

	uc := appcalendar.NewGetCalendarByTokenUsecase(mockCalRepo, mockEventRepo, mockBdRepo, mockEntryRepo)

	validToken := common.NewPublicToken()
	input := appcalendar.GetCalendarByTokenInput{
		Token: validToken.String(),
	}

	result, err := uc.Execute(context.Background(), input)

	if err == nil {
		t.Error("expected error, got nil")
	}
	if result != nil {
		t.Errorf("expected nil result, got %v", result)
	}
}

func TestGetCalendarByTokenUsecase_ErrorWhenInvalidToken(t *testing.T) {
	mockCalRepo := &mockCalendarRepository{}
	mockEventRepo := &mockEventRepository{}
	mockBdRepo := &mockBusinessDayRepository{}
	mockEntryRepo := &mockCalendarEntryRepository{}

	uc := appcalendar.NewGetCalendarByTokenUsecase(mockCalRepo, mockEventRepo, mockBdRepo, mockEntryRepo)

	input := appcalendar.GetCalendarByTokenInput{
		Token: "invalid-token",
	}

	result, err := uc.Execute(context.Background(), input)

	if err == nil {
		t.Error("expected error, got nil")
	}
	if result != nil {
		t.Errorf("expected nil result, got %v", result)
	}
}

// =============================================================================
// ListCalendarsUsecase Tests
// =============================================================================

func TestListCalendarsUsecase_Success(t *testing.T) {
	tenantID := createTestTenantID(t)
	eventID := createTestEventID(t)
	testCalendar1 := createTestCalendar(t, tenantID, []common.EventID{eventID})
	testCalendar2 := createTestCalendar(t, tenantID, []common.EventID{})

	mockCalRepo := &mockCalendarRepository{
		findByTenantIDFunc: func(ctx context.Context, tid common.TenantID) ([]*calendar.Calendar, error) {
			return []*calendar.Calendar{testCalendar1, testCalendar2}, nil
		},
	}

	uc := appcalendar.NewListCalendarsUsecase(mockCalRepo)

	input := appcalendar.ListCalendarsInput{
		TenantID: tenantID.String(),
	}

	result, err := uc.Execute(context.Background(), input)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if result == nil {
		t.Fatal("expected result, got nil")
	}
	if len(result) != 2 {
		t.Errorf("expected 2 calendars, got %d", len(result))
	}
}

func TestListCalendarsUsecase_EmptyList(t *testing.T) {
	tenantID := createTestTenantID(t)

	mockCalRepo := &mockCalendarRepository{
		findByTenantIDFunc: func(ctx context.Context, tid common.TenantID) ([]*calendar.Calendar, error) {
			return []*calendar.Calendar{}, nil
		},
	}

	uc := appcalendar.NewListCalendarsUsecase(mockCalRepo)

	input := appcalendar.ListCalendarsInput{
		TenantID: tenantID.String(),
	}

	result, err := uc.Execute(context.Background(), input)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if result == nil {
		t.Fatal("expected result, got nil")
	}
	if len(result) != 0 {
		t.Errorf("expected 0 calendars, got %d", len(result))
	}
}
