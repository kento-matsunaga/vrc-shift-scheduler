package role_test

import (
	"testing"
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/role"
)

// =====================================================
// Role Entity Tests
// =====================================================

func TestNewRole_Success(t *testing.T) {
	tenantID := common.NewTenantID()

	r, err := role.NewRole(
		time.Now(),
		tenantID,
		"スタッフ",
		"イベントスタッフロール",
		"#FF5733",
		1,
	)

	if err != nil {
		t.Fatalf("NewRole() should succeed, got error: %v", err)
	}

	if r.RoleID() == "" {
		t.Error("RoleID should be set")
	}

	if r.TenantID() != tenantID {
		t.Errorf("TenantID mismatch: got %v, want %v", r.TenantID(), tenantID)
	}

	if r.Name() != "スタッフ" {
		t.Errorf("Name mismatch: got %v, want 'スタッフ'", r.Name())
	}

	if r.Description() != "イベントスタッフロール" {
		t.Errorf("Description mismatch: got %v, want 'イベントスタッフロール'", r.Description())
	}

	if r.Color() != "#FF5733" {
		t.Errorf("Color mismatch: got %v, want '#FF5733'", r.Color())
	}

	if r.DisplayOrder() != 1 {
		t.Errorf("DisplayOrder mismatch: got %v, want 1", r.DisplayOrder())
	}

	if r.DeletedAt() != nil {
		t.Error("New role should not be deleted")
	}
}

func TestNewRole_ErrorWhenNameEmpty(t *testing.T) {
	tenantID := common.NewTenantID()

	_, err := role.NewRole(
		time.Now(),
		tenantID,
		"", // Empty name
		"Description",
		"#FF5733",
		1,
	)

	if err == nil {
		t.Fatal("NewRole() should fail when name is empty")
	}
}

func TestNewRole_ErrorWhenNameTooLong(t *testing.T) {
	tenantID := common.NewTenantID()
	longName := make([]byte, 101)
	for i := range longName {
		longName[i] = 'a'
	}

	_, err := role.NewRole(
		time.Now(),
		tenantID,
		string(longName),
		"Description",
		"#FF5733",
		1,
	)

	if err == nil {
		t.Fatal("NewRole() should fail when name is too long")
	}
}

func TestNewRole_ErrorWhenDescriptionTooLong(t *testing.T) {
	tenantID := common.NewTenantID()
	longDesc := make([]byte, 501)
	for i := range longDesc {
		longDesc[i] = 'a'
	}

	_, err := role.NewRole(
		time.Now(),
		tenantID,
		"Valid Name",
		string(longDesc),
		"#FF5733",
		1,
	)

	if err == nil {
		t.Fatal("NewRole() should fail when description is too long")
	}
}

func TestNewRole_ErrorWhenColorTooLong(t *testing.T) {
	tenantID := common.NewTenantID()
	longColor := make([]byte, 21)
	for i := range longColor {
		longColor[i] = 'a'
	}

	_, err := role.NewRole(
		time.Now(),
		tenantID,
		"Valid Name",
		"Description",
		string(longColor),
		1,
	)

	if err == nil {
		t.Fatal("NewRole() should fail when color is too long")
	}
}

func TestNewRole_SuccessWithEmptyOptionalFields(t *testing.T) {
	tenantID := common.NewTenantID()

	r, err := role.NewRole(
		time.Now(),
		tenantID,
		"Minimal Role",
		"", // Empty description (optional)
		"", // Empty color (optional)
		0,  // Zero display order
	)

	if err != nil {
		t.Fatalf("NewRole() should succeed with empty optional fields, got error: %v", err)
	}

	if r.Description() != "" {
		t.Errorf("Description should be empty: got %v", r.Description())
	}

	if r.Color() != "" {
		t.Errorf("Color should be empty: got %v", r.Color())
	}
}

func TestRole_UpdateDetails_Success(t *testing.T) {
	tenantID := common.NewTenantID()
	r, _ := role.NewRole(time.Now(), tenantID, "Original", "Original Desc", "#000000", 1)

	err := r.UpdateDetails(time.Now(),"Updated", "Updated Desc", "#FFFFFF", 2)

	if err != nil {
		t.Fatalf("UpdateDetails() should succeed, got error: %v", err)
	}

	if r.Name() != "Updated" {
		t.Errorf("Name should be updated: got %v, want 'Updated'", r.Name())
	}

	if r.Description() != "Updated Desc" {
		t.Errorf("Description should be updated: got %v, want 'Updated Desc'", r.Description())
	}

	if r.Color() != "#FFFFFF" {
		t.Errorf("Color should be updated: got %v, want '#FFFFFF'", r.Color())
	}

	if r.DisplayOrder() != 2 {
		t.Errorf("DisplayOrder should be updated: got %v, want 2", r.DisplayOrder())
	}
}

func TestRole_UpdateDetails_ErrorWhenNameEmpty(t *testing.T) {
	tenantID := common.NewTenantID()
	r, _ := role.NewRole(time.Now(), tenantID, "Original", "Original Desc", "#000000", 1)

	err := r.UpdateDetails(time.Now(),"", "Updated Desc", "#FFFFFF", 2)

	if err == nil {
		t.Fatal("UpdateDetails() should fail when name is empty")
	}
}

func TestRole_Delete(t *testing.T) {
	tenantID := common.NewTenantID()
	r, _ := role.NewRole(time.Now(), tenantID, "Test Role", "Description", "#FF5733", 1)

	if r.DeletedAt() != nil {
		t.Error("New role should not be deleted")
	}

	r.Delete(time.Now())

	if r.DeletedAt() == nil {
		t.Error("DeletedAt should be set after Delete()")
	}
}

func TestReconstructRole_Success(t *testing.T) {
	tenantID := common.NewTenantID()
	roleID := common.NewRoleID()
	now := time.Now()

	r, err := role.ReconstructRole(
		roleID,
		tenantID,
		"Reconstructed Role",
		"Description",
		"#FF5733",
		1,
		now,
		now,
		nil,
	)

	if err != nil {
		t.Fatalf("ReconstructRole() should succeed, got error: %v", err)
	}

	if r.RoleID() != roleID {
		t.Errorf("RoleID mismatch: got %v, want %v", r.RoleID(), roleID)
	}

	if r.Name() != "Reconstructed Role" {
		t.Errorf("Name mismatch: got %v, want 'Reconstructed Role'", r.Name())
	}
}

func TestReconstructRole_WithDeletedAt(t *testing.T) {
	tenantID := common.NewTenantID()
	roleID := common.NewRoleID()
	now := time.Now()
	deletedAt := now.Add(-1 * time.Hour)

	r, err := role.ReconstructRole(
		roleID,
		tenantID,
		"Deleted Role",
		"Description",
		"#FF5733",
		1,
		now,
		now,
		&deletedAt,
	)

	if err != nil {
		t.Fatalf("ReconstructRole() should succeed, got error: %v", err)
	}

	if r.DeletedAt() == nil {
		t.Error("DeletedAt should be set")
	}
}
