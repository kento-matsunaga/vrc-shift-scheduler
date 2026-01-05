package attendance_test

import (
	"context"
	"errors"
	"testing"
	"time"

	appattendance "github.com/erenoa/vrc-shift-scheduler/backend/internal/app/attendance"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/attendance"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/role"
)

// =====================================================
// Mock Clock
// =====================================================

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
// Mock Repository
// =====================================================

type MockAttendanceCollectionRepository struct {
	saveFunc                 func(ctx context.Context, c *attendance.AttendanceCollection) error
	findByIDFunc             func(ctx context.Context, tenantID common.TenantID, collectionID common.CollectionID) (*attendance.AttendanceCollection, error)
	findByPublicTokenFunc    func(ctx context.Context, token common.PublicToken) (*attendance.AttendanceCollection, error)
	saveTargetDatesFunc      func(ctx context.Context, collectionID common.CollectionID, dates []*attendance.TargetDate) error
	saveGroupAssignmentsFunc func(ctx context.Context, collectionID common.CollectionID, assignments []*attendance.CollectionGroupAssignment) error
}

func (m *MockAttendanceCollectionRepository) Save(ctx context.Context, c *attendance.AttendanceCollection) error {
	if m.saveFunc != nil {
		return m.saveFunc(ctx, c)
	}
	return nil
}

func (m *MockAttendanceCollectionRepository) FindByID(ctx context.Context, tenantID common.TenantID, collectionID common.CollectionID) (*attendance.AttendanceCollection, error) {
	if m.findByIDFunc != nil {
		return m.findByIDFunc(ctx, tenantID, collectionID)
	}
	return nil, errors.New("not implemented")
}

func (m *MockAttendanceCollectionRepository) FindByPublicToken(ctx context.Context, token common.PublicToken) (*attendance.AttendanceCollection, error) {
	if m.findByPublicTokenFunc != nil {
		return m.findByPublicTokenFunc(ctx, token)
	}
	return nil, errors.New("not implemented")
}

func (m *MockAttendanceCollectionRepository) FindByToken(ctx context.Context, token common.PublicToken) (*attendance.AttendanceCollection, error) {
	if m.findByPublicTokenFunc != nil {
		return m.findByPublicTokenFunc(ctx, token)
	}
	return nil, errors.New("not implemented")
}

func (m *MockAttendanceCollectionRepository) SaveTargetDates(ctx context.Context, collectionID common.CollectionID, dates []*attendance.TargetDate) error {
	if m.saveTargetDatesFunc != nil {
		return m.saveTargetDatesFunc(ctx, collectionID, dates)
	}
	return nil
}

func (m *MockAttendanceCollectionRepository) SaveGroupAssignments(ctx context.Context, collectionID common.CollectionID, assignments []*attendance.CollectionGroupAssignment) error {
	if m.saveGroupAssignmentsFunc != nil {
		return m.saveGroupAssignmentsFunc(ctx, collectionID, assignments)
	}
	return nil
}

func (m *MockAttendanceCollectionRepository) FindByTenantID(ctx context.Context, tenantID common.TenantID) ([]*attendance.AttendanceCollection, error) {
	return nil, nil
}

func (m *MockAttendanceCollectionRepository) UpsertResponse(ctx context.Context, response *attendance.AttendanceResponse) error {
	return nil
}

func (m *MockAttendanceCollectionRepository) FindResponsesByCollectionID(ctx context.Context, collectionID common.CollectionID) ([]*attendance.AttendanceResponse, error) {
	return nil, nil
}

func (m *MockAttendanceCollectionRepository) FindResponsesByMemberID(ctx context.Context, tenantID common.TenantID, memberID common.MemberID) ([]*attendance.AttendanceResponse, error) {
	return nil, nil
}

func (m *MockAttendanceCollectionRepository) FindTargetDatesByCollectionID(ctx context.Context, collectionID common.CollectionID) ([]*attendance.TargetDate, error) {
	return nil, nil
}

func (m *MockAttendanceCollectionRepository) FindGroupAssignmentsByCollectionID(ctx context.Context, collectionID common.CollectionID) ([]*attendance.CollectionGroupAssignment, error) {
	return nil, nil
}

func (m *MockAttendanceCollectionRepository) SaveRoleAssignments(ctx context.Context, collectionID common.CollectionID, assignments []*attendance.CollectionRoleAssignment) error {
	return nil
}

func (m *MockAttendanceCollectionRepository) FindRoleAssignmentsByCollectionID(ctx context.Context, collectionID common.CollectionID) ([]*attendance.CollectionRoleAssignment, error) {
	return nil, nil
}

// =====================================================
// Mock Role Repository
// =====================================================

type MockRoleRepository struct {
	findByIDFunc func(ctx context.Context, tenantID common.TenantID, roleID common.RoleID) (*role.Role, error)
}

func (m *MockRoleRepository) Save(ctx context.Context, r *role.Role) error {
	return nil
}

func (m *MockRoleRepository) FindByID(ctx context.Context, tenantID common.TenantID, roleID common.RoleID) (*role.Role, error) {
	if m.findByIDFunc != nil {
		return m.findByIDFunc(ctx, tenantID, roleID)
	}
	// Default: return a mock role (role exists)
	return &role.Role{}, nil
}

func (m *MockRoleRepository) FindByTenantID(ctx context.Context, tenantID common.TenantID) ([]*role.Role, error) {
	return nil, nil
}

func (m *MockRoleRepository) Delete(ctx context.Context, tenantID common.TenantID, roleID common.RoleID) error {
	return nil
}

// =====================================================
// CreateCollectionUsecase Tests
// =====================================================

func TestCreateCollectionUsecase_Execute_Success(t *testing.T) {
	tenantID := common.NewTenantID()
	now := time.Now()

	clock := &MockClock{
		nowFunc: func() time.Time { return now },
	}

	repo := &MockAttendanceCollectionRepository{
		saveFunc: func(ctx context.Context, c *attendance.AttendanceCollection) error {
			return nil
		},
	}

	roleRepo := &MockRoleRepository{}

	usecase := appattendance.NewCreateCollectionUsecase(repo, roleRepo, clock)

	input := appattendance.CreateCollectionInput{
		TenantID:    tenantID.String(),
		Title:       "12月イベント出欠確認",
		Description: "12月のイベントへの参加可否を回答してください",
		TargetType:  "event",
		TargetID:    "event-123",
	}

	result, err := usecase.Execute(context.Background(), input)

	if err != nil {
		t.Fatalf("Execute() should succeed, got error: %v", err)
	}

	if result == nil {
		t.Fatal("Result should not be nil")
	}

	if result.Title != "12月イベント出欠確認" {
		t.Errorf("Title mismatch: got %v, want '12月イベント出欠確認'", result.Title)
	}

	if result.Status != "open" {
		t.Errorf("Status should be 'open': got %v", result.Status)
	}

	if result.PublicToken == "" {
		t.Error("PublicToken should be set")
	}
}

func TestCreateCollectionUsecase_Execute_WithDeadline(t *testing.T) {
	tenantID := common.NewTenantID()
	now := time.Now()
	deadline := now.Add(7 * 24 * time.Hour)

	clock := &MockClock{
		nowFunc: func() time.Time { return now },
	}

	repo := &MockAttendanceCollectionRepository{
		saveFunc: func(ctx context.Context, c *attendance.AttendanceCollection) error {
			return nil
		},
	}

	roleRepo := &MockRoleRepository{}

	usecase := appattendance.NewCreateCollectionUsecase(repo, roleRepo, clock)

	input := appattendance.CreateCollectionInput{
		TenantID:    tenantID.String(),
		Title:       "締切付き出欠確認",
		Description: "",
		TargetType:  "event",
		TargetID:    "",
		Deadline:    &deadline,
	}

	result, err := usecase.Execute(context.Background(), input)

	if err != nil {
		t.Fatalf("Execute() should succeed, got error: %v", err)
	}

	if result.Deadline == nil {
		t.Error("Deadline should be set")
	}
}

func TestCreateCollectionUsecase_Execute_ErrorWhenTitleEmpty(t *testing.T) {
	tenantID := common.NewTenantID()
	now := time.Now()

	clock := &MockClock{
		nowFunc: func() time.Time { return now },
	}

	repo := &MockAttendanceCollectionRepository{}
	roleRepo := &MockRoleRepository{}

	usecase := appattendance.NewCreateCollectionUsecase(repo, roleRepo, clock)

	input := appattendance.CreateCollectionInput{
		TenantID:    tenantID.String(),
		Title:       "", // Empty title
		Description: "Description",
		TargetType:  "event",
		TargetID:    "",
	}

	_, err := usecase.Execute(context.Background(), input)

	if err == nil {
		t.Fatal("Execute() should fail when title is empty")
	}
}

func TestCreateCollectionUsecase_Execute_ErrorWhenInvalidTargetType(t *testing.T) {
	tenantID := common.NewTenantID()
	now := time.Now()

	clock := &MockClock{
		nowFunc: func() time.Time { return now },
	}

	repo := &MockAttendanceCollectionRepository{}
	roleRepo := &MockRoleRepository{}

	usecase := appattendance.NewCreateCollectionUsecase(repo, roleRepo, clock)

	input := appattendance.CreateCollectionInput{
		TenantID:    tenantID.String(),
		Title:       "Test Title",
		Description: "Description",
		TargetType:  "invalid_type", // Invalid target type
		TargetID:    "",
	}

	_, err := usecase.Execute(context.Background(), input)

	if err == nil {
		t.Fatal("Execute() should fail when target type is invalid")
	}
}

func TestCreateCollectionUsecase_Execute_ErrorWhenSaveFails(t *testing.T) {
	tenantID := common.NewTenantID()
	now := time.Now()

	clock := &MockClock{
		nowFunc: func() time.Time { return now },
	}

	repo := &MockAttendanceCollectionRepository{
		saveFunc: func(ctx context.Context, c *attendance.AttendanceCollection) error {
			return errors.New("database error")
		},
	}
	roleRepo := &MockRoleRepository{}

	usecase := appattendance.NewCreateCollectionUsecase(repo, roleRepo, clock)

	input := appattendance.CreateCollectionInput{
		TenantID:    tenantID.String(),
		Title:       "Test Title",
		Description: "Description",
		TargetType:  "event",
		TargetID:    "",
	}

	_, err := usecase.Execute(context.Background(), input)

	if err == nil {
		t.Fatal("Execute() should fail when save fails")
	}
}

func TestCreateCollectionUsecase_Execute_ErrorWhenInvalidTenantID(t *testing.T) {
	now := time.Now()

	clock := &MockClock{
		nowFunc: func() time.Time { return now },
	}

	repo := &MockAttendanceCollectionRepository{}
	roleRepo := &MockRoleRepository{}

	usecase := appattendance.NewCreateCollectionUsecase(repo, roleRepo, clock)

	input := appattendance.CreateCollectionInput{
		TenantID:    "invalid-tenant-id", // Invalid tenant ID format
		Title:       "Test Title",
		Description: "Description",
		TargetType:  "event",
		TargetID:    "",
	}

	_, err := usecase.Execute(context.Background(), input)

	if err == nil {
		t.Fatal("Execute() should fail when tenant ID is invalid")
	}
}

func TestCreateCollectionUsecase_Execute_WithTargetDates(t *testing.T) {
	tenantID := common.NewTenantID()
	now := time.Now()

	targetDates := []time.Time{
		now.Add(24 * time.Hour),
		now.Add(48 * time.Hour),
	}

	clock := &MockClock{
		nowFunc: func() time.Time { return now },
	}

	var savedTargetDates []*attendance.TargetDate

	repo := &MockAttendanceCollectionRepository{
		saveFunc: func(ctx context.Context, c *attendance.AttendanceCollection) error {
			return nil
		},
		saveTargetDatesFunc: func(ctx context.Context, collectionID common.CollectionID, dates []*attendance.TargetDate) error {
			savedTargetDates = dates
			return nil
		},
	}
	roleRepo := &MockRoleRepository{}

	usecase := appattendance.NewCreateCollectionUsecase(repo, roleRepo, clock)

	input := appattendance.CreateCollectionInput{
		TenantID:    tenantID.String(),
		Title:       "複数日程の出欠確認",
		Description: "",
		TargetType:  "event",
		TargetID:    "",
		TargetDates: targetDates,
	}

	_, err := usecase.Execute(context.Background(), input)

	if err != nil {
		t.Fatalf("Execute() should succeed, got error: %v", err)
	}

	if len(savedTargetDates) != 2 {
		t.Errorf("Expected 2 target dates to be saved, got %d", len(savedTargetDates))
	}
}

func TestCreateCollectionUsecase_Execute_BusinessDayTarget(t *testing.T) {
	tenantID := common.NewTenantID()
	now := time.Now()

	clock := &MockClock{
		nowFunc: func() time.Time { return now },
	}

	repo := &MockAttendanceCollectionRepository{
		saveFunc: func(ctx context.Context, c *attendance.AttendanceCollection) error {
			return nil
		},
	}
	roleRepo := &MockRoleRepository{}

	usecase := appattendance.NewCreateCollectionUsecase(repo, roleRepo, clock)

	input := appattendance.CreateCollectionInput{
		TenantID:    tenantID.String(),
		Title:       "営業日出欠確認",
		Description: "",
		TargetType:  "business_day",
		TargetID:    "bd-456",
	}

	result, err := usecase.Execute(context.Background(), input)

	if err != nil {
		t.Fatalf("Execute() should succeed, got error: %v", err)
	}

	if result.TargetType != "business_day" {
		t.Errorf("TargetType mismatch: got %v, want 'business_day'", result.TargetType)
	}

	if result.TargetID != "bd-456" {
		t.Errorf("TargetID mismatch: got %v, want 'bd-456'", result.TargetID)
	}
}

// =====================================================
// DeleteCollectionUsecase Tests
// =====================================================

func createTestCollection(t *testing.T, tenantID common.TenantID) *attendance.AttendanceCollection {
	t.Helper()
	now := time.Now()

	collection, err := attendance.NewAttendanceCollection(
		now,
		tenantID,
		"Test Collection",
		"Test Description",
		attendance.TargetTypeEvent,
		"event-123",
		nil,
	)
	if err != nil {
		t.Fatalf("Failed to create test collection: %v", err)
	}
	return collection
}

func TestDeleteCollectionUsecase_Execute_Success(t *testing.T) {
	tenantID := common.NewTenantID()
	testCollection := createTestCollection(t, tenantID)

	repo := &MockAttendanceCollectionRepository{
		findByIDFunc: func(ctx context.Context, tid common.TenantID, cid common.CollectionID) (*attendance.AttendanceCollection, error) {
			return testCollection, nil
		},
		saveFunc: func(ctx context.Context, c *attendance.AttendanceCollection) error {
			return nil
		},
	}

	clock := &MockClock{nowFunc: func() time.Time { return time.Now() }}

	usecase := appattendance.NewDeleteCollectionUsecase(repo, clock)

	input := appattendance.DeleteCollectionInput{
		TenantID:     tenantID.String(),
		CollectionID: testCollection.CollectionID().String(),
	}

	result, err := usecase.Execute(context.Background(), input)

	if err != nil {
		t.Fatalf("Execute() should succeed, got error: %v", err)
	}

	if result.CollectionID != testCollection.CollectionID().String() {
		t.Errorf("CollectionID mismatch: got %v, want %v", result.CollectionID, testCollection.CollectionID().String())
	}

	if result.DeletedAt == nil {
		t.Error("DeletedAt should not be nil")
	}
}

func TestDeleteCollectionUsecase_Execute_NotFound(t *testing.T) {
	tenantID := common.NewTenantID()
	collectionID := common.NewCollectionID()

	repo := &MockAttendanceCollectionRepository{
		findByIDFunc: func(ctx context.Context, tid common.TenantID, cid common.CollectionID) (*attendance.AttendanceCollection, error) {
			return nil, common.NewNotFoundError("collection", cid.String())
		},
	}

	clock := &MockClock{nowFunc: func() time.Time { return time.Now() }}

	usecase := appattendance.NewDeleteCollectionUsecase(repo, clock)

	input := appattendance.DeleteCollectionInput{
		TenantID:     tenantID.String(),
		CollectionID: collectionID.String(),
	}

	_, err := usecase.Execute(context.Background(), input)

	if err == nil {
		t.Fatal("Execute() should fail when collection not found")
	}
}

func TestDeleteCollectionUsecase_Execute_AlreadyDeleted(t *testing.T) {
	tenantID := common.NewTenantID()
	testCollection := createTestCollection(t, tenantID)

	// Delete the collection first
	now := time.Now()
	_ = testCollection.Delete(now)

	repo := &MockAttendanceCollectionRepository{
		findByIDFunc: func(ctx context.Context, tid common.TenantID, cid common.CollectionID) (*attendance.AttendanceCollection, error) {
			return testCollection, nil
		},
	}

	clock := &MockClock{nowFunc: func() time.Time { return time.Now() }}

	usecase := appattendance.NewDeleteCollectionUsecase(repo, clock)

	input := appattendance.DeleteCollectionInput{
		TenantID:     tenantID.String(),
		CollectionID: testCollection.CollectionID().String(),
	}

	_, err := usecase.Execute(context.Background(), input)

	if err == nil {
		t.Fatal("Execute() should fail when collection is already deleted")
	}

	if !errors.Is(err, attendance.ErrAlreadyDeleted) {
		t.Errorf("Expected ErrAlreadyDeleted, got: %v", err)
	}
}

func TestDeleteCollectionUsecase_Execute_ErrorWhenSaveFails(t *testing.T) {
	tenantID := common.NewTenantID()
	testCollection := createTestCollection(t, tenantID)

	repo := &MockAttendanceCollectionRepository{
		findByIDFunc: func(ctx context.Context, tid common.TenantID, cid common.CollectionID) (*attendance.AttendanceCollection, error) {
			return testCollection, nil
		},
		saveFunc: func(ctx context.Context, c *attendance.AttendanceCollection) error {
			return errors.New("database error")
		},
	}

	clock := &MockClock{nowFunc: func() time.Time { return time.Now() }}

	usecase := appattendance.NewDeleteCollectionUsecase(repo, clock)

	input := appattendance.DeleteCollectionInput{
		TenantID:     tenantID.String(),
		CollectionID: testCollection.CollectionID().String(),
	}

	_, err := usecase.Execute(context.Background(), input)

	if err == nil {
		t.Fatal("Execute() should fail when save fails")
	}
}
