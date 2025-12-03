package rest

import (
	"encoding/json"
	"net/http"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/app"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/shift"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/infra/db"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ShiftAssignmentHandler handles shift assignment-related HTTP requests
type ShiftAssignmentHandler struct {
	assignmentService *app.ShiftAssignmentService
	assignmentRepo    *db.ShiftAssignmentRepository
}

// NewShiftAssignmentHandler creates a new ShiftAssignmentHandler
func NewShiftAssignmentHandler(dbPool *pgxpool.Pool) *ShiftAssignmentHandler {
	return &ShiftAssignmentHandler{
		assignmentService: app.NewShiftAssignmentService(dbPool),
		assignmentRepo:    db.NewShiftAssignmentRepository(dbPool),
	}
}

// ConfirmAssignmentRequest represents the request body for confirming a shift assignment
type ConfirmAssignmentRequest struct {
	SlotID   string `json:"slot_id"`
	MemberID string `json:"member_id"`
	Note     string `json:"note"`
}

// ShiftAssignmentResponse represents a shift assignment in API responses
type ShiftAssignmentResponse struct {
	AssignmentID      string `json:"assignment_id"`
	SlotID            string `json:"slot_id"`
	MemberID          string `json:"member_id"`
	MemberDisplayName string `json:"member_display_name,omitempty"`
	SlotName          string `json:"slot_name,omitempty"`
	TargetDate        string `json:"target_date,omitempty"`
	StartTime         string `json:"start_time,omitempty"`
	EndTime           string `json:"end_time,omitempty"`
	AssignmentStatus  string `json:"assignment_status"`
	AssignmentMethod  string `json:"assignment_method"`
	AssignedAt        string `json:"assigned_at"`
	NotificationSent  bool   `json:"notification_sent"`
}

// ConfirmAssignment handles POST /api/v1/shift-assignments
func (h *ShiftAssignmentHandler) ConfirmAssignment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// テナントIDの取得
	tenantID, ok := getTenantIDFromContext(ctx)
	if !ok {
		writeError(w, http.StatusForbidden, "ERR_FORBIDDEN", "Tenant ID is required", nil)
		return
	}

	// アクター（操作者）IDの取得
	actorID, ok := getMemberIDFromContext(ctx)
	if !ok {
		writeError(w, http.StatusForbidden, "ERR_FORBIDDEN", "Member ID is required", nil)
		return
	}

	// リクエストボディのパース
	var req ConfirmAssignmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", "Invalid request body", nil)
		return
	}

	// バリデーション
	if req.SlotID == "" {
		writeError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", "slot_id is required", nil)
		return
	}

	if req.MemberID == "" {
		writeError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", "member_id is required", nil)
		return
	}

	// ID のパース
	slotID, err := shift.ParseSlotID(req.SlotID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", "Invalid slot_id format", nil)
		return
	}

	memberID, err := common.ParseMemberID(req.MemberID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", "Invalid member_id format", nil)
		return
	}

	// Application Service 経由でシフト確定
	assignment, err := h.assignmentService.ConfirmManualAssignment(
		ctx,
		tenantID,
		slotID,
		memberID,
		actorID,
		req.Note,
	)
	if err != nil {
		// ドメインエラーを適切な HTTP ステータスに変換
		if domainErr, ok := err.(*common.DomainError); ok {
			switch domainErr.Code() {
			case common.ErrConflict:
				writeError(w, http.StatusConflict, "ERR_SLOT_FULL", domainErr.Error(), nil)
				return
			case common.ErrNotFound:
				writeError(w, http.StatusNotFound, "ERR_NOT_FOUND", domainErr.Error(), nil)
				return
			case common.ErrInvalidInput:
				writeError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", domainErr.Error(), nil)
				return
			}
		}
		writeError(w, http.StatusInternalServerError, "ERR_INTERNAL", "Failed to confirm shift assignment", nil)
		return
	}

	// レスポンス
	resp := ShiftAssignmentResponse{
		AssignmentID:     assignment.AssignmentID().String(),
		SlotID:           assignment.SlotID().String(),
		MemberID:         assignment.MemberID().String(),
		AssignmentStatus: "confirmed",
		AssignmentMethod: "manual",
		AssignedAt:       assignment.AssignedAt().Format("2006-01-02T15:04:05Z07:00"),
		NotificationSent: false, // stub
	}

	writeSuccess(w, http.StatusCreated, resp)
}

// GetAssignments handles GET /api/v1/shift-assignments
func (h *ShiftAssignmentHandler) GetAssignments(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// テナントIDの取得
	tenantID, ok := getTenantIDFromContext(ctx)
	if !ok {
		writeError(w, http.StatusForbidden, "ERR_FORBIDDEN", "Tenant ID is required", nil)
		return
	}

	// クエリパラメータの取得
	memberIDStr := r.URL.Query().Get("member_id")
	slotIDStr := r.URL.Query().Get("slot_id")
	statusStr := r.URL.Query().Get("assignment_status")

	// フィルタ処理（簡易実装）
	// TODO: 日付範囲フィルタ（start_date, end_date）は v1.1 で実装

	var assignments []ShiftAssignmentResponse

	// member_id でフィルタ
	if memberIDStr != "" {
		memberID, err := common.ParseMemberID(memberIDStr)
		if err != nil {
			writeError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", "Invalid member_id format", nil)
			return
		}

		assignmentList, err := h.assignmentRepo.FindByMemberID(ctx, tenantID, memberID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "ERR_INTERNAL", "Failed to fetch assignments", nil)
			return
		}

		for _, a := range assignmentList {
			if statusStr == "" || (statusStr == "confirmed" && !a.IsCancelled()) || (statusStr == "cancelled" && a.IsCancelled()) {
				assignments = append(assignments, ShiftAssignmentResponse{
					AssignmentID:     a.AssignmentID().String(),
					SlotID:           a.SlotID().String(),
					MemberID:         a.MemberID().String(),
					AssignmentStatus: map[bool]string{true: "cancelled", false: "confirmed"}[a.IsCancelled()],
					AssignmentMethod: "manual",
					AssignedAt:       a.AssignedAt().Format("2006-01-02T15:04:05Z07:00"),
				})
			}
		}
	} else if slotIDStr != "" {
		// slot_id でフィルタ
		slotID, err := shift.ParseSlotID(slotIDStr)
		if err != nil {
			writeError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", "Invalid slot_id format", nil)
			return
		}

		assignmentList, err := h.assignmentRepo.FindBySlotID(ctx, tenantID, slotID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "ERR_INTERNAL", "Failed to fetch assignments", nil)
			return
		}

		for _, a := range assignmentList {
			if statusStr == "" || (statusStr == "confirmed" && !a.IsCancelled()) || (statusStr == "cancelled" && a.IsCancelled()) {
				assignments = append(assignments, ShiftAssignmentResponse{
					AssignmentID:     a.AssignmentID().String(),
					SlotID:           a.SlotID().String(),
					MemberID:         a.MemberID().String(),
					AssignmentStatus: map[bool]string{true: "cancelled", false: "confirmed"}[a.IsCancelled()],
					AssignmentMethod: "manual",
					AssignedAt:       a.AssignedAt().Format("2006-01-02T15:04:05Z07:00"),
				})
			}
		}
	} else {
		// フィルタなし（全件取得は非推奨だが、テスト用に許可）
		writeError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", "member_id or slot_id is required", nil)
		return
	}

	writeSuccess(w, http.StatusOK, map[string]interface{}{
		"assignments": assignments,
		"count":       len(assignments),
	})
}

// GetAssignmentDetail handles GET /api/v1/shift-assignments/:assignment_id
func (h *ShiftAssignmentHandler) GetAssignmentDetail(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// テナントIDの取得
	tenantID, ok := getTenantIDFromContext(ctx)
	if !ok {
		writeError(w, http.StatusForbidden, "ERR_FORBIDDEN", "Tenant ID is required", nil)
		return
	}

	// assignment_id の取得
	assignmentIDStr := chi.URLParam(r, "assignment_id")
	if assignmentIDStr == "" {
		writeError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", "assignment_id is required", nil)
		return
	}

	assignmentID, err := shift.ParseAssignmentID(assignmentIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", "Invalid assignment_id format", nil)
		return
	}

	// 割り当ての取得
	assignment, err := h.assignmentRepo.FindByID(ctx, tenantID, assignmentID)
	if err != nil {
		if err.Error() == "shift assignment not found" {
			writeError(w, http.StatusNotFound, "ERR_NOT_FOUND", "Shift assignment not found", nil)
			return
		}
		writeError(w, http.StatusInternalServerError, "ERR_INTERNAL", "Failed to fetch shift assignment", nil)
		return
	}

	// レスポンス
	resp := ShiftAssignmentResponse{
		AssignmentID:     assignment.AssignmentID().String(),
		SlotID:           assignment.SlotID().String(),
		MemberID:         assignment.MemberID().String(),
		AssignmentStatus: map[bool]string{true: "cancelled", false: "confirmed"}[assignment.IsCancelled()],
		AssignmentMethod: "manual",
		AssignedAt:       assignment.AssignedAt().Format("2006-01-02T15:04:05Z07:00"),
	}

	writeSuccess(w, http.StatusOK, resp)
}

