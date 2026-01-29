package db_test

import (
	"context"
	"testing"
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/event"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/member"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/infra/db"
)

// =====================================================
// Security Tests - SQLインジェクション防御
// =====================================================

// TestSQLInjection_MemberDisplayName tests that SQL injection patterns
// in member display names are safely handled by parameterized queries.
func TestSQLInjection_MemberDisplayName(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	repo := db.NewMemberRepository(pool)
	ctx := context.Background()

	tenantID := common.NewTenantID()
	createTestTenant(t, pool, tenantID)

	// SQLインジェクションパターンをテスト
	injectionPatterns := []struct {
		name        string
		displayName string
	}{
		{"UNION SELECT", "'; UNION SELECT * FROM admins; --"},
		{"DROP TABLE", "'; DROP TABLE members; --"},
		{"Single Quote", "Test'Name"},
		{"Double Quote", `Test"Name`},
		{"Semicolon", "Test;Name"},
		{"Comment", "Test--Name"},
		{"Null Byte", "Test\x00Name"},
		{"Backslash", "Test\\Name"},
	}

	for _, tc := range injectionPatterns {
		t.Run(tc.name, func(t *testing.T) {
			now := time.Now()
			// メンバー作成時にエラーが出る可能性があるが、SQLインジェクションは発生しない
			mem, err := member.NewMember(now, tenantID, tc.displayName, "discord_"+tc.name, tc.name+"@example.com")
			if err != nil {
				// ドメインバリデーションでエラーになるのは正常
				t.Logf("Domain validation rejected input (expected): %v", err)
				return
			}

			// 保存を試行 - SQLインジェクションが発生しないことを確認
			err = repo.Save(ctx, mem)
			if err != nil {
				// DBエラーは許容（制約違反など）、ただしSQLインジェクションは発生しない
				t.Logf("Save failed (expected for some patterns): %v", err)
				return
			}

			// 保存成功した場合、正しく取得できることを確認
			found, err := repo.FindByID(ctx, tenantID, mem.MemberID())
			if err != nil {
				t.Errorf("Failed to find saved member: %v", err)
				return
			}

			// 保存した値がそのまま取得できる（エスケープされて安全に処理された）
			if found.DisplayName() != tc.displayName {
				t.Errorf("DisplayName mismatch: got %q, want %q", found.DisplayName(), tc.displayName)
			}
		})
	}
}

// TestSQLInjection_EventTitle tests that SQL injection patterns
// in event titles are safely handled.
func TestSQLInjection_EventTitle(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	repo := db.NewEventRepository(pool)
	ctx := context.Background()

	tenantID := common.NewTenantID()
	createTestTenant(t, pool, tenantID)

	injectionPatterns := []struct {
		name  string
		title string
	}{
		{"UNION SELECT", "'; UNION SELECT * FROM admins; --"},
		{"DROP TABLE", "'; DROP TABLE events; --"},
		{"Single Quote", "Event's Title"},
		{"Double Quote", `Event "Special" Title`},
		{"Semicolon", "Event;Title"},
	}

	for _, tc := range injectionPatterns {
		t.Run(tc.name, func(t *testing.T) {
			now := time.Now()
			evt, err := event.NewEvent(
				now,
				tenantID,
				tc.title,
				event.EventTypeNormal,
				"Test Description",
				event.RecurrenceTypeNone,
				nil, nil, nil, nil,
			)
			if err != nil {
				t.Logf("Domain validation rejected input (expected): %v", err)
				return
			}

			err = repo.Save(ctx, evt)
			if err != nil {
				t.Logf("Save failed (expected for some patterns): %v", err)
				return
			}

			found, err := repo.FindByID(ctx, tenantID, evt.EventID())
			if err != nil {
				t.Errorf("Failed to find saved event: %v", err)
				return
			}

			if found.EventName() != tc.title {
				t.Errorf("EventName mismatch: got %q, want %q", found.EventName(), tc.title)
			}
		})
	}
}

// =====================================================
// Security Tests - テナント分離（クロステナントアクセス防止）
// =====================================================

// TestTenantIsolation_MemberAccess tests that members from one tenant
// cannot be accessed by another tenant.
func TestTenantIsolation_MemberAccess(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	repo := db.NewMemberRepository(pool)
	ctx := context.Background()

	// 2つの異なるテナントを作成
	tenantA := common.NewTenantID()
	tenantB := common.NewTenantID()
	createTestTenant(t, pool, tenantA)
	createTestTenant(t, pool, tenantB)

	// テナントAにメンバーを作成
	now := time.Now()
	memberA, err := member.NewMember(now, tenantA, "Member in Tenant A", "discord_a", "a@example.com")
	if err != nil {
		t.Fatalf("Failed to create member: %v", err)
	}

	err = repo.Save(ctx, memberA)
	if err != nil {
		t.Fatalf("Failed to save member: %v", err)
	}

	// テナントAからは取得可能
	found, err := repo.FindByID(ctx, tenantA, memberA.MemberID())
	if err != nil {
		t.Errorf("Should be able to find member from same tenant: %v", err)
	}
	if found == nil {
		t.Error("Member should be found from same tenant")
	}

	// テナントBからは取得不可（テナント分離）
	found, err = repo.FindByID(ctx, tenantB, memberA.MemberID())
	if err == nil && found != nil {
		t.Error("SECURITY VIOLATION: Should NOT be able to access member from different tenant")
	}
	// エラーまたはnilが返ることを期待
	t.Logf("Cross-tenant access correctly prevented (error: %v, found: %v)", err, found)
}

// TestTenantIsolation_EventAccess tests that events from one tenant
// cannot be accessed by another tenant.
func TestTenantIsolation_EventAccess(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	repo := db.NewEventRepository(pool)
	ctx := context.Background()

	// 2つの異なるテナントを作成
	tenantA := common.NewTenantID()
	tenantB := common.NewTenantID()
	createTestTenant(t, pool, tenantA)
	createTestTenant(t, pool, tenantB)

	// テナントAにイベントを作成
	now := time.Now()
	eventA, err := event.NewEvent(
		now,
		tenantA,
		"Event in Tenant A",
		event.EventTypeNormal,
		"Description",
		event.RecurrenceTypeNone,
		nil, nil, nil, nil,
	)
	if err != nil {
		t.Fatalf("Failed to create event: %v", err)
	}

	err = repo.Save(ctx, eventA)
	if err != nil {
		t.Fatalf("Failed to save event: %v", err)
	}

	// テナントAからは取得可能
	found, err := repo.FindByID(ctx, tenantA, eventA.EventID())
	if err != nil {
		t.Errorf("Should be able to find event from same tenant: %v", err)
	}
	if found == nil {
		t.Error("Event should be found from same tenant")
	}

	// テナントBからは取得不可（テナント分離）
	found, err = repo.FindByID(ctx, tenantB, eventA.EventID())
	if err == nil && found != nil {
		t.Error("SECURITY VIOLATION: Should NOT be able to access event from different tenant")
	}
	t.Logf("Cross-tenant access correctly prevented (error: %v, found: %v)", err, found)
}

// TestTenantIsolation_MemberList tests that listing members only returns
// members from the requesting tenant.
func TestTenantIsolation_MemberList(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	repo := db.NewMemberRepository(pool)
	ctx := context.Background()

	// 2つの異なるテナントを作成
	tenantA := common.NewTenantID()
	tenantB := common.NewTenantID()
	createTestTenant(t, pool, tenantA)
	createTestTenant(t, pool, tenantB)

	now := time.Now()

	// テナントAに2人のメンバーを作成
	memberA1, _ := member.NewMember(now, tenantA, "Member A1", "discord_a1", "a1@example.com")
	memberA2, _ := member.NewMember(now, tenantA, "Member A2", "discord_a2", "a2@example.com")
	repo.Save(ctx, memberA1)
	repo.Save(ctx, memberA2)

	// テナントBに1人のメンバーを作成
	memberB1, _ := member.NewMember(now, tenantB, "Member B1", "discord_b1", "b1@example.com")
	repo.Save(ctx, memberB1)

	// テナントAのメンバー一覧を取得
	membersA, err := repo.FindByTenantID(ctx, tenantA)
	if err != nil {
		t.Fatalf("Failed to list members for tenant A: %v", err)
	}

	// テナントAのメンバーのみが返される
	for _, m := range membersA {
		if m.TenantID() != tenantA {
			t.Errorf("SECURITY VIOLATION: Member from different tenant returned: %v", m.TenantID())
		}
	}

	// テナントBのメンバー一覧を取得
	membersB, err := repo.FindByTenantID(ctx, tenantB)
	if err != nil {
		t.Fatalf("Failed to list members for tenant B: %v", err)
	}

	// テナントBのメンバーのみが返される
	for _, m := range membersB {
		if m.TenantID() != tenantB {
			t.Errorf("SECURITY VIOLATION: Member from different tenant returned: %v", m.TenantID())
		}
	}

	// 件数の確認
	if len(membersA) < 2 {
		t.Errorf("Tenant A should have at least 2 members, got %d", len(membersA))
	}
	if len(membersB) < 1 {
		t.Errorf("Tenant B should have at least 1 member, got %d", len(membersB))
	}
}
