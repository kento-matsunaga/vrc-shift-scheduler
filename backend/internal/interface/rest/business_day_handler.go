package rest

import (
	"encoding/json"
	"net/http"
	"time"

	appevent "github.com/erenoa/vrc-shift-scheduler/backend/internal/app/event"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/event"
	"github.com/go-chi/chi/v5"
)

// BusinessDayHandler handles business day-related HTTP requests
type BusinessDayHandler struct {
	createBusinessDayUC *appevent.CreateBusinessDayUsecase
	listBusinessDaysUC  *appevent.ListBusinessDaysUsecase
	getBusinessDayUC    *appevent.GetBusinessDayUsecase
	applyTemplateUC     *appevent.ApplyTemplateUsecase
	deleteBusinessDayUC *appevent.DeleteBusinessDayUsecase
}

// NewBusinessDayHandler creates a new BusinessDayHandler with injected usecases
func NewBusinessDayHandler(
	createBusinessDayUC *appevent.CreateBusinessDayUsecase,
	listBusinessDaysUC *appevent.ListBusinessDaysUsecase,
	getBusinessDayUC *appevent.GetBusinessDayUsecase,
	applyTemplateUC *appevent.ApplyTemplateUsecase,
	deleteBusinessDayUC *appevent.DeleteBusinessDayUsecase,
) *BusinessDayHandler {
	return &BusinessDayHandler{
		createBusinessDayUC: createBusinessDayUC,
		listBusinessDaysUC:  listBusinessDaysUC,
		getBusinessDayUC:    getBusinessDayUC,
		applyTemplateUC:     applyTemplateUC,
		deleteBusinessDayUC: deleteBusinessDayUC,
	}
}

// CreateBusinessDayRequest represents the request body for creating a business day
type CreateBusinessDayRequest struct {
	TargetDate     string  `json:"target_date"`     // YYYY-MM-DD
	StartTime      string  `json:"start_time"`      // HH:MM
	EndTime        string  `json:"end_time"`        // HH:MM
	OccurrenceType string  `json:"occurrence_type"` // recurring or special
	TemplateID     *string `json:"template_id"`     // optional: テンプレートからシフト枠を作成
}

// BusinessDayResponse represents a business day in API responses
type BusinessDayResponse struct {
	BusinessDayID  string `json:"business_day_id"`
	TenantID       string `json:"tenant_id"`
	EventID        string `json:"event_id"`
	TargetDate     string `json:"target_date"` // YYYY-MM-DD
	StartTime      string `json:"start_time"`  // HH:MM:SS
	EndTime        string `json:"end_time"`    // HH:MM:SS
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

	// テンプレートIDのパース
	var templateID *common.ShiftSlotTemplateID
	if req.TemplateID != nil && *req.TemplateID != "" {
		tid := common.ShiftSlotTemplateID(*req.TemplateID)
		if err := tid.Validate(); err != nil {
			RespondBadRequest(w, "Invalid template_id format")
			return
		}
		templateID = &tid
	}

	// Usecaseの実行
	input := appevent.CreateBusinessDayInput{
		TenantID:       tenantID,
		EventID:        eventID,
		TargetDate:     targetDate,
		StartTime:      startTime,
		EndTime:        endTime,
		OccurrenceType: event.OccurrenceTypeSpecial, // 手動作成は常にspecial
		TemplateID:     templateID,
	}

	newBusinessDay, err := h.createBusinessDayUC.Execute(ctx, input)
	if err != nil {
		RespondDomainError(w, err)
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

	var startDate, endDate *time.Time
	if startDateStr != "" && endDateStr != "" {
		parsedStartDate, parseErr := time.Parse("2006-01-02", startDateStr)
		if parseErr != nil {
			RespondBadRequest(w, "Invalid start_date format (expected YYYY-MM-DD)")
			return
		}

		parsedEndDate, parseErr := time.Parse("2006-01-02", endDateStr)
		if parseErr != nil {
			RespondBadRequest(w, "Invalid end_date format (expected YYYY-MM-DD)")
			return
		}

		startDate = &parsedStartDate
		endDate = &parsedEndDate
	}

	// Usecaseの実行
	input := appevent.ListBusinessDaysInput{
		TenantID:  tenantID,
		EventID:   eventID,
		StartDate: startDate,
		EndDate:   endDate,
	}

	businessDays, err := h.listBusinessDaysUC.Execute(ctx, input)
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

	// Usecaseの実行
	input := appevent.GetBusinessDayInput{
		TenantID:      tenantID,
		BusinessDayID: businessDayID,
	}

	foundBusinessDay, err := h.getBusinessDayUC.Execute(ctx, input)
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

// ApplyTemplateRequest represents the request body for applying a template to a business day
type ApplyTemplateRequest struct {
	TemplateID string `json:"template_id"`
}

// ApplyTemplate handles POST /api/v1/business-days/:business_day_id/apply-template
func (h *BusinessDayHandler) ApplyTemplate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// テナントIDの取得
	tenantID, ok := GetTenantID(ctx)
	if !ok {
		RespondBadRequest(w, "tenant_id is required")
		return
	}

	// 営業日IDの取得
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

	// リクエストボディのパース
	var req ApplyTemplateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondBadRequest(w, "Invalid request body")
		return
	}

	// テンプレートIDのバリデーション
	if req.TemplateID == "" {
		RespondBadRequest(w, "template_id is required")
		return
	}

	templateID := common.ShiftSlotTemplateID(req.TemplateID)
	if err := templateID.Validate(); err != nil {
		RespondBadRequest(w, "Invalid template_id format")
		return
	}

	// Usecaseの実行
	input := appevent.ApplyTemplateInput{
		TenantID:      tenantID,
		BusinessDayID: businessDayID,
		TemplateID:    templateID,
	}

	itemsCount, err := h.applyTemplateUC.Execute(ctx, input)
	if err != nil {
		RespondDomainError(w, err)
		return
	}

	// 成功レスポンス
	RespondSuccess(w, map[string]interface{}{
		"message":         "Template applied successfully",
		"business_day_id": businessDayID.String(),
		"template_id":     templateID.String(),
		"items_count":     itemsCount,
	})
}

// DeleteBusinessDay handles DELETE /api/v1/business-days/:business_day_id
func (h *BusinessDayHandler) DeleteBusinessDay(w http.ResponseWriter, r *http.Request) {
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

	// Usecaseの実行
	input := appevent.DeleteBusinessDayInput{
		TenantID:      tenantID,
		BusinessDayID: businessDayID,
	}

	if err := h.deleteBusinessDayUC.Execute(ctx, input); err != nil {
		RespondDomainError(w, err)
		return
	}

	// 成功レスポンス（204 No Content）
	w.WriteHeader(http.StatusNoContent)
}
