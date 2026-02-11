package db_test

import (
	"context"
	"testing"
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/role"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/infra/db"
	"github.com/jackc/pgx/v5/pgxpool"
)

// =====================================================
// Test Helpers for RoleGroup Integration Tests
// =====================================================

func createTestRoleGroup(t *testing.T, pool *pgxpool.Pool, tenantID common.TenantID) *role.RoleGroup {
	t.Helper()
	now := time.Now()
	uniqueName := "テストロールグループ_" + common.NewULID()
	group, err := role.NewRoleGroup(now, tenantID, uniqueName, "説明", "", 0)
	if err != nil {
		t.Fatalf("Failed to create test role group: %v", err)
	}
	repo := db.NewRoleGroupRepository(pool)
	if err := repo.Save(context.Background(), group); err != nil {
		t.Fatalf("Failed to save test role group: %v", err)
	}
	return group
}

// =====================================================
// RoleGroupRepository Integration Tests
// =====================================================

func TestRoleGroupRepository_AssignRole_Success(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	tenantID := common.NewTenantID()
	createTestTenant(t, pool, tenantID)

	r := createTestRole(t, pool, tenantID)
	group := createTestRoleGroup(t, pool, tenantID)

	repo := db.NewRoleGroupRepository(pool)
	ctx := context.Background()

	err := repo.AssignRole(ctx, group.GroupID(), r.RoleID())
	if err != nil {
		t.Fatalf("AssignRole() should succeed: %v", err)
	}

	// Verify assignment
	roleIDs, err := repo.FindRoleIDsByGroupID(ctx, group.GroupID())
	if err != nil {
		t.Fatalf("FindRoleIDsByGroupID() failed: %v", err)
	}
	if len(roleIDs) != 1 {
		t.Fatalf("Expected 1 role, got %d", len(roleIDs))
	}
	if roleIDs[0] != r.RoleID() {
		t.Errorf("RoleID mismatch: got %v, want %v", roleIDs[0], r.RoleID())
	}
}

func TestRoleGroupRepository_AssignRole_CrossTenantRejected(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	tenantA := common.NewTenantID()
	tenantB := common.NewTenantID()
	createTestTenant(t, pool, tenantA)
	createTestTenant(t, pool, tenantB)

	roleA := createTestRole(t, pool, tenantA)
	groupB := createTestRoleGroup(t, pool, tenantB)

	repo := db.NewRoleGroupRepository(pool)
	ctx := context.Background()

	// Cross-tenant assignment should not error but should insert 0 rows
	err := repo.AssignRole(ctx, groupB.GroupID(), roleA.RoleID())
	if err != nil {
		t.Fatalf("AssignRole() should not error: %v", err)
	}

	// Verify no assignment was made
	roleIDs, err := repo.FindRoleIDsByGroupID(ctx, groupB.GroupID())
	if err != nil {
		t.Fatalf("FindRoleIDsByGroupID() failed: %v", err)
	}
	if len(roleIDs) != 0 {
		t.Errorf("Cross-tenant role should not be assigned, got %d roles", len(roleIDs))
	}
}

func TestRoleGroupRepository_RemoveRole_Success(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	tenantID := common.NewTenantID()
	createTestTenant(t, pool, tenantID)

	r := createTestRole(t, pool, tenantID)
	group := createTestRoleGroup(t, pool, tenantID)

	repo := db.NewRoleGroupRepository(pool)
	ctx := context.Background()

	// Assign then remove
	err := repo.AssignRole(ctx, group.GroupID(), r.RoleID())
	if err != nil {
		t.Fatalf("AssignRole() failed: %v", err)
	}

	err = repo.RemoveRole(ctx, group.GroupID(), r.RoleID())
	if err != nil {
		t.Fatalf("RemoveRole() should succeed: %v", err)
	}

	// Verify removal
	roleIDs, err := repo.FindRoleIDsByGroupID(ctx, group.GroupID())
	if err != nil {
		t.Fatalf("FindRoleIDsByGroupID() failed: %v", err)
	}
	if len(roleIDs) != 0 {
		t.Errorf("Expected 0 roles after removal, got %d", len(roleIDs))
	}
}

func TestRoleGroupRepository_FindRoleIDsByGroupID(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	tenantA := common.NewTenantID()
	tenantB := common.NewTenantID()
	createTestTenant(t, pool, tenantA)
	createTestTenant(t, pool, tenantB)

	roleA1 := createTestRole(t, pool, tenantA)
	roleA2 := createTestRole(t, pool, tenantA)
	roleB := createTestRole(t, pool, tenantB)
	groupA := createTestRoleGroup(t, pool, tenantA)

	repo := db.NewRoleGroupRepository(pool)
	ctx := context.Background()

	// Assign same-tenant roles
	_ = repo.AssignRole(ctx, groupA.GroupID(), roleA1.RoleID())
	_ = repo.AssignRole(ctx, groupA.GroupID(), roleA2.RoleID())
	// Cross-tenant assign (should be silently rejected)
	_ = repo.AssignRole(ctx, groupA.GroupID(), roleB.RoleID())

	roleIDs, err := repo.FindRoleIDsByGroupID(ctx, groupA.GroupID())
	if err != nil {
		t.Fatalf("FindRoleIDsByGroupID() failed: %v", err)
	}

	// Only same-tenant roles should be returned
	if len(roleIDs) != 2 {
		t.Errorf("Expected 2 roles (same tenant only), got %d", len(roleIDs))
	}

	for _, rid := range roleIDs {
		if rid == roleB.RoleID() {
			t.Error("Cross-tenant role should not be in results")
		}
	}
}

func TestRoleGroupRepository_FindGroupIDsByRoleID(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	tenantA := common.NewTenantID()
	tenantB := common.NewTenantID()
	createTestTenant(t, pool, tenantA)
	createTestTenant(t, pool, tenantB)

	roleA := createTestRole(t, pool, tenantA)
	groupA1 := createTestRoleGroup(t, pool, tenantA)
	groupA2 := createTestRoleGroup(t, pool, tenantA)

	repo := db.NewRoleGroupRepository(pool)
	ctx := context.Background()

	// Assign role to two same-tenant groups
	_ = repo.AssignRole(ctx, groupA1.GroupID(), roleA.RoleID())
	_ = repo.AssignRole(ctx, groupA2.GroupID(), roleA.RoleID())

	groupIDs, err := repo.FindGroupIDsByRoleID(ctx, roleA.RoleID())
	if err != nil {
		t.Fatalf("FindGroupIDsByRoleID() failed: %v", err)
	}
	if len(groupIDs) != 2 {
		t.Errorf("Expected 2 groups, got %d", len(groupIDs))
	}
}

func TestRoleGroupRepository_SetGroupRoles(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	tenantA := common.NewTenantID()
	tenantB := common.NewTenantID()
	createTestTenant(t, pool, tenantA)
	createTestTenant(t, pool, tenantB)

	groupA := createTestRoleGroup(t, pool, tenantA)
	roleA1 := createTestRole(t, pool, tenantA)
	roleA2 := createTestRole(t, pool, tenantA)
	roleA3 := createTestRole(t, pool, tenantA)
	roleB := createTestRole(t, pool, tenantB)

	repo := db.NewRoleGroupRepository(pool)
	ctx := context.Background()

	// Initial assignment
	err := repo.SetGroupRoles(ctx, groupA.GroupID(), []common.RoleID{roleA1.RoleID(), roleA2.RoleID()})
	if err != nil {
		t.Fatalf("SetGroupRoles() failed: %v", err)
	}

	roleIDs, err := repo.FindRoleIDsByGroupID(ctx, groupA.GroupID())
	if err != nil {
		t.Fatalf("FindRoleIDsByGroupID() failed: %v", err)
	}
	if len(roleIDs) != 2 {
		t.Errorf("Expected 2 roles after initial set, got %d", len(roleIDs))
	}

	// Replace with new set including cross-tenant role (which should be silently ignored)
	err = repo.SetGroupRoles(ctx, groupA.GroupID(), []common.RoleID{roleA3.RoleID(), roleB.RoleID()})
	if err != nil {
		t.Fatalf("SetGroupRoles() failed: %v", err)
	}

	roleIDs, err = repo.FindRoleIDsByGroupID(ctx, groupA.GroupID())
	if err != nil {
		t.Fatalf("FindRoleIDsByGroupID() failed: %v", err)
	}

	// Only same-tenant role should be assigned
	if len(roleIDs) != 1 {
		t.Errorf("Expected 1 role (cross-tenant ignored), got %d", len(roleIDs))
	}
	if len(roleIDs) > 0 && roleIDs[0] != roleA3.RoleID() {
		t.Errorf("Expected roleA3, got %v", roleIDs[0])
	}
}
