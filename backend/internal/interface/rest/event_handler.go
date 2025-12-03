package rest

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/event"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/infra/db"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// EventHandler handles event-related HTTP requests
type EventHandler struct {
	eventRepo *db.EventRepository
}

// NewEventHandler creates a new EventHandler
func NewEventHandler(dbPool *pgxpool.Pool) *EventHandler {
	return &EventHandler{
		eventRepo: db.NewEventRepository(dbPool),
	}
}

// CreateEventRequest represents the request body for creating an event
type CreateEventRequest struct {
	EventName   string `json:"event_name"`
	EventType   string `json:"event_type"`
	Description string `json:"description"`
}

// EventResponse represents an event in API responses
type EventResponse struct {
	EventID     string `json:"event_id"`
	TenantID    string `json:"tenant_id"`
	EventName   string `json:"event_name"`
	EventType   string `json:"event_type"`
	Description string `json:"description"`
	IsActive    bool   `json:"is_active"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
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

	// EventType のデフォルト値
	eventType := event.EventTypeNormal
	if req.EventType != "" {
		eventType = event.EventType(req.EventType)
	}

	// イベント名の重複チェック
	exists, err := h.eventRepo.ExistsByName(ctx, tenantID, req.EventName)
	if err != nil {
		RespondInternalError(w)
		return
	}
	if exists {
		RespondConflict(w, "Event with this name already exists")
		return
	}

	// イベントの作成
	newEvent, err := event.NewEvent(tenantID, req.EventName, eventType, req.Description)
	if err != nil {
		RespondDomainError(w, err)
		return
	}

	// 保存
	if err := h.eventRepo.Save(ctx, newEvent); err != nil {
		RespondInternalError(w)
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

	// イベント一覧の取得
	events, err := h.eventRepo.FindByTenantID(ctx, tenantID)
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

	// イベントの取得
	foundEvent, err := h.eventRepo.FindByID(ctx, tenantID, eventID)
	if err != nil {
		RespondDomainError(w, err)
		return
	}

	// レスポンス
	RespondSuccess(w, toEventResponse(foundEvent))
}

// toEventResponse converts an Event entity to EventResponse
func toEventResponse(e *event.Event) EventResponse {
	return EventResponse{
		EventID:     e.EventID().String(),
		TenantID:    e.TenantID().String(),
		EventName:   e.EventName(),
		EventType:   string(e.EventType()),
		Description: e.Description(),
		IsActive:    e.IsActive(),
		CreatedAt:   e.CreatedAt().Format(time.RFC3339),
		UpdatedAt:   e.UpdatedAt().Format(time.RFC3339),
	}
}

