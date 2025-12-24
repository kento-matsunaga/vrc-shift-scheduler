package tenant_test

import (
	"testing"
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/tenant"
)

// =====================================================
// Tenant Entity Tests
// =====================================================

func TestNewTenant_Success(t *testing.T) {
	now := time.Now()

	ten, err := tenant.NewTenant(now, "Test Organization", "Asia/Tokyo")

	if err != nil {
		t.Fatalf("NewTenant() should succeed, got error: %v", err)
	}

	if ten.TenantID() == "" {
		t.Error("TenantID should be set")
	}

	if ten.TenantName() != "Test Organization" {
		t.Errorf("TenantName mismatch: got %v, want 'Test Organization'", ten.TenantName())
	}

	if ten.Timezone() != "Asia/Tokyo" {
		t.Errorf("Timezone mismatch: got %v, want 'Asia/Tokyo'", ten.Timezone())
	}

	if !ten.IsActive() {
		t.Error("New tenant should be active")
	}

	if ten.Status() != tenant.TenantStatusActive {
		t.Errorf("Status should be active: got %v", ten.Status())
	}

	if ten.IsDeleted() {
		t.Error("New tenant should not be deleted")
	}
}

func TestNewTenant_ErrorWhenNameEmpty(t *testing.T) {
	now := time.Now()

	_, err := tenant.NewTenant(now, "", "Asia/Tokyo")

	if err == nil {
		t.Fatal("NewTenant() should fail when name is empty")
	}
}

func TestNewTenant_ErrorWhenNameTooLong(t *testing.T) {
	now := time.Now()
	longName := make([]byte, 256)
	for i := range longName {
		longName[i] = 'a'
	}

	_, err := tenant.NewTenant(now, string(longName), "Asia/Tokyo")

	if err == nil {
		t.Fatal("NewTenant() should fail when name is too long")
	}
}

func TestNewTenant_ErrorWhenTimezoneEmpty(t *testing.T) {
	now := time.Now()

	_, err := tenant.NewTenant(now, "Test Organization", "")

	if err == nil {
		t.Fatal("NewTenant() should fail when timezone is empty")
	}
}

func TestTenant_UpdateTenantName_Success(t *testing.T) {
	now := time.Now()
	ten, _ := tenant.NewTenant(now, "Original Name", "Asia/Tokyo")

	updateTime := now.Add(1 * time.Hour)
	err := ten.UpdateTenantName(updateTime, "New Name")

	if err != nil {
		t.Fatalf("UpdateTenantName() should succeed, got error: %v", err)
	}

	if ten.TenantName() != "New Name" {
		t.Errorf("TenantName should be updated: got %v, want 'New Name'", ten.TenantName())
	}
}

func TestTenant_UpdateTenantName_ErrorWhenEmpty(t *testing.T) {
	now := time.Now()
	ten, _ := tenant.NewTenant(now, "Original Name", "Asia/Tokyo")

	err := ten.UpdateTenantName(now, "")

	if err == nil {
		t.Fatal("UpdateTenantName() should fail when name is empty")
	}
}

func TestTenant_UpdateTimezone_Success(t *testing.T) {
	now := time.Now()
	ten, _ := tenant.NewTenant(now, "Test Organization", "Asia/Tokyo")

	updateTime := now.Add(1 * time.Hour)
	err := ten.UpdateTimezone(updateTime, "America/New_York")

	if err != nil {
		t.Fatalf("UpdateTimezone() should succeed, got error: %v", err)
	}

	if ten.Timezone() != "America/New_York" {
		t.Errorf("Timezone should be updated: got %v, want 'America/New_York'", ten.Timezone())
	}
}

func TestTenant_UpdateTimezone_ErrorWhenEmpty(t *testing.T) {
	now := time.Now()
	ten, _ := tenant.NewTenant(now, "Test Organization", "Asia/Tokyo")

	err := ten.UpdateTimezone(now, "")

	if err == nil {
		t.Fatal("UpdateTimezone() should fail when timezone is empty")
	}
}

func TestTenant_ActivateDeactivate(t *testing.T) {
	now := time.Now()
	ten, _ := tenant.NewTenant(now, "Test Organization", "Asia/Tokyo")

	// Initially active
	if !ten.IsActive() {
		t.Error("New tenant should be active")
	}

	// Deactivate
	ten.Deactivate(now)
	if ten.IsActive() {
		t.Error("Tenant should be inactive after Deactivate()")
	}

	// Activate again
	ten.Activate(now)
	if !ten.IsActive() {
		t.Error("Tenant should be active after Activate()")
	}
}

func TestTenant_Delete(t *testing.T) {
	now := time.Now()
	ten, _ := tenant.NewTenant(now, "Test Organization", "Asia/Tokyo")

	if ten.IsDeleted() {
		t.Error("New tenant should not be deleted")
	}

	ten.Delete(now)

	if !ten.IsDeleted() {
		t.Error("Tenant should be deleted after Delete()")
	}

	if ten.DeletedAt() == nil {
		t.Error("DeletedAt should be set after Delete()")
	}
}

func TestTenant_StatusTransitions(t *testing.T) {
	now := time.Now()
	ten, _ := tenant.NewTenant(now, "Test Organization", "Asia/Tokyo")

	// Initially active
	if ten.Status() != tenant.TenantStatusActive {
		t.Errorf("Initial status should be active: got %v", ten.Status())
	}

	// Set to grace
	graceUntil := now.Add(7 * 24 * time.Hour) // 7 days grace period
	ten.SetStatusGrace(now, graceUntil)

	if ten.Status() != tenant.TenantStatusGrace {
		t.Errorf("Status should be grace: got %v", ten.Status())
	}
	if ten.GraceUntil() == nil {
		t.Error("GraceUntil should be set")
	}
	if !ten.IsInGracePeriod() {
		t.Error("IsInGracePeriod() should return true")
	}
	if ten.IsActive() {
		t.Error("Tenant should not be active in grace period")
	}

	// Set to suspended
	ten.SetStatusSuspended(now)

	if ten.Status() != tenant.TenantStatusSuspended {
		t.Errorf("Status should be suspended: got %v", ten.Status())
	}
	if !ten.IsSuspended() {
		t.Error("IsSuspended() should return true")
	}
	if ten.GraceUntil() != nil {
		t.Error("GraceUntil should be nil when suspended")
	}

	// Set back to active
	ten.SetStatusActive(now)

	if ten.Status() != tenant.TenantStatusActive {
		t.Errorf("Status should be active: got %v", ten.Status())
	}
	if !ten.IsActive() {
		t.Error("Tenant should be active")
	}
}

func TestTenant_CanWriteCanRead(t *testing.T) {
	now := time.Now()
	ten, _ := tenant.NewTenant(now, "Test Organization", "Asia/Tokyo")

	// Active - can read and write
	if !ten.CanWrite() {
		t.Error("Active tenant should be able to write")
	}
	if !ten.CanRead() {
		t.Error("Active tenant should be able to read")
	}

	// Grace - can read, cannot write
	ten.SetStatusGrace(now, now.Add(7*24*time.Hour))
	if ten.CanWrite() {
		t.Error("Grace tenant should not be able to write")
	}
	if !ten.CanRead() {
		t.Error("Grace tenant should be able to read")
	}

	// Suspended - can read, cannot write
	ten.SetStatusSuspended(now)
	if ten.CanWrite() {
		t.Error("Suspended tenant should not be able to write")
	}
	if !ten.CanRead() {
		t.Error("Suspended tenant should be able to read")
	}

	// Deleted - cannot read or write
	ten.SetStatusActive(now)
	ten.Delete(now)
	if ten.CanWrite() {
		t.Error("Deleted tenant should not be able to write")
	}
	if ten.CanRead() {
		t.Error("Deleted tenant should not be able to read")
	}
}

func TestTenantStatus_IsValid(t *testing.T) {
	tests := []struct {
		status   tenant.TenantStatus
		expected bool
	}{
		{tenant.TenantStatusActive, true},
		{tenant.TenantStatusGrace, true},
		{tenant.TenantStatusSuspended, true},
		{tenant.TenantStatus("invalid"), false},
		{tenant.TenantStatus(""), false},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			result := tt.status.IsValid()
			if result != tt.expected {
				t.Errorf("TenantStatus(%q).IsValid() = %v, want %v", tt.status, result, tt.expected)
			}
		})
	}
}

func TestReconstructTenant_Success(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()

	ten, err := tenant.ReconstructTenant(
		tenantID,
		"Reconstructed Org",
		"Asia/Tokyo",
		true,
		tenant.TenantStatusActive,
		nil,
		now,
		now,
		nil,
	)

	if err != nil {
		t.Fatalf("ReconstructTenant() should succeed, got error: %v", err)
	}

	if ten.TenantName() != "Reconstructed Org" {
		t.Errorf("TenantName mismatch: got %v, want 'Reconstructed Org'", ten.TenantName())
	}
}
