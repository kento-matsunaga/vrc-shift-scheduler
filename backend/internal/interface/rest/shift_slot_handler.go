package rest

import (
	"encoding/json"
	"log"
	"net/http"

	appshift "github.com/erenoa/vrc-shift-scheduler/backend/internal/app/shift"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/event"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/shift"
	"github.com/go-chi/chi/v5"
)

// ShiftSlotHandler handles shift slot-related HTTP requests
type ShiftSlotHandler struct {
	createShiftSlotUC *appshift.CreateShiftSlotUsecase
	listShiftSlotsUC  *appshift.ListShiftSlotsUsecase
	getShiftSlotUC    *appshift.GetShiftSlotUsecase
	deleteShiftSlotUC *appshift.DeleteShiftSlotUsecase
}

// NewShiftSlotHandler creates a new ShiftSlotHandler with injected usecases
func NewShiftSlotHandler(
	createShiftSlotUC *appshift.CreateShiftSlotUsecase,
	listShiftSlotsUC *appshift.ListShiftSlotsUsecase,
	getShiftSlotUC *appshift.GetShiftSlotUsecase,
	deleteShiftSlotUC *appshift.DeleteShiftSlotUsecase,
) *ShiftSlotHandler {
	return &ShiftSlotHandler{
		createShiftSlotUC: createShiftSlotUC,
		listShiftSlotsUC:  listShiftSlotsUC,
		getShiftSlotUC:    getShiftSlotUC,
		deleteShiftSlotUC: deleteShiftSlotUC,
	}
}

// CreateShiftSlotRequest represents the request body for creating a shift slot
type CreateShiftSlotRequest struct {
	SlotName      string  `json:"slot_name"`
	InstanceID    *string `json:"instance_id,omitempty"` // optional - existing instance ID
	InstanceName  string  `json:"instance_name"`
	StartTime     string  `json:"start_time"` // HH:MM
	EndTime       string  `json:"end_time"`   // HH:MM
	RequiredCount int     `json:"required_count"`
	Priority      int     `json:"priority"`
}

// ShiftSlotResponse represents a shift slot in API responses
type ShiftSlotResponse struct {
	SlotID        string  `json:"slot_id"`
	TenantID      string  `json:"tenant_id"`
	BusinessDayID string  `json:"business_day_id"`
	SlotName      string  `json:"slot_name"`
	InstanceName  string  `json:"instance_name"`
	InstanceID    *string `json:"instance_id,omitempty"` // インスタンスへの参照（FK）
	StartTime     string  `json:"start_time"`
	EndTime       string  `json:"end_time"`
	RequiredCount int     `json:"required_count"`
	AssignedCount int     `json:"assigned_count"` // 実際の割り当て数
	Priority      int     `json:"priority"`
	IsOvernight   bool    `json:"is_overnight"`
	CreatedAt     string  `json:"created_at"`
	UpdatedAt     string  `json:"updated_at"`
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

	// 時刻のパース (HH:MM or HH:MM:SS 形式)
	startTime, err := ParseTimeFlexible(req.StartTime)
	if err != nil {
		writeError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", "開始時刻の形式が正しくありません（HH:MMまたはHH:MM:SS）", nil)
		return
	}

	endTime, err := ParseTimeFlexible(req.EndTime)
	if err != nil {
		writeError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", "終了時刻の形式が正しくありません（HH:MMまたはHH:MM:SS）", nil)
		return
	}

	// InstanceID のパース（オプショナル）
	var instanceID *shift.InstanceID
	if req.InstanceID != nil && *req.InstanceID != "" {
		parsedID, err := shift.ParseInstanceID(*req.InstanceID)
		if err != nil {
			writeError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", "Invalid instance_id format", nil)
			return
		}
		instanceID = &parsedID
	}

	// Usecaseの実行
	// Priority のデフォルト値はユースケース層で設定される
	input := appshift.CreateShiftSlotInput{
		TenantID:      tenantID,
		BusinessDayID: businessDayID,
		InstanceID:    instanceID,
		SlotName:      req.SlotName,
		InstanceName:  req.InstanceName,
		StartTime:     startTime,
		EndTime:       endTime,
		RequiredCount: req.RequiredCount,
		Priority:      req.Priority,
	}

	newSlot, err := h.createShiftSlotUC.Execute(ctx, input)
	if err != nil {
		log.Printf("CreateShiftSlot error: %+v", err)
		RespondDomainError(w, err)
		return
	}

	// レスポンス
	resp := ShiftSlotResponse{
		SlotID:        newSlot.SlotID().String(),
		TenantID:      newSlot.TenantID().String(),
		BusinessDayID: newSlot.BusinessDayID().String(),
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
	if newSlot.InstanceID() != nil {
		instanceIDStr := newSlot.InstanceID().String()
		resp.InstanceID = &instanceIDStr
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

	// Usecaseの実行
	input := appshift.ListShiftSlotsInput{
		TenantID:      tenantID,
		BusinessDayID: businessDayID,
	}

	slots, err := h.listShiftSlotsUC.Execute(ctx, input)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "ERR_INTERNAL", "Failed to fetch shift slots", nil)
		return
	}

	// レスポンス構築
	slotResponses := make([]ShiftSlotResponse, 0, len(slots))
	for _, s := range slots {
		resp := ShiftSlotResponse{
			SlotID:        s.Slot.SlotID().String(),
			TenantID:      s.Slot.TenantID().String(),
			BusinessDayID: s.Slot.BusinessDayID().String(),
			SlotName:      s.Slot.SlotName(),
			InstanceName:  s.Slot.InstanceName(),
			StartTime:     s.Slot.StartTime().Format("15:04:05"),
			EndTime:       s.Slot.EndTime().Format("15:04:05"),
			RequiredCount: s.Slot.RequiredCount(),
			AssignedCount: s.AssignedCount,
			Priority:      s.Slot.Priority(),
			IsOvernight:   s.Slot.IsOvernight(),
			CreatedAt:     s.Slot.CreatedAt().Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:     s.Slot.UpdatedAt().Format("2006-01-02T15:04:05Z07:00"),
		}
		if s.Slot.InstanceID() != nil {
			instanceIDStr := s.Slot.InstanceID().String()
			resp.InstanceID = &instanceIDStr
		}
		slotResponses = append(slotResponses, resp)
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

	// Usecaseの実行
	input := appshift.GetShiftSlotInput{
		TenantID: tenantID,
		SlotID:   slotID,
	}

	result, err := h.getShiftSlotUC.Execute(ctx, input)
	if err != nil {
		RespondDomainError(w, err)
		return
	}

	// レスポンス
	resp := ShiftSlotResponse{
		SlotID:        result.Slot.SlotID().String(),
		TenantID:      result.Slot.TenantID().String(),
		BusinessDayID: result.Slot.BusinessDayID().String(),
		SlotName:      result.Slot.SlotName(),
		InstanceName:  result.Slot.InstanceName(),
		StartTime:     result.Slot.StartTime().Format("15:04:05"),
		EndTime:       result.Slot.EndTime().Format("15:04:05"),
		RequiredCount: result.Slot.RequiredCount(),
		AssignedCount: result.AssignedCount,
		Priority:      result.Slot.Priority(),
		IsOvernight:   result.Slot.IsOvernight(),
		CreatedAt:     result.Slot.CreatedAt().Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:     result.Slot.UpdatedAt().Format("2006-01-02T15:04:05Z07:00"),
	}
	if result.Slot.InstanceID() != nil {
		instanceIDStr := result.Slot.InstanceID().String()
		resp.InstanceID = &instanceIDStr
	}

	writeSuccess(w, http.StatusOK, resp)
}

// DeleteShiftSlot handles DELETE /api/v1/shift-slots/:slot_id
func (h *ShiftSlotHandler) DeleteShiftSlot(w http.ResponseWriter, r *http.Request) {
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

	// Usecaseの実行
	input := appshift.DeleteShiftSlotInput{
		TenantID: tenantID,
		SlotID:   slotID,
	}

	if err := h.deleteShiftSlotUC.Execute(ctx, input); err != nil {
		log.Printf("DeleteShiftSlot error: %+v", err)
		RespondDomainError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
