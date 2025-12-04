# バックエンド統合テスト方針

## 概要

このドキュメントは、Go の `httptest` を使った「REST API 統合テスト」のテストケース設計と、実際のテストコードのひな型をまとめたものです。

---

## 1. 既存のテスト状況

### 1.1 テストファイルの存在確認

**確認結果**:

- **ドメインレイヤー**: テストあり
  - `backend/internal/domain/event/event_test.go` - Event ドメインのユニットテスト
  - `backend/internal/domain/shift/shift_slot_test.go` - ShiftSlot ドメインのユニットテスト

- **インフラレイヤー**: 統合テストあり
  - `backend/internal/infra/db/event_repository_integration_test.go` - EventRepository の統合テスト（実際のDB接続が必要）

- **インターフェースレイヤー（REST API）**: **テストなし**
  - `backend/internal/interface/rest/` 配下に `_test.go` ファイルが存在しない

### 1.2 テスト実行方法

**ドメインテスト**:
```bash
cd backend
go test ./internal/domain/...
```

**統合テスト**:
```bash
cd backend
DATABASE_URL=postgres://user:pass@localhost:5432/testdb go test -v ./internal/infra/db/
```

**REST API統合テスト**: 未実装

---

## 2. REST API 統合テストのテストケース設計

### 2.1 テスト対象エンドポイント

以下のエンドポイントについて、最低限「正常系」のテストケースを設計します。

1. **Event API**
   - `POST /api/v1/events` - イベント作成
   - `GET /api/v1/events` - イベント一覧取得

2. **BusinessDay API**
   - `POST /api/v1/events/:event_id/business-days` - 営業日作成
   - `GET /api/v1/events/:event_id/business-days` - 営業日一覧取得

3. **ShiftSlot API**
   - `POST /api/v1/business-days/:business_day_id/shift-slots` - シフト枠作成
   - `GET /api/v1/business-days/:business_day_id/shift-slots` - シフト枠一覧取得

4. **ShiftAssignment API**
   - `POST /api/v1/shift-assignments` - シフト割り当て確定
   - `GET /api/v1/shift-assignments?member_id=...` - シフト割り当て一覧取得

### 2.2 テストケース一覧表

#### Event API

| エンドポイント | 前提データ | 入力 | 期待ステータス | 期待されるレスポンスの主なフィールド |
|----------------|-----------|------|----------------|--------------------------------------|
| POST /api/v1/events | tenantが存在する（ヘッダーで指定） | { event_name: "Test Event", event_type: "normal", description: "Test Description" } | 201 Created | event_id, event_name, event_type, description, is_active, created_at |
| GET /api/v1/events | 上記イベントが存在 | - | 200 OK | events 配列（length >= 1）、先ほど作成したイベントを含む |

#### BusinessDay API

| エンドポイント | 前提データ | 入力 | 期待ステータス | 期待されるレスポンスの主なフィールド |
|----------------|-----------|------|----------------|--------------------------------------|
| POST /api/v1/events/:event_id/business-days | 上記イベントが存在 | { target_date: "2025-01-15", start_time: "21:30", end_time: "23:00", occurrence_type: "special" } | 201 Created | business_day_id, event_id, target_date, start_time, end_time, occurrence_type |
| GET /api/v1/events/:event_id/business-days | 上記営業日が存在 | - | 200 OK | business_days 配列（length >= 1）、先ほど作成した営業日を含む |

#### ShiftSlot API

| エンドポイント | 前提データ | 入力 | 期待ステータス | 期待されるレスポンスの主なフィールド |
|----------------|-----------|------|----------------|--------------------------------------|
| POST /api/v1/business-days/:business_day_id/shift-slots | 上記営業日が存在、positionが存在 | { position_id: "...", slot_name: "早番スタッフ", instance_name: "Instance A", start_time: "21:30:00", end_time: "23:00:00", required_count: 2, priority: 1 } | 201 Created | slot_id, business_day_id, position_id, slot_name, instance_name, start_time, end_time, required_count |
| GET /api/v1/business-days/:business_day_id/shift-slots | 上記シフト枠が存在 | - | 200 OK | shift_slots 配列（length >= 1）、先ほど作成したシフト枠を含む |

#### ShiftAssignment API

| エンドポイント | 前提データ | 入力 | 期待ステータス | 期待されるレスポンスの主なフィールド |
|----------------|-----------|------|----------------|--------------------------------------|
| POST /api/v1/shift-assignments | 上記シフト枠が存在、memberが存在 | { slot_id: "...", member_id: "...", note: "Test note" } | 201 Created | assignment_id, slot_id, member_id, assignment_status: "confirmed", assignment_method: "manual" |
| GET /api/v1/shift-assignments?member_id=... | 上記割り当てが存在 | - | 200 OK | assignments 配列（length >= 1）、先ほど作成した割り当てを含む |

### 2.3 テストデータの準備

**前提条件**:
- テスト用のテナントID（ULID形式）
- テスト用のメンバーID（ULID形式）
- テスト用のポジションID（ULID形式、シフト枠作成時に必要）

**注意**: 実際のDB接続が必要なため、テスト用のデータベースを用意する必要がある。

---

## 3. テストコードのひな型

### 3.1 ファイル構成

**推奨ファイル名**: `backend/internal/interface/rest/api_integration_test.go`

### 3.2 テストコードのひな型

```go
package rest_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/interface/rest"
	"github.com/jackc/pgx/v5/pgxpool"
)

// setupTestRouter creates a test router with a test database connection
func setupTestRouter(t *testing.T) (http.Handler, *pgxpool.Pool, func()) {
	t.Helper()

	// テスト用のデータベース接続
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

	// ルーターの作成
	router := rest.NewRouter(pool)

	cleanup := func() {
		pool.Close()
	}

	return router, pool, cleanup
}

// TestHealthCheck tests the health check endpoint
func TestHealthCheck(t *testing.T) {
	router, _, cleanup := setupTestRouter(t)
	defer cleanup()

	// リクエストの作成
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	// リクエストの実行
	router.ServeHTTP(w, req)

	// ステータスコードの確認
	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	// レスポンスボディの確認
	var response map[string]string
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response["status"] != "ok" {
		t.Errorf("Expected status 'ok', got '%s'", response["status"])
	}
}

// TestCreateEvent tests POST /api/v1/events
func TestCreateEvent(t *testing.T) {
	router, pool, cleanup := setupTestRouter(t)
	defer cleanup()

	// テスト用のテナントIDを生成
	tenantID := common.NewTenantID()

	// リクエストボディの作成
	requestBody := map[string]string{
		"event_name":   "Test Event",
		"event_type":   "normal",
		"description":  "Test Description",
	}
	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		t.Fatalf("Failed to marshal request body: %v", err)
	}

	// リクエストの作成
	req := httptest.NewRequest("POST", "/api/v1/events", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tenant-ID", tenantID.String())
	w := httptest.NewRecorder()

	// リクエストの実行
	router.ServeHTTP(w, req)

	// ステータスコードの確認
	if w.Code != http.StatusCreated {
		t.Errorf("Expected status code %d, got %d. Body: %s", http.StatusCreated, w.Code, w.Body.String())
	}

	// レスポンスボディの確認
	var response struct {
		EventID     string `json:"event_id"`
		TenantID    string `json:"tenant_id"`
		EventName   string `json:"event_name"`
		EventType   string `json:"event_type"`
		Description string `json:"description"`
		IsActive    bool   `json:"is_active"`
		CreatedAt   string `json:"created_at"`
		UpdatedAt   string `json:"updated_at"`
	}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// フィールドの検証
	if response.EventName != "Test Event" {
		t.Errorf("Expected EventName 'Test Event', got '%s'", response.EventName)
	}
	if response.EventType != "normal" {
		t.Errorf("Expected EventType 'normal', got '%s'", response.EventType)
	}
	if response.TenantID != tenantID.String() {
		t.Errorf("Expected TenantID '%s', got '%s'", tenantID.String(), response.TenantID)
	}
	if response.EventID == "" {
		t.Error("Expected EventID to be set, got empty string")
	}
}

// TestListEvents tests GET /api/v1/events
func TestListEvents(t *testing.T) {
	router, pool, cleanup := setupTestRouter(t)
	defer cleanup()

	// テスト用のテナントIDを生成
	tenantID := common.NewTenantID()

	// まずイベントを作成（TestCreateEvent と同じロジック）
	// ここでは簡略化のため、直接DBにデータを挿入するか、TestCreateEvent を呼び出す

	// リクエストの作成
	req := httptest.NewRequest("GET", "/api/v1/events", nil)
	req.Header.Set("X-Tenant-ID", tenantID.String())
	w := httptest.NewRecorder()

	// リクエストの実行
	router.ServeHTTP(w, req)

	// ステータスコードの確認
	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d. Body: %s", http.StatusOK, w.Code, w.Body.String())
	}

	// レスポンスボディの確認
	var response struct {
		Events []struct {
			EventID     string `json:"event_id"`
			EventName   string `json:"event_name"`
			EventType   string `json:"event_type"`
		} `json:"events"`
		Count int `json:"count"`
	}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// 最低でも0件以上であることを確認
	if response.Count < 0 {
		t.Errorf("Expected count >= 0, got %d", response.Count)
	}
	if len(response.Events) != response.Count {
		t.Errorf("Expected events length %d, got %d", response.Count, len(response.Events))
	}
}

// TestCreateEvent_InvalidRequest tests POST /api/v1/events with invalid request
func TestCreateEvent_InvalidRequest(t *testing.T) {
	router, _, cleanup := setupTestRouter(t)
	defer cleanup()

	tenantID := common.NewTenantID()

	// リクエストボディ（event_name が空）
	requestBody := map[string]string{
		"event_name":   "",  // 空文字
		"event_type":   "normal",
		"description":  "Test Description",
	}
	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		t.Fatalf("Failed to marshal request body: %v", err)
	}

	// リクエストの作成
	req := httptest.NewRequest("POST", "/api/v1/events", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tenant-ID", tenantID.String())
	w := httptest.NewRecorder()

	// リクエストの実行
	router.ServeHTTP(w, req)

	// ステータスコードの確認（400 Bad Request が期待される）
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d. Body: %s", http.StatusBadRequest, w.Code, w.Body.String())
	}
}

// TestCreateEvent_MissingTenantID tests POST /api/v1/events without X-Tenant-ID header
func TestCreateEvent_MissingTenantID(t *testing.T) {
	router, _, cleanup := setupTestRouter(t)
	defer cleanup()

	// リクエストボディ
	requestBody := map[string]string{
		"event_name":   "Test Event",
		"event_type":   "normal",
		"description":  "Test Description",
	}
	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		t.Fatalf("Failed to marshal request body: %v", err)
	}

	// リクエストの作成（X-Tenant-ID ヘッダーなし）
	req := httptest.NewRequest("POST", "/api/v1/events", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// リクエストの実行
	router.ServeHTTP(w, req)

	// ステータスコードの確認（400 Bad Request が期待される）
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d. Body: %s", http.StatusBadRequest, w.Code, w.Body.String())
	}
}
```

### 3.3 テスト実行方法

**前提条件**:
- テスト用のPostgreSQLデータベースが起動している
- `DATABASE_URL` 環境変数が設定されている

**実行コマンド**:
```bash
cd backend
DATABASE_URL=postgres://user:pass@localhost:5432/testdb go test -v ./internal/interface/rest/
```

**注意**: テスト用のデータベースは、テスト実行前にマイグレーションを実行しておく必要がある。

---

## 4. 追加のテストケース（将来実装）

### 4.1 異常系テスト

- **バリデーションエラー**: 必須フィールドが空、型が不正など
- **認証エラー**: `X-Tenant-ID` ヘッダーがない、不正な形式など
- **権限エラー**: 他テナントのデータにアクセスしようとした場合（現状は実装されていない）
- **リソース不存在**: 存在しないIDを指定した場合（404 Not Found）

### 4.2 エッジケーステスト

- **深夜営業**: `start_time > end_time` の場合（例: 22:00 - 02:00）
- **満員チェック**: `required_count` を超えた割り当てを試みた場合（409 Conflict）
- **重複チェック**: 同じ名前のイベントを重複作成しようとした場合（409 Conflict）

### 4.3 パフォーマンステスト

- **大量データ**: 1000件以上のイベントを取得する場合のレスポンス時間
- **同時リクエスト**: 複数のリクエストが同時に来た場合の動作

---

## 5. まとめ

### 5.1 現状

- **ドメインレイヤー**: ユニットテストあり
- **インフラレイヤー**: 統合テストあり
- **インターフェースレイヤー（REST API）**: **テストなし**

### 5.2 推奨される実装順序

1. **Phase 1**: ヘルスチェックエンドポイントのテスト（`TestHealthCheck`）
2. **Phase 2**: Event API の正常系テスト（`TestCreateEvent`, `TestListEvents`）
3. **Phase 3**: Event API の異常系テスト（`TestCreateEvent_InvalidRequest`, `TestCreateEvent_MissingTenantID`）
4. **Phase 4**: 他のエンドポイント（BusinessDay, ShiftSlot, ShiftAssignment）のテスト

### 5.3 テスト環境の整備

- **テスト用データベース**: 開発環境とは別のデータベースを使用
- **データクリーンアップ**: テスト実行後にデータをクリーンアップする仕組み（トランザクションのロールバックなど）
- **CI/CD統合**: GitHub Actions などで自動実行

---

**作成日**: 2025-01-XX  
**作成者**: 検証専用アシスタント（Auto）

