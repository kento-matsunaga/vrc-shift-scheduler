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
// ChangePasswordUsecase Tests - Success Cases
// =====================================================

func TestChangePasswordUsecase_Execute_Success(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	testAdmin, _ := auth.NewAdmin(now, tenantID, "test@example.com", "$2a$10$oldhash", "Test Admin", auth.RoleOwner)
	adminID := testAdmin.AdminID()

	var savedAdmin *auth.Admin

	mockRepo := &MockAdminRepository{
		findByIDWithTenantFunc: func(ctx context.Context, tID common.TenantID, aID common.AdminID) (*auth.Admin, error) {
			if tID == tenantID && aID == adminID {
				return testAdmin, nil
			}
			return nil, errors.New("not found")
		},
		saveFunc: func(ctx context.Context, admin *auth.Admin) error {
			savedAdmin = admin
			return nil
		},
	}

	mockHasher := &MockPasswordHasher{
		compareFunc: func(hash, password string) error {
			if password == "currentpassword" {
				return nil
			}
			return errors.New("password mismatch")
		},
		hashFunc: func(password string) (string, error) {
			return "$2a$10$newhash", nil
		},
	}

	usecase := NewChangePasswordUsecase(mockRepo, mockHasher)

	input := ChangePasswordInput{
		AdminID:         adminID,
		TenantID:        tenantID,
		CurrentPassword: "currentpassword",
		NewPassword:     "newpassword123",
	}

	err := usecase.Execute(context.Background(), input)

	if err != nil {
		t.Fatalf("Execute() should succeed, but got error: %v", err)
	}

	if savedAdmin == nil {
		t.Fatal("Admin should have been saved")
	}

	if savedAdmin.PasswordHash() != "$2a$10$newhash" {
		t.Errorf("PasswordHash: expected $2a$10$newhash, got %s", savedAdmin.PasswordHash())
	}
}

// =====================================================
// ChangePasswordUsecase Tests - Error Cases
// =====================================================

func TestChangePasswordUsecase_Execute_ErrorWhenAdminNotFound(t *testing.T) {
	tenantID := common.NewTenantID()
	adminID := common.NewAdminID()

	mockRepo := &MockAdminRepository{
		findByIDWithTenantFunc: func(ctx context.Context, tID common.TenantID, aID common.AdminID) (*auth.Admin, error) {
			return nil, errors.New("not found")
		},
	}

	mockHasher := &MockPasswordHasher{}

	usecase := NewChangePasswordUsecase(mockRepo, mockHasher)

	input := ChangePasswordInput{
		AdminID:         adminID,
		TenantID:        tenantID,
		CurrentPassword: "password",
		NewPassword:     "newpassword",
	}

	err := usecase.Execute(context.Background(), input)

	if err == nil {
		t.Fatal("Execute() should return error when admin is not found")
	}
}

func TestChangePasswordUsecase_Execute_ErrorWhenCurrentPasswordIncorrect(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	testAdmin, _ := auth.NewAdmin(now, tenantID, "test@example.com", "$2a$10$oldhash", "Test Admin", auth.RoleOwner)
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

	usecase := NewChangePasswordUsecase(mockRepo, mockHasher)

	input := ChangePasswordInput{
		AdminID:         adminID,
		TenantID:        tenantID,
		CurrentPassword: "wrongpassword",
		NewPassword:     "newpassword",
	}

	err := usecase.Execute(context.Background(), input)

	if !errors.Is(err, ErrInvalidCredentials) {
		t.Errorf("Expected ErrInvalidCredentials, got %v", err)
	}
}

func TestChangePasswordUsecase_Execute_ErrorWhenHashFails(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	testAdmin, _ := auth.NewAdmin(now, tenantID, "test@example.com", "$2a$10$oldhash", "Test Admin", auth.RoleOwner)
	adminID := testAdmin.AdminID()

	mockRepo := &MockAdminRepository{
		findByIDWithTenantFunc: func(ctx context.Context, tID common.TenantID, aID common.AdminID) (*auth.Admin, error) {
			return testAdmin, nil
		},
	}

	mockHasher := &MockPasswordHasher{
		compareFunc: func(hash, password string) error {
			return nil // Current password is correct
		},
		hashFunc: func(password string) (string, error) {
			return "", errors.New("hash failed")
		},
	}

	usecase := NewChangePasswordUsecase(mockRepo, mockHasher)

	input := ChangePasswordInput{
		AdminID:         adminID,
		TenantID:        tenantID,
		CurrentPassword: "currentpassword",
		NewPassword:     "newpassword",
	}

	err := usecase.Execute(context.Background(), input)

	if err == nil {
		t.Fatal("Execute() should return error when hashing fails")
	}
}

func TestChangePasswordUsecase_Execute_ErrorWhenSaveFails(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	testAdmin, _ := auth.NewAdmin(now, tenantID, "test@example.com", "$2a$10$oldhash", "Test Admin", auth.RoleOwner)
	adminID := testAdmin.AdminID()

	saveError := errors.New("database error")

	mockRepo := &MockAdminRepository{
		findByIDWithTenantFunc: func(ctx context.Context, tID common.TenantID, aID common.AdminID) (*auth.Admin, error) {
			return testAdmin, nil
		},
		saveFunc: func(ctx context.Context, admin *auth.Admin) error {
			return saveError
		},
	}

	mockHasher := &MockPasswordHasher{
		compareFunc: func(hash, password string) error {
			return nil
		},
		hashFunc: func(password string) (string, error) {
			return "$2a$10$newhash", nil
		},
	}

	usecase := NewChangePasswordUsecase(mockRepo, mockHasher)

	input := ChangePasswordInput{
		AdminID:         adminID,
		TenantID:        tenantID,
		CurrentPassword: "currentpassword",
		NewPassword:     "newpassword",
	}

	err := usecase.Execute(context.Background(), input)

	if !errors.Is(err, saveError) {
		t.Errorf("Expected save error, got %v", err)
	}
}

// =====================================================
// ChangePasswordUsecase Tests - Cross-Tenant Security
// =====================================================

func TestChangePasswordUsecase_Execute_ErrorWhenTenantMismatch(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	differentTenantID := common.NewTenantID()
	testAdmin, _ := auth.NewAdmin(now, tenantID, "test@example.com", "$2a$10$oldhash", "Test Admin", auth.RoleOwner)
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

	usecase := NewChangePasswordUsecase(mockRepo, mockHasher)

	// Try to change password with a different tenant ID
	input := ChangePasswordInput{
		AdminID:         adminID,
		TenantID:        differentTenantID, // Wrong tenant
		CurrentPassword: "currentpassword",
		NewPassword:     "newpassword",
	}

	err := usecase.Execute(context.Background(), input)

	if err == nil {
		t.Fatal("Execute() should return error when tenant ID doesn't match")
	}
}
