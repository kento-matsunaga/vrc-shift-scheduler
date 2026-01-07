package shift

import (
	"context"
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/event"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/shift"
)

// TemplateItemInput represents input for a template item
type TemplateItemInput struct {
	SlotName      string
	InstanceName  string
	StartTime     time.Time
	EndTime       time.Time
	RequiredCount int
	Priority      int
}

// CreateShiftTemplateInput represents the input for creating a shift template
type CreateShiftTemplateInput struct {
	TenantID     common.TenantID
	EventID      common.EventID
	TemplateName string
	Description  string
	Items        []TemplateItemInput
}

// CreateShiftTemplateUsecase handles shift template creation
type CreateShiftTemplateUsecase struct {
	templateRepo shift.ShiftSlotTemplateRepository
}

// NewCreateShiftTemplateUsecase creates a new CreateShiftTemplateUsecase
func NewCreateShiftTemplateUsecase(templateRepo shift.ShiftSlotTemplateRepository) *CreateShiftTemplateUsecase {
	return &CreateShiftTemplateUsecase{
		templateRepo: templateRepo,
	}
}

// Execute creates a new shift template
func (uc *CreateShiftTemplateUsecase) Execute(ctx context.Context, input CreateShiftTemplateInput) (*shift.ShiftSlotTemplate, error) {
	// Create template first to get the template ID
	template, err := shift.NewShiftSlotTemplate(
		input.TenantID,
		input.EventID,
		input.TemplateName,
		input.Description,
		[]*shift.ShiftSlotTemplateItem{}, // empty items initially
	)
	if err != nil {
		return nil, err
	}

	// Create template items using the template's ID
	var items []*shift.ShiftSlotTemplateItem
	for _, itemInput := range input.Items {
		item, err := shift.NewShiftSlotTemplateItem(
			template.TemplateID(),
			itemInput.SlotName,
			itemInput.InstanceName,
			itemInput.StartTime,
			itemInput.EndTime,
			itemInput.RequiredCount,
			itemInput.Priority,
		)
		if err != nil {
			return nil, err
		}

		items = append(items, item)
	}

	// Update template with items
	if err := template.UpdateDetails(input.TemplateName, input.Description, items); err != nil {
		return nil, err
	}

	// Save
	if err := uc.templateRepo.Save(ctx, template); err != nil {
		return nil, err
	}

	return template, nil
}

// ListShiftTemplatesInput represents the input for listing shift templates
type ListShiftTemplatesInput struct {
	TenantID common.TenantID
	EventID  common.EventID
}

// ListShiftTemplatesUsecase handles shift template listing
type ListShiftTemplatesUsecase struct {
	templateRepo shift.ShiftSlotTemplateRepository
}

// NewListShiftTemplatesUsecase creates a new ListShiftTemplatesUsecase
func NewListShiftTemplatesUsecase(templateRepo shift.ShiftSlotTemplateRepository) *ListShiftTemplatesUsecase {
	return &ListShiftTemplatesUsecase{
		templateRepo: templateRepo,
	}
}

// Execute retrieves all templates for an event
func (uc *ListShiftTemplatesUsecase) Execute(ctx context.Context, input ListShiftTemplatesInput) ([]*shift.ShiftSlotTemplate, error) {
	templates, err := uc.templateRepo.FindByEventID(ctx, input.TenantID, input.EventID)
	if err != nil {
		return nil, err
	}

	return templates, nil
}

// GetShiftTemplateInput represents the input for getting a shift template
type GetShiftTemplateInput struct {
	TenantID   common.TenantID
	TemplateID common.ShiftSlotTemplateID
}

// GetShiftTemplateUsecase handles shift template retrieval
type GetShiftTemplateUsecase struct {
	templateRepo shift.ShiftSlotTemplateRepository
}

// NewGetShiftTemplateUsecase creates a new GetShiftTemplateUsecase
func NewGetShiftTemplateUsecase(templateRepo shift.ShiftSlotTemplateRepository) *GetShiftTemplateUsecase {
	return &GetShiftTemplateUsecase{
		templateRepo: templateRepo,
	}
}

// Execute retrieves a template by ID
func (uc *GetShiftTemplateUsecase) Execute(ctx context.Context, input GetShiftTemplateInput) (*shift.ShiftSlotTemplate, error) {
	template, err := uc.templateRepo.FindByID(ctx, input.TenantID, input.TemplateID)
	if err != nil {
		return nil, err
	}

	return template, nil
}

// UpdateShiftTemplateInput represents the input for updating a shift template
type UpdateShiftTemplateInput struct {
	TenantID     common.TenantID
	TemplateID   common.ShiftSlotTemplateID
	TemplateName string
	Description  string
	Items        []TemplateItemInput
}

// UpdateShiftTemplateUsecase handles shift template update
type UpdateShiftTemplateUsecase struct {
	templateRepo shift.ShiftSlotTemplateRepository
}

// NewUpdateShiftTemplateUsecase creates a new UpdateShiftTemplateUsecase
func NewUpdateShiftTemplateUsecase(templateRepo shift.ShiftSlotTemplateRepository) *UpdateShiftTemplateUsecase {
	return &UpdateShiftTemplateUsecase{
		templateRepo: templateRepo,
	}
}

// Execute updates an existing shift template
func (uc *UpdateShiftTemplateUsecase) Execute(ctx context.Context, input UpdateShiftTemplateInput) (*shift.ShiftSlotTemplate, error) {
	// Fetch existing template
	template, err := uc.templateRepo.FindByID(ctx, input.TenantID, input.TemplateID)
	if err != nil {
		return nil, err
	}

	// Create new template items
	var items []*shift.ShiftSlotTemplateItem
	for _, itemInput := range input.Items {
		item, err := shift.NewShiftSlotTemplateItem(
			template.TemplateID(),
			itemInput.SlotName,
			itemInput.InstanceName,
			itemInput.StartTime,
			itemInput.EndTime,
			itemInput.RequiredCount,
			itemInput.Priority,
		)
		if err != nil {
			return nil, err
		}

		items = append(items, item)
	}

	// Update template details
	if err := template.UpdateDetails(input.TemplateName, input.Description, items); err != nil {
		return nil, err
	}

	// Save
	if err := uc.templateRepo.Save(ctx, template); err != nil {
		return nil, err
	}

	return template, nil
}

// DeleteShiftTemplateInput represents the input for deleting a shift template
type DeleteShiftTemplateInput struct {
	TenantID   common.TenantID
	TemplateID common.ShiftSlotTemplateID
}

// DeleteShiftTemplateUsecase handles shift template deletion
type DeleteShiftTemplateUsecase struct {
	templateRepo shift.ShiftSlotTemplateRepository
}

// NewDeleteShiftTemplateUsecase creates a new DeleteShiftTemplateUsecase
func NewDeleteShiftTemplateUsecase(templateRepo shift.ShiftSlotTemplateRepository) *DeleteShiftTemplateUsecase {
	return &DeleteShiftTemplateUsecase{
		templateRepo: templateRepo,
	}
}

// Execute deletes a shift template
func (uc *DeleteShiftTemplateUsecase) Execute(ctx context.Context, input DeleteShiftTemplateInput) error {
	// Delete the template
	if err := uc.templateRepo.Delete(ctx, input.TenantID, input.TemplateID); err != nil {
		return err
	}

	return nil
}

// SaveBusinessDayAsTemplateInput represents the input for saving a business day as template
type SaveBusinessDayAsTemplateInput struct {
	TenantID      common.TenantID
	BusinessDayID event.BusinessDayID
	TemplateName  string
	Description   string
}

// SaveBusinessDayAsTemplateUsecase handles saving a business day as a template
type SaveBusinessDayAsTemplateUsecase struct {
	templateRepo    shift.ShiftSlotTemplateRepository
	businessDayRepo event.EventBusinessDayRepository
	slotRepo        shift.ShiftSlotRepository
}

// NewSaveBusinessDayAsTemplateUsecase creates a new SaveBusinessDayAsTemplateUsecase
func NewSaveBusinessDayAsTemplateUsecase(
	templateRepo shift.ShiftSlotTemplateRepository,
	businessDayRepo event.EventBusinessDayRepository,
	slotRepo shift.ShiftSlotRepository,
) *SaveBusinessDayAsTemplateUsecase {
	return &SaveBusinessDayAsTemplateUsecase{
		templateRepo:    templateRepo,
		businessDayRepo: businessDayRepo,
		slotRepo:        slotRepo,
	}
}

// Execute saves a business day's shift slots as a template
func (uc *SaveBusinessDayAsTemplateUsecase) Execute(ctx context.Context, input SaveBusinessDayAsTemplateInput) (*shift.ShiftSlotTemplate, error) {
	// Find business day
	businessDay, err := uc.businessDayRepo.FindByID(ctx, input.TenantID, input.BusinessDayID)
	if err != nil {
		return nil, err
	}

	// Find shift slots for this business day
	slots, err := uc.slotRepo.FindByBusinessDayID(ctx, input.TenantID, input.BusinessDayID)
	if err != nil {
		return nil, err
	}

	if len(slots) == 0 {
		return nil, common.NewValidationError("Business day has no shift slots to save as template", nil)
	}

	// Create template items from shift slots
	var items []*shift.ShiftSlotTemplateItem
	templateID := common.NewShiftSlotTemplateID()

	for _, slot := range slots {
		item, err := shift.NewShiftSlotTemplateItem(
			templateID,
			slot.SlotName(),
			slot.InstanceName(),
			slot.StartTime(),
			slot.EndTime(),
			slot.RequiredCount(),
			slot.Priority(),
		)
		if err != nil {
			return nil, err
		}

		items = append(items, item)
	}

	// Create template
	template, err := shift.NewShiftSlotTemplate(
		input.TenantID,
		businessDay.EventID(),
		input.TemplateName,
		input.Description,
		items,
	)
	if err != nil {
		return nil, err
	}

	// Save
	if err := uc.templateRepo.Save(ctx, template); err != nil {
		return nil, err
	}

	return template, nil
}
