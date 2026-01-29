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

func TestNewTenant_ErrorWhenTimezoneInvalid(t *testing.T) {
	now := time.Now()

	_, err := tenant.NewTenant(now, "Test Organization", "Invalid/Timezone")

	if err == nil {
		t.Fatal("NewTenant() should fail when timezone is invalid IANA format")
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

	// Set to grace (valid: active -> grace)
	graceUntil := now.Add(7 * 24 * time.Hour) // 7 days grace period
	if err := ten.SetStatusGrace(now, graceUntil); err != nil {
		t.Errorf("SetStatusGrace should succeed: %v", err)
	}

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

	// Set to suspended (valid: grace -> suspended)
	if err := ten.SetStatusSuspended(now); err != nil {
		t.Errorf("SetStatusSuspended should succeed: %v", err)
	}

	if ten.Status() != tenant.TenantStatusSuspended {
		t.Errorf("Status should be suspended: got %v", ten.Status())
	}
	if !ten.IsSuspended() {
		t.Error("IsSuspended() should return true")
	}
	if ten.GraceUntil() != nil {
		t.Error("GraceUntil should be nil when suspended")
	}

	// Set back to active (valid: suspended -> active)
	if err := ten.SetStatusActive(now); err != nil {
		t.Errorf("SetStatusActive should succeed: %v", err)
	}

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

	// Grace - can read, cannot write (valid: active -> grace)
	_ = ten.SetStatusGrace(now, now.Add(7*24*time.Hour))
	if ten.CanWrite() {
		t.Error("Grace tenant should not be able to write")
	}
	if !ten.CanRead() {
		t.Error("Grace tenant should be able to read")
	}

	// Suspended - can read, cannot write (valid: grace -> suspended)
	_ = ten.SetStatusSuspended(now)
	if ten.CanWrite() {
		t.Error("Suspended tenant should not be able to write")
	}
	if !ten.CanRead() {
		t.Error("Suspended tenant should be able to read")
	}

	// Deleted - cannot read or write (valid: suspended -> active, then delete)
	_ = ten.SetStatusActive(now)
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
		{tenant.TenantStatusPendingPayment, true},
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
		nil, // graceUntil
		nil, // pendingExpiresAt
		nil, // pendingStripeSessionID
		now,
		now,
		nil, // deletedAt
	)

	if err != nil {
		t.Fatalf("ReconstructTenant() should succeed, got error: %v", err)
	}

	if ten.TenantName() != "Reconstructed Org" {
		t.Errorf("TenantName mismatch: got %v, want 'Reconstructed Org'", ten.TenantName())
	}
}

func TestTenant_PendingPaymentStatus(t *testing.T) {
	now := time.Now()
	expiresAt := now.Add(30 * time.Minute)

	// NewTenantPendingPaymentのテスト
	ten, err := tenant.NewTenantPendingPayment(now, "Test Organization", "Asia/Tokyo", "cs_test_xxx", expiresAt)

	if err != nil {
		t.Fatalf("NewTenantPendingPayment() should succeed, got error: %v", err)
	}

	if ten.Status() != tenant.TenantStatusPendingPayment {
		t.Errorf("Status should be pending_payment: got %v", ten.Status())
	}

	if !ten.IsPendingPayment() {
		t.Error("IsPendingPayment() should return true")
	}

	if ten.PendingStripeSessionID() == nil || *ten.PendingStripeSessionID() != "cs_test_xxx" {
		t.Error("PendingStripeSessionID should be set")
	}

	if ten.PendingExpiresAt() == nil {
		t.Error("PendingExpiresAt should be set")
	}

	if ten.IsActive() {
		t.Error("Pending payment tenant should not be active")
	}

	// SetStatusActiveでpendingフィールドがクリアされることを確認 (valid: pending_payment -> active)
	if err := ten.SetStatusActive(now); err != nil {
		t.Errorf("SetStatusActive should succeed: %v", err)
	}

	if ten.PendingStripeSessionID() != nil {
		t.Error("PendingStripeSessionID should be cleared after SetStatusActive")
	}

	if ten.PendingExpiresAt() != nil {
		t.Error("PendingExpiresAt should be cleared after SetStatusActive")
	}
}

func TestCalculateGraceUntil(t *testing.T) {
	// Test that grace period is exactly 14 days
	periodEnd := time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)
	graceUntil := tenant.CalculateGraceUntil(periodEnd)

	expected := time.Date(2026, 2, 14, 0, 0, 0, 0, time.UTC)
	if !graceUntil.Equal(expected) {
		t.Errorf("CalculateGraceUntil() = %v, want %v", graceUntil, expected)
	}
}

func TestTransitionToGraceAfterSubscriptionEnd(t *testing.T) {
	now := time.Now()
	periodEnd := time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)

	ten, _ := tenant.NewTenant(now, "Test Organization", "Asia/Tokyo")
	if err := ten.TransitionToGraceAfterSubscriptionEnd(now, periodEnd); err != nil {
		t.Errorf("TransitionToGraceAfterSubscriptionEnd should succeed: %v", err)
	}

	if ten.Status() != tenant.TenantStatusGrace {
		t.Errorf("Status should be grace: got %v", ten.Status())
	}

	if ten.GraceUntil() == nil {
		t.Fatal("GraceUntil should be set")
	}

	expected := time.Date(2026, 2, 14, 0, 0, 0, 0, time.UTC)
	if !ten.GraceUntil().Equal(expected) {
		t.Errorf("GraceUntil() = %v, want %v", ten.GraceUntil(), expected)
	}

	if ten.IsActive() {
		t.Error("Tenant should not be active after transitioning to grace")
	}
}

func TestTenantStatus_CanTransitionTo(t *testing.T) {
	tests := []struct {
		name     string
		from     tenant.TenantStatus
		to       tenant.TenantStatus
		expected bool
	}{
		// Valid transitions from pending_payment
		{"pending_payment -> active", tenant.TenantStatusPendingPayment, tenant.TenantStatusActive, true},
		{"pending_payment -> suspended", tenant.TenantStatusPendingPayment, tenant.TenantStatusSuspended, true},
		{"pending_payment -> grace (invalid)", tenant.TenantStatusPendingPayment, tenant.TenantStatusGrace, false},

		// Valid transitions from active
		{"active -> grace", tenant.TenantStatusActive, tenant.TenantStatusGrace, true},
		{"active -> suspended", tenant.TenantStatusActive, tenant.TenantStatusSuspended, true},
		{"active -> pending_payment (invalid)", tenant.TenantStatusActive, tenant.TenantStatusPendingPayment, false},

		// Valid transitions from grace
		{"grace -> active", tenant.TenantStatusGrace, tenant.TenantStatusActive, true},
		{"grace -> suspended", tenant.TenantStatusGrace, tenant.TenantStatusSuspended, true},
		{"grace -> pending_payment (invalid)", tenant.TenantStatusGrace, tenant.TenantStatusPendingPayment, false},

		// Valid transitions from suspended
		{"suspended -> pending_payment", tenant.TenantStatusSuspended, tenant.TenantStatusPendingPayment, true},
		{"suspended -> active", tenant.TenantStatusSuspended, tenant.TenantStatusActive, true},
		{"suspended -> grace (invalid)", tenant.TenantStatusSuspended, tenant.TenantStatusGrace, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.from.CanTransitionTo(tt.to)
			if result != tt.expected {
				t.Errorf("CanTransitionTo(%s -> %s) = %v, want %v",
					tt.from, tt.to, result, tt.expected)
			}
		})
	}
}

func TestTenant_SetStatus_InvalidTransitions(t *testing.T) {
	now := time.Now()

	// Test: active -> pending_payment (invalid)
	t.Run("active -> pending_payment should fail", func(t *testing.T) {
		ten, _ := tenant.NewTenant(now, "Test Organization", "Asia/Tokyo")
		err := ten.SetStatusPendingPayment(now, "cs_test_xxx", now.Add(30*time.Minute))
		if err == nil {
			t.Error("SetStatusPendingPayment should fail for active -> pending_payment")
		}
	})

	// Test: grace -> pending_payment (invalid)
	t.Run("grace -> pending_payment should fail", func(t *testing.T) {
		ten, _ := tenant.NewTenant(now, "Test Organization", "Asia/Tokyo")
		_ = ten.SetStatusGrace(now, now.Add(7*24*time.Hour))
		err := ten.SetStatusPendingPayment(now, "cs_test_xxx", now.Add(30*time.Minute))
		if err == nil {
			t.Error("SetStatusPendingPayment should fail for grace -> pending_payment")
		}
	})

	// Test: pending_payment -> grace (invalid)
	t.Run("pending_payment -> grace should fail", func(t *testing.T) {
		ten, _ := tenant.NewTenantPendingPayment(now, "Test Organization", "Asia/Tokyo", "cs_test_xxx", now.Add(30*time.Minute))
		err := ten.SetStatusGrace(now, now.Add(7*24*time.Hour))
		if err == nil {
			t.Error("SetStatusGrace should fail for pending_payment -> grace")
		}
	})

	// Test: suspended -> grace (invalid)
	t.Run("suspended -> grace should fail", func(t *testing.T) {
		ten, _ := tenant.NewTenant(now, "Test Organization", "Asia/Tokyo")
		_ = ten.SetStatusGrace(now, now.Add(7*24*time.Hour))
		_ = ten.SetStatusSuspended(now)
		err := ten.SetStatusGrace(now, now.Add(7*24*time.Hour))
		if err == nil {
			t.Error("SetStatusGrace should fail for suspended -> grace")
		}
	})
}
