package db_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/event"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/infra/db"
	"github.com/jackc/pgx/v5/pgxpool"
)

// このテストは実際のPostgreSQLデータベースが必要です
// DATABASE_URL 環境変数を設定して実行してください
// 例: DATABASE_URL=postgres://user:pass@localhost:5432/testdb go test -v ./internal/infra/db/

func setupTestDB(t *testing.T) (*pgxpool.Pool, func()) {
	t.Helper()

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Skip("DATABASE_URL not set, skipping integration test")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// データベース接続確認
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		t.Fatalf("Failed to ping test database: %v", err)
	}

	cleanup := func() {
		pool.Close()
	}

	return pool, cleanup
}

// createTestTenant はテスト用のテナントを作成します
func createTestTenant(t *testing.T, pool *pgxpool.Pool, tenantID common.TenantID) {
	t.Helper()
	ctx := context.Background()
	query := `
		INSERT INTO tenants (tenant_id, tenant_name, timezone, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
		ON CONFLICT (tenant_id) DO NOTHING
	`
	_, err := pool.Exec(ctx, query, string(tenantID), "Test Tenant", "Asia/Tokyo", true)
	if err != nil {
		t.Fatalf("Failed to create test tenant: %v", err)
	}
}

func TestEventRepository_SaveAndFind(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	repo := db.NewEventRepository(pool)
	ctx := context.Background()

	// テスト用のテナントとイベントを作成
	tenantID := common.NewTenantID()
	createTestTenant(t, pool, tenantID)
	now := time.Now()
	testEvent, err := event.NewEvent(
		now,
		tenantID,
		"週末VRChat集会",
		event.EventTypeNormal,
		"毎週末に開催するVRChat集会イベント",
		event.RecurrenceTypeNone,
		nil, nil, nil, nil,
	)
	if err != nil {
		t.Fatalf("Failed to create test event: %v", err)
	}

	// Save
	err = repo.Save(ctx, testEvent)
	if err != nil {
		t.Fatalf("Failed to save event: %v", err)
	}

	// FindByID
	foundEvent, err := repo.FindByID(ctx, tenantID, testEvent.EventID())
	if err != nil {
		t.Fatalf("Failed to find event: %v", err)
	}

	// 値の検証
	if foundEvent.EventID() != testEvent.EventID() {
		t.Errorf("EventID mismatch: got %v, want %v", foundEvent.EventID(), testEvent.EventID())
	}

	if foundEvent.EventName() != testEvent.EventName() {
		t.Errorf("EventName mismatch: got %v, want %v", foundEvent.EventName(), testEvent.EventName())
	}

	if foundEvent.EventType() != testEvent.EventType() {
		t.Errorf("EventType mismatch: got %v, want %v", foundEvent.EventType(), testEvent.EventType())
	}

	if foundEvent.Description() != testEvent.Description() {
		t.Errorf("Description mismatch: got %v, want %v", foundEvent.Description(), testEvent.Description())
	}

	if foundEvent.IsActive() != testEvent.IsActive() {
		t.Errorf("IsActive mismatch: got %v, want %v", foundEvent.IsActive(), testEvent.IsActive())
	}
}

func TestEventRepository_FindByTenantID(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	repo := db.NewEventRepository(pool)
	ctx := context.Background()

	// テスト用のテナントを作成
	tenantID := common.NewTenantID()
	createTestTenant(t, pool, tenantID)

	// 複数のイベントを作成
	now := time.Now()
	events := make([]*event.Event, 3)
	for i := 0; i < 3; i++ {
		e, err := event.NewEvent(
			now,
			tenantID,
			fmt.Sprintf("テストイベント%d", i+1),
			event.EventTypeNormal,
			fmt.Sprintf("説明%d", i+1),
			event.RecurrenceTypeNone,
			nil, nil, nil, nil,
		)
		if err != nil {
			t.Fatalf("Failed to create test event: %v", err)
		}
		events[i] = e

		err = repo.Save(ctx, e)
		if err != nil {
			t.Fatalf("Failed to save event: %v", err)
		}
	}

	// FindByTenantID
	foundEvents, err := repo.FindByTenantID(ctx, tenantID)
	if err != nil {
		t.Fatalf("Failed to find events by tenant: %v", err)
	}

	// 最低でも作成した3件が見つかるはず
	if len(foundEvents) < 3 {
		t.Errorf("Expected at least 3 events, got %d", len(foundEvents))
	}
}

func TestEventRepository_Update(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	repo := db.NewEventRepository(pool)
	ctx := context.Background()

	tenantID := common.NewTenantID()
	createTestTenant(t, pool, tenantID)
	now := time.Now()
	testEvent, err := event.NewEvent(
		now,
		tenantID,
		"元の名前",
		event.EventTypeNormal,
		"元の説明",
		event.RecurrenceTypeNone,
		nil, nil, nil, nil,
	)
	if err != nil {
		t.Fatalf("Failed to create test event: %v", err)
	}

	// 最初の保存
	err = repo.Save(ctx, testEvent)
	if err != nil {
		t.Fatalf("Failed to save event: %v", err)
	}

	// 更新
	err = testEvent.UpdateEventName("新しい名前")
	if err != nil {
		t.Fatalf("Failed to update event name: %v", err)
	}

	// 再度保存（更新）
	err = repo.Save(ctx, testEvent)
	if err != nil {
		t.Fatalf("Failed to update event: %v", err)
	}

	// 取得して確認
	foundEvent, err := repo.FindByID(ctx, tenantID, testEvent.EventID())
	if err != nil {
		t.Fatalf("Failed to find event: %v", err)
	}

	if foundEvent.EventName() != "新しい名前" {
		t.Errorf("EventName not updated: got %v, want '新しい名前'", foundEvent.EventName())
	}
}

func TestEventRepository_ExistsByName(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	repo := db.NewEventRepository(pool)
	ctx := context.Background()

	tenantID := common.NewTenantID()
	createTestTenant(t, pool, tenantID)
	eventName := "ユニークなイベント名"

	// 最初は存在しない
	exists, err := repo.ExistsByName(ctx, tenantID, eventName)
	if err != nil {
		t.Fatalf("Failed to check existence: %v", err)
	}
	if exists {
		t.Error("Event should not exist yet")
	}

	// イベントを作成・保存
	now := time.Now()
	testEvent, err := event.NewEvent(now, tenantID, eventName, event.EventTypeNormal, "説明", event.RecurrenceTypeNone, nil, nil, nil, nil)
	if err != nil {
		t.Fatalf("Failed to create event: %v", err)
	}

	err = repo.Save(ctx, testEvent)
	if err != nil {
		t.Fatalf("Failed to save event: %v", err)
	}

	// 存在確認
	exists, err = repo.ExistsByName(ctx, tenantID, eventName)
	if err != nil {
		t.Fatalf("Failed to check existence: %v", err)
	}
	if !exists {
		t.Error("Event should exist now")
	}
}
