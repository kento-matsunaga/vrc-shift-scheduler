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
// ChangeEmailUsecase Tests - Success Cases
// =====================================================

func TestChangeEmailUsecase_Execute_Success(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	testAdmin, _ := auth.NewAdmin(now, tenantID, "old@example.com", "$2a$10$hash", "Test Admin", auth.RoleOwner)
	adminID := testAdmin.AdminID()

	var savedAdmin *auth.Admin

	mockRepo := &MockAdminRepository{
		findByIDWithTenantFunc: func(ctx context.Context, tID common.TenantID, aID common.AdminID) (*auth.Admin, error) {
			if tID == tenantID && aID == adminID {
				return testAdmin, nil
			}
			return nil, errors.New("not found")
		},
		existsByEmailGlobalFunc: func(ctx context.Context, email string) (bool, error) {
			return false, nil // Email does not exist
		},
		saveFunc: func(ctx context.Context, admin *auth.Admin) error {
			savedAdmin = admin
			return nil
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

	usecase := NewChangeEmailUsecase(mockRepo, mockHasher)

	input := ChangeEmailInput{
		AdminID:         adminID,
		TenantID:        tenantID,
		CurrentPassword: "correctpassword",
		NewEmail:        "new@example.com",
	}

	err := usecase.Execute(context.Background(), input)

	if err != nil {
		t.Fatalf("Execute() should succeed, but got error: %v", err)
	}

	if savedAdmin == nil {
		t.Fatal("Admin should have been saved")
	}

	if savedAdmin.Email() != "new@example.com" {
		t.Errorf("Email: expected new@example.com, got %s", savedAdmin.Email())
	}
}

// =====================================================
// ChangeEmailUsecase Tests - Error Cases
// =====================================================

func TestChangeEmailUsecase_Execute_ErrorWhenAdminNotFound(t *testing.T) {
	tenantID := common.NewTenantID()
	adminID := common.NewAdminID()

	mockRepo := &MockAdminRepository{
		findByIDWithTenantFunc: func(ctx context.Context, tID common.TenantID, aID common.AdminID) (*auth.Admin, error) {
			return nil, errors.New("not found")
		},
	}

	mockHasher := &MockPasswordHasher{}

	usecase := NewChangeEmailUsecase(mockRepo, mockHasher)

	input := ChangeEmailInput{
		AdminID:         adminID,
		TenantID:        tenantID,
		CurrentPassword: "password",
		NewEmail:        "new@example.com",
	}

	err := usecase.Execute(context.Background(), input)

	if err == nil {
		t.Fatal("Execute() should return error when admin is not found")
	}
}

func TestChangeEmailUsecase_Execute_ErrorWhenPasswordIncorrect(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	testAdmin, _ := auth.NewAdmin(now, tenantID, "old@example.com", "$2a$10$hash", "Test Admin", auth.RoleOwner)
	adminID := testAdmin.AdminID()

	mockRepo := &MockAdminRepository{
		findByIDWithTenantFunc: func(ctx context.Context, tID common.TenantID, aID common.AdminID) (*auth.Admin, error) {
			return testAdmin, nil
		},
	}

	mockHasher := &MockPasswordHasher{
		compareFunc: func(hash, password string) error {
			return errors.New("password mismatch")
		},
	}

	usecase := NewChangeEmailUsecase(mockRepo, mockHasher)

	input := ChangeEmailInput{
		AdminID:         adminID,
		TenantID:        tenantID,
		CurrentPassword: "wrongpassword",
		NewEmail:        "new@example.com",
	}

	err := usecase.Execute(context.Background(), input)

	if !errors.Is(err, ErrInvalidCredentials) {
		t.Errorf("Expected ErrInvalidCredentials, got %v", err)
	}
}

func TestChangeEmailUsecase_Execute_ErrorWhenInvalidEmailFormat(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	testAdmin, _ := auth.NewAdmin(now, tenantID, "old@example.com", "$2a$10$hash", "Test Admin", auth.RoleOwner)
	adminID := testAdmin.AdminID()

	mockRepo := &MockAdminRepository{
		findByIDWithTenantFunc: func(ctx context.Context, tID common.TenantID, aID common.AdminID) (*auth.Admin, error) {
			return testAdmin, nil
		},
	}

	mockHasher := &MockPasswordHasher{
		compareFunc: func(hash, password string) error {
			return nil // Password is correct
		},
	}

	usecase := NewChangeEmailUsecase(mockRepo, mockHasher)

	tests := []struct {
		name     string
		newEmail string
	}{
		{"missing @ symbol", "invalidemail.com"},
		{"missing domain", "invalid@"},
		{"missing local part", "@example.com"},
		{"missing dot in domain", "test@examplecom"},
		{"too short", "a@b"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := ChangeEmailInput{
				AdminID:         adminID,
				TenantID:        tenantID,
				CurrentPassword: "correctpassword",
				NewEmail:        tt.newEmail,
			}

			err := usecase.Execute(context.Background(), input)

			if err == nil {
				t.Errorf("Execute() should return error for invalid email: %s", tt.newEmail)
			}

			// Should be a validation error
			var validationErr *common.DomainError
			if !errors.As(err, &validationErr) {
				t.Errorf("Expected validation error for %s, got %v", tt.newEmail, err)
			}
		})
	}
}

func TestChangeEmailUsecase_Execute_ErrorWhenSameEmail(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	testAdmin, _ := auth.NewAdmin(now, tenantID, "same@example.com", "$2a$10$hash", "Test Admin", auth.RoleOwner)
	adminID := testAdmin.AdminID()

	mockRepo := &MockAdminRepository{
		findByIDWithTenantFunc: func(ctx context.Context, tID common.TenantID, aID common.AdminID) (*auth.Admin, error) {
			return testAdmin, nil
		},
	}

	mockHasher := &MockPasswordHasher{
		compareFunc: func(hash, password string) error {
			return nil // Password is correct
		},
	}

	usecase := NewChangeEmailUsecase(mockRepo, mockHasher)

	input := ChangeEmailInput{
		AdminID:         adminID,
		TenantID:        tenantID,
		CurrentPassword: "correctpassword",
		NewEmail:        "same@example.com", // Same as current email
	}

	err := usecase.Execute(context.Background(), input)

	if err == nil {
		t.Fatal("Execute() should return error when new email is same as current")
	}

	// Should be a validation error
	var validationErr *common.DomainError
	if !errors.As(err, &validationErr) {
		t.Errorf("Expected validation error, got %v", err)
	}
}

func TestChangeEmailUsecase_Execute_ErrorWhenEmailAlreadyExists(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	testAdmin, _ := auth.NewAdmin(now, tenantID, "old@example.com", "$2a$10$hash", "Test Admin", auth.RoleOwner)
	adminID := testAdmin.AdminID()

	mockRepo := &MockAdminRepository{
		findByIDWithTenantFunc: func(ctx context.Context, tID common.TenantID, aID common.AdminID) (*auth.Admin, error) {
			return testAdmin, nil
		},
		existsByEmailGlobalFunc: func(ctx context.Context, email string) (bool, error) {
			return true, nil // Email already exists
		},
	}

	mockHasher := &MockPasswordHasher{
		compareFunc: func(hash, password string) error {
			return nil // Password is correct
		},
	}

	usecase := NewChangeEmailUsecase(mockRepo, mockHasher)

	input := ChangeEmailInput{
		AdminID:         adminID,
		TenantID:        tenantID,
		CurrentPassword: "correctpassword",
		NewEmail:        "existing@example.com",
	}

	err := usecase.Execute(context.Background(), input)

	if !errors.Is(err, ErrEmailAlreadyExists) {
		t.Errorf("Expected ErrEmailAlreadyExists, got %v", err)
	}
}

func TestChangeEmailUsecase_Execute_ErrorWhenSaveFails(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	testAdmin, _ := auth.NewAdmin(now, tenantID, "old@example.com", "$2a$10$hash", "Test Admin", auth.RoleOwner)
	adminID := testAdmin.AdminID()

	saveError := errors.New("database error")

	mockRepo := &MockAdminRepository{
		findByIDWithTenantFunc: func(ctx context.Context, tID common.TenantID, aID common.AdminID) (*auth.Admin, error) {
			return testAdmin, nil
		},
		existsByEmailGlobalFunc: func(ctx context.Context, email string) (bool, error) {
			return false, nil
		},
		saveFunc: func(ctx context.Context, admin *auth.Admin) error {
			return saveError
		},
	}

	mockHasher := &MockPasswordHasher{
		compareFunc: func(hash, password string) error {
			return nil
		},
	}

	usecase := NewChangeEmailUsecase(mockRepo, mockHasher)

	input := ChangeEmailInput{
		AdminID:         adminID,
		TenantID:        tenantID,
		CurrentPassword: "correctpassword",
		NewEmail:        "new@example.com",
	}

	err := usecase.Execute(context.Background(), input)

	if !errors.Is(err, saveError) {
		t.Errorf("Expected save error, got %v", err)
	}
}

// =====================================================
// ChangeEmailUsecase Tests - Cross-Tenant Security
// =====================================================

func TestChangeEmailUsecase_Execute_ErrorWhenTenantMismatch(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	differentTenantID := common.NewTenantID()
	testAdmin, _ := auth.NewAdmin(now, tenantID, "old@example.com", "$2a$10$hash", "Test Admin", auth.RoleOwner)
	adminID := testAdmin.AdminID()

	mockRepo := &MockAdminRepository{
		findByIDWithTenantFunc: func(ctx context.Context, tID common.TenantID, aID common.AdminID) (*auth.Admin, error) {
			// Only return admin when tenant matches
			if tID == tenantID && aID == adminID {
				return testAdmin, nil
			}
			return nil, errors.New("not found")
		},
	}

	mockHasher := &MockPasswordHasher{}

	usecase := NewChangeEmailUsecase(mockRepo, mockHasher)

	// Try to change email with a different tenant ID
	input := ChangeEmailInput{
		AdminID:         adminID,
		TenantID:        differentTenantID, // Wrong tenant
		CurrentPassword: "correctpassword",
		NewEmail:        "new@example.com",
	}

	err := usecase.Execute(context.Background(), input)

	if err == nil {
		t.Fatal("Execute() should return error when tenant ID doesn't match")
	}
}
