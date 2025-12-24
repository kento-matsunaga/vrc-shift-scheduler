package billing

import (
	"testing"
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
)

// =====================================================
// NewEntitlementID Tests
// =====================================================

func TestNewEntitlementID_Success(t *testing.T) {
	id := NewEntitlementID()

	if id == "" {
		t.Error("NewEntitlementID() should not return empty string")
	}
}

func TestNewEntitlementIDWithTime_Success(t *testing.T) {
	fixedTime := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	id := NewEntitlementIDWithTime(fixedTime)

	if id == "" {
		t.Error("NewEntitlementIDWithTime() should not return empty string")
	}
}

func TestParseEntitlementID(t *testing.T) {
	t.Run("valid ID", func(t *testing.T) {
		validID := "01ARZ3NDEKTSV4RRFFQ69G5FAV"
		id, err := ParseEntitlementID(validID)
		if err != nil {
			t.Errorf("ParseEntitlementID(%q) unexpected error: %v", validID, err)
		}
		if id.String() != validID {
			t.Errorf("ParseEntitlementID(%q) = %q, want %q", validID, id.String(), validID)
		}
	})

	t.Run("invalid ID", func(t *testing.T) {
		invalidID := "invalid"
		_, err := ParseEntitlementID(invalidID)
		if err == nil {
			t.Errorf("ParseEntitlementID(%q) expected error, got nil", invalidID)
		}
	})
}

// =====================================================
// EntitlementSource Tests
// =====================================================

func TestEntitlementSource_IsValid(t *testing.T) {
	tests := []struct {
		name   string
		source EntitlementSource
		want   bool
	}{
		{"booth is valid", EntitlementSourceBooth, true},
		{"stripe is valid", EntitlementSourceStripe, true},
		{"manual is valid", EntitlementSourceManual, true},
		{"invalid source", EntitlementSource("invalid"), false},
		{"empty source", EntitlementSource(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.source.IsValid()
			if got != tt.want {
				t.Errorf("IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEntitlementSource_String(t *testing.T) {
	tests := []struct {
		source EntitlementSource
		want   string
	}{
		{EntitlementSourceBooth, "booth"},
		{EntitlementSourceStripe, "stripe"},
		{EntitlementSourceManual, "manual"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.source.String()
			if got != tt.want {
				t.Errorf("String() = %s, want %s", got, tt.want)
			}
		})
	}
}

// =====================================================
// NewEntitlement Tests - Success Cases
// =====================================================

func TestNewEntitlement_Success(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	planCode := "LIFETIME"
	source := EntitlementSourceBooth

	entitlement, err := NewEntitlement(now, tenantID, planCode, source, nil)

	if err != nil {
		t.Fatalf("NewEntitlement() should succeed, but got error: %v", err)
	}

	if entitlement == nil {
		t.Fatal("NewEntitlement() returned nil")
	}

	// Basic field validation
	if entitlement.TenantID() != tenantID {
		t.Errorf("TenantID: expected %s, got %s", tenantID, entitlement.TenantID())
	}
	if entitlement.PlanCode() != planCode {
		t.Errorf("PlanCode: expected %s, got %s", planCode, entitlement.PlanCode())
	}
	if entitlement.Source() != source {
		t.Errorf("Source: expected %s, got %s", source, entitlement.Source())
	}

	// Default values
	if entitlement.IsRevoked() {
		t.Error("IsRevoked should be false by default")
	}
	if !entitlement.IsLifetime() {
		t.Error("IsLifetime should be true when endsAt is nil")
	}

	// ID should be generated
	if entitlement.EntitlementID() == "" {
		t.Error("EntitlementID should not be empty")
	}

	// Timestamps
	if entitlement.CreatedAt().IsZero() {
		t.Error("CreatedAt should not be zero")
	}
	if entitlement.StartsAt().IsZero() {
		t.Error("StartsAt should not be zero")
	}
}

func TestNewEntitlement_SuccessWithEndsAt(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	planCode := "SUB_980"
	source := EntitlementSourceStripe
	endsAt := now.Add(30 * 24 * time.Hour) // 30 days later

	entitlement, err := NewEntitlement(now, tenantID, planCode, source, &endsAt)

	if err != nil {
		t.Fatalf("NewEntitlement() should succeed, but got error: %v", err)
	}

	if entitlement.IsLifetime() {
		t.Error("IsLifetime should be false when endsAt is set")
	}
	if entitlement.EndsAt() == nil {
		t.Fatal("EndsAt should not be nil")
	}
	if !entitlement.EndsAt().Equal(endsAt) {
		t.Errorf("EndsAt: expected %v, got %v", endsAt, entitlement.EndsAt())
	}
}

// =====================================================
// NewEntitlement Tests - Error Cases
// =====================================================

func TestNewEntitlement_ErrorWhenPlanCodeEmpty(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	planCode := "" // Empty
	source := EntitlementSourceBooth

	entitlement, err := NewEntitlement(now, tenantID, planCode, source, nil)

	if err == nil {
		t.Fatal("NewEntitlement() should return error when plan_code is empty")
	}
	if entitlement != nil {
		t.Error("NewEntitlement() should return nil when validation fails")
	}
}

func TestNewEntitlement_ErrorWhenSourceInvalid(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	planCode := "LIFETIME"
	source := EntitlementSource("invalid") // Invalid

	entitlement, err := NewEntitlement(now, tenantID, planCode, source, nil)

	if err == nil {
		t.Fatal("NewEntitlement() should return error when source is invalid")
	}
	if entitlement != nil {
		t.Error("NewEntitlement() should return nil when validation fails")
	}
}

// =====================================================
// IsActive Tests
// =====================================================

func TestEntitlement_IsActive_ActiveLifetime(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	entitlement, _ := NewEntitlement(now, tenantID, "LIFETIME", EntitlementSourceBooth, nil)

	if !entitlement.IsActive(now) {
		t.Error("IsActive() should return true for active lifetime entitlement")
	}

	// Also active in the future
	future := now.Add(365 * 24 * time.Hour)
	if !entitlement.IsActive(future) {
		t.Error("IsActive() should return true for lifetime entitlement even in far future")
	}
}

func TestEntitlement_IsActive_WithinEndsAt(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	endsAt := now.Add(30 * 24 * time.Hour)
	entitlement, _ := NewEntitlement(now, tenantID, "SUB_980", EntitlementSourceStripe, &endsAt)

	// Active now
	if !entitlement.IsActive(now) {
		t.Error("IsActive() should return true before endsAt")
	}

	// Active just before expiry
	justBefore := endsAt.Add(-1 * time.Minute)
	if !entitlement.IsActive(justBefore) {
		t.Error("IsActive() should return true just before endsAt")
	}
}

func TestEntitlement_IsActive_ExpiredEndsAt(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	endsAt := now.Add(30 * 24 * time.Hour)
	entitlement, _ := NewEntitlement(now, tenantID, "SUB_980", EntitlementSourceStripe, &endsAt)

	// After expiry
	afterExpiry := endsAt.Add(1 * time.Minute)
	if entitlement.IsActive(afterExpiry) {
		t.Error("IsActive() should return false after endsAt")
	}
}

func TestEntitlement_IsActive_WhenRevoked(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	entitlement, _ := NewEntitlement(now, tenantID, "LIFETIME", EntitlementSourceBooth, nil)

	// Revoke it
	entitlement.Revoke(now, "refund requested")

	if entitlement.IsActive(now) {
		t.Error("IsActive() should return false when entitlement is revoked")
	}
}

func TestEntitlement_IsActive_BeforeStartsAt(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	entitlement, _ := NewEntitlement(now, tenantID, "LIFETIME", EntitlementSourceBooth, nil)

	// Check before startsAt
	beforeStart := now.Add(-1 * time.Hour)
	if entitlement.IsActive(beforeStart) {
		t.Error("IsActive() should return false before startsAt")
	}
}

// =====================================================
// Revoke Tests
// =====================================================

func TestEntitlement_Revoke(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	entitlement, _ := NewEntitlement(now, tenantID, "LIFETIME", EntitlementSourceBooth, nil)

	// Initially not revoked
	if entitlement.IsRevoked() {
		t.Error("Entitlement should not be revoked by default")
	}
	if entitlement.RevokedAt() != nil {
		t.Error("RevokedAt should be nil by default")
	}
	if entitlement.RevokedReason() != nil {
		t.Error("RevokedReason should be nil by default")
	}

	// Revoke
	revokeTime := now.Add(1 * time.Hour)
	reason := "chargeback"
	entitlement.Revoke(revokeTime, reason)

	if !entitlement.IsRevoked() {
		t.Error("Entitlement should be revoked after Revoke()")
	}
	if entitlement.RevokedAt() == nil {
		t.Fatal("RevokedAt should not be nil after Revoke()")
	}
	if !entitlement.RevokedAt().Equal(revokeTime) {
		t.Errorf("RevokedAt: expected %v, got %v", revokeTime, entitlement.RevokedAt())
	}
	if entitlement.RevokedReason() == nil {
		t.Fatal("RevokedReason should not be nil after Revoke()")
	}
	if *entitlement.RevokedReason() != reason {
		t.Errorf("RevokedReason: expected %s, got %s", reason, *entitlement.RevokedReason())
	}
}

// =====================================================
// ExtendEndsAt Tests
// =====================================================

func TestEntitlement_ExtendEndsAt(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	originalEndsAt := now.Add(30 * 24 * time.Hour)
	entitlement, _ := NewEntitlement(now, tenantID, "SUB_980", EntitlementSourceStripe, &originalEndsAt)

	newEndsAt := now.Add(60 * 24 * time.Hour)
	entitlement.ExtendEndsAt(now, newEndsAt)

	if entitlement.EndsAt() == nil {
		t.Fatal("EndsAt should not be nil")
	}
	if !entitlement.EndsAt().Equal(newEndsAt) {
		t.Errorf("EndsAt: expected %v, got %v", newEndsAt, entitlement.EndsAt())
	}
}

// =====================================================
// ReconstructEntitlement Tests
// =====================================================

func TestReconstructEntitlement_Success(t *testing.T) {
	entitlementID := NewEntitlementID()
	tenantID := common.NewTenantID()
	now := time.Now()

	entitlement, err := ReconstructEntitlement(
		entitlementID,
		tenantID,
		"LIFETIME",
		EntitlementSourceBooth,
		now,
		nil,
		nil,
		nil,
		now,
		now,
	)

	if err != nil {
		t.Fatalf("ReconstructEntitlement() should succeed, but got error: %v", err)
	}
	if entitlement == nil {
		t.Fatal("ReconstructEntitlement() returned nil")
	}
	if entitlement.EntitlementID() != entitlementID {
		t.Errorf("EntitlementID: expected %s, got %s", entitlementID, entitlement.EntitlementID())
	}
}

func TestReconstructEntitlement_WithRevokedState(t *testing.T) {
	entitlementID := NewEntitlementID()
	tenantID := common.NewTenantID()
	now := time.Now()
	revokedAt := now.Add(-1 * time.Hour)
	reason := "refund"

	entitlement, err := ReconstructEntitlement(
		entitlementID,
		tenantID,
		"LIFETIME",
		EntitlementSourceBooth,
		now,
		nil,
		&revokedAt,
		&reason,
		now,
		now,
	)

	if err != nil {
		t.Fatalf("ReconstructEntitlement() should succeed, but got error: %v", err)
	}
	if !entitlement.IsRevoked() {
		t.Error("Entitlement should be marked as revoked")
	}
}

func TestReconstructEntitlement_ErrorWhenValidationFails(t *testing.T) {
	entitlementID := NewEntitlementID()
	tenantID := common.NewTenantID()
	now := time.Now()

	entitlement, err := ReconstructEntitlement(
		entitlementID,
		tenantID,
		"", // Invalid - empty plan code
		EntitlementSourceBooth,
		now,
		nil,
		nil,
		nil,
		now,
		now,
	)

	if err == nil {
		t.Fatal("ReconstructEntitlement() should return error when validation fails")
	}
	if entitlement != nil {
		t.Error("ReconstructEntitlement() should return nil when validation fails")
	}
}
