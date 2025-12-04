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
// DATABASE_URL 環境変数が必要です。
// フォーマット: postgres://user:password@host:port/database?sslmode=disable
// 例: postgres://vrcshift:vrcshift@localhost:5432/vrcshift_test?sslmode=disable
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
	router, _, cleanup := setupTestRouter(t)
	defer cleanup()

	// テスト用のテナントIDを生成
	tenantID := common.NewTenantID()

	// リクエストボディの作成
	requestBody := map[string]string{
		"event_name":  "Test Event",
		"event_type":  "normal",
		"description": "Test Description",
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

	// レスポンスボディの確認（SuccessResponse でラップされている）
	var responseWrapper struct {
		Data struct {
			EventID     string `json:"event_id"`
			TenantID    string `json:"tenant_id"`
			EventName   string `json:"event_name"`
			EventType   string `json:"event_type"`
			Description string `json:"description"`
			IsActive    bool   `json:"is_active"`
			CreatedAt   string `json:"created_at"`
			UpdatedAt   string `json:"updated_at"`
		} `json:"data"`
	}
	if err := json.NewDecoder(w.Body).Decode(&responseWrapper); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	response := responseWrapper.Data

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
	if !response.IsActive {
		t.Error("Expected IsActive to be true, got false")
	}
}

// TestListEvents tests GET /api/v1/events
func TestListEvents(t *testing.T) {
	router, _, cleanup := setupTestRouter(t)
	defer cleanup()

	// テスト用のテナントIDを生成
	tenantID := common.NewTenantID()

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

	// レスポンスボディの確認（SuccessResponse でラップされている）
	var responseWrapper struct {
		Data struct {
			Events []struct {
				EventID     string `json:"event_id"`
				EventName   string `json:"event_name"`
				EventType   string `json:"event_type"`
			} `json:"events"`
			Count int `json:"count"`
		} `json:"data"`
	}
	if err := json.NewDecoder(w.Body).Decode(&responseWrapper); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	response := responseWrapper.Data

	// 最低でも0件以上であることを確認
	if response.Count < 0 {
		t.Errorf("Expected count >= 0, got %d", response.Count)
	}
	if len(response.Events) != response.Count {
		t.Errorf("Expected events length %d, got %d", response.Count, len(response.Events))
	}
}

