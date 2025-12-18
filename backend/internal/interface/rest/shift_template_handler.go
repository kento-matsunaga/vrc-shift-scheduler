package rest

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/event"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/shift"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/infra/db"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ShiftTemplateHandler handles shift template-related HTTP requests
type ShiftTemplateHandler struct {
	templateRepo    *db.ShiftSlotTemplateRepository
	slotRepo        *db.ShiftSlotRepository
	businessDayRepo *db.EventBusinessDayRepository
}

// NewShiftTemplateHandler creates a new ShiftTemplateHandler
func NewShiftTemplateHandler(dbPool *pgxpool.Pool) *ShiftTemplateHandler {
	return &ShiftTemplateHandler{
		templateRepo:    db.NewShiftSlotTemplateRepository(dbPool),
		slotRepo:        db.NewShiftSlotRepository(dbPool),
		businessDayRepo: db.NewEventBusinessDayRepository(dbPool),
	}
}

// TemplateItemRequest represents a single template item in request
type TemplateItemRequest struct {
	PositionID    string `json:"position_id"`
	SlotName      string `json:"slot_name"`
	InstanceName  string `json:"instance_name"`
	StartTime     string `json:"start_time"` // HH:MM:SS
	EndTime       string `json:"end_time"`   // HH:MM:SS
	RequiredCount int    `json:"required_count"`
	Priority      int    `json:"priority"`
}

// CreateTemplateRequest represents the request body for creating a template
type CreateTemplateRequest struct {
	TemplateName string                `json:"template_name"`
	Description  string                `json:"description"`
	Items        []TemplateItemRequest `json:"items"`
}

// UpdateTemplateRequest represents the request body for updating a template
type UpdateTemplateRequest struct {
	TemplateName string                `json:"template_name"`
	Description  string                `json:"description"`
	Items        []TemplateItemRequest `json:"items"`
}

// SaveAsTemplateRequest represents the request body for saving a business day as template
type SaveAsTemplateRequest struct {
	TemplateName string `json:"template_name"`
	Description  string `json:"description"`
}

// TemplateItemResponse represents a template item in API responses
type TemplateItemResponse struct {
	ItemID        string `json:"item_id"`
	PositionID    string `json:"position_id"`
	SlotName      string `json:"slot_name"`
	InstanceName  string `json:"instance_name"`
	StartTime     string `json:"start_time"`
	EndTime       string `json:"end_time"`
	RequiredCount int    `json:"required_count"`
	Priority      int    `json:"priority"`
}

// TemplateResponse represents a template in API responses
type TemplateResponse struct {
	TemplateID   string                 `json:"template_id"`
	TenantID     string                 `json:"tenant_id"`
	EventID      string                 `json:"event_id"`
	TemplateName string                 `json:"template_name"`
	Description  string                 `json:"description"`
	Items        []TemplateItemResponse `json:"items"`
	CreatedAt    string                 `json:"created_at"`
	UpdatedAt    string                 `json:"updated_at"`
}

// CreateTemplate handles POST /api/v1/events/:event_id/templates
func (h *ShiftTemplateHandler) CreateTemplate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get tenant ID
	tenantID, ok := getTenantIDFromContext(ctx)
	if !ok {
		writeError(w, http.StatusForbidden, "ERR_FORBIDDEN", "Tenant ID is required", nil)
		return
	}

	// Get event_id from URL
	eventIDStr := chi.URLParam(r, "event_id")
	if eventIDStr == "" {
		writeError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", "event_id is required", nil)
		return
	}

	eventID, err := common.ParseEventID(eventIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", "Invalid event_id format", nil)
		return
	}

	// Parse request body
	var req CreateTemplateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", "Invalid request body", nil)
		return
	}

	// Validate
	if req.TemplateName == "" {
		writeError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", "template_name is required", nil)
		return
	}

	if len(req.Items) == 0 {
		writeError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", "At least one template item is required", nil)
		return
	}

	// Create template first to get the template ID
	template, err := shift.NewShiftSlotTemplate(
		tenantID,
		eventID,
		req.TemplateName,
		req.Description,
		[]*shift.ShiftSlotTemplateItem{}, // empty items initially
	)
	if err != nil {
		writeError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", err.Error(), nil)
		return
	}

	// Create template items using the template's ID
	var items []*shift.ShiftSlotTemplateItem
	for _, itemReq := range req.Items {
		positionID, err := shift.ParsePositionID(itemReq.PositionID)
		if err != nil {
			writeError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", "Invalid position_id format", nil)
			return
		}

		startTime, err := time.Parse("15:04:05", itemReq.StartTime)
		if err != nil {
			writeError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", "Invalid start_time format (expected HH:MM:SS)", nil)
			return
		}

		endTime, err := time.Parse("15:04:05", itemReq.EndTime)
		if err != nil {
			writeError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", "Invalid end_time format (expected HH:MM:SS)", nil)
			return
		}

		item, err := shift.NewShiftSlotTemplateItem(
			template.TemplateID(),
			positionID,
			itemReq.SlotName,
			itemReq.InstanceName,
			startTime,
			endTime,
			itemReq.RequiredCount,
			itemReq.Priority,
		)
		if err != nil {
			writeError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", err.Error(), nil)
			return
		}

		items = append(items, item)
	}

	// Update template with items
	if err := template.UpdateDetails(req.TemplateName, req.Description, items); err != nil {
		writeError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", err.Error(), nil)
		return
	}

	// Save
	if err := h.templateRepo.Save(ctx, template); err != nil {
		// Log the actual error for debugging
		writeError(w, http.StatusInternalServerError, "ERR_INTERNAL", err.Error(), nil)
		return
	}

	// Return response
	response := h.toTemplateResponse(template)
	RespondCreated(w, response)
}

// ListTemplates handles GET /api/v1/events/:event_id/templates
func (h *ShiftTemplateHandler) ListTemplates(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get tenant ID
	tenantID, ok := getTenantIDFromContext(ctx)
	if !ok {
		writeError(w, http.StatusForbidden, "ERR_FORBIDDEN", "Tenant ID is required", nil)
		return
	}

	// Get event_id from URL
	eventIDStr := chi.URLParam(r, "event_id")
	if eventIDStr == "" {
		writeError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", "event_id is required", nil)
		return
	}

	eventID, err := common.ParseEventID(eventIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", "Invalid event_id format", nil)
		return
	}

	// Find templates
	templates, err := h.templateRepo.FindByEventID(ctx, tenantID, eventID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "ERR_INTERNAL", "Failed to fetch templates", nil)
		return
	}

	// Convert to responses
	var responses []TemplateResponse
	for _, template := range templates {
		responses = append(responses, h.toTemplateResponse(template))
	}

	RespondSuccess(w, map[string]interface{}{
		"count":     len(responses),
		"templates": responses,
	})
}

// GetTemplate handles GET /api/v1/events/:event_id/templates/:template_id
func (h *ShiftTemplateHandler) GetTemplate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get tenant ID
	tenantID, ok := getTenantIDFromContext(ctx)
	if !ok {
		writeError(w, http.StatusForbidden, "ERR_FORBIDDEN", "Tenant ID is required", nil)
		return
	}

	// Get template_id from URL
	templateIDStr := chi.URLParam(r, "template_id")
	if templateIDStr == "" {
		writeError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", "template_id is required", nil)
		return
	}

	templateID, err := common.ParseShiftSlotTemplateID(templateIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", "Invalid template_id format", nil)
		return
	}

	// Find template
	template, err := h.templateRepo.FindByID(ctx, tenantID, templateID)
	if err != nil {
		if err.Error() == "ShiftSlotTemplate not found" {
			writeError(w, http.StatusNotFound, "ERR_NOT_FOUND", "Template not found", nil)
			return
		}
		writeError(w, http.StatusInternalServerError, "ERR_INTERNAL", "Failed to fetch template", nil)
		return
	}

	response := h.toTemplateResponse(template)
	RespondSuccess(w, response)
}

// UpdateTemplate handles PUT /api/v1/events/:event_id/templates/:template_id
func (h *ShiftTemplateHandler) UpdateTemplate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get tenant ID
	tenantID, ok := getTenantIDFromContext(ctx)
	if !ok {
		writeError(w, http.StatusForbidden, "ERR_FORBIDDEN", "Tenant ID is required", nil)
		return
	}

	// Get template_id from URL
	templateIDStr := chi.URLParam(r, "template_id")
	if templateIDStr == "" {
		writeError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", "template_id is required", nil)
		return
	}

	templateID, err := common.ParseShiftSlotTemplateID(templateIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", "Invalid template_id format", nil)
		return
	}

	// Parse request body
	var req UpdateTemplateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", "Invalid request body", nil)
		return
	}

	// Validate
	if req.TemplateName == "" {
		writeError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", "template_name is required", nil)
		return
	}

	if len(req.Items) == 0 {
		writeError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", "At least one template item is required", nil)
		return
	}

	// Find existing template
	template, err := h.templateRepo.FindByID(ctx, tenantID, templateID)
	if err != nil {
		if err.Error() == "ShiftSlotTemplate not found" {
			writeError(w, http.StatusNotFound, "ERR_NOT_FOUND", "Template not found", nil)
			return
		}
		writeError(w, http.StatusInternalServerError, "ERR_INTERNAL", "Failed to fetch template", nil)
		return
	}

	// Create new template items
	var items []*shift.ShiftSlotTemplateItem
	for _, itemReq := range req.Items {
		positionID, err := shift.ParsePositionID(itemReq.PositionID)
		if err != nil {
			writeError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", "Invalid position_id format", nil)
			return
		}

		startTime, err := time.Parse("15:04:05", itemReq.StartTime)
		if err != nil {
			writeError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", "Invalid start_time format (expected HH:MM:SS)", nil)
			return
		}

		endTime, err := time.Parse("15:04:05", itemReq.EndTime)
		if err != nil {
			writeError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", "Invalid end_time format (expected HH:MM:SS)", nil)
			return
		}

		item, err := shift.NewShiftSlotTemplateItem(
			templateID,
			positionID,
			itemReq.SlotName,
			itemReq.InstanceName,
			startTime,
			endTime,
			itemReq.RequiredCount,
			itemReq.Priority,
		)
		if err != nil {
			writeError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", err.Error(), nil)
			return
		}

		items = append(items, item)
	}

	// Update template
	if err := template.UpdateDetails(req.TemplateName, req.Description, items); err != nil {
		writeError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", err.Error(), nil)
		return
	}

	// Save
	if err := h.templateRepo.Save(ctx, template); err != nil {
		writeError(w, http.StatusInternalServerError, "ERR_INTERNAL", "Failed to update template", nil)
		return
	}

	response := h.toTemplateResponse(template)
	RespondSuccess(w, response)
}

// DeleteTemplate handles DELETE /api/v1/events/:event_id/templates/:template_id
func (h *ShiftTemplateHandler) DeleteTemplate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get tenant ID
	tenantID, ok := getTenantIDFromContext(ctx)
	if !ok {
		writeError(w, http.StatusForbidden, "ERR_FORBIDDEN", "Tenant ID is required", nil)
		return
	}

	// Get template_id from URL
	templateIDStr := chi.URLParam(r, "template_id")
	if templateIDStr == "" {
		writeError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", "template_id is required", nil)
		return
	}

	templateID, err := common.ParseShiftSlotTemplateID(templateIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", "Invalid template_id format", nil)
		return
	}

	// Delete template
	if err := h.templateRepo.Delete(ctx, tenantID, templateID); err != nil {
		if err.Error() == "ShiftSlotTemplate not found" {
			writeError(w, http.StatusNotFound, "ERR_NOT_FOUND", "Template not found", nil)
			return
		}
		writeError(w, http.StatusInternalServerError, "ERR_INTERNAL", "Failed to delete template", nil)
		return
	}

	RespondSuccess(w, map[string]string{
		"message": "Template deleted successfully",
	})
}

// SaveBusinessDayAsTemplate handles POST /api/v1/business-days/:business_day_id/save-as-template
func (h *ShiftTemplateHandler) SaveBusinessDayAsTemplate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get tenant ID
	tenantID, ok := getTenantIDFromContext(ctx)
	if !ok {
		writeError(w, http.StatusForbidden, "ERR_FORBIDDEN", "Tenant ID is required", nil)
		return
	}

	// Get business_day_id from URL
	businessDayIDStr := chi.URLParam(r, "business_day_id")
	if businessDayIDStr == "" {
		writeError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", "business_day_id is required", nil)
		return
	}

	businessDayID, err := event.ParseBusinessDayID(businessDayIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", "Invalid business_day_id format", nil)
		return
	}

	// Parse request body
	var req SaveAsTemplateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", "Invalid request body", nil)
		return
	}

	// Validate
	if req.TemplateName == "" {
		writeError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", "template_name is required", nil)
		return
	}

	// Find business day
	businessDay, err := h.businessDayRepo.FindByID(ctx, tenantID, businessDayID)
	if err != nil {
		if err.Error() == "business day not found" {
			writeError(w, http.StatusNotFound, "ERR_NOT_FOUND", "Business day not found", nil)
			return
		}
		writeError(w, http.StatusInternalServerError, "ERR_INTERNAL", "Failed to fetch business day", nil)
		return
	}

	// Find shift slots for this business day
	slots, err := h.slotRepo.FindByBusinessDayID(ctx, tenantID, businessDayID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "ERR_INTERNAL", "Failed to fetch shift slots", nil)
		return
	}

	if len(slots) == 0 {
		writeError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", "Business day has no shift slots to save as template", nil)
		return
	}

	// Create template items from shift slots
	var items []*shift.ShiftSlotTemplateItem
	templateID := common.NewShiftSlotTemplateID()

	for _, slot := range slots {
		item, err := shift.NewShiftSlotTemplateItem(
			templateID,
			slot.PositionID(),
			slot.SlotName(),
			slot.InstanceName(),
			slot.StartTime(),
			slot.EndTime(),
			slot.RequiredCount(),
			slot.Priority(),
		)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "ERR_INTERNAL", "Failed to create template item", nil)
			return
		}

		items = append(items, item)
	}

	// Create template
	template, err := shift.NewShiftSlotTemplate(
		tenantID,
		businessDay.EventID(),
		req.TemplateName,
		req.Description,
		items,
	)
	if err != nil {
		writeError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", err.Error(), nil)
		return
	}

	// Save
	if err := h.templateRepo.Save(ctx, template); err != nil {
		writeError(w, http.StatusInternalServerError, "ERR_INTERNAL", "Failed to save template", nil)
		return
	}

	response := h.toTemplateResponse(template)
	RespondCreated(w, response)
}

// toTemplateResponse converts a template entity to API response
func (h *ShiftTemplateHandler) toTemplateResponse(template *shift.ShiftSlotTemplate) TemplateResponse {
	var items []TemplateItemResponse
	for _, item := range template.Items() {
		items = append(items, TemplateItemResponse{
			ItemID:        item.ItemID().String(),
			PositionID:    item.PositionID().String(),
			SlotName:      item.SlotName(),
			InstanceName:  item.InstanceName(),
			StartTime:     item.StartTime().Format("15:04:05"),
			EndTime:       item.EndTime().Format("15:04:05"),
			RequiredCount: item.RequiredCount(),
			Priority:      item.Priority(),
		})
	}

	return TemplateResponse{
		TemplateID:   template.TemplateID().String(),
		TenantID:     template.TenantID().String(),
		EventID:      template.EventID().String(),
		TemplateName: template.TemplateName(),
		Description:  template.Description(),
		Items:        items,
		CreatedAt:    template.CreatedAt().Format(time.RFC3339),
		UpdatedAt:    template.UpdatedAt().Format(time.RFC3339),
	}
}
