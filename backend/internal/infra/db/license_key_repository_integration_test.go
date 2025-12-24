package db_test

import (
	"context"
	"testing"
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/billing"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/infra/db"
)

// =====================================================
// LicenseKeyRepository Integration Tests
// =====================================================

func TestLicenseKeyRepository_SaveAndFindByID(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	repo := db.NewLicenseKeyRepository(pool)
	ctx := context.Background()

	// テスト用のライセンスキーを作成
	now := time.Now()
	keyHash := billing.HashLicenseKey("ABCD1234EF567890")
	key, err := billing.NewLicenseKey(now, keyHash, nil, "Test memo")
	if err != nil {
		t.Fatalf("Failed to create test license key: %v", err)
	}

	// Save
	err = repo.Save(ctx, key)
	if err != nil {
		t.Fatalf("Failed to save license key: %v", err)
	}

	// FindByID
	foundKey, err := repo.FindByID(ctx, key.KeyID())
	if err != nil {
		t.Fatalf("Failed to find license key by ID: %v", err)
	}

	if foundKey == nil {
		t.Fatal("License key should be found")
	}

	// 値の検証
	if foundKey.KeyID() != key.KeyID() {
		t.Errorf("KeyID mismatch: got %v, want %v", foundKey.KeyID(), key.KeyID())
	}

	if foundKey.KeyHash() != key.KeyHash() {
		t.Errorf("KeyHash mismatch: got %v, want %v", foundKey.KeyHash(), key.KeyHash())
	}

	if foundKey.Status() != key.Status() {
		t.Errorf("Status mismatch: got %v, want %v", foundKey.Status(), key.Status())
	}

	if foundKey.Memo() != key.Memo() {
		t.Errorf("Memo mismatch: got %v, want %v", foundKey.Memo(), key.Memo())
	}

	if !foundKey.IsUnused() {
		t.Error("Key should be unused")
	}
}

func TestLicenseKeyRepository_FindByHashForUpdate(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	repo := db.NewLicenseKeyRepository(pool)
	ctx := context.Background()

	now := time.Now()
	keyHash := billing.HashLicenseKey("HASH1234TEST5678")
	key, err := billing.NewLicenseKey(now, keyHash, nil, "")
	if err != nil {
		t.Fatalf("Failed to create test license key: %v", err)
	}

	err = repo.Save(ctx, key)
	if err != nil {
		t.Fatalf("Failed to save license key: %v", err)
	}

	// FindByHashForUpdate
	foundKey, err := repo.FindByHashForUpdate(ctx, keyHash)
	if err != nil {
		t.Fatalf("Failed to find license key by hash: %v", err)
	}

	if foundKey == nil {
		t.Fatal("License key should be found")
	}

	if foundKey.KeyHash() != keyHash {
		t.Errorf("KeyHash mismatch: got %v, want %v", foundKey.KeyHash(), keyHash)
	}

	// 存在しないハッシュで検索
	notFoundKey, err := repo.FindByHashForUpdate(ctx, "nonexistenthash")
	if err != nil {
		t.Fatalf("Should not return error for not found: %v", err)
	}
	if notFoundKey != nil {
		t.Error("Should return nil for not found key")
	}
}

func TestLicenseKeyRepository_MarkAsUsed(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	repo := db.NewLicenseKeyRepository(pool)
	ctx := context.Background()

	now := time.Now()
	keyHash := billing.HashLicenseKey("MARK1234USED5678")
	key, err := billing.NewLicenseKey(now, keyHash, nil, "")
	if err != nil {
		t.Fatalf("Failed to create test license key: %v", err)
	}

	err = repo.Save(ctx, key)
	if err != nil {
		t.Fatalf("Failed to save license key: %v", err)
	}

	// テナントを作成
	tenantID := common.NewTenantID()
	createTestTenant(t, pool, tenantID)

	// MarkAsUsed
	useTime := time.Now()
	err = key.MarkAsUsed(useTime, tenantID)
	if err != nil {
		t.Fatalf("Failed to mark as used: %v", err)
	}

	// 更新を保存
	err = repo.Save(ctx, key)
	if err != nil {
		t.Fatalf("Failed to save updated license key: %v", err)
	}

	// 再取得して確認
	foundKey, err := repo.FindByID(ctx, key.KeyID())
	if err != nil {
		t.Fatalf("Failed to find license key: %v", err)
	}

	if !foundKey.IsUsed() {
		t.Error("Key should be marked as used")
	}

	if foundKey.UsedTenantID() == nil {
		t.Fatal("UsedTenantID should not be nil")
	}

	if *foundKey.UsedTenantID() != tenantID {
		t.Errorf("UsedTenantID mismatch: got %v, want %v", *foundKey.UsedTenantID(), tenantID)
	}

	if foundKey.UsedAt() == nil {
		t.Error("UsedAt should not be nil")
	}
}

func TestLicenseKeyRepository_Revoke(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	repo := db.NewLicenseKeyRepository(pool)
	ctx := context.Background()

	now := time.Now()
	keyHash := billing.HashLicenseKey("REVO1234TEST5678")
	key, err := billing.NewLicenseKey(now, keyHash, nil, "")
	if err != nil {
		t.Fatalf("Failed to create test license key: %v", err)
	}

	err = repo.Save(ctx, key)
	if err != nil {
		t.Fatalf("Failed to save license key: %v", err)
	}

	// Revoke
	revokeTime := time.Now()
	err = key.Revoke(revokeTime)
	if err != nil {
		t.Fatalf("Failed to revoke: %v", err)
	}

	// 更新を保存
	err = repo.Save(ctx, key)
	if err != nil {
		t.Fatalf("Failed to save revoked license key: %v", err)
	}

	// 再取得して確認
	foundKey, err := repo.FindByID(ctx, key.KeyID())
	if err != nil {
		t.Fatalf("Failed to find license key: %v", err)
	}

	if !foundKey.IsRevoked() {
		t.Error("Key should be revoked")
	}

	if foundKey.RevokedAt() == nil {
		t.Error("RevokedAt should not be nil")
	}
}

func TestLicenseKeyRepository_SaveBatch(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	repo := db.NewLicenseKeyRepository(pool)
	ctx := context.Background()

	now := time.Now()
	// Note: batchID would be used in a real scenario with FindByBatchID
	// For this test, we just create individual keys

	// 5つのキーを作成
	keys := make([]*billing.LicenseKey, 5)
	for i := 0; i < 5; i++ {
		keyStr := "BATC" + string(rune('0'+i)) + "123456789AB"
		keyHash := billing.HashLicenseKey(keyStr)
		key, err := billing.NewLicenseKey(now, keyHash, nil, "Batch test")
		if err != nil {
			t.Fatalf("Failed to create test license key: %v", err)
		}
		keys[i] = key
	}

	// SaveBatch
	err := repo.SaveBatch(ctx, keys)
	if err != nil {
		t.Fatalf("Failed to save batch: %v", err)
	}

	// 各キーが保存されていることを確認
	for _, k := range keys {
		foundKey, err := repo.FindByID(ctx, k.KeyID())
		if err != nil {
			t.Fatalf("Failed to find key %s: %v", k.KeyID(), err)
		}
		if foundKey == nil {
			t.Errorf("Key %s should be found", k.KeyID())
		}
	}
}

func TestLicenseKeyRepository_CountByStatus(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	repo := db.NewLicenseKeyRepository(pool)
	ctx := context.Background()

	// 初期カウントを取得
	initialCount, err := repo.CountByStatus(ctx, billing.LicenseKeyStatusUnused)
	if err != nil {
		t.Fatalf("Failed to count: %v", err)
	}

	// 新しいキーを追加
	now := time.Now()
	keyHash := billing.HashLicenseKey("COUN1234TEST5678")
	key, err := billing.NewLicenseKey(now, keyHash, nil, "")
	if err != nil {
		t.Fatalf("Failed to create test license key: %v", err)
	}

	err = repo.Save(ctx, key)
	if err != nil {
		t.Fatalf("Failed to save license key: %v", err)
	}

	// カウントが増えたことを確認
	newCount, err := repo.CountByStatus(ctx, billing.LicenseKeyStatusUnused)
	if err != nil {
		t.Fatalf("Failed to count: %v", err)
	}

	if newCount != initialCount+1 {
		t.Errorf("Count should have increased by 1: got %d, want %d", newCount, initialCount+1)
	}
}

func TestLicenseKeyRepository_List(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	repo := db.NewLicenseKeyRepository(pool)
	ctx := context.Background()

	// テスト用のキーを追加
	now := time.Now()
	for i := 0; i < 5; i++ {
		keyStr := "LIST" + string(rune('A'+i)) + "23456789AB"
		keyHash := billing.HashLicenseKey(keyStr)
		key, err := billing.NewLicenseKey(now, keyHash, nil, "List test")
		if err != nil {
			t.Fatalf("Failed to create test license key: %v", err)
		}
		err = repo.Save(ctx, key)
		if err != nil {
			t.Fatalf("Failed to save license key: %v", err)
		}
	}

	// List（全件）
	keys, total, err := repo.List(ctx, nil, 10, 0)
	if err != nil {
		t.Fatalf("Failed to list license keys: %v", err)
	}

	if len(keys) == 0 {
		t.Error("Should return at least some keys")
	}

	if total < len(keys) {
		t.Errorf("Total count should be >= returned keys count: total=%d, returned=%d", total, len(keys))
	}

	// List（ステータスフィルター付き）
	unusedStatus := billing.LicenseKeyStatusUnused
	unusedKeys, unusedTotal, err := repo.List(ctx, &unusedStatus, 10, 0)
	if err != nil {
		t.Fatalf("Failed to list unused license keys: %v", err)
	}

	// 全ての結果がunusedであることを確認
	for _, k := range unusedKeys {
		if k.Status() != billing.LicenseKeyStatusUnused {
			t.Errorf("Key should be unused: got %v", k.Status())
		}
	}

	if unusedTotal > total {
		t.Errorf("Unused total should be <= total: unused=%d, total=%d", unusedTotal, total)
	}
}

func TestLicenseKeyRepository_ListWithPagination(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	repo := db.NewLicenseKeyRepository(pool)
	ctx := context.Background()

	// テスト前にテーブルをクリア
	_, err := pool.Exec(ctx, "DELETE FROM license_keys")
	if err != nil {
		t.Fatalf("Failed to clean up license_keys table: %v", err)
	}

	// テスト用のキーを追加
	now := time.Now()
	for i := 0; i < 10; i++ {
		keyStr := "PAGE" + string(rune('A'+i)) + "23456789AB"
		keyHash := billing.HashLicenseKey(keyStr)
		key, err := billing.NewLicenseKey(now, keyHash, nil, "Pagination test")
		if err != nil {
			t.Fatalf("Failed to create test license key: %v", err)
		}
		err = repo.Save(ctx, key)
		if err != nil {
			t.Fatalf("Failed to save license key: %v", err)
		}
	}

	// 最初のページ
	firstPage, total, err := repo.List(ctx, nil, 3, 0)
	if err != nil {
		t.Fatalf("Failed to list first page: %v", err)
	}

	if len(firstPage) > 3 {
		t.Errorf("First page should have at most 3 items: got %d", len(firstPage))
	}

	// 2ページ目
	secondPage, _, err := repo.List(ctx, nil, 3, 3)
	if err != nil {
		t.Fatalf("Failed to list second page: %v", err)
	}

	// 結果が重複していないことを確認（IDで比較）
	firstPageIDs := make(map[string]bool)
	for _, k := range firstPage {
		firstPageIDs[k.KeyID().String()] = true
	}

	for _, k := range secondPage {
		if firstPageIDs[k.KeyID().String()] {
			t.Errorf("Key %s appears in both pages", k.KeyID())
		}
	}

	t.Logf("Total: %d, First page: %d items, Second page: %d items", total, len(firstPage), len(secondPage))
}

func TestLicenseKeyRepository_WithExpiration(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	repo := db.NewLicenseKeyRepository(pool)
	ctx := context.Background()

	now := time.Now()
	expiresAt := now.Add(30 * 24 * time.Hour) // 30日後に期限切れ
	keyHash := billing.HashLicenseKey("EXPI1234TEST5678")
	key, err := billing.NewLicenseKey(now, keyHash, &expiresAt, "Expiration test")
	if err != nil {
		t.Fatalf("Failed to create test license key: %v", err)
	}

	err = repo.Save(ctx, key)
	if err != nil {
		t.Fatalf("Failed to save license key: %v", err)
	}

	// 再取得して確認
	foundKey, err := repo.FindByID(ctx, key.KeyID())
	if err != nil {
		t.Fatalf("Failed to find license key: %v", err)
	}

	if foundKey.ExpiresAt() == nil {
		t.Fatal("ExpiresAt should not be nil")
	}

	// 時間の比較（秒単位まで）
	if foundKey.ExpiresAt().Truncate(time.Second).Unix() != expiresAt.Truncate(time.Second).Unix() {
		t.Errorf("ExpiresAt mismatch: got %v, want %v", foundKey.ExpiresAt(), expiresAt)
	}
}
