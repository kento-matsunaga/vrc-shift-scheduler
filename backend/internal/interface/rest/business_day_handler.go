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

// BusinessDayHandler handles business day-related HTTP requests
type BusinessDayHandler struct {
	businessDayRepo *db.EventBusinessDayRepository
	eventRepo       *db.EventRepository
}

// NewBusinessDayHandler creates a new BusinessDayHandler
func NewBusinessDayHandler(dbPool *pgxpool.Pool) *BusinessDayHandler {
	return &BusinessDayHandler{
		businessDayRepo: db.NewEventBusinessDayRepository(dbPool),
		eventRepo:       db.NewEventRepository(dbPool),
	}
}

// CreateBusinessDayRequest represents the request body for creating a business day
type CreateBusinessDayRequest struct {
	TargetDate     string `json:"target_date"`      // YYYY-MM-DD
	StartTime      string `json:"start_time"`       // HH:MM
	EndTime        string `json:"end_time"`         // HH:MM
	OccurrenceType string `json:"occurrence_type"`  // recurring or special
}

// BusinessDayResponse represents a business day in API responses
type BusinessDayResponse struct {
	BusinessDayID  string `json:"business_day_id"`
	TenantID       string `json:"tenant_id"`
	EventID        string `json:"event_id"`
	TargetDate     string `json:"target_date"`      // YYYY-MM-DD
	StartTime      string `json:"start_time"`       // HH:MM:SS
	EndTime        string `json:"end_time"`         // HH:MM:SS
	OccurrenceType string `json:"occurrence_type"`
	IsActive       bool   `json:"is_active"`
	CreatedAt      string `json:"created_at"`
	UpdatedAt      string `json:"updated_at"`
}

// CreateBusinessDay handles POST /api/v1/events/:event_id/business-days
func (h *BusinessDayHandler) CreateBusinessDay(w http.ResponseWriter, r *http.Request) {
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

	// イベントの存在確認
	_, err := h.eventRepo.FindByID(ctx, tenantID, eventID)
	if err != nil {
		RespondDomainError(w, err)
		return
	}

	// リクエストボディのパース
	var req CreateBusinessDayRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondBadRequest(w, "Invalid request body")
		return
	}

	// バリデーション
	if req.TargetDate == "" || req.StartTime == "" || req.EndTime == "" {
		RespondBadRequest(w, "target_date, start_time, and end_time are required")
		return
	}

	// 日付と時刻のパース
	targetDate, err := time.Parse("2006-01-02", req.TargetDate)
	if err != nil {
		RespondBadRequest(w, "Invalid target_date format (expected YYYY-MM-DD)")
		return
	}

	startTime, err := time.Parse("15:04", req.StartTime)
	if err != nil {
		RespondBadRequest(w, "Invalid start_time format (expected HH:MM)")
		return
	}

	endTime, err := time.Parse("15:04", req.EndTime)
	if err != nil {
		RespondBadRequest(w, "Invalid end_time format (expected HH:MM)")
		return
	}

	// OccurrenceType のデフォルト値
	occurrenceType := event.OccurrenceTypeSpecial // 手動作成の場合は special
	if req.OccurrenceType != "" {
		occurrenceType = event.OccurrenceType(req.OccurrenceType)
	}

	// 重複チェック
	exists, err := h.businessDayRepo.ExistsByEventIDAndDate(ctx, tenantID, eventID, targetDate, startTime)
	if err != nil {
		RespondInternalError(w)
		return
	}
	if exists {
		RespondConflict(w, "Business day already exists for this date and time")
		return
	}

	// BusinessDay の作成
	newBusinessDay, err := event.NewEventBusinessDay(
		tenantID,
		eventID,
		targetDate,
		startTime,
		endTime,
		occurrenceType,
		nil, // recurring_pattern_id は手動作成では nil
	)
	if err != nil {
		RespondDomainError(w, err)
		return
	}

	// 保存
	if err := h.businessDayRepo.Save(ctx, newBusinessDay); err != nil {
		RespondInternalError(w)
		return
	}

	// レスポンス
	RespondCreated(w, toBusinessDayResponse(newBusinessDay))
}

// ListBusinessDays handles GET /api/v1/events/:event_id/business-days
func (h *BusinessDayHandler) ListBusinessDays(w http.ResponseWriter, r *http.Request) {
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

	// クエリパラメータの取得（日付範囲フィルタ）
	startDateStr := r.URL.Query().Get("start_date")
	endDateStr := r.URL.Query().Get("end_date")

	var businessDays []*event.EventBusinessDay
	var err error

	if startDateStr != "" && endDateStr != "" {
		// 日付範囲で検索
		startDate, parseErr := time.Parse("2006-01-02", startDateStr)
		if parseErr != nil {
			RespondBadRequest(w, "Invalid start_date format (expected YYYY-MM-DD)")
			return
		}

		endDate, parseErr := time.Parse("2006-01-02", endDateStr)
		if parseErr != nil {
			RespondBadRequest(w, "Invalid end_date format (expected YYYY-MM-DD)")
			return
		}

		businessDays, err = h.businessDayRepo.FindByEventIDAndDateRange(ctx, tenantID, eventID, startDate, endDate)
	} else {
		// 全件取得
		businessDays, err = h.businessDayRepo.FindByEventID(ctx, tenantID, eventID)
	}

	if err != nil {
		RespondInternalError(w)
		return
	}

	// レスポンス
	var businessDayResponses []BusinessDayResponse
	for _, bd := range businessDays {
		businessDayResponses = append(businessDayResponses, toBusinessDayResponse(bd))
	}

	RespondSuccess(w, map[string]interface{}{
		"business_days": businessDayResponses,
		"count":         len(businessDayResponses),
	})
}

// GetBusinessDay handles GET /api/v1/business-days/:business_day_id
func (h *BusinessDayHandler) GetBusinessDay(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// テナントIDの取得
	tenantID, ok := GetTenantID(ctx)
	if !ok {
		RespondBadRequest(w, "tenant_id is required")
		return
	}

	// BusinessDayIDの取得
	businessDayIDStr := chi.URLParam(r, "business_day_id")
	if businessDayIDStr == "" {
		RespondBadRequest(w, "business_day_id is required")
		return
	}

	businessDayID := event.BusinessDayID(businessDayIDStr)
	if err := businessDayID.Validate(); err != nil {
		RespondBadRequest(w, "Invalid business_day_id format")
		return
	}

	// BusinessDayの取得
	foundBusinessDay, err := h.businessDayRepo.FindByID(ctx, tenantID, businessDayID)
	if err != nil {
		RespondDomainError(w, err)
		return
	}

	// レスポンス
	RespondSuccess(w, toBusinessDayResponse(foundBusinessDay))
}

// toBusinessDayResponse converts an EventBusinessDay entity to BusinessDayResponse
func toBusinessDayResponse(bd *event.EventBusinessDay) BusinessDayResponse {
	return BusinessDayResponse{
		BusinessDayID:  bd.BusinessDayID().String(),
		TenantID:       bd.TenantID().String(),
		EventID:        bd.EventID().String(),
		TargetDate:     bd.TargetDate().Format("2006-01-02"),
		StartTime:      bd.StartTime().Format("15:04:05"),
		EndTime:        bd.EndTime().Format("15:04:05"),
		OccurrenceType: string(bd.OccurrenceType()),
		IsActive:       bd.IsActive(),
		CreatedAt:      bd.CreatedAt().Format(time.RFC3339),
		UpdatedAt:      bd.UpdatedAt().Format(time.RFC3339),
	}
}

