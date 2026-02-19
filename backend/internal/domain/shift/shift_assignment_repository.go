package shift

import (
	"context"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/event"
)

// ShiftAssignmentRepository defines the interface for ShiftAssignment persistence
type ShiftAssignmentRepository interface {
	// Save saves a shift assignment (insert or update)
	Save(ctx context.Context, assignment *ShiftAssignment) error

	// FindByID finds a shift assignment by ID within a tenant
	FindByID(ctx context.Context, tenantID common.TenantID, assignmentID AssignmentID) (*ShiftAssignment, error)

	// FindBySlotID finds all shift assignments for a slot
	FindBySlotID(ctx context.Context, tenantID common.TenantID, slotID SlotID) ([]*ShiftAssignment, error)

	// FindConfirmedBySlotID finds all confirmed shift assignments for a slot
	// 同時実行制御のため、FOR UPDATE でロックする必要がある場合もある
	FindConfirmedBySlotID(ctx context.Context, tenantID common.TenantID, slotID SlotID) ([]*ShiftAssignment, error)

	// FindByMemberID finds all shift assignments for a member
	FindByMemberID(ctx context.Context, tenantID common.TenantID, memberID common.MemberID) ([]*ShiftAssignment, error)

	// FindConfirmedByMemberID finds all confirmed shift assignments for a member
	FindConfirmedByMemberID(ctx context.Context, tenantID common.TenantID, memberID common.MemberID) ([]*ShiftAssignment, error)

	// FindByPlanID finds all shift assignments for a plan
	FindByPlanID(ctx context.Context, tenantID common.TenantID, planID PlanID) ([]*ShiftAssignment, error)

	// CountConfirmedBySlotID counts confirmed assignments for a slot
	// required_count 制御に使用
	CountConfirmedBySlotID(ctx context.Context, tenantID common.TenantID, slotID SlotID) (int, error)

	// Delete deletes a shift assignment (physical delete)
	// 通常は ShiftAssignment.Delete() で論理削除を使用するため、このメソッドは稀に使用
	Delete(ctx context.Context, tenantID common.TenantID, assignmentID AssignmentID) error

	// ExistsBySlotIDAndMemberID checks if a confirmed assignment exists for the given slot and member
	// 重複割り当てチェックに使用
	ExistsBySlotIDAndMemberID(ctx context.Context, tenantID common.TenantID, slotID SlotID, memberID common.MemberID) (bool, error)

	// HasConfirmedByMemberAndBusinessDayID checks if a confirmed assignment exists for the given member and business day
	// Used for actual attendance calculation
	HasConfirmedByMemberAndBusinessDayID(ctx context.Context, tenantID common.TenantID, memberID common.MemberID, businessDayID event.BusinessDayID) (bool, error)

	// FindByBusinessDayID finds all shift assignments for a business day
	// Used for bulk fetching assignments to avoid N+1 problem
	FindByBusinessDayID(ctx context.Context, tenantID common.TenantID, businessDayID event.BusinessDayID) ([]*ShiftAssignment, error)
}
