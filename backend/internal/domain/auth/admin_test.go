package auth

import (
	"testing"
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
)

// =====================================================
// NewAdmin Tests - Success Cases
// =====================================================

func TestNewAdmin_Success(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	email := "test@example.com"
	passwordHash := "$2a$10$hashedpassword"
	displayName := "Test Admin"
	role := RoleOwner

	admin, err := NewAdmin(now, tenantID, email, passwordHash, displayName, role)

	if err != nil {
		t.Fatalf("NewAdmin() should succeed, but got error: %v", err)
	}

	if admin == nil {
		t.Fatal("NewAdmin() returned nil")
	}

	// Basic field validation
	if admin.TenantID() != tenantID {
		t.Errorf("TenantID: expected %s, got %s", tenantID, admin.TenantID())
	}
	if admin.Email() != email {
		t.Errorf("Email: expected %s, got %s", email, admin.Email())
	}
	if admin.PasswordHash() != passwordHash {
		t.Errorf("PasswordHash: expected %s, got %s", passwordHash, admin.PasswordHash())
	}
	if admin.DisplayName() != displayName {
		t.Errorf("DisplayName: expected %s, got %s", displayName, admin.DisplayName())
	}
	if admin.Role() != role {
		t.Errorf("Role: expected %s, got %s", role, admin.Role())
	}

	// Default values
	if !admin.IsActive() {
		t.Error("IsActive should be true by default")
	}
	if admin.IsDeleted() {
		t.Error("IsDeleted should be false by default")
	}

	// ID should be generated
	if admin.AdminID() == "" {
		t.Error("AdminID should not be empty")
	}

	// Timestamps
	if admin.CreatedAt().IsZero() {
		t.Error("CreatedAt should not be zero")
	}
	if admin.UpdatedAt().IsZero() {
		t.Error("UpdatedAt should not be zero")
	}
}

// =====================================================
// NewAdmin Tests - Error Cases
// =====================================================

func TestNewAdmin_ErrorWhenTenantIDEmpty(t *testing.T) {
	now := time.Now()
	tenantID := common.TenantID("") // Empty
	email := "test@example.com"
	passwordHash := "$2a$10$hashedpassword"
	displayName := "Test Admin"
	role := RoleOwner

	admin, err := NewAdmin(now, tenantID, email, passwordHash, displayName, role)

	if err == nil {
		t.Fatal("NewAdmin() should return error when tenant_id is empty")
	}
	if admin != nil {
		t.Error("NewAdmin() should return nil when validation fails")
	}
}

func TestNewAdmin_ErrorWhenEmailEmpty(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	email := "" // Empty
	passwordHash := "$2a$10$hashedpassword"
	displayName := "Test Admin"
	role := RoleOwner

	admin, err := NewAdmin(now, tenantID, email, passwordHash, displayName, role)

	if err == nil {
		t.Fatal("NewAdmin() should return error when email is empty")
	}
	if admin != nil {
		t.Error("NewAdmin() should return nil when validation fails")
	}
}

func TestNewAdmin_ErrorWhenEmailTooLong(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	email := string(make([]byte, 256)) + "@example.com" // Too long
	passwordHash := "$2a$10$hashedpassword"
	displayName := "Test Admin"
	role := RoleOwner

	admin, err := NewAdmin(now, tenantID, email, passwordHash, displayName, role)

	if err == nil {
		t.Fatal("NewAdmin() should return error when email is too long")
	}
	if admin != nil {
		t.Error("NewAdmin() should return nil when validation fails")
	}
}

func TestNewAdmin_ErrorWhenPasswordHashEmpty(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	email := "test@example.com"
	passwordHash := "" // Empty
	displayName := "Test Admin"
	role := RoleOwner

	admin, err := NewAdmin(now, tenantID, email, passwordHash, displayName, role)

	if err == nil {
		t.Fatal("NewAdmin() should return error when password_hash is empty")
	}
	if admin != nil {
		t.Error("NewAdmin() should return nil when validation fails")
	}
}

func TestNewAdmin_ErrorWhenDisplayNameEmpty(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	email := "test@example.com"
	passwordHash := "$2a$10$hashedpassword"
	displayName := "" // Empty
	role := RoleOwner

	admin, err := NewAdmin(now, tenantID, email, passwordHash, displayName, role)

	if err == nil {
		t.Fatal("NewAdmin() should return error when display_name is empty")
	}
	if admin != nil {
		t.Error("NewAdmin() should return nil when validation fails")
	}
}

func TestNewAdmin_ErrorWhenDisplayNameTooLong(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	email := "test@example.com"
	passwordHash := "$2a$10$hashedpassword"
	displayName := string(make([]byte, 256)) // Too long
	role := RoleOwner

	admin, err := NewAdmin(now, tenantID, email, passwordHash, displayName, role)

	if err == nil {
		t.Fatal("NewAdmin() should return error when display_name is too long")
	}
	if admin != nil {
		t.Error("NewAdmin() should return nil when validation fails")
	}
}

func TestNewAdmin_ErrorWhenRoleInvalid(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	email := "test@example.com"
	passwordHash := "$2a$10$hashedpassword"
	displayName := "Test Admin"
	role := Role("invalid") // Invalid role

	admin, err := NewAdmin(now, tenantID, email, passwordHash, displayName, role)

	if err == nil {
		t.Fatal("NewAdmin() should return error when role is invalid")
	}
	if admin != nil {
		t.Error("NewAdmin() should return nil when validation fails")
	}
}

// =====================================================
// Admin.CanLogin Tests
// =====================================================

func TestAdmin_CanLogin_ActiveAndNotDeleted(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	admin, _ := NewAdmin(now, tenantID, "test@example.com", "$2a$10$hash", "Admin", RoleOwner)

	if !admin.CanLogin() {
		t.Error("CanLogin() should return true for active, non-deleted admin")
	}
}

func TestAdmin_CanLogin_InactiveAdmin(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	admin, _ := NewAdmin(now, tenantID, "test@example.com", "$2a$10$hash", "Admin", RoleOwner)

	admin.Deactivate(now)

	if admin.CanLogin() {
		t.Error("CanLogin() should return false for inactive admin")
	}
}

func TestAdmin_CanLogin_DeletedAdmin(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	admin, _ := NewAdmin(now, tenantID, "test@example.com", "$2a$10$hash", "Admin", RoleOwner)

	admin.Delete(now)

	if admin.CanLogin() {
		t.Error("CanLogin() should return false for deleted admin")
	}
}

// =====================================================
// Admin Update Methods Tests
// =====================================================

func TestAdmin_UpdateEmail_Success(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	admin, _ := NewAdmin(now, tenantID, "old@example.com", "$2a$10$hash", "Admin", RoleOwner)

	newEmail := "new@example.com"
	err := admin.UpdateEmail(now, newEmail)

	if err != nil {
		t.Fatalf("UpdateEmail() should succeed, but got error: %v", err)
	}
	if admin.Email() != newEmail {
		t.Errorf("Email: expected %s, got %s", newEmail, admin.Email())
	}
}

func TestAdmin_UpdateEmail_ErrorWhenEmpty(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	admin, _ := NewAdmin(now, tenantID, "old@example.com", "$2a$10$hash", "Admin", RoleOwner)

	err := admin.UpdateEmail(now, "")

	if err == nil {
		t.Fatal("UpdateEmail() should return error when email is empty")
	}
}

func TestAdmin_UpdateEmail_ErrorWhenTooLong(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	admin, _ := NewAdmin(now, tenantID, "old@example.com", "$2a$10$hash", "Admin", RoleOwner)

	err := admin.UpdateEmail(now, string(make([]byte, 256))+"@example.com")

	if err == nil {
		t.Fatal("UpdateEmail() should return error when email is too long")
	}
}

func TestAdmin_UpdatePasswordHash_Success(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	admin, _ := NewAdmin(now, tenantID, "test@example.com", "$2a$10$oldhash", "Admin", RoleOwner)

	newHash := "$2a$10$newhash"
	err := admin.UpdatePasswordHash(now, newHash)

	if err != nil {
		t.Fatalf("UpdatePasswordHash() should succeed, but got error: %v", err)
	}
	if admin.PasswordHash() != newHash {
		t.Errorf("PasswordHash: expected %s, got %s", newHash, admin.PasswordHash())
	}
}

func TestAdmin_UpdatePasswordHash_ErrorWhenEmpty(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	admin, _ := NewAdmin(now, tenantID, "test@example.com", "$2a$10$hash", "Admin", RoleOwner)

	err := admin.UpdatePasswordHash(now, "")

	if err == nil {
		t.Fatal("UpdatePasswordHash() should return error when password_hash is empty")
	}
}

func TestAdmin_UpdateDisplayName_Success(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	admin, _ := NewAdmin(now, tenantID, "test@example.com", "$2a$10$hash", "Old Name", RoleOwner)

	newName := "New Name"
	err := admin.UpdateDisplayName(now, newName)

	if err != nil {
		t.Fatalf("UpdateDisplayName() should succeed, but got error: %v", err)
	}
	if admin.DisplayName() != newName {
		t.Errorf("DisplayName: expected %s, got %s", newName, admin.DisplayName())
	}
}

func TestAdmin_UpdateDisplayName_ErrorWhenEmpty(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	admin, _ := NewAdmin(now, tenantID, "test@example.com", "$2a$10$hash", "Old Name", RoleOwner)

	err := admin.UpdateDisplayName(now, "")

	if err == nil {
		t.Fatal("UpdateDisplayName() should return error when display_name is empty")
	}
}

func TestAdmin_UpdateRole_Success(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	admin, _ := NewAdmin(now, tenantID, "test@example.com", "$2a$10$hash", "Admin", RoleOwner)

	newRole := RoleManager
	err := admin.UpdateRole(now, newRole)

	if err != nil {
		t.Fatalf("UpdateRole() should succeed, but got error: %v", err)
	}
	if admin.Role() != newRole {
		t.Errorf("Role: expected %s, got %s", newRole, admin.Role())
	}
}

func TestAdmin_UpdateRole_ErrorWhenInvalid(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	admin, _ := NewAdmin(now, tenantID, "test@example.com", "$2a$10$hash", "Admin", RoleOwner)

	err := admin.UpdateRole(now, Role("invalid"))

	if err == nil {
		t.Fatal("UpdateRole() should return error when role is invalid")
	}
}

// =====================================================
// Admin Activate/Deactivate/Delete Tests
// =====================================================

func TestAdmin_ActivateDeactivate(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	admin, _ := NewAdmin(now, tenantID, "test@example.com", "$2a$10$hash", "Admin", RoleOwner)

	// Initially active
	if !admin.IsActive() {
		t.Error("Admin should be active by default")
	}

	// Deactivate
	admin.Deactivate(now)
	if admin.IsActive() {
		t.Error("Admin should be inactive after Deactivate()")
	}

	// Activate
	admin.Activate(now)
	if !admin.IsActive() {
		t.Error("Admin should be active after Activate()")
	}
}

func TestAdmin_Delete(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	admin, _ := NewAdmin(now, tenantID, "test@example.com", "$2a$10$hash", "Admin", RoleOwner)

	// Initially not deleted
	if admin.IsDeleted() {
		t.Error("Admin should not be deleted by default")
	}
	if admin.DeletedAt() != nil {
		t.Error("DeletedAt should be nil by default")
	}

	// Delete
	admin.Delete(now)

	if !admin.IsDeleted() {
		t.Error("Admin should be deleted after Delete()")
	}
	if admin.DeletedAt() == nil {
		t.Error("DeletedAt should not be nil after Delete()")
	}
}

// =====================================================
// Role Tests
// =====================================================

func TestRole_Validate(t *testing.T) {
	tests := []struct {
		name    string
		role    Role
		wantErr bool
	}{
		{"Owner is valid", RoleOwner, false},
		{"Manager is valid", RoleManager, false},
		{"Invalid role returns error", Role("invalid"), true},
		{"Empty role returns error", Role(""), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.role.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// =====================================================
// ReconstructAdmin Tests
// =====================================================

func TestReconstructAdmin_Success(t *testing.T) {
	adminID := common.NewAdminID()
	tenantID := common.NewTenantID()
	now := time.Now()

	admin, err := ReconstructAdmin(
		adminID,
		tenantID,
		"test@example.com",
		"$2a$10$hash",
		"Admin",
		RoleOwner,
		true,
		now,
		now,
		nil,
	)

	if err != nil {
		t.Fatalf("ReconstructAdmin() should succeed, but got error: %v", err)
	}
	if admin == nil {
		t.Fatal("ReconstructAdmin() returned nil")
	}
	if admin.AdminID() != adminID {
		t.Errorf("AdminID: expected %s, got %s", adminID, admin.AdminID())
	}
}

func TestReconstructAdmin_WithDeletedAt(t *testing.T) {
	adminID := common.NewAdminID()
	tenantID := common.NewTenantID()
	now := time.Now()
	deletedAt := now.Add(-time.Hour)

	admin, err := ReconstructAdmin(
		adminID,
		tenantID,
		"test@example.com",
		"$2a$10$hash",
		"Admin",
		RoleOwner,
		false,
		now,
		now,
		&deletedAt,
	)

	if err != nil {
		t.Fatalf("ReconstructAdmin() should succeed, but got error: %v", err)
	}
	if !admin.IsDeleted() {
		t.Error("Admin should be marked as deleted")
	}
	// Deleted admin should not be able to login
	if admin.CanLogin() {
		t.Error("Deleted admin should not be able to login")
	}
}
