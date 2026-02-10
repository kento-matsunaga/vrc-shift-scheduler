package shift_test

import (
	"testing"
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/shift"
)

// =====================================================
// AssignmentStatus Tests
// =====================================================

func TestAssignmentStatus_Validate_Success(t *testing.T) {
	testCases := []struct {
		name   string
		status shift.AssignmentStatus
	}{
		{"confirmed", shift.AssignmentStatusConfirmed},
		{"cancelled", shift.AssignmentStatusCancelled},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.status.Validate()
			if err != nil {
				t.Errorf("Validate() should succeed for %s, got error: %v", tc.name, err)
			}
		})
	}
}

func TestAssignmentStatus_Validate_Error(t *testing.T) {
	testCases := []struct {
		name   string
		status shift.AssignmentStatus
	}{
		{"empty", shift.AssignmentStatus("")},
		{"invalid", shift.AssignmentStatus("invalid")},
		{"CONFIRMED", shift.AssignmentStatus("CONFIRMED")},
		{"pending", shift.AssignmentStatus("pending")},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.status.Validate()
			if err == nil {
				t.Errorf("Validate() should fail for %s", tc.name)
			}
		})
	}
}

// =====================================================
// AssignmentMethod Tests
// =====================================================

func TestAssignmentMethod_Validate_Success(t *testing.T) {
	testCases := []struct {
		name   string
		method shift.AssignmentMethod
	}{
		{"auto", shift.AssignmentMethodAuto},
		{"manual", shift.AssignmentMethodManual},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.method.Validate()
			if err != nil {
				t.Errorf("Validate() should succeed for %s, got error: %v", tc.name, err)
			}
		})
	}
}

func TestAssignmentMethod_Validate_Error(t *testing.T) {
	testCases := []struct {
		name   string
		method shift.AssignmentMethod
	}{
		{"empty", shift.AssignmentMethod("")},
		{"invalid", shift.AssignmentMethod("invalid")},
		{"AUTO", shift.AssignmentMethod("AUTO")},
		{"semi-auto", shift.AssignmentMethod("semi-auto")},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.method.Validate()
			if err == nil {
				t.Errorf("Validate() should fail for %s", tc.name)
			}
		})
	}
}

// =====================================================
// AssignmentID Tests
// =====================================================

func TestNewAssignmentID(t *testing.T) {
	id := shift.NewAssignmentID()

	if id.String() == "" {
		t.Error("NewAssignmentID should generate a non-empty ID")
	}

	if err := id.Validate(); err != nil {
		t.Errorf("NewAssignmentID should generate a valid ID: %v", err)
	}
}

func TestNewAssignmentIDWithTime(t *testing.T) {
	now := time.Now()
	id := shift.NewAssignmentIDWithTime(now)

	if id.String() == "" {
		t.Error("NewAssignmentIDWithTime should generate a non-empty ID")
	}

	if err := id.Validate(); err != nil {
		t.Errorf("NewAssignmentIDWithTime should generate a valid ID: %v", err)
	}
}

func TestAssignmentID_Validate_Error(t *testing.T) {
	testCases := []struct {
		name string
		id   shift.AssignmentID
	}{
		{"empty", shift.AssignmentID("")},
		{"invalid_format", shift.AssignmentID("invalid")},
		{"too_short", shift.AssignmentID("0123456789")},
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

func TestParseAssignmentID_Success(t *testing.T) {
	validID := shift.NewAssignmentID()

	parsed, err := shift.ParseAssignmentID(validID.String())
	if err != nil {
		t.Fatalf("ParseAssignmentID() should succeed: %v", err)
	}

	if parsed != validID {
		t.Errorf("ParseAssignmentID mismatch: got %v, want %v", parsed, validID)
	}
}

func TestParseAssignmentID_Error(t *testing.T) {
	testCases := []struct {
		name string
		id   string
	}{
		{"empty", ""},
		{"invalid_format", "invalid"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := shift.ParseAssignmentID(tc.id)
			if err == nil {
				t.Errorf("ParseAssignmentID() should fail for %s", tc.name)
			}
		})
	}
}

// =====================================================
// PlanID Tests
// =====================================================

func TestNewPlanID(t *testing.T) {
	id := shift.NewPlanID()

	if id.String() == "" {
		t.Error("NewPlanID should generate a non-empty ID")
	}

	if err := id.Validate(); err != nil {
		t.Errorf("NewPlanID should generate a valid ID: %v", err)
	}
}

func TestNewPlanIDWithTime(t *testing.T) {
	now := time.Now()
	id := shift.NewPlanIDWithTime(now)

	if id.String() == "" {
		t.Error("NewPlanIDWithTime should generate a non-empty ID")
	}

	if err := id.Validate(); err != nil {
		t.Errorf("NewPlanIDWithTime should generate a valid ID: %v", err)
	}
}

func TestPlanID_Validate_Error(t *testing.T) {
	testCases := []struct {
		name string
		id   shift.PlanID
	}{
		{"empty", shift.PlanID("")},
		{"invalid_format", shift.PlanID("invalid")},
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
// NewShiftAssignment Tests
// =====================================================

func TestNewShiftAssignment_Success_Auto(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	planID := shift.NewPlanID()
	slotID := shift.NewSlotID()
	memberID := common.NewMemberID()

	assignment, err := shift.NewShiftAssignment(
		now,
		tenantID,
		planID,
		slotID,
		memberID,
		shift.AssignmentMethodAuto,
		false,
	)

	if err != nil {
		t.Fatalf("NewShiftAssignment() should succeed: %v", err)
	}

	if assignment.TenantID() != tenantID {
		t.Errorf("TenantID mismatch: got %v, want %v", assignment.TenantID(), tenantID)
	}

	if assignment.PlanID() != planID {
		t.Errorf("PlanID mismatch: got %v, want %v", assignment.PlanID(), planID)
	}

	if assignment.SlotID() != slotID {
		t.Errorf("SlotID mismatch: got %v, want %v", assignment.SlotID(), slotID)
	}

	if assignment.MemberID() != memberID {
		t.Errorf("MemberID mismatch: got %v, want %v", assignment.MemberID(), memberID)
	}

	if assignment.AssignmentMethod() != shift.AssignmentMethodAuto {
		t.Errorf("AssignmentMethod mismatch: got %v, want auto", assignment.AssignmentMethod())
	}

	if assignment.AssignmentStatus() != shift.AssignmentStatusConfirmed {
		t.Errorf("AssignmentStatus should be confirmed, got %v", assignment.AssignmentStatus())
	}

	if assignment.IsOutsidePreference() {
		t.Error("IsOutsidePreference should be false")
	}

	if assignment.AssignmentID().String() == "" {
		t.Error("AssignmentID should be generated")
	}

	if assignment.AssignedAt().IsZero() {
		t.Error("AssignedAt should be set")
	}

	if assignment.CancelledAt() != nil {
		t.Error("CancelledAt should be nil for new assignment")
	}

	if assignment.IsDeleted() {
		t.Error("IsDeleted should be false for new assignment")
	}

	if !assignment.IsConfirmed() {
		t.Error("IsConfirmed should be true for new assignment")
	}

	if assignment.IsCancelled() {
		t.Error("IsCancelled should be false for new assignment")
	}
}

func TestNewShiftAssignment_Success_Manual(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	planID := shift.NewPlanID()
	slotID := shift.NewSlotID()
	memberID := common.NewMemberID()

	assignment, err := shift.NewShiftAssignment(
		now,
		tenantID,
		planID,
		slotID,
		memberID,
		shift.AssignmentMethodManual,
		true,
	)

	if err != nil {
		t.Fatalf("NewShiftAssignment() should succeed: %v", err)
	}

	if assignment.AssignmentMethod() != shift.AssignmentMethodManual {
		t.Errorf("AssignmentMethod mismatch: got %v, want manual", assignment.AssignmentMethod())
	}

	if !assignment.IsOutsidePreference() {
		t.Error("IsOutsidePreference should be true")
	}
}

func TestNewShiftAssignment_Success_EmptyPlanID(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	slotID := shift.NewSlotID()
	memberID := common.NewMemberID()

	// Empty PlanID is allowed
	assignment, err := shift.NewShiftAssignment(
		now,
		tenantID,
		shift.PlanID(""), // Empty plan ID
		slotID,
		memberID,
		shift.AssignmentMethodManual,
		false,
	)

	if err != nil {
		t.Fatalf("NewShiftAssignment() should succeed with empty PlanID: %v", err)
	}

	if assignment.PlanID().String() != "" {
		t.Error("PlanID should be empty")
	}
}

func TestNewShiftAssignment_ErrorWhenInvalidTenantID(t *testing.T) {
	now := time.Now()
	planID := shift.NewPlanID()
	slotID := shift.NewSlotID()
	memberID := common.NewMemberID()

	_, err := shift.NewShiftAssignment(
		now,
		common.TenantID(""), // Invalid
		planID,
		slotID,
		memberID,
		shift.AssignmentMethodManual,
		false,
	)

	if err == nil {
		t.Error("NewShiftAssignment() should fail when tenant_id is invalid")
	}
}

func TestNewShiftAssignment_ErrorWhenInvalidSlotID(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	planID := shift.NewPlanID()
	memberID := common.NewMemberID()

	_, err := shift.NewShiftAssignment(
		now,
		tenantID,
		planID,
		shift.SlotID(""), // Invalid
		memberID,
		shift.AssignmentMethodManual,
		false,
	)

	if err == nil {
		t.Error("NewShiftAssignment() should fail when slot_id is invalid")
	}
}

func TestNewShiftAssignment_ErrorWhenInvalidMemberID(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	planID := shift.NewPlanID()
	slotID := shift.NewSlotID()

	_, err := shift.NewShiftAssignment(
		now,
		tenantID,
		planID,
		slotID,
		common.MemberID(""), // Invalid
		shift.AssignmentMethodManual,
		false,
	)

	if err == nil {
		t.Error("NewShiftAssignment() should fail when member_id is invalid")
	}
}

func TestNewShiftAssignment_ErrorWhenInvalidAssignmentMethod(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	planID := shift.NewPlanID()
	slotID := shift.NewSlotID()
	memberID := common.NewMemberID()

	_, err := shift.NewShiftAssignment(
		now,
		tenantID,
		planID,
		slotID,
		memberID,
		shift.AssignmentMethod("invalid"), // Invalid
		false,
	)

	if err == nil {
		t.Error("NewShiftAssignment() should fail when assignment_method is invalid")
	}
}

// =====================================================
// ReconstructShiftAssignment Tests
// =====================================================

func TestReconstructShiftAssignment_Success(t *testing.T) {
	now := time.Now()
	assignmentID := shift.NewAssignmentID()
	tenantID := common.NewTenantID()
	planID := shift.NewPlanID()
	slotID := shift.NewSlotID()
	memberID := common.NewMemberID()

	assignment, err := shift.ReconstructShiftAssignment(
		assignmentID,
		tenantID,
		planID,
		slotID,
		memberID,
		shift.AssignmentStatusConfirmed,
		shift.AssignmentMethodAuto,
		false,
		now,
		nil,
		now,
		now,
		nil,
	)

	if err != nil {
		t.Fatalf("ReconstructShiftAssignment() should succeed: %v", err)
	}

	if assignment.AssignmentID() != assignmentID {
		t.Errorf("AssignmentID mismatch: got %v, want %v", assignment.AssignmentID(), assignmentID)
	}
}

func TestReconstructShiftAssignment_Success_Cancelled(t *testing.T) {
	now := time.Now()
	assignmentID := shift.NewAssignmentID()
	tenantID := common.NewTenantID()
	planID := shift.NewPlanID()
	slotID := shift.NewSlotID()
	memberID := common.NewMemberID()
	cancelledAt := now

	assignment, err := shift.ReconstructShiftAssignment(
		assignmentID,
		tenantID,
		planID,
		slotID,
		memberID,
		shift.AssignmentStatusCancelled,
		shift.AssignmentMethodManual,
		true,
		now.Add(-time.Hour),
		&cancelledAt,
		now.Add(-time.Hour),
		now,
		nil,
	)

	if err != nil {
		t.Fatalf("ReconstructShiftAssignment() should succeed: %v", err)
	}

	if !assignment.IsCancelled() {
		t.Error("IsCancelled should be true")
	}

	if assignment.CancelledAt() == nil {
		t.Error("CancelledAt should be set")
	}
}

func TestReconstructShiftAssignment_Success_Deleted(t *testing.T) {
	now := time.Now()
	assignmentID := shift.NewAssignmentID()
	tenantID := common.NewTenantID()
	planID := shift.NewPlanID()
	slotID := shift.NewSlotID()
	memberID := common.NewMemberID()
	deletedAt := now

	assignment, err := shift.ReconstructShiftAssignment(
		assignmentID,
		tenantID,
		planID,
		slotID,
		memberID,
		shift.AssignmentStatusConfirmed,
		shift.AssignmentMethodAuto,
		false,
		now,
		nil,
		now,
		now,
		&deletedAt,
	)

	if err != nil {
		t.Fatalf("ReconstructShiftAssignment() should succeed: %v", err)
	}

	if !assignment.IsDeleted() {
		t.Error("IsDeleted should be true")
	}

	if assignment.IsConfirmed() {
		t.Error("IsConfirmed should be false when deleted")
	}
}

func TestReconstructShiftAssignment_ErrorWhenCancelledWithoutCancelledAt(t *testing.T) {
	now := time.Now()
	assignmentID := shift.NewAssignmentID()
	tenantID := common.NewTenantID()
	planID := shift.NewPlanID()
	slotID := shift.NewSlotID()
	memberID := common.NewMemberID()

	_, err := shift.ReconstructShiftAssignment(
		assignmentID,
		tenantID,
		planID,
		slotID,
		memberID,
		shift.AssignmentStatusCancelled, // Cancelled but no cancelledAt
		shift.AssignmentMethodAuto,
		false,
		now,
		nil, // cancelledAt is nil
		now,
		now,
		nil,
	)

	if err == nil {
		t.Error("ReconstructShiftAssignment() should fail when status is cancelled but cancelledAt is nil")
	}
}

func TestReconstructShiftAssignment_ErrorWhenConfirmedWithCancelledAt(t *testing.T) {
	now := time.Now()
	assignmentID := shift.NewAssignmentID()
	tenantID := common.NewTenantID()
	planID := shift.NewPlanID()
	slotID := shift.NewSlotID()
	memberID := common.NewMemberID()
	cancelledAt := now

	_, err := shift.ReconstructShiftAssignment(
		assignmentID,
		tenantID,
		planID,
		slotID,
		memberID,
		shift.AssignmentStatusConfirmed, // Confirmed but with cancelledAt
		shift.AssignmentMethodAuto,
		false,
		now,
		&cancelledAt, // Should be nil for confirmed
		now,
		now,
		nil,
	)

	if err == nil {
		t.Error("ReconstructShiftAssignment() should fail when status is confirmed but cancelledAt is set")
	}
}

// =====================================================
// ShiftAssignment Methods Tests
// =====================================================

func TestShiftAssignment_Cancel_Success(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	planID := shift.NewPlanID()
	slotID := shift.NewSlotID()
	memberID := common.NewMemberID()

	assignment, _ := shift.NewShiftAssignment(
		now,
		tenantID,
		planID,
		slotID,
		memberID,
		shift.AssignmentMethodAuto,
		false,
	)

	if assignment.IsCancelled() {
		t.Error("IsCancelled should be false before Cancel()")
	}

	err := assignment.Cancel(time.Now())
	if err != nil {
		t.Fatalf("Cancel() should succeed: %v", err)
	}

	if !assignment.IsCancelled() {
		t.Error("IsCancelled should be true after Cancel()")
	}

	if assignment.CancelledAt() == nil {
		t.Error("CancelledAt should be set after Cancel()")
	}

	if assignment.AssignmentStatus() != shift.AssignmentStatusCancelled {
		t.Errorf("AssignmentStatus should be cancelled, got %v", assignment.AssignmentStatus())
	}
}

func TestShiftAssignment_Cancel_ErrorWhenAlreadyCancelled(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	planID := shift.NewPlanID()
	slotID := shift.NewSlotID()
	memberID := common.NewMemberID()

	assignment, _ := shift.NewShiftAssignment(
		now,
		tenantID,
		planID,
		slotID,
		memberID,
		shift.AssignmentMethodAuto,
		false,
	)

	// Cancel first time
	_ = assignment.Cancel(time.Now())

	// Cancel second time should fail
	err := assignment.Cancel(time.Now())
	if err == nil {
		t.Error("Cancel() should fail when already cancelled")
	}
}

func TestShiftAssignment_Delete(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	planID := shift.NewPlanID()
	slotID := shift.NewSlotID()
	memberID := common.NewMemberID()

	assignment, _ := shift.NewShiftAssignment(
		now,
		tenantID,
		planID,
		slotID,
		memberID,
		shift.AssignmentMethodAuto,
		false,
	)

	if assignment.IsDeleted() {
		t.Error("IsDeleted should be false before Delete()")
	}

	assignment.Delete(time.Now())

	if !assignment.IsDeleted() {
		t.Error("IsDeleted should be true after Delete()")
	}

	if assignment.DeletedAt() == nil {
		t.Error("DeletedAt should be set after Delete()")
	}
}

func TestShiftAssignment_IsConfirmed_FalseWhenDeleted(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	planID := shift.NewPlanID()
	slotID := shift.NewSlotID()
	memberID := common.NewMemberID()

	assignment, _ := shift.NewShiftAssignment(
		now,
		tenantID,
		planID,
		slotID,
		memberID,
		shift.AssignmentMethodAuto,
		false,
	)

	if !assignment.IsConfirmed() {
		t.Error("IsConfirmed should be true for new assignment")
	}

	assignment.Delete(time.Now())

	if assignment.IsConfirmed() {
		t.Error("IsConfirmed should be false after Delete()")
	}
}
