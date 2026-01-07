package event

import (
	"context"
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/event"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/shift"
)

// CreateBusinessDayInput represents the input for creating a business day
type CreateBusinessDayInput struct {
	TenantID       common.TenantID
	EventID        common.EventID
	TargetDate     time.Time
	StartTime      time.Time
	EndTime        time.Time
	OccurrenceType event.OccurrenceType
	TemplateID     *common.ShiftSlotTemplateID // optional
}

// CreateBusinessDayUsecase handles the business day creation use case
type CreateBusinessDayUsecase struct {
	businessDayRepo event.EventBusinessDayRepository
	eventRepo       event.EventRepository
	templateRepo    shift.ShiftSlotTemplateRepository
	slotRepo        shift.ShiftSlotRepository
}

// NewCreateBusinessDayUsecase creates a new CreateBusinessDayUsecase
func NewCreateBusinessDayUsecase(
	businessDayRepo event.EventBusinessDayRepository,
	eventRepo event.EventRepository,
	templateRepo shift.ShiftSlotTemplateRepository,
	slotRepo shift.ShiftSlotRepository,
) *CreateBusinessDayUsecase {
	return &CreateBusinessDayUsecase{
		businessDayRepo: businessDayRepo,
		eventRepo:       eventRepo,
		templateRepo:    templateRepo,
		slotRepo:        slotRepo,
	}
}

// Execute creates a new business day
func (uc *CreateBusinessDayUsecase) Execute(ctx context.Context, input CreateBusinessDayInput) (*event.EventBusinessDay, error) {
	// イベントの存在確認
	_, err := uc.eventRepo.FindByID(ctx, input.TenantID, input.EventID)
	if err != nil {
		return nil, err
	}

	// 重複チェック
	exists, err := uc.businessDayRepo.ExistsByEventIDAndDate(ctx, input.TenantID, input.EventID, input.TargetDate, input.StartTime)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, common.NewConflictError("Business day already exists for this date and time")
	}

	// BusinessDay の作成
	newBusinessDay, err := event.NewEventBusinessDay(
		time.Now(),
		input.TenantID,
		input.EventID,
		input.TargetDate,
		input.StartTime,
		input.EndTime,
		input.OccurrenceType,
		nil, // recurring_pattern_id は手動作成では nil
	)
	if err != nil {
		return nil, err
	}

	// 保存
	if err := uc.businessDayRepo.Save(ctx, newBusinessDay); err != nil {
		return nil, err
	}

	// テンプレートが指定されている場合、テンプレートからシフト枠を作成
	if input.TemplateID != nil {
		template, err := uc.templateRepo.FindByID(ctx, input.TenantID, *input.TemplateID)
		if err != nil {
			return nil, err
		}

		if err := uc.createShiftSlotsFromTemplate(ctx, newBusinessDay, template); err != nil {
			return nil, err
		}
	}

	return newBusinessDay, nil
}

// createShiftSlotsFromTemplate creates shift slots from a template for a business day
func (uc *CreateBusinessDayUsecase) createShiftSlotsFromTemplate(ctx context.Context, businessDay *event.EventBusinessDay, template *shift.ShiftSlotTemplate) error {
	// テンプレートの各アイテムからシフト枠を作成
	for _, item := range template.Items() {
		// テンプレートの時刻を営業日の日付と組み合わせてDateTimeを作成
		startDateTime := time.Date(
			businessDay.TargetDate().Year(),
			businessDay.TargetDate().Month(),
			businessDay.TargetDate().Day(),
			item.StartTime().Hour(),
			item.StartTime().Minute(),
			item.StartTime().Second(),
			0,
			time.Local,
		)

		endDateTime := time.Date(
			businessDay.TargetDate().Year(),
			businessDay.TargetDate().Month(),
			businessDay.TargetDate().Day(),
			item.EndTime().Hour(),
			item.EndTime().Minute(),
			item.EndTime().Second(),
			0,
			time.Local,
		)

		// シフト枠を作成
		shiftSlot, err := shift.NewShiftSlot(
			time.Now(),
			businessDay.TenantID(),
			businessDay.BusinessDayID(),
			item.SlotName(),
			item.InstanceName(),
			startDateTime,
			endDateTime,
			item.RequiredCount(),
			item.Priority(),
		)
		if err != nil {
			return err
		}

		// シフト枠を保存
		if err := uc.slotRepo.Save(ctx, shiftSlot); err != nil {
			return err
		}
	}

	return nil
}

// ListBusinessDaysInput represents the input for listing business days
type ListBusinessDaysInput struct {
	TenantID  common.TenantID
	EventID   common.EventID
	StartDate *time.Time
	EndDate   *time.Time
}

// ListBusinessDaysUsecase handles the business day listing use case
type ListBusinessDaysUsecase struct {
	businessDayRepo event.EventBusinessDayRepository
}

// NewListBusinessDaysUsecase creates a new ListBusinessDaysUsecase
func NewListBusinessDaysUsecase(businessDayRepo event.EventBusinessDayRepository) *ListBusinessDaysUsecase {
	return &ListBusinessDaysUsecase{
		businessDayRepo: businessDayRepo,
	}
}

// Execute retrieves business days for an event
func (uc *ListBusinessDaysUsecase) Execute(ctx context.Context, input ListBusinessDaysInput) ([]*event.EventBusinessDay, error) {
	var businessDays []*event.EventBusinessDay
	var err error

	if input.StartDate != nil && input.EndDate != nil {
		// 日付範囲で検索
		businessDays, err = uc.businessDayRepo.FindByEventIDAndDateRange(ctx, input.TenantID, input.EventID, *input.StartDate, *input.EndDate)
	} else {
		// 全件取得
		businessDays, err = uc.businessDayRepo.FindByEventID(ctx, input.TenantID, input.EventID)
	}

	if err != nil {
		return nil, err
	}

	return businessDays, nil
}

// GetBusinessDayInput represents the input for getting a business day
type GetBusinessDayInput struct {
	TenantID      common.TenantID
	BusinessDayID event.BusinessDayID
}

// GetBusinessDayUsecase handles the business day retrieval use case
type GetBusinessDayUsecase struct {
	businessDayRepo event.EventBusinessDayRepository
}

// NewGetBusinessDayUsecase creates a new GetBusinessDayUsecase
func NewGetBusinessDayUsecase(businessDayRepo event.EventBusinessDayRepository) *GetBusinessDayUsecase {
	return &GetBusinessDayUsecase{
		businessDayRepo: businessDayRepo,
	}
}

// Execute retrieves a business day by ID
func (uc *GetBusinessDayUsecase) Execute(ctx context.Context, input GetBusinessDayInput) (*event.EventBusinessDay, error) {
	foundBusinessDay, err := uc.businessDayRepo.FindByID(ctx, input.TenantID, input.BusinessDayID)
	if err != nil {
		return nil, err
	}

	return foundBusinessDay, nil
}

// ApplyTemplateInput represents the input for applying a template to a business day
type ApplyTemplateInput struct {
	TenantID      common.TenantID
	BusinessDayID event.BusinessDayID
	TemplateID    common.ShiftSlotTemplateID
}

// ApplyTemplateUsecase handles applying a template to an existing business day
type ApplyTemplateUsecase struct {
	businessDayRepo event.EventBusinessDayRepository
	templateRepo    shift.ShiftSlotTemplateRepository
	slotRepo        shift.ShiftSlotRepository
}

// NewApplyTemplateUsecase creates a new ApplyTemplateUsecase
func NewApplyTemplateUsecase(
	businessDayRepo event.EventBusinessDayRepository,
	templateRepo shift.ShiftSlotTemplateRepository,
	slotRepo shift.ShiftSlotRepository,
) *ApplyTemplateUsecase {
	return &ApplyTemplateUsecase{
		businessDayRepo: businessDayRepo,
		templateRepo:    templateRepo,
		slotRepo:        slotRepo,
	}
}

// Execute applies a template to an existing business day
func (uc *ApplyTemplateUsecase) Execute(ctx context.Context, input ApplyTemplateInput) (int, error) {
	// 営業日の取得
	businessDay, err := uc.businessDayRepo.FindByID(ctx, input.TenantID, input.BusinessDayID)
	if err != nil {
		return 0, err
	}

	// テンプレートの取得
	template, err := uc.templateRepo.FindByID(ctx, input.TenantID, input.TemplateID)
	if err != nil {
		return 0, err
	}

	// テンプレートからシフト枠を作成
	if err := uc.createShiftSlotsFromTemplate(ctx, businessDay, template); err != nil {
		return 0, err
	}

	return len(template.Items()), nil
}

// createShiftSlotsFromTemplate creates shift slots from a template for a business day
func (uc *ApplyTemplateUsecase) createShiftSlotsFromTemplate(ctx context.Context, businessDay *event.EventBusinessDay, template *shift.ShiftSlotTemplate) error {
	// テンプレートの各アイテムからシフト枠を作成
	for _, item := range template.Items() {
		// テンプレートの時刻を営業日の日付と組み合わせてDateTimeを作成
		startDateTime := time.Date(
			businessDay.TargetDate().Year(),
			businessDay.TargetDate().Month(),
			businessDay.TargetDate().Day(),
			item.StartTime().Hour(),
			item.StartTime().Minute(),
			item.StartTime().Second(),
			0,
			time.Local,
		)

		endDateTime := time.Date(
			businessDay.TargetDate().Year(),
			businessDay.TargetDate().Month(),
			businessDay.TargetDate().Day(),
			item.EndTime().Hour(),
			item.EndTime().Minute(),
			item.EndTime().Second(),
			0,
			time.Local,
		)

		// シフト枠を作成
		shiftSlot, err := shift.NewShiftSlot(
			time.Now(),
			businessDay.TenantID(),
			businessDay.BusinessDayID(),
			item.SlotName(),
			item.InstanceName(),
			startDateTime,
			endDateTime,
			item.RequiredCount(),
			item.Priority(),
		)
		if err != nil {
			return err
		}

		// シフト枠を保存
		if err := uc.slotRepo.Save(ctx, shiftSlot); err != nil {
			return err
		}
	}

	return nil
}
