package rest

import (
	"encoding/json"
	"net/http"
	"time"

	appcalendar "github.com/erenoa/vrc-shift-scheduler/backend/internal/app/calendar"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/go-chi/chi/v5"
)

// CalendarHandler handles calendar-related HTTP requests
type CalendarHandler struct {
	createCalendarUC       *appcalendar.CreateCalendarUsecase
	getCalendarUC          *appcalendar.GetCalendarUsecase
	listCalendarsUC        *appcalendar.ListCalendarsUsecase
	updateCalendarUC       *appcalendar.UpdateCalendarUsecase
	deleteCalendarUC       *appcalendar.DeleteCalendarUsecase
	getCalendarByTokenUC   *appcalendar.GetCalendarByTokenUsecase
}

// NewCalendarHandler creates a new CalendarHandler with injected usecases
func NewCalendarHandler(
	createCalendarUC *appcalendar.CreateCalendarUsecase,
	getCalendarUC *appcalendar.GetCalendarUsecase,
	listCalendarsUC *appcalendar.ListCalendarsUsecase,
	updateCalendarUC *appcalendar.UpdateCalendarUsecase,
	deleteCalendarUC *appcalendar.DeleteCalendarUsecase,
	getCalendarByTokenUC *appcalendar.GetCalendarByTokenUsecase,
) *CalendarHandler {
	return &CalendarHandler{
		createCalendarUC:       createCalendarUC,
		getCalendarUC:          getCalendarUC,
		listCalendarsUC:        listCalendarsUC,
		updateCalendarUC:       updateCalendarUC,
		deleteCalendarUC:       deleteCalendarUC,
		getCalendarByTokenUC:   getCalendarByTokenUC,
	}
}

// CreateCalendarRequest represents the request body for creating a calendar
type CreateCalendarRequest struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	EventIDs    []string `json:"event_ids"`
}

// UpdateCalendarRequest represents the request body for updating a calendar
type UpdateCalendarRequest struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	EventIDs    []string `json:"event_ids"`
	IsPublic    bool     `json:"is_public"`
}

// CalendarResponse represents a calendar in API responses
type CalendarResponse struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	IsPublic    bool      `json:"is_public"`
	PublicToken *string   `json:"public_token,omitempty"`
	PublicURL   *string   `json:"public_url,omitempty"`
	EventIDs    []string  `json:"event_ids"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// PublicCalendarResponse represents a public calendar with events
type PublicCalendarResponse struct {
	Title       string                `json:"title"`
	Description string                `json:"description"`
	Events      []PublicEventResponse `json:"events"`
}

// PublicEventResponse represents an event in public calendar
type PublicEventResponse struct {
	Title        string                    `json:"title"`
	Description  string                    `json:"description"`
	BusinessDays []PublicBusinessDayResponse `json:"business_days"`
}

// PublicBusinessDayResponse represents a business day in public calendar
type PublicBusinessDayResponse struct {
	Date      string `json:"date"`
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
}

// Create handles POST /api/v1/calendars
func (h *CalendarHandler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tenantID, ok := GetTenantID(ctx)
	if !ok {
		RespondBadRequest(w, "tenant_id is required")
		return
	}

	var req CreateCalendarRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondBadRequest(w, "Invalid request body")
		return
	}

	if req.Title == "" {
		RespondBadRequest(w, "title is required")
		return
	}
	if len(req.Title) > 255 {
		RespondBadRequest(w, "title must be 255 characters or less")
		return
	}
	if len(req.EventIDs) == 0 {
		RespondBadRequest(w, "at least one event_id is required")
		return
	}

	input := appcalendar.CreateCalendarInput{
		TenantID:    tenantID.String(),
		Title:       req.Title,
		Description: req.Description,
		EventIDs:    req.EventIDs,
	}

	output, err := h.createCalendarUC.Execute(ctx, input)
	if err != nil {
		RespondDomainError(w, err)
		return
	}

	RespondCreated(w, toCalendarResponse(output))
}

// List handles GET /api/v1/calendars
func (h *CalendarHandler) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tenantID, ok := GetTenantID(ctx)
	if !ok {
		RespondBadRequest(w, "tenant_id is required")
		return
	}

	input := appcalendar.ListCalendarsInput{
		TenantID: tenantID.String(),
	}

	outputs, err := h.listCalendarsUC.Execute(ctx, input)
	if err != nil {
		RespondDomainError(w, err)
		return
	}

	var calendars []CalendarResponse
	for _, output := range outputs {
		calendars = append(calendars, toCalendarResponse(output))
	}

	RespondSuccess(w, map[string]interface{}{
		"calendars": calendars,
		"count":     len(calendars),
	})
}

// GetByID handles GET /api/v1/calendars/:id
func (h *CalendarHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tenantID, ok := GetTenantID(ctx)
	if !ok {
		RespondBadRequest(w, "tenant_id is required")
		return
	}

	calendarIDStr := chi.URLParam(r, "id")
	if calendarIDStr == "" {
		RespondBadRequest(w, "calendar id is required")
		return
	}

	calendarID := common.CalendarID(calendarIDStr)
	if err := calendarID.Validate(); err != nil {
		RespondBadRequest(w, "Invalid calendar id format")
		return
	}

	input := appcalendar.GetCalendarInput{
		TenantID:   tenantID.String(),
		CalendarID: calendarIDStr,
	}

	output, err := h.getCalendarUC.Execute(ctx, input)
	if err != nil {
		RespondDomainError(w, err)
		return
	}

	resp := CalendarResponse{
		ID:          output.Calendar.ID,
		Title:       output.Calendar.Title,
		Description: output.Calendar.Description,
		IsPublic:    output.Calendar.IsPublic,
		PublicToken: output.Calendar.PublicToken,
		EventIDs:    output.Calendar.EventIDs,
		CreatedAt:   output.Calendar.CreatedAt,
		UpdatedAt:   output.Calendar.UpdatedAt,
	}

	if output.Calendar.PublicToken != nil {
		publicURL := "/api/v1/public/calendar/" + *output.Calendar.PublicToken
		resp.PublicURL = &publicURL
	}

	RespondSuccess(w, resp)
}

// Update handles PUT /api/v1/calendars/:id
func (h *CalendarHandler) Update(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tenantID, ok := GetTenantID(ctx)
	if !ok {
		RespondBadRequest(w, "tenant_id is required")
		return
	}

	calendarIDStr := chi.URLParam(r, "id")
	if calendarIDStr == "" {
		RespondBadRequest(w, "calendar id is required")
		return
	}

	calendarID := common.CalendarID(calendarIDStr)
	if err := calendarID.Validate(); err != nil {
		RespondBadRequest(w, "Invalid calendar id format")
		return
	}

	var req UpdateCalendarRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondBadRequest(w, "Invalid request body")
		return
	}

	if req.Title == "" {
		RespondBadRequest(w, "title is required")
		return
	}
	if len(req.Title) > 255 {
		RespondBadRequest(w, "title must be 255 characters or less")
		return
	}
	if len(req.EventIDs) == 0 {
		RespondBadRequest(w, "at least one event_id is required")
		return
	}

	input := appcalendar.UpdateCalendarInput{
		TenantID:    tenantID.String(),
		CalendarID:  calendarIDStr,
		Title:       req.Title,
		Description: req.Description,
		EventIDs:    req.EventIDs,
		IsPublic:    req.IsPublic,
	}

	output, err := h.updateCalendarUC.Execute(ctx, input)
	if err != nil {
		RespondDomainError(w, err)
		return
	}

	RespondSuccess(w, toCalendarResponse(output))
}

// Delete handles DELETE /api/v1/calendars/:id
func (h *CalendarHandler) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tenantID, ok := GetTenantID(ctx)
	if !ok {
		RespondBadRequest(w, "tenant_id is required")
		return
	}

	calendarIDStr := chi.URLParam(r, "id")
	if calendarIDStr == "" {
		RespondBadRequest(w, "calendar id is required")
		return
	}

	calendarID := common.CalendarID(calendarIDStr)
	if err := calendarID.Validate(); err != nil {
		RespondBadRequest(w, "Invalid calendar id format")
		return
	}

	input := appcalendar.DeleteCalendarInput{
		TenantID:   tenantID.String(),
		CalendarID: calendarIDStr,
	}

	if err := h.deleteCalendarUC.Execute(ctx, input); err != nil {
		RespondDomainError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetByPublicToken handles GET /api/v1/public/calendar/:token
func (h *CalendarHandler) GetByPublicToken(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	token := chi.URLParam(r, "token")
	if token == "" {
		RespondNotFound(w, "Calendar not found")
		return
	}

	input := appcalendar.GetCalendarByTokenInput{
		Token: token,
	}

	output, err := h.getCalendarByTokenUC.Execute(ctx, input)
	if err != nil {
		RespondNotFound(w, "Calendar not found")
		return
	}

	var events []PublicEventResponse
	for _, evt := range output.Events {
		var businessDays []PublicBusinessDayResponse
		for _, bd := range evt.BusinessDays {
			businessDays = append(businessDays, PublicBusinessDayResponse{
				Date:      bd.Date.Format("2006-01-02"),
				StartTime: bd.StartTime,
				EndTime:   bd.EndTime,
			})
		}
		events = append(events, PublicEventResponse{
			Title:        evt.Title,
			Description:  evt.Description,
			BusinessDays: businessDays,
		})
	}

	RespondSuccess(w, PublicCalendarResponse{
		Title:       output.Calendar.Title,
		Description: output.Calendar.Description,
		Events:      events,
	})
}

// toCalendarResponse converts CalendarOutput to CalendarResponse
func toCalendarResponse(output *appcalendar.CalendarOutput) CalendarResponse {
	resp := CalendarResponse{
		ID:          output.ID,
		Title:       output.Title,
		Description: output.Description,
		IsPublic:    output.IsPublic,
		PublicToken: output.PublicToken,
		EventIDs:    output.EventIDs,
		CreatedAt:   output.CreatedAt,
		UpdatedAt:   output.UpdatedAt,
	}

	if output.PublicToken != nil {
		publicURL := "/api/v1/public/calendar/" + *output.PublicToken
		resp.PublicURL = &publicURL
	}

	return resp
}
