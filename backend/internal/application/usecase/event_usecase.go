package usecase

import (
	"context"
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/event"
)

// EventRepository defines the interface for event persistence
type EventRepository interface {
	Save(ctx context.Context, event *event.Event) error
	FindByID(ctx context.Context, tenantID common.TenantID, eventID common.EventID) (*event.Event, error)
	FindByTenantID(ctx context.Context, tenantID common.TenantID) ([]*event.Event, error)
	ExistsByName(ctx context.Context, tenantID common.TenantID, eventName string) (bool, error)
}

// CreateEventInput represents the input for creating an event
type CreateEventInput struct {
	TenantID            common.TenantID
	EventName           string
	EventType           event.EventType
	Description         string
	RecurrenceType      event.RecurrenceType
	RecurrenceStartDate *time.Time
	RecurrenceDayOfWeek *int
	DefaultStartTime    *time.Time
	DefaultEndTime      *time.Time
}

// CreateEventUsecase handles the event creation use case
type CreateEventUsecase struct {
	eventRepo       EventRepository
	businessDayRepo EventBusinessDayRepository
}

// NewCreateEventUsecase creates a new CreateEventUsecase
func NewCreateEventUsecase(eventRepo EventRepository, businessDayRepo EventBusinessDayRepository) *CreateEventUsecase {
	return &CreateEventUsecase{
		eventRepo:       eventRepo,
		businessDayRepo: businessDayRepo,
	}
}

// Execute creates a new event
func (uc *CreateEventUsecase) Execute(ctx context.Context, input CreateEventInput) (*event.Event, error) {
	// イベント名の重複チェック
	exists, err := uc.eventRepo.ExistsByName(ctx, input.TenantID, input.EventName)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, common.NewConflictError("Event with this name already exists")
	}

	// イベントの作成
	newEvent, err := event.NewEvent(
		time.Now(),
		input.TenantID,
		input.EventName,
		input.EventType,
		input.Description,
		input.RecurrenceType,
		input.RecurrenceStartDate,
		input.RecurrenceDayOfWeek,
		input.DefaultStartTime,
		input.DefaultEndTime,
	)
	if err != nil {
		return nil, err
	}

	// 保存
	if err := uc.eventRepo.Save(ctx, newEvent); err != nil {
		return nil, err
	}

	// 定期設定がある場合、営業日を自動生成
	if newEvent.HasRecurrence() {
		if err := uc.generateBusinessDays(ctx, newEvent); err != nil {
			return nil, err
		}
	}

	return newEvent, nil
}

// generateBusinessDays generates business days for recurring events
// 今月と来月末までの営業日を自動生成
func (uc *CreateEventUsecase) generateBusinessDays(ctx context.Context, e *event.Event) error {
	if !e.HasRecurrence() {
		return nil
	}

	if e.RecurrenceStartDate() == nil || e.RecurrenceDayOfWeek() == nil ||
		e.DefaultStartTime() == nil || e.DefaultEndTime() == nil {
		return common.NewValidationError("recurrence fields are incomplete", nil)
	}

	now := time.Now()
	startDate := *e.RecurrenceStartDate()
	targetDayOfWeek := time.Weekday(*e.RecurrenceDayOfWeek())

	// 今月の最初の日と来月末の日を計算
	currentMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.Local)
	nextMonthEnd := currentMonth.AddDate(0, 2, 0).AddDate(0, 0, -1)

	// 定期開始日から次の指定曜日を見つける
	candidateDate := startDate
	for candidateDate.Weekday() != targetDayOfWeek {
		candidateDate = candidateDate.AddDate(0, 0, 1)
	}

	// 営業日を生成
	interval := 7 // 毎週の場合は7日間隔
	if e.RecurrenceType() == event.RecurrenceTypeBiweekly {
		interval = 14 // 隔週の場合は14日間隔
	}

	for candidateDate.Before(nextMonthEnd) || candidateDate.Equal(nextMonthEnd) {
		// 開始日より前の日付はスキップ
		if candidateDate.Before(startDate) {
			candidateDate = candidateDate.AddDate(0, 0, interval)
			continue
		}

		// 重複チェック
		exists, err := uc.businessDayRepo.ExistsByEventIDAndDate(
			ctx,
			e.TenantID(),
			e.EventID(),
			candidateDate,
			*e.DefaultStartTime(),
		)
		if err != nil {
			return err
		}

		if !exists {
			// 営業日を作成（普通営業 = recurring）
			businessDay, err := event.NewEventBusinessDay(
				time.Now(),
				e.TenantID(),
				e.EventID(),
				candidateDate,
				*e.DefaultStartTime(),
				*e.DefaultEndTime(),
				event.OccurrenceTypeRecurring,
				nil, // recurring_pattern_idはnilで良い（イベント自体が定期情報を持っているため）
			)
			if err != nil {
				return err
			}

			// 保存
			if err := uc.businessDayRepo.Save(ctx, businessDay); err != nil {
				return err
			}
		}

		candidateDate = candidateDate.AddDate(0, 0, interval)
	}

	return nil
}

// ListEventsInput represents the input for listing events
type ListEventsInput struct {
	TenantID common.TenantID
}

// ListEventsUsecase handles the event listing use case
type ListEventsUsecase struct {
	eventRepo EventRepository
}

// NewListEventsUsecase creates a new ListEventsUsecase
func NewListEventsUsecase(eventRepo EventRepository) *ListEventsUsecase {
	return &ListEventsUsecase{
		eventRepo: eventRepo,
	}
}

// Execute retrieves all events for a tenant
func (uc *ListEventsUsecase) Execute(ctx context.Context, input ListEventsInput) ([]*event.Event, error) {
	events, err := uc.eventRepo.FindByTenantID(ctx, input.TenantID)
	if err != nil {
		return nil, err
	}

	return events, nil
}

// GetEventInput represents the input for getting an event
type GetEventInput struct {
	TenantID common.TenantID
	EventID  common.EventID
}

// GetEventUsecase handles the event retrieval use case
type GetEventUsecase struct {
	eventRepo EventRepository
}

// NewGetEventUsecase creates a new GetEventUsecase
func NewGetEventUsecase(eventRepo EventRepository) *GetEventUsecase {
	return &GetEventUsecase{
		eventRepo: eventRepo,
	}
}

// Execute retrieves an event by ID
func (uc *GetEventUsecase) Execute(ctx context.Context, input GetEventInput) (*event.Event, error) {
	foundEvent, err := uc.eventRepo.FindByID(ctx, input.TenantID, input.EventID)
	if err != nil {
		return nil, err
	}

	return foundEvent, nil
}

// GenerateBusinessDaysInput represents the input for generating business days
type GenerateBusinessDaysInput struct {
	TenantID common.TenantID
	EventID  common.EventID
}

// GenerateBusinessDaysOutput represents the output of generating business days
type GenerateBusinessDaysOutput struct {
	GeneratedCount int
	Event          *event.Event
}

// GenerateBusinessDaysUsecase handles generating business days for recurring events
type GenerateBusinessDaysUsecase struct {
	eventRepo       EventRepository
	businessDayRepo EventBusinessDayRepository
}

// NewGenerateBusinessDaysUsecase creates a new GenerateBusinessDaysUsecase
func NewGenerateBusinessDaysUsecase(eventRepo EventRepository, businessDayRepo EventBusinessDayRepository) *GenerateBusinessDaysUsecase {
	return &GenerateBusinessDaysUsecase{
		eventRepo:       eventRepo,
		businessDayRepo: businessDayRepo,
	}
}

// Execute generates business days for a recurring event
func (uc *GenerateBusinessDaysUsecase) Execute(ctx context.Context, input GenerateBusinessDaysInput) (*GenerateBusinessDaysOutput, error) {
	// イベントを取得
	e, err := uc.eventRepo.FindByID(ctx, input.TenantID, input.EventID)
	if err != nil {
		return nil, err
	}

	// 定期設定がない場合はエラー
	if !e.HasRecurrence() {
		return nil, common.NewValidationError("Event does not have recurrence settings", nil)
	}

	// 営業日を生成
	generatedCount, err := uc.generateBusinessDays(ctx, e)
	if err != nil {
		return nil, err
	}

	return &GenerateBusinessDaysOutput{
		GeneratedCount: generatedCount,
		Event:          e,
	}, nil
}

// generateBusinessDays generates business days for recurring events
// 今月と来月末までの営業日を自動生成し、生成された件数を返す
func (uc *GenerateBusinessDaysUsecase) generateBusinessDays(ctx context.Context, e *event.Event) (int, error) {
	if !e.HasRecurrence() {
		return 0, nil
	}

	if e.RecurrenceStartDate() == nil || e.RecurrenceDayOfWeek() == nil ||
		e.DefaultStartTime() == nil || e.DefaultEndTime() == nil {
		return 0, common.NewValidationError("recurrence fields are incomplete", nil)
	}

	now := time.Now()
	startDate := *e.RecurrenceStartDate()
	targetDayOfWeek := time.Weekday(*e.RecurrenceDayOfWeek())

	// 今月の最初の日と来月末の日を計算
	currentMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.Local)
	nextMonthEnd := currentMonth.AddDate(0, 2, 0).AddDate(0, 0, -1)

	// 定期開始日から次の指定曜日を見つける
	candidateDate := startDate
	for candidateDate.Weekday() != targetDayOfWeek {
		candidateDate = candidateDate.AddDate(0, 0, 1)
	}

	// 営業日を生成
	interval := 7 // 毎週の場合は7日間隔
	if e.RecurrenceType() == event.RecurrenceTypeBiweekly {
		interval = 14 // 隔週の場合は14日間隔
	}

	generatedCount := 0

	for candidateDate.Before(nextMonthEnd) || candidateDate.Equal(nextMonthEnd) {
		// 開始日より前の日付はスキップ
		if candidateDate.Before(startDate) {
			candidateDate = candidateDate.AddDate(0, 0, interval)
			continue
		}

		// 重複チェック
		exists, err := uc.businessDayRepo.ExistsByEventIDAndDate(
			ctx,
			e.TenantID(),
			e.EventID(),
			candidateDate,
			*e.DefaultStartTime(),
		)
		if err != nil {
			return generatedCount, err
		}

		if !exists {
			// 営業日を作成（普通営業 = recurring）
			businessDay, err := event.NewEventBusinessDay(
				time.Now(),
				e.TenantID(),
				e.EventID(),
				candidateDate,
				*e.DefaultStartTime(),
				*e.DefaultEndTime(),
				event.OccurrenceTypeRecurring,
				nil,
			)
			if err != nil {
				return generatedCount, err
			}

			// 保存
			if err := uc.businessDayRepo.Save(ctx, businessDay); err != nil {
				return generatedCount, err
			}

			generatedCount++
		}

		candidateDate = candidateDate.AddDate(0, 0, interval)
	}

	return generatedCount, nil
}
