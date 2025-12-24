package db_test

import (
	"context"
	"testing"
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/auth"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/infra/db"
)

// =====================================================
// AdminRepository Integration Tests
// =====================================================

func TestAdminRepository_SaveAndFindByID(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	repo := db.NewAdminRepository(pool)
	ctx := context.Background()

	// テスト用のテナントを作成
	tenantID := common.NewTenantID()
	createTestTenant(t, pool, tenantID)

	// テスト用の管理者を作成
	now := time.Now()
	admin, err := auth.NewAdmin(now, tenantID, "test@example.com", "$2a$10$hash", "Test Admin", auth.RoleOwner)
	if err != nil {
		t.Fatalf("Failed to create test admin: %v", err)
	}

	// Save
	err = repo.Save(ctx, admin)
	if err != nil {
		t.Fatalf("Failed to save admin: %v", err)
	}

	// FindByID
	foundAdmin, err := repo.FindByID(ctx, admin.AdminID())
	if err != nil {
		t.Fatalf("Failed to find admin by ID: %v", err)
	}

	// 値の検証
	if foundAdmin.AdminID() != admin.AdminID() {
		t.Errorf("AdminID mismatch: got %v, want %v", foundAdmin.AdminID(), admin.AdminID())
	}

	if foundAdmin.TenantID() != admin.TenantID() {
		t.Errorf("TenantID mismatch: got %v, want %v", foundAdmin.TenantID(), admin.TenantID())
	}

	if foundAdmin.Email() != admin.Email() {
		t.Errorf("Email mismatch: got %v, want %v", foundAdmin.Email(), admin.Email())
	}

	if foundAdmin.DisplayName() != admin.DisplayName() {
		t.Errorf("DisplayName mismatch: got %v, want %v", foundAdmin.DisplayName(), admin.DisplayName())
	}

	if foundAdmin.Role() != admin.Role() {
		t.Errorf("Role mismatch: got %v, want %v", foundAdmin.Role(), admin.Role())
	}

	if foundAdmin.IsActive() != admin.IsActive() {
		t.Errorf("IsActive mismatch: got %v, want %v", foundAdmin.IsActive(), admin.IsActive())
	}
}

func TestAdminRepository_FindByIDWithTenant(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	repo := db.NewAdminRepository(pool)
	ctx := context.Background()

	tenantID := common.NewTenantID()
	createTestTenant(t, pool, tenantID)

	now := time.Now()
	admin, err := auth.NewAdmin(now, tenantID, "test-tenant@example.com", "$2a$10$hash", "Test Admin", auth.RoleOwner)
	if err != nil {
		t.Fatalf("Failed to create test admin: %v", err)
	}

	err = repo.Save(ctx, admin)
	if err != nil {
		t.Fatalf("Failed to save admin: %v", err)
	}

	// 正しいテナントIDで検索
	foundAdmin, err := repo.FindByIDWithTenant(ctx, tenantID, admin.AdminID())
	if err != nil {
		t.Fatalf("Failed to find admin with tenant: %v", err)
	}

	if foundAdmin.AdminID() != admin.AdminID() {
		t.Errorf("AdminID mismatch: got %v, want %v", foundAdmin.AdminID(), admin.AdminID())
	}

	// 異なるテナントIDで検索（見つからないはず）
	differentTenantID := common.NewTenantID()
	createTestTenant(t, pool, differentTenantID)
	_, err = repo.FindByIDWithTenant(ctx, differentTenantID, admin.AdminID())
	if err == nil {
		t.Error("Expected error when finding admin with different tenant ID")
	}
}

func TestAdminRepository_FindByEmail(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	repo := db.NewAdminRepository(pool)
	ctx := context.Background()

	tenantID := common.NewTenantID()
	createTestTenant(t, pool, tenantID)

	email := "email-test@example.com"
	now := time.Now()
	admin, err := auth.NewAdmin(now, tenantID, email, "$2a$10$hash", "Email Test Admin", auth.RoleOwner)
	if err != nil {
		t.Fatalf("Failed to create test admin: %v", err)
	}

	err = repo.Save(ctx, admin)
	if err != nil {
		t.Fatalf("Failed to save admin: %v", err)
	}

	// メールアドレスで検索
	foundAdmin, err := repo.FindByEmail(ctx, tenantID, email)
	if err != nil {
		t.Fatalf("Failed to find admin by email: %v", err)
	}

	if foundAdmin.Email() != email {
		t.Errorf("Email mismatch: got %v, want %v", foundAdmin.Email(), email)
	}

	// 存在しないメールアドレスで検索
	_, err = repo.FindByEmail(ctx, tenantID, "nonexistent@example.com")
	if err == nil {
		t.Error("Expected error when finding admin with nonexistent email")
	}
}

func TestAdminRepository_FindByEmailGlobal(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	repo := db.NewAdminRepository(pool)
	ctx := context.Background()

	tenantID := common.NewTenantID()
	createTestTenant(t, pool, tenantID)

	email := "global-email-test@example.com"
	now := time.Now()
	admin, err := auth.NewAdmin(now, tenantID, email, "$2a$10$hash", "Global Test Admin", auth.RoleOwner)
	if err != nil {
		t.Fatalf("Failed to create test admin: %v", err)
	}

	err = repo.Save(ctx, admin)
	if err != nil {
		t.Fatalf("Failed to save admin: %v", err)
	}

	// グローバル検索（テナントIDなし）
	foundAdmin, err := repo.FindByEmailGlobal(ctx, email)
	if err != nil {
		t.Fatalf("Failed to find admin globally by email: %v", err)
	}

	if foundAdmin.Email() != email {
		t.Errorf("Email mismatch: got %v, want %v", foundAdmin.Email(), email)
	}

	if foundAdmin.TenantID() != tenantID {
		t.Errorf("TenantID mismatch: got %v, want %v", foundAdmin.TenantID(), tenantID)
	}
}

func TestAdminRepository_FindByTenantID(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	repo := db.NewAdminRepository(pool)
	ctx := context.Background()

	tenantID := common.NewTenantID()
	createTestTenant(t, pool, tenantID)

	// 3人の管理者を作成
	now := time.Now()
	for i := 0; i < 3; i++ {
		admin, err := auth.NewAdmin(now, tenantID,
			"tenant-admin-"+string(rune('0'+i))+"@example.com",
			"$2a$10$hash",
			"Admin "+string(rune('0'+i)),
			auth.RoleOwner)
		if err != nil {
			t.Fatalf("Failed to create test admin: %v", err)
		}
		err = repo.Save(ctx, admin)
		if err != nil {
			t.Fatalf("Failed to save admin: %v", err)
		}
	}

	// テナントIDで検索
	admins, err := repo.FindByTenantID(ctx, tenantID)
	if err != nil {
		t.Fatalf("Failed to find admins by tenant ID: %v", err)
	}

	if len(admins) < 3 {
		t.Errorf("Expected at least 3 admins, got %d", len(admins))
	}

	// 全ての結果が同じテナントIDを持つことを確認
	for _, a := range admins {
		if a.TenantID() != tenantID {
			t.Errorf("Admin has wrong tenant ID: got %v, want %v", a.TenantID(), tenantID)
		}
	}
}

func TestAdminRepository_FindActiveByTenantID(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	repo := db.NewAdminRepository(pool)
	ctx := context.Background()

	tenantID := common.NewTenantID()
	createTestTenant(t, pool, tenantID)

	now := time.Now()

	// アクティブな管理者を作成
	activeAdmin, err := auth.NewAdmin(now, tenantID, "active@example.com", "$2a$10$hash", "Active Admin", auth.RoleOwner)
	if err != nil {
		t.Fatalf("Failed to create active admin: %v", err)
	}
	err = repo.Save(ctx, activeAdmin)
	if err != nil {
		t.Fatalf("Failed to save active admin: %v", err)
	}

	// 非アクティブな管理者を作成
	inactiveAdmin, err := auth.NewAdmin(now, tenantID, "inactive@example.com", "$2a$10$hash", "Inactive Admin", auth.RoleManager)
	if err != nil {
		t.Fatalf("Failed to create inactive admin: %v", err)
	}
	inactiveAdmin.Deactivate(now)
	err = repo.Save(ctx, inactiveAdmin)
	if err != nil {
		t.Fatalf("Failed to save inactive admin: %v", err)
	}

	// アクティブな管理者のみを検索
	activeAdmins, err := repo.FindActiveByTenantID(ctx, tenantID)
	if err != nil {
		t.Fatalf("Failed to find active admins: %v", err)
	}

	// 結果にアクティブな管理者のみが含まれることを確認
	for _, a := range activeAdmins {
		if !a.IsActive() {
			t.Errorf("Found inactive admin in active admins result: %s", a.Email())
		}
	}

	// 少なくとも1人のアクティブな管理者が見つかること
	found := false
	for _, a := range activeAdmins {
		if a.Email() == "active@example.com" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected to find active@example.com in results")
	}
}

func TestAdminRepository_Update(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	repo := db.NewAdminRepository(pool)
	ctx := context.Background()

	tenantID := common.NewTenantID()
	createTestTenant(t, pool, tenantID)

	now := time.Now()
	admin, err := auth.NewAdmin(now, tenantID, "update-test@example.com", "$2a$10$oldhash", "Original Name", auth.RoleOwner)
	if err != nil {
		t.Fatalf("Failed to create test admin: %v", err)
	}

	// 最初の保存
	err = repo.Save(ctx, admin)
	if err != nil {
		t.Fatalf("Failed to save admin: %v", err)
	}

	// 更新
	updateTime := time.Now()
	err = admin.UpdateDisplayName(updateTime, "New Name")
	if err != nil {
		t.Fatalf("Failed to update display name: %v", err)
	}
	err = admin.UpdatePasswordHash(updateTime, "$2a$10$newhash")
	if err != nil {
		t.Fatalf("Failed to update password hash: %v", err)
	}

	// 再度保存（更新）
	err = repo.Save(ctx, admin)
	if err != nil {
		t.Fatalf("Failed to update admin: %v", err)
	}

	// 取得して確認
	foundAdmin, err := repo.FindByID(ctx, admin.AdminID())
	if err != nil {
		t.Fatalf("Failed to find admin: %v", err)
	}

	if foundAdmin.DisplayName() != "New Name" {
		t.Errorf("DisplayName not updated: got %v, want 'New Name'", foundAdmin.DisplayName())
	}

	if foundAdmin.PasswordHash() != "$2a$10$newhash" {
		t.Errorf("PasswordHash not updated: got %v, want '$2a$10$newhash'", foundAdmin.PasswordHash())
	}
}

func TestAdminRepository_ExistsByEmail(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	repo := db.NewAdminRepository(pool)
	ctx := context.Background()

	tenantID := common.NewTenantID()
	createTestTenant(t, pool, tenantID)

	email := "exists-test@example.com"

	// 最初は存在しない
	exists, err := repo.ExistsByEmail(ctx, tenantID, email)
	if err != nil {
		t.Fatalf("Failed to check email existence: %v", err)
	}
	if exists {
		t.Error("Email should not exist yet")
	}

	// 管理者を作成・保存
	now := time.Now()
	admin, err := auth.NewAdmin(now, tenantID, email, "$2a$10$hash", "Exists Test Admin", auth.RoleOwner)
	if err != nil {
		t.Fatalf("Failed to create admin: %v", err)
	}

	err = repo.Save(ctx, admin)
	if err != nil {
		t.Fatalf("Failed to save admin: %v", err)
	}

	// 存在確認
	exists, err = repo.ExistsByEmail(ctx, tenantID, email)
	if err != nil {
		t.Fatalf("Failed to check email existence: %v", err)
	}
	if !exists {
		t.Error("Email should exist now")
	}

	// 異なるテナントでは存在しない
	otherTenantID := common.NewTenantID()
	createTestTenant(t, pool, otherTenantID)
	exists, err = repo.ExistsByEmail(ctx, otherTenantID, email)
	if err != nil {
		t.Fatalf("Failed to check email existence in other tenant: %v", err)
	}
	if exists {
		t.Error("Email should not exist in other tenant")
	}
}

func TestAdminRepository_Delete(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	repo := db.NewAdminRepository(pool)
	ctx := context.Background()

	tenantID := common.NewTenantID()
	createTestTenant(t, pool, tenantID)

	now := time.Now()
	admin, err := auth.NewAdmin(now, tenantID, "delete-test@example.com", "$2a$10$hash", "Delete Test Admin", auth.RoleOwner)
	if err != nil {
		t.Fatalf("Failed to create test admin: %v", err)
	}

	err = repo.Save(ctx, admin)
	if err != nil {
		t.Fatalf("Failed to save admin: %v", err)
	}

	// 削除前に存在確認
	_, err = repo.FindByID(ctx, admin.AdminID())
	if err != nil {
		t.Fatalf("Admin should exist before delete: %v", err)
	}

	// 削除
	err = repo.Delete(ctx, tenantID, admin.AdminID())
	if err != nil {
		t.Fatalf("Failed to delete admin: %v", err)
	}

	// 削除後は見つからない
	_, err = repo.FindByID(ctx, admin.AdminID())
	if err == nil {
		t.Error("Admin should not exist after delete")
	}
}

func TestAdminRepository_DeleteNotFound(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	repo := db.NewAdminRepository(pool)
	ctx := context.Background()

	tenantID := common.NewTenantID()
	createTestTenant(t, pool, tenantID)

	// 存在しない管理者を削除しようとする
	nonExistentID := common.NewAdminID()
	err := repo.Delete(ctx, tenantID, nonExistentID)
	if err == nil {
		t.Error("Expected error when deleting non-existent admin")
	}
}
