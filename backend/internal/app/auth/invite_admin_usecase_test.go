package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/auth"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/services"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/tenant"
)

// =====================================================
// Mock Implementations for InviteAdminUsecase
// =====================================================

// MockInvitationRepository is a mock implementation of auth.InvitationRepository
type MockInvitationRepository struct {
	saveFunc               func(ctx context.Context, invitation *auth.Invitation) error
	findByTokenFunc        func(ctx context.Context, token string) (*auth.Invitation, error)
	findByTenantIDFunc     func(ctx context.Context, tenantID common.TenantID) ([]*auth.Invitation, error)
	existsPendingByEmailFunc func(ctx context.Context, tenantID common.TenantID, email string) (bool, error)
	deleteFunc             func(ctx context.Context, invitationID auth.InvitationID) error
}

func (m *MockInvitationRepository) Save(ctx context.Context, invitation *auth.Invitation) error {
	if m.saveFunc != nil {
		return m.saveFunc(ctx, invitation)
	}
	return nil
}

func (m *MockInvitationRepository) FindByToken(ctx context.Context, token string) (*auth.Invitation, error) {
	if m.findByTokenFunc != nil {
		return m.findByTokenFunc(ctx, token)
	}
	return nil, errors.New("not implemented")
}

func (m *MockInvitationRepository) FindByTenantID(ctx context.Context, tenantID common.TenantID) ([]*auth.Invitation, error) {
	if m.findByTenantIDFunc != nil {
		return m.findByTenantIDFunc(ctx, tenantID)
	}
	return nil, errors.New("not implemented")
}

func (m *MockInvitationRepository) ExistsPendingByEmail(ctx context.Context, tenantID common.TenantID, email string) (bool, error) {
	if m.existsPendingByEmailFunc != nil {
		return m.existsPendingByEmailFunc(ctx, tenantID, email)
	}
	return false, nil
}

func (m *MockInvitationRepository) Delete(ctx context.Context, invitationID auth.InvitationID) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, invitationID)
	}
	return nil
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

// MockTenantRepository is a mock implementation of tenant.TenantRepository
type MockTenantRepository struct {
	findByIDFunc                      func(ctx context.Context, tenantID common.TenantID) (*tenant.Tenant, error)
	findByPendingStripeSessionIDFunc  func(ctx context.Context, sessionID string) (*tenant.Tenant, error)
	saveFunc                          func(ctx context.Context, t *tenant.Tenant) error
	listAllFunc                       func(ctx context.Context, status *tenant.TenantStatus, limit, offset int) ([]*tenant.Tenant, int, error)
}

func (m *MockTenantRepository) FindByID(ctx context.Context, tenantID common.TenantID) (*tenant.Tenant, error) {
	if m.findByIDFunc != nil {
		return m.findByIDFunc(ctx, tenantID)
	}
	return nil, errors.New("not implemented")
}

func (m *MockTenantRepository) FindByPendingStripeSessionID(ctx context.Context, sessionID string) (*tenant.Tenant, error) {
	if m.findByPendingStripeSessionIDFunc != nil {
		return m.findByPendingStripeSessionIDFunc(ctx, sessionID)
	}
	return nil, errors.New("not implemented")
}

func (m *MockTenantRepository) Save(ctx context.Context, t *tenant.Tenant) error {
	if m.saveFunc != nil {
		return m.saveFunc(ctx, t)
	}
	return nil
}

func (m *MockTenantRepository) ListAll(ctx context.Context, status *tenant.TenantStatus, limit, offset int) ([]*tenant.Tenant, int, error) {
	if m.listAllFunc != nil {
		return m.listAllFunc(ctx, status, limit, offset)
	}
	return nil, 0, errors.New("not implemented")
}

// MockEmailService is a mock implementation of services.EmailService
type MockEmailService struct {
	sendInvitationEmailFunc     func(ctx context.Context, input services.SendInvitationEmailInput) error
	sendPasswordResetEmailFunc  func(ctx context.Context, input services.SendPasswordResetEmailInput) error
}

func (m *MockEmailService) SendInvitationEmail(ctx context.Context, input services.SendInvitationEmailInput) error {
	if m.sendInvitationEmailFunc != nil {
		return m.sendInvitationEmailFunc(ctx, input)
	}
	return nil
}

func (m *MockEmailService) SendPasswordResetEmail(ctx context.Context, input services.SendPasswordResetEmailInput) error {
	if m.sendPasswordResetEmailFunc != nil {
		return m.sendPasswordResetEmailFunc(ctx, input)
	}
	return nil
}

// =====================================================
// Test Helper Functions
// =====================================================

func createTestInviterAdminWithTenantID(t *testing.T, tenantID common.TenantID) *auth.Admin {
	t.Helper()
	now := time.Now()

	admin, err := auth.NewAdmin(now, tenantID, "inviter@example.com", "$2a$10$hash", "Inviter Admin", auth.RoleOwner)
	if err != nil {
		t.Fatalf("Failed to create test inviter admin: %v", err)
	}
	return admin
}

func createTestInviterAdmin(t *testing.T) *auth.Admin {
	t.Helper()
	return createTestInviterAdminWithTenantID(t, common.NewTenantID())
}

func createTestTenant(t *testing.T, tenantID common.TenantID) *tenant.Tenant {
	t.Helper()
	now := time.Now()

	testTenant, err := tenant.ReconstructTenant(
		tenantID,
		"Test Tenant",
		"Asia/Tokyo",
		true,
		tenant.TenantStatusActive,
		nil,  // graceUntil
		nil,  // pendingExpiresAt
		nil,  // pendingStripeSessionID
		now,
		now,
		nil,
	)
	if err != nil {
		t.Fatalf("Failed to create test tenant: %v", err)
	}
	return testTenant
}

// =====================================================
// InviteAdminUsecase Tests
// =====================================================

func TestInviteAdminUsecase_Execute_Success(t *testing.T) {
	tenantID := common.NewTenantID()
	inviter := createTestInviterAdminWithTenantID(t, tenantID)
	testTenant := createTestTenant(t, tenantID)

	adminRepo := &MockAdminRepository{
		findByIDWithTenantFunc: func(ctx context.Context, tenantID common.TenantID, adminID common.AdminID) (*auth.Admin, error) {
			return nil, errors.New("not implemented")
		},
		findByEmailGlobalFunc: func(ctx context.Context, email string) (*auth.Admin, error) {
			return nil, common.NewNotFoundError("Admin", email) // Not found = good
		},
	}
	// Override FindByID to return inviter
	adminRepo2 := &mockAdminRepoWithFindByID{
		MockAdminRepository: adminRepo,
		findByIDFunc: func(ctx context.Context, adminID common.AdminID) (*auth.Admin, error) {
			return inviter, nil
		},
	}

	invitationRepo := &MockInvitationRepository{
		existsPendingByEmailFunc: func(ctx context.Context, tenantID common.TenantID, email string) (bool, error) {
			return false, nil // No pending invitation
		},
		saveFunc: func(ctx context.Context, invitation *auth.Invitation) error {
			return nil
		},
	}

	tenantRepo := &MockTenantRepository{
		findByIDFunc: func(ctx context.Context, tid common.TenantID) (*tenant.Tenant, error) {
			return testTenant, nil
		},
	}

	emailService := &MockEmailService{
		sendInvitationEmailFunc: func(ctx context.Context, input services.SendInvitationEmailInput) error {
			return nil
		},
	}

	clock := &MockClock{nowFunc: func() time.Time { return time.Now() }}

	usecase := NewInviteAdminUsecase(adminRepo2, invitationRepo, tenantRepo, emailService, clock)

	input := InviteAdminInput{
		InviterAdminID: inviter.AdminID().String(),
		Email:          "newadmin@example.com",
		Role:           "manager",
	}

	output, err := usecase.Execute(context.Background(), input)

	if err != nil {
		t.Fatalf("Execute() should succeed, got error: %v", err)
	}

	if output == nil {
		t.Fatal("Output should not be nil")
	}

	if output.Email != "newadmin@example.com" {
		t.Errorf("Email mismatch: got %v, want 'newadmin@example.com'", output.Email)
	}

	if output.Role != "manager" {
		t.Errorf("Role mismatch: got %v, want 'manager'", output.Role)
	}

	if output.Token == "" {
		t.Error("Token should be set")
	}

	if output.InvitationID == "" {
		t.Error("InvitationID should be set")
	}
}

func TestInviteAdminUsecase_Execute_ErrorWhenInviterNotFound(t *testing.T) {
	adminRepo := &mockAdminRepoWithFindByID{
		MockAdminRepository: &MockAdminRepository{},
		findByIDFunc: func(ctx context.Context, adminID common.AdminID) (*auth.Admin, error) {
			return nil, common.NewNotFoundError("Admin", adminID.String())
		},
	}

	invitationRepo := &MockInvitationRepository{}
	tenantRepo := &MockTenantRepository{}
	emailService := &MockEmailService{}
	clock := &MockClock{nowFunc: func() time.Time { return time.Now() }}

	usecase := NewInviteAdminUsecase(adminRepo, invitationRepo, tenantRepo, emailService, clock)

	input := InviteAdminInput{
		InviterAdminID: common.NewAdminID().String(),
		Email:          "newadmin@example.com",
		Role:           "manager",
	}

	_, err := usecase.Execute(context.Background(), input)

	if err == nil {
		t.Fatal("Execute() should fail when inviter not found")
	}
}

func TestInviteAdminUsecase_Execute_ErrorWhenInvalidRole(t *testing.T) {
	inviter := createTestInviterAdmin(t)

	adminRepo := &mockAdminRepoWithFindByID{
		MockAdminRepository: &MockAdminRepository{},
		findByIDFunc: func(ctx context.Context, adminID common.AdminID) (*auth.Admin, error) {
			return inviter, nil
		},
	}

	invitationRepo := &MockInvitationRepository{}
	tenantRepo := &MockTenantRepository{}
	emailService := &MockEmailService{}
	clock := &MockClock{nowFunc: func() time.Time { return time.Now() }}

	usecase := NewInviteAdminUsecase(adminRepo, invitationRepo, tenantRepo, emailService, clock)

	input := InviteAdminInput{
		InviterAdminID: inviter.AdminID().String(),
		Email:          "newadmin@example.com",
		Role:           "invalid_role", // Invalid role
	}

	_, err := usecase.Execute(context.Background(), input)

	if err == nil {
		t.Fatal("Execute() should fail when role is invalid")
	}
}

func TestInviteAdminUsecase_Execute_ErrorWhenAdminEmailAlreadyExists(t *testing.T) {
	inviter := createTestInviterAdmin(t)
	existingAdmin := createTestInviterAdmin(t)

	adminRepo := &mockAdminRepoWithFindByID{
		MockAdminRepository: &MockAdminRepository{
			findByEmailGlobalFunc: func(ctx context.Context, email string) (*auth.Admin, error) {
				return existingAdmin, nil // Admin already exists
			},
		},
		findByIDFunc: func(ctx context.Context, adminID common.AdminID) (*auth.Admin, error) {
			return inviter, nil
		},
	}

	invitationRepo := &MockInvitationRepository{}
	tenantRepo := &MockTenantRepository{}
	emailService := &MockEmailService{}
	clock := &MockClock{nowFunc: func() time.Time { return time.Now() }}

	usecase := NewInviteAdminUsecase(adminRepo, invitationRepo, tenantRepo, emailService, clock)

	input := InviteAdminInput{
		InviterAdminID: inviter.AdminID().String(),
		Email:          "existing@example.com",
		Role:           "manager",
	}

	_, err := usecase.Execute(context.Background(), input)

	if err == nil {
		t.Fatal("Execute() should fail when admin with this email already exists")
	}
}

func TestInviteAdminUsecase_Execute_ErrorWhenPendingInvitationExists(t *testing.T) {
	inviter := createTestInviterAdmin(t)

	adminRepo := &mockAdminRepoWithFindByID{
		MockAdminRepository: &MockAdminRepository{
			findByEmailGlobalFunc: func(ctx context.Context, email string) (*auth.Admin, error) {
				return nil, common.NewNotFoundError("Admin", email)
			},
		},
		findByIDFunc: func(ctx context.Context, adminID common.AdminID) (*auth.Admin, error) {
			return inviter, nil
		},
	}

	invitationRepo := &MockInvitationRepository{
		existsPendingByEmailFunc: func(ctx context.Context, tenantID common.TenantID, email string) (bool, error) {
			return true, nil // Pending invitation exists
		},
	}

	tenantRepo := &MockTenantRepository{}
	emailService := &MockEmailService{}
	clock := &MockClock{nowFunc: func() time.Time { return time.Now() }}

	usecase := NewInviteAdminUsecase(adminRepo, invitationRepo, tenantRepo, emailService, clock)

	input := InviteAdminInput{
		InviterAdminID: inviter.AdminID().String(),
		Email:          "pending@example.com",
		Role:           "manager",
	}

	_, err := usecase.Execute(context.Background(), input)

	if err == nil {
		t.Fatal("Execute() should fail when pending invitation for this email exists")
	}
}

func TestInviteAdminUsecase_Execute_ErrorWhenSaveFails(t *testing.T) {
	inviter := createTestInviterAdmin(t)

	adminRepo := &mockAdminRepoWithFindByID{
		MockAdminRepository: &MockAdminRepository{
			findByEmailGlobalFunc: func(ctx context.Context, email string) (*auth.Admin, error) {
				return nil, common.NewNotFoundError("Admin", email)
			},
		},
		findByIDFunc: func(ctx context.Context, adminID common.AdminID) (*auth.Admin, error) {
			return inviter, nil
		},
	}

	invitationRepo := &MockInvitationRepository{
		existsPendingByEmailFunc: func(ctx context.Context, tenantID common.TenantID, email string) (bool, error) {
			return false, nil
		},
		saveFunc: func(ctx context.Context, invitation *auth.Invitation) error {
			return errors.New("database error")
		},
	}

	tenantRepo := &MockTenantRepository{}
	emailService := &MockEmailService{}
	clock := &MockClock{nowFunc: func() time.Time { return time.Now() }}

	usecase := NewInviteAdminUsecase(adminRepo, invitationRepo, tenantRepo, emailService, clock)

	input := InviteAdminInput{
		InviterAdminID: inviter.AdminID().String(),
		Email:          "newadmin@example.com",
		Role:           "manager",
	}

	_, err := usecase.Execute(context.Background(), input)

	if err == nil {
		t.Fatal("Execute() should fail when save fails")
	}
}

func TestInviteAdminUsecase_Execute_ErrorWhenInvalidInviterAdminID(t *testing.T) {
	adminRepo := &mockAdminRepoWithFindByID{
		MockAdminRepository: &MockAdminRepository{},
		findByIDFunc:        nil,
	}
	invitationRepo := &MockInvitationRepository{}
	tenantRepo := &MockTenantRepository{}
	emailService := &MockEmailService{}
	clock := &MockClock{nowFunc: func() time.Time { return time.Now() }}

	usecase := NewInviteAdminUsecase(adminRepo, invitationRepo, tenantRepo, emailService, clock)

	input := InviteAdminInput{
		InviterAdminID: "invalid-ulid", // Invalid ULID
		Email:          "newadmin@example.com",
		Role:           "manager",
	}

	_, err := usecase.Execute(context.Background(), input)

	if err == nil {
		t.Fatal("Execute() should fail when inviter admin ID is invalid")
	}
}

func TestInviteAdminUsecase_Execute_ErrorWhenEmailSendFails(t *testing.T) {
	tenantID := common.NewTenantID()
	inviter := createTestInviterAdminWithTenantID(t, tenantID)
	testTenant := createTestTenant(t, tenantID)

	var savedInvitationID *auth.InvitationID
	deleteCalled := false

	adminRepo := &mockAdminRepoWithFindByID{
		MockAdminRepository: &MockAdminRepository{
			findByEmailGlobalFunc: func(ctx context.Context, email string) (*auth.Admin, error) {
				return nil, common.NewNotFoundError("Admin", email)
			},
		},
		findByIDFunc: func(ctx context.Context, adminID common.AdminID) (*auth.Admin, error) {
			return inviter, nil
		},
	}

	invitationRepo := &MockInvitationRepository{
		existsPendingByEmailFunc: func(ctx context.Context, tenantID common.TenantID, email string) (bool, error) {
			return false, nil
		},
		saveFunc: func(ctx context.Context, invitation *auth.Invitation) error {
			id := invitation.InvitationID()
			savedInvitationID = &id
			return nil
		},
		deleteFunc: func(ctx context.Context, invitationID auth.InvitationID) error {
			deleteCalled = true
			if savedInvitationID != nil && invitationID == *savedInvitationID {
				return nil
			}
			return errors.New("unexpected invitation ID")
		},
	}

	tenantRepo := &MockTenantRepository{
		findByIDFunc: func(ctx context.Context, tid common.TenantID) (*tenant.Tenant, error) {
			return testTenant, nil
		},
	}

	emailService := &MockEmailService{
		sendInvitationEmailFunc: func(ctx context.Context, input services.SendInvitationEmailInput) error {
			return errors.New("SES error: email send failed")
		},
	}

	clock := &MockClock{nowFunc: func() time.Time { return time.Now() }}

	usecase := NewInviteAdminUsecase(adminRepo, invitationRepo, tenantRepo, emailService, clock)

	input := InviteAdminInput{
		InviterAdminID: inviter.AdminID().String(),
		Email:          "newadmin@example.com",
		Role:           "manager",
	}

	_, err := usecase.Execute(context.Background(), input)

	if err == nil {
		t.Fatal("Execute() should fail when email send fails")
	}

	if !deleteCalled {
		t.Error("Invitation should be deleted (rolled back) when email send fails")
	}

	// Verify error message contains email failure info
	if !contains(err.Error(), "email") {
		t.Errorf("Error message should mention email failure, got: %v", err)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// mockAdminRepoWithFindByID wraps MockAdminRepository to add FindByID
type mockAdminRepoWithFindByID struct {
	*MockAdminRepository
	findByIDFunc func(ctx context.Context, adminID common.AdminID) (*auth.Admin, error)
}

func (m *mockAdminRepoWithFindByID) FindByID(ctx context.Context, adminID common.AdminID) (*auth.Admin, error) {
	if m.findByIDFunc != nil {
		return m.findByIDFunc(ctx, adminID)
	}
	return nil, errors.New("not implemented")
}
