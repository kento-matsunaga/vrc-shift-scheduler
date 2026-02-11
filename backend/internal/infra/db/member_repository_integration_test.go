package db_test

import (
	"context"
	"testing"
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/member"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/infra/db"
)

// =====================================================
// MemberRepository Integration Tests
// =====================================================

func TestMemberRepository_SaveAndFindByID(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	repo := db.NewMemberRepository(pool)
	ctx := context.Background()

	// テスト用のテナントを作成
	tenantID := common.NewTenantID()
	createTestTenant(t, pool, tenantID)

	// テスト用のメンバーを作成
	now := time.Now()
	mem, err := member.NewMember(now, tenantID, "テストメンバー", "discord_123", "test@example.com")
	if err != nil {
		t.Fatalf("Failed to create test member: %v", err)
	}

	// Save
	err = repo.Save(ctx, mem)
	if err != nil {
		t.Fatalf("Failed to save member: %v", err)
	}

	// FindByID
	foundMember, err := repo.FindByID(ctx, tenantID, mem.MemberID())
	if err != nil {
		t.Fatalf("Failed to find member by ID: %v", err)
	}

	// 値の検証
	if foundMember.MemberID() != mem.MemberID() {
		t.Errorf("MemberID mismatch: got %v, want %v", foundMember.MemberID(), mem.MemberID())
	}

	if foundMember.TenantID() != mem.TenantID() {
		t.Errorf("TenantID mismatch: got %v, want %v", foundMember.TenantID(), mem.TenantID())
	}

	if foundMember.DisplayName() != mem.DisplayName() {
		t.Errorf("DisplayName mismatch: got %v, want %v", foundMember.DisplayName(), mem.DisplayName())
	}

	if foundMember.DiscordUserID() != mem.DiscordUserID() {
		t.Errorf("DiscordUserID mismatch: got %v, want %v", foundMember.DiscordUserID(), mem.DiscordUserID())
	}

	if foundMember.Email() != mem.Email() {
		t.Errorf("Email mismatch: got %v, want %v", foundMember.Email(), mem.Email())
	}

	if foundMember.IsActive() != mem.IsActive() {
		t.Errorf("IsActive mismatch: got %v, want %v", foundMember.IsActive(), mem.IsActive())
	}
}

func TestMemberRepository_FindByTenantID(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	repo := db.NewMemberRepository(pool)
	ctx := context.Background()

	tenantID := common.NewTenantID()
	createTestTenant(t, pool, tenantID)

	// 3人のメンバーを作成
	now := time.Now()
	for i := 0; i < 3; i++ {
		mem, err := member.NewMember(
			now,
			tenantID,
			"テストメンバー"+string(rune('A'+i)),
			"discord_"+string(rune('1'+i)),
			"member"+string(rune('a'+i))+"@example.com",
		)
		if err != nil {
			t.Fatalf("Failed to create test member: %v", err)
		}

		err = repo.Save(ctx, mem)
		if err != nil {
			t.Fatalf("Failed to save member: %v", err)
		}
	}

	// テナントIDで検索
	members, err := repo.FindByTenantID(ctx, tenantID)
	if err != nil {
		t.Fatalf("Failed to find members by tenant ID: %v", err)
	}

	if len(members) < 3 {
		t.Errorf("Expected at least 3 members, got %d", len(members))
	}

	// 全ての結果が同じテナントIDを持つことを確認
	for _, m := range members {
		if m.TenantID() != tenantID {
			t.Errorf("Member has wrong tenant ID: got %v, want %v", m.TenantID(), tenantID)
		}
	}
}

func TestMemberRepository_FindActiveByTenantID(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	repo := db.NewMemberRepository(pool)
	ctx := context.Background()

	tenantID := common.NewTenantID()
	createTestTenant(t, pool, tenantID)

	now := time.Now()

	// アクティブなメンバーを作成
	activeMember, err := member.NewMember(now, tenantID, "アクティブメンバー", "discord_active", "active@example.com")
	if err != nil {
		t.Fatalf("Failed to create active member: %v", err)
	}
	err = repo.Save(ctx, activeMember)
	if err != nil {
		t.Fatalf("Failed to save active member: %v", err)
	}

	// 非アクティブなメンバーを作成
	inactiveMember, err := member.NewMember(now, tenantID, "非アクティブメンバー", "discord_inactive", "inactive@example.com")
	if err != nil {
		t.Fatalf("Failed to create inactive member: %v", err)
	}
	inactiveMember.Deactivate(time.Now())
	err = repo.Save(ctx, inactiveMember)
	if err != nil {
		t.Fatalf("Failed to save inactive member: %v", err)
	}

	// アクティブなメンバーのみを検索
	activeMembers, err := repo.FindActiveByTenantID(ctx, tenantID)
	if err != nil {
		t.Fatalf("Failed to find active members: %v", err)
	}

	// 結果にアクティブなメンバーのみが含まれることを確認
	for _, m := range activeMembers {
		if !m.IsActive() {
			t.Errorf("Found inactive member in active members result: %s", m.DisplayName())
		}
	}

	// アクティブメンバーが見つかることを確認
	found := false
	for _, m := range activeMembers {
		if m.DiscordUserID() == "discord_active" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected to find active member in results")
	}
}

func TestMemberRepository_FindByDiscordUserID(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	repo := db.NewMemberRepository(pool)
	ctx := context.Background()

	tenantID := common.NewTenantID()
	createTestTenant(t, pool, tenantID)

	discordUserID := "discord_user_12345"
	now := time.Now()
	mem, err := member.NewMember(now, tenantID, "Discord検索テスト", discordUserID, "discord-test@example.com")
	if err != nil {
		t.Fatalf("Failed to create test member: %v", err)
	}

	err = repo.Save(ctx, mem)
	if err != nil {
		t.Fatalf("Failed to save member: %v", err)
	}

	// Discord User IDで検索
	foundMember, err := repo.FindByDiscordUserID(ctx, tenantID, discordUserID)
	if err != nil {
		t.Fatalf("Failed to find member by Discord User ID: %v", err)
	}

	if foundMember.DiscordUserID() != discordUserID {
		t.Errorf("DiscordUserID mismatch: got %v, want %v", foundMember.DiscordUserID(), discordUserID)
	}

	// 存在しないDiscord User IDで検索
	_, err = repo.FindByDiscordUserID(ctx, tenantID, "nonexistent_discord_id")
	if err == nil {
		t.Error("Expected error when finding member with nonexistent Discord User ID")
	}
}

func TestMemberRepository_FindByEmail(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	repo := db.NewMemberRepository(pool)
	ctx := context.Background()

	tenantID := common.NewTenantID()
	createTestTenant(t, pool, tenantID)

	email := "email-search@example.com"
	now := time.Now()
	mem, err := member.NewMember(now, tenantID, "Email検索テスト", "discord_email_test", email)
	if err != nil {
		t.Fatalf("Failed to create test member: %v", err)
	}

	err = repo.Save(ctx, mem)
	if err != nil {
		t.Fatalf("Failed to save member: %v", err)
	}

	// Emailで検索
	foundMember, err := repo.FindByEmail(ctx, tenantID, email)
	if err != nil {
		t.Fatalf("Failed to find member by email: %v", err)
	}

	if foundMember.Email() != email {
		t.Errorf("Email mismatch: got %v, want %v", foundMember.Email(), email)
	}

	// 存在しないEmailで検索
	_, err = repo.FindByEmail(ctx, tenantID, "nonexistent@example.com")
	if err == nil {
		t.Error("Expected error when finding member with nonexistent email")
	}
}

func TestMemberRepository_ExistsByDiscordUserID(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	repo := db.NewMemberRepository(pool)
	ctx := context.Background()

	tenantID := common.NewTenantID()
	createTestTenant(t, pool, tenantID)

	discordUserID := "discord_exists_test_123"

	// 最初は存在しない
	exists, err := repo.ExistsByDiscordUserID(ctx, tenantID, discordUserID)
	if err != nil {
		t.Fatalf("Failed to check Discord User ID existence: %v", err)
	}
	if exists {
		t.Error("Discord User ID should not exist yet")
	}

	// メンバーを作成・保存
	now := time.Now()
	mem, err := member.NewMember(now, tenantID, "Discord存在チェック", discordUserID, "exists-check@example.com")
	if err != nil {
		t.Fatalf("Failed to create member: %v", err)
	}

	err = repo.Save(ctx, mem)
	if err != nil {
		t.Fatalf("Failed to save member: %v", err)
	}

	// 存在確認
	exists, err = repo.ExistsByDiscordUserID(ctx, tenantID, discordUserID)
	if err != nil {
		t.Fatalf("Failed to check Discord User ID existence: %v", err)
	}
	if !exists {
		t.Error("Discord User ID should exist now")
	}

	// 異なるテナントでは存在しない
	otherTenantID := common.NewTenantID()
	createTestTenant(t, pool, otherTenantID)
	exists, err = repo.ExistsByDiscordUserID(ctx, otherTenantID, discordUserID)
	if err != nil {
		t.Fatalf("Failed to check Discord User ID existence in other tenant: %v", err)
	}
	if exists {
		t.Error("Discord User ID should not exist in other tenant")
	}
}

func TestMemberRepository_ExistsByEmail(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	repo := db.NewMemberRepository(pool)
	ctx := context.Background()

	tenantID := common.NewTenantID()
	createTestTenant(t, pool, tenantID)

	email := "email-exists-test@example.com"

	// 最初は存在しない
	exists, err := repo.ExistsByEmail(ctx, tenantID, email)
	if err != nil {
		t.Fatalf("Failed to check email existence: %v", err)
	}
	if exists {
		t.Error("Email should not exist yet")
	}

	// メンバーを作成・保存
	now := time.Now()
	mem, err := member.NewMember(now, tenantID, "Email存在チェック", "discord_email_exists", email)
	if err != nil {
		t.Fatalf("Failed to create member: %v", err)
	}

	err = repo.Save(ctx, mem)
	if err != nil {
		t.Fatalf("Failed to save member: %v", err)
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

func TestMemberRepository_Update(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	repo := db.NewMemberRepository(pool)
	ctx := context.Background()

	tenantID := common.NewTenantID()
	createTestTenant(t, pool, tenantID)

	now := time.Now()
	mem, err := member.NewMember(now, tenantID, "元の名前", "discord_update_test", "update@example.com")
	if err != nil {
		t.Fatalf("Failed to create test member: %v", err)
	}

	// 最初の保存
	err = repo.Save(ctx, mem)
	if err != nil {
		t.Fatalf("Failed to save member: %v", err)
	}

	// 更新
	err = mem.UpdateDisplayName(time.Now(), "新しい名前")
	if err != nil {
		t.Fatalf("Failed to update display name: %v", err)
	}

	// 再度保存（更新）
	err = repo.Save(ctx, mem)
	if err != nil {
		t.Fatalf("Failed to update member: %v", err)
	}

	// 取得して確認
	foundMember, err := repo.FindByID(ctx, tenantID, mem.MemberID())
	if err != nil {
		t.Fatalf("Failed to find member: %v", err)
	}

	if foundMember.DisplayName() != "新しい名前" {
		t.Errorf("DisplayName not updated: got %v, want '新しい名前'", foundMember.DisplayName())
	}
}

func TestMemberRepository_Delete(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	repo := db.NewMemberRepository(pool)
	ctx := context.Background()

	tenantID := common.NewTenantID()
	createTestTenant(t, pool, tenantID)

	now := time.Now()
	mem, err := member.NewMember(now, tenantID, "削除テストメンバー", "discord_delete_test", "delete@example.com")
	if err != nil {
		t.Fatalf("Failed to create test member: %v", err)
	}

	err = repo.Save(ctx, mem)
	if err != nil {
		t.Fatalf("Failed to save member: %v", err)
	}

	// 削除前に存在確認
	_, err = repo.FindByID(ctx, tenantID, mem.MemberID())
	if err != nil {
		t.Fatalf("Member should exist before delete: %v", err)
	}

	// 削除
	err = repo.Delete(ctx, tenantID, mem.MemberID())
	if err != nil {
		t.Fatalf("Failed to delete member: %v", err)
	}

	// 削除後は見つからない
	_, err = repo.FindByID(ctx, tenantID, mem.MemberID())
	if err == nil {
		t.Error("Member should not exist after delete")
	}
}

func TestMemberRepository_DeleteNotFound(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	repo := db.NewMemberRepository(pool)
	ctx := context.Background()

	tenantID := common.NewTenantID()
	createTestTenant(t, pool, tenantID)

	// 存在しないメンバーを削除しようとする
	nonExistentID := common.NewMemberID()
	err := repo.Delete(ctx, tenantID, nonExistentID)
	if err == nil {
		t.Error("Expected error when deleting non-existent member")
	}
}

func TestMemberRepository_SoftDelete(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	repo := db.NewMemberRepository(pool)
	ctx := context.Background()

	tenantID := common.NewTenantID()
	createTestTenant(t, pool, tenantID)

	now := time.Now()
	mem, err := member.NewMember(now, tenantID, "論理削除テスト", "discord_soft_delete", "soft-delete@example.com")
	if err != nil {
		t.Fatalf("Failed to create test member: %v", err)
	}

	err = repo.Save(ctx, mem)
	if err != nil {
		t.Fatalf("Failed to save member: %v", err)
	}

	// 論理削除
	mem.Delete(time.Now())
	err = repo.Save(ctx, mem)
	if err != nil {
		t.Fatalf("Failed to soft delete member: %v", err)
	}

	// FindByIDでは見つからない（deleted_at IS NULLの条件があるため）
	_, err = repo.FindByID(ctx, tenantID, mem.MemberID())
	if err == nil {
		t.Error("Soft deleted member should not be found by FindByID")
	}

	// FindByTenantIDでも見つからない
	members, err := repo.FindByTenantID(ctx, tenantID)
	if err != nil {
		t.Fatalf("Failed to find members: %v", err)
	}

	for _, m := range members {
		if m.MemberID() == mem.MemberID() {
			t.Error("Soft deleted member should not appear in FindByTenantID results")
		}
	}
}

func TestMemberRepository_TenantIsolation(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	repo := db.NewMemberRepository(pool)
	ctx := context.Background()

	// 2つのテナントを作成
	tenantID1 := common.NewTenantID()
	tenantID2 := common.NewTenantID()
	createTestTenant(t, pool, tenantID1)
	createTestTenant(t, pool, tenantID2)

	now := time.Now()

	// テナント1のメンバーを作成
	member1, err := member.NewMember(now, tenantID1, "テナント1のメンバー", "discord_tenant1", "tenant1@example.com")
	if err != nil {
		t.Fatalf("Failed to create member for tenant 1: %v", err)
	}
	err = repo.Save(ctx, member1)
	if err != nil {
		t.Fatalf("Failed to save member for tenant 1: %v", err)
	}

	// テナント2のメンバーを作成
	member2, err := member.NewMember(now, tenantID2, "テナント2のメンバー", "discord_tenant2", "tenant2@example.com")
	if err != nil {
		t.Fatalf("Failed to create member for tenant 2: %v", err)
	}
	err = repo.Save(ctx, member2)
	if err != nil {
		t.Fatalf("Failed to save member for tenant 2: %v", err)
	}

	// テナント1のメンバーはテナント2からは見えない
	_, err = repo.FindByID(ctx, tenantID2, member1.MemberID())
	if err == nil {
		t.Error("Member from tenant 1 should not be visible from tenant 2")
	}

	// テナント2のメンバーはテナント1からは見えない
	_, err = repo.FindByID(ctx, tenantID1, member2.MemberID())
	if err == nil {
		t.Error("Member from tenant 2 should not be visible from tenant 1")
	}

	// 各テナントのメンバー一覧は自分のテナントのメンバーのみ
	members1, err := repo.FindByTenantID(ctx, tenantID1)
	if err != nil {
		t.Fatalf("Failed to find members for tenant 1: %v", err)
	}
	for _, m := range members1 {
		if m.TenantID() != tenantID1 {
			t.Errorf("Found member from wrong tenant: got %v, want %v", m.TenantID(), tenantID1)
		}
	}

	members2, err := repo.FindByTenantID(ctx, tenantID2)
	if err != nil {
		t.Fatalf("Failed to find members for tenant 2: %v", err)
	}
	for _, m := range members2 {
		if m.TenantID() != tenantID2 {
			t.Errorf("Found member from wrong tenant: got %v, want %v", m.TenantID(), tenantID2)
		}
	}
}
