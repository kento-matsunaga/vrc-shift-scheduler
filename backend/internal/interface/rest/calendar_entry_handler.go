package rest

import (
	"encoding/json"
	"net/http"

	appcalendar "github.com/erenoa/vrc-shift-scheduler/backend/internal/app/calendar"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/go-chi/chi/v5"
)

// CalendarEntryHandler handles calendar entry HTTP requests
type CalendarEntryHandler struct {
	createEntryUC *appcalendar.CreateCalendarEntryUsecase
	listEntriesUC *appcalendar.ListCalendarEntriesUsecase
	updateEntryUC *appcalendar.UpdateCalendarEntryUsecase
	deleteEntryUC *appcalendar.DeleteCalendarEntryUsecase
}

// NewCalendarEntryHandler creates a new CalendarEntryHandler with injected usecases
func NewCalendarEntryHandler(
	createEntryUC *appcalendar.CreateCalendarEntryUsecase,
	listEntriesUC *appcalendar.ListCalendarEntriesUsecase,
	updateEntryUC *appcalendar.UpdateCalendarEntryUsecase,
	deleteEntryUC *appcalendar.DeleteCalendarEntryUsecase,
) *CalendarEntryHandler {
	return &CalendarEntryHandler{
		createEntryUC: createEntryUC,
		listEntriesUC: listEntriesUC,
		updateEntryUC: updateEntryUC,
		deleteEntryUC: deleteEntryUC,
	}
}

// CreateCalendarEntryRequest represents the request body for creating a calendar entry
type CreateCalendarEntryRequest struct {
	Title     string  `json:"title"`
	Date      string  `json:"date"`       // YYYY-MM-DD
	StartTime *string `json:"start_time"` // HH:MM (optional)
	EndTime   *string `json:"end_time"`   // HH:MM (optional)
	Note      string  `json:"note"`
}

// UpdateCalendarEntryRequest represents the request body for updating a calendar entry
type UpdateCalendarEntryRequest struct {
	Title     string  `json:"title"`
	Date      string  `json:"date"`       // YYYY-MM-DD
	StartTime *string `json:"start_time"` // HH:MM (optional)
	EndTime   *string `json:"end_time"`   // HH:MM (optional)
	Note      string  `json:"note"`
}

// CalendarEntryResponse represents a calendar entry in API responses
type CalendarEntryResponse struct {
	EntryID    string  `json:"entry_id"`
	CalendarID string  `json:"calendar_id"`
	Title      string  `json:"title"`
	Date       string  `json:"date"`
	StartTime  *string `json:"start_time,omitempty"`
	EndTime    *string `json:"end_time,omitempty"`
	Note       string  `json:"note"`
	CreatedAt  string  `json:"created_at"`
	UpdatedAt  string  `json:"updated_at"`
}

// CreateCalendarEntry handles POST /calendars/:calendar_id/entries
func (h *CalendarEntryHandler) CreateCalendarEntry(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tenantID, ok := GetTenantID(ctx)
	if !ok {
		RespondBadRequest(w, "tenant_id is required")
		return
	}

	calendarIDStr := chi.URLParam(r, "calendar_id")
	if calendarIDStr == "" {
		RespondBadRequest(w, "calendar_id is required")
		return
	}

	calendarID := common.CalendarID(calendarIDStr)
	if err := calendarID.Validate(); err != nil {
		RespondBadRequest(w, "Invalid calendar_id format")
		return
	}

	var req CreateCalendarEntryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondBadRequest(w, "Invalid request body")
		return
	}

	// Validation
	if req.Title == "" {
		RespondBadRequest(w, "title is required")
		return
	}
	if len(req.Title) > 255 {
		RespondBadRequest(w, "title must be 255 characters or less")
		return
	}
	if req.Date == "" {
		RespondBadRequest(w, "date is required")
		return
	}
	if len(req.Note) > 2000 {
		RespondBadRequest(w, "note must be 2000 characters or less")
		return
	}

	input := appcalendar.CreateCalendarEntryInput{
		TenantID:   tenantID.String(),
		CalendarID: calendarIDStr,
		Title:      req.Title,
		Date:       req.Date,
		StartTime:  req.StartTime,
		EndTime:    req.EndTime,
		Note:       req.Note,
	}

	output, err := h.createEntryUC.Execute(ctx, input)
	if err != nil {
		RespondDomainError(w, err)
		return
	}

	RespondCreated(w, toCalendarEntryResponse(output))
}

// ListCalendarEntries handles GET /calendars/:calendar_id/entries
func (h *CalendarEntryHandler) ListCalendarEntries(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tenantID, ok := GetTenantID(ctx)
	if !ok {
		RespondBadRequest(w, "tenant_id is required")
		return
	}

	calendarIDStr := chi.URLParam(r, "calendar_id")
	if calendarIDStr == "" {
		RespondBadRequest(w, "calendar_id is required")
		return
	}

	calendarID := common.CalendarID(calendarIDStr)
	if err := calendarID.Validate(); err != nil {
		RespondBadRequest(w, "Invalid calendar_id format")
		return
	}

	input := appcalendar.ListCalendarEntriesInput{
		TenantID:   tenantID.String(),
		CalendarID: calendarIDStr,
	}

	outputs, err := h.listEntriesUC.Execute(ctx, input)
	if err != nil {
		RespondDomainError(w, err)
		return
	}

	entries := make([]CalendarEntryResponse, 0, len(outputs))
	for _, output := range outputs {
		entries = append(entries, toCalendarEntryResponse(&output))
	}

	RespondSuccess(w, map[string]interface{}{
		"entries": entries,
		"count":   len(entries),
	})
}

// UpdateCalendarEntry handles PUT /calendars/:calendar_id/entries/:entry_id
func (h *CalendarEntryHandler) UpdateCalendarEntry(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tenantID, ok := GetTenantID(ctx)
	if !ok {
		RespondBadRequest(w, "tenant_id is required")
		return
	}

	entryIDStr := chi.URLParam(r, "entry_id")
	if entryIDStr == "" {
		RespondBadRequest(w, "entry_id is required")
		return
	}

	entryID := common.CalendarEntryID(entryIDStr)
	if err := entryID.Validate(); err != nil {
		RespondBadRequest(w, "Invalid entry_id format")
		return
	}

	var req UpdateCalendarEntryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondBadRequest(w, "Invalid request body")
		return
	}

	// Validation
	if req.Title == "" {
		RespondBadRequest(w, "title is required")
		return
	}
	if len(req.Title) > 255 {
		RespondBadRequest(w, "title must be 255 characters or less")
		return
	}
	if req.Date == "" {
		RespondBadRequest(w, "date is required")
		return
	}
	if len(req.Note) > 2000 {
		RespondBadRequest(w, "note must be 2000 characters or less")
		return
	}

	input := appcalendar.UpdateCalendarEntryInput{
		TenantID:  tenantID.String(),
		EntryID:   entryIDStr,
		Title:     req.Title,
		Date:      req.Date,
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
		Note:      req.Note,
	}

	output, err := h.updateEntryUC.Execute(ctx, input)
	if err != nil {
		RespondDomainError(w, err)
		return
	}

	RespondSuccess(w, toCalendarEntryResponse(output))
}

// DeleteCalendarEntry handles DELETE /calendars/:calendar_id/entries/:entry_id
func (h *CalendarEntryHandler) DeleteCalendarEntry(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tenantID, ok := GetTenantID(ctx)
	if !ok {
		RespondBadRequest(w, "tenant_id is required")
		return
	}

	entryIDStr := chi.URLParam(r, "entry_id")
	if entryIDStr == "" {
		RespondBadRequest(w, "entry_id is required")
		return
	}

	entryID := common.CalendarEntryID(entryIDStr)
	if err := entryID.Validate(); err != nil {
		RespondBadRequest(w, "Invalid entry_id format")
		return
	}

	input := appcalendar.DeleteCalendarEntryInput{
		TenantID: tenantID.String(),
		EntryID:  entryIDStr,
	}

	if err := h.deleteEntryUC.Execute(ctx, input); err != nil {
		RespondDomainError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// toCalendarEntryResponse converts CalendarEntryDTO to CalendarEntryResponse
func toCalendarEntryResponse(dto *appcalendar.CalendarEntryDTO) CalendarEntryResponse {
	return CalendarEntryResponse{
		EntryID:    dto.EntryID,
		CalendarID: dto.CalendarID,
		Title:      dto.Title,
		Date:       dto.Date,
		StartTime:  dto.StartTime,
		EndTime:    dto.EndTime,
		Note:       dto.Note,
		CreatedAt:  dto.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:  dto.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}
