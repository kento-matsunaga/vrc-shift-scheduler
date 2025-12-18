package shift

import (
	"fmt"
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
)

// ShiftSlotTemplate represents a template for shift slots
type ShiftSlotTemplate struct {
	templateID   common.ShiftSlotTemplateID
	tenantID     common.TenantID
	eventID      common.EventID
	templateName string
	description  string
	items        []*ShiftSlotTemplateItem
	createdAt    time.Time
	updatedAt    time.Time
	deletedAt    *time.Time
}

// ShiftSlotTemplateItem represents an individual shift slot in a template
type ShiftSlotTemplateItem struct {
	itemID        common.ShiftSlotTemplateItemID
	templateID    common.ShiftSlotTemplateID
	positionID    PositionID
	slotName      string
	instanceName  string
	startTime     time.Time
	endTime       time.Time
	requiredCount int
	priority      int
	createdAt     time.Time
	updatedAt     time.Time
}

// NewShiftSlotTemplate creates a new shift slot template
func NewShiftSlotTemplate(
	tenantID common.TenantID,
	eventID common.EventID,
	templateName string,
	description string,
	items []*ShiftSlotTemplateItem,
) (*ShiftSlotTemplate, error) {
	if err := tenantID.Validate(); err != nil {
		return nil, fmt.Errorf("invalid tenant_id: %w", err)
	}

	if err := eventID.Validate(); err != nil {
		return nil, fmt.Errorf("invalid event_id: %w", err)
	}

	if templateName == "" {
		return nil, fmt.Errorf("template_name is required")
	}

	if len(templateName) > 100 {
		return nil, fmt.Errorf("template_name must be 100 characters or less")
	}

	// Note: items can be empty initially and added later via UpdateDetails
	// if len(items) == 0 {
	// 	return nil, fmt.Errorf("at least one template item is required")
	// }

	now := time.Now()
	template := &ShiftSlotTemplate{
		templateID:   common.NewShiftSlotTemplateID(),
		tenantID:     tenantID,
		eventID:      eventID,
		templateName: templateName,
		description:  description,
		items:        items,
		createdAt:    now,
		updatedAt:    now,
	}

	return template, nil
}

// NewShiftSlotTemplateItem creates a new shift slot template item
func NewShiftSlotTemplateItem(
	templateID common.ShiftSlotTemplateID,
	positionID PositionID,
	slotName string,
	instanceName string,
	startTime time.Time,
	endTime time.Time,
	requiredCount int,
	priority int,
) (*ShiftSlotTemplateItem, error) {
	if err := templateID.Validate(); err != nil {
		return nil, fmt.Errorf("invalid template_id: %w", err)
	}

	if err := positionID.Validate(); err != nil {
		return nil, fmt.Errorf("invalid position_id: %w", err)
	}

	if slotName == "" {
		return nil, fmt.Errorf("slot_name is required")
	}

	if requiredCount <= 0 {
		return nil, fmt.Errorf("required_count must be positive")
	}

	now := time.Now()
	item := &ShiftSlotTemplateItem{
		itemID:        common.NewShiftSlotTemplateItemID(),
		templateID:    templateID,
		positionID:    positionID,
		slotName:      slotName,
		instanceName:  instanceName,
		startTime:     startTime,
		endTime:       endTime,
		requiredCount: requiredCount,
		priority:      priority,
		createdAt:     now,
		updatedAt:     now,
	}

	return item, nil
}

// ReconstituteShiftSlotTemplate reconstitutes a shift slot template from persistence
func ReconstituteShiftSlotTemplate(
	templateID common.ShiftSlotTemplateID,
	tenantID common.TenantID,
	eventID common.EventID,
	templateName string,
	description string,
	items []*ShiftSlotTemplateItem,
	createdAt time.Time,
	updatedAt time.Time,
	deletedAt *time.Time,
) *ShiftSlotTemplate {
	return &ShiftSlotTemplate{
		templateID:   templateID,
		tenantID:     tenantID,
		eventID:      eventID,
		templateName: templateName,
		description:  description,
		items:        items,
		createdAt:    createdAt,
		updatedAt:    updatedAt,
		deletedAt:    deletedAt,
	}
}

// ReconstituteShiftSlotTemplateItem reconstitutes a template item from persistence
func ReconstituteShiftSlotTemplateItem(
	itemID common.ShiftSlotTemplateItemID,
	templateID common.ShiftSlotTemplateID,
	positionID PositionID,
	slotName string,
	instanceName string,
	startTime time.Time,
	endTime time.Time,
	requiredCount int,
	priority int,
	createdAt time.Time,
	updatedAt time.Time,
) *ShiftSlotTemplateItem {
	return &ShiftSlotTemplateItem{
		itemID:        itemID,
		templateID:    templateID,
		positionID:    positionID,
		slotName:      slotName,
		instanceName:  instanceName,
		startTime:     startTime,
		endTime:       endTime,
		requiredCount: requiredCount,
		priority:      priority,
		createdAt:     createdAt,
		updatedAt:     updatedAt,
	}
}

// UpdateDetails updates the template details
func (t *ShiftSlotTemplate) UpdateDetails(templateName, description string, items []*ShiftSlotTemplateItem) error {
	if templateName == "" {
		return fmt.Errorf("template_name is required")
	}

	if len(templateName) > 100 {
		return fmt.Errorf("template_name must be 100 characters or less")
	}

	if len(items) == 0 {
		return fmt.Errorf("at least one template item is required")
	}

	t.templateName = templateName
	t.description = description
	t.items = items
	t.updatedAt = time.Now()

	return nil
}

// Delete soft-deletes the template
func (t *ShiftSlotTemplate) Delete() {
	now := time.Now()
	t.deletedAt = &now
	t.updatedAt = now
}

// Getters
func (t *ShiftSlotTemplate) TemplateID() common.ShiftSlotTemplateID {
	return t.templateID
}

func (t *ShiftSlotTemplate) TenantID() common.TenantID {
	return t.tenantID
}

func (t *ShiftSlotTemplate) EventID() common.EventID {
	return t.eventID
}

func (t *ShiftSlotTemplate) TemplateName() string {
	return t.templateName
}

func (t *ShiftSlotTemplate) Description() string {
	return t.description
}

func (t *ShiftSlotTemplate) Items() []*ShiftSlotTemplateItem {
	return t.items
}

func (t *ShiftSlotTemplate) CreatedAt() time.Time {
	return t.createdAt
}

func (t *ShiftSlotTemplate) UpdatedAt() time.Time {
	return t.updatedAt
}

func (t *ShiftSlotTemplate) DeletedAt() *time.Time {
	return t.deletedAt
}

// Template Item Getters
func (i *ShiftSlotTemplateItem) ItemID() common.ShiftSlotTemplateItemID {
	return i.itemID
}

func (i *ShiftSlotTemplateItem) TemplateID() common.ShiftSlotTemplateID {
	return i.templateID
}

func (i *ShiftSlotTemplateItem) PositionID() PositionID {
	return i.positionID
}

func (i *ShiftSlotTemplateItem) SlotName() string {
	return i.slotName
}

func (i *ShiftSlotTemplateItem) InstanceName() string {
	return i.instanceName
}

func (i *ShiftSlotTemplateItem) StartTime() time.Time {
	return i.startTime
}

func (i *ShiftSlotTemplateItem) EndTime() time.Time {
	return i.endTime
}

func (i *ShiftSlotTemplateItem) RequiredCount() int {
	return i.requiredCount
}

func (i *ShiftSlotTemplateItem) Priority() int {
	return i.priority
}

func (i *ShiftSlotTemplateItem) CreatedAt() time.Time {
	return i.createdAt
}

func (i *ShiftSlotTemplateItem) UpdatedAt() time.Time {
	return i.updatedAt
}
