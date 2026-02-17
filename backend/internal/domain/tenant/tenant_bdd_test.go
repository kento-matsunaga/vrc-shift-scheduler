package tenant_test

import (
	"testing"
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/tenant"
)

// =====================================================
// BDD-style tests for Tenant Status Transitions
// These tests correspond to: features/tenant_status.feature
// =====================================================

func TestBDD_TenantStatus_SubscriptionEnd_TransitionsToGrace(t *testing.T) {
	// Scenario: サブスク解約後、猶予期間に入る

	// Given: テナント「VRChat Japan」のステータスが「active」である
	now := time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)
	ten, err := tenant.NewTenant(now, "VRChat Japan", "Asia/Tokyo")
	if err != nil {
		t.Fatalf("Given failed: %v", err)
	}
	if ten.Status() != tenant.TenantStatusActive {
		t.Fatalf("Given failed: expected active, got %s", ten.Status())
	}

	// When: サブスクリプション期間が終了する（期間終了日: 2026-01-31）
	periodEnd := time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)
	err = ten.TransitionToGraceAfterSubscriptionEnd(now, periodEnd)

	// Then: テナントのステータスが「grace」に遷移する
	if err != nil {
		t.Fatalf("Then failed: transition should succeed: %v", err)
	}
	if ten.Status() != tenant.TenantStatusGrace {
		t.Errorf("Then failed: expected grace, got %s", ten.Status())
	}

	// And: 猶予期限が「2026-02-14」に設定される（期間終了日 + 14日）
	expectedGraceUntil := time.Date(2026, 2, 14, 0, 0, 0, 0, time.UTC)
	if ten.GraceUntil() == nil || !ten.GraceUntil().Equal(expectedGraceUntil) {
		t.Errorf("Then failed: grace_until should be %v, got %v", expectedGraceUntil, ten.GraceUntil())
	}

	// And: テナントのデータは読み取り可能である
	if !ten.CanRead() {
		t.Error("Then failed: grace tenant should be able to read")
	}

	// And: テナントへの書き込みは拒否される
	if ten.CanWrite() {
		t.Error("Then failed: grace tenant should not be able to write")
	}
}

func TestBDD_TenantStatus_GraceToActive_ResubscribeSuccess(t *testing.T) {
	// Scenario: 猶予期間中に再決済して復帰する

	// Given: テナント「VRChat Japan」のステータスが「grace」である
	now := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	ten, _ := tenant.NewTenant(now, "VRChat Japan", "Asia/Tokyo")
	graceUntil := now.Add(14 * 24 * time.Hour)
	_ = ten.SetStatusGrace(now, graceUntil)
	if ten.Status() != tenant.TenantStatusGrace {
		t.Fatalf("Given failed: expected grace, got %s", ten.Status())
	}

	// When: 管理者が再決済を完了する
	err := ten.SetStatusActive(now)

	// Then: テナントのステータスが「active」に遷移する
	if err != nil {
		t.Fatalf("Then failed: transition should succeed: %v", err)
	}
	if ten.Status() != tenant.TenantStatusActive {
		t.Errorf("Then failed: expected active, got %s", ten.Status())
	}

	// And: 猶予期限がクリアされる
	if ten.GraceUntil() != nil {
		t.Error("Then failed: grace_until should be cleared")
	}

	// And: テナントの全機能が利用可能になる
	if !ten.CanRead() || !ten.CanWrite() {
		t.Error("Then failed: active tenant should have full access")
	}
}

func TestBDD_TenantStatus_GraceExpired_TransitionsToSuspended(t *testing.T) {
	// Scenario: 猶予期間が切れてサービス停止になる

	// Given: テナント「VRChat Japan」のステータスが「grace」である
	now := time.Date(2026, 2, 15, 0, 0, 0, 0, time.UTC)
	ten, _ := tenant.NewTenant(now, "VRChat Japan", "Asia/Tokyo")
	_ = ten.SetStatusGrace(now, now.Add(14*24*time.Hour))

	// When: 猶予期間（14日間）が終了する
	err := ten.SetStatusSuspended(now)

	// Then: テナントのステータスが「suspended」に遷移する
	if err != nil {
		t.Fatalf("Then failed: transition should succeed: %v", err)
	}
	if ten.Status() != tenant.TenantStatusSuspended {
		t.Errorf("Then failed: expected suspended, got %s", ten.Status())
	}

	// And: 猶予期限がクリアされる
	if ten.GraceUntil() != nil {
		t.Error("Then failed: grace_until should be cleared")
	}
}

func TestBDD_TenantStatus_SuspendedToPendingPayment(t *testing.T) {
	// Scenario: 停止状態から再決済を開始する

	// Given: テナント「VRChat Japan」のステータスが「suspended」である
	now := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	ten, _ := tenant.NewTenant(now, "VRChat Japan", "Asia/Tokyo")
	_ = ten.SetStatusGrace(now, now.Add(14*24*time.Hour))
	_ = ten.SetStatusSuspended(now)
	if ten.Status() != tenant.TenantStatusSuspended {
		t.Fatalf("Given failed: expected suspended, got %s", ten.Status())
	}

	// When: 管理者が再決済を開始する（Stripe Session: cs_test_xxx）
	expiresAt := now.Add(30 * time.Minute)
	err := ten.SetStatusPendingPayment(now, "cs_test_xxx", expiresAt)

	// Then: テナントのステータスが「pending_payment」に遷移する
	if err != nil {
		t.Fatalf("Then failed: transition should succeed: %v", err)
	}
	if ten.Status() != tenant.TenantStatusPendingPayment {
		t.Errorf("Then failed: expected pending_payment, got %s", ten.Status())
	}

	// And: 決済セッションIDが保存される
	if ten.PendingStripeSessionID() == nil || *ten.PendingStripeSessionID() != "cs_test_xxx" {
		t.Error("Then failed: stripe session ID should be saved")
	}

	// And: 決済有効期限が設定される
	if ten.PendingExpiresAt() == nil {
		t.Error("Then failed: pending expires_at should be set")
	}
}

func TestBDD_TenantStatus_PendingPaymentToActive(t *testing.T) {
	// Scenario: 決済完了でアクティブに復帰する

	// Given: テナント「VRChat Japan」のステータスが「pending_payment」である
	now := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	expiresAt := now.Add(30 * time.Minute)
	ten, _ := tenant.NewTenantPendingPayment(now, "VRChat Japan", "Asia/Tokyo", "cs_test_xxx", expiresAt)
	if ten.Status() != tenant.TenantStatusPendingPayment {
		t.Fatalf("Given failed: expected pending_payment, got %s", ten.Status())
	}

	// When: Stripe決済が完了する
	err := ten.SetStatusActive(now)

	// Then: テナントのステータスが「active」に遷移する
	if err != nil {
		t.Fatalf("Then failed: transition should succeed: %v", err)
	}
	if ten.Status() != tenant.TenantStatusActive {
		t.Errorf("Then failed: expected active, got %s", ten.Status())
	}

	// And: 決済セッション情報がクリアされる
	if ten.PendingStripeSessionID() != nil {
		t.Error("Then failed: stripe session ID should be cleared")
	}
	if ten.PendingExpiresAt() != nil {
		t.Error("Then failed: pending expires_at should be cleared")
	}

	// And: テナントの全機能が利用可能になる
	if !ten.CanRead() || !ten.CanWrite() {
		t.Error("Then failed: active tenant should have full access")
	}
}

// =====================================================
// 禁止された状態遷移
// =====================================================

func TestBDD_TenantStatus_InvalidTransition_ActiveToPendingPayment(t *testing.T) {
	// Scenario: active から直接 pending_payment には遷移できない

	// Given: テナント「VRChat Japan」のステータスが「active」である
	now := time.Now()
	ten, _ := tenant.NewTenant(now, "VRChat Japan", "Asia/Tokyo")

	// When: ステータスを「pending_payment」に変更しようとする
	err := ten.SetStatusPendingPayment(now, "cs_test_xxx", now.Add(30*time.Minute))

	// Then: エラー「invalid status transition from active to pending_payment」が返される
	if err == nil {
		t.Fatal("Then failed: transition should be rejected")
	}

	// And: ステータスは「active」のまま変更されない
	if ten.Status() != tenant.TenantStatusActive {
		t.Errorf("Then failed: status should remain active, got %s", ten.Status())
	}
}

func TestBDD_TenantStatus_InvalidTransition_GraceToPendingPayment(t *testing.T) {
	// Scenario: grace から pending_payment には遷移できない

	// Given: テナント「VRChat Japan」のステータスが「grace」である
	now := time.Now()
	ten, _ := tenant.NewTenant(now, "VRChat Japan", "Asia/Tokyo")
	_ = ten.SetStatusGrace(now, now.Add(14*24*time.Hour))

	// When: ステータスを「pending_payment」に変更しようとする
	err := ten.SetStatusPendingPayment(now, "cs_test_xxx", now.Add(30*time.Minute))

	// Then: エラーが返される
	if err == nil {
		t.Fatal("Then failed: transition should be rejected")
	}

	// And: ステータスは「grace」のまま変更されない
	if ten.Status() != tenant.TenantStatusGrace {
		t.Errorf("Then failed: status should remain grace, got %s", ten.Status())
	}
}

func TestBDD_TenantStatus_InvalidTransition_SuspendedToGrace(t *testing.T) {
	// Scenario: suspended から直接 grace には遷移できない

	// Given: テナント「VRChat Japan」のステータスが「suspended」である
	now := time.Now()
	ten, _ := tenant.NewTenant(now, "VRChat Japan", "Asia/Tokyo")
	_ = ten.SetStatusGrace(now, now.Add(14*24*time.Hour))
	_ = ten.SetStatusSuspended(now)

	// When: ステータスを「grace」に変更しようとする
	err := ten.SetStatusGrace(now, now.Add(14*24*time.Hour))

	// Then: エラーが返される
	if err == nil {
		t.Fatal("Then failed: transition should be rejected")
	}

	// And: ステータスは「suspended」のまま変更されない
	if ten.Status() != tenant.TenantStatusSuspended {
		t.Errorf("Then failed: status should remain suspended, got %s", ten.Status())
	}
}

// =====================================================
// アクセス制御
// =====================================================

func TestBDD_TenantStatus_AccessControl(t *testing.T) {
	// Scenario Outline: ステータスごとのアクセス制御

	now := time.Now()
	expiresAt := now.Add(30 * time.Minute)

	tests := []struct {
		name     string
		status   string
		canRead  bool
		canWrite bool
	}{
		{"active: 読み書き可能", "active", true, true},
		{"grace: 読み取りのみ", "grace", true, false},
		{"suspended: 読み取りのみ", "suspended", true, false},
		{"pending_payment: 全操作不可", "pending_payment", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given: テナントのステータスが「<status>」である
			var ten *tenant.Tenant
			switch tt.status {
			case "active":
				ten, _ = tenant.NewTenant(now, "VRChat Japan", "Asia/Tokyo")
			case "grace":
				ten, _ = tenant.NewTenant(now, "VRChat Japan", "Asia/Tokyo")
				_ = ten.SetStatusGrace(now, now.Add(14*24*time.Hour))
			case "suspended":
				ten, _ = tenant.NewTenant(now, "VRChat Japan", "Asia/Tokyo")
				_ = ten.SetStatusGrace(now, now.Add(14*24*time.Hour))
				_ = ten.SetStatusSuspended(now)
			case "pending_payment":
				ten, _ = tenant.NewTenantPendingPayment(now, "VRChat Japan", "Asia/Tokyo", "cs_test_xxx", expiresAt)
			}

			// Then: データの読み取りは「<can_read>」である
			if ten.CanRead() != tt.canRead {
				t.Errorf("CanRead() = %v, want %v", ten.CanRead(), tt.canRead)
			}

			// And: データの書き込みは「<can_write>」である
			if ten.CanWrite() != tt.canWrite {
				t.Errorf("CanWrite() = %v, want %v", ten.CanWrite(), tt.canWrite)
			}
		})
	}
}

// =====================================================
// ソフトデリート
// =====================================================

func TestBDD_TenantStatus_SoftDelete_AllOperationsDenied(t *testing.T) {
	// Scenario: 削除されたテナントは全操作が不可

	// Given: テナント「VRChat Japan」のステータスが「active」である
	now := time.Now()
	ten, _ := tenant.NewTenant(now, "VRChat Japan", "Asia/Tokyo")

	// When: テナントが削除される（ソフトデリート）
	ten.Delete(now)

	// Then: データの読み取りは不可である
	if ten.CanRead() {
		t.Error("Then failed: deleted tenant should not be readable")
	}

	// And: データの書き込みは不可である
	if ten.CanWrite() {
		t.Error("Then failed: deleted tenant should not be writable")
	}

	// And: deleted_at が記録される
	if ten.DeletedAt() == nil {
		t.Error("Then failed: deleted_at should be set")
	}
}
