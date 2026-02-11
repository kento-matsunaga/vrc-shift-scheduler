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
	now time.Time,
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

	if err := template.validate(); err != nil {
		return nil, err
	}

	return template, nil
}

// NewShiftSlotTemplateItem creates a new shift slot template item
func NewShiftSlotTemplateItem(
	now time.Time,
	templateID common.ShiftSlotTemplateID,
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

	if slotName == "" {
		return nil, fmt.Errorf("slot_name is required")
	}

	if requiredCount <= 0 {
		return nil, fmt.Errorf("required_count must be positive")
	}

	if priority < 1 {
		return nil, fmt.Errorf("priority must be at least 1")
	}

	item := &ShiftSlotTemplateItem{
		itemID:        common.NewShiftSlotTemplateItemID(),
		templateID:    templateID,
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

// ReconstructShiftSlotTemplate reconstructs a shift slot template from persistence
func ReconstructShiftSlotTemplate(
	templateID common.ShiftSlotTemplateID,
	tenantID common.TenantID,
	eventID common.EventID,
	templateName string,
	description string,
	items []*ShiftSlotTemplateItem,
	createdAt time.Time,
	updatedAt time.Time,
	deletedAt *time.Time,
) (*ShiftSlotTemplate, error) {
	t := &ShiftSlotTemplate{
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

	if err := t.validate(); err != nil {
		return nil, err
	}

	return t, nil
}

// ReconstructShiftSlotTemplateItem reconstructs a template item from persistence
func ReconstructShiftSlotTemplateItem(
	itemID common.ShiftSlotTemplateItemID,
	templateID common.ShiftSlotTemplateID,
	slotName string,
	instanceName string,
	startTime time.Time,
	endTime time.Time,
	requiredCount int,
	priority int,
	createdAt time.Time,
	updatedAt time.Time,
) (*ShiftSlotTemplateItem, error) {
	if slotName == "" {
		return nil, fmt.Errorf("slot_name is required")
	}
	return &ShiftSlotTemplateItem{
		itemID:        itemID,
		templateID:    templateID,
		slotName:      slotName,
		instanceName:  instanceName,
		startTime:     startTime,
		endTime:       endTime,
		requiredCount: requiredCount,
		priority:      priority,
		createdAt:     createdAt,
		updatedAt:     updatedAt,
	}, nil
}

// UpdateDetails updates the template details
func (t *ShiftSlotTemplate) UpdateDetails(now time.Time, templateName, description string, items []*ShiftSlotTemplateItem) error {
	// Validate before mutating using a temporary copy
	tmp := *t
	tmp.templateName = templateName
	tmp.description = description
	tmp.items = items
	tmp.updatedAt = now
	if err := tmp.validate(); err != nil {
		return err
	}

	if len(items) == 0 {
		return fmt.Errorf("at least one template item is required")
	}

	// Apply validated changes
	t.templateName = templateName
	t.description = description
	t.items = items
	t.updatedAt = now

	return nil
}

func (t *ShiftSlotTemplate) validate() error {
	if t.templateName == "" {
		return fmt.Errorf("template_name is required")
	}
	if len(t.templateName) > 100 {
		return fmt.Errorf("template_name must be 100 characters or less")
	}
	return nil
}

// Delete soft-deletes the template
func (t *ShiftSlotTemplate) Delete(now time.Time) {
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
