package db_test

import (
	"context"
	"testing"
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/member"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/infra/db"
	"github.com/jackc/pgx/v5/pgxpool"
)

// =====================================================
// Test Helpers for MemberGroup Integration Tests
// =====================================================

func createTestMemberGroup(t *testing.T, pool *pgxpool.Pool, tenantID common.TenantID) *member.MemberGroup {
	t.Helper()
	now := time.Now()
	group, err := member.NewMemberGroup(now, tenantID, "テストグループ", "説明", "", 0)
	if err != nil {
		t.Fatalf("Failed to create test member group: %v", err)
	}
	repo := db.NewMemberGroupRepository(pool)
	if err := repo.Save(context.Background(), group); err != nil {
		t.Fatalf("Failed to save test member group: %v", err)
	}
	return group
}

// =====================================================
// MemberGroupRepository Integration Tests
// =====================================================

func TestMemberGroupRepository_AssignMember_Success(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	tenantID := common.NewTenantID()
	createTestTenant(t, pool, tenantID)

	mem := createTestMember(t, pool, tenantID)
	group := createTestMemberGroup(t, pool, tenantID)

	repo := db.NewMemberGroupRepository(pool)
	ctx := context.Background()

	err := repo.AssignMember(ctx, group.GroupID(), mem.MemberID())
	if err != nil {
		t.Fatalf("AssignMember() should succeed: %v", err)
	}

	// Verify assignment
	memberIDs, err := repo.FindMemberIDsByGroupID(ctx, group.GroupID())
	if err != nil {
		t.Fatalf("FindMemberIDsByGroupID() failed: %v", err)
	}
	if len(memberIDs) != 1 {
		t.Fatalf("Expected 1 member, got %d", len(memberIDs))
	}
	if memberIDs[0] != mem.MemberID() {
		t.Errorf("MemberID mismatch: got %v, want %v", memberIDs[0], mem.MemberID())
	}
}

func TestMemberGroupRepository_AssignMember_CrossTenantRejected(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	tenantA := common.NewTenantID()
	tenantB := common.NewTenantID()
	createTestTenant(t, pool, tenantA)
	createTestTenant(t, pool, tenantB)

	memberA := createTestMember(t, pool, tenantA)
	groupB := createTestMemberGroup(t, pool, tenantB)

	repo := db.NewMemberGroupRepository(pool)
	ctx := context.Background()

	// Cross-tenant assignment should not error but should insert 0 rows
	err := repo.AssignMember(ctx, groupB.GroupID(), memberA.MemberID())
	if err != nil {
		t.Fatalf("AssignMember() should not error: %v", err)
	}

	// Verify no assignment was made
	memberIDs, err := repo.FindMemberIDsByGroupID(ctx, groupB.GroupID())
	if err != nil {
		t.Fatalf("FindMemberIDsByGroupID() failed: %v", err)
	}
	if len(memberIDs) != 0 {
		t.Errorf("Cross-tenant member should not be assigned, got %d members", len(memberIDs))
	}
}

func TestMemberGroupRepository_RemoveMember_Success(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	tenantID := common.NewTenantID()
	createTestTenant(t, pool, tenantID)

	mem := createTestMember(t, pool, tenantID)
	group := createTestMemberGroup(t, pool, tenantID)

	repo := db.NewMemberGroupRepository(pool)
	ctx := context.Background()

	// Assign then remove
	err := repo.AssignMember(ctx, group.GroupID(), mem.MemberID())
	if err != nil {
		t.Fatalf("AssignMember() failed: %v", err)
	}

	err = repo.RemoveMember(ctx, group.GroupID(), mem.MemberID())
	if err != nil {
		t.Fatalf("RemoveMember() should succeed: %v", err)
	}

	// Verify removal
	memberIDs, err := repo.FindMemberIDsByGroupID(ctx, group.GroupID())
	if err != nil {
		t.Fatalf("FindMemberIDsByGroupID() failed: %v", err)
	}
	if len(memberIDs) != 0 {
		t.Errorf("Expected 0 members after removal, got %d", len(memberIDs))
	}
}

func TestMemberGroupRepository_FindMemberIDsByGroupID(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	tenantA := common.NewTenantID()
	tenantB := common.NewTenantID()
	createTestTenant(t, pool, tenantA)
	createTestTenant(t, pool, tenantB)

	memberA1 := createTestMember(t, pool, tenantA)
	memberA2 := createTestMember(t, pool, tenantA)
	memberB := createTestMember(t, pool, tenantB)
	groupA := createTestMemberGroup(t, pool, tenantA)

	repo := db.NewMemberGroupRepository(pool)
	ctx := context.Background()

	// Assign same-tenant members
	_ = repo.AssignMember(ctx, groupA.GroupID(), memberA1.MemberID())
	_ = repo.AssignMember(ctx, groupA.GroupID(), memberA2.MemberID())
	// Cross-tenant assign (should be silently rejected)
	_ = repo.AssignMember(ctx, groupA.GroupID(), memberB.MemberID())

	memberIDs, err := repo.FindMemberIDsByGroupID(ctx, groupA.GroupID())
	if err != nil {
		t.Fatalf("FindMemberIDsByGroupID() failed: %v", err)
	}

	// Only same-tenant members should be returned
	if len(memberIDs) != 2 {
		t.Errorf("Expected 2 members (same tenant only), got %d", len(memberIDs))
	}

	for _, mid := range memberIDs {
		if mid == memberB.MemberID() {
			t.Error("Cross-tenant member should not be in results")
		}
	}
}

func TestMemberGroupRepository_FindGroupIDsByMemberID(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	tenantA := common.NewTenantID()
	tenantB := common.NewTenantID()
	createTestTenant(t, pool, tenantA)
	createTestTenant(t, pool, tenantB)

	memberA := createTestMember(t, pool, tenantA)
	groupA1 := createTestMemberGroup(t, pool, tenantA)
	groupA2 := createTestMemberGroup(t, pool, tenantA)

	repo := db.NewMemberGroupRepository(pool)
	ctx := context.Background()

	// Assign member to two same-tenant groups
	_ = repo.AssignMember(ctx, groupA1.GroupID(), memberA.MemberID())
	_ = repo.AssignMember(ctx, groupA2.GroupID(), memberA.MemberID())

	groupIDs, err := repo.FindGroupIDsByMemberID(ctx, memberA.MemberID())
	if err != nil {
		t.Fatalf("FindGroupIDsByMemberID() failed: %v", err)
	}
	if len(groupIDs) != 2 {
		t.Errorf("Expected 2 groups, got %d", len(groupIDs))
	}
}

func TestMemberGroupRepository_SetMemberGroups(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	tenantA := common.NewTenantID()
	tenantB := common.NewTenantID()
	createTestTenant(t, pool, tenantA)
	createTestTenant(t, pool, tenantB)

	memberA := createTestMember(t, pool, tenantA)
	groupA1 := createTestMemberGroup(t, pool, tenantA)
	groupA2 := createTestMemberGroup(t, pool, tenantA)
	groupA3 := createTestMemberGroup(t, pool, tenantA)
	groupB := createTestMemberGroup(t, pool, tenantB)

	repo := db.NewMemberGroupRepository(pool)
	ctx := context.Background()

	// Initial assignment
	err := repo.SetMemberGroups(ctx, memberA.MemberID(), []common.MemberGroupID{groupA1.GroupID(), groupA2.GroupID()})
	if err != nil {
		t.Fatalf("SetMemberGroups() failed: %v", err)
	}

	groupIDs, err := repo.FindGroupIDsByMemberID(ctx, memberA.MemberID())
	if err != nil {
		t.Fatalf("FindGroupIDsByMemberID() failed: %v", err)
	}
	if len(groupIDs) != 2 {
		t.Errorf("Expected 2 groups after initial set, got %d", len(groupIDs))
	}

	// Replace with new set including cross-tenant group (which should be silently ignored)
	err = repo.SetMemberGroups(ctx, memberA.MemberID(), []common.MemberGroupID{groupA3.GroupID(), groupB.GroupID()})
	if err != nil {
		t.Fatalf("SetMemberGroups() failed: %v", err)
	}

	groupIDs, err = repo.FindGroupIDsByMemberID(ctx, memberA.MemberID())
	if err != nil {
		t.Fatalf("FindGroupIDsByMemberID() failed: %v", err)
	}

	// Only same-tenant group should be assigned
	if len(groupIDs) != 1 {
		t.Errorf("Expected 1 group (cross-tenant ignored), got %d", len(groupIDs))
	}
	if len(groupIDs) > 0 && groupIDs[0] != groupA3.GroupID() {
		t.Errorf("Expected groupA3, got %v", groupIDs[0])
	}
}
