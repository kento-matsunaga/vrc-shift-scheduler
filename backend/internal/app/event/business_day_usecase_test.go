package event_test

import (
	"context"
	"errors"
	"testing"
	"time"

	appevent "github.com/erenoa/vrc-shift-scheduler/backend/internal/app/event"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/event"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/shift"
)

// =====================================================
// MockTxManager
// =====================================================

// MockTxManager is a mock implementation of TxManager for testing
type MockTxManager struct {
	withTxFunc func(ctx context.Context, fn func(context.Context) error) error
}

func (m *MockTxManager) WithTx(ctx context.Context, fn func(context.Context) error) error {
	if m.withTxFunc != nil {
		return m.withTxFunc(ctx, fn)
	}
	// Default implementation: just call the function without actual transaction
	return fn(ctx)
}

// =====================================================
// Additional Mock Repositories for BusinessDay Usecases
// =====================================================

// MockShiftSlotTemplateRepository is a mock implementation of shift.ShiftSlotTemplateRepository
type MockShiftSlotTemplateRepository struct {
	saveFunc          func(ctx context.Context, template *shift.ShiftSlotTemplate) error
	findByIDFunc      func(ctx context.Context, tenantID common.TenantID, templateID common.ShiftSlotTemplateID) (*shift.ShiftSlotTemplate, error)
	findByEventIDFunc func(ctx context.Context, tenantID common.TenantID, eventID common.EventID) ([]*shift.ShiftSlotTemplate, error)
	deleteFunc        func(ctx context.Context, tenantID common.TenantID, templateID common.ShiftSlotTemplateID) error
}

func (m *MockShiftSlotTemplateRepository) Save(ctx context.Context, template *shift.ShiftSlotTemplate) error {
	if m.saveFunc != nil {
		return m.saveFunc(ctx, template)
	}
	return nil
}

func (m *MockShiftSlotTemplateRepository) FindByID(ctx context.Context, tenantID common.TenantID, templateID common.ShiftSlotTemplateID) (*shift.ShiftSlotTemplate, error) {
	if m.findByIDFunc != nil {
		return m.findByIDFunc(ctx, tenantID, templateID)
	}
	return nil, errors.New("not implemented")
}

func (m *MockShiftSlotTemplateRepository) FindByEventID(ctx context.Context, tenantID common.TenantID, eventID common.EventID) ([]*shift.ShiftSlotTemplate, error) {
	if m.findByEventIDFunc != nil {
		return m.findByEventIDFunc(ctx, tenantID, eventID)
	}
	return nil, errors.New("not implemented")
}

func (m *MockShiftSlotTemplateRepository) Delete(ctx context.Context, tenantID common.TenantID, templateID common.ShiftSlotTemplateID) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, tenantID, templateID)
	}
	return nil
}

// MockShiftSlotRepository is a mock implementation of shift.ShiftSlotRepository
type MockShiftSlotRepository struct {
	saveFunc                             func(ctx context.Context, slot *shift.ShiftSlot) error
	findByIDFunc                         func(ctx context.Context, tenantID common.TenantID, slotID shift.SlotID) (*shift.ShiftSlot, error)
	findByBusinessDayIDFunc              func(ctx context.Context, tenantID common.TenantID, businessDayID event.BusinessDayID) ([]*shift.ShiftSlot, error)
	findByInstanceIDFunc                 func(ctx context.Context, tenantID common.TenantID, instanceID shift.InstanceID) ([]*shift.ShiftSlot, error)
	findByBusinessDayIDAndInstanceIDFunc func(ctx context.Context, tenantID common.TenantID, businessDayID event.BusinessDayID, instanceID shift.InstanceID) ([]*shift.ShiftSlot, error)
	deleteFunc                           func(ctx context.Context, tenantID common.TenantID, slotID shift.SlotID) error
}

func (m *MockShiftSlotRepository) Save(ctx context.Context, slot *shift.ShiftSlot) error {
	if m.saveFunc != nil {
		return m.saveFunc(ctx, slot)
	}
	return nil
}

func (m *MockShiftSlotRepository) FindByID(ctx context.Context, tenantID common.TenantID, slotID shift.SlotID) (*shift.ShiftSlot, error) {
	if m.findByIDFunc != nil {
		return m.findByIDFunc(ctx, tenantID, slotID)
	}
	return nil, errors.New("not implemented")
}

func (m *MockShiftSlotRepository) FindByBusinessDayID(ctx context.Context, tenantID common.TenantID, businessDayID event.BusinessDayID) ([]*shift.ShiftSlot, error) {
	if m.findByBusinessDayIDFunc != nil {
		return m.findByBusinessDayIDFunc(ctx, tenantID, businessDayID)
	}
	return nil, errors.New("not implemented")
}

func (m *MockShiftSlotRepository) FindByInstanceID(ctx context.Context, tenantID common.TenantID, instanceID shift.InstanceID) ([]*shift.ShiftSlot, error) {
	if m.findByInstanceIDFunc != nil {
		return m.findByInstanceIDFunc(ctx, tenantID, instanceID)
	}
	return nil, nil
}

func (m *MockShiftSlotRepository) FindByBusinessDayIDAndInstanceID(ctx context.Context, tenantID common.TenantID, businessDayID event.BusinessDayID, instanceID shift.InstanceID) ([]*shift.ShiftSlot, error) {
	if m.findByBusinessDayIDAndInstanceIDFunc != nil {
		return m.findByBusinessDayIDAndInstanceIDFunc(ctx, tenantID, businessDayID, instanceID)
	}
	return nil, nil
}

func (m *MockShiftSlotRepository) Delete(ctx context.Context, tenantID common.TenantID, slotID shift.SlotID) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, tenantID, slotID)
	}
	return nil
}

// MockInstanceRepository is a mock implementation of shift.InstanceRepository
type MockInstanceRepository struct {
	saveFunc                 func(ctx context.Context, instance *shift.Instance) error
	findByIDFunc             func(ctx context.Context, tenantID common.TenantID, instanceID shift.InstanceID) (*shift.Instance, error)
	findByEventIDFunc        func(ctx context.Context, tenantID common.TenantID, eventID common.EventID) ([]*shift.Instance, error)
	findByEventIDAndNameFunc func(ctx context.Context, tenantID common.TenantID, eventID common.EventID, name string) (*shift.Instance, error)
	deleteFunc               func(ctx context.Context, tenantID common.TenantID, instanceID shift.InstanceID) error
}

func (m *MockInstanceRepository) Save(ctx context.Context, instance *shift.Instance) error {
	if m.saveFunc != nil {
		return m.saveFunc(ctx, instance)
	}
	return nil
}

func (m *MockInstanceRepository) FindByID(ctx context.Context, tenantID common.TenantID, instanceID shift.InstanceID) (*shift.Instance, error) {
	if m.findByIDFunc != nil {
		return m.findByIDFunc(ctx, tenantID, instanceID)
	}
	return nil, errors.New("not implemented")
}

func (m *MockInstanceRepository) FindByEventID(ctx context.Context, tenantID common.TenantID, eventID common.EventID) ([]*shift.Instance, error) {
	if m.findByEventIDFunc != nil {
		return m.findByEventIDFunc(ctx, tenantID, eventID)
	}
	return nil, errors.New("not implemented")
}

func (m *MockInstanceRepository) FindByEventIDAndName(ctx context.Context, tenantID common.TenantID, eventID common.EventID, name string) (*shift.Instance, error) {
	if m.findByEventIDAndNameFunc != nil {
		return m.findByEventIDAndNameFunc(ctx, tenantID, eventID, name)
	}
	return nil, nil // Not found is not an error
}

func (m *MockInstanceRepository) Delete(ctx context.Context, tenantID common.TenantID, instanceID shift.InstanceID) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, tenantID, instanceID)
	}
	return nil
}

// =====================================================
// Test Helper Functions
// =====================================================

func createTestBusinessDay(t *testing.T, tenantID common.TenantID, eventID common.EventID) *event.EventBusinessDay {
	t.Helper()
	now := time.Now()
	targetDate := now.AddDate(0, 0, 7) // 7 days from now
	startTime := time.Date(targetDate.Year(), targetDate.Month(), targetDate.Day(), 20, 0, 0, 0, time.Local)
	endTime := time.Date(targetDate.Year(), targetDate.Month(), targetDate.Day(), 22, 0, 0, 0, time.Local)

	bd, err := event.NewEventBusinessDay(now, tenantID, eventID, targetDate, startTime, endTime, event.OccurrenceTypeSpecial, nil)
	if err != nil {
		t.Fatalf("Failed to create test business day: %v", err)
	}
	return bd
}

// =====================================================
// CreateBusinessDayUsecase Tests
// =====================================================

func TestCreateBusinessDayUsecase_Execute_Success(t *testing.T) {
	tenantID := common.NewTenantID()
	eventID := common.NewEventID()
	now := time.Now()

	testEvent, _ := event.NewEvent(now, tenantID, "Test Event", event.EventTypeNormal, "Desc", event.RecurrenceTypeNone, nil, nil, nil, nil)

	eventRepo := &MockEventRepository{
		findByIDFunc: func(ctx context.Context, tid common.TenantID, eid common.EventID) (*event.Event, error) {
			return testEvent, nil
		},
	}

	bdRepo := &MockBusinessDayRepository{
		existsByEventIDAndDate: func(ctx context.Context, tid common.TenantID, eid common.EventID, date time.Time, startTime time.Time) (bool, error) {
			return false, nil // No duplicate
		},
		saveFunc: func(ctx context.Context, bd *event.EventBusinessDay) error {
			return nil
		},
	}

	templateRepo := &MockShiftSlotTemplateRepository{}
	slotRepo := &MockShiftSlotRepository{}

	instanceRepo := &MockInstanceRepository{}
	usecase := appevent.NewCreateBusinessDayUsecase(bdRepo, eventRepo, templateRepo, slotRepo, instanceRepo, &MockTxManager{})

	targetDate := now.AddDate(0, 0, 7)
	startTime := time.Date(targetDate.Year(), targetDate.Month(), targetDate.Day(), 20, 0, 0, 0, time.Local)
	endTime := time.Date(targetDate.Year(), targetDate.Month(), targetDate.Day(), 22, 0, 0, 0, time.Local)

	input := appevent.CreateBusinessDayInput{
		TenantID:       tenantID,
		EventID:        eventID,
		TargetDate:     targetDate,
		StartTime:      startTime,
		EndTime:        endTime,
		OccurrenceType: event.OccurrenceTypeSpecial,
		TemplateID:     nil,
	}

	result, err := usecase.Execute(context.Background(), input)

	if err != nil {
		t.Fatalf("Execute() should succeed, got error: %v", err)
	}

	if result == nil {
		t.Fatal("Result should not be nil")
	}

	if result.TenantID() != tenantID {
		t.Errorf("TenantID mismatch: got %v, want %v", result.TenantID(), tenantID)
	}

	if result.EventID() != eventID {
		t.Errorf("EventID mismatch: got %v, want %v", result.EventID(), eventID)
	}
}

func TestCreateBusinessDayUsecase_Execute_ErrorWhenEventNotFound(t *testing.T) {
	tenantID := common.NewTenantID()
	eventID := common.NewEventID()
	now := time.Now()

	eventRepo := &MockEventRepository{
		findByIDFunc: func(ctx context.Context, tid common.TenantID, eid common.EventID) (*event.Event, error) {
			return nil, common.NewNotFoundError("event", eid.String())
		},
	}

	bdRepo := &MockBusinessDayRepository{}
	templateRepo := &MockShiftSlotTemplateRepository{}
	slotRepo := &MockShiftSlotRepository{}

	instanceRepo := &MockInstanceRepository{}
	usecase := appevent.NewCreateBusinessDayUsecase(bdRepo, eventRepo, templateRepo, slotRepo, instanceRepo, &MockTxManager{})

	targetDate := now.AddDate(0, 0, 7)
	startTime := time.Date(targetDate.Year(), targetDate.Month(), targetDate.Day(), 20, 0, 0, 0, time.Local)
	endTime := time.Date(targetDate.Year(), targetDate.Month(), targetDate.Day(), 22, 0, 0, 0, time.Local)

	input := appevent.CreateBusinessDayInput{
		TenantID:       tenantID,
		EventID:        eventID,
		TargetDate:     targetDate,
		StartTime:      startTime,
		EndTime:        endTime,
		OccurrenceType: event.OccurrenceTypeSpecial,
		TemplateID:     nil,
	}

	_, err := usecase.Execute(context.Background(), input)

	if err == nil {
		t.Fatal("Execute() should fail when event not found")
	}
}

func TestCreateBusinessDayUsecase_Execute_ErrorWhenDuplicate(t *testing.T) {
	tenantID := common.NewTenantID()
	eventID := common.NewEventID()
	now := time.Now()

	testEvent, _ := event.NewEvent(now, tenantID, "Test Event", event.EventTypeNormal, "Desc", event.RecurrenceTypeNone, nil, nil, nil, nil)

	eventRepo := &MockEventRepository{
		findByIDFunc: func(ctx context.Context, tid common.TenantID, eid common.EventID) (*event.Event, error) {
			return testEvent, nil
		},
	}

	bdRepo := &MockBusinessDayRepository{
		existsByEventIDAndDate: func(ctx context.Context, tid common.TenantID, eid common.EventID, date time.Time, startTime time.Time) (bool, error) {
			return true, nil // Already exists
		},
	}

	templateRepo := &MockShiftSlotTemplateRepository{}
	slotRepo := &MockShiftSlotRepository{}

	instanceRepo := &MockInstanceRepository{}
	usecase := appevent.NewCreateBusinessDayUsecase(bdRepo, eventRepo, templateRepo, slotRepo, instanceRepo, &MockTxManager{})

	targetDate := now.AddDate(0, 0, 7)
	startTime := time.Date(targetDate.Year(), targetDate.Month(), targetDate.Day(), 20, 0, 0, 0, time.Local)
	endTime := time.Date(targetDate.Year(), targetDate.Month(), targetDate.Day(), 22, 0, 0, 0, time.Local)

	input := appevent.CreateBusinessDayInput{
		TenantID:       tenantID,
		EventID:        eventID,
		TargetDate:     targetDate,
		StartTime:      startTime,
		EndTime:        endTime,
		OccurrenceType: event.OccurrenceTypeSpecial,
		TemplateID:     nil,
	}

	_, err := usecase.Execute(context.Background(), input)

	if err == nil {
		t.Fatal("Execute() should fail when business day already exists")
	}
}

func TestCreateBusinessDayUsecase_Execute_ErrorWhenSaveFails(t *testing.T) {
	tenantID := common.NewTenantID()
	eventID := common.NewEventID()
	now := time.Now()

	testEvent, _ := event.NewEvent(now, tenantID, "Test Event", event.EventTypeNormal, "Desc", event.RecurrenceTypeNone, nil, nil, nil, nil)

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
			return errors.New("database error")
		},
	}

	templateRepo := &MockShiftSlotTemplateRepository{}
	slotRepo := &MockShiftSlotRepository{}

	instanceRepo := &MockInstanceRepository{}
	usecase := appevent.NewCreateBusinessDayUsecase(bdRepo, eventRepo, templateRepo, slotRepo, instanceRepo, &MockTxManager{})

	targetDate := now.AddDate(0, 0, 7)
	startTime := time.Date(targetDate.Year(), targetDate.Month(), targetDate.Day(), 20, 0, 0, 0, time.Local)
	endTime := time.Date(targetDate.Year(), targetDate.Month(), targetDate.Day(), 22, 0, 0, 0, time.Local)

	input := appevent.CreateBusinessDayInput{
		TenantID:       tenantID,
		EventID:        eventID,
		TargetDate:     targetDate,
		StartTime:      startTime,
		EndTime:        endTime,
		OccurrenceType: event.OccurrenceTypeSpecial,
		TemplateID:     nil,
	}

	_, err := usecase.Execute(context.Background(), input)

	if err == nil {
		t.Fatal("Execute() should fail when save fails")
	}
}

// =====================================================
// ListBusinessDaysUsecase Tests
// =====================================================

func TestListBusinessDaysUsecase_Execute_Success(t *testing.T) {
	tenantID := common.NewTenantID()
	eventID := common.NewEventID()

	testBusinessDays := []*event.EventBusinessDay{
		createTestBusinessDay(t, tenantID, eventID),
		createTestBusinessDay(t, tenantID, eventID),
	}

	bdRepo := &MockBusinessDayRepository{
		findByEventIDFunc: func(ctx context.Context, tid common.TenantID, eid common.EventID) ([]*event.EventBusinessDay, error) {
			return testBusinessDays, nil
		},
	}

	usecase := appevent.NewListBusinessDaysUsecase(bdRepo)

	input := appevent.ListBusinessDaysInput{
		TenantID: tenantID,
		EventID:  eventID,
	}

	result, err := usecase.Execute(context.Background(), input)

	if err != nil {
		t.Fatalf("Execute() should succeed, got error: %v", err)
	}

	if len(result) != 2 {
		t.Errorf("Expected 2 business days, got %d", len(result))
	}
}

func TestListBusinessDaysUsecase_Execute_WithDateRange(t *testing.T) {
	tenantID := common.NewTenantID()
	eventID := common.NewEventID()

	testBusinessDays := []*event.EventBusinessDay{
		createTestBusinessDay(t, tenantID, eventID),
	}

	bdRepo := &MockBusinessDayRepository{
		findByEventIDFunc: func(ctx context.Context, tid common.TenantID, eid common.EventID) ([]*event.EventBusinessDay, error) {
			return testBusinessDays, nil
		},
	}

	// Add findByEventIDAndDateRange
	bdRepoWithRange := &MockBusinessDayRepository{
		findByEventIDFunc: bdRepo.findByEventIDFunc,
	}

	usecase := appevent.NewListBusinessDaysUsecase(bdRepoWithRange)

	now := time.Now()
	startDate := now.AddDate(0, 0, -7)
	endDate := now.AddDate(0, 0, 14)

	input := appevent.ListBusinessDaysInput{
		TenantID:  tenantID,
		EventID:   eventID,
		StartDate: &startDate,
		EndDate:   &endDate,
	}

	_, err := usecase.Execute(context.Background(), input)

	// Note: This will return nil due to mock returning nil for FindByEventIDAndDateRange
	// This just tests that the logic branch works
	if err != nil {
		t.Fatalf("Execute() should succeed, got error: %v", err)
	}
}

func TestListBusinessDaysUsecase_Execute_EmptyList(t *testing.T) {
	tenantID := common.NewTenantID()
	eventID := common.NewEventID()

	bdRepo := &MockBusinessDayRepository{
		findByEventIDFunc: func(ctx context.Context, tid common.TenantID, eid common.EventID) ([]*event.EventBusinessDay, error) {
			return []*event.EventBusinessDay{}, nil
		},
	}

	usecase := appevent.NewListBusinessDaysUsecase(bdRepo)

	input := appevent.ListBusinessDaysInput{
		TenantID: tenantID,
		EventID:  eventID,
	}

	result, err := usecase.Execute(context.Background(), input)

	if err != nil {
		t.Fatalf("Execute() should succeed, got error: %v", err)
	}

	if len(result) != 0 {
		t.Errorf("Expected 0 business days, got %d", len(result))
	}
}

func TestListBusinessDaysUsecase_Execute_ErrorWhenFindFails(t *testing.T) {
	tenantID := common.NewTenantID()
	eventID := common.NewEventID()

	bdRepo := &MockBusinessDayRepository{
		findByEventIDFunc: func(ctx context.Context, tid common.TenantID, eid common.EventID) ([]*event.EventBusinessDay, error) {
			return nil, errors.New("database error")
		},
	}

	usecase := appevent.NewListBusinessDaysUsecase(bdRepo)

	input := appevent.ListBusinessDaysInput{
		TenantID: tenantID,
		EventID:  eventID,
	}

	_, err := usecase.Execute(context.Background(), input)

	if err == nil {
		t.Fatal("Execute() should fail when find fails")
	}
}

// =====================================================
// GetBusinessDayUsecase Tests
// =====================================================

func TestGetBusinessDayUsecase_Execute_Success(t *testing.T) {
	tenantID := common.NewTenantID()
	eventID := common.NewEventID()

	testBusinessDay := createTestBusinessDay(t, tenantID, eventID)

	bdRepo := &MockBusinessDayRepository{}
	// Override FindByID behavior using a wrapper
	bdRepoWithFindByID := &mockBusinessDayRepoWithFindByID{
		MockBusinessDayRepository: bdRepo,
		findByIDFunc: func(ctx context.Context, tid common.TenantID, bdID event.BusinessDayID) (*event.EventBusinessDay, error) {
			return testBusinessDay, nil
		},
	}

	usecase := appevent.NewGetBusinessDayUsecase(bdRepoWithFindByID)

	input := appevent.GetBusinessDayInput{
		TenantID:      tenantID,
		BusinessDayID: testBusinessDay.BusinessDayID(),
	}

	result, err := usecase.Execute(context.Background(), input)

	if err != nil {
		t.Fatalf("Execute() should succeed, got error: %v", err)
	}

	if result.BusinessDayID() != testBusinessDay.BusinessDayID() {
		t.Errorf("BusinessDayID mismatch: got %v, want %v", result.BusinessDayID(), testBusinessDay.BusinessDayID())
	}
}

func TestGetBusinessDayUsecase_Execute_NotFound(t *testing.T) {
	tenantID := common.NewTenantID()
	bdID := event.NewBusinessDayID()

	bdRepo := &MockBusinessDayRepository{}
	bdRepoWithFindByID := &mockBusinessDayRepoWithFindByID{
		MockBusinessDayRepository: bdRepo,
		findByIDFunc: func(ctx context.Context, tid common.TenantID, id event.BusinessDayID) (*event.EventBusinessDay, error) {
			return nil, common.NewNotFoundError("business_day", id.String())
		},
	}

	usecase := appevent.NewGetBusinessDayUsecase(bdRepoWithFindByID)

	input := appevent.GetBusinessDayInput{
		TenantID:      tenantID,
		BusinessDayID: bdID,
	}

	_, err := usecase.Execute(context.Background(), input)

	if err == nil {
		t.Fatal("Execute() should fail when business day not found")
	}
}

// =====================================================
// ApplyTemplateUsecase Tests
// =====================================================

func TestApplyTemplateUsecase_Execute_Success(t *testing.T) {
	tenantID := common.NewTenantID()
	eventID := common.NewEventID()

	testBusinessDay := createTestBusinessDay(t, tenantID, eventID)

	// Create a template with items - but since creating ShiftSlotTemplate is complex,
	// we'll simulate an empty template for this test
	bdRepoWithFindByID := &mockBusinessDayRepoWithFindByID{
		MockBusinessDayRepository: &MockBusinessDayRepository{},
		findByIDFunc: func(ctx context.Context, tid common.TenantID, bdID event.BusinessDayID) (*event.EventBusinessDay, error) {
			return testBusinessDay, nil
		},
	}

	templateID := common.NewShiftSlotTemplateID()

	// Create a mock template
	templateRepo := &MockShiftSlotTemplateRepository{
		findByIDFunc: func(ctx context.Context, tid common.TenantID, tmplID common.ShiftSlotTemplateID) (*shift.ShiftSlotTemplate, error) {
			// Return a template with no items (empty template for simpler testing)
			tmpl, _ := shift.NewShiftSlotTemplate(
				time.Now(),
				tenantID,
				eventID,
				"Test Template",
				"Description",
				[]*shift.ShiftSlotTemplateItem{}, // Empty items
			)
			return tmpl, nil
		},
	}

	slotRepo := &MockShiftSlotRepository{
		saveFunc: func(ctx context.Context, slot *shift.ShiftSlot) error {
			return nil
		},
	}

	instanceRepo := &MockInstanceRepository{}
	usecase := appevent.NewApplyTemplateUsecase(bdRepoWithFindByID, templateRepo, slotRepo, instanceRepo, &MockTxManager{})

	input := appevent.ApplyTemplateInput{
		TenantID:      tenantID,
		BusinessDayID: testBusinessDay.BusinessDayID(),
		TemplateID:    templateID,
	}

	count, err := usecase.Execute(context.Background(), input)

	if err != nil {
		t.Fatalf("Execute() should succeed, got error: %v", err)
	}

	// Template has 0 items in our mock
	if count != 0 {
		t.Errorf("Expected 0 slots created, got %d", count)
	}
}

func TestApplyTemplateUsecase_Execute_ErrorWhenBusinessDayNotFound(t *testing.T) {
	tenantID := common.NewTenantID()
	bdID := event.NewBusinessDayID()
	templateID := common.NewShiftSlotTemplateID()

	bdRepoWithFindByID := &mockBusinessDayRepoWithFindByID{
		MockBusinessDayRepository: &MockBusinessDayRepository{},
		findByIDFunc: func(ctx context.Context, tid common.TenantID, id event.BusinessDayID) (*event.EventBusinessDay, error) {
			return nil, common.NewNotFoundError("business_day", id.String())
		},
	}

	templateRepo := &MockShiftSlotTemplateRepository{}
	slotRepo := &MockShiftSlotRepository{}

	instanceRepo := &MockInstanceRepository{}
	usecase := appevent.NewApplyTemplateUsecase(bdRepoWithFindByID, templateRepo, slotRepo, instanceRepo, &MockTxManager{})

	input := appevent.ApplyTemplateInput{
		TenantID:      tenantID,
		BusinessDayID: bdID,
		TemplateID:    templateID,
	}

	_, err := usecase.Execute(context.Background(), input)

	if err == nil {
		t.Fatal("Execute() should fail when business day not found")
	}
}

func TestApplyTemplateUsecase_Execute_ErrorWhenTemplateNotFound(t *testing.T) {
	tenantID := common.NewTenantID()
	eventID := common.NewEventID()

	testBusinessDay := createTestBusinessDay(t, tenantID, eventID)
	templateID := common.NewShiftSlotTemplateID()

	bdRepoWithFindByID := &mockBusinessDayRepoWithFindByID{
		MockBusinessDayRepository: &MockBusinessDayRepository{},
		findByIDFunc: func(ctx context.Context, tid common.TenantID, bdID event.BusinessDayID) (*event.EventBusinessDay, error) {
			return testBusinessDay, nil
		},
	}

	templateRepo := &MockShiftSlotTemplateRepository{
		findByIDFunc: func(ctx context.Context, tid common.TenantID, tmplID common.ShiftSlotTemplateID) (*shift.ShiftSlotTemplate, error) {
			return nil, common.NewNotFoundError("template", tmplID.String())
		},
	}

	slotRepo := &MockShiftSlotRepository{}

	instanceRepo := &MockInstanceRepository{}
	usecase := appevent.NewApplyTemplateUsecase(bdRepoWithFindByID, templateRepo, slotRepo, instanceRepo, &MockTxManager{})

	input := appevent.ApplyTemplateInput{
		TenantID:      tenantID,
		BusinessDayID: testBusinessDay.BusinessDayID(),
		TemplateID:    templateID,
	}

	_, err := usecase.Execute(context.Background(), input)

	if err == nil {
		t.Fatal("Execute() should fail when template not found")
	}
}

// =====================================================
// Mock Wrapper for FindByID
// =====================================================

type mockBusinessDayRepoWithFindByID struct {
	*MockBusinessDayRepository
	findByIDFunc func(ctx context.Context, tenantID common.TenantID, businessDayID event.BusinessDayID) (*event.EventBusinessDay, error)
}

func (m *mockBusinessDayRepoWithFindByID) FindByID(ctx context.Context, tenantID common.TenantID, businessDayID event.BusinessDayID) (*event.EventBusinessDay, error) {
	if m.findByIDFunc != nil {
		return m.findByIDFunc(ctx, tenantID, businessDayID)
	}
	return nil, errors.New("not implemented")
}
