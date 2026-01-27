package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/auth"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/services"
)

// =====================================================
// Mock Implementations for RequestPasswordResetUsecase
// =====================================================

// MockAdminRepositoryForPasswordReset is a mock implementation for password reset tests
type MockAdminRepositoryForPasswordReset struct {
	findByEmailGlobalFunc func(ctx context.Context, email string) (*auth.Admin, error)
	findByIDFunc          func(ctx context.Context, id common.AdminID) (*auth.Admin, error)
	saveFunc              func(ctx context.Context, admin *auth.Admin) error
}

func (m *MockAdminRepositoryForPasswordReset) Save(ctx context.Context, admin *auth.Admin) error {
	if m.saveFunc != nil {
		return m.saveFunc(ctx, admin)
	}
	return nil
}

func (m *MockAdminRepositoryForPasswordReset) FindByID(ctx context.Context, id common.AdminID) (*auth.Admin, error) {
	if m.findByIDFunc != nil {
		return m.findByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockAdminRepositoryForPasswordReset) FindByIDWithTenant(ctx context.Context, tenantID common.TenantID, adminID common.AdminID) (*auth.Admin, error) {
	return nil, nil
}

func (m *MockAdminRepositoryForPasswordReset) FindByEmail(ctx context.Context, tenantID common.TenantID, email string) (*auth.Admin, error) {
	return nil, nil
}

func (m *MockAdminRepositoryForPasswordReset) FindByEmailGlobal(ctx context.Context, email string) (*auth.Admin, error) {
	if m.findByEmailGlobalFunc != nil {
		return m.findByEmailGlobalFunc(ctx, email)
	}
	return nil, nil
}

func (m *MockAdminRepositoryForPasswordReset) FindByTenantID(ctx context.Context, tenantID common.TenantID) ([]*auth.Admin, error) {
	return nil, nil
}

func (m *MockAdminRepositoryForPasswordReset) Delete(ctx context.Context, tenantID common.TenantID, adminID common.AdminID) error {
	return nil
}

func (m *MockAdminRepositoryForPasswordReset) ExistsByEmail(ctx context.Context, tenantID common.TenantID, email string) (bool, error) {
	return false, nil
}

func (m *MockAdminRepositoryForPasswordReset) ExistsByEmailGlobal(ctx context.Context, email string) (bool, error) {
	return false, nil
}

func (m *MockAdminRepositoryForPasswordReset) FindActiveByTenantID(ctx context.Context, tenantID common.TenantID) ([]*auth.Admin, error) {
	return nil, nil
}

// MockPasswordResetTokenRepository is a mock implementation for password reset token tests
type MockPasswordResetTokenRepository struct {
	saveFunc                   func(ctx context.Context, token *auth.PasswordResetToken) error
	findByTokenFunc            func(ctx context.Context, token string) (*auth.PasswordResetToken, error)
	findValidByAdminIDFunc     func(ctx context.Context, adminID common.AdminID) (*auth.PasswordResetToken, error)
	invalidateAllByAdminIDFunc func(ctx context.Context, adminID common.AdminID) error
}

func (m *MockPasswordResetTokenRepository) Save(ctx context.Context, token *auth.PasswordResetToken) error {
	if m.saveFunc != nil {
		return m.saveFunc(ctx, token)
	}
	return nil
}

func (m *MockPasswordResetTokenRepository) FindByToken(ctx context.Context, token string) (*auth.PasswordResetToken, error) {
	if m.findByTokenFunc != nil {
		return m.findByTokenFunc(ctx, token)
	}
	return nil, nil
}

func (m *MockPasswordResetTokenRepository) FindValidByAdminID(ctx context.Context, adminID common.AdminID) (*auth.PasswordResetToken, error) {
	if m.findValidByAdminIDFunc != nil {
		return m.findValidByAdminIDFunc(ctx, adminID)
	}
	return nil, nil
}

func (m *MockPasswordResetTokenRepository) InvalidateAllByAdminID(ctx context.Context, adminID common.AdminID) error {
	if m.invalidateAllByAdminIDFunc != nil {
		return m.invalidateAllByAdminIDFunc(ctx, adminID)
	}
	return nil
}

// MockPasswordResetClock is a mock clock for password reset tests
type MockPasswordResetClock struct {
	now time.Time
}

func (m *MockPasswordResetClock) Now() time.Time {
	return m.now
}

// createTestAdminForPasswordReset creates a test admin for testing
func createTestAdminForPasswordReset(t *testing.T, now time.Time, isActive bool, isDeleted bool) *auth.Admin {
	t.Helper()
	tenantID := common.NewTenantID()
	admin, err := auth.NewAdmin(
		now,
		tenantID,
		"test@example.com",
		"hashedpassword123",
		"Test Admin",
		auth.RoleOwner,
	)
	if err != nil {
		t.Fatalf("Failed to create test admin: %v", err)
	}

	if !isActive {
		admin.Deactivate(now)
	}
	if isDeleted {
		admin.Delete(now)
	}

	return admin
}

// =====================================================
// Tests for RequestPasswordResetUsecase
// =====================================================

func TestRequestPasswordResetUsecase_Execute_Success(t *testing.T) {
	now := time.Now()
	testAdmin := createTestAdminForPasswordReset(t, now, true, false)
	emailSent := false

	adminRepo := &MockAdminRepositoryForPasswordReset{
		findByEmailGlobalFunc: func(ctx context.Context, email string) (*auth.Admin, error) {
			return testAdmin, nil
		},
	}

	tokenRepo := &MockPasswordResetTokenRepository{
		invalidateAllByAdminIDFunc: func(ctx context.Context, adminID common.AdminID) error {
			return nil
		},
		saveFunc: func(ctx context.Context, token *auth.PasswordResetToken) error {
			return nil
		},
	}

	emailService := &MockEmailService{
		sendPasswordResetEmailFunc: func(ctx context.Context, input services.SendPasswordResetEmailInput) error {
			emailSent = true
			return nil
		},
	}

	clock := &MockPasswordResetClock{now: now}

	usecase := NewRequestPasswordResetUsecase(adminRepo, tokenRepo, emailService, clock)

	output, err := usecase.Execute(context.Background(), RequestPasswordResetInput{
		Email: "test@example.com",
	})

	if err != nil {
		t.Fatalf("Execute() should succeed, got error: %v", err)
	}

	if !output.Success {
		t.Error("Output.Success should be true")
	}

	if !emailSent {
		t.Error("Email should have been sent")
	}
}

func TestRequestPasswordResetUsecase_Execute_NonExistentEmail(t *testing.T) {
	now := time.Now()

	adminRepo := &MockAdminRepositoryForPasswordReset{
		findByEmailGlobalFunc: func(ctx context.Context, email string) (*auth.Admin, error) {
			return nil, common.NewDomainError(common.ErrNotFound, "Admin not found")
		},
	}

	tokenRepo := &MockPasswordResetTokenRepository{}
	emailService := &MockEmailService{}
	clock := &MockPasswordResetClock{now: now}

	usecase := NewRequestPasswordResetUsecase(adminRepo, tokenRepo, emailService, clock)

	output, err := usecase.Execute(context.Background(), RequestPasswordResetInput{
		Email: "nonexistent@example.com",
	})

	// Should return success to prevent user enumeration
	if err != nil {
		t.Fatalf("Execute() should succeed even for non-existent email, got error: %v", err)
	}

	if !output.Success {
		t.Error("Output.Success should be true (user enumeration prevention)")
	}
}

func TestRequestPasswordResetUsecase_Execute_InactiveAdmin(t *testing.T) {
	now := time.Now()
	testAdmin := createTestAdminForPasswordReset(t, now, false, false) // inactive

	adminRepo := &MockAdminRepositoryForPasswordReset{
		findByEmailGlobalFunc: func(ctx context.Context, email string) (*auth.Admin, error) {
			return testAdmin, nil
		},
	}

	tokenRepo := &MockPasswordResetTokenRepository{}
	emailService := &MockEmailService{}
	clock := &MockPasswordResetClock{now: now}

	usecase := NewRequestPasswordResetUsecase(adminRepo, tokenRepo, emailService, clock)

	output, err := usecase.Execute(context.Background(), RequestPasswordResetInput{
		Email: "test@example.com",
	})

	// Should return success to prevent user enumeration
	if err != nil {
		t.Fatalf("Execute() should succeed for inactive admin, got error: %v", err)
	}

	if !output.Success {
		t.Error("Output.Success should be true (user enumeration prevention)")
	}
}

func TestRequestPasswordResetUsecase_Execute_DeletedAdmin(t *testing.T) {
	now := time.Now()
	testAdmin := createTestAdminForPasswordReset(t, now, true, true) // deleted

	adminRepo := &MockAdminRepositoryForPasswordReset{
		findByEmailGlobalFunc: func(ctx context.Context, email string) (*auth.Admin, error) {
			return testAdmin, nil
		},
	}

	tokenRepo := &MockPasswordResetTokenRepository{}
	emailService := &MockEmailService{}
	clock := &MockPasswordResetClock{now: now}

	usecase := NewRequestPasswordResetUsecase(adminRepo, tokenRepo, emailService, clock)

	output, err := usecase.Execute(context.Background(), RequestPasswordResetInput{
		Email: "test@example.com",
	})

	// Should return success to prevent user enumeration
	if err != nil {
		t.Fatalf("Execute() should succeed for deleted admin, got error: %v", err)
	}

	if !output.Success {
		t.Error("Output.Success should be true (user enumeration prevention)")
	}
}

func TestRequestPasswordResetUsecase_Execute_EmailSendFailure(t *testing.T) {
	now := time.Now()
	testAdmin := createTestAdminForPasswordReset(t, now, true, false)

	adminRepo := &MockAdminRepositoryForPasswordReset{
		findByEmailGlobalFunc: func(ctx context.Context, email string) (*auth.Admin, error) {
			return testAdmin, nil
		},
	}

	tokenRepo := &MockPasswordResetTokenRepository{
		invalidateAllByAdminIDFunc: func(ctx context.Context, adminID common.AdminID) error {
			return nil
		},
		saveFunc: func(ctx context.Context, token *auth.PasswordResetToken) error {
			return nil
		},
	}

	emailService := &MockEmailService{
		sendPasswordResetEmailFunc: func(ctx context.Context, input services.SendPasswordResetEmailInput) error {
			return errors.New("email send failed")
		},
	}

	clock := &MockPasswordResetClock{now: now}

	usecase := NewRequestPasswordResetUsecase(adminRepo, tokenRepo, emailService, clock)

	output, err := usecase.Execute(context.Background(), RequestPasswordResetInput{
		Email: "test@example.com",
	})

	// Should return success even if email fails (to prevent timing attacks)
	if err != nil {
		t.Fatalf("Execute() should succeed even if email fails, got error: %v", err)
	}

	if !output.Success {
		t.Error("Output.Success should be true (timing attack prevention)")
	}
}
