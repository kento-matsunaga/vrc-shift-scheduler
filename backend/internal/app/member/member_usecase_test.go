package member_test

import (
	"context"
	"errors"
	"testing"
	"time"

	appmember "github.com/erenoa/vrc-shift-scheduler/backend/internal/app/member"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/member"
)

// =====================================================
// Mock Repositories
// =====================================================

type MockMemberRepository struct {
	saveFunc                  func(ctx context.Context, m *member.Member) error
	findByIDFunc              func(ctx context.Context, tenantID common.TenantID, memberID common.MemberID) (*member.Member, error)
	findByTenantIDFunc        func(ctx context.Context, tenantID common.TenantID) ([]*member.Member, error)
	findActiveByTenantIDFunc  func(ctx context.Context, tenantID common.TenantID) ([]*member.Member, error)
	existsByDiscordUserIDFunc func(ctx context.Context, tenantID common.TenantID, discordUserID string) (bool, error)
	existsByEmailFunc         func(ctx context.Context, tenantID common.TenantID, email string) (bool, error)
}

func (m *MockMemberRepository) Save(ctx context.Context, mem *member.Member) error {
	if m.saveFunc != nil {
		return m.saveFunc(ctx, mem)
	}
	return nil
}

func (m *MockMemberRepository) FindByID(ctx context.Context, tenantID common.TenantID, memberID common.MemberID) (*member.Member, error) {
	if m.findByIDFunc != nil {
		return m.findByIDFunc(ctx, tenantID, memberID)
	}
	return nil, errors.New("not implemented")
}

func (m *MockMemberRepository) FindByTenantID(ctx context.Context, tenantID common.TenantID) ([]*member.Member, error) {
	if m.findByTenantIDFunc != nil {
		return m.findByTenantIDFunc(ctx, tenantID)
	}
	return nil, nil
}

func (m *MockMemberRepository) FindActiveByTenantID(ctx context.Context, tenantID common.TenantID) ([]*member.Member, error) {
	if m.findActiveByTenantIDFunc != nil {
		return m.findActiveByTenantIDFunc(ctx, tenantID)
	}
	return nil, nil
}

func (m *MockMemberRepository) FindByDiscordUserID(ctx context.Context, tenantID common.TenantID, discordUserID string) (*member.Member, error) {
	return nil, nil
}

func (m *MockMemberRepository) FindByEmail(ctx context.Context, tenantID common.TenantID, email string) (*member.Member, error) {
	return nil, nil
}

func (m *MockMemberRepository) ExistsByDiscordUserID(ctx context.Context, tenantID common.TenantID, discordUserID string) (bool, error) {
	if m.existsByDiscordUserIDFunc != nil {
		return m.existsByDiscordUserIDFunc(ctx, tenantID, discordUserID)
	}
	return false, nil
}

func (m *MockMemberRepository) ExistsByEmail(ctx context.Context, tenantID common.TenantID, email string) (bool, error) {
	if m.existsByEmailFunc != nil {
		return m.existsByEmailFunc(ctx, tenantID, email)
	}
	return false, nil
}

type MockMemberRoleRepository struct {
	findRolesByMemberIDFunc func(ctx context.Context, memberID common.MemberID) ([]common.RoleID, error)
	setMemberRolesFunc      func(ctx context.Context, memberID common.MemberID, roleIDs []common.RoleID) error
}

func (m *MockMemberRoleRepository) FindRolesByMemberID(ctx context.Context, memberID common.MemberID) ([]common.RoleID, error) {
	if m.findRolesByMemberIDFunc != nil {
		return m.findRolesByMemberIDFunc(ctx, memberID)
	}
	return []common.RoleID{}, nil
}

func (m *MockMemberRoleRepository) SetMemberRoles(ctx context.Context, memberID common.MemberID, roleIDs []common.RoleID) error {
	if m.setMemberRolesFunc != nil {
		return m.setMemberRolesFunc(ctx, memberID, roleIDs)
	}
	return nil
}

// =====================================================
// Helper functions
// =====================================================

func createTestMember(t *testing.T, tenantID common.TenantID, displayName string) *member.Member {
	t.Helper()
	now := time.Now()
	mem, err := member.NewMember(
		now,
		tenantID,
		displayName,
		"discord_user_123",
		"test@example.com",
	)
	if err != nil {
		t.Fatalf("Failed to create test member: %v", err)
	}
	return mem
}

// =====================================================
// CreateMemberUsecase Tests
// =====================================================

func TestCreateMemberUsecase_Execute_Success(t *testing.T) {
	tenantID := common.NewTenantID()

	memberRepo := &MockMemberRepository{
		existsByDiscordUserIDFunc: func(ctx context.Context, tid common.TenantID, discordUserID string) (bool, error) {
			return false, nil
		},
		existsByEmailFunc: func(ctx context.Context, tid common.TenantID, email string) (bool, error) {
			return false, nil
		},
		saveFunc: func(ctx context.Context, m *member.Member) error {
			return nil
		},
	}

	usecase := appmember.NewCreateMemberUsecase(memberRepo)

	input := appmember.CreateMemberInput{
		TenantID:      tenantID,
		DisplayName:   "テストメンバー",
		DiscordUserID: "discord_user_123",
		Email:         "test@example.com",
	}

	result, err := usecase.Execute(context.Background(), input)

	if err != nil {
		t.Fatalf("Execute() should succeed, got error: %v", err)
	}

	if result == nil {
		t.Fatal("Result should not be nil")
	}

	if result.DisplayName() != "テストメンバー" {
		t.Errorf("DisplayName mismatch: got %v, want 'テストメンバー'", result.DisplayName())
	}
}

func TestCreateMemberUsecase_Execute_ErrorWhenDiscordUserIDExists(t *testing.T) {
	tenantID := common.NewTenantID()

	memberRepo := &MockMemberRepository{
		existsByDiscordUserIDFunc: func(ctx context.Context, tid common.TenantID, discordUserID string) (bool, error) {
			return true, nil // Already exists
		},
	}

	usecase := appmember.NewCreateMemberUsecase(memberRepo)

	input := appmember.CreateMemberInput{
		TenantID:      tenantID,
		DisplayName:   "テストメンバー",
		DiscordUserID: "discord_user_123",
		Email:         "",
	}

	_, err := usecase.Execute(context.Background(), input)

	if err == nil {
		t.Fatal("Execute() should fail when Discord user ID already exists")
	}
}

func TestCreateMemberUsecase_Execute_ErrorWhenEmailExists(t *testing.T) {
	tenantID := common.NewTenantID()

	memberRepo := &MockMemberRepository{
		existsByDiscordUserIDFunc: func(ctx context.Context, tid common.TenantID, discordUserID string) (bool, error) {
			return false, nil
		},
		existsByEmailFunc: func(ctx context.Context, tid common.TenantID, email string) (bool, error) {
			return true, nil // Already exists
		},
	}

	usecase := appmember.NewCreateMemberUsecase(memberRepo)

	input := appmember.CreateMemberInput{
		TenantID:      tenantID,
		DisplayName:   "テストメンバー",
		DiscordUserID: "",
		Email:         "test@example.com",
	}

	_, err := usecase.Execute(context.Background(), input)

	if err == nil {
		t.Fatal("Execute() should fail when email already exists")
	}
}

func TestCreateMemberUsecase_Execute_ErrorWhenSaveFails(t *testing.T) {
	tenantID := common.NewTenantID()

	memberRepo := &MockMemberRepository{
		existsByDiscordUserIDFunc: func(ctx context.Context, tid common.TenantID, discordUserID string) (bool, error) {
			return false, nil
		},
		existsByEmailFunc: func(ctx context.Context, tid common.TenantID, email string) (bool, error) {
			return false, nil
		},
		saveFunc: func(ctx context.Context, m *member.Member) error {
			return errors.New("database error")
		},
	}

	usecase := appmember.NewCreateMemberUsecase(memberRepo)

	input := appmember.CreateMemberInput{
		TenantID:      tenantID,
		DisplayName:   "テストメンバー",
		DiscordUserID: "",
		Email:         "",
	}

	_, err := usecase.Execute(context.Background(), input)

	if err == nil {
		t.Fatal("Execute() should fail when save fails")
	}
}

// =====================================================
// ListMembersUsecase Tests
// =====================================================

func TestListMembersUsecase_Execute_Success(t *testing.T) {
	tenantID := common.NewTenantID()
	testMembers := []*member.Member{
		createTestMember(t, tenantID, "メンバー1"),
		createTestMember(t, tenantID, "メンバー2"),
		createTestMember(t, tenantID, "メンバー3"),
	}

	memberRepo := &MockMemberRepository{
		findByTenantIDFunc: func(ctx context.Context, tid common.TenantID) ([]*member.Member, error) {
			return testMembers, nil
		},
	}

	memberRoleRepo := &MockMemberRoleRepository{
		findRolesByMemberIDFunc: func(ctx context.Context, memberID common.MemberID) ([]common.RoleID, error) {
			return []common.RoleID{common.NewRoleID()}, nil
		},
	}

	usecase := appmember.NewListMembersUsecase(memberRepo, memberRoleRepo)

	input := appmember.ListMembersInput{
		TenantID: tenantID,
	}

	result, err := usecase.Execute(context.Background(), input)

	if err != nil {
		t.Fatalf("Execute() should succeed, got error: %v", err)
	}

	if len(result) != 3 {
		t.Errorf("Expected 3 members, got %d", len(result))
	}

	// Check role aggregation
	for _, mr := range result {
		if len(mr.RoleIDs) != 1 {
			t.Errorf("Expected 1 role per member, got %d", len(mr.RoleIDs))
		}
	}
}

func TestListMembersUsecase_Execute_WithIsActiveFilter(t *testing.T) {
	tenantID := common.NewTenantID()
	activeMember := createTestMember(t, tenantID, "アクティブメンバー")
	inactiveMember := createTestMember(t, tenantID, "非アクティブメンバー")
	inactiveMember.Deactivate()

	testMembers := []*member.Member{activeMember, inactiveMember}

	memberRepo := &MockMemberRepository{
		findByTenantIDFunc: func(ctx context.Context, tid common.TenantID) ([]*member.Member, error) {
			return testMembers, nil
		},
	}

	memberRoleRepo := &MockMemberRoleRepository{}

	usecase := appmember.NewListMembersUsecase(memberRepo, memberRoleRepo)

	// Filter active only
	isActive := true
	input := appmember.ListMembersInput{
		TenantID: tenantID,
		IsActive: &isActive,
	}

	result, err := usecase.Execute(context.Background(), input)

	if err != nil {
		t.Fatalf("Execute() should succeed, got error: %v", err)
	}

	if len(result) != 1 {
		t.Errorf("Expected 1 active member, got %d", len(result))
	}

	if result[0].Member.DisplayName() != "アクティブメンバー" {
		t.Errorf("Expected active member, got %v", result[0].Member.DisplayName())
	}
}

func TestListMembersUsecase_Execute_EmptyList(t *testing.T) {
	tenantID := common.NewTenantID()

	memberRepo := &MockMemberRepository{
		findByTenantIDFunc: func(ctx context.Context, tid common.TenantID) ([]*member.Member, error) {
			return []*member.Member{}, nil
		},
	}

	memberRoleRepo := &MockMemberRoleRepository{}

	usecase := appmember.NewListMembersUsecase(memberRepo, memberRoleRepo)

	input := appmember.ListMembersInput{
		TenantID: tenantID,
	}

	result, err := usecase.Execute(context.Background(), input)

	if err != nil {
		t.Fatalf("Execute() should succeed, got error: %v", err)
	}

	if len(result) != 0 {
		t.Errorf("Expected 0 members, got %d", len(result))
	}
}

// =====================================================
// GetMemberUsecase Tests
// =====================================================

func TestGetMemberUsecase_Execute_Success(t *testing.T) {
	tenantID := common.NewTenantID()
	testMember := createTestMember(t, tenantID, "テストメンバー")
	roleID := common.NewRoleID()

	memberRepo := &MockMemberRepository{
		findByIDFunc: func(ctx context.Context, tid common.TenantID, memID common.MemberID) (*member.Member, error) {
			return testMember, nil
		},
	}

	memberRoleRepo := &MockMemberRoleRepository{
		findRolesByMemberIDFunc: func(ctx context.Context, memberID common.MemberID) ([]common.RoleID, error) {
			return []common.RoleID{roleID}, nil
		},
	}

	usecase := appmember.NewGetMemberUsecase(memberRepo, memberRoleRepo)

	input := appmember.GetMemberInput{
		TenantID: tenantID,
		MemberID: testMember.MemberID(),
	}

	result, err := usecase.Execute(context.Background(), input)

	if err != nil {
		t.Fatalf("Execute() should succeed, got error: %v", err)
	}

	if result.Member.MemberID() != testMember.MemberID() {
		t.Errorf("MemberID mismatch: got %v, want %v", result.Member.MemberID(), testMember.MemberID())
	}

	if len(result.RoleIDs) != 1 {
		t.Errorf("Expected 1 role, got %d", len(result.RoleIDs))
	}
}

func TestGetMemberUsecase_Execute_NotFound(t *testing.T) {
	tenantID := common.NewTenantID()
	memberID := common.NewMemberID()

	memberRepo := &MockMemberRepository{
		findByIDFunc: func(ctx context.Context, tid common.TenantID, memID common.MemberID) (*member.Member, error) {
			return nil, common.NewNotFoundError("member", memID.String())
		},
	}

	memberRoleRepo := &MockMemberRoleRepository{}

	usecase := appmember.NewGetMemberUsecase(memberRepo, memberRoleRepo)

	input := appmember.GetMemberInput{
		TenantID: tenantID,
		MemberID: memberID,
	}

	_, err := usecase.Execute(context.Background(), input)

	if err == nil {
		t.Fatal("Execute() should fail when member not found")
	}
}

// =====================================================
// DeleteMemberUsecase Tests
// =====================================================

func TestDeleteMemberUsecase_Execute_Success(t *testing.T) {
	tenantID := common.NewTenantID()
	testMember := createTestMember(t, tenantID, "テストメンバー")

	var savedMember *member.Member

	memberRepo := &MockMemberRepository{
		findByIDFunc: func(ctx context.Context, tid common.TenantID, memID common.MemberID) (*member.Member, error) {
			return testMember, nil
		},
		saveFunc: func(ctx context.Context, m *member.Member) error {
			savedMember = m
			return nil
		},
	}

	usecase := appmember.NewDeleteMemberUsecase(memberRepo)

	input := appmember.DeleteMemberInput{
		TenantID: tenantID,
		MemberID: testMember.MemberID(),
	}

	err := usecase.Execute(context.Background(), input)

	if err != nil {
		t.Fatalf("Execute() should succeed, got error: %v", err)
	}

	if savedMember == nil {
		t.Fatal("Member should be saved")
	}

	if !savedMember.IsDeleted() {
		t.Error("Member should be deleted")
	}
}

func TestDeleteMemberUsecase_Execute_ErrorWhenNotFound(t *testing.T) {
	tenantID := common.NewTenantID()
	memberID := common.NewMemberID()

	memberRepo := &MockMemberRepository{
		findByIDFunc: func(ctx context.Context, tid common.TenantID, memID common.MemberID) (*member.Member, error) {
			return nil, common.NewNotFoundError("member", memID.String())
		},
	}

	usecase := appmember.NewDeleteMemberUsecase(memberRepo)

	input := appmember.DeleteMemberInput{
		TenantID: tenantID,
		MemberID: memberID,
	}

	err := usecase.Execute(context.Background(), input)

	if err == nil {
		t.Fatal("Execute() should fail when member not found")
	}
}

// =====================================================
// BulkImportMembersUsecase Tests
// =====================================================

func TestBulkImportMembersUsecase_Execute_Success(t *testing.T) {
	tenantID := common.NewTenantID()

	memberRepo := &MockMemberRepository{
		saveFunc: func(ctx context.Context, m *member.Member) error {
			return nil
		},
	}

	memberRoleRepo := &MockMemberRoleRepository{
		setMemberRolesFunc: func(ctx context.Context, memberID common.MemberID, roleIDs []common.RoleID) error {
			return nil
		},
	}

	usecase := appmember.NewBulkImportMembersUsecase(memberRepo, memberRoleRepo)

	input := appmember.BulkImportMembersInput{
		TenantID: tenantID,
		Members: []appmember.BulkImportMemberInput{
			{DisplayName: "メンバー1"},
			{DisplayName: "メンバー2"},
			{DisplayName: "メンバー3"},
		},
	}

	result, err := usecase.Execute(context.Background(), input)

	if err != nil {
		t.Fatalf("Execute() should succeed, got error: %v", err)
	}

	if result.TotalCount != 3 {
		t.Errorf("TotalCount should be 3, got %d", result.TotalCount)
	}

	if result.SuccessCount != 3 {
		t.Errorf("SuccessCount should be 3, got %d", result.SuccessCount)
	}

	if result.FailedCount != 0 {
		t.Errorf("FailedCount should be 0, got %d", result.FailedCount)
	}
}

func TestBulkImportMembersUsecase_Execute_PartialFailure(t *testing.T) {
	tenantID := common.NewTenantID()

	memberRepo := &MockMemberRepository{
		saveFunc: func(ctx context.Context, m *member.Member) error {
			return nil
		},
	}

	memberRoleRepo := &MockMemberRoleRepository{}

	usecase := appmember.NewBulkImportMembersUsecase(memberRepo, memberRoleRepo)

	input := appmember.BulkImportMembersInput{
		TenantID: tenantID,
		Members: []appmember.BulkImportMemberInput{
			{DisplayName: "メンバー1"},
			{DisplayName: ""},       // Empty name - should fail
			{DisplayName: "メンバー3"},
		},
	}

	result, err := usecase.Execute(context.Background(), input)

	if err != nil {
		t.Fatalf("Execute() should succeed even with partial failures, got error: %v", err)
	}

	if result.TotalCount != 3 {
		t.Errorf("TotalCount should be 3, got %d", result.TotalCount)
	}

	if result.SuccessCount != 2 {
		t.Errorf("SuccessCount should be 2, got %d", result.SuccessCount)
	}

	if result.FailedCount != 1 {
		t.Errorf("FailedCount should be 1, got %d", result.FailedCount)
	}
}

func TestBulkImportMembersUsecase_Execute_DisplayNameTooLong(t *testing.T) {
	tenantID := common.NewTenantID()

	memberRepo := &MockMemberRepository{
		saveFunc: func(ctx context.Context, m *member.Member) error {
			return nil
		},
	}

	memberRoleRepo := &MockMemberRoleRepository{}

	usecase := appmember.NewBulkImportMembersUsecase(memberRepo, memberRoleRepo)

	longName := make([]byte, 51)
	for i := range longName {
		longName[i] = 'a'
	}

	input := appmember.BulkImportMembersInput{
		TenantID: tenantID,
		Members: []appmember.BulkImportMemberInput{
			{DisplayName: string(longName)}, // Too long - should fail
		},
	}

	result, err := usecase.Execute(context.Background(), input)

	if err != nil {
		t.Fatalf("Execute() should succeed even with failures, got error: %v", err)
	}

	if result.FailedCount != 1 {
		t.Errorf("FailedCount should be 1, got %d", result.FailedCount)
	}

	if result.Results[0].Error != "display_name must be 50 characters or less" {
		t.Errorf("Error message mismatch: got %v", result.Results[0].Error)
	}
}

func TestBulkImportMembersUsecase_Execute_WithRoles(t *testing.T) {
	tenantID := common.NewTenantID()
	roleID := common.NewRoleID()

	var assignedRoleIDs []common.RoleID

	memberRepo := &MockMemberRepository{
		saveFunc: func(ctx context.Context, m *member.Member) error {
			return nil
		},
	}

	memberRoleRepo := &MockMemberRoleRepository{
		setMemberRolesFunc: func(ctx context.Context, memberID common.MemberID, roleIDs []common.RoleID) error {
			assignedRoleIDs = roleIDs
			return nil
		},
	}

	usecase := appmember.NewBulkImportMembersUsecase(memberRepo, memberRoleRepo)

	input := appmember.BulkImportMembersInput{
		TenantID: tenantID,
		Members: []appmember.BulkImportMemberInput{
			{
				DisplayName: "メンバー1",
				RoleIDs:     []string{roleID.String()},
			},
		},
	}

	result, err := usecase.Execute(context.Background(), input)

	if err != nil {
		t.Fatalf("Execute() should succeed, got error: %v", err)
	}

	if result.SuccessCount != 1 {
		t.Errorf("SuccessCount should be 1, got %d", result.SuccessCount)
	}

	if len(assignedRoleIDs) != 1 {
		t.Errorf("Should have assigned 1 role, got %d", len(assignedRoleIDs))
	}
}
