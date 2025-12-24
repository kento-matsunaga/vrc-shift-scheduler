package shift_test

import (
	"strings"
	"testing"
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/shift"
)

// =====================================================
// PositionID Tests
// =====================================================

func TestNewPositionID(t *testing.T) {
	id := shift.NewPositionID()

	if id.String() == "" {
		t.Error("NewPositionID should generate a non-empty ID")
	}

	if err := id.Validate(); err != nil {
		t.Errorf("NewPositionID should generate a valid ID: %v", err)
	}
}

func TestNewPositionIDWithTime(t *testing.T) {
	now := time.Now()
	id := shift.NewPositionIDWithTime(now)

	if id.String() == "" {
		t.Error("NewPositionIDWithTime should generate a non-empty ID")
	}

	if err := id.Validate(); err != nil {
		t.Errorf("NewPositionIDWithTime should generate a valid ID: %v", err)
	}
}

func TestPositionID_Validate_Error(t *testing.T) {
	testCases := []struct {
		name string
		id   shift.PositionID
	}{
		{"empty", shift.PositionID("")},
		{"invalid_format", shift.PositionID("invalid")},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.id.Validate()
			if err == nil {
				t.Errorf("Validate() should fail for %s", tc.name)
			}
		})
	}
}

// =====================================================
// NewPosition Tests
// =====================================================

func TestNewPosition_Success(t *testing.T) {
	tenantID := common.NewTenantID()

	position, err := shift.NewPosition(
		tenantID,
		"Staff",
		"General staff position",
		1,
	)

	if err != nil {
		t.Fatalf("NewPosition() should succeed: %v", err)
	}

	if position.TenantID() != tenantID {
		t.Errorf("TenantID mismatch: got %v, want %v", position.TenantID(), tenantID)
	}

	if position.PositionName() != "Staff" {
		t.Errorf("PositionName mismatch: got %v, want 'Staff'", position.PositionName())
	}

	if position.Description() != "General staff position" {
		t.Errorf("Description mismatch: got %v, want 'General staff position'", position.Description())
	}

	if position.DisplayOrder() != 1 {
		t.Errorf("DisplayOrder mismatch: got %v, want 1", position.DisplayOrder())
	}

	if !position.IsActive() {
		t.Error("IsActive should be true for new position")
	}

	if position.PositionID().String() == "" {
		t.Error("PositionID should be generated")
	}

	if position.CreatedAt().IsZero() {
		t.Error("CreatedAt should be set")
	}

	if position.UpdatedAt().IsZero() {
		t.Error("UpdatedAt should be set")
	}

	if position.DeletedAt() != nil {
		t.Error("DeletedAt should be nil for new position")
	}

	if position.IsDeleted() {
		t.Error("IsDeleted should be false for new position")
	}
}

func TestNewPosition_Success_EmptyDescription(t *testing.T) {
	tenantID := common.NewTenantID()

	position, err := shift.NewPosition(
		tenantID,
		"Staff",
		"", // Empty description is allowed
		0,
	)

	if err != nil {
		t.Fatalf("NewPosition() should succeed with empty description: %v", err)
	}

	if position.Description() != "" {
		t.Error("Description should be empty")
	}
}

func TestNewPosition_ErrorWhenInvalidTenantID(t *testing.T) {
	_, err := shift.NewPosition(
		common.TenantID(""), // Invalid
		"Staff",
		"",
		0,
	)

	if err == nil {
		t.Error("NewPosition() should fail when tenant_id is invalid")
	}
}

func TestNewPosition_ErrorWhenEmptyPositionName(t *testing.T) {
	tenantID := common.NewTenantID()

	_, err := shift.NewPosition(
		tenantID,
		"", // Empty position name
		"",
		0,
	)

	if err == nil {
		t.Error("NewPosition() should fail when position_name is empty")
	}
}

func TestNewPosition_ErrorWhenPositionNameTooLong(t *testing.T) {
	tenantID := common.NewTenantID()
	longName := strings.Repeat("a", 256) // 256 characters (max is 255)

	_, err := shift.NewPosition(
		tenantID,
		longName,
		"",
		0,
	)

	if err == nil {
		t.Error("NewPosition() should fail when position_name is too long")
	}
}

// =====================================================
// ReconstructPosition Tests
// =====================================================

func TestReconstructPosition_Success(t *testing.T) {
	positionID := shift.NewPositionID()
	tenantID := common.NewTenantID()
	createdAt := time.Now().Add(-time.Hour)
	updatedAt := time.Now()

	position, err := shift.ReconstructPosition(
		positionID,
		tenantID,
		"Security",
		"Security position",
		2,
		true,
		createdAt,
		updatedAt,
		nil,
	)

	if err != nil {
		t.Fatalf("ReconstructPosition() should succeed: %v", err)
	}

	if position.PositionID() != positionID {
		t.Errorf("PositionID mismatch: got %v, want %v", position.PositionID(), positionID)
	}

	if position.PositionName() != "Security" {
		t.Errorf("PositionName mismatch: got %v, want 'Security'", position.PositionName())
	}
}

func TestReconstructPosition_Success_Inactive(t *testing.T) {
	positionID := shift.NewPositionID()
	tenantID := common.NewTenantID()
	createdAt := time.Now()
	updatedAt := time.Now()

	position, err := shift.ReconstructPosition(
		positionID,
		tenantID,
		"Inactive Position",
		"",
		0,
		false, // Inactive
		createdAt,
		updatedAt,
		nil,
	)

	if err != nil {
		t.Fatalf("ReconstructPosition() should succeed: %v", err)
	}

	if position.IsActive() {
		t.Error("IsActive should be false")
	}
}

func TestReconstructPosition_Success_Deleted(t *testing.T) {
	positionID := shift.NewPositionID()
	tenantID := common.NewTenantID()
	createdAt := time.Now()
	updatedAt := time.Now()
	deletedAt := time.Now()

	position, err := shift.ReconstructPosition(
		positionID,
		tenantID,
		"Deleted Position",
		"",
		0,
		false,
		createdAt,
		updatedAt,
		&deletedAt,
	)

	if err != nil {
		t.Fatalf("ReconstructPosition() should succeed: %v", err)
	}

	if !position.IsDeleted() {
		t.Error("IsDeleted should be true")
	}

	if position.DeletedAt() == nil {
		t.Error("DeletedAt should be set")
	}
}

func TestReconstructPosition_ErrorWhenInvalidTenantID(t *testing.T) {
	positionID := shift.NewPositionID()
	createdAt := time.Now()
	updatedAt := time.Now()

	_, err := shift.ReconstructPosition(
		positionID,
		common.TenantID(""), // Invalid
		"Staff",
		"",
		0,
		true,
		createdAt,
		updatedAt,
		nil,
	)

	if err == nil {
		t.Error("ReconstructPosition() should fail when tenant_id is invalid")
	}
}

// =====================================================
// Position Methods Tests
// =====================================================

func TestPosition_UpdatePositionName_Success(t *testing.T) {
	tenantID := common.NewTenantID()

	position, _ := shift.NewPosition(tenantID, "Original", "", 0)
	originalUpdatedAt := position.UpdatedAt()

	time.Sleep(time.Millisecond)

	err := position.UpdatePositionName("Updated Name")
	if err != nil {
		t.Fatalf("UpdatePositionName() should succeed: %v", err)
	}

	if position.PositionName() != "Updated Name" {
		t.Errorf("PositionName should be updated: got %v, want 'Updated Name'", position.PositionName())
	}

	if !position.UpdatedAt().After(originalUpdatedAt) {
		t.Error("UpdatedAt should be updated")
	}
}

func TestPosition_UpdatePositionName_ErrorWhenEmpty(t *testing.T) {
	tenantID := common.NewTenantID()
	position, _ := shift.NewPosition(tenantID, "Original", "", 0)

	err := position.UpdatePositionName("")
	if err == nil {
		t.Error("UpdatePositionName() should fail when name is empty")
	}
}

func TestPosition_UpdatePositionName_ErrorWhenTooLong(t *testing.T) {
	tenantID := common.NewTenantID()
	position, _ := shift.NewPosition(tenantID, "Original", "", 0)

	longName := strings.Repeat("a", 256)
	err := position.UpdatePositionName(longName)
	if err == nil {
		t.Error("UpdatePositionName() should fail when name is too long")
	}
}

func TestPosition_UpdateDescription(t *testing.T) {
	tenantID := common.NewTenantID()
	position, _ := shift.NewPosition(tenantID, "Staff", "Original", 0)
	originalUpdatedAt := position.UpdatedAt()

	time.Sleep(time.Millisecond)

	position.UpdateDescription("New Description")

	if position.Description() != "New Description" {
		t.Errorf("Description should be updated: got %v, want 'New Description'", position.Description())
	}

	if !position.UpdatedAt().After(originalUpdatedAt) {
		t.Error("UpdatedAt should be updated")
	}
}

func TestPosition_UpdateDisplayOrder(t *testing.T) {
	tenantID := common.NewTenantID()
	position, _ := shift.NewPosition(tenantID, "Staff", "", 1)
	originalUpdatedAt := position.UpdatedAt()

	time.Sleep(time.Millisecond)

	position.UpdateDisplayOrder(5)

	if position.DisplayOrder() != 5 {
		t.Errorf("DisplayOrder should be updated: got %v, want 5", position.DisplayOrder())
	}

	if !position.UpdatedAt().After(originalUpdatedAt) {
		t.Error("UpdatedAt should be updated")
	}
}

func TestPosition_Activate(t *testing.T) {
	tenantID := common.NewTenantID()
	position, _ := shift.NewPosition(tenantID, "Staff", "", 0)

	// Deactivate first
	position.Deactivate()

	if position.IsActive() {
		t.Error("IsActive should be false after Deactivate()")
	}

	// Now activate
	position.Activate()

	if !position.IsActive() {
		t.Error("IsActive should be true after Activate()")
	}
}

func TestPosition_Deactivate(t *testing.T) {
	tenantID := common.NewTenantID()
	position, _ := shift.NewPosition(tenantID, "Staff", "", 0)

	if !position.IsActive() {
		t.Error("IsActive should be true for new position")
	}

	position.Deactivate()

	if position.IsActive() {
		t.Error("IsActive should be false after Deactivate()")
	}
}

func TestPosition_Delete(t *testing.T) {
	tenantID := common.NewTenantID()
	position, _ := shift.NewPosition(tenantID, "Staff", "", 0)

	if position.IsDeleted() {
		t.Error("IsDeleted should be false for new position")
	}

	position.Delete()

	if !position.IsDeleted() {
		t.Error("IsDeleted should be true after Delete()")
	}

	if position.DeletedAt() == nil {
		t.Error("DeletedAt should be set after Delete()")
	}
}
