package schedule_test

import (
	"context"
	"errors"
	"testing"
	"time"

	appschedule "github.com/erenoa/vrc-shift-scheduler/backend/internal/app/schedule"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/schedule"
)

// =====================================================
// Mock Implementations
// =====================================================

// MockDateScheduleRepository is a mock implementation of schedule.DateScheduleRepository
type MockDateScheduleRepository struct {
	saveFunc                        func(ctx context.Context, sch *schedule.DateSchedule) error
	findByIDFunc                    func(ctx context.Context, tenantID common.TenantID, id common.ScheduleID) (*schedule.DateSchedule, error)
	findByTokenFunc                 func(ctx context.Context, token common.PublicToken) (*schedule.DateSchedule, error)
	findByTenantIDFunc              func(ctx context.Context, tenantID common.TenantID) ([]*schedule.DateSchedule, error)
	upsertResponseFunc              func(ctx context.Context, response *schedule.DateScheduleResponse) error
	findResponsesByScheduleIDFunc   func(ctx context.Context, scheduleID common.ScheduleID) ([]*schedule.DateScheduleResponse, error)
	findCandidatesByScheduleIDFunc  func(ctx context.Context, scheduleID common.ScheduleID) ([]*schedule.CandidateDate, error)
	saveGroupAssignmentsFunc        func(ctx context.Context, scheduleID common.ScheduleID, assignments []*schedule.ScheduleGroupAssignment) error
	findGroupAssignmentsByScheduleIDFunc func(ctx context.Context, scheduleID common.ScheduleID) ([]*schedule.ScheduleGroupAssignment, error)
}

func (m *MockDateScheduleRepository) Save(ctx context.Context, sch *schedule.DateSchedule) error {
	if m.saveFunc != nil {
		return m.saveFunc(ctx, sch)
	}
	return nil
}

func (m *MockDateScheduleRepository) FindByID(ctx context.Context, tenantID common.TenantID, id common.ScheduleID) (*schedule.DateSchedule, error) {
	if m.findByIDFunc != nil {
		return m.findByIDFunc(ctx, tenantID, id)
	}
	return nil, errors.New("not implemented")
}

func (m *MockDateScheduleRepository) FindByToken(ctx context.Context, token common.PublicToken) (*schedule.DateSchedule, error) {
	if m.findByTokenFunc != nil {
		return m.findByTokenFunc(ctx, token)
	}
	return nil, errors.New("not implemented")
}

func (m *MockDateScheduleRepository) FindByTenantID(ctx context.Context, tenantID common.TenantID) ([]*schedule.DateSchedule, error) {
	if m.findByTenantIDFunc != nil {
		return m.findByTenantIDFunc(ctx, tenantID)
	}
	return nil, errors.New("not implemented")
}

func (m *MockDateScheduleRepository) UpsertResponse(ctx context.Context, response *schedule.DateScheduleResponse) error {
	if m.upsertResponseFunc != nil {
		return m.upsertResponseFunc(ctx, response)
	}
	return nil
}

func (m *MockDateScheduleRepository) FindResponsesByScheduleID(ctx context.Context, scheduleID common.ScheduleID) ([]*schedule.DateScheduleResponse, error) {
	if m.findResponsesByScheduleIDFunc != nil {
		return m.findResponsesByScheduleIDFunc(ctx, scheduleID)
	}
	return []*schedule.DateScheduleResponse{}, nil
}

func (m *MockDateScheduleRepository) FindCandidatesByScheduleID(ctx context.Context, scheduleID common.ScheduleID) ([]*schedule.CandidateDate, error) {
	if m.findCandidatesByScheduleIDFunc != nil {
		return m.findCandidatesByScheduleIDFunc(ctx, scheduleID)
	}
	return nil, errors.New("not implemented")
}

func (m *MockDateScheduleRepository) SaveGroupAssignments(ctx context.Context, scheduleID common.ScheduleID, assignments []*schedule.ScheduleGroupAssignment) error {
	if m.saveGroupAssignmentsFunc != nil {
		return m.saveGroupAssignmentsFunc(ctx, scheduleID, assignments)
	}
	return nil
}

func (m *MockDateScheduleRepository) FindGroupAssignmentsByScheduleID(ctx context.Context, scheduleID common.ScheduleID) ([]*schedule.ScheduleGroupAssignment, error) {
	if m.findGroupAssignmentsByScheduleIDFunc != nil {
		return m.findGroupAssignmentsByScheduleIDFunc(ctx, scheduleID)
	}
	return []*schedule.ScheduleGroupAssignment{}, nil
}

// MockClock is a mock implementation of services.Clock
type MockClock struct {
	nowFunc func() time.Time
}

func (m *MockClock) Now() time.Time {
	if m.nowFunc != nil {
		return m.nowFunc()
	}
	return time.Now()
}

// =====================================================
// Test Helper Functions
// =====================================================

func createTestSchedule(t *testing.T, tenantID common.TenantID) *schedule.DateSchedule {
	t.Helper()
	now := time.Now()
	scheduleID := common.NewScheduleID()

	// Create candidates
	candidates := make([]*schedule.CandidateDate, 2)
	for i := 0; i < 2; i++ {
		candidate, err := schedule.NewCandidateDate(now, scheduleID, now.AddDate(0, 0, i+1), nil, nil, i)
		if err != nil {
			t.Fatalf("Failed to create test candidate: %v", err)
		}
		candidates[i] = candidate
	}

	sch, err := schedule.NewDateSchedule(now, scheduleID, tenantID, "Test Schedule", "Test Description", nil, candidates, nil)
	if err != nil {
		t.Fatalf("Failed to create test schedule: %v", err)
	}
	return sch
}

// =====================================================
// CreateScheduleUsecase Tests
// =====================================================

func TestCreateScheduleUsecase_Execute_Success(t *testing.T) {
	tenantID := common.NewTenantID()

	repo := &MockDateScheduleRepository{
		saveFunc: func(ctx context.Context, sch *schedule.DateSchedule) error {
			return nil
		},
	}

	clock := &MockClock{nowFunc: func() time.Time { return time.Now() }}

	usecase := appschedule.NewCreateScheduleUsecase(repo, clock)

	now := time.Now()
	input := appschedule.CreateScheduleInput{
		TenantID:    tenantID.String(),
		Title:       "Test Schedule",
		Description: "Test Description",
		Candidates: []appschedule.CandidateInput{
			{Date: now.AddDate(0, 0, 1)},
			{Date: now.AddDate(0, 0, 2)},
		},
	}

	result, err := usecase.Execute(context.Background(), input)

	if err != nil {
		t.Fatalf("Execute() should succeed, got error: %v", err)
	}

	if result == nil {
		t.Fatal("Result should not be nil")
	}

	if result.Title != "Test Schedule" {
		t.Errorf("Title mismatch: got %v, want 'Test Schedule'", result.Title)
	}

	if len(result.Candidates) != 2 {
		t.Errorf("Expected 2 candidates, got %d", len(result.Candidates))
	}
}

func TestCreateScheduleUsecase_Execute_ErrorWhenInvalidTenantID(t *testing.T) {
	repo := &MockDateScheduleRepository{}
	clock := &MockClock{nowFunc: func() time.Time { return time.Now() }}

	usecase := appschedule.NewCreateScheduleUsecase(repo, clock)

	input := appschedule.CreateScheduleInput{
		TenantID:    "invalid-ulid",
		Title:       "Test Schedule",
		Description: "Test Description",
		Candidates:  []appschedule.CandidateInput{},
	}

	_, err := usecase.Execute(context.Background(), input)

	if err == nil {
		t.Fatal("Execute() should fail when tenant ID is invalid")
	}
}

func TestCreateScheduleUsecase_Execute_ErrorWhenSaveFails(t *testing.T) {
	tenantID := common.NewTenantID()

	repo := &MockDateScheduleRepository{
		saveFunc: func(ctx context.Context, sch *schedule.DateSchedule) error {
			return errors.New("database error")
		},
	}

	clock := &MockClock{nowFunc: func() time.Time { return time.Now() }}

	usecase := appschedule.NewCreateScheduleUsecase(repo, clock)

	now := time.Now()
	input := appschedule.CreateScheduleInput{
		TenantID:    tenantID.String(),
		Title:       "Test Schedule",
		Description: "Test Description",
		Candidates: []appschedule.CandidateInput{
			{Date: now.AddDate(0, 0, 1)},
		},
	}

	_, err := usecase.Execute(context.Background(), input)

	if err == nil {
		t.Fatal("Execute() should fail when save fails")
	}
}

// =====================================================
// ListSchedulesUsecase Tests
// =====================================================

func TestListSchedulesUsecase_Execute_Success(t *testing.T) {
	tenantID := common.NewTenantID()

	testSchedules := []*schedule.DateSchedule{
		createTestSchedule(t, tenantID),
		createTestSchedule(t, tenantID),
	}

	repo := &MockDateScheduleRepository{
		findByTenantIDFunc: func(ctx context.Context, tid common.TenantID) ([]*schedule.DateSchedule, error) {
			return testSchedules, nil
		},
		findResponsesByScheduleIDFunc: func(ctx context.Context, scheduleID common.ScheduleID) ([]*schedule.DateScheduleResponse, error) {
			return []*schedule.DateScheduleResponse{}, nil
		},
	}

	usecase := appschedule.NewListSchedulesUsecase(repo)

	input := appschedule.ListSchedulesInput{
		TenantID: tenantID.String(),
	}

	result, err := usecase.Execute(context.Background(), input)

	if err != nil {
		t.Fatalf("Execute() should succeed, got error: %v", err)
	}

	if len(result.Schedules) != 2 {
		t.Errorf("Expected 2 schedules, got %d", len(result.Schedules))
	}
}

func TestListSchedulesUsecase_Execute_EmptyList(t *testing.T) {
	tenantID := common.NewTenantID()

	repo := &MockDateScheduleRepository{
		findByTenantIDFunc: func(ctx context.Context, tid common.TenantID) ([]*schedule.DateSchedule, error) {
			return []*schedule.DateSchedule{}, nil
		},
	}

	usecase := appschedule.NewListSchedulesUsecase(repo)

	input := appschedule.ListSchedulesInput{
		TenantID: tenantID.String(),
	}

	result, err := usecase.Execute(context.Background(), input)

	if err != nil {
		t.Fatalf("Execute() should succeed, got error: %v", err)
	}

	if len(result.Schedules) != 0 {
		t.Errorf("Expected 0 schedules, got %d", len(result.Schedules))
	}
}

func TestListSchedulesUsecase_Execute_ErrorWhenInvalidTenantID(t *testing.T) {
	repo := &MockDateScheduleRepository{}

	usecase := appschedule.NewListSchedulesUsecase(repo)

	input := appschedule.ListSchedulesInput{
		TenantID: "invalid-ulid",
	}

	_, err := usecase.Execute(context.Background(), input)

	if err == nil {
		t.Fatal("Execute() should fail when tenant ID is invalid")
	}
}

func TestListSchedulesUsecase_Execute_ErrorWhenFindFails(t *testing.T) {
	tenantID := common.NewTenantID()

	repo := &MockDateScheduleRepository{
		findByTenantIDFunc: func(ctx context.Context, tid common.TenantID) ([]*schedule.DateSchedule, error) {
			return nil, errors.New("database error")
		},
	}

	usecase := appschedule.NewListSchedulesUsecase(repo)

	input := appschedule.ListSchedulesInput{
		TenantID: tenantID.String(),
	}

	_, err := usecase.Execute(context.Background(), input)

	if err == nil {
		t.Fatal("Execute() should fail when find fails")
	}
}

// =====================================================
// GetScheduleUsecase Tests
// =====================================================

func TestGetScheduleUsecase_Execute_Success(t *testing.T) {
	tenantID := common.NewTenantID()
	testSchedule := createTestSchedule(t, tenantID)

	repo := &MockDateScheduleRepository{
		findByIDFunc: func(ctx context.Context, tid common.TenantID, id common.ScheduleID) (*schedule.DateSchedule, error) {
			return testSchedule, nil
		},
		findGroupAssignmentsByScheduleIDFunc: func(ctx context.Context, scheduleID common.ScheduleID) ([]*schedule.ScheduleGroupAssignment, error) {
			return []*schedule.ScheduleGroupAssignment{}, nil
		},
	}

	usecase := appschedule.NewGetScheduleUsecase(repo)

	input := appschedule.GetScheduleInput{
		TenantID:   tenantID.String(),
		ScheduleID: testSchedule.ScheduleID().String(),
	}

	result, err := usecase.Execute(context.Background(), input)

	if err != nil {
		t.Fatalf("Execute() should succeed, got error: %v", err)
	}

	if result.ScheduleID != testSchedule.ScheduleID().String() {
		t.Errorf("ScheduleID mismatch: got %v, want %v", result.ScheduleID, testSchedule.ScheduleID().String())
	}

	if result.Title != "Test Schedule" {
		t.Errorf("Title mismatch: got %v, want 'Test Schedule'", result.Title)
	}
}

func TestGetScheduleUsecase_Execute_NotFound(t *testing.T) {
	tenantID := common.NewTenantID()
	scheduleID := common.NewScheduleID()

	repo := &MockDateScheduleRepository{
		findByIDFunc: func(ctx context.Context, tid common.TenantID, id common.ScheduleID) (*schedule.DateSchedule, error) {
			return nil, common.NewNotFoundError("schedule", id.String())
		},
	}

	usecase := appschedule.NewGetScheduleUsecase(repo)

	input := appschedule.GetScheduleInput{
		TenantID:   tenantID.String(),
		ScheduleID: scheduleID.String(),
	}

	_, err := usecase.Execute(context.Background(), input)

	if err == nil {
		t.Fatal("Execute() should fail when schedule not found")
	}
}

func TestGetScheduleUsecase_Execute_ErrorWhenInvalidTenantID(t *testing.T) {
	repo := &MockDateScheduleRepository{}

	usecase := appschedule.NewGetScheduleUsecase(repo)

	input := appschedule.GetScheduleInput{
		TenantID:   "invalid-ulid",
		ScheduleID: common.NewScheduleID().String(),
	}

	_, err := usecase.Execute(context.Background(), input)

	if err == nil {
		t.Fatal("Execute() should fail when tenant ID is invalid")
	}
}

func TestGetScheduleUsecase_Execute_ErrorWhenInvalidScheduleID(t *testing.T) {
	tenantID := common.NewTenantID()
	repo := &MockDateScheduleRepository{}

	usecase := appschedule.NewGetScheduleUsecase(repo)

	input := appschedule.GetScheduleInput{
		TenantID:   tenantID.String(),
		ScheduleID: "invalid-ulid",
	}

	_, err := usecase.Execute(context.Background(), input)

	if err == nil {
		t.Fatal("Execute() should fail when schedule ID is invalid")
	}
}

// =====================================================
// CloseScheduleUsecase Tests
// =====================================================

func TestCloseScheduleUsecase_Execute_Success(t *testing.T) {
	tenantID := common.NewTenantID()
	testSchedule := createTestSchedule(t, tenantID)

	repo := &MockDateScheduleRepository{
		findByIDFunc: func(ctx context.Context, tid common.TenantID, id common.ScheduleID) (*schedule.DateSchedule, error) {
			return testSchedule, nil
		},
		saveFunc: func(ctx context.Context, sch *schedule.DateSchedule) error {
			return nil
		},
	}

	clock := &MockClock{nowFunc: func() time.Time { return time.Now() }}

	usecase := appschedule.NewCloseScheduleUsecase(repo, clock)

	input := appschedule.CloseScheduleInput{
		TenantID:   tenantID.String(),
		ScheduleID: testSchedule.ScheduleID().String(),
	}

	result, err := usecase.Execute(context.Background(), input)

	if err != nil {
		t.Fatalf("Execute() should succeed, got error: %v", err)
	}

	if result.ScheduleID != testSchedule.ScheduleID().String() {
		t.Errorf("ScheduleID mismatch: got %v, want %v", result.ScheduleID, testSchedule.ScheduleID().String())
	}
}

func TestCloseScheduleUsecase_Execute_NotFound(t *testing.T) {
	tenantID := common.NewTenantID()
	scheduleID := common.NewScheduleID()

	repo := &MockDateScheduleRepository{
		findByIDFunc: func(ctx context.Context, tid common.TenantID, id common.ScheduleID) (*schedule.DateSchedule, error) {
			return nil, common.NewNotFoundError("schedule", id.String())
		},
	}

	clock := &MockClock{nowFunc: func() time.Time { return time.Now() }}

	usecase := appschedule.NewCloseScheduleUsecase(repo, clock)

	input := appschedule.CloseScheduleInput{
		TenantID:   tenantID.String(),
		ScheduleID: scheduleID.String(),
	}

	_, err := usecase.Execute(context.Background(), input)

	if err == nil {
		t.Fatal("Execute() should fail when schedule not found")
	}
}

func TestCloseScheduleUsecase_Execute_ErrorWhenSaveFails(t *testing.T) {
	tenantID := common.NewTenantID()
	testSchedule := createTestSchedule(t, tenantID)

	repo := &MockDateScheduleRepository{
		findByIDFunc: func(ctx context.Context, tid common.TenantID, id common.ScheduleID) (*schedule.DateSchedule, error) {
			return testSchedule, nil
		},
		saveFunc: func(ctx context.Context, sch *schedule.DateSchedule) error {
			return errors.New("database error")
		},
	}

	clock := &MockClock{nowFunc: func() time.Time { return time.Now() }}

	usecase := appschedule.NewCloseScheduleUsecase(repo, clock)

	input := appschedule.CloseScheduleInput{
		TenantID:   tenantID.String(),
		ScheduleID: testSchedule.ScheduleID().String(),
	}

	_, err := usecase.Execute(context.Background(), input)

	if err == nil {
		t.Fatal("Execute() should fail when save fails")
	}
}

// =====================================================
// DecideScheduleUsecase Tests
// =====================================================

func TestDecideScheduleUsecase_Execute_Success(t *testing.T) {
	tenantID := common.NewTenantID()
	testSchedule := createTestSchedule(t, tenantID)

	// Get the first candidate ID from the test schedule
	candidateID := testSchedule.Candidates()[0].CandidateID()

	repo := &MockDateScheduleRepository{
		findByIDFunc: func(ctx context.Context, tid common.TenantID, id common.ScheduleID) (*schedule.DateSchedule, error) {
			return testSchedule, nil
		},
		saveFunc: func(ctx context.Context, sch *schedule.DateSchedule) error {
			return nil
		},
	}

	clock := &MockClock{nowFunc: func() time.Time { return time.Now() }}

	usecase := appschedule.NewDecideScheduleUsecase(repo, clock)

	input := appschedule.DecideScheduleInput{
		TenantID:    tenantID.String(),
		ScheduleID:  testSchedule.ScheduleID().String(),
		CandidateID: candidateID.String(),
	}

	result, err := usecase.Execute(context.Background(), input)

	if err != nil {
		t.Fatalf("Execute() should succeed, got error: %v", err)
	}

	if result.ScheduleID != testSchedule.ScheduleID().String() {
		t.Errorf("ScheduleID mismatch: got %v, want %v", result.ScheduleID, testSchedule.ScheduleID().String())
	}

	if result.DecidedCandidateID != candidateID.String() {
		t.Errorf("DecidedCandidateID mismatch: got %v, want %v", result.DecidedCandidateID, candidateID.String())
	}
}

func TestDecideScheduleUsecase_Execute_NotFound(t *testing.T) {
	tenantID := common.NewTenantID()
	scheduleID := common.NewScheduleID()
	candidateID := common.NewCandidateID()

	repo := &MockDateScheduleRepository{
		findByIDFunc: func(ctx context.Context, tid common.TenantID, id common.ScheduleID) (*schedule.DateSchedule, error) {
			return nil, common.NewNotFoundError("schedule", id.String())
		},
	}

	clock := &MockClock{nowFunc: func() time.Time { return time.Now() }}

	usecase := appschedule.NewDecideScheduleUsecase(repo, clock)

	input := appschedule.DecideScheduleInput{
		TenantID:    tenantID.String(),
		ScheduleID:  scheduleID.String(),
		CandidateID: candidateID.String(),
	}

	_, err := usecase.Execute(context.Background(), input)

	if err == nil {
		t.Fatal("Execute() should fail when schedule not found")
	}
}

func TestDecideScheduleUsecase_Execute_ErrorWhenInvalidCandidateID(t *testing.T) {
	tenantID := common.NewTenantID()
	testSchedule := createTestSchedule(t, tenantID)

	repo := &MockDateScheduleRepository{
		findByIDFunc: func(ctx context.Context, tid common.TenantID, id common.ScheduleID) (*schedule.DateSchedule, error) {
			return testSchedule, nil
		},
	}

	clock := &MockClock{nowFunc: func() time.Time { return time.Now() }}

	usecase := appschedule.NewDecideScheduleUsecase(repo, clock)

	// Use a candidate ID that doesn't exist in the schedule
	input := appschedule.DecideScheduleInput{
		TenantID:    tenantID.String(),
		ScheduleID:  testSchedule.ScheduleID().String(),
		CandidateID: common.NewCandidateID().String(), // Different candidate
	}

	_, err := usecase.Execute(context.Background(), input)

	if err == nil {
		t.Fatal("Execute() should fail when candidate ID is not in the schedule")
	}
}
