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
	saveFunc              func(ctx context.Context, e *event.Event) error
	findByIDFunc          func(ctx context.Context, tenantID common.TenantID, eventID common.EventID) (*event.Event, error)
	findByTenantFunc      func(ctx context.Context, tenantID common.TenantID) ([]*event.Event, error)
	findActiveByTenantFunc func(ctx context.Context, tenantID common.TenantID) ([]*event.Event, error)
	existsByNameFunc      func(ctx context.Context, tenantID common.TenantID, name string) (bool, error)
	deleteFunc            func(ctx context.Context, tenantID common.TenantID, eventID common.EventID) error
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
