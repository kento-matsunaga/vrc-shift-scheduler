package event_test

import (
	"context"
	"errors"
	"testing"
	"time"

	appevent "github.com/erenoa/vrc-shift-scheduler/backend/internal/app/event"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/event"
)

// =====================================================
// Mock Repositories
// =====================================================

type MockEventRepository struct {
	saveFunc               func(ctx context.Context, e *event.Event) error
	findByIDFunc           func(ctx context.Context, tenantID common.TenantID, eventID common.EventID) (*event.Event, error)
	findByTenantFunc       func(ctx context.Context, tenantID common.TenantID) ([]*event.Event, error)
	findActiveByTenantFunc func(ctx context.Context, tenantID common.TenantID) ([]*event.Event, error)
	existsByNameFunc       func(ctx context.Context, tenantID common.TenantID, name string) (bool, error)
	deleteFunc             func(ctx context.Context, tenantID common.TenantID, eventID common.EventID) error
}

func (m *MockEventRepository) Save(ctx context.Context, e *event.Event) error {
	if m.saveFunc != nil {
		return m.saveFunc(ctx, e)
	}
	return nil
}

func (m *MockEventRepository) FindByID(ctx context.Context, tenantID common.TenantID, eventID common.EventID) (*event.Event, error) {
	if m.findByIDFunc != nil {
		return m.findByIDFunc(ctx, tenantID, eventID)
	}
	return nil, errors.New("not implemented")
}

func (m *MockEventRepository) FindByTenantID(ctx context.Context, tenantID common.TenantID) ([]*event.Event, error) {
	if m.findByTenantFunc != nil {
		return m.findByTenantFunc(ctx, tenantID)
	}
	return nil, errors.New("not implemented")
}

func (m *MockEventRepository) FindActiveByTenantID(ctx context.Context, tenantID common.TenantID) ([]*event.Event, error) {
	if m.findActiveByTenantFunc != nil {
		return m.findActiveByTenantFunc(ctx, tenantID)
	}
	return nil, errors.New("not implemented")
}

func (m *MockEventRepository) ExistsByName(ctx context.Context, tenantID common.TenantID, name string) (bool, error) {
	if m.existsByNameFunc != nil {
		return m.existsByNameFunc(ctx, tenantID, name)
	}
	return false, nil
}

func (m *MockEventRepository) Delete(ctx context.Context, tenantID common.TenantID, eventID common.EventID) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, tenantID, eventID)
	}
	return nil
}

type MockBusinessDayRepository struct {
	saveFunc               func(ctx context.Context, bd *event.EventBusinessDay) error
	findByEventIDFunc      func(ctx context.Context, tenantID common.TenantID, eventID common.EventID) ([]*event.EventBusinessDay, error)
	existsByEventIDAndDate func(ctx context.Context, tenantID common.TenantID, eventID common.EventID, date time.Time, startTime time.Time) (bool, error)
}

func (m *MockBusinessDayRepository) Save(ctx context.Context, bd *event.EventBusinessDay) error {
	if m.saveFunc != nil {
		return m.saveFunc(ctx, bd)
	}
	return nil
}

func (m *MockBusinessDayRepository) FindByEventID(ctx context.Context, tenantID common.TenantID, eventID common.EventID) ([]*event.EventBusinessDay, error) {
	if m.findByEventIDFunc != nil {
		return m.findByEventIDFunc(ctx, tenantID, eventID)
	}
	return nil, nil
}

func (m *MockBusinessDayRepository) ExistsByEventIDAndDate(ctx context.Context, tenantID common.TenantID, eventID common.EventID, date time.Time, startTime time.Time) (bool, error) {
	if m.existsByEventIDAndDate != nil {
		return m.existsByEventIDAndDate(ctx, tenantID, eventID, date, startTime)
	}
	return false, nil
}

func (m *MockBusinessDayRepository) FindByID(ctx context.Context, tenantID common.TenantID, businessDayID event.BusinessDayID) (*event.EventBusinessDay, error) {
	return nil, errors.New("not implemented")
}

func (m *MockBusinessDayRepository) FindByEventIDAndDateRange(ctx context.Context, tenantID common.TenantID, eventID common.EventID, startDate, endDate time.Time) ([]*event.EventBusinessDay, error) {
	return nil, nil
}

func (m *MockBusinessDayRepository) FindActiveByEventID(ctx context.Context, tenantID common.TenantID, eventID common.EventID) ([]*event.EventBusinessDay, error) {
	return nil, nil
}

func (m *MockBusinessDayRepository) FindByTenantIDAndDate(ctx context.Context, tenantID common.TenantID, date time.Time) ([]*event.EventBusinessDay, error) {
	return nil, nil
}

func (m *MockBusinessDayRepository) Delete(ctx context.Context, tenantID common.TenantID, businessDayID event.BusinessDayID) error {
	return nil
}

func (m *MockBusinessDayRepository) FindRecentByTenantID(ctx context.Context, tenantID common.TenantID, limit int) ([]*event.EventBusinessDay, error) {
	return nil, nil
}

func (m *MockBusinessDayRepository) FindRecentByEventID(ctx context.Context, tenantID common.TenantID, eventID common.EventID, limit int, includeFuture bool) ([]*event.EventBusinessDay, error) {
	return nil, nil
}

// =====================================================
// CreateEventUsecase Tests
// =====================================================

func TestCreateEventUsecase_Execute_Success(t *testing.T) {
	tenantID := common.NewTenantID()

	eventRepo := &MockEventRepository{
		existsByNameFunc: func(ctx context.Context, tid common.TenantID, name string) (bool, error) {
			return false, nil // Event does not exist
		},
		saveFunc: func(ctx context.Context, e *event.Event) error {
			return nil
		},
	}

	bdRepo := &MockBusinessDayRepository{}

	usecase := appevent.NewCreateEventUsecase(eventRepo, bdRepo)

	input := appevent.CreateEventInput{
		TenantID:       tenantID,
		EventName:      "Test Event",
		EventType:      event.EventTypeNormal,
		Description:    "Test Description",
		RecurrenceType: event.RecurrenceTypeNone,
	}

	result, err := usecase.Execute(context.Background(), input)

	if err != nil {
		t.Fatalf("Execute() should succeed, got error: %v", err)
	}

	if result == nil {
		t.Fatal("Result should not be nil")
	}

	if result.EventName() != "Test Event" {
		t.Errorf("EventName mismatch: got %v, want 'Test Event'", result.EventName())
	}

	if result.TenantID() != tenantID {
		t.Errorf("TenantID mismatch: got %v, want %v", result.TenantID(), tenantID)
	}
}

func TestCreateEventUsecase_Execute_ErrorWhenNameExists(t *testing.T) {
	tenantID := common.NewTenantID()

	eventRepo := &MockEventRepository{
		existsByNameFunc: func(ctx context.Context, tid common.TenantID, name string) (bool, error) {
			return true, nil // Event already exists
		},
	}

	bdRepo := &MockBusinessDayRepository{}

	usecase := appevent.NewCreateEventUsecase(eventRepo, bdRepo)

	input := appevent.CreateEventInput{
		TenantID:       tenantID,
		EventName:      "Existing Event",
		EventType:      event.EventTypeNormal,
		Description:    "Test Description",
		RecurrenceType: event.RecurrenceTypeNone,
	}

	_, err := usecase.Execute(context.Background(), input)

	if err == nil {
		t.Fatal("Execute() should fail when event name already exists")
	}
}

func TestCreateEventUsecase_Execute_ErrorWhenSaveFails(t *testing.T) {
	tenantID := common.NewTenantID()

	eventRepo := &MockEventRepository{
		existsByNameFunc: func(ctx context.Context, tid common.TenantID, name string) (bool, error) {
			return false, nil
		},
		saveFunc: func(ctx context.Context, e *event.Event) error {
			return errors.New("database error")
		},
	}

	bdRepo := &MockBusinessDayRepository{}

	usecase := appevent.NewCreateEventUsecase(eventRepo, bdRepo)

	input := appevent.CreateEventInput{
		TenantID:       tenantID,
		EventName:      "Test Event",
		EventType:      event.EventTypeNormal,
		Description:    "Test Description",
		RecurrenceType: event.RecurrenceTypeNone,
	}

	_, err := usecase.Execute(context.Background(), input)

	if err == nil {
		t.Fatal("Execute() should fail when save fails")
	}
}

// =====================================================
// ListEventsUsecase Tests
// =====================================================

func TestListEventsUsecase_Execute_Success(t *testing.T) {
	tenantID := common.NewTenantID()
	now := time.Now()

	testEvents := []*event.Event{}
	for i := 0; i < 3; i++ {
		e, _ := event.NewEvent(now, tenantID, "Event "+string(rune('A'+i)), event.EventTypeNormal, "Desc", event.RecurrenceTypeNone, nil, nil, nil, nil)
		testEvents = append(testEvents, e)
	}

	eventRepo := &MockEventRepository{
		findByTenantFunc: func(ctx context.Context, tid common.TenantID) ([]*event.Event, error) {
			return testEvents, nil
		},
	}

	usecase := appevent.NewListEventsUsecase(eventRepo)

	input := appevent.ListEventsInput{
		TenantID: tenantID,
	}

	result, err := usecase.Execute(context.Background(), input)

	if err != nil {
		t.Fatalf("Execute() should succeed, got error: %v", err)
	}

	if len(result) != 3 {
		t.Errorf("Expected 3 events, got %d", len(result))
	}
}

func TestListEventsUsecase_Execute_EmptyList(t *testing.T) {
	tenantID := common.NewTenantID()

	eventRepo := &MockEventRepository{
		findByTenantFunc: func(ctx context.Context, tid common.TenantID) ([]*event.Event, error) {
			return []*event.Event{}, nil
		},
	}

	usecase := appevent.NewListEventsUsecase(eventRepo)

	input := appevent.ListEventsInput{
		TenantID: tenantID,
	}

	result, err := usecase.Execute(context.Background(), input)

	if err != nil {
		t.Fatalf("Execute() should succeed, got error: %v", err)
	}

	if len(result) != 0 {
		t.Errorf("Expected 0 events, got %d", len(result))
	}
}

func TestListEventsUsecase_Execute_ErrorWhenFindFails(t *testing.T) {
	tenantID := common.NewTenantID()

	eventRepo := &MockEventRepository{
		findByTenantFunc: func(ctx context.Context, tid common.TenantID) ([]*event.Event, error) {
			return nil, errors.New("database error")
		},
	}

	usecase := appevent.NewListEventsUsecase(eventRepo)

	input := appevent.ListEventsInput{
		TenantID: tenantID,
	}

	_, err := usecase.Execute(context.Background(), input)

	if err == nil {
		t.Fatal("Execute() should fail when find fails")
	}
}

// =====================================================
// GetEventUsecase Tests
// =====================================================

func TestGetEventUsecase_Execute_Success(t *testing.T) {
	tenantID := common.NewTenantID()
	now := time.Now()

	testEvent, _ := event.NewEvent(now, tenantID, "Test Event", event.EventTypeNormal, "Desc", event.RecurrenceTypeNone, nil, nil, nil, nil)

	eventRepo := &MockEventRepository{
		findByIDFunc: func(ctx context.Context, tid common.TenantID, eid common.EventID) (*event.Event, error) {
			return testEvent, nil
		},
	}

	usecase := appevent.NewGetEventUsecase(eventRepo)

	input := appevent.GetEventInput{
		TenantID: tenantID,
		EventID:  testEvent.EventID(),
	}

	result, err := usecase.Execute(context.Background(), input)

	if err != nil {
		t.Fatalf("Execute() should succeed, got error: %v", err)
	}

	if result.EventID() != testEvent.EventID() {
		t.Errorf("EventID mismatch: got %v, want %v", result.EventID(), testEvent.EventID())
	}
}

func TestGetEventUsecase_Execute_NotFound(t *testing.T) {
	tenantID := common.NewTenantID()
	eventID := common.NewEventID()

	eventRepo := &MockEventRepository{
		findByIDFunc: func(ctx context.Context, tid common.TenantID, eid common.EventID) (*event.Event, error) {
			return nil, common.NewNotFoundError("event", eid.String())
		},
	}

	usecase := appevent.NewGetEventUsecase(eventRepo)

	input := appevent.GetEventInput{
		TenantID: tenantID,
		EventID:  eventID,
	}

	_, err := usecase.Execute(context.Background(), input)

	if err == nil {
		t.Fatal("Execute() should fail when event not found")
	}
}

// =====================================================
// UpdateEventUsecase Tests
// =====================================================

func TestUpdateEventUsecase_Execute_Success(t *testing.T) {
	tenantID := common.NewTenantID()
	now := time.Now()

	testEvent, _ := event.NewEvent(now, tenantID, "Original Name", event.EventTypeNormal, "Desc", event.RecurrenceTypeNone, nil, nil, nil, nil)

	eventRepo := &MockEventRepository{
		findByIDFunc: func(ctx context.Context, tid common.TenantID, eid common.EventID) (*event.Event, error) {
			return testEvent, nil
		},
		saveFunc: func(ctx context.Context, e *event.Event) error {
			return nil
		},
	}

	usecase := appevent.NewUpdateEventUsecase(eventRepo)

	input := appevent.UpdateEventInput{
		TenantID:  tenantID,
		EventID:   testEvent.EventID(),
		EventName: "Updated Name",
	}

	result, err := usecase.Execute(context.Background(), input)

	if err != nil {
		t.Fatalf("Execute() should succeed, got error: %v", err)
	}

	if result.EventName() != "Updated Name" {
		t.Errorf("EventName should be updated: got %v, want 'Updated Name'", result.EventName())
	}
}

func TestUpdateEventUsecase_Execute_ErrorWhenNotFound(t *testing.T) {
	tenantID := common.NewTenantID()
	eventID := common.NewEventID()

	eventRepo := &MockEventRepository{
		findByIDFunc: func(ctx context.Context, tid common.TenantID, eid common.EventID) (*event.Event, error) {
			return nil, common.NewNotFoundError("event", eid.String())
		},
	}

	usecase := appevent.NewUpdateEventUsecase(eventRepo)

	input := appevent.UpdateEventInput{
		TenantID:  tenantID,
		EventID:   eventID,
		EventName: "Updated Name",
	}

	_, err := usecase.Execute(context.Background(), input)

	if err == nil {
		t.Fatal("Execute() should fail when event not found")
	}
}

// =====================================================
// DeleteEventUsecase Tests
// =====================================================

func TestDeleteEventUsecase_Execute_Success(t *testing.T) {
	tenantID := common.NewTenantID()
	now := time.Now()

	testEvent, _ := event.NewEvent(now, tenantID, "Test Event", event.EventTypeNormal, "Desc", event.RecurrenceTypeNone, nil, nil, nil, nil)

	eventRepo := &MockEventRepository{
		findByIDFunc: func(ctx context.Context, tid common.TenantID, eid common.EventID) (*event.Event, error) {
			return testEvent, nil
		},
		saveFunc: func(ctx context.Context, e *event.Event) error {
			return nil
		},
	}

	usecase := appevent.NewDeleteEventUsecase(eventRepo)

	input := appevent.DeleteEventInput{
		TenantID: tenantID,
		EventID:  testEvent.EventID(),
	}

	err := usecase.Execute(context.Background(), input)

	if err != nil {
		t.Fatalf("Execute() should succeed, got error: %v", err)
	}
}

func TestDeleteEventUsecase_Execute_ErrorWhenNotFound(t *testing.T) {
	tenantID := common.NewTenantID()
	eventID := common.NewEventID()

	eventRepo := &MockEventRepository{
		findByIDFunc: func(ctx context.Context, tid common.TenantID, eid common.EventID) (*event.Event, error) {
			return nil, common.NewNotFoundError("event", eid.String())
		},
	}

	usecase := appevent.NewDeleteEventUsecase(eventRepo)

	input := appevent.DeleteEventInput{
		TenantID: tenantID,
		EventID:  eventID,
	}

	err := usecase.Execute(context.Background(), input)

	if err == nil {
		t.Fatal("Execute() should fail when event not found")
	}
}

// =====================================================
// GenerateBusinessDaysUsecase Tests
// =====================================================

// createEventWithRecurrence creates an event with weekly recurrence for testing
func createEventWithRecurrence(t *testing.T, tenantID common.TenantID) *event.Event {
	t.Helper()
	now := time.Now()
	recurrenceStart := now
	dayOfWeek := int(now.Weekday()) // 今日の曜日
	startTime := time.Date(0, 1, 1, 21, 30, 0, 0, time.UTC)
	endTime := time.Date(0, 1, 1, 23, 0, 0, 0, time.UTC)

	e, err := event.NewEvent(
		now,
		tenantID,
		"Recurring Event",
		event.EventTypeNormal,
		"Test recurring event",
		event.RecurrenceTypeWeekly,
		&recurrenceStart,
		&dayOfWeek,
		&startTime,
		&endTime,
	)
	if err != nil {
		t.Fatalf("Failed to create event with recurrence: %v", err)
	}
	return e
}

func TestGenerateBusinessDaysUsecase_Execute_Success(t *testing.T) {
	tenantID := common.NewTenantID()
	testEvent := createEventWithRecurrence(t, tenantID)

	var savedCount int
	eventRepo := &MockEventRepository{
		findByIDFunc: func(ctx context.Context, tid common.TenantID, eid common.EventID) (*event.Event, error) {
			return testEvent, nil
		},
	}

	bdRepo := &MockBusinessDayRepository{
		existsByEventIDAndDate: func(ctx context.Context, tid common.TenantID, eid common.EventID, date time.Time, startTime time.Time) (bool, error) {
			return false, nil // No existing business days
		},
		saveFunc: func(ctx context.Context, bd *event.EventBusinessDay) error {
			savedCount++
			return nil
		},
	}

	usecase := appevent.NewGenerateBusinessDaysUsecase(eventRepo, bdRepo)

	input := appevent.GenerateBusinessDaysInput{
		TenantID: tenantID,
		EventID:  testEvent.EventID(),
		Months:   3,
	}

	result, err := usecase.Execute(context.Background(), input)

	if err != nil {
		t.Fatalf("Execute() should succeed, got error: %v", err)
	}

	if result == nil {
		t.Fatal("Result should not be nil")
	}

	// 3ヶ月分なので少なくとも12個（週1回 × 12週程度）以上生成されるはず
	if result.GeneratedCount < 10 {
		t.Errorf("Expected at least 10 business days for 3 months, got %d", result.GeneratedCount)
	}

	if savedCount != result.GeneratedCount {
		t.Errorf("Saved count mismatch: repo saved %d, result says %d", savedCount, result.GeneratedCount)
	}
}

func TestGenerateBusinessDaysUsecase_Execute_DefaultMonths(t *testing.T) {
	tenantID := common.NewTenantID()
	testEvent := createEventWithRecurrence(t, tenantID)

	var savedCount int
	eventRepo := &MockEventRepository{
		findByIDFunc: func(ctx context.Context, tid common.TenantID, eid common.EventID) (*event.Event, error) {
			return testEvent, nil
		},
	}

	bdRepo := &MockBusinessDayRepository{
		existsByEventIDAndDate: func(ctx context.Context, tid common.TenantID, eid common.EventID, date time.Time, startTime time.Time) (bool, error) {
			return false, nil
		},
		saveFunc: func(ctx context.Context, bd *event.EventBusinessDay) error {
			savedCount++
			return nil
		},
	}

	usecase := appevent.NewGenerateBusinessDaysUsecase(eventRepo, bdRepo)

	// months=0 → デフォルト2ヶ月に設定される
	input := appevent.GenerateBusinessDaysInput{
		TenantID: tenantID,
		EventID:  testEvent.EventID(),
		Months:   0,
	}

	result, err := usecase.Execute(context.Background(), input)

	if err != nil {
		t.Fatalf("Execute() should succeed, got error: %v", err)
	}

	// 2ヶ月分なので8個以上生成されるはず
	if result.GeneratedCount < 6 {
		t.Errorf("Expected at least 6 business days for default 2 months, got %d", result.GeneratedCount)
	}
}

func TestGenerateBusinessDaysUsecase_Execute_MaxMonthsLimit(t *testing.T) {
	tenantID := common.NewTenantID()
	testEvent := createEventWithRecurrence(t, tenantID)

	var savedCount int
	eventRepo := &MockEventRepository{
		findByIDFunc: func(ctx context.Context, tid common.TenantID, eid common.EventID) (*event.Event, error) {
			return testEvent, nil
		},
	}

	bdRepo := &MockBusinessDayRepository{
		existsByEventIDAndDate: func(ctx context.Context, tid common.TenantID, eid common.EventID, date time.Time, startTime time.Time) (bool, error) {
			return false, nil
		},
		saveFunc: func(ctx context.Context, bd *event.EventBusinessDay) error {
			savedCount++
			return nil
		},
	}

	usecase := appevent.NewGenerateBusinessDaysUsecase(eventRepo, bdRepo)

	// months=30 → 24ヶ月に制限される
	input := appevent.GenerateBusinessDaysInput{
		TenantID: tenantID,
		EventID:  testEvent.EventID(),
		Months:   30,
	}

	result, err := usecase.Execute(context.Background(), input)

	if err != nil {
		t.Fatalf("Execute() should succeed, got error: %v", err)
	}

	// 24ヶ月（最大）の場合、約100個程度生成されるはず
	// 最大24ヶ月で約104週
	if result.GeneratedCount > 110 {
		t.Errorf("Should be limited to 24 months max, got %d business days (too many)", result.GeneratedCount)
	}
}

func TestGenerateBusinessDaysUsecase_Execute_NegativeMonths(t *testing.T) {
	tenantID := common.NewTenantID()
	testEvent := createEventWithRecurrence(t, tenantID)

	var savedCount int
	eventRepo := &MockEventRepository{
		findByIDFunc: func(ctx context.Context, tid common.TenantID, eid common.EventID) (*event.Event, error) {
			return testEvent, nil
		},
	}

	bdRepo := &MockBusinessDayRepository{
		existsByEventIDAndDate: func(ctx context.Context, tid common.TenantID, eid common.EventID, date time.Time, startTime time.Time) (bool, error) {
			return false, nil
		},
		saveFunc: func(ctx context.Context, bd *event.EventBusinessDay) error {
			savedCount++
			return nil
		},
	}

	usecase := appevent.NewGenerateBusinessDaysUsecase(eventRepo, bdRepo)

	// months=-5 → デフォルト2ヶ月に設定される
	input := appevent.GenerateBusinessDaysInput{
		TenantID: tenantID,
		EventID:  testEvent.EventID(),
		Months:   -5,
	}

	result, err := usecase.Execute(context.Background(), input)

	if err != nil {
		t.Fatalf("Execute() should succeed, got error: %v", err)
	}

	// 2ヶ月分（デフォルト）なので6個以上生成されるはず
	if result.GeneratedCount < 6 {
		t.Errorf("Expected at least 6 business days for default 2 months (from negative), got %d", result.GeneratedCount)
	}
}

func TestGenerateBusinessDaysUsecase_Execute_NoRecurrence(t *testing.T) {
	tenantID := common.NewTenantID()
	now := time.Now()

	// 定期設定なしのイベント
	testEvent, _ := event.NewEvent(
		now,
		tenantID,
		"Non-recurring Event",
		event.EventTypeNormal,
		"Test event without recurrence",
		event.RecurrenceTypeNone,
		nil, nil, nil, nil,
	)

	eventRepo := &MockEventRepository{
		findByIDFunc: func(ctx context.Context, tid common.TenantID, eid common.EventID) (*event.Event, error) {
			return testEvent, nil
		},
	}

	bdRepo := &MockBusinessDayRepository{}

	usecase := appevent.NewGenerateBusinessDaysUsecase(eventRepo, bdRepo)

	input := appevent.GenerateBusinessDaysInput{
		TenantID: tenantID,
		EventID:  testEvent.EventID(),
		Months:   3,
	}

	_, err := usecase.Execute(context.Background(), input)

	if err == nil {
		t.Fatal("Execute() should fail when event has no recurrence")
	}

	// DomainError であることを確認（ValidationError = code INVALID_INPUT）
	domainErr, ok := err.(*common.DomainError)
	if !ok {
		t.Errorf("Expected DomainError, got: %T", err)
	} else if domainErr.Code() != common.ErrInvalidInput {
		t.Errorf("Expected error code %s, got: %s", common.ErrInvalidInput, domainErr.Code())
	}
}
