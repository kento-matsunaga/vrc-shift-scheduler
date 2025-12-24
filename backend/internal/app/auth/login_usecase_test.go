package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/auth"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
)

// =====================================================
// Mock Implementations
// =====================================================

// MockAdminRepository is a mock implementation of auth.AdminRepository
type MockAdminRepository struct {
	findByEmailGlobalFunc  func(ctx context.Context, email string) (*auth.Admin, error)
	findByIDWithTenantFunc func(ctx context.Context, tenantID common.TenantID, adminID common.AdminID) (*auth.Admin, error)
	saveFunc               func(ctx context.Context, admin *auth.Admin) error
}

func (m *MockAdminRepository) FindByEmailGlobal(ctx context.Context, email string) (*auth.Admin, error) {
	if m.findByEmailGlobalFunc != nil {
		return m.findByEmailGlobalFunc(ctx, email)
	}
	return nil, errors.New("not implemented")
}

func (m *MockAdminRepository) FindByIDWithTenant(ctx context.Context, tenantID common.TenantID, adminID common.AdminID) (*auth.Admin, error) {
	if m.findByIDWithTenantFunc != nil {
		return m.findByIDWithTenantFunc(ctx, tenantID, adminID)
	}
	return nil, errors.New("not implemented")
}

func (m *MockAdminRepository) Save(ctx context.Context, admin *auth.Admin) error {
	if m.saveFunc != nil {
		return m.saveFunc(ctx, admin)
	}
	return nil
}

// Unused methods - just satisfy the interface
func (m *MockAdminRepository) FindByID(ctx context.Context, adminID common.AdminID) (*auth.Admin, error) {
	return nil, errors.New("not implemented")
}
func (m *MockAdminRepository) FindByEmail(ctx context.Context, tenantID common.TenantID, email string) (*auth.Admin, error) {
	return nil, errors.New("not implemented")
}
func (m *MockAdminRepository) FindByTenantID(ctx context.Context, tenantID common.TenantID) ([]*auth.Admin, error) {
	return nil, errors.New("not implemented")
}
func (m *MockAdminRepository) FindActiveByTenantID(ctx context.Context, tenantID common.TenantID) ([]*auth.Admin, error) {
	return nil, errors.New("not implemented")
}
func (m *MockAdminRepository) Delete(ctx context.Context, tenantID common.TenantID, adminID common.AdminID) error {
	return errors.New("not implemented")
}
func (m *MockAdminRepository) ExistsByEmail(ctx context.Context, tenantID common.TenantID, email string) (bool, error) {
	return false, errors.New("not implemented")
}

// MockPasswordHasher is a mock implementation of services.PasswordHasher
type MockPasswordHasher struct {
	hashFunc    func(password string) (string, error)
	compareFunc func(hash, password string) error
}

func (m *MockPasswordHasher) Hash(password string) (string, error) {
	if m.hashFunc != nil {
		return m.hashFunc(password)
	}
	return "$2a$10$mockhash", nil
}

func (m *MockPasswordHasher) Compare(hash, password string) error {
	if m.compareFunc != nil {
		return m.compareFunc(hash, password)
	}
	return nil
}

// MockTokenIssuer is a mock implementation of services.TokenIssuer
type MockTokenIssuer struct {
	issueFunc func(adminID, tenantID, role string) (string, time.Time, error)
}

func (m *MockTokenIssuer) Issue(adminID, tenantID, role string) (string, time.Time, error) {
	if m.issueFunc != nil {
		return m.issueFunc(adminID, tenantID, role)
	}
	return "mock-token", time.Now().Add(24 * time.Hour), nil
}

// =====================================================
// Test Helper Functions
// =====================================================

func createTestAdmin(t *testing.T, email string, passwordHash string, isActive bool) *auth.Admin {
	now := time.Now()
	tenantID := common.NewTenantID()
	admin, err := auth.NewAdmin(now, tenantID, email, passwordHash, "Test Admin", auth.RoleOwner)
	if err != nil {
		t.Fatalf("Failed to create test admin: %v", err)
	}
	if !isActive {
		admin.Deactivate(now)
	}
	return admin
}

// =====================================================
// LoginUsecase Tests - Success Cases
// =====================================================

func TestLoginUsecase_Execute_Success(t *testing.T) {
	testAdmin := createTestAdmin(t, "test@example.com", "$2a$10$hashedpassword", true)

	mockRepo := &MockAdminRepository{
		findByEmailGlobalFunc: func(ctx context.Context, email string) (*auth.Admin, error) {
			if email == "test@example.com" {
				return testAdmin, nil
			}
			return nil, errors.New("not found")
		},
	}

	mockHasher := &MockPasswordHasher{
		compareFunc: func(hash, password string) error {
			if password == "correctpassword" {
				return nil
			}
			return errors.New("password mismatch")
		},
	}

	expectedToken := "jwt-token-abc123"
	expectedExpires := time.Now().Add(24 * time.Hour)
	mockIssuer := &MockTokenIssuer{
		issueFunc: func(adminID, tenantID, role string) (string, time.Time, error) {
			return expectedToken, expectedExpires, nil
		},
	}

	usecase := NewLoginUsecase(mockRepo, mockHasher, mockIssuer)

	input := LoginInput{
		Email:    "test@example.com",
		Password: "correctpassword",
	}

	output, err := usecase.Execute(context.Background(), input)

	if err != nil {
		t.Fatalf("Execute() should succeed, but got error: %v", err)
	}

	if output == nil {
		t.Fatal("Execute() returned nil output")
	}

	if output.Token != expectedToken {
		t.Errorf("Token: expected %s, got %s", expectedToken, output.Token)
	}

	if output.AdminID != testAdmin.AdminID().String() {
		t.Errorf("AdminID: expected %s, got %s", testAdmin.AdminID().String(), output.AdminID)
	}

	if output.TenantID != testAdmin.TenantID().String() {
		t.Errorf("TenantID: expected %s, got %s", testAdmin.TenantID().String(), output.TenantID)
	}

	if output.Role != testAdmin.Role().String() {
		t.Errorf("Role: expected %s, got %s", testAdmin.Role().String(), output.Role)
	}
}

// =====================================================
// LoginUsecase Tests - Error Cases
// =====================================================

func TestLoginUsecase_Execute_ErrorWhenEmailNotFound(t *testing.T) {
	mockRepo := &MockAdminRepository{
		findByEmailGlobalFunc: func(ctx context.Context, email string) (*auth.Admin, error) {
			return nil, errors.New("not found")
		},
	}

	mockHasher := &MockPasswordHasher{}
	mockIssuer := &MockTokenIssuer{}

	usecase := NewLoginUsecase(mockRepo, mockHasher, mockIssuer)

	input := LoginInput{
		Email:    "nonexistent@example.com",
		Password: "password",
	}

	output, err := usecase.Execute(context.Background(), input)

	if !errors.Is(err, ErrInvalidCredentials) {
		t.Errorf("Expected ErrInvalidCredentials, got %v", err)
	}

	if output != nil {
		t.Error("Output should be nil when error occurs")
	}
}

func TestLoginUsecase_Execute_ErrorWhenPasswordIncorrect(t *testing.T) {
	testAdmin := createTestAdmin(t, "test@example.com", "$2a$10$hashedpassword", true)

	mockRepo := &MockAdminRepository{
		findByEmailGlobalFunc: func(ctx context.Context, email string) (*auth.Admin, error) {
			return testAdmin, nil
		},
	}

	mockHasher := &MockPasswordHasher{
		compareFunc: func(hash, password string) error {
			return errors.New("password mismatch")
		},
	}

	mockIssuer := &MockTokenIssuer{}

	usecase := NewLoginUsecase(mockRepo, mockHasher, mockIssuer)

	input := LoginInput{
		Email:    "test@example.com",
		Password: "wrongpassword",
	}

	output, err := usecase.Execute(context.Background(), input)

	if !errors.Is(err, ErrInvalidCredentials) {
		t.Errorf("Expected ErrInvalidCredentials, got %v", err)
	}

	if output != nil {
		t.Error("Output should be nil when error occurs")
	}
}

func TestLoginUsecase_Execute_ErrorWhenAccountDisabled(t *testing.T) {
	// Create inactive admin
	testAdmin := createTestAdmin(t, "test@example.com", "$2a$10$hashedpassword", false)

	mockRepo := &MockAdminRepository{
		findByEmailGlobalFunc: func(ctx context.Context, email string) (*auth.Admin, error) {
			return testAdmin, nil
		},
	}

	mockHasher := &MockPasswordHasher{}
	mockIssuer := &MockTokenIssuer{}

	usecase := NewLoginUsecase(mockRepo, mockHasher, mockIssuer)

	input := LoginInput{
		Email:    "test@example.com",
		Password: "password",
	}

	output, err := usecase.Execute(context.Background(), input)

	if !errors.Is(err, ErrAccountDisabled) {
		t.Errorf("Expected ErrAccountDisabled, got %v", err)
	}

	if output != nil {
		t.Error("Output should be nil when error occurs")
	}
}

func TestLoginUsecase_Execute_ErrorWhenTokenIssueFails(t *testing.T) {
	testAdmin := createTestAdmin(t, "test@example.com", "$2a$10$hashedpassword", true)

	mockRepo := &MockAdminRepository{
		findByEmailGlobalFunc: func(ctx context.Context, email string) (*auth.Admin, error) {
			return testAdmin, nil
		},
	}

	mockHasher := &MockPasswordHasher{
		compareFunc: func(hash, password string) error {
			return nil
		},
	}

	tokenError := errors.New("token issue failed")
	mockIssuer := &MockTokenIssuer{
		issueFunc: func(adminID, tenantID, role string) (string, time.Time, error) {
			return "", time.Time{}, tokenError
		},
	}

	usecase := NewLoginUsecase(mockRepo, mockHasher, mockIssuer)

	input := LoginInput{
		Email:    "test@example.com",
		Password: "password",
	}

	output, err := usecase.Execute(context.Background(), input)

	if err == nil {
		t.Fatal("Expected error when token issue fails")
	}

	if output != nil {
		t.Error("Output should be nil when error occurs")
	}
}

// =====================================================
// LoginUsecase Tests - Security Considerations
// =====================================================

func TestLoginUsecase_Execute_SameErrorForNonExistentAndWrongPassword(t *testing.T) {
	// This test ensures that the error message doesn't leak information
	// about whether the email exists or not

	testAdmin := createTestAdmin(t, "test@example.com", "$2a$10$hashedpassword", true)

	mockRepo := &MockAdminRepository{
		findByEmailGlobalFunc: func(ctx context.Context, email string) (*auth.Admin, error) {
			if email == "test@example.com" {
				return testAdmin, nil
			}
			return nil, errors.New("not found")
		},
	}

	mockHasher := &MockPasswordHasher{
		compareFunc: func(hash, password string) error {
			return errors.New("password mismatch")
		},
	}

	mockIssuer := &MockTokenIssuer{}

	usecase := NewLoginUsecase(mockRepo, mockHasher, mockIssuer)

	// Test with non-existent email
	_, errNonExistent := usecase.Execute(context.Background(), LoginInput{
		Email:    "nonexistent@example.com",
		Password: "password",
	})

	// Test with existing email but wrong password
	_, errWrongPassword := usecase.Execute(context.Background(), LoginInput{
		Email:    "test@example.com",
		Password: "wrongpassword",
	})

	// Both should return the same error type (security best practice)
	if !errors.Is(errNonExistent, ErrInvalidCredentials) {
		t.Errorf("Non-existent email should return ErrInvalidCredentials, got %v", errNonExistent)
	}

	if !errors.Is(errWrongPassword, ErrInvalidCredentials) {
		t.Errorf("Wrong password should return ErrInvalidCredentials, got %v", errWrongPassword)
	}
}
