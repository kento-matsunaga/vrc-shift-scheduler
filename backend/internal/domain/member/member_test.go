package member

import (
	"testing"
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
)

// =====================================================
// NewMember Tests - Success Cases
// =====================================================

func TestNewMember_Success(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	displayName := "Test Member"
	discordUserID := "123456789012345678"
	email := "test@example.com"

	member, err := NewMember(now, tenantID, displayName, discordUserID, email)

	if err != nil {
		t.Fatalf("NewMember() should succeed, but got error: %v", err)
	}

	if member == nil {
		t.Fatal("NewMember() returned nil")
	}

	// Basic field validation
	if member.TenantID() != tenantID {
		t.Errorf("TenantID: expected %s, got %s", tenantID, member.TenantID())
	}
	if member.DisplayName() != displayName {
		t.Errorf("DisplayName: expected %s, got %s", displayName, member.DisplayName())
	}
	if member.DiscordUserID() != discordUserID {
		t.Errorf("DiscordUserID: expected %s, got %s", discordUserID, member.DiscordUserID())
	}
	if member.Email() != email {
		t.Errorf("Email: expected %s, got %s", email, member.Email())
	}

	// Default values
	if !member.IsActive() {
		t.Error("IsActive should be true by default")
	}
	if member.IsDeleted() {
		t.Error("IsDeleted should be false by default")
	}

	// ID should be generated
	if member.MemberID() == "" {
		t.Error("MemberID should not be empty")
	}

	// Timestamps
	if member.CreatedAt().IsZero() {
		t.Error("CreatedAt should not be zero")
	}
	if member.UpdatedAt().IsZero() {
		t.Error("UpdatedAt should not be zero")
	}
	if member.DeletedAt() != nil {
		t.Error("DeletedAt should be nil by default")
	}
}

func TestNewMember_SuccessWithOptionalFieldsEmpty(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	displayName := "Test Member"
	discordUserID := "" // Optional, can be empty
	email := ""         // Optional, can be empty

	member, err := NewMember(now, tenantID, displayName, discordUserID, email)

	if err != nil {
		t.Fatalf("NewMember() should succeed with empty optional fields, but got error: %v", err)
	}

	if member.DiscordUserID() != "" {
		t.Error("DiscordUserID should be empty")
	}
	if member.Email() != "" {
		t.Error("Email should be empty")
	}
}

// =====================================================
// NewMember Tests - Error Cases
// =====================================================

func TestNewMember_ErrorWhenTenantIDEmpty(t *testing.T) {
	now := time.Now()
	tenantID := common.TenantID("") // Empty
	displayName := "Test Member"

	member, err := NewMember(now, tenantID, displayName, "", "")

	if err == nil {
		t.Fatal("NewMember() should return error when tenant_id is empty")
	}
	if member != nil {
		t.Error("NewMember() should return nil when validation fails")
	}
}

func TestNewMember_ErrorWhenDisplayNameEmpty(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	displayName := "" // Empty

	member, err := NewMember(now, tenantID, displayName, "", "")

	if err == nil {
		t.Fatal("NewMember() should return error when display_name is empty")
	}
	if member != nil {
		t.Error("NewMember() should return nil when validation fails")
	}
}

func TestNewMember_ErrorWhenDisplayNameTooLong(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	displayName := string(make([]byte, 256)) // Too long (>255)

	member, err := NewMember(now, tenantID, displayName, "", "")

	if err == nil {
		t.Fatal("NewMember() should return error when display_name is too long")
	}
	if member != nil {
		t.Error("NewMember() should return nil when validation fails")
	}
}

func TestNewMember_ErrorWhenEmailTooLong(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	displayName := "Test Member"
	email := string(make([]byte, 256)) + "@example.com" // Too long (>255)

	member, err := NewMember(now, tenantID, displayName, "", email)

	if err == nil {
		t.Fatal("NewMember() should return error when email is too long")
	}
	if member != nil {
		t.Error("NewMember() should return nil when validation fails")
	}
}

func TestNewMember_ErrorWhenDiscordUserIDTooLong(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	displayName := "Test Member"
	discordUserID := string(make([]byte, 101)) // Too long (>100)

	member, err := NewMember(now, tenantID, displayName, discordUserID, "")

	if err == nil {
		t.Fatal("NewMember() should return error when discord_user_id is too long")
	}
	if member != nil {
		t.Error("NewMember() should return nil when validation fails")
	}
}

// =====================================================
// UpdateDisplayName Tests
// =====================================================

func TestMember_UpdateDisplayName_Success(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	member, _ := NewMember(now, tenantID, "Old Name", "", "")

	newName := "New Name"
	err := member.UpdateDisplayName(newName)

	if err != nil {
		t.Fatalf("UpdateDisplayName() should succeed, but got error: %v", err)
	}
	if member.DisplayName() != newName {
		t.Errorf("DisplayName: expected %s, got %s", newName, member.DisplayName())
	}
}

func TestMember_UpdateDisplayName_ErrorWhenEmpty(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	member, _ := NewMember(now, tenantID, "Old Name", "", "")

	err := member.UpdateDisplayName("")

	if err == nil {
		t.Fatal("UpdateDisplayName() should return error when display_name is empty")
	}
}

func TestMember_UpdateDisplayName_ErrorWhenTooLong(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	member, _ := NewMember(now, tenantID, "Old Name", "", "")

	err := member.UpdateDisplayName(string(make([]byte, 256)))

	if err == nil {
		t.Fatal("UpdateDisplayName() should return error when display_name is too long")
	}
}

// =====================================================
// UpdateDiscordUserID Tests
// =====================================================

func TestMember_UpdateDiscordUserID_Success(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	member, _ := NewMember(now, tenantID, "Member", "", "")

	newID := "987654321098765432"
	err := member.UpdateDiscordUserID(newID)

	if err != nil {
		t.Fatalf("UpdateDiscordUserID() should succeed, but got error: %v", err)
	}
	if member.DiscordUserID() != newID {
		t.Errorf("DiscordUserID: expected %s, got %s", newID, member.DiscordUserID())
	}
}

func TestMember_UpdateDiscordUserID_SuccessWhenEmpty(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	member, _ := NewMember(now, tenantID, "Member", "123456789012345678", "")

	// Empty is allowed (optional field)
	err := member.UpdateDiscordUserID("")

	if err != nil {
		t.Fatalf("UpdateDiscordUserID() should succeed with empty value, but got error: %v", err)
	}
	if member.DiscordUserID() != "" {
		t.Error("DiscordUserID should be empty")
	}
}

func TestMember_UpdateDiscordUserID_ErrorWhenTooLong(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	member, _ := NewMember(now, tenantID, "Member", "", "")

	err := member.UpdateDiscordUserID(string(make([]byte, 101)))

	if err == nil {
		t.Fatal("UpdateDiscordUserID() should return error when discord_user_id is too long")
	}
}

// =====================================================
// UpdateEmail Tests
// =====================================================

func TestMember_UpdateEmail_Success(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	member, _ := NewMember(now, tenantID, "Member", "", "old@example.com")

	newEmail := "new@example.com"
	err := member.UpdateEmail(newEmail)

	if err != nil {
		t.Fatalf("UpdateEmail() should succeed, but got error: %v", err)
	}
	if member.Email() != newEmail {
		t.Errorf("Email: expected %s, got %s", newEmail, member.Email())
	}
}

func TestMember_UpdateEmail_SuccessWhenEmpty(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	member, _ := NewMember(now, tenantID, "Member", "", "old@example.com")

	// Empty is allowed (optional field)
	err := member.UpdateEmail("")

	if err != nil {
		t.Fatalf("UpdateEmail() should succeed with empty value, but got error: %v", err)
	}
	if member.Email() != "" {
		t.Error("Email should be empty")
	}
}

func TestMember_UpdateEmail_ErrorWhenTooLong(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	member, _ := NewMember(now, tenantID, "Member", "", "")

	err := member.UpdateEmail(string(make([]byte, 256)) + "@example.com")

	if err == nil {
		t.Fatal("UpdateEmail() should return error when email is too long")
	}
}

// =====================================================
// UpdateDetails Tests
// =====================================================

func TestMember_UpdateDetails_Success(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	member, _ := NewMember(now, tenantID, "Old Name", "", "")

	newDisplayName := "New Name"
	newDiscordUserID := "123456789012345678"
	newEmail := "new@example.com"
	newIsActive := false

	err := member.UpdateDetails(newDisplayName, newDiscordUserID, newEmail, newIsActive)

	if err != nil {
		t.Fatalf("UpdateDetails() should succeed, but got error: %v", err)
	}
	if member.DisplayName() != newDisplayName {
		t.Errorf("DisplayName: expected %s, got %s", newDisplayName, member.DisplayName())
	}
	if member.DiscordUserID() != newDiscordUserID {
		t.Errorf("DiscordUserID: expected %s, got %s", newDiscordUserID, member.DiscordUserID())
	}
	if member.Email() != newEmail {
		t.Errorf("Email: expected %s, got %s", newEmail, member.Email())
	}
	if member.IsActive() != newIsActive {
		t.Errorf("IsActive: expected %t, got %t", newIsActive, member.IsActive())
	}
}

func TestMember_UpdateDetails_ErrorWhenDisplayNameEmpty(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	member, _ := NewMember(now, tenantID, "Old Name", "", "")

	err := member.UpdateDetails("", "", "", true) // Empty display name

	if err == nil {
		t.Fatal("UpdateDetails() should return error when display_name is empty")
	}
}

func TestMember_UpdateDetails_ErrorWhenDiscordUserIDTooLong(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	member, _ := NewMember(now, tenantID, "Name", "", "")

	err := member.UpdateDetails("Name", string(make([]byte, 101)), "", true)

	if err == nil {
		t.Fatal("UpdateDetails() should return error when discord_user_id is too long")
	}
}

// =====================================================
// Activate/Deactivate Tests
// =====================================================

func TestMember_ActivateDeactivate(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	member, _ := NewMember(now, tenantID, "Member", "", "")

	// Initially active
	if !member.IsActive() {
		t.Error("Member should be active by default")
	}

	// Deactivate
	member.Deactivate()
	if member.IsActive() {
		t.Error("Member should be inactive after Deactivate()")
	}

	// Activate
	member.Activate()
	if !member.IsActive() {
		t.Error("Member should be active after Activate()")
	}
}

// =====================================================
// Delete Tests
// =====================================================

func TestMember_Delete(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	member, _ := NewMember(now, tenantID, "Member", "", "")

	// Initially not deleted
	if member.IsDeleted() {
		t.Error("Member should not be deleted by default")
	}
	if member.DeletedAt() != nil {
		t.Error("DeletedAt should be nil by default")
	}

	// Delete
	member.Delete()

	if !member.IsDeleted() {
		t.Error("Member should be deleted after Delete()")
	}
	if member.DeletedAt() == nil {
		t.Error("DeletedAt should not be nil after Delete()")
	}
}

// =====================================================
// ReconstructMember Tests
// =====================================================

func TestReconstructMember_Success(t *testing.T) {
	memberID := common.NewMemberID()
	tenantID := common.NewTenantID()
	now := time.Now()
	displayName := "Test Member"
	discordUserID := "123456789012345678"
	email := "test@example.com"

	member, err := ReconstructMember(
		memberID,
		tenantID,
		displayName,
		discordUserID,
		email,
		true,
		now,
		now,
		nil,
	)

	if err != nil {
		t.Fatalf("ReconstructMember() should succeed, but got error: %v", err)
	}
	if member == nil {
		t.Fatal("ReconstructMember() returned nil")
	}
	if member.MemberID() != memberID {
		t.Errorf("MemberID: expected %s, got %s", memberID, member.MemberID())
	}
	if member.TenantID() != tenantID {
		t.Errorf("TenantID: expected %s, got %s", tenantID, member.TenantID())
	}
	if member.DisplayName() != displayName {
		t.Errorf("DisplayName: expected %s, got %s", displayName, member.DisplayName())
	}
}

func TestReconstructMember_WithDeletedAt(t *testing.T) {
	memberID := common.NewMemberID()
	tenantID := common.NewTenantID()
	now := time.Now()
	deletedAt := now.Add(-time.Hour)

	member, err := ReconstructMember(
		memberID,
		tenantID,
		"Deleted Member",
		"",
		"",
		false,
		now,
		now,
		&deletedAt,
	)

	if err != nil {
		t.Fatalf("ReconstructMember() should succeed, but got error: %v", err)
	}
	if !member.IsDeleted() {
		t.Error("Member should be marked as deleted")
	}
	if member.DeletedAt() == nil {
		t.Error("DeletedAt should not be nil")
	}
}

func TestReconstructMember_ErrorWhenValidationFails(t *testing.T) {
	memberID := common.NewMemberID()
	tenantID := common.TenantID("") // Invalid - empty
	now := time.Now()

	member, err := ReconstructMember(
		memberID,
		tenantID,
		"Member",
		"",
		"",
		true,
		now,
		now,
		nil,
	)

	if err == nil {
		t.Fatal("ReconstructMember() should return error when validation fails")
	}
	if member != nil {
		t.Error("ReconstructMember() should return nil when validation fails")
	}
}

// =====================================================
// UpdatedAt Timestamp Tests
// =====================================================

func TestMember_UpdateMethodsUpdateTimestamp(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	member, _ := NewMember(now, tenantID, "Member", "", "")

	originalUpdatedAt := member.UpdatedAt()

	// Wait a tiny bit to ensure time difference
	time.Sleep(1 * time.Millisecond)

	// UpdateDisplayName should update timestamp
	_ = member.UpdateDisplayName("New Name")

	if !member.UpdatedAt().After(originalUpdatedAt) {
		t.Error("UpdatedAt should be updated after UpdateDisplayName()")
	}
}

func TestMember_DeactivateUpdatesTimestamp(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	member, _ := NewMember(now, tenantID, "Member", "", "")

	originalUpdatedAt := member.UpdatedAt()

	// Wait a tiny bit to ensure time difference
	time.Sleep(1 * time.Millisecond)

	member.Deactivate()

	if !member.UpdatedAt().After(originalUpdatedAt) {
		t.Error("UpdatedAt should be updated after Deactivate()")
	}
}
