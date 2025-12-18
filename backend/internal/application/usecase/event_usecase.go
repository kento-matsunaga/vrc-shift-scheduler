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
	TenantID    common.TenantID
	EventName   string
	EventType   event.EventType
	Description string
}

// CreateEventUsecase handles the event creation use case
type CreateEventUsecase struct {
	eventRepo EventRepository
}

// NewCreateEventUsecase creates a new CreateEventUsecase
func NewCreateEventUsecase(eventRepo EventRepository) *CreateEventUsecase {
	return &CreateEventUsecase{
		eventRepo: eventRepo,
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
	newEvent, err := event.NewEvent(time.Now(), input.TenantID, input.EventName, input.EventType, input.Description)
	if err != nil {
		return nil, err
	}

	// 保存
	if err := uc.eventRepo.Save(ctx, newEvent); err != nil {
		return nil, err
	}

	return newEvent, nil
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
