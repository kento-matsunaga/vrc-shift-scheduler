package db_test

import (
	"context"
	"testing"
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/member"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/role"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/infra/db"
	"github.com/jackc/pgx/v5/pgxpool"
)

// =====================================================
// Test Helpers for MemberRole Integration Tests
// =====================================================

func createTestMember(t *testing.T, pool *pgxpool.Pool, tenantID common.TenantID) *member.Member {
	t.Helper()
	now := time.Now()
	uniqueName := "テストメンバー_" + common.NewULID()
	mem, err := member.NewMember(now, tenantID, uniqueName, "", "")
	if err != nil {
		t.Fatalf("Failed to create test member: %v", err)
	}
	repo := db.NewMemberRepository(pool)
	if err := repo.Save(context.Background(), mem); err != nil {
		t.Fatalf("Failed to save test member: %v", err)
	}
	return mem
}

func createTestRole(t *testing.T, pool *pgxpool.Pool, tenantID common.TenantID) *role.Role {
	t.Helper()
	now := time.Now()
	uniqueName := "テストロール_" + common.NewULID()
	r, err := role.NewRole(now, tenantID, uniqueName, "", "", 0)
	if err != nil {
		t.Fatalf("Failed to create test role: %v", err)
	}
	repo := db.NewRoleRepository(pool)
	if err := repo.Save(context.Background(), r); err != nil {
		t.Fatalf("Failed to save test role: %v", err)
	}
	return r
}

// =====================================================
// MemberRoleRepository Integration Tests
// =====================================================

func TestMemberRoleRepository_AssignRole_Success(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	tenantID := common.NewTenantID()
	createTestTenant(t, pool, tenantID)

	mem := createTestMember(t, pool, tenantID)
	r := createTestRole(t, pool, tenantID)

	repo := db.NewMemberRoleRepository(pool)
	ctx := context.Background()

	err := repo.AssignRole(ctx, mem.MemberID(), r.RoleID())
	if err != nil {
		t.Fatalf("AssignRole() should succeed: %v", err)
	}

	// Verify assignment
	roles, err := repo.FindRolesByMemberID(ctx, mem.MemberID())
	if err != nil {
		t.Fatalf("FindRolesByMemberID() failed: %v", err)
	}
	if len(roles) != 1 {
		t.Fatalf("Expected 1 role, got %d", len(roles))
	}
	if roles[0] != r.RoleID() {
		t.Errorf("RoleID mismatch: got %v, want %v", roles[0], r.RoleID())
	}
}

func TestMemberRoleRepository_AssignRole_CrossTenantRejected(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	tenantA := common.NewTenantID()
	tenantB := common.NewTenantID()
	createTestTenant(t, pool, tenantA)
	createTestTenant(t, pool, tenantB)

	memberA := createTestMember(t, pool, tenantA)
	roleB := createTestRole(t, pool, tenantB)

	repo := db.NewMemberRoleRepository(pool)
	ctx := context.Background()

	// Cross-tenant assignment should not error but should insert 0 rows
	err := repo.AssignRole(ctx, memberA.MemberID(), roleB.RoleID())
	if err != nil {
		t.Fatalf("AssignRole() should not error: %v", err)
	}

	// Verify no assignment was made
	roles, err := repo.FindRolesByMemberID(ctx, memberA.MemberID())
	if err != nil {
		t.Fatalf("FindRolesByMemberID() failed: %v", err)
	}
	if len(roles) != 0 {
		t.Errorf("Cross-tenant role should not be assigned, got %d roles", len(roles))
	}
}

func TestMemberRoleRepository_RemoveRole_Success(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	tenantID := common.NewTenantID()
	createTestTenant(t, pool, tenantID)

	mem := createTestMember(t, pool, tenantID)
	r := createTestRole(t, pool, tenantID)

	repo := db.NewMemberRoleRepository(pool)
	ctx := context.Background()

	// Assign then remove
	err := repo.AssignRole(ctx, mem.MemberID(), r.RoleID())
	if err != nil {
		t.Fatalf("AssignRole() failed: %v", err)
	}

	err = repo.RemoveRole(ctx, mem.MemberID(), r.RoleID())
	if err != nil {
		t.Fatalf("RemoveRole() should succeed: %v", err)
	}

	// Verify removal
	roles, err := repo.FindRolesByMemberID(ctx, mem.MemberID())
	if err != nil {
		t.Fatalf("FindRolesByMemberID() failed: %v", err)
	}
	if len(roles) != 0 {
		t.Errorf("Expected 0 roles after removal, got %d", len(roles))
	}
}

func TestMemberRoleRepository_RemoveRole_CrossTenantRejected(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	tenantA := common.NewTenantID()
	tenantB := common.NewTenantID()
	createTestTenant(t, pool, tenantA)
	createTestTenant(t, pool, tenantB)

	memberA := createTestMember(t, pool, tenantA)
	roleA := createTestRole(t, pool, tenantA)
	roleB := createTestRole(t, pool, tenantB)

	repo := db.NewMemberRoleRepository(pool)
	ctx := context.Background()

	// Assign same-tenant role
	err := repo.AssignRole(ctx, memberA.MemberID(), roleA.RoleID())
	if err != nil {
		t.Fatalf("AssignRole() failed: %v", err)
	}

	// Try to remove cross-tenant role (should return NotFound)
	err = repo.RemoveRole(ctx, memberA.MemberID(), roleB.RoleID())
	if err == nil {
		t.Error("RemoveRole() should fail for cross-tenant removal")
	}

	// Verify same-tenant assignment still exists
	roles, err := repo.FindRolesByMemberID(ctx, memberA.MemberID())
	if err != nil {
		t.Fatalf("FindRolesByMemberID() failed: %v", err)
	}
	if len(roles) != 1 {
		t.Errorf("Same-tenant role should still exist, got %d roles", len(roles))
	}
}

func TestMemberRoleRepository_FindRolesByMemberID(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	tenantA := common.NewTenantID()
	tenantB := common.NewTenantID()
	createTestTenant(t, pool, tenantA)
	createTestTenant(t, pool, tenantB)

	memberA := createTestMember(t, pool, tenantA)
	roleA1 := createTestRole(t, pool, tenantA)
	roleA2 := createTestRole(t, pool, tenantA)

	repo := db.NewMemberRoleRepository(pool)
	ctx := context.Background()

	// Assign two same-tenant roles
	_ = repo.AssignRole(ctx, memberA.MemberID(), roleA1.RoleID())
	_ = repo.AssignRole(ctx, memberA.MemberID(), roleA2.RoleID())

	roles, err := repo.FindRolesByMemberID(ctx, memberA.MemberID())
	if err != nil {
		t.Fatalf("FindRolesByMemberID() failed: %v", err)
	}
	if len(roles) != 2 {
		t.Errorf("Expected 2 roles, got %d", len(roles))
	}
}

func TestMemberRoleRepository_FindMemberIDsByRoleID(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	tenantA := common.NewTenantID()
	tenantB := common.NewTenantID()
	createTestTenant(t, pool, tenantA)
	createTestTenant(t, pool, tenantB)

	memberA1 := createTestMember(t, pool, tenantA)
	memberA2 := createTestMember(t, pool, tenantA)
	memberB := createTestMember(t, pool, tenantB)
	roleA := createTestRole(t, pool, tenantA)

	repo := db.NewMemberRoleRepository(pool)
	ctx := context.Background()

	// Assign same-tenant members
	_ = repo.AssignRole(ctx, memberA1.MemberID(), roleA.RoleID())
	_ = repo.AssignRole(ctx, memberA2.MemberID(), roleA.RoleID())
	// Cross-tenant assign (should be silently rejected)
	_ = repo.AssignRole(ctx, memberB.MemberID(), roleA.RoleID())

	memberIDs, err := repo.FindMemberIDsByRoleID(ctx, roleA.RoleID())
	if err != nil {
		t.Fatalf("FindMemberIDsByRoleID() failed: %v", err)
	}

	// Only same-tenant members should be returned
	if len(memberIDs) != 2 {
		t.Errorf("Expected 2 members (same tenant only), got %d", len(memberIDs))
	}

	// Verify cross-tenant member is not included
	for _, mid := range memberIDs {
		if mid == memberB.MemberID() {
			t.Error("Cross-tenant member should not be in results")
		}
	}
}

func TestMemberRoleRepository_SetMemberRoles(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	tenantA := common.NewTenantID()
	tenantB := common.NewTenantID()
	createTestTenant(t, pool, tenantA)
	createTestTenant(t, pool, tenantB)

	memberA := createTestMember(t, pool, tenantA)
	roleA1 := createTestRole(t, pool, tenantA)
	roleA2 := createTestRole(t, pool, tenantA)
	roleA3 := createTestRole(t, pool, tenantA)
	roleB := createTestRole(t, pool, tenantB)

	repo := db.NewMemberRoleRepository(pool)
	ctx := context.Background()

	// Initial assignment
	err := repo.SetMemberRoles(ctx, memberA.MemberID(), []common.RoleID{roleA1.RoleID(), roleA2.RoleID()})
	if err != nil {
		t.Fatalf("SetMemberRoles() failed: %v", err)
	}

	roles, err := repo.FindRolesByMemberID(ctx, memberA.MemberID())
	if err != nil {
		t.Fatalf("FindRolesByMemberID() failed: %v", err)
	}
	if len(roles) != 2 {
		t.Errorf("Expected 2 roles after initial set, got %d", len(roles))
	}

	// Replace with new set including cross-tenant role (which should be silently ignored)
	err = repo.SetMemberRoles(ctx, memberA.MemberID(), []common.RoleID{roleA3.RoleID(), roleB.RoleID()})
	if err != nil {
		t.Fatalf("SetMemberRoles() failed: %v", err)
	}

	roles, err = repo.FindRolesByMemberID(ctx, memberA.MemberID())
	if err != nil {
		t.Fatalf("FindRolesByMemberID() failed: %v", err)
	}

	// Only same-tenant role should be assigned
	if len(roles) != 1 {
		t.Errorf("Expected 1 role (cross-tenant ignored), got %d", len(roles))
	}
	if len(roles) > 0 && roles[0] != roleA3.RoleID() {
		t.Errorf("Expected roleA3, got %v", roles[0])
	}
}
