package event

import (
	"fmt"
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
)

// EventType represents the type of event
type EventType string

const (
	EventTypeNormal  EventType = "normal"  // 通常営業
	EventTypeSpecial EventType = "special" // 特別イベント
)

func (t EventType) Validate() error {
	switch t {
	case EventTypeNormal, EventTypeSpecial:
		return nil
	default:
		return fmt.Errorf("invalid event type: %s", t)
	}
}

// Event represents an event entity (aggregate root)
// イベントはVRChatイベントの定義を表す集約ルート
type Event struct {
	eventID     common.EventID
	tenantID    common.TenantID
	eventName   string
	eventType   EventType
	description string
	isActive    bool
	createdAt   time.Time
	updatedAt   time.Time
	deletedAt   *time.Time
}

// NewEvent creates a new Event entity
func NewEvent(
	tenantID common.TenantID,
	eventName string,
	eventType EventType,
	description string,
) (*Event, error) {
	event := &Event{
		eventID:     common.NewEventID(),
		tenantID:    tenantID,
		eventName:   eventName,
		eventType:   eventType,
		description: description,
		isActive:    true,
		createdAt:   time.Now(),
		updatedAt:   time.Now(),
	}

	if err := event.validate(); err != nil {
		return nil, err
	}

	return event, nil
}

// ReconstructEvent reconstructs an Event entity from persistence
func ReconstructEvent(
	eventID common.EventID,
	tenantID common.TenantID,
	eventName string,
	eventType EventType,
	description string,
	isActive bool,
	createdAt time.Time,
	updatedAt time.Time,
	deletedAt *time.Time,
) (*Event, error) {
	event := &Event{
		eventID:     eventID,
		tenantID:    tenantID,
		eventName:   eventName,
		eventType:   eventType,
		description: description,
		isActive:    isActive,
		createdAt:   createdAt,
		updatedAt:   updatedAt,
		deletedAt:   deletedAt,
	}

	if err := event.validate(); err != nil {
		return nil, err
	}

	return event, nil
}

// validate checks invariants
func (e *Event) validate() error {
	// TenantID の必須性チェック
	if err := e.tenantID.Validate(); err != nil {
		return common.NewValidationError("tenant_id is required", err)
	}

	// EventName の必須性チェック
	if e.eventName == "" {
		return common.NewValidationError("event_name is required", nil)
	}

	if len(e.eventName) > 255 {
		return common.NewValidationError("event_name must be less than 255 characters", nil)
	}

	// EventType のバリデーション
	if err := e.eventType.Validate(); err != nil {
		return common.NewValidationError("invalid event_type", err)
	}

	return nil
}

// Getters

func (e *Event) EventID() common.EventID {
	return e.eventID
}

func (e *Event) TenantID() common.TenantID {
	return e.tenantID
}

func (e *Event) EventName() string {
	return e.eventName
}

func (e *Event) EventType() EventType {
	return e.eventType
}

func (e *Event) Description() string {
	return e.description
}

func (e *Event) IsActive() bool {
	return e.isActive
}

func (e *Event) CreatedAt() time.Time {
	return e.createdAt
}

func (e *Event) UpdatedAt() time.Time {
	return e.updatedAt
}

func (e *Event) DeletedAt() *time.Time {
	return e.deletedAt
}

func (e *Event) IsDeleted() bool {
	return e.deletedAt != nil
}

// UpdateEventName updates the event name
func (e *Event) UpdateEventName(eventName string) error {
	if eventName == "" {
		return common.NewValidationError("event_name is required", nil)
	}
	if len(eventName) > 255 {
		return common.NewValidationError("event_name must be less than 255 characters", nil)
	}

	e.eventName = eventName
	e.updatedAt = time.Now()
	return nil
}

// UpdateDescription updates the description
func (e *Event) UpdateDescription(description string) {
	e.description = description
	e.updatedAt = time.Now()
}

// Activate activates the event
func (e *Event) Activate() {
	e.isActive = true
	e.updatedAt = time.Now()
}

// Deactivate deactivates the event
func (e *Event) Deactivate() {
	e.isActive = false
	e.updatedAt = time.Now()
}

// Delete marks the event as deleted (soft delete)
func (e *Event) Delete() {
	now := time.Now()
	e.deletedAt = &now
	e.updatedAt = now
}

