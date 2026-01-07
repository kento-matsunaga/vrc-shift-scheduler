package rest

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	appevent "github.com/erenoa/vrc-shift-scheduler/backend/internal/app/event"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/event"
	"github.com/go-chi/chi/v5"
)

// EventHandler handles event-related HTTP requests
type EventHandler struct {
	createEventUC            *appevent.CreateEventUsecase
	listEventsUC             *appevent.ListEventsUsecase
	getEventUC               *appevent.GetEventUsecase
	updateEventUC            *appevent.UpdateEventUsecase
	deleteEventUC            *appevent.DeleteEventUsecase
	generateBusinessDaysUC   *appevent.GenerateBusinessDaysUsecase
	getGroupAssignmentsUC    *appevent.GetEventGroupAssignmentsUsecase
	updateGroupAssignmentsUC *appevent.UpdateEventGroupAssignmentsUsecase
}

// NewEventHandler creates a new EventHandler with injected usecases
func NewEventHandler(
	createEventUC *appevent.CreateEventUsecase,
	listEventsUC *appevent.ListEventsUsecase,
	getEventUC *appevent.GetEventUsecase,
	updateEventUC *appevent.UpdateEventUsecase,
	deleteEventUC *appevent.DeleteEventUsecase,
	generateBusinessDaysUC *appevent.GenerateBusinessDaysUsecase,
	getGroupAssignmentsUC *appevent.GetEventGroupAssignmentsUsecase,
	updateGroupAssignmentsUC *appevent.UpdateEventGroupAssignmentsUsecase,
) *EventHandler {
	return &EventHandler{
		createEventUC:            createEventUC,
		listEventsUC:             listEventsUC,
		getEventUC:               getEventUC,
		updateEventUC:            updateEventUC,
		deleteEventUC:            deleteEventUC,
		generateBusinessDaysUC:   generateBusinessDaysUC,
		getGroupAssignmentsUC:    getGroupAssignmentsUC,
		updateGroupAssignmentsUC: updateGroupAssignmentsUC,
	}
}

// CreateEventRequest represents the request body for creating an event
type CreateEventRequest struct {
	EventName           string  `json:"event_name"`
	EventType           string  `json:"event_type"`
	Description         string  `json:"description"`
	RecurrenceType      string  `json:"recurrence_type,omitempty"`       // "none", "weekly", "biweekly"
	RecurrenceStartDate *string `json:"recurrence_start_date,omitempty"` // YYYY-MM-DD
	RecurrenceDayOfWeek *int    `json:"recurrence_day_of_week,omitempty"`// 0-6
	DefaultStartTime    *string `json:"default_start_time,omitempty"`    // HH:MM:SS
	DefaultEndTime      *string `json:"default_end_time,omitempty"`      // HH:MM:SS
}

// UpdateEventRequest represents the request body for updating an event
type UpdateEventRequest struct {
	EventName string `json:"event_name"`
}

// EventResponse represents an event in API responses
type EventResponse struct {
	EventID             string  `json:"event_id"`
	TenantID            string  `json:"tenant_id"`
	EventName           string  `json:"event_name"`
	EventType           string  `json:"event_type"`
	Description         string  `json:"description"`
	IsActive            bool    `json:"is_active"`
	RecurrenceType      string  `json:"recurrence_type"`
	RecurrenceStartDate *string `json:"recurrence_start_date,omitempty"`
	RecurrenceDayOfWeek *int    `json:"recurrence_day_of_week,omitempty"`
	DefaultStartTime    *string `json:"default_start_time,omitempty"`
	DefaultEndTime      *string `json:"default_end_time,omitempty"`
	CreatedAt           string  `json:"created_at"`
	UpdatedAt           string  `json:"updated_at"`
}

// CreateEvent handles POST /api/v1/events
func (h *EventHandler) CreateEvent(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// テナントIDの取得
	tenantID, ok := GetTenantID(ctx)
	if !ok {
		RespondBadRequest(w, "tenant_id is required")
		return
	}

	// リクエストボディのパース
	var req CreateEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondBadRequest(w, "Invalid request body")
		return
	}

	// バリデーション
	if req.EventName == "" {
		RespondBadRequest(w, "event_name is required")
		return
	}

	// 長さ制限（DoS対策）
	if len(req.EventName) > 255 {
		RespondBadRequest(w, "event_name must be 255 characters or less")
		return
	}
	if len(req.Description) > 2000 {
		RespondBadRequest(w, "description must be 2000 characters or less")
		return
	}

	// EventType のデフォルト値
	eventType := event.EventTypeNormal
	if req.EventType != "" {
		eventType = event.EventType(req.EventType)
	}

	// RecurrenceType のデフォルト値
	recurrenceType := event.RecurrenceTypeNone
	if req.RecurrenceType != "" {
		recurrenceType = event.RecurrenceType(req.RecurrenceType)
	}

	// 定期設定のパース
	var recurrenceStartDate *time.Time
	var recurrenceDayOfWeek *int
	var defaultStartTime *time.Time
	var defaultEndTime *time.Time

	if recurrenceType != event.RecurrenceTypeNone {
		// 定期開始日のパース
		if req.RecurrenceStartDate != nil {
			t, err := time.Parse("2006-01-02", *req.RecurrenceStartDate)
			if err != nil {
				RespondBadRequest(w, "Invalid recurrence_start_date format (expected YYYY-MM-DD)")
				return
			}
			recurrenceStartDate = &t
		}

		// 曜日のコピー
		recurrenceDayOfWeek = req.RecurrenceDayOfWeek

		// 開始時刻のパース
		if req.DefaultStartTime != nil {
			t, err := ParseTimeFlexible(*req.DefaultStartTime)
			if err != nil {
				RespondBadRequest(w, "デフォルト開始時刻の形式が正しくありません（HH:MMまたはHH:MM:SS）")
				return
			}
			defaultStartTime = &t
		}

		// 終了時刻のパース
		if req.DefaultEndTime != nil {
			t, err := ParseTimeFlexible(*req.DefaultEndTime)
			if err != nil {
				RespondBadRequest(w, "デフォルト終了時刻の形式が正しくありません（HH:MMまたはHH:MM:SS）")
				return
			}
			defaultEndTime = &t
		}
	}

	// Usecaseの実行
	input := appevent.CreateEventInput{
		TenantID:            tenantID,
		EventName:           req.EventName,
		EventType:           eventType,
		Description:         req.Description,
		RecurrenceType:      recurrenceType,
		RecurrenceStartDate: recurrenceStartDate,
		RecurrenceDayOfWeek: recurrenceDayOfWeek,
		DefaultStartTime:    defaultStartTime,
		DefaultEndTime:      defaultEndTime,
	}

	newEvent, err := h.createEventUC.Execute(ctx, input)
	if err != nil {
		RespondDomainError(w, err)
		return
	}

	// レスポンス
	RespondCreated(w, toEventResponse(newEvent))
}

// ListEvents handles GET /api/v1/events
func (h *EventHandler) ListEvents(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// テナントIDの取得
	tenantID, ok := GetTenantID(ctx)
	if !ok {
		RespondBadRequest(w, "tenant_id is required")
		return
	}

	// Usecaseの実行
	input := appevent.ListEventsInput{
		TenantID: tenantID,
	}

	events, err := h.listEventsUC.Execute(ctx, input)
	if err != nil {
		RespondInternalError(w)
		return
	}

	// レスポンス
	var eventResponses []EventResponse
	for _, e := range events {
		eventResponses = append(eventResponses, toEventResponse(e))
	}

	RespondSuccess(w, map[string]interface{}{
		"events": eventResponses,
		"count":  len(eventResponses),
	})
}

// GetEvent handles GET /api/v1/events/:event_id
func (h *EventHandler) GetEvent(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// テナントIDの取得
	tenantID, ok := GetTenantID(ctx)
	if !ok {
		RespondBadRequest(w, "tenant_id is required")
		return
	}

	// イベントIDの取得
	eventIDStr := chi.URLParam(r, "event_id")
	if eventIDStr == "" {
		RespondBadRequest(w, "event_id is required")
		return
	}

	eventID := common.EventID(eventIDStr)
	if err := eventID.Validate(); err != nil {
		RespondBadRequest(w, "Invalid event_id format")
		return
	}

	// Usecaseの実行
	input := appevent.GetEventInput{
		TenantID: tenantID,
		EventID:  eventID,
	}

	foundEvent, err := h.getEventUC.Execute(ctx, input)
	if err != nil {
		RespondDomainError(w, err)
		return
	}

	// レスポンス
	RespondSuccess(w, toEventResponse(foundEvent))
}

// UpdateEvent handles PUT /api/v1/events/:event_id
func (h *EventHandler) UpdateEvent(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// テナントIDの取得
	tenantID, ok := GetTenantID(ctx)
	if !ok {
		RespondBadRequest(w, "tenant_id is required")
		return
	}

	// イベントIDの取得
	eventIDStr := chi.URLParam(r, "event_id")
	if eventIDStr == "" {
		RespondBadRequest(w, "event_id is required")
		return
	}

	eventID := common.EventID(eventIDStr)
	if err := eventID.Validate(); err != nil {
		RespondBadRequest(w, "Invalid event_id format")
		return
	}

	// リクエストボディのパース
	var req UpdateEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondBadRequest(w, "Invalid request body")
		return
	}

	// バリデーション
	if req.EventName == "" {
		RespondBadRequest(w, "event_name is required")
		return
	}

	// 長さ制限
	if len(req.EventName) > 255 {
		RespondBadRequest(w, "event_name must be 255 characters or less")
		return
	}

	// Usecaseの実行
	input := appevent.UpdateEventInput{
		TenantID:  tenantID,
		EventID:   eventID,
		EventName: req.EventName,
	}

	updatedEvent, err := h.updateEventUC.Execute(ctx, input)
	if err != nil {
		RespondDomainError(w, err)
		return
	}

	// レスポンス
	RespondSuccess(w, toEventResponse(updatedEvent))
}

// DeleteEvent handles DELETE /api/v1/events/:event_id
func (h *EventHandler) DeleteEvent(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// テナントIDの取得
	tenantID, ok := GetTenantID(ctx)
	if !ok {
		RespondBadRequest(w, "tenant_id is required")
		return
	}

	// イベントIDの取得
	eventIDStr := chi.URLParam(r, "event_id")
	if eventIDStr == "" {
		RespondBadRequest(w, "event_id is required")
		return
	}

	eventID := common.EventID(eventIDStr)
	if err := eventID.Validate(); err != nil {
		RespondBadRequest(w, "Invalid event_id format")
		return
	}

	// Usecaseの実行
	input := appevent.DeleteEventInput{
		TenantID: tenantID,
		EventID:  eventID,
	}

	if err := h.deleteEventUC.Execute(ctx, input); err != nil {
		RespondDomainError(w, err)
		return
	}

	// 204 No Content で応答
	w.WriteHeader(http.StatusNoContent)
}

// toEventResponse converts an Event entity to EventResponse
func toEventResponse(e *event.Event) EventResponse {
	resp := EventResponse{
		EventID:        e.EventID().String(),
		TenantID:       e.TenantID().String(),
		EventName:      e.EventName(),
		EventType:      string(e.EventType()),
		Description:    e.Description(),
		IsActive:       e.IsActive(),
		RecurrenceType: string(e.RecurrenceType()),
		CreatedAt:      e.CreatedAt().Format(time.RFC3339),
		UpdatedAt:      e.UpdatedAt().Format(time.RFC3339),
	}

	// 定期設定フィールドの変換
	if e.RecurrenceStartDate() != nil {
		dateStr := e.RecurrenceStartDate().Format("2006-01-02")
		resp.RecurrenceStartDate = &dateStr
	}

	resp.RecurrenceDayOfWeek = e.RecurrenceDayOfWeek()

	if e.DefaultStartTime() != nil {
		timeStr := e.DefaultStartTime().Format("15:04:05")
		resp.DefaultStartTime = &timeStr
	}

	if e.DefaultEndTime() != nil {
		timeStr := e.DefaultEndTime().Format("15:04:05")
		resp.DefaultEndTime = &timeStr
	}

	return resp
}

// GenerateBusinessDaysResponse represents the response for generating business days
type GenerateBusinessDaysResponse struct {
	GeneratedCount int           `json:"generated_count"`
	Message        string        `json:"message"`
	Event          EventResponse `json:"event"`
}

// GenerateBusinessDays handles POST /api/v1/events/:event_id/generate-business-days
func (h *EventHandler) GenerateBusinessDays(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// テナントIDの取得
	tenantID, ok := GetTenantID(ctx)
	if !ok {
		RespondBadRequest(w, "tenant_id is required")
		return
	}

	// イベントIDの取得
	eventIDStr := chi.URLParam(r, "event_id")
	if eventIDStr == "" {
		RespondBadRequest(w, "event_id is required")
		return
	}

	eventID := common.EventID(eventIDStr)
	if err := eventID.Validate(); err != nil {
		RespondBadRequest(w, "Invalid event_id format")
		return
	}

	// Usecaseの実行
	input := appevent.GenerateBusinessDaysInput{
		TenantID: tenantID,
		EventID:  eventID,
	}

	output, err := h.generateBusinessDaysUC.Execute(ctx, input)
	if err != nil {
		log.Printf("GenerateBusinessDays error for event %s, tenant %s: %v", eventID, tenantID, err)
		RespondDomainError(w, err)
		return
	}

	// メッセージを生成
	message := "営業日の生成が完了しました"
	if output.GeneratedCount == 0 {
		message = "新しい営業日はありませんでした（既に生成済み）"
	}

	// レスポンス
	RespondSuccess(w, GenerateBusinessDaysResponse{
		GeneratedCount: output.GeneratedCount,
		Message:        message,
		Event:          toEventResponse(output.Event),
	})
}

// EventGroupAssignmentsRequest represents the request body for updating group assignments
type EventGroupAssignmentsRequest struct {
	MemberGroupIDs []string `json:"member_group_ids"`
	RoleGroupIDs   []string `json:"role_group_ids"`
}

// EventGroupAssignmentsResponse represents the response for group assignments
type EventGroupAssignmentsResponse struct {
	MemberGroupIDs []string `json:"member_group_ids"`
	RoleGroupIDs   []string `json:"role_group_ids"`
}

// GetGroupAssignments handles GET /api/v1/events/:event_id/groups
func (h *EventHandler) GetGroupAssignments(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// テナントIDの取得
	tenantID, ok := GetTenantID(ctx)
	if !ok {
		RespondBadRequest(w, "tenant_id is required")
		return
	}

	// イベントIDの取得
	eventIDStr := chi.URLParam(r, "event_id")
	if eventIDStr == "" {
		RespondBadRequest(w, "event_id is required")
		return
	}

	eventID := common.EventID(eventIDStr)
	if err := eventID.Validate(); err != nil {
		RespondBadRequest(w, "Invalid event_id format")
		return
	}

	// Usecaseの実行
	input := appevent.GetEventGroupAssignmentsInput{
		TenantID: tenantID,
		EventID:  eventID,
	}

	output, err := h.getGroupAssignmentsUC.Execute(ctx, input)
	if err != nil {
		RespondDomainError(w, err)
		return
	}

	// レスポンス
	RespondSuccess(w, EventGroupAssignmentsResponse{
		MemberGroupIDs: output.MemberGroupIDs,
		RoleGroupIDs:   output.RoleGroupIDs,
	})
}

// UpdateGroupAssignments handles PUT /api/v1/events/:event_id/groups
func (h *EventHandler) UpdateGroupAssignments(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// テナントIDの取得
	tenantID, ok := GetTenantID(ctx)
	if !ok {
		RespondBadRequest(w, "tenant_id is required")
		return
	}

	// イベントIDの取得
	eventIDStr := chi.URLParam(r, "event_id")
	if eventIDStr == "" {
		RespondBadRequest(w, "event_id is required")
		return
	}

	eventID := common.EventID(eventIDStr)
	if err := eventID.Validate(); err != nil {
		RespondBadRequest(w, "Invalid event_id format")
		return
	}

	// リクエストボディのパース
	var req EventGroupAssignmentsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondBadRequest(w, "Invalid request body")
		return
	}

	// グループIDの最大数制限（DoS対策）
	if len(req.MemberGroupIDs) > 100 || len(req.RoleGroupIDs) > 100 {
		RespondBadRequest(w, "Too many group IDs (max 100)")
		return
	}

	// Usecaseの実行
	input := appevent.UpdateEventGroupAssignmentsInput{
		TenantID:       tenantID,
		EventID:        eventID,
		MemberGroupIDs: req.MemberGroupIDs,
		RoleGroupIDs:   req.RoleGroupIDs,
	}

	if err := h.updateGroupAssignmentsUC.Execute(ctx, input); err != nil {
		RespondDomainError(w, err)
		return
	}

	// 更新後のデータを取得して返す
	getInput := appevent.GetEventGroupAssignmentsInput{
		TenantID: tenantID,
		EventID:  eventID,
	}

	output, err := h.getGroupAssignmentsUC.Execute(ctx, getInput)
	if err != nil {
		RespondDomainError(w, err)
		return
	}

	RespondSuccess(w, EventGroupAssignmentsResponse{
		MemberGroupIDs: output.MemberGroupIDs,
		RoleGroupIDs:   output.RoleGroupIDs,
	})
}

