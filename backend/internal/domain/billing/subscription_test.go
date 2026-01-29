package billing_test

import (
	"testing"
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/billing"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
)

func TestSubscriptionStatus_CanTransitionTo(t *testing.T) {
	tests := []struct {
		name     string
		from     billing.SubscriptionStatus
		to       billing.SubscriptionStatus
		expected bool
	}{
		// Valid transitions from incomplete
		{"incomplete -> active", billing.SubscriptionStatusIncomplete, billing.SubscriptionStatusActive, true},
		{"incomplete -> canceled", billing.SubscriptionStatusIncomplete, billing.SubscriptionStatusCanceled, true},
		{"incomplete -> past_due (invalid)", billing.SubscriptionStatusIncomplete, billing.SubscriptionStatusPastDue, false},

		// Valid transitions from trialing
		{"trialing -> active", billing.SubscriptionStatusTrialing, billing.SubscriptionStatusActive, true},
		{"trialing -> past_due", billing.SubscriptionStatusTrialing, billing.SubscriptionStatusPastDue, true},
		{"trialing -> canceled", billing.SubscriptionStatusTrialing, billing.SubscriptionStatusCanceled, true},
		{"trialing -> unpaid (invalid)", billing.SubscriptionStatusTrialing, billing.SubscriptionStatusUnpaid, false},

		// Valid transitions from active
		{"active -> past_due", billing.SubscriptionStatusActive, billing.SubscriptionStatusPastDue, true},
		{"active -> canceled", billing.SubscriptionStatusActive, billing.SubscriptionStatusCanceled, true},
		{"active -> unpaid", billing.SubscriptionStatusActive, billing.SubscriptionStatusUnpaid, true},
		{"active -> incomplete (invalid)", billing.SubscriptionStatusActive, billing.SubscriptionStatusIncomplete, false},

		// Valid transitions from past_due
		{"past_due -> active", billing.SubscriptionStatusPastDue, billing.SubscriptionStatusActive, true},
		{"past_due -> canceled", billing.SubscriptionStatusPastDue, billing.SubscriptionStatusCanceled, true},
		{"past_due -> unpaid", billing.SubscriptionStatusPastDue, billing.SubscriptionStatusUnpaid, true},
		{"past_due -> trialing (invalid)", billing.SubscriptionStatusPastDue, billing.SubscriptionStatusTrialing, false},

		// Valid transitions from unpaid
		{"unpaid -> active", billing.SubscriptionStatusUnpaid, billing.SubscriptionStatusActive, true},
		{"unpaid -> canceled", billing.SubscriptionStatusUnpaid, billing.SubscriptionStatusCanceled, true},
		{"unpaid -> past_due (invalid)", billing.SubscriptionStatusUnpaid, billing.SubscriptionStatusPastDue, false},

		// Canceled is terminal state
		{"canceled -> active (invalid)", billing.SubscriptionStatusCanceled, billing.SubscriptionStatusActive, false},
		{"canceled -> unpaid (invalid)", billing.SubscriptionStatusCanceled, billing.SubscriptionStatusUnpaid, false},
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

func TestSubscription_UpdateStatus_ValidTransition(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	periodEnd := now.Add(30 * 24 * time.Hour)

	// Create subscription with active status
	sub, err := billing.NewSubscription(
		now,
		tenantID,
		"cus_test123",
		"sub_test123",
		billing.SubscriptionStatusActive,
		&periodEnd,
	)
	if err != nil {
		t.Fatalf("NewSubscription failed: %v", err)
	}

	// Valid transition: active -> past_due
	err = sub.UpdateStatus(now, billing.SubscriptionStatusPastDue, &periodEnd)
	if err != nil {
		t.Errorf("UpdateStatus should succeed for valid transition: %v", err)
	}
	if sub.Status() != billing.SubscriptionStatusPastDue {
		t.Errorf("Status should be past_due: got %v", sub.Status())
	}

	// Valid transition: past_due -> active
	err = sub.UpdateStatus(now, billing.SubscriptionStatusActive, &periodEnd)
	if err != nil {
		t.Errorf("UpdateStatus should succeed for valid transition: %v", err)
	}
	if sub.Status() != billing.SubscriptionStatusActive {
		t.Errorf("Status should be active: got %v", sub.Status())
	}

	// Valid transition: active -> canceled
	err = sub.UpdateStatus(now, billing.SubscriptionStatusCanceled, &periodEnd)
	if err != nil {
		t.Errorf("UpdateStatus should succeed for valid transition: %v", err)
	}
	if sub.Status() != billing.SubscriptionStatusCanceled {
		t.Errorf("Status should be canceled: got %v", sub.Status())
	}
}

func TestSubscription_UpdateStatus_InvalidTransition(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	periodEnd := now.Add(30 * 24 * time.Hour)

	// Create subscription with active status
	sub, err := billing.NewSubscription(
		now,
		tenantID,
		"cus_test123",
		"sub_test123",
		billing.SubscriptionStatusActive,
		&periodEnd,
	)
	if err != nil {
		t.Fatalf("NewSubscription failed: %v", err)
	}

	// Invalid transition: active -> incomplete
	err = sub.UpdateStatus(now, billing.SubscriptionStatusIncomplete, &periodEnd)
	if err == nil {
		t.Error("UpdateStatus should fail for invalid transition active -> incomplete")
	}

	// Status should remain active
	if sub.Status() != billing.SubscriptionStatusActive {
		t.Errorf("Status should remain active after failed transition: got %v", sub.Status())
	}
}

func TestSubscription_UpdateStatus_SameStatusAllowed(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	periodEnd := now.Add(30 * 24 * time.Hour)

	// Create subscription with active status
	sub, _ := billing.NewSubscription(
		now,
		tenantID,
		"cus_test123",
		"sub_test123",
		billing.SubscriptionStatusActive,
		&periodEnd,
	)

	// Same status transition (e.g., subscription renewal) should be allowed
	err := sub.UpdateStatus(now, billing.SubscriptionStatusActive, &periodEnd)
	if err != nil {
		t.Errorf("UpdateStatus should allow same status transition (renewal): %v", err)
	}
}

func TestSubscription_UpdateStatus_CanceledIsTerminal(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	periodEnd := now.Add(30 * 24 * time.Hour)

	// Create subscription with active status and transition to canceled
	sub, _ := billing.NewSubscription(
		now,
		tenantID,
		"cus_test123",
		"sub_test123",
		billing.SubscriptionStatusActive,
		&periodEnd,
	)
	_ = sub.UpdateStatus(now, billing.SubscriptionStatusCanceled, &periodEnd)

	// Try to transition from canceled (should fail for any status)
	err := sub.UpdateStatus(now, billing.SubscriptionStatusActive, &periodEnd)
	if err == nil {
		t.Error("UpdateStatus should fail for transition from canceled status")
	}
}
