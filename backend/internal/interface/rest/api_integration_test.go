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
	createTestTenant(t, pool, tenantID)

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
	router, pool, cleanup := setupTestRouter(t)
	defer cleanup()

	// テスト用のテナントIDを生成
	tenantID := common.NewTenantID()
	createTestTenant(t, pool, tenantID)

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
				EventID   string `json:"event_id"`
				EventName string `json:"event_name"`
				EventType string `json:"event_type"`
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

// =====================================================
// Authentication Tests
// =====================================================

// createTestAdmin はテスト用の管理者を作成します
func createTestAdmin(t *testing.T, pool *pgxpool.Pool, tenantID common.TenantID, email, passwordHash string) common.AdminID {
	t.Helper()
	ctx := context.Background()
	adminID := common.NewAdminID()
	query := `
		INSERT INTO admins (admin_id, tenant_id, email, password_hash, display_name, role, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
	`
	_, err := pool.Exec(ctx, query, adminID.String(), tenantID.String(), email, passwordHash, "Test Admin", "owner", true)
	if err != nil {
		t.Fatalf("Failed to create test admin: %v", err)
	}
	return adminID
}

// TestLogin_Success tests successful login
func TestLogin_Success(t *testing.T) {
	router, pool, cleanup := setupTestRouter(t)
	defer cleanup()

	// テスト用のテナントと管理者を作成
	tenantID := common.NewTenantID()
	createTestTenant(t, pool, tenantID)
	// bcrypt hash of "password123"
	passwordHash := "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZRGdjGj/n3.E8y6iYAXxIz8p7qmOy"
	createTestAdmin(t, pool, tenantID, "login-test@example.com", passwordHash)

	// リクエストボディの作成
	requestBody := map[string]string{
		"email":    "login-test@example.com",
		"password": "password123",
	}
	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		t.Fatalf("Failed to marshal request body: %v", err)
	}

	// リクエストの作成
	req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// リクエストの実行
	router.ServeHTTP(w, req)

	// ステータスコードの確認
	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d. Body: %s", http.StatusOK, w.Code, w.Body.String())
	}

	// レスポンスボディの確認
	var responseWrapper struct {
		Data struct {
			Token     string `json:"token"`
			AdminID   string `json:"admin_id"`
			TenantID  string `json:"tenant_id"`
			Email     string `json:"email"`
			Role      string `json:"role"`
			ExpiresAt string `json:"expires_at"`
		} `json:"data"`
	}
	if err := json.NewDecoder(w.Body).Decode(&responseWrapper); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	response := responseWrapper.Data

	if response.Token == "" {
		t.Error("Expected token to be set")
	}
	if response.TenantID != tenantID.String() {
		t.Errorf("Expected TenantID '%s', got '%s'", tenantID.String(), response.TenantID)
	}
	if response.Email != "login-test@example.com" {
		t.Errorf("Expected Email 'login-test@example.com', got '%s'", response.Email)
	}
}

// TestLogin_InvalidCredentials tests login with wrong password
func TestLogin_InvalidCredentials(t *testing.T) {
	router, pool, cleanup := setupTestRouter(t)
	defer cleanup()

	// テスト用のテナントと管理者を作成
	tenantID := common.NewTenantID()
	createTestTenant(t, pool, tenantID)
	passwordHash := "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZRGdjGj/n3.E8y6iYAXxIz8p7qmOy"
	createTestAdmin(t, pool, tenantID, "invalid-login@example.com", passwordHash)

	// 間違ったパスワードでログイン
	requestBody := map[string]string{
		"email":    "invalid-login@example.com",
		"password": "wrongpassword",
	}
	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		t.Fatalf("Failed to marshal request body: %v", err)
	}

	req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// 401 Unauthorized を期待
	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status code %d, got %d. Body: %s", http.StatusUnauthorized, w.Code, w.Body.String())
	}
}

// TestLogin_EmailNotFound tests login with non-existent email
func TestLogin_EmailNotFound(t *testing.T) {
	router, _, cleanup := setupTestRouter(t)
	defer cleanup()

	requestBody := map[string]string{
		"email":    "nonexistent@example.com",
		"password": "password123",
	}
	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		t.Fatalf("Failed to marshal request body: %v", err)
	}

	req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// 401 Unauthorized を期待（セキュリティのため存在しないユーザーも同じエラー）
	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status code %d, got %d. Body: %s", http.StatusUnauthorized, w.Code, w.Body.String())
	}
}

// =====================================================
// Error Handling Tests
// =====================================================

// TestCreateEvent_ValidationError tests event creation with invalid data
func TestCreateEvent_ValidationError(t *testing.T) {
	router, pool, cleanup := setupTestRouter(t)
	defer cleanup()

	tenantID := common.NewTenantID()
	createTestTenant(t, pool, tenantID)

	// 空のイベント名でリクエスト
	requestBody := map[string]string{
		"event_name":  "", // 空の名前
		"event_type":  "normal",
		"description": "Test Description",
	}
	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		t.Fatalf("Failed to marshal request body: %v", err)
	}

	req := httptest.NewRequest("POST", "/api/v1/events", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tenant-ID", tenantID.String())
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// 400 Bad Request を期待
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d. Body: %s", http.StatusBadRequest, w.Code, w.Body.String())
	}

	// エラーレスポンスの確認
	var responseWrapper struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.NewDecoder(w.Body).Decode(&responseWrapper); err != nil {
		t.Fatalf("Failed to decode error response: %v", err)
	}

	if responseWrapper.Error.Code == "" {
		t.Error("Expected error code to be set")
	}
}

// TestCreateEvent_InvalidEventType tests event creation with invalid event type
func TestCreateEvent_InvalidEventType(t *testing.T) {
	router, pool, cleanup := setupTestRouter(t)
	defer cleanup()

	tenantID := common.NewTenantID()
	createTestTenant(t, pool, tenantID)

	requestBody := map[string]string{
		"event_name":  "Test Event",
		"event_type":  "invalid_type", // 無効なイベントタイプ
		"description": "Test Description",
	}
	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		t.Fatalf("Failed to marshal request body: %v", err)
	}

	req := httptest.NewRequest("POST", "/api/v1/events", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tenant-ID", tenantID.String())
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// 400 Bad Request を期待
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d. Body: %s", http.StatusBadRequest, w.Code, w.Body.String())
	}
}

// TestMissingTenantID tests request without tenant ID
func TestMissingTenantID(t *testing.T) {
	router, _, cleanup := setupTestRouter(t)
	defer cleanup()

	// X-Tenant-ID ヘッダーなしでリクエスト
	req := httptest.NewRequest("GET", "/api/v1/events", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// 400 または 401 を期待（認証前なので認可エラー）
	if w.Code != http.StatusBadRequest && w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status code 400 or 401, got %d. Body: %s", w.Code, w.Body.String())
	}
}

// TestInvalidJSON tests request with invalid JSON body
func TestInvalidJSON(t *testing.T) {
	router, pool, cleanup := setupTestRouter(t)
	defer cleanup()

	tenantID := common.NewTenantID()
	createTestTenant(t, pool, tenantID)

	// 不正なJSONでリクエスト
	req := httptest.NewRequest("POST", "/api/v1/events", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tenant-ID", tenantID.String())
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// 400 Bad Request を期待
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d. Body: %s", http.StatusBadRequest, w.Code, w.Body.String())
	}
}

// =====================================================
// Not Found Tests
// =====================================================

// TestGetEvent_NotFound tests getting a non-existent event
func TestGetEvent_NotFound(t *testing.T) {
	router, pool, cleanup := setupTestRouter(t)
	defer cleanup()

	tenantID := common.NewTenantID()
	createTestTenant(t, pool, tenantID)

	// 存在しないイベントIDでリクエスト
	nonExistentID := common.NewULID()
	req := httptest.NewRequest("GET", "/api/v1/events/"+nonExistentID, nil)
	req.Header.Set("X-Tenant-ID", tenantID.String())
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// 404 Not Found を期待
	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status code %d, got %d. Body: %s", http.StatusNotFound, w.Code, w.Body.String())
	}
}
