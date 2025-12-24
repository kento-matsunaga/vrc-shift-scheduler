package auth_test

import (
	"testing"
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/auth"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
)

// =====================================================
// Invitation Entity Tests
// =====================================================

func createTestAdmin(t *testing.T) *auth.Admin {
	t.Helper()
	now := time.Now()
	tenantID := common.NewTenantID()
	admin, err := auth.NewAdmin(now, tenantID, "inviter@example.com", "$2a$10$hash", "Inviter Admin", auth.RoleOwner)
	if err != nil {
		t.Fatalf("Failed to create test admin: %v", err)
	}
	return admin
}

func TestNewInvitation_Success(t *testing.T) {
	now := time.Now()
	admin := createTestAdmin(t)

	invitation, err := auth.NewInvitation(
		now,
		admin,
		"invited@example.com",
		auth.RoleManager,
		24*time.Hour, // 24 hours expiration
	)

	if err != nil {
		t.Fatalf("NewInvitation() should succeed, got error: %v", err)
	}

	if invitation.InvitationID() == "" {
		t.Error("InvitationID should be set")
	}

	if invitation.TenantID() != admin.TenantID() {
		t.Errorf("TenantID should match inviter's tenant: got %v, want %v", invitation.TenantID(), admin.TenantID())
	}

	if invitation.Email() != "invited@example.com" {
		t.Errorf("Email mismatch: got %v, want 'invited@example.com'", invitation.Email())
	}

	if invitation.Role() != auth.RoleManager {
		t.Errorf("Role mismatch: got %v, want %v", invitation.Role(), auth.RoleManager)
	}

	if invitation.Token() == "" {
		t.Error("Token should be generated")
	}

	if len(invitation.Token()) != 64 { // 32 bytes = 64 hex characters
		t.Errorf("Token should be 64 characters (32 bytes hex), got %d", len(invitation.Token()))
	}

	if invitation.CreatedByAdminID() != admin.AdminID() {
		t.Errorf("CreatedByAdminID mismatch: got %v, want %v", invitation.CreatedByAdminID(), admin.AdminID())
	}

	if invitation.AcceptedAt() != nil {
		t.Error("AcceptedAt should be nil initially")
	}

	if invitation.IsAccepted() {
		t.Error("IsAccepted() should return false initially")
	}
}

func TestNewInvitation_ErrorWhenInviterInactive(t *testing.T) {
	now := time.Now()
	admin := createTestAdmin(t)
	admin.Deactivate(now)

	_, err := auth.NewInvitation(
		now,
		admin,
		"invited@example.com",
		auth.RoleManager,
		24*time.Hour,
	)

	if err == nil {
		t.Fatal("NewInvitation() should fail when inviter is inactive")
	}
}

func TestNewInvitation_ErrorWhenInviterDeleted(t *testing.T) {
	now := time.Now()
	admin := createTestAdmin(t)
	admin.Delete(now)

	_, err := auth.NewInvitation(
		now,
		admin,
		"invited@example.com",
		auth.RoleManager,
		24*time.Hour,
	)

	if err == nil {
		t.Fatal("NewInvitation() should fail when inviter is deleted")
	}
}

func TestInvitation_IsExpired(t *testing.T) {
	now := time.Now()
	admin := createTestAdmin(t)

	invitation, err := auth.NewInvitation(
		now,
		admin,
		"invited@example.com",
		auth.RoleManager,
		1*time.Hour, // 1 hour expiration
	)
	if err != nil {
		t.Fatalf("NewInvitation() failed: %v", err)
	}

	// Before expiration
	if invitation.IsExpired(now.Add(30 * time.Minute)) {
		t.Error("Invitation should not be expired 30 minutes after creation")
	}

	// After expiration
	if !invitation.IsExpired(now.Add(2 * time.Hour)) {
		t.Error("Invitation should be expired 2 hours after creation")
	}
}

func TestInvitation_CanAccept_Success(t *testing.T) {
	now := time.Now()
	admin := createTestAdmin(t)

	invitation, _ := auth.NewInvitation(
		now,
		admin,
		"invited@example.com",
		auth.RoleManager,
		24*time.Hour,
	)

	err := invitation.CanAccept(now.Add(1 * time.Hour))
	if err != nil {
		t.Errorf("CanAccept() should succeed, got error: %v", err)
	}
}

func TestInvitation_CanAccept_ErrorWhenExpired(t *testing.T) {
	now := time.Now()
	admin := createTestAdmin(t)

	invitation, _ := auth.NewInvitation(
		now,
		admin,
		"invited@example.com",
		auth.RoleManager,
		1*time.Hour,
	)

	err := invitation.CanAccept(now.Add(2 * time.Hour))
	if err == nil {
		t.Error("CanAccept() should fail when invitation is expired")
	}
}

func TestInvitation_CanAccept_ErrorWhenAlreadyAccepted(t *testing.T) {
	now := time.Now()
	admin := createTestAdmin(t)

	invitation, _ := auth.NewInvitation(
		now,
		admin,
		"invited@example.com",
		auth.RoleManager,
		24*time.Hour,
	)

	// First accept
	err := invitation.Accept(now.Add(1 * time.Hour))
	if err != nil {
		t.Fatalf("First Accept() should succeed: %v", err)
	}

	// Try to check if can accept again
	err = invitation.CanAccept(now.Add(2 * time.Hour))
	if err == nil {
		t.Error("CanAccept() should fail when invitation is already accepted")
	}
}

func TestInvitation_Accept_Success(t *testing.T) {
	now := time.Now()
	admin := createTestAdmin(t)

	invitation, _ := auth.NewInvitation(
		now,
		admin,
		"invited@example.com",
		auth.RoleManager,
		24*time.Hour,
	)

	acceptTime := now.Add(1 * time.Hour)
	err := invitation.Accept(acceptTime)

	if err != nil {
		t.Fatalf("Accept() should succeed, got error: %v", err)
	}

	if !invitation.IsAccepted() {
		t.Error("IsAccepted() should return true after Accept()")
	}

	if invitation.AcceptedAt() == nil {
		t.Error("AcceptedAt() should not be nil after Accept()")
	}
}

func TestInvitation_Accept_ErrorWhenExpired(t *testing.T) {
	now := time.Now()
	admin := createTestAdmin(t)

	invitation, _ := auth.NewInvitation(
		now,
		admin,
		"invited@example.com",
		auth.RoleManager,
		1*time.Hour,
	)

	err := invitation.Accept(now.Add(2 * time.Hour))
	if err == nil {
		t.Error("Accept() should fail when invitation is expired")
	}
}

func TestInvitation_Accept_ErrorWhenAlreadyAccepted(t *testing.T) {
	now := time.Now()
	admin := createTestAdmin(t)

	invitation, _ := auth.NewInvitation(
		now,
		admin,
		"invited@example.com",
		auth.RoleManager,
		24*time.Hour,
	)

	// First accept
	_ = invitation.Accept(now.Add(1 * time.Hour))

	// Second accept attempt
	err := invitation.Accept(now.Add(2 * time.Hour))
	if err == nil {
		t.Error("Accept() should fail when invitation is already accepted")
	}
}

func TestReconstructInvitation_Success(t *testing.T) {
	tenantID := common.NewTenantID()
	adminID := common.NewAdminID()
	now := time.Now()
	expiresAt := now.Add(24 * time.Hour)

	invitation, err := auth.ReconstructInvitation(
		auth.InvitationID("01INVITATIONIDTEST"),
		tenantID,
		"test@example.com",
		auth.RoleManager,
		"abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
		adminID,
		expiresAt,
		nil,
		now,
	)

	if err != nil {
		t.Fatalf("ReconstructInvitation() should succeed, got error: %v", err)
	}

	if invitation.Email() != "test@example.com" {
		t.Errorf("Email mismatch: got %v, want 'test@example.com'", invitation.Email())
	}

	if invitation.Role() != auth.RoleManager {
		t.Errorf("Role mismatch: got %v, want %v", invitation.Role(), auth.RoleManager)
	}
}

func TestReconstructInvitation_ErrorWhenEmailEmpty(t *testing.T) {
	tenantID := common.NewTenantID()
	adminID := common.NewAdminID()
	now := time.Now()
	expiresAt := now.Add(24 * time.Hour)

	_, err := auth.ReconstructInvitation(
		auth.InvitationID("01INVITATIONIDTEST"),
		tenantID,
		"", // Empty email
		auth.RoleManager,
		"abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
		adminID,
		expiresAt,
		nil,
		now,
	)

	if err == nil {
		t.Error("ReconstructInvitation() should fail when email is empty")
	}
}
