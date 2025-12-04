package rest

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/event"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/shift"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/infra/db"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ShiftSlotHandler handles shift slot-related HTTP requests
type ShiftSlotHandler struct {
	slotRepo        *db.ShiftSlotRepository
	businessDayRepo *db.EventBusinessDayRepository
	assignmentRepo  *db.ShiftAssignmentRepository
}

// NewShiftSlotHandler creates a new ShiftSlotHandler
func NewShiftSlotHandler(dbPool *pgxpool.Pool) *ShiftSlotHandler {
	return &ShiftSlotHandler{
		slotRepo:        db.NewShiftSlotRepository(dbPool),
		businessDayRepo: db.NewEventBusinessDayRepository(dbPool),
		assignmentRepo:  db.NewShiftAssignmentRepository(dbPool),
	}
}

// CreateShiftSlotRequest represents the request body for creating a shift slot
type CreateShiftSlotRequest struct {
	PositionID    string `json:"position_id"`
	SlotName      string `json:"slot_name"`
	InstanceName  string `json:"instance_name"`
	StartTime     string `json:"start_time"` // HH:MM
	EndTime       string `json:"end_time"`   // HH:MM
	RequiredCount int    `json:"required_count"`
	Priority      int    `json:"priority"`
}

// ShiftSlotResponse represents a shift slot in API responses
type ShiftSlotResponse struct {
	SlotID         string `json:"slot_id"`
	TenantID       string `json:"tenant_id"`
	BusinessDayID  string `json:"business_day_id"`
	PositionID     string `json:"position_id"`
	SlotName       string `json:"slot_name"`
	InstanceName   string `json:"instance_name"`
	StartTime      string `json:"start_time"`
	EndTime        string `json:"end_time"`
	RequiredCount  int    `json:"required_count"`
	AssignedCount  int    `json:"assigned_count,omitempty"` // JOIN で取得する場合
	Priority       int    `json:"priority"`
	IsOvernight    bool   `json:"is_overnight"`
	CreatedAt      string `json:"created_at"`
	UpdatedAt      string `json:"updated_at"`
}

// CreateShiftSlot handles POST /api/v1/business-days/:business_day_id/shift-slots
func (h *ShiftSlotHandler) CreateShiftSlot(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// テナントIDの取得
	tenantID, ok := getTenantIDFromContext(ctx)
	if !ok {
		writeError(w, http.StatusForbidden, "ERR_FORBIDDEN", "Tenant ID is required", nil)
		return
	}

	// business_day_id の取得
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

	// BusinessDay の存在確認（tenant_id チェック）
	_, err = h.businessDayRepo.FindByID(ctx, tenantID, businessDayID)
	if err != nil {
		if err.Error() == "business day not found" {
			writeError(w, http.StatusNotFound, "ERR_NOT_FOUND", "Business day not found", nil)
			return
		}
		writeError(w, http.StatusInternalServerError, "ERR_INTERNAL", "Failed to fetch business day", nil)
		return
	}

	// リクエストボディのパース
	var req CreateShiftSlotRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", "Invalid request body", nil)
		return
	}

	// バリデーション
	if req.SlotName == "" {
		writeError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", "slot_name is required", nil)
		return
	}

	if req.InstanceName == "" {
		writeError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", "instance_name is required", nil)
		return
	}

	if req.StartTime == "" || req.EndTime == "" {
		writeError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", "start_time and end_time are required", nil)
		return
	}

	if req.RequiredCount < 1 {
		writeError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", "required_count must be at least 1", nil)
		return
	}

	// Position ID のパース
	positionID, err := shift.ParsePositionID(req.PositionID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", "Invalid position_id format", nil)
		return
	}

	// 時刻のパース (HH:MM:SS 形式)
	startTime, err := time.Parse("15:04:05", req.StartTime)
	if err != nil {
		writeError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", "Invalid start_time format (expected HH:MM:SS)", nil)
		return
	}

	endTime, err := time.Parse("15:04:05", req.EndTime)
	if err != nil {
		writeError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", "Invalid end_time format (expected HH:MM:SS)", nil)
		return
	}

	// ShiftSlot エンティティの作成
	newSlot, err := shift.NewShiftSlot(
		tenantID,
		businessDayID,
		positionID,
		req.SlotName,
		req.InstanceName,
		startTime,
		endTime,
		req.RequiredCount,
		req.Priority,
	)
	if err != nil {
		writeError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", err.Error(), nil)
		return
	}

	// 保存
	if err := h.slotRepo.Save(ctx, newSlot); err != nil {
		writeError(w, http.StatusInternalServerError, "ERR_INTERNAL", "Failed to create shift slot", nil)
		return
	}

	// レスポンス
	resp := ShiftSlotResponse{
		SlotID:        newSlot.SlotID().String(),
		TenantID:      newSlot.TenantID().String(),
		BusinessDayID: newSlot.BusinessDayID().String(),
		PositionID:    newSlot.PositionID().String(),
		SlotName:      newSlot.SlotName(),
		InstanceName:  newSlot.InstanceName(),
		StartTime:     newSlot.StartTime().Format("15:04:05"),
		EndTime:       newSlot.EndTime().Format("15:04:05"),
		RequiredCount: newSlot.RequiredCount(),
		Priority:      newSlot.Priority(),
		IsOvernight:   newSlot.IsOvernight(),
		CreatedAt:     newSlot.CreatedAt().Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:     newSlot.UpdatedAt().Format("2006-01-02T15:04:05Z07:00"),
	}

	writeSuccess(w, http.StatusCreated, resp)
}

// GetShiftSlots handles GET /api/v1/business-days/:business_day_id/shift-slots
func (h *ShiftSlotHandler) GetShiftSlots(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// テナントIDの取得
	tenantID, ok := getTenantIDFromContext(ctx)
	if !ok {
		writeError(w, http.StatusForbidden, "ERR_FORBIDDEN", "Tenant ID is required", nil)
		return
	}

	// business_day_id の取得
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

	// シフト枠一覧を取得
	slots, err := h.slotRepo.FindByBusinessDayID(ctx, tenantID, businessDayID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "ERR_INTERNAL", "Failed to fetch shift slots", nil)
		return
	}

	// レスポンス構築
	slotResponses := make([]ShiftSlotResponse, 0, len(slots))
	for _, s := range slots {
		// assigned_count を取得
		assignedCount, err := h.assignmentRepo.CountConfirmedBySlotID(ctx, tenantID, s.SlotID())
		if err != nil {
			writeError(w, http.StatusInternalServerError, "ERR_INTERNAL", "Failed to count assignments", nil)
			return
		}

		slotResponses = append(slotResponses, ShiftSlotResponse{
			SlotID:        s.SlotID().String(),
			TenantID:      s.TenantID().String(),
			BusinessDayID: s.BusinessDayID().String(),
			PositionID:    s.PositionID().String(),
			SlotName:      s.SlotName(),
			InstanceName:  s.InstanceName(),
			StartTime:     s.StartTime().Format("15:04:05"),
			EndTime:       s.EndTime().Format("15:04:05"),
			RequiredCount: s.RequiredCount(),
			AssignedCount: assignedCount,
			Priority:      s.Priority(),
			IsOvernight:   s.IsOvernight(),
			CreatedAt:     s.CreatedAt().Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:     s.UpdatedAt().Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	writeSuccess(w, http.StatusOK, map[string]interface{}{
		"shift_slots": slotResponses,
		"count":       len(slotResponses),
	})
}

// GetShiftSlotDetail handles GET /api/v1/shift-slots/:slot_id
func (h *ShiftSlotHandler) GetShiftSlotDetail(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// テナントIDの取得
	tenantID, ok := getTenantIDFromContext(ctx)
	if !ok {
		writeError(w, http.StatusForbidden, "ERR_FORBIDDEN", "Tenant ID is required", nil)
		return
	}

	// slot_id の取得
	slotIDStr := chi.URLParam(r, "slot_id")
	if slotIDStr == "" {
		writeError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", "slot_id is required", nil)
		return
	}

	slotID, err := shift.ParseSlotID(slotIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", "Invalid slot_id format", nil)
		return
	}

	// シフト枠の取得
	slot, err := h.slotRepo.FindByID(ctx, tenantID, slotID)
	if err != nil {
		if err.Error() == "shift slot not found" {
			writeError(w, http.StatusNotFound, "ERR_NOT_FOUND", "Shift slot not found", nil)
			return
		}
		writeError(w, http.StatusInternalServerError, "ERR_INTERNAL", "Failed to fetch shift slot", nil)
		return
	}

	// assigned_count を取得
	assignedCount, err := h.assignmentRepo.CountConfirmedBySlotID(ctx, tenantID, slot.SlotID())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "ERR_INTERNAL", "Failed to count assignments", nil)
		return
	}

	// レスポンス
	resp := ShiftSlotResponse{
		SlotID:        slot.SlotID().String(),
		TenantID:      slot.TenantID().String(),
		BusinessDayID: slot.BusinessDayID().String(),
		PositionID:    slot.PositionID().String(),
		SlotName:      slot.SlotName(),
		InstanceName:  slot.InstanceName(),
		StartTime:     slot.StartTime().Format("15:04:05"),
		EndTime:       slot.EndTime().Format("15:04:05"),
		RequiredCount: slot.RequiredCount(),
		AssignedCount: assignedCount,
		Priority:      slot.Priority(),
		IsOvernight:   slot.IsOvernight(),
		CreatedAt:     slot.CreatedAt().Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:     slot.UpdatedAt().Format("2006-01-02T15:04:05Z07:00"),
	}

	writeSuccess(w, http.StatusOK, resp)
}

