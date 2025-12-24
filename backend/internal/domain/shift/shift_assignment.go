package shift

import (
	"fmt"
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
)

// AssignmentStatus represents the status of a shift assignment
type AssignmentStatus string

const (
	AssignmentStatusConfirmed AssignmentStatus = "confirmed" // 確定
	AssignmentStatusCancelled AssignmentStatus = "cancelled" // キャンセル
)

func (s AssignmentStatus) Validate() error {
	switch s {
	case AssignmentStatusConfirmed, AssignmentStatusCancelled:
		return nil
	default:
		return fmt.Errorf("invalid assignment status: %s", s)
	}
}

// AssignmentMethod represents the method of assignment
type AssignmentMethod string

const (
	AssignmentMethodAuto   AssignmentMethod = "auto"   // 自動割り当て
	AssignmentMethodManual AssignmentMethod = "manual" // 手動割り当て
)

func (m AssignmentMethod) Validate() error {
	switch m {
	case AssignmentMethodAuto, AssignmentMethodManual:
		return nil
	default:
		return fmt.Errorf("invalid assignment method: %s", m)
	}
}

// AssignmentID represents a shift assignment identifier
type AssignmentID string

// NewAssignmentIDWithTime creates a new AssignmentID using the provided time.
func NewAssignmentIDWithTime(t time.Time) AssignmentID {
	return AssignmentID(common.NewULIDWithTime(t))
}

// NewAssignmentID creates a new AssignmentID using the current time.
// Deprecated: Use NewAssignmentIDWithTime for better testability.
func NewAssignmentID() AssignmentID {
	return AssignmentID(common.NewULID())
}

func (id AssignmentID) String() string {
	return string(id)
}

func (id AssignmentID) Validate() error {
	if id == "" {
		return fmt.Errorf("assignment_id is required")
	}
	return common.ValidateULID(string(id))
}

func ParseAssignmentID(s string) (AssignmentID, error) {
	if err := common.ValidateULID(s); err != nil {
		return "", err
	}
	return AssignmentID(s), nil
}

// PlanID represents a shift plan identifier
type PlanID string

// NewPlanIDWithTime creates a new PlanID using the provided time.
func NewPlanIDWithTime(t time.Time) PlanID {
	return PlanID(common.NewULIDWithTime(t))
}

// NewPlanID creates a new PlanID using the current time.
// Deprecated: Use NewPlanIDWithTime for better testability.
func NewPlanID() PlanID {
	return PlanID(common.NewULID())
}

func (id PlanID) String() string {
	return string(id)
}

func (id PlanID) Validate() error {
	if id == "" {
		return fmt.Errorf("plan_id is required")
	}
	return common.ValidateULID(string(id))
}

// ShiftAssignment represents a shift assignment entity
// ShiftPlan 集約内のエンティティ
type ShiftAssignment struct {
	assignmentID        AssignmentID
	tenantID            common.TenantID
	planID              PlanID
	slotID              SlotID
	memberID            common.MemberID
	assignmentStatus    AssignmentStatus
	assignmentMethod    AssignmentMethod
	isOutsidePreference bool
	assignedAt          time.Time
	cancelledAt         *time.Time
	createdAt           time.Time
	updatedAt           time.Time
	deletedAt           *time.Time
}

// NewShiftAssignment creates a new ShiftAssignment entity
func NewShiftAssignment(
	now time.Time,
	tenantID common.TenantID,
	planID PlanID,
	slotID SlotID,
	memberID common.MemberID,
	assignmentMethod AssignmentMethod,
	isOutsidePreference bool,
) (*ShiftAssignment, error) {
	assignment := &ShiftAssignment{
		assignmentID:        NewAssignmentIDWithTime(now),
		tenantID:            tenantID,
		planID:              planID,
		slotID:              slotID,
		memberID:            memberID,
		assignmentStatus:    AssignmentStatusConfirmed,
		assignmentMethod:    assignmentMethod,
		isOutsidePreference: isOutsidePreference,
		assignedAt:          now,
		createdAt:           now,
		updatedAt:           now,
	}

	if err := assignment.validate(); err != nil {
		return nil, err
	}

	return assignment, nil
}

// ReconstructShiftAssignment reconstructs a ShiftAssignment from persistence
func ReconstructShiftAssignment(
	assignmentID AssignmentID,
	tenantID common.TenantID,
	planID PlanID,
	slotID SlotID,
	memberID common.MemberID,
	assignmentStatus AssignmentStatus,
	assignmentMethod AssignmentMethod,
	isOutsidePreference bool,
	assignedAt time.Time,
	cancelledAt *time.Time,
	createdAt time.Time,
	updatedAt time.Time,
	deletedAt *time.Time,
) (*ShiftAssignment, error) {
	assignment := &ShiftAssignment{
		assignmentID:        assignmentID,
		tenantID:            tenantID,
		planID:              planID,
		slotID:              slotID,
		memberID:            memberID,
		assignmentStatus:    assignmentStatus,
		assignmentMethod:    assignmentMethod,
		isOutsidePreference: isOutsidePreference,
		assignedAt:          assignedAt,
		cancelledAt:         cancelledAt,
		createdAt:           createdAt,
		updatedAt:           updatedAt,
		deletedAt:           deletedAt,
	}

	if err := assignment.validate(); err != nil {
		return nil, err
	}

	return assignment, nil
}

func (a *ShiftAssignment) validate() error {
	// TenantID の必須性チェック
	if err := a.tenantID.Validate(); err != nil {
		return common.NewValidationError("tenant_id is required", err)
	}

	// PlanID のバリデーション（空文字列の場合はスキップ - 簡易実装でNULLを許可）
	if a.planID.String() != "" {
		if err := a.planID.Validate(); err != nil {
			return common.NewValidationError("invalid plan_id format", err)
		}
	}

	// SlotID の必須性チェック
	if err := a.slotID.Validate(); err != nil {
		return common.NewValidationError("slot_id is required", err)
	}

	// MemberID の必須性チェック
	if err := a.memberID.Validate(); err != nil {
		return common.NewValidationError("member_id is required", err)
	}

	// AssignmentStatus のバリデーション
	if err := a.assignmentStatus.Validate(); err != nil {
		return common.NewValidationError("invalid assignment_status", err)
	}

	// AssignmentMethod のバリデーション
	if err := a.assignmentMethod.Validate(); err != nil {
		return common.NewValidationError("invalid assignment_method", err)
	}

	// cancelled の場合は cancelledAt が必須
	if a.assignmentStatus == AssignmentStatusCancelled {
		if a.cancelledAt == nil {
			return common.NewValidationError("cancelled_at is required when status is cancelled", nil)
		}
	} else if a.assignmentStatus == AssignmentStatusConfirmed {
		if a.cancelledAt != nil {
			return common.NewValidationError("cancelled_at must be null when status is confirmed", nil)
		}
	}

	return nil
}

// Getters

func (a *ShiftAssignment) AssignmentID() AssignmentID {
	return a.assignmentID
}

func (a *ShiftAssignment) TenantID() common.TenantID {
	return a.tenantID
}

func (a *ShiftAssignment) PlanID() PlanID {
	return a.planID
}

func (a *ShiftAssignment) SlotID() SlotID {
	return a.slotID
}

func (a *ShiftAssignment) MemberID() common.MemberID {
	return a.memberID
}

func (a *ShiftAssignment) AssignmentStatus() AssignmentStatus {
	return a.assignmentStatus
}

func (a *ShiftAssignment) AssignmentMethod() AssignmentMethod {
	return a.assignmentMethod
}

func (a *ShiftAssignment) IsOutsidePreference() bool {
	return a.isOutsidePreference
}

func (a *ShiftAssignment) AssignedAt() time.Time {
	return a.assignedAt
}

func (a *ShiftAssignment) CancelledAt() *time.Time {
	return a.cancelledAt
}

func (a *ShiftAssignment) CreatedAt() time.Time {
	return a.createdAt
}

func (a *ShiftAssignment) UpdatedAt() time.Time {
	return a.updatedAt
}

func (a *ShiftAssignment) DeletedAt() *time.Time {
	return a.deletedAt
}

func (a *ShiftAssignment) IsDeleted() bool {
	return a.deletedAt != nil
}

func (a *ShiftAssignment) IsConfirmed() bool {
	return a.assignmentStatus == AssignmentStatusConfirmed && !a.IsDeleted()
}

func (a *ShiftAssignment) IsCancelled() bool {
	return a.assignmentStatus == AssignmentStatusCancelled
}

// Cancel marks the assignment as cancelled
func (a *ShiftAssignment) Cancel() error {
	if a.assignmentStatus == AssignmentStatusCancelled {
		return common.NewInvariantViolationError("assignment is already cancelled")
	}

	now := time.Now()
	a.assignmentStatus = AssignmentStatusCancelled
	a.cancelledAt = &now
	a.updatedAt = now
	return nil
}

// Delete marks the assignment as deleted (soft delete)
// cancelledとdeleted_atの違い:
// - cancelled: メンバーがキャンセルした（履歴として残し、UIにも表示可能）
// - deleted_at: 管理者が誤って作成した割り当てを削除（履歴から除外）
func (a *ShiftAssignment) Delete() {
	now := time.Now()
	a.deletedAt = &now
	a.updatedAt = now
}

