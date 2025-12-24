package db_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/attendance"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/member"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/infra/db"
)

// =====================================================
// AttendanceRepository Integration Tests
// =====================================================

func TestAttendanceRepository_SaveAndFindByID(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	repo := db.NewAttendanceRepository(pool)
	ctx := context.Background()

	// テスト用のテナントを作成
	tenantID := common.NewTenantID()
	createTestTenant(t, pool, tenantID)

	// テスト用のコレクションを作成
	now := time.Now()
	collection, err := attendance.NewAttendanceCollection(
		now,
		tenantID,
		"12月イベント出欠確認",
		"12月のイベントへの参加可否を回答してください",
		attendance.TargetTypeEvent,
		"event-123",
		nil,
	)
	if err != nil {
		t.Fatalf("Failed to create test collection: %v", err)
	}

	// Save
	err = repo.Save(ctx, collection)
	if err != nil {
		t.Fatalf("Failed to save collection: %v", err)
	}

	// FindByID
	foundCollection, err := repo.FindByID(ctx, tenantID, collection.CollectionID())
	if err != nil {
		t.Fatalf("Failed to find collection by ID: %v", err)
	}

	// 値の検証
	if foundCollection.CollectionID() != collection.CollectionID() {
		t.Errorf("CollectionID mismatch: got %v, want %v", foundCollection.CollectionID(), collection.CollectionID())
	}

	if foundCollection.TenantID() != collection.TenantID() {
		t.Errorf("TenantID mismatch: got %v, want %v", foundCollection.TenantID(), collection.TenantID())
	}

	if foundCollection.Title() != collection.Title() {
		t.Errorf("Title mismatch: got %v, want %v", foundCollection.Title(), collection.Title())
	}

	if foundCollection.Description() != collection.Description() {
		t.Errorf("Description mismatch: got %v, want %v", foundCollection.Description(), collection.Description())
	}

	if foundCollection.TargetType() != collection.TargetType() {
		t.Errorf("TargetType mismatch: got %v, want %v", foundCollection.TargetType(), collection.TargetType())
	}

	if foundCollection.Status() != collection.Status() {
		t.Errorf("Status mismatch: got %v, want %v", foundCollection.Status(), collection.Status())
	}

	if foundCollection.PublicToken() != collection.PublicToken() {
		t.Errorf("PublicToken mismatch: got %v, want %v", foundCollection.PublicToken(), collection.PublicToken())
	}
}

func TestAttendanceRepository_FindByToken(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	repo := db.NewAttendanceRepository(pool)
	ctx := context.Background()

	tenantID := common.NewTenantID()
	createTestTenant(t, pool, tenantID)

	now := time.Now()
	collection, err := attendance.NewAttendanceCollection(
		now,
		tenantID,
		"トークン検索テスト",
		"",
		attendance.TargetTypeEvent,
		"",
		nil,
	)
	if err != nil {
		t.Fatalf("Failed to create test collection: %v", err)
	}

	err = repo.Save(ctx, collection)
	if err != nil {
		t.Fatalf("Failed to save collection: %v", err)
	}

	// PublicTokenで検索
	foundCollection, err := repo.FindByToken(ctx, collection.PublicToken())
	if err != nil {
		t.Fatalf("Failed to find collection by token: %v", err)
	}

	if foundCollection.CollectionID() != collection.CollectionID() {
		t.Errorf("CollectionID mismatch: got %v, want %v", foundCollection.CollectionID(), collection.CollectionID())
	}

	// 存在しないトークンで検索
	nonExistentToken := common.NewPublicToken()
	_, err = repo.FindByToken(ctx, nonExistentToken)
	if err == nil {
		t.Error("Expected error when finding collection with nonexistent token")
	}
}

func TestAttendanceRepository_FindByTenantID(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	repo := db.NewAttendanceRepository(pool)
	ctx := context.Background()

	tenantID := common.NewTenantID()
	createTestTenant(t, pool, tenantID)

	// 3つのコレクションを作成
	now := time.Now()
	for i := 0; i < 3; i++ {
		collection, err := attendance.NewAttendanceCollection(
			now,
			tenantID,
			"テストコレクション"+string(rune('1'+i)),
			"",
			attendance.TargetTypeEvent,
			"",
			nil,
		)
		if err != nil {
			t.Fatalf("Failed to create test collection: %v", err)
		}

		err = repo.Save(ctx, collection)
		if err != nil {
			t.Fatalf("Failed to save collection: %v", err)
		}
	}

	// テナントIDで検索
	collections, err := repo.FindByTenantID(ctx, tenantID)
	if err != nil {
		t.Fatalf("Failed to find collections by tenant ID: %v", err)
	}

	if len(collections) < 3 {
		t.Errorf("Expected at least 3 collections, got %d", len(collections))
	}

	// 全ての結果が同じテナントIDを持つことを確認
	for _, c := range collections {
		if c.TenantID() != tenantID {
			t.Errorf("Collection has wrong tenant ID: got %v, want %v", c.TenantID(), tenantID)
		}
	}
}

func TestAttendanceRepository_SaveWithDeadline(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	repo := db.NewAttendanceRepository(pool)
	ctx := context.Background()

	tenantID := common.NewTenantID()
	createTestTenant(t, pool, tenantID)

	now := time.Now()
	deadline := now.Add(7 * 24 * time.Hour) // 1週間後

	collection, err := attendance.NewAttendanceCollection(
		now,
		tenantID,
		"締切付きコレクション",
		"",
		attendance.TargetTypeEvent,
		"",
		&deadline,
	)
	if err != nil {
		t.Fatalf("Failed to create test collection: %v", err)
	}

	err = repo.Save(ctx, collection)
	if err != nil {
		t.Fatalf("Failed to save collection: %v", err)
	}

	foundCollection, err := repo.FindByID(ctx, tenantID, collection.CollectionID())
	if err != nil {
		t.Fatalf("Failed to find collection: %v", err)
	}

	if foundCollection.Deadline() == nil {
		t.Error("Deadline should not be nil")
	}
}

func TestAttendanceRepository_BusinessDayTargetType(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	repo := db.NewAttendanceRepository(pool)
	ctx := context.Background()

	tenantID := common.NewTenantID()
	createTestTenant(t, pool, tenantID)

	now := time.Now()
	collection, err := attendance.NewAttendanceCollection(
		now,
		tenantID,
		"営業日出欠確認",
		"",
		attendance.TargetTypeBusinessDay,
		"bd-456",
		nil,
	)
	if err != nil {
		t.Fatalf("Failed to create test collection: %v", err)
	}

	err = repo.Save(ctx, collection)
	if err != nil {
		t.Fatalf("Failed to save collection: %v", err)
	}

	foundCollection, err := repo.FindByID(ctx, tenantID, collection.CollectionID())
	if err != nil {
		t.Fatalf("Failed to find collection: %v", err)
	}

	if foundCollection.TargetType() != attendance.TargetTypeBusinessDay {
		t.Errorf("TargetType mismatch: got %v, want %v", foundCollection.TargetType(), attendance.TargetTypeBusinessDay)
	}

	if strings.TrimSpace(foundCollection.TargetID()) != "bd-456" {
		t.Errorf("TargetID mismatch: got %q, want 'bd-456'", foundCollection.TargetID())
	}
}

func TestAttendanceRepository_UpsertResponse(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	repo := db.NewAttendanceRepository(pool)
	memberRepo := db.NewMemberRepository(pool)
	ctx := context.Background()

	tenantID := common.NewTenantID()
	createTestTenant(t, pool, tenantID)

	// コレクションを作成
	now := time.Now()
	collection, err := attendance.NewAttendanceCollection(
		now,
		tenantID,
		"回答テスト",
		"",
		attendance.TargetTypeEvent,
		"",
		nil,
	)
	if err != nil {
		t.Fatalf("Failed to create test collection: %v", err)
	}
	err = repo.Save(ctx, collection)
	if err != nil {
		t.Fatalf("Failed to save collection: %v", err)
	}

	// メンバーを作成
	mem, err := member.NewMember(now, tenantID, "テスト回答者", "discord_response_test", "response@example.com")
	if err != nil {
		t.Fatalf("Failed to create test member: %v", err)
	}
	err = memberRepo.Save(ctx, mem)
	if err != nil {
		t.Fatalf("Failed to save member: %v", err)
	}

	// 対象日を作成
	targetDate, err := attendance.NewTargetDate(now, collection.CollectionID(), now.Add(24*time.Hour), 1)
	if err != nil {
		t.Fatalf("Failed to create target date: %v", err)
	}
	err = repo.SaveTargetDates(ctx, collection.CollectionID(), []*attendance.TargetDate{targetDate})
	if err != nil {
		t.Fatalf("Failed to save target dates: %v", err)
	}

	// 回答を作成
	response, err := attendance.NewAttendanceResponse(
		now,
		collection.CollectionID(),
		tenantID,
		mem.MemberID(),
		targetDate.TargetDateID(),
		attendance.ResponseTypeAttending,
		"参加します",
	)
	if err != nil {
		t.Fatalf("Failed to create response: %v", err)
	}

	// Upsert
	err = repo.UpsertResponse(ctx, response)
	if err != nil {
		t.Fatalf("Failed to upsert response: %v", err)
	}

	// 回答を取得して確認
	responses, err := repo.FindResponsesByCollectionID(ctx, collection.CollectionID())
	if err != nil {
		t.Fatalf("Failed to find responses: %v", err)
	}

	if len(responses) != 1 {
		t.Errorf("Expected 1 response, got %d", len(responses))
	}

	if responses[0].Response() != attendance.ResponseTypeAttending {
		t.Errorf("Response mismatch: got %v, want %v", responses[0].Response(), attendance.ResponseTypeAttending)
	}

	// 同じメンバー・対象日で回答を更新
	updatedResponse, err := attendance.NewAttendanceResponse(
		now,
		collection.CollectionID(),
		tenantID,
		mem.MemberID(),
		targetDate.TargetDateID(),
		attendance.ResponseTypeAbsent,
		"予定が入りました",
	)
	if err != nil {
		t.Fatalf("Failed to create updated response: %v", err)
	}

	err = repo.UpsertResponse(ctx, updatedResponse)
	if err != nil {
		t.Fatalf("Failed to upsert updated response: %v", err)
	}

	// 更新されているか確認
	responses, err = repo.FindResponsesByCollectionID(ctx, collection.CollectionID())
	if err != nil {
		t.Fatalf("Failed to find responses after update: %v", err)
	}

	if len(responses) != 1 {
		t.Errorf("Expected 1 response after upsert, got %d", len(responses))
	}

	if responses[0].Response() != attendance.ResponseTypeAbsent {
		t.Errorf("Response not updated: got %v, want %v", responses[0].Response(), attendance.ResponseTypeAbsent)
	}
}

func TestAttendanceRepository_FindResponsesByMemberID(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	repo := db.NewAttendanceRepository(pool)
	memberRepo := db.NewMemberRepository(pool)
	ctx := context.Background()

	tenantID := common.NewTenantID()
	createTestTenant(t, pool, tenantID)

	now := time.Now()

	// メンバーを作成
	mem, err := member.NewMember(now, tenantID, "メンバー回答検索", "discord_member_responses", "member-responses@example.com")
	if err != nil {
		t.Fatalf("Failed to create test member: %v", err)
	}
	err = memberRepo.Save(ctx, mem)
	if err != nil {
		t.Fatalf("Failed to save member: %v", err)
	}

	// 2つのコレクションを作成し、それぞれに回答
	for i := 0; i < 2; i++ {
		collection, err := attendance.NewAttendanceCollection(
			now,
			tenantID,
			"コレクション"+string(rune('A'+i)),
			"",
			attendance.TargetTypeEvent,
			"",
			nil,
		)
		if err != nil {
			t.Fatalf("Failed to create collection %d: %v", i, err)
		}
		err = repo.Save(ctx, collection)
		if err != nil {
			t.Fatalf("Failed to save collection %d: %v", i, err)
		}

		// 対象日を作成
		targetDate, err := attendance.NewTargetDate(now, collection.CollectionID(), now.Add(time.Duration(i+1)*24*time.Hour), 1)
		if err != nil {
			t.Fatalf("Failed to create target date %d: %v", i, err)
		}
		err = repo.SaveTargetDates(ctx, collection.CollectionID(), []*attendance.TargetDate{targetDate})
		if err != nil {
			t.Fatalf("Failed to save target dates %d: %v", i, err)
		}

		response, err := attendance.NewAttendanceResponse(
			now,
			collection.CollectionID(),
			tenantID,
			mem.MemberID(),
			targetDate.TargetDateID(),
			attendance.ResponseTypeAttending,
			"",
		)
		if err != nil {
			t.Fatalf("Failed to create response %d: %v", i, err)
		}
		err = repo.UpsertResponse(ctx, response)
		if err != nil {
			t.Fatalf("Failed to upsert response %d: %v", i, err)
		}
	}

	// メンバーIDで回答を検索
	responses, err := repo.FindResponsesByMemberID(ctx, tenantID, mem.MemberID())
	if err != nil {
		t.Fatalf("Failed to find responses by member ID: %v", err)
	}

	if len(responses) < 2 {
		t.Errorf("Expected at least 2 responses, got %d", len(responses))
	}

	// 全ての回答が同じメンバーIDを持つことを確認
	for _, r := range responses {
		if r.MemberID() != mem.MemberID() {
			t.Errorf("Response has wrong member ID: got %v, want %v", r.MemberID(), mem.MemberID())
		}
	}
}

func TestAttendanceRepository_SaveTargetDates(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	repo := db.NewAttendanceRepository(pool)
	ctx := context.Background()

	tenantID := common.NewTenantID()
	createTestTenant(t, pool, tenantID)

	now := time.Now()
	collection, err := attendance.NewAttendanceCollection(
		now,
		tenantID,
		"対象日テスト",
		"",
		attendance.TargetTypeEvent,
		"",
		nil,
	)
	if err != nil {
		t.Fatalf("Failed to create test collection: %v", err)
	}
	err = repo.Save(ctx, collection)
	if err != nil {
		t.Fatalf("Failed to save collection: %v", err)
	}

	// 3つの対象日を作成
	var targetDates []*attendance.TargetDate
	for i := 0; i < 3; i++ {
		td, err := attendance.NewTargetDate(
			now,
			collection.CollectionID(),
			now.Add(time.Duration(i+1)*24*time.Hour),
			i+1,
		)
		if err != nil {
			t.Fatalf("Failed to create target date %d: %v", i, err)
		}
		targetDates = append(targetDates, td)
	}

	// 保存
	err = repo.SaveTargetDates(ctx, collection.CollectionID(), targetDates)
	if err != nil {
		t.Fatalf("Failed to save target dates: %v", err)
	}

	// 取得して確認
	foundDates, err := repo.FindTargetDatesByCollectionID(ctx, collection.CollectionID())
	if err != nil {
		t.Fatalf("Failed to find target dates: %v", err)
	}

	if len(foundDates) != 3 {
		t.Errorf("Expected 3 target dates, got %d", len(foundDates))
	}

	// 新しい対象日で上書き（1つだけ）
	newTargetDate, err := attendance.NewTargetDate(now, collection.CollectionID(), now.Add(10*24*time.Hour), 1)
	if err != nil {
		t.Fatalf("Failed to create new target date: %v", err)
	}

	err = repo.SaveTargetDates(ctx, collection.CollectionID(), []*attendance.TargetDate{newTargetDate})
	if err != nil {
		t.Fatalf("Failed to save new target dates: %v", err)
	}

	foundDates, err = repo.FindTargetDatesByCollectionID(ctx, collection.CollectionID())
	if err != nil {
		t.Fatalf("Failed to find target dates after update: %v", err)
	}

	if len(foundDates) != 1 {
		t.Errorf("Expected 1 target date after update, got %d", len(foundDates))
	}
}

func TestAttendanceRepository_TenantIsolation(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	repo := db.NewAttendanceRepository(pool)
	ctx := context.Background()

	// 2つのテナントを作成
	tenantID1 := common.NewTenantID()
	tenantID2 := common.NewTenantID()
	createTestTenant(t, pool, tenantID1)
	createTestTenant(t, pool, tenantID2)

	now := time.Now()

	// テナント1のコレクションを作成
	collection1, err := attendance.NewAttendanceCollection(
		now,
		tenantID1,
		"テナント1のコレクション",
		"",
		attendance.TargetTypeEvent,
		"",
		nil,
	)
	if err != nil {
		t.Fatalf("Failed to create collection for tenant 1: %v", err)
	}
	err = repo.Save(ctx, collection1)
	if err != nil {
		t.Fatalf("Failed to save collection for tenant 1: %v", err)
	}

	// テナント2のコレクションを作成
	collection2, err := attendance.NewAttendanceCollection(
		now,
		tenantID2,
		"テナント2のコレクション",
		"",
		attendance.TargetTypeEvent,
		"",
		nil,
	)
	if err != nil {
		t.Fatalf("Failed to create collection for tenant 2: %v", err)
	}
	err = repo.Save(ctx, collection2)
	if err != nil {
		t.Fatalf("Failed to save collection for tenant 2: %v", err)
	}

	// テナント1のコレクションはテナント2からは見えない
	_, err = repo.FindByID(ctx, tenantID2, collection1.CollectionID())
	if err == nil {
		t.Error("Collection from tenant 1 should not be visible from tenant 2")
	}

	// テナント2のコレクションはテナント1からは見えない
	_, err = repo.FindByID(ctx, tenantID1, collection2.CollectionID())
	if err == nil {
		t.Error("Collection from tenant 2 should not be visible from tenant 1")
	}

	// 各テナントのコレクション一覧は自分のテナントのコレクションのみ
	collections1, err := repo.FindByTenantID(ctx, tenantID1)
	if err != nil {
		t.Fatalf("Failed to find collections for tenant 1: %v", err)
	}
	for _, c := range collections1 {
		if c.TenantID() != tenantID1 {
			t.Errorf("Found collection from wrong tenant: got %v, want %v", c.TenantID(), tenantID1)
		}
	}

	collections2, err := repo.FindByTenantID(ctx, tenantID2)
	if err != nil {
		t.Fatalf("Failed to find collections for tenant 2: %v", err)
	}
	for _, c := range collections2 {
		if c.TenantID() != tenantID2 {
			t.Errorf("Found collection from wrong tenant: got %v, want %v", c.TenantID(), tenantID2)
		}
	}
}

func TestAttendanceRepository_UpdateStatus(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	repo := db.NewAttendanceRepository(pool)
	ctx := context.Background()

	tenantID := common.NewTenantID()
	createTestTenant(t, pool, tenantID)

	now := time.Now()
	collection, err := attendance.NewAttendanceCollection(
		now,
		tenantID,
		"ステータス更新テスト",
		"",
		attendance.TargetTypeEvent,
		"",
		nil,
	)
	if err != nil {
		t.Fatalf("Failed to create test collection: %v", err)
	}

	// 最初は open
	if collection.Status() != attendance.StatusOpen {
		t.Errorf("Initial status should be open: got %v", collection.Status())
	}

	err = repo.Save(ctx, collection)
	if err != nil {
		t.Fatalf("Failed to save collection: %v", err)
	}

	// ステータスを closed に変更
	err = collection.Close(now)
	if err != nil {
		t.Fatalf("Failed to close collection: %v", err)
	}

	err = repo.Save(ctx, collection)
	if err != nil {
		t.Fatalf("Failed to update collection: %v", err)
	}

	// 取得して確認
	foundCollection, err := repo.FindByID(ctx, tenantID, collection.CollectionID())
	if err != nil {
		t.Fatalf("Failed to find collection: %v", err)
	}

	if foundCollection.Status() != attendance.StatusClosed {
		t.Errorf("Status should be closed: got %v", foundCollection.Status())
	}
}
