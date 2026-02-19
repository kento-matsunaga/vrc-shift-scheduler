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

// DeleteShiftSlotInput represents the input for deleting a shift slot
type DeleteShiftSlotInput struct {
	TenantID common.TenantID
	SlotID   shift.SlotID
}

// DeleteShiftSlotUsecase handles the shift slot deletion use case
type DeleteShiftSlotUsecase struct {
	slotRepo       shift.ShiftSlotRepository
	assignmentRepo shift.ShiftAssignmentRepository
}

// NewDeleteShiftSlotUsecase creates a new DeleteShiftSlotUsecase
func NewDeleteShiftSlotUsecase(
	slotRepo shift.ShiftSlotRepository,
	assignmentRepo shift.ShiftAssignmentRepository,
) *DeleteShiftSlotUsecase {
	return &DeleteShiftSlotUsecase{
		slotRepo:       slotRepo,
		assignmentRepo: assignmentRepo,
	}
}

// Execute deletes a shift slot if no assignments exist
func (uc *DeleteShiftSlotUsecase) Execute(ctx context.Context, input DeleteShiftSlotInput) error {
	// シフト枠の存在確認
	slot, err := uc.slotRepo.FindByID(ctx, input.TenantID, input.SlotID)
	if err != nil {
		return err
	}

	// 割り当てが存在するかチェック
	assignedCount, err := uc.assignmentRepo.CountConfirmedBySlotID(ctx, input.TenantID, slot.SlotID())
	if err != nil {
		return err
	}

	if assignedCount > 0 {
		return common.NewConflictError("割り当てが存在するシフト枠は削除できません")
	}

	// ソフトデリート
	now := time.Now()
	slot.Delete(now)
	if err := uc.slotRepo.Save(ctx, slot); err != nil {
		return err
	}

	return nil
}

// DeleteSlotsByInstanceInput represents the input for bulk deleting shift slots by instance
type DeleteSlotsByInstanceInput struct {
	TenantID      common.TenantID
	BusinessDayID event.BusinessDayID
	InstanceID    shift.InstanceID
}

// DeleteSlotsByInstanceResult represents the result of checking if slots can be deleted
type DeleteSlotsByInstanceResult struct {
	CanDelete      bool
	SlotCount      int
	AssignedSlots  int
	BlockingReason string
}

// DeleteSlotsByInstanceUsecase handles bulk deletion of shift slots by business day and instance
type DeleteSlotsByInstanceUsecase struct {
	txManager      TxManager
	slotRepo       shift.ShiftSlotRepository
	assignmentRepo shift.ShiftAssignmentRepository
}

// NewDeleteSlotsByInstanceUsecase creates a new DeleteSlotsByInstanceUsecase
func NewDeleteSlotsByInstanceUsecase(
	txManager TxManager,
	slotRepo shift.ShiftSlotRepository,
	assignmentRepo shift.ShiftAssignmentRepository,
) *DeleteSlotsByInstanceUsecase {
	return &DeleteSlotsByInstanceUsecase{
		txManager:      txManager,
		slotRepo:       slotRepo,
		assignmentRepo: assignmentRepo,
	}
}

// CheckDeletable checks if slots can be deleted and returns details
func (uc *DeleteSlotsByInstanceUsecase) CheckDeletable(ctx context.Context, input DeleteSlotsByInstanceInput) (*DeleteSlotsByInstanceResult, error) {
	// 営業日+インスタンスに紐づくシフト枠を取得
	slots, err := uc.slotRepo.FindByBusinessDayIDAndInstanceID(ctx, input.TenantID, input.BusinessDayID, input.InstanceID)
	if err != nil {
		return nil, err
	}

	if len(slots) == 0 {
		return &DeleteSlotsByInstanceResult{
			CanDelete: true,
			SlotCount: 0,
		}, nil
	}

	// 各シフト枠に担当があるかチェック
	assignedSlots := 0
	for _, slot := range slots {
		count, err := uc.assignmentRepo.CountConfirmedBySlotID(ctx, input.TenantID, slot.SlotID())
		if err != nil {
			return nil, err
		}
		if count > 0 {
			assignedSlots++
		}
	}

	if assignedSlots > 0 {
		return &DeleteSlotsByInstanceResult{
			CanDelete:      false,
			SlotCount:      len(slots),
			AssignedSlots:  assignedSlots,
			BlockingReason: "cannot delete: some shift slots have assignments",
		}, nil
	}

	return &DeleteSlotsByInstanceResult{
		CanDelete:     true,
		SlotCount:     len(slots),
		AssignedSlots: 0,
	}, nil
}

// Execute deletes all shift slots for a business day and instance
func (uc *DeleteSlotsByInstanceUsecase) Execute(ctx context.Context, input DeleteSlotsByInstanceInput) error {
	// 削除可能かチェック（トランザクション外で実行）
	result, err := uc.CheckDeletable(ctx, input)
	if err != nil {
		return err
	}

	if !result.CanDelete {
		return common.NewConflictError(result.BlockingReason)
	}

	// トランザクション内で削除処理を実行
	return uc.txManager.WithTx(ctx, func(txCtx context.Context) error {
		// 営業日+インスタンスに紐づくシフト枠を取得
		slots, err := uc.slotRepo.FindByBusinessDayIDAndInstanceID(txCtx, input.TenantID, input.BusinessDayID, input.InstanceID)
		if err != nil {
			return err
		}

		// シフト枠をソフトデリート
		now := time.Now()
		for _, slot := range slots {
			slot.Delete(now)
			if err := uc.slotRepo.Save(txCtx, slot); err != nil {
				return err
			}
		}

		return nil
	})
}
