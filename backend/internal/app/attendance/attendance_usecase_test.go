package attendance_test

import (
	"context"
	"errors"
	"testing"
	"time"

	appattendance "github.com/erenoa/vrc-shift-scheduler/backend/internal/app/attendance"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/attendance"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/member"
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
	saveFunc                        func(ctx context.Context, c *attendance.AttendanceCollection) error
	findByIDFunc                    func(ctx context.Context, tenantID common.TenantID, collectionID common.CollectionID) (*attendance.AttendanceCollection, error)
	findByPublicTokenFunc           func(ctx context.Context, token common.PublicToken) (*attendance.AttendanceCollection, error)
	saveTargetDatesFunc             func(ctx context.Context, collectionID common.CollectionID, dates []*attendance.TargetDate) error
	saveGroupAssignmentsFunc        func(ctx context.Context, collectionID common.CollectionID, assignments []*attendance.CollectionGroupAssignment) error
	findResponsesByCollectionIDFunc func(ctx context.Context, collectionID common.CollectionID) ([]*attendance.AttendanceResponse, error)
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
	if m.findResponsesByCollectionIDFunc != nil {
		return m.findResponsesByCollectionIDFunc(ctx, collectionID)
	}
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

func (m *MockAttendanceCollectionRepository) FindResponsesByCollectionIDAndMemberID(ctx context.Context, tenantID common.TenantID, collectionID common.CollectionID, memberID common.MemberID) ([]*attendance.AttendanceResponse, error) {
	return nil, nil
}

// =====================================================
// Mock Role Repository
// =====================================================

type MockRoleRepository struct {
	findByIDFunc  func(ctx context.Context, tenantID common.TenantID, roleID common.RoleID) (*role.Role, error)
	findByIDsFunc func(ctx context.Context, tenantID common.TenantID, roleIDs []common.RoleID) ([]*role.Role, error)
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

func (m *MockRoleRepository) FindByIDs(ctx context.Context, tenantID common.TenantID, roleIDs []common.RoleID) ([]*role.Role, error) {
	if m.findByIDsFunc != nil {
		return m.findByIDsFunc(ctx, tenantID, roleIDs)
	}
	// Default: return mock roles for all requested IDs
	roles := make([]*role.Role, 0, len(roleIDs))
	for range roleIDs {
		roles = append(roles, &role.Role{})
	}
	return roles, nil
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
	txManager := &MockTxManager{}

	usecase := appattendance.NewCreateCollectionUsecase(repo, roleRepo, txManager, clock)

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
	txManager := &MockTxManager{}

	usecase := appattendance.NewCreateCollectionUsecase(repo, roleRepo, txManager, clock)

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
	txManager := &MockTxManager{}

	usecase := appattendance.NewCreateCollectionUsecase(repo, roleRepo, txManager, clock)

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
	txManager := &MockTxManager{}

	usecase := appattendance.NewCreateCollectionUsecase(repo, roleRepo, txManager, clock)

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
	txManager := &MockTxManager{}

	usecase := appattendance.NewCreateCollectionUsecase(repo, roleRepo, txManager, clock)

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
	txManager := &MockTxManager{}

	usecase := appattendance.NewCreateCollectionUsecase(repo, roleRepo, txManager, clock)

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

	targetDates := []appattendance.TargetDateInput{
		{TargetDate: now.Add(24 * time.Hour)},
		{TargetDate: now.Add(48 * time.Hour)},
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
	txManager := &MockTxManager{}

	usecase := appattendance.NewCreateCollectionUsecase(repo, roleRepo, txManager, clock)

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
	txManager := &MockTxManager{}

	usecase := appattendance.NewCreateCollectionUsecase(repo, roleRepo, txManager, clock)

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

// =====================================================
// Mock TxManager
// =====================================================

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
// Mock Member Repository
// =====================================================

type MockMemberRepository struct {
	findByIDFunc       func(ctx context.Context, tenantID common.TenantID, memberID common.MemberID) (*member.Member, error)
	findByTenantIDFunc func(ctx context.Context, tenantID common.TenantID) ([]*member.Member, error)
}

func (m *MockMemberRepository) Save(ctx context.Context, mem *member.Member) error {
	return nil
}

func (m *MockMemberRepository) FindByID(ctx context.Context, tenantID common.TenantID, memberID common.MemberID) (*member.Member, error) {
	if m.findByIDFunc != nil {
		return m.findByIDFunc(ctx, tenantID, memberID)
	}
	// Return a dummy member to indicate member exists
	return &member.Member{}, nil
}

func (m *MockMemberRepository) FindByTenantID(ctx context.Context, tenantID common.TenantID) ([]*member.Member, error) {
	if m.findByTenantIDFunc != nil {
		return m.findByTenantIDFunc(ctx, tenantID)
	}
	return nil, nil
}

func (m *MockMemberRepository) FindActiveByTenantID(ctx context.Context, tenantID common.TenantID) ([]*member.Member, error) {
	return nil, nil
}

func (m *MockMemberRepository) Delete(ctx context.Context, tenantID common.TenantID, memberID common.MemberID) error {
	return nil
}

func (m *MockMemberRepository) ExistsByDiscordUserID(ctx context.Context, tenantID common.TenantID, discordUserID string) (bool, error) {
	return false, nil
}

func (m *MockMemberRepository) FindByDiscordUserID(ctx context.Context, tenantID common.TenantID, discordUserID string) (*member.Member, error) {
	return nil, nil
}

func (m *MockMemberRepository) FindByEmail(ctx context.Context, tenantID common.TenantID, email string) (*member.Member, error) {
	return nil, nil
}

func (m *MockMemberRepository) ExistsByEmail(ctx context.Context, tenantID common.TenantID, email string) (bool, error) {
	return false, nil
}

// =====================================================
// AdminUpdateResponseUsecase Tests
// =====================================================

func TestAdminUpdateResponseUsecase_Execute_Success(t *testing.T) {
	tenantID := common.NewTenantID()
	memberID := common.NewMemberID()
	now := time.Now()
	testCollection := createTestCollection(t, tenantID)

	clock := &MockClock{
		nowFunc: func() time.Time { return now },
	}

	repo := &MockAttendanceCollectionRepository{
		findByIDFunc: func(ctx context.Context, tid common.TenantID, cid common.CollectionID) (*attendance.AttendanceCollection, error) {
			return testCollection, nil
		},
	}

	memberRepo := &MockMemberRepository{
		findByIDFunc: func(ctx context.Context, tid common.TenantID, mid common.MemberID) (*member.Member, error) {
			return &member.Member{}, nil // Member exists
		},
	}

	txManager := &MockTxManager{}

	usecase := appattendance.NewAdminUpdateResponseUsecase(repo, memberRepo, txManager, clock)

	// Create a target date ID for the test
	targetDateID := common.NewTargetDateID()

	input := appattendance.AdminUpdateResponseInput{
		TenantID:     tenantID.String(),
		CollectionID: testCollection.CollectionID().String(),
		MemberID:     memberID.String(),
		TargetDateID: targetDateID.String(),
		Response:     "attending",
		Note:         "テストノート",
	}

	result, err := usecase.Execute(context.Background(), input)

	if err != nil {
		t.Fatalf("Execute() should succeed, got error: %v", err)
	}

	if result == nil {
		t.Fatal("Result should not be nil")
	}

	if result.Response != "attending" {
		t.Errorf("Response mismatch: got %v, want 'attending'", result.Response)
	}

	if result.Note != "テストノート" {
		t.Errorf("Note mismatch: got %v, want 'テストノート'", result.Note)
	}
}

func TestAdminUpdateResponseUsecase_Execute_CollectionNotFound(t *testing.T) {
	tenantID := common.NewTenantID()
	memberID := common.NewMemberID()
	collectionID := common.NewCollectionID()
	targetDateID := common.NewTargetDateID()
	now := time.Now()

	clock := &MockClock{
		nowFunc: func() time.Time { return now },
	}

	repo := &MockAttendanceCollectionRepository{
		findByIDFunc: func(ctx context.Context, tid common.TenantID, cid common.CollectionID) (*attendance.AttendanceCollection, error) {
			return nil, common.NewNotFoundError("collection", cid.String())
		},
	}

	memberRepo := &MockMemberRepository{}
	txManager := &MockTxManager{}

	usecase := appattendance.NewAdminUpdateResponseUsecase(repo, memberRepo, txManager, clock)

	input := appattendance.AdminUpdateResponseInput{
		TenantID:     tenantID.String(),
		CollectionID: collectionID.String(),
		MemberID:     memberID.String(),
		TargetDateID: targetDateID.String(),
		Response:     "attending",
	}

	_, err := usecase.Execute(context.Background(), input)

	if err == nil {
		t.Fatal("Execute() should fail when collection not found")
	}

	if !errors.Is(err, appattendance.ErrCollectionNotFound) {
		t.Errorf("Expected ErrCollectionNotFound, got: %v", err)
	}
}

func TestAdminUpdateResponseUsecase_Execute_MemberNotFound(t *testing.T) {
	tenantID := common.NewTenantID()
	memberID := common.NewMemberID()
	targetDateID := common.NewTargetDateID()
	now := time.Now()
	testCollection := createTestCollection(t, tenantID)

	clock := &MockClock{
		nowFunc: func() time.Time { return now },
	}

	repo := &MockAttendanceCollectionRepository{
		findByIDFunc: func(ctx context.Context, tid common.TenantID, cid common.CollectionID) (*attendance.AttendanceCollection, error) {
			return testCollection, nil
		},
	}

	memberRepo := &MockMemberRepository{
		findByIDFunc: func(ctx context.Context, tid common.TenantID, mid common.MemberID) (*member.Member, error) {
			return nil, common.NewNotFoundError("member", mid.String())
		},
	}

	txManager := &MockTxManager{}

	usecase := appattendance.NewAdminUpdateResponseUsecase(repo, memberRepo, txManager, clock)

	input := appattendance.AdminUpdateResponseInput{
		TenantID:     tenantID.String(),
		CollectionID: testCollection.CollectionID().String(),
		MemberID:     memberID.String(),
		TargetDateID: targetDateID.String(),
		Response:     "attending",
	}

	_, err := usecase.Execute(context.Background(), input)

	if err == nil {
		t.Fatal("Execute() should fail when member not found")
	}

	if !errors.Is(err, appattendance.ErrMemberNotFound) {
		t.Errorf("Expected ErrMemberNotFound, got: %v", err)
	}
}

func TestAdminUpdateResponseUsecase_Execute_InvalidTenantID(t *testing.T) {
	now := time.Now()

	clock := &MockClock{
		nowFunc: func() time.Time { return now },
	}

	repo := &MockAttendanceCollectionRepository{}
	memberRepo := &MockMemberRepository{}
	txManager := &MockTxManager{}

	usecase := appattendance.NewAdminUpdateResponseUsecase(repo, memberRepo, txManager, clock)

	input := appattendance.AdminUpdateResponseInput{
		TenantID:     "invalid-tenant-id",
		CollectionID: common.NewCollectionID().String(),
		MemberID:     common.NewMemberID().String(),
		TargetDateID: common.NewTargetDateID().String(),
		Response:     "attending",
	}

	_, err := usecase.Execute(context.Background(), input)

	if err == nil {
		t.Fatal("Execute() should fail when tenant ID is invalid")
	}
}

func TestAdminUpdateResponseUsecase_Execute_InvalidCollectionID(t *testing.T) {
	tenantID := common.NewTenantID()
	now := time.Now()

	clock := &MockClock{
		nowFunc: func() time.Time { return now },
	}

	repo := &MockAttendanceCollectionRepository{}
	memberRepo := &MockMemberRepository{}
	txManager := &MockTxManager{}

	usecase := appattendance.NewAdminUpdateResponseUsecase(repo, memberRepo, txManager, clock)

	input := appattendance.AdminUpdateResponseInput{
		TenantID:     tenantID.String(),
		CollectionID: "invalid-collection-id",
		MemberID:     common.NewMemberID().String(),
		TargetDateID: common.NewTargetDateID().String(),
		Response:     "attending",
	}

	_, err := usecase.Execute(context.Background(), input)

	if err == nil {
		t.Fatal("Execute() should fail when collection ID is invalid")
	}
}

func TestAdminUpdateResponseUsecase_Execute_InvalidMemberID(t *testing.T) {
	tenantID := common.NewTenantID()
	now := time.Now()

	clock := &MockClock{
		nowFunc: func() time.Time { return now },
	}

	repo := &MockAttendanceCollectionRepository{}
	memberRepo := &MockMemberRepository{}
	txManager := &MockTxManager{}

	usecase := appattendance.NewAdminUpdateResponseUsecase(repo, memberRepo, txManager, clock)

	input := appattendance.AdminUpdateResponseInput{
		TenantID:     tenantID.String(),
		CollectionID: common.NewCollectionID().String(),
		MemberID:     "invalid-member-id",
		TargetDateID: common.NewTargetDateID().String(),
		Response:     "attending",
	}

	_, err := usecase.Execute(context.Background(), input)

	if err == nil {
		t.Fatal("Execute() should fail when member ID is invalid")
	}
}

func TestAdminUpdateResponseUsecase_Execute_InvalidTargetDateID(t *testing.T) {
	tenantID := common.NewTenantID()
	now := time.Now()

	clock := &MockClock{
		nowFunc: func() time.Time { return now },
	}

	repo := &MockAttendanceCollectionRepository{}
	memberRepo := &MockMemberRepository{}
	txManager := &MockTxManager{}

	usecase := appattendance.NewAdminUpdateResponseUsecase(repo, memberRepo, txManager, clock)

	input := appattendance.AdminUpdateResponseInput{
		TenantID:     tenantID.String(),
		CollectionID: common.NewCollectionID().String(),
		MemberID:     common.NewMemberID().String(),
		TargetDateID: "invalid-target-date-id",
		Response:     "attending",
	}

	_, err := usecase.Execute(context.Background(), input)

	if err == nil {
		t.Fatal("Execute() should fail when target date ID is invalid")
	}
}

func TestAdminUpdateResponseUsecase_Execute_InvalidResponseType(t *testing.T) {
	tenantID := common.NewTenantID()
	now := time.Now()

	clock := &MockClock{
		nowFunc: func() time.Time { return now },
	}

	repo := &MockAttendanceCollectionRepository{}
	memberRepo := &MockMemberRepository{}
	txManager := &MockTxManager{}

	usecase := appattendance.NewAdminUpdateResponseUsecase(repo, memberRepo, txManager, clock)

	input := appattendance.AdminUpdateResponseInput{
		TenantID:     tenantID.String(),
		CollectionID: common.NewCollectionID().String(),
		MemberID:     common.NewMemberID().String(),
		TargetDateID: common.NewTargetDateID().String(),
		Response:     "invalid_response_type",
	}

	_, err := usecase.Execute(context.Background(), input)

	if err == nil {
		t.Fatal("Execute() should fail when response type is invalid")
	}
}

func TestAdminUpdateResponseUsecase_Execute_WithOptionalTimes(t *testing.T) {
	tenantID := common.NewTenantID()
	memberID := common.NewMemberID()
	targetDateID := common.NewTargetDateID()
	now := time.Now()
	testCollection := createTestCollection(t, tenantID)

	clock := &MockClock{
		nowFunc: func() time.Time { return now },
	}

	repo := &MockAttendanceCollectionRepository{
		findByIDFunc: func(ctx context.Context, tid common.TenantID, cid common.CollectionID) (*attendance.AttendanceCollection, error) {
			return testCollection, nil
		},
	}

	memberRepo := &MockMemberRepository{
		findByIDFunc: func(ctx context.Context, tid common.TenantID, mid common.MemberID) (*member.Member, error) {
			return &member.Member{}, nil // Member exists
		},
	}

	txManager := &MockTxManager{}

	usecase := appattendance.NewAdminUpdateResponseUsecase(repo, memberRepo, txManager, clock)

	availableFrom := "09:00"
	availableTo := "18:00"

	input := appattendance.AdminUpdateResponseInput{
		TenantID:      tenantID.String(),
		CollectionID:  testCollection.CollectionID().String(),
		MemberID:      memberID.String(),
		TargetDateID:  targetDateID.String(),
		Response:      "attending",
		Note:          "時間指定あり",
		AvailableFrom: &availableFrom,
		AvailableTo:   &availableTo,
	}

	result, err := usecase.Execute(context.Background(), input)

	if err != nil {
		t.Fatalf("Execute() should succeed, got error: %v", err)
	}

	if result == nil {
		t.Fatal("Result should not be nil")
	}

	if result.AvailableFrom == nil || *result.AvailableFrom != "09:00" {
		t.Errorf("AvailableFrom mismatch: got %v, want '09:00'", result.AvailableFrom)
	}

	if result.AvailableTo == nil || *result.AvailableTo != "18:00" {
		t.Errorf("AvailableTo mismatch: got %v, want '18:00'", result.AvailableTo)
	}
}

// =====================================================
// CreateCollectionUsecase Role Tests (#113)
// =====================================================

func TestCreateCollectionUsecase_Execute_WithRoleAssignment(t *testing.T) {
	tenantID := common.NewTenantID()
	roleID := common.NewRoleID()
	now := time.Now()

	clock := &MockClock{
		nowFunc: func() time.Time { return now },
	}

	repo := &MockAttendanceCollectionRepository{
		saveFunc: func(ctx context.Context, c *attendance.AttendanceCollection) error {
			return nil
		},
	}

	roleRepo := &MockRoleRepository{
		findByIDsFunc: func(ctx context.Context, tid common.TenantID, roleIDs []common.RoleID) ([]*role.Role, error) {
			roles := make([]*role.Role, 0, len(roleIDs))
			for _, rid := range roleIDs {
				r, _ := role.ReconstructRole(rid, tid, "Test Role", "", "", 0, now, now, nil)
				roles = append(roles, r)
			}
			return roles, nil
		},
	}
	txManager := &MockTxManager{}

	usecase := appattendance.NewCreateCollectionUsecase(repo, roleRepo, txManager, clock)

	input := appattendance.CreateCollectionInput{
		TenantID:    tenantID.String(),
		Title:       "ロール割り当てテスト",
		Description: "",
		TargetType:  "event",
		TargetID:    "",
		RoleIDs:     []string{roleID.String()},
	}

	result, err := usecase.Execute(context.Background(), input)

	if err != nil {
		t.Fatalf("Execute() should succeed with role assignment, got error: %v", err)
	}

	if result == nil {
		t.Fatal("Result should not be nil")
	}

	if result.Title != "ロール割り当てテスト" {
		t.Errorf("Title mismatch: got %v", result.Title)
	}
}

func TestCreateCollectionUsecase_Execute_WithMultipleRoles(t *testing.T) {
	tenantID := common.NewTenantID()
	roleID1 := common.NewRoleID()
	roleID2 := common.NewRoleID()
	roleID3 := common.NewRoleID()
	now := time.Now()

	clock := &MockClock{
		nowFunc: func() time.Time { return now },
	}

	repo := &MockAttendanceCollectionRepository{
		saveFunc: func(ctx context.Context, c *attendance.AttendanceCollection) error {
			return nil
		},
	}

	roleRepo := &MockRoleRepository{
		findByIDsFunc: func(ctx context.Context, tid common.TenantID, roleIDs []common.RoleID) ([]*role.Role, error) {
			roles := make([]*role.Role, 0, len(roleIDs))
			for _, rid := range roleIDs {
				r, _ := role.ReconstructRole(rid, tid, "Test Role", "", "", 0, now, now, nil)
				roles = append(roles, r)
			}
			return roles, nil
		},
	}
	txManager := &MockTxManager{}

	usecase := appattendance.NewCreateCollectionUsecase(repo, roleRepo, txManager, clock)

	input := appattendance.CreateCollectionInput{
		TenantID:    tenantID.String(),
		Title:       "複数ロール割り当てテスト",
		Description: "",
		TargetType:  "event",
		TargetID:    "",
		RoleIDs:     []string{roleID1.String(), roleID2.String(), roleID3.String()},
	}

	result, err := usecase.Execute(context.Background(), input)

	if err != nil {
		t.Fatalf("Execute() should succeed with multiple role assignments, got error: %v", err)
	}

	if result == nil {
		t.Fatal("Result should not be nil")
	}
}

func TestCreateCollectionUsecase_Execute_WithInvalidRoleID(t *testing.T) {
	tenantID := common.NewTenantID()
	nonExistentRoleID := common.NewRoleID()
	now := time.Now()

	clock := &MockClock{
		nowFunc: func() time.Time { return now },
	}

	repo := &MockAttendanceCollectionRepository{
		saveFunc: func(ctx context.Context, c *attendance.AttendanceCollection) error {
			return nil
		},
	}

	roleRepo := &MockRoleRepository{
		findByIDsFunc: func(ctx context.Context, tid common.TenantID, roleIDs []common.RoleID) ([]*role.Role, error) {
			return []*role.Role{}, nil
		},
	}
	txManager := &MockTxManager{}

	usecase := appattendance.NewCreateCollectionUsecase(repo, roleRepo, txManager, clock)

	input := appattendance.CreateCollectionInput{
		TenantID:    tenantID.String(),
		Title:       "無効なロールIDテスト",
		Description: "",
		TargetType:  "event",
		TargetID:    "",
		RoleIDs:     []string{nonExistentRoleID.String()},
	}

	_, err := usecase.Execute(context.Background(), input)

	if err == nil {
		t.Fatal("Execute() should fail when role ID is invalid/not found")
	}
}

func TestCreateCollectionUsecase_Execute_WithBothGroupAndRole(t *testing.T) {
	tenantID := common.NewTenantID()
	roleID := common.NewRoleID()
	groupID := common.NewMemberGroupID()
	now := time.Now()

	clock := &MockClock{
		nowFunc: func() time.Time { return now },
	}

	repo := &MockAttendanceCollectionRepository{
		saveFunc: func(ctx context.Context, c *attendance.AttendanceCollection) error {
			return nil
		},
	}

	roleRepo := &MockRoleRepository{
		findByIDsFunc: func(ctx context.Context, tid common.TenantID, roleIDs []common.RoleID) ([]*role.Role, error) {
			roles := make([]*role.Role, 0, len(roleIDs))
			for _, rid := range roleIDs {
				r, _ := role.ReconstructRole(rid, tid, "Test Role", "", "", 0, now, now, nil)
				roles = append(roles, r)
			}
			return roles, nil
		},
	}
	txManager := &MockTxManager{}

	usecase := appattendance.NewCreateCollectionUsecase(repo, roleRepo, txManager, clock)

	input := appattendance.CreateCollectionInput{
		TenantID:    tenantID.String(),
		Title:       "グループとロール両方指定テスト",
		Description: "",
		TargetType:  "event",
		TargetID:    "",
		GroupIDs:    []string{groupID.String()},
		RoleIDs:     []string{roleID.String()},
	}

	result, err := usecase.Execute(context.Background(), input)

	if err != nil {
		t.Fatalf("Execute() should succeed with both group and role, got error: %v", err)
	}

	if result == nil {
		t.Fatal("Result should not be nil")
	}
}

func TestCreateCollectionUsecase_Execute_WithoutRoleAssignment(t *testing.T) {
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
	txManager := &MockTxManager{}

	usecase := appattendance.NewCreateCollectionUsecase(repo, roleRepo, txManager, clock)

	input := appattendance.CreateCollectionInput{
		TenantID:    tenantID.String(),
		Title:       "ロールなし（後方互換性）",
		Description: "",
		TargetType:  "event",
		TargetID:    "",
	}

	result, err := usecase.Execute(context.Background(), input)

	if err != nil {
		t.Fatalf("Execute() should succeed without role assignment (backward compatibility), got error: %v", err)
	}

	if result == nil {
		t.Fatal("Result should not be nil")
	}
}

// =====================================================
// GetAllPublicResponsesUsecase Tests (#133)
// =====================================================

func TestGetAllPublicResponsesUsecase_Execute_Success(t *testing.T) {
	tenantID := common.NewTenantID()
	memberID := common.NewMemberID()
	now := time.Now()
	testCollection := createTestCollection(t, tenantID)
	targetDateID := common.NewTargetDateID()

	testResponse, err := attendance.NewAttendanceResponse(
		now,
		testCollection.CollectionID(),
		tenantID,
		memberID,
		targetDateID,
		attendance.ResponseTypeAttending,
		"テストノート",
		nil,
		nil,
	)
	if err != nil {
		t.Fatalf("Failed to create test response: %v", err)
	}

	repo := &MockAttendanceCollectionRepository{
		findByPublicTokenFunc: func(ctx context.Context, token common.PublicToken) (*attendance.AttendanceCollection, error) {
			return testCollection, nil
		},
		findResponsesByCollectionIDFunc: func(ctx context.Context, collectionID common.CollectionID) ([]*attendance.AttendanceResponse, error) {
			return []*attendance.AttendanceResponse{testResponse}, nil
		},
	}

	memberRepo := &MockMemberRepository{
		findByTenantIDFunc: func(ctx context.Context, tid common.TenantID) ([]*member.Member, error) {
			m, _ := member.ReconstructMember(memberID, tid, "テストメンバー", "", "", true, now, now, nil)
			return []*member.Member{m}, nil
		},
	}

	usecase := appattendance.NewGetAllPublicResponsesUsecase(repo, memberRepo)

	input := appattendance.GetAllPublicResponsesInput{
		PublicToken: testCollection.PublicToken().String(),
	}

	result, err := usecase.Execute(context.Background(), input)

	if err != nil {
		t.Fatalf("Execute() should succeed, got error: %v", err)
	}

	if result == nil {
		t.Fatal("Result should not be nil")
	}

	if len(result.Responses) != 1 {
		t.Errorf("Expected 1 response, got %d", len(result.Responses))
	}

	if result.Responses[0].MemberName != "テストメンバー" {
		t.Errorf("MemberName mismatch: got %v, want 'テストメンバー'", result.Responses[0].MemberName)
	}
}

func TestGetAllPublicResponsesUsecase_Execute_InvalidToken(t *testing.T) {
	repo := &MockAttendanceCollectionRepository{}
	memberRepo := &MockMemberRepository{}

	usecase := appattendance.NewGetAllPublicResponsesUsecase(repo, memberRepo)

	input := appattendance.GetAllPublicResponsesInput{
		PublicToken: "invalid-token-format",
	}

	_, err := usecase.Execute(context.Background(), input)

	if err == nil {
		t.Fatal("Execute() should fail with invalid token format")
	}
}

func TestGetAllPublicResponsesUsecase_Execute_NotFoundToken(t *testing.T) {
	repo := &MockAttendanceCollectionRepository{
		findByPublicTokenFunc: func(ctx context.Context, token common.PublicToken) (*attendance.AttendanceCollection, error) {
			return nil, common.NewNotFoundError("AttendanceCollection", token.String())
		},
	}
	memberRepo := &MockMemberRepository{}

	usecase := appattendance.NewGetAllPublicResponsesUsecase(repo, memberRepo)

	validToken := common.NewPublicToken()
	input := appattendance.GetAllPublicResponsesInput{
		PublicToken: validToken.String(),
	}

	_, err := usecase.Execute(context.Background(), input)

	if err == nil {
		t.Fatal("Execute() should fail when token is not found")
	}
}

func TestGetAllPublicResponsesUsecase_Execute_EmptyResponses(t *testing.T) {
	tenantID := common.NewTenantID()
	testCollection := createTestCollection(t, tenantID)

	repo := &MockAttendanceCollectionRepository{
		findByPublicTokenFunc: func(ctx context.Context, token common.PublicToken) (*attendance.AttendanceCollection, error) {
			return testCollection, nil
		},
		findResponsesByCollectionIDFunc: func(ctx context.Context, collectionID common.CollectionID) ([]*attendance.AttendanceResponse, error) {
			return []*attendance.AttendanceResponse{}, nil
		},
	}

	memberRepo := &MockMemberRepository{
		findByTenantIDFunc: func(ctx context.Context, tid common.TenantID) ([]*member.Member, error) {
			return []*member.Member{}, nil
		},
	}

	usecase := appattendance.NewGetAllPublicResponsesUsecase(repo, memberRepo)

	input := appattendance.GetAllPublicResponsesInput{
		PublicToken: testCollection.PublicToken().String(),
	}

	result, err := usecase.Execute(context.Background(), input)

	if err != nil {
		t.Fatalf("Execute() should succeed even with empty responses, got error: %v", err)
	}

	if result == nil {
		t.Fatal("Result should not be nil")
	}

	if len(result.Responses) != 0 {
		t.Errorf("Expected 0 responses, got %d", len(result.Responses))
	}
}

func TestGetAllPublicResponsesUsecase_Execute_MemberNotFound(t *testing.T) {
	tenantID := common.NewTenantID()
	memberID := common.NewMemberID()
	now := time.Now()
	testCollection := createTestCollection(t, tenantID)
	targetDateID := common.NewTargetDateID()

	testResponse, err := attendance.NewAttendanceResponse(
		now,
		testCollection.CollectionID(),
		tenantID,
		memberID,
		targetDateID,
		attendance.ResponseTypeAttending,
		"",
		nil,
		nil,
	)
	if err != nil {
		t.Fatalf("Failed to create test response: %v", err)
	}

	repo := &MockAttendanceCollectionRepository{
		findByPublicTokenFunc: func(ctx context.Context, token common.PublicToken) (*attendance.AttendanceCollection, error) {
			return testCollection, nil
		},
		findResponsesByCollectionIDFunc: func(ctx context.Context, collectionID common.CollectionID) ([]*attendance.AttendanceResponse, error) {
			return []*attendance.AttendanceResponse{testResponse}, nil
		},
	}

	memberRepo := &MockMemberRepository{
		findByTenantIDFunc: func(ctx context.Context, tid common.TenantID) ([]*member.Member, error) {
			return []*member.Member{}, nil
		},
	}

	usecase := appattendance.NewGetAllPublicResponsesUsecase(repo, memberRepo)

	input := appattendance.GetAllPublicResponsesInput{
		PublicToken: testCollection.PublicToken().String(),
	}

	result, err := usecase.Execute(context.Background(), input)

	if err != nil {
		t.Fatalf("Execute() should succeed even when member not found (fallback to ID), got error: %v", err)
	}

	if result == nil {
		t.Fatal("Result should not be nil")
	}

	if result.Responses[0].MemberName != memberID.String() {
		t.Errorf("MemberName should fallback to MemberID when member not found: got %v, want %v", result.Responses[0].MemberName, memberID.String())
	}
}
