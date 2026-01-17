package shift

import (
	"context"
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/event"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/shift"
)

// commonを使用しているのでインポートを維持

// CreateShiftSlotInput represents the input for creating a shift slot
type CreateShiftSlotInput struct {
	TenantID      common.TenantID
	BusinessDayID event.BusinessDayID
	InstanceID    *shift.InstanceID // optional - nil if not linking to an instance
	SlotName      string
	InstanceName  string
	StartTime     time.Time
	EndTime       time.Time
	RequiredCount int
	Priority      int
}

// CreateShiftSlotUsecase handles the shift slot creation use case
type CreateShiftSlotUsecase struct {
	slotRepo        shift.ShiftSlotRepository
	businessDayRepo event.EventBusinessDayRepository
	instanceRepo    shift.InstanceRepository
}

// NewCreateShiftSlotUsecase creates a new CreateShiftSlotUsecase
func NewCreateShiftSlotUsecase(
	slotRepo shift.ShiftSlotRepository,
	businessDayRepo event.EventBusinessDayRepository,
	instanceRepo shift.InstanceRepository,
) *CreateShiftSlotUsecase {
	return &CreateShiftSlotUsecase{
		slotRepo:        slotRepo,
		businessDayRepo: businessDayRepo,
		instanceRepo:    instanceRepo,
	}
}

// DefaultPriority is the default priority value for new shift slots
const DefaultPriority = 1

// Execute creates a new shift slot
func (uc *CreateShiftSlotUsecase) Execute(ctx context.Context, input CreateShiftSlotInput) (*shift.ShiftSlot, error) {
	// BusinessDay の存在確認
	businessDay, err := uc.businessDayRepo.FindByID(ctx, input.TenantID, input.BusinessDayID)
	if err != nil {
		return nil, err
	}

	// InstanceID が指定されている場合、同じイベントに属しているか検証
	if input.InstanceID != nil {
		instance, err := uc.instanceRepo.FindByID(ctx, input.TenantID, *input.InstanceID)
		if err != nil {
			return nil, err
		}

		// インスタンスが同じイベントに属しているか確認
		if instance.EventID() != businessDay.EventID() {
			return nil, common.NewValidationError("instance does not belong to the same event as the business day", nil)
		}
	}

	// Priority のデフォルト値設定（未指定の場合は1）
	priority := input.Priority
	if priority == 0 {
		priority = DefaultPriority
	}

	// ShiftSlot エンティティの作成
	newSlot, err := shift.NewShiftSlot(
		time.Now(),
		input.TenantID,
		input.BusinessDayID,
		input.InstanceID,
		input.SlotName,
		input.InstanceName,
		input.StartTime,
		input.EndTime,
		input.RequiredCount,
		priority,
	)
	if err != nil {
		return nil, err
	}

	// 保存
	if err := uc.slotRepo.Save(ctx, newSlot); err != nil {
		return nil, err
	}

	return newSlot, nil
}

// ListShiftSlotsInput represents the input for listing shift slots
type ListShiftSlotsInput struct {
	TenantID      common.TenantID
	BusinessDayID event.BusinessDayID
}

// ShiftSlotWithAssignmentCount represents a shift slot with its assignment count
type ShiftSlotWithAssignmentCount struct {
	Slot          *shift.ShiftSlot
	AssignedCount int
}

// ListShiftSlotsUsecase handles the shift slot listing use case
type ListShiftSlotsUsecase struct {
	slotRepo       shift.ShiftSlotRepository
	assignmentRepo shift.ShiftAssignmentRepository
}

// NewListShiftSlotsUsecase creates a new ListShiftSlotsUsecase
func NewListShiftSlotsUsecase(
	slotRepo shift.ShiftSlotRepository,
	assignmentRepo shift.ShiftAssignmentRepository,
) *ListShiftSlotsUsecase {
	return &ListShiftSlotsUsecase{
		slotRepo:       slotRepo,
		assignmentRepo: assignmentRepo,
	}
}

// Execute retrieves shift slots for a business day with their assignment counts
func (uc *ListShiftSlotsUsecase) Execute(ctx context.Context, input ListShiftSlotsInput) ([]*ShiftSlotWithAssignmentCount, error) {
	// シフト枠一覧を取得
	slots, err := uc.slotRepo.FindByBusinessDayID(ctx, input.TenantID, input.BusinessDayID)
	if err != nil {
		return nil, err
	}

	// 各シフト枠の割り当て数を取得
	result := make([]*ShiftSlotWithAssignmentCount, 0, len(slots))
	for _, s := range slots {
		assignedCount, err := uc.assignmentRepo.CountConfirmedBySlotID(ctx, input.TenantID, s.SlotID())
		if err != nil {
			return nil, err
		}

		result = append(result, &ShiftSlotWithAssignmentCount{
			Slot:          s,
			AssignedCount: assignedCount,
		})
	}

	return result, nil
}

// GetShiftSlotInput represents the input for getting a shift slot
type GetShiftSlotInput struct {
	TenantID common.TenantID
	SlotID   shift.SlotID
}

// GetShiftSlotUsecase handles the shift slot retrieval use case
type GetShiftSlotUsecase struct {
	slotRepo       shift.ShiftSlotRepository
	assignmentRepo shift.ShiftAssignmentRepository
}

// NewGetShiftSlotUsecase creates a new GetShiftSlotUsecase
func NewGetShiftSlotUsecase(
	slotRepo shift.ShiftSlotRepository,
	assignmentRepo shift.ShiftAssignmentRepository,
) *GetShiftSlotUsecase {
	return &GetShiftSlotUsecase{
		slotRepo:       slotRepo,
		assignmentRepo: assignmentRepo,
	}
}

// Execute retrieves a shift slot by ID with its assignment count
func (uc *GetShiftSlotUsecase) Execute(ctx context.Context, input GetShiftSlotInput) (*ShiftSlotWithAssignmentCount, error) {
	// シフト枠の取得
	slot, err := uc.slotRepo.FindByID(ctx, input.TenantID, input.SlotID)
	if err != nil {
		return nil, err
	}

	// 割り当て数を取得
	assignedCount, err := uc.assignmentRepo.CountConfirmedBySlotID(ctx, input.TenantID, slot.SlotID())
	if err != nil {
		return nil, err
	}

	return &ShiftSlotWithAssignmentCount{
		Slot:          slot,
		AssignedCount: assignedCount,
	}, nil
}
