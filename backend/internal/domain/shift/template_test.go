package shift_test

import (
	"strings"
	"testing"
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/shift"
)

// =====================================================
// NewShiftSlotTemplate Tests
// =====================================================

func TestNewShiftSlotTemplate_Success(t *testing.T) {
	tenantID := common.NewTenantID()
	eventID := common.NewEventID()

	template, err := shift.NewShiftSlotTemplate(
		time.Now(),
		tenantID,
		eventID,
		"Test Template",
		"Test description",
		[]*shift.ShiftSlotTemplateItem{},
	)

	if err != nil {
		t.Fatalf("NewShiftSlotTemplate() should succeed: %v", err)
	}

	if template.TenantID() != tenantID {
		t.Errorf("TenantID mismatch: got %v, want %v", template.TenantID(), tenantID)
	}

	if template.EventID() != eventID {
		t.Errorf("EventID mismatch: got %v, want %v", template.EventID(), eventID)
	}

	if template.TemplateName() != "Test Template" {
		t.Errorf("TemplateName mismatch: got %v, want 'Test Template'", template.TemplateName())
	}

	if template.Description() != "Test description" {
		t.Errorf("Description mismatch: got %v, want 'Test description'", template.Description())
	}

	if template.TemplateID().String() == "" {
		t.Error("TemplateID should be generated")
	}

	if template.CreatedAt().IsZero() {
		t.Error("CreatedAt should be set")
	}

	if template.UpdatedAt().IsZero() {
		t.Error("UpdatedAt should be set")
	}

	if template.DeletedAt() != nil {
		t.Error("DeletedAt should be nil for new template")
	}
}

func TestNewShiftSlotTemplate_Success_EmptyDescription(t *testing.T) {
	tenantID := common.NewTenantID()
	eventID := common.NewEventID()

	template, err := shift.NewShiftSlotTemplate(
		time.Now(),
		tenantID,
		eventID,
		"Test Template",
		"", // Empty description is allowed
		[]*shift.ShiftSlotTemplateItem{},
	)

	if err != nil {
		t.Fatalf("NewShiftSlotTemplate() should succeed with empty description: %v", err)
	}

	if template.Description() != "" {
		t.Error("Description should be empty")
	}
}

func TestNewShiftSlotTemplate_ErrorWhenInvalidTenantID(t *testing.T) {
	eventID := common.NewEventID()

	_, err := shift.NewShiftSlotTemplate(
		time.Now(),
		common.TenantID(""), // Invalid
		eventID,
		"Test Template",
		"",
		[]*shift.ShiftSlotTemplateItem{},
	)

	if err == nil {
		t.Error("NewShiftSlotTemplate() should fail when tenant_id is invalid")
	}
}

func TestNewShiftSlotTemplate_ErrorWhenInvalidEventID(t *testing.T) {
	tenantID := common.NewTenantID()

	_, err := shift.NewShiftSlotTemplate(
		time.Now(),
		tenantID,
		common.EventID(""), // Invalid
		"Test Template",
		"",
		[]*shift.ShiftSlotTemplateItem{},
	)

	if err == nil {
		t.Error("NewShiftSlotTemplate() should fail when event_id is invalid")
	}
}

func TestNewShiftSlotTemplate_ErrorWhenEmptyTemplateName(t *testing.T) {
	tenantID := common.NewTenantID()
	eventID := common.NewEventID()

	_, err := shift.NewShiftSlotTemplate(
		time.Now(),
		tenantID,
		eventID,
		"", // Empty template name
		"",
		[]*shift.ShiftSlotTemplateItem{},
	)

	if err == nil {
		t.Error("NewShiftSlotTemplate() should fail when template_name is empty")
	}
}

func TestNewShiftSlotTemplate_ErrorWhenTemplateNameTooLong(t *testing.T) {
	tenantID := common.NewTenantID()
	eventID := common.NewEventID()

	longName := strings.Repeat("a", 101) // 101 characters

	_, err := shift.NewShiftSlotTemplate(
		time.Now(),
		tenantID,
		eventID,
		longName,
		"",
		[]*shift.ShiftSlotTemplateItem{},
	)

	if err == nil {
		t.Error("NewShiftSlotTemplate() should fail when template_name is too long")
	}
}

// =====================================================
// NewShiftSlotTemplateItem Tests
// =====================================================

func TestNewShiftSlotTemplateItem_Success(t *testing.T) {
	templateID := common.NewShiftSlotTemplateID()
	startTime := time.Date(2000, 1, 1, 20, 0, 0, 0, time.UTC)
	endTime := time.Date(2000, 1, 1, 22, 0, 0, 0, time.UTC)

	item, err := shift.NewShiftSlotTemplateItem(
		time.Now(),
		templateID,
		"DJ Slot",
		"Main Stage",
		startTime,
		endTime,
		2,
		1,
	)

	if err != nil {
		t.Fatalf("NewShiftSlotTemplateItem() should succeed: %v", err)
	}

	if item.TemplateID() != templateID {
		t.Errorf("TemplateID mismatch: got %v, want %v", item.TemplateID(), templateID)
	}

	if item.SlotName() != "DJ Slot" {
		t.Errorf("SlotName mismatch: got %v, want 'DJ Slot'", item.SlotName())
	}

	if item.InstanceName() != "Main Stage" {
		t.Errorf("InstanceName mismatch: got %v, want 'Main Stage'", item.InstanceName())
	}

	if item.RequiredCount() != 2 {
		t.Errorf("RequiredCount mismatch: got %v, want 2", item.RequiredCount())
	}

	if item.Priority() != 1 {
		t.Errorf("Priority mismatch: got %v, want 1", item.Priority())
	}

	if item.ItemID().String() == "" {
		t.Error("ItemID should be generated")
	}
}

func TestNewShiftSlotTemplateItem_Success_EmptyInstanceName(t *testing.T) {
	templateID := common.NewShiftSlotTemplateID()
	startTime := time.Date(2000, 1, 1, 20, 0, 0, 0, time.UTC)
	endTime := time.Date(2000, 1, 1, 22, 0, 0, 0, time.UTC)

	item, err := shift.NewShiftSlotTemplateItem(
		time.Now(),
		templateID,
		"DJ Slot",
		"", // Empty instance name is allowed
		startTime,
		endTime,
		1,
		1,
	)

	if err != nil {
		t.Fatalf("NewShiftSlotTemplateItem() should succeed with empty instance_name: %v", err)
	}

	if item.InstanceName() != "" {
		t.Error("InstanceName should be empty")
	}
}

func TestNewShiftSlotTemplateItem_ErrorWhenInvalidTemplateID(t *testing.T) {
	startTime := time.Date(2000, 1, 1, 20, 0, 0, 0, time.UTC)
	endTime := time.Date(2000, 1, 1, 22, 0, 0, 0, time.UTC)

	_, err := shift.NewShiftSlotTemplateItem(
		time.Now(),
		common.ShiftSlotTemplateID(""), // Invalid
		"DJ Slot",
		"",
		startTime,
		endTime,
		1,
		1,
	)

	if err == nil {
		t.Error("NewShiftSlotTemplateItem() should fail when template_id is invalid")
	}
}

func TestNewShiftSlotTemplateItem_ErrorWhenEmptySlotName(t *testing.T) {
	templateID := common.NewShiftSlotTemplateID()
	startTime := time.Date(2000, 1, 1, 20, 0, 0, 0, time.UTC)
	endTime := time.Date(2000, 1, 1, 22, 0, 0, 0, time.UTC)

	_, err := shift.NewShiftSlotTemplateItem(
		time.Now(),
		templateID,
		"", // Empty slot name
		"",
		startTime,
		endTime,
		1,
		1,
	)

	if err == nil {
		t.Error("NewShiftSlotTemplateItem() should fail when slot_name is empty")
	}
}

func TestNewShiftSlotTemplateItem_ErrorWhenRequiredCountZero(t *testing.T) {
	templateID := common.NewShiftSlotTemplateID()
	startTime := time.Date(2000, 1, 1, 20, 0, 0, 0, time.UTC)
	endTime := time.Date(2000, 1, 1, 22, 0, 0, 0, time.UTC)

	_, err := shift.NewShiftSlotTemplateItem(
		time.Now(),
		templateID,
		"DJ Slot",
		"",
		startTime,
		endTime,
		0, // Zero is not allowed
		1,
	)

	if err == nil {
		t.Error("NewShiftSlotTemplateItem() should fail when required_count is zero")
	}
}

func TestNewShiftSlotTemplateItem_ErrorWhenRequiredCountNegative(t *testing.T) {
	templateID := common.NewShiftSlotTemplateID()
	startTime := time.Date(2000, 1, 1, 20, 0, 0, 0, time.UTC)
	endTime := time.Date(2000, 1, 1, 22, 0, 0, 0, time.UTC)

	_, err := shift.NewShiftSlotTemplateItem(
		time.Now(),
		templateID,
		"DJ Slot",
		"",
		startTime,
		endTime,
		-1, // Negative is not allowed
		1,
	)

	if err == nil {
		t.Error("NewShiftSlotTemplateItem() should fail when required_count is negative")
	}
}

func TestNewShiftSlotTemplateItem_ErrorWhenPriorityZero(t *testing.T) {
	templateID := common.NewShiftSlotTemplateID()
	startTime := time.Date(2000, 1, 1, 20, 0, 0, 0, time.UTC)
	endTime := time.Date(2000, 1, 1, 22, 0, 0, 0, time.UTC)

	_, err := shift.NewShiftSlotTemplateItem(
		time.Now(),
		templateID,
		"DJ Slot",
		"",
		startTime,
		endTime,
		1,
		0, // Zero is not allowed (must be at least 1)
	)

	if err == nil {
		t.Error("NewShiftSlotTemplateItem() should fail when priority is zero")
	}
}

func TestNewShiftSlotTemplateItem_ErrorWhenPriorityNegative(t *testing.T) {
	templateID := common.NewShiftSlotTemplateID()
	startTime := time.Date(2000, 1, 1, 20, 0, 0, 0, time.UTC)
	endTime := time.Date(2000, 1, 1, 22, 0, 0, 0, time.UTC)

	_, err := shift.NewShiftSlotTemplateItem(
		time.Now(),
		templateID,
		"DJ Slot",
		"",
		startTime,
		endTime,
		1,
		-1, // Negative is not allowed
	)

	if err == nil {
		t.Error("NewShiftSlotTemplateItem() should fail when priority is negative")
	}
}

// =====================================================
// ReconstructShiftSlotTemplate Tests
// =====================================================

func TestReconstructShiftSlotTemplate_Success(t *testing.T) {
	templateID := common.NewShiftSlotTemplateID()
	tenantID := common.NewTenantID()
	eventID := common.NewEventID()
	createdAt := time.Now().Add(-time.Hour)
	updatedAt := time.Now()

	template, err := shift.ReconstructShiftSlotTemplate(
		templateID,
		tenantID,
		eventID,
		"Test Template",
		"Description",
		[]*shift.ShiftSlotTemplateItem{},
		createdAt,
		updatedAt,
		nil,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if template.TemplateID() != templateID {
		t.Errorf("TemplateID mismatch: got %v, want %v", template.TemplateID(), templateID)
	}

	if template.TemplateName() != "Test Template" {
		t.Errorf("TemplateName mismatch: got %v, want 'Test Template'", template.TemplateName())
	}

	if template.DeletedAt() != nil {
		t.Error("DeletedAt should be nil")
	}
}

func TestReconstructShiftSlotTemplate_Success_Deleted(t *testing.T) {
	templateID := common.NewShiftSlotTemplateID()
	tenantID := common.NewTenantID()
	eventID := common.NewEventID()
	createdAt := time.Now().Add(-time.Hour)
	updatedAt := time.Now()
	deletedAt := time.Now()

	template, err := shift.ReconstructShiftSlotTemplate(
		templateID,
		tenantID,
		eventID,
		"Test Template",
		"",
		[]*shift.ShiftSlotTemplateItem{},
		createdAt,
		updatedAt,
		&deletedAt,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if template.DeletedAt() == nil {
		t.Error("DeletedAt should be set")
	}
}

// =====================================================
// ReconstructShiftSlotTemplateItem Tests
// =====================================================

func TestReconstructShiftSlotTemplateItem_Success(t *testing.T) {
	itemID := common.NewShiftSlotTemplateItemID()
	templateID := common.NewShiftSlotTemplateID()
	startTime := time.Date(2000, 1, 1, 20, 0, 0, 0, time.UTC)
	endTime := time.Date(2000, 1, 1, 22, 0, 0, 0, time.UTC)
	createdAt := time.Now()
	updatedAt := time.Now()

	item, err := shift.ReconstructShiftSlotTemplateItem(
		itemID,
		templateID,
		"DJ Slot",
		"Main Stage",
		startTime,
		endTime,
		3,
		2,
		createdAt,
		updatedAt,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if item.ItemID() != itemID {
		t.Errorf("ItemID mismatch: got %v, want %v", item.ItemID(), itemID)
	}

	if item.TemplateID() != templateID {
		t.Errorf("TemplateID mismatch: got %v, want %v", item.TemplateID(), templateID)
	}

	if item.SlotName() != "DJ Slot" {
		t.Errorf("SlotName mismatch: got %v, want 'DJ Slot'", item.SlotName())
	}

	if item.RequiredCount() != 3 {
		t.Errorf("RequiredCount mismatch: got %v, want 3", item.RequiredCount())
	}

	if item.Priority() != 2 {
		t.Errorf("Priority mismatch: got %v, want 2", item.Priority())
	}
}

// =====================================================
// ShiftSlotTemplate Methods Tests
// =====================================================

func TestShiftSlotTemplate_UpdateDetails_Success(t *testing.T) {
	tenantID := common.NewTenantID()
	eventID := common.NewEventID()

	template, _ := shift.NewShiftSlotTemplate(
		time.Now(),
		tenantID,
		eventID,
		"Original Name",
		"Original Description",
		[]*shift.ShiftSlotTemplateItem{},
	)

	templateID := template.TemplateID()
	startTime := time.Date(2000, 1, 1, 20, 0, 0, 0, time.UTC)
	endTime := time.Date(2000, 1, 1, 22, 0, 0, 0, time.UTC)

	item, _ := shift.NewShiftSlotTemplateItem(
		time.Now(),
		templateID,
		"DJ Slot",
		"",
		startTime,
		endTime,
		1,
		1,
	)

	originalUpdatedAt := template.UpdatedAt()
	time.Sleep(time.Millisecond)

	err := template.UpdateDetails(time.Now(),
		"Updated Name",
		"Updated Description",
		[]*shift.ShiftSlotTemplateItem{item},
	)

	if err != nil {
		t.Fatalf("UpdateDetails() should succeed: %v", err)
	}

	if template.TemplateName() != "Updated Name" {
		t.Errorf("TemplateName should be updated: got %v, want 'Updated Name'", template.TemplateName())
	}

	if template.Description() != "Updated Description" {
		t.Errorf("Description should be updated: got %v, want 'Updated Description'", template.Description())
	}

	if len(template.Items()) != 1 {
		t.Errorf("Items length mismatch: got %v, want 1", len(template.Items()))
	}

	if !template.UpdatedAt().After(originalUpdatedAt) {
		t.Error("UpdatedAt should be updated")
	}
}

func TestShiftSlotTemplate_UpdateDetails_ErrorWhenEmptyName(t *testing.T) {
	tenantID := common.NewTenantID()
	eventID := common.NewEventID()

	template, _ := shift.NewShiftSlotTemplate(
		time.Now(),
		tenantID,
		eventID,
		"Original Name",
		"",
		[]*shift.ShiftSlotTemplateItem{},
	)

	templateID := template.TemplateID()
	startTime := time.Date(2000, 1, 1, 20, 0, 0, 0, time.UTC)
	endTime := time.Date(2000, 1, 1, 22, 0, 0, 0, time.UTC)

	item, _ := shift.NewShiftSlotTemplateItem(
		time.Now(),
		templateID,
		"DJ Slot",
		"",
		startTime,
		endTime,
		1,
		1,
	)

	err := template.UpdateDetails(time.Now(),"", "", []*shift.ShiftSlotTemplateItem{item})

	if err == nil {
		t.Error("UpdateDetails() should fail when template_name is empty")
	}
}

func TestShiftSlotTemplate_UpdateDetails_ErrorWhenNameTooLong(t *testing.T) {
	tenantID := common.NewTenantID()
	eventID := common.NewEventID()

	template, _ := shift.NewShiftSlotTemplate(
		time.Now(),
		tenantID,
		eventID,
		"Original Name",
		"",
		[]*shift.ShiftSlotTemplateItem{},
	)

	templateID := template.TemplateID()
	startTime := time.Date(2000, 1, 1, 20, 0, 0, 0, time.UTC)
	endTime := time.Date(2000, 1, 1, 22, 0, 0, 0, time.UTC)

	item, _ := shift.NewShiftSlotTemplateItem(
		time.Now(),
		templateID,
		"DJ Slot",
		"",
		startTime,
		endTime,
		1,
		1,
	)

	longName := strings.Repeat("a", 101)
	err := template.UpdateDetails(time.Now(),longName, "", []*shift.ShiftSlotTemplateItem{item})

	if err == nil {
		t.Error("UpdateDetails() should fail when template_name is too long")
	}
}

func TestShiftSlotTemplate_UpdateDetails_ErrorWhenNoItems(t *testing.T) {
	tenantID := common.NewTenantID()
	eventID := common.NewEventID()

	template, _ := shift.NewShiftSlotTemplate(
		time.Now(),
		tenantID,
		eventID,
		"Original Name",
		"",
		[]*shift.ShiftSlotTemplateItem{},
	)

	err := template.UpdateDetails(time.Now(),"Updated Name", "", []*shift.ShiftSlotTemplateItem{})

	if err == nil {
		t.Error("UpdateDetails() should fail when items is empty")
	}
}

func TestShiftSlotTemplate_Delete(t *testing.T) {
	tenantID := common.NewTenantID()
	eventID := common.NewEventID()

	template, _ := shift.NewShiftSlotTemplate(
		time.Now(),
		tenantID,
		eventID,
		"Test Template",
		"",
		[]*shift.ShiftSlotTemplateItem{},
	)

	if template.DeletedAt() != nil {
		t.Error("DeletedAt should be nil before Delete()")
	}

	template.Delete(time.Now())

	if template.DeletedAt() == nil {
		t.Error("DeletedAt should be set after Delete()")
	}
}
