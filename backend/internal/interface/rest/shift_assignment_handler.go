package rest

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

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
	dbPool            *pgxpool.Pool
}

// NewShiftAssignmentHandler creates a new ShiftAssignmentHandler
func NewShiftAssignmentHandler(dbPool *pgxpool.Pool) *ShiftAssignmentHandler {
	return &ShiftAssignmentHandler{
		assignmentService: app.NewShiftAssignmentService(dbPool),
		assignmentRepo:    db.NewShiftAssignmentRepository(dbPool),
		dbPool:            dbPool,
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
	AssignmentID        string  `json:"assignment_id"`
	TenantID            string  `json:"tenant_id"`
	SlotID              string  `json:"slot_id"`
	MemberID            string  `json:"member_id"`
	MemberDisplayName   string  `json:"member_display_name,omitempty"`
	SlotName            string  `json:"slot_name,omitempty"`
	TargetDate          string  `json:"target_date,omitempty"`
	StartTime           string  `json:"start_time,omitempty"`
	EndTime             string  `json:"end_time,omitempty"`
	AssignmentStatus    string  `json:"assignment_status"`
	AssignmentMethod    string  `json:"assignment_method"`
	IsOutsidePreference bool    `json:"is_outside_preference"`
	AssignedAt          string  `json:"assigned_at"`
	CancelledAt         *string `json:"cancelled_at,omitempty"`
	CreatedAt           string  `json:"created_at"`
	UpdatedAt           string  `json:"updated_at"`
	NotificationSent    bool    `json:"notification_sent"`
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
	// Admin (JWT認証) または Member (X-Member-ID認証) のいずれかを取得
	var actorID common.MemberID
	if adminID, ok := GetAdminIDFromContext(ctx); ok {
		// Admin の場合は admin_id を MemberID として使用
		actorID = common.MemberID(adminID)
	} else if memberID, ok := getMemberIDFromContext(ctx); ok {
		// Member の場合は member_id を使用
		actorID = memberID
	} else {
		writeError(w, http.StatusForbidden, "ERR_FORBIDDEN", "Member ID or Admin ID is required", nil)
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
		log.Printf("ConfirmAssignment error: %+v", err)
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

	// レスポンス（JOIN データを含む）
	resp, err := h.buildAssignmentResponse(ctx, assignment, nil, nil)
	if err != nil {
		// JOIN エラーは無視して最小限のレスポンスを返す
		resp = &ShiftAssignmentResponse{
			AssignmentID:        assignment.AssignmentID().String(),
			TenantID:            assignment.TenantID().String(),
			SlotID:              assignment.SlotID().String(),
			MemberID:            assignment.MemberID().String(),
			AssignmentStatus:    "confirmed",
			AssignmentMethod:    "manual",
			IsOutsidePreference: assignment.IsOutsidePreference(),
			AssignedAt:          assignment.AssignedAt().Format(time.RFC3339),
			CreatedAt:           assignment.CreatedAt().Format(time.RFC3339),
			UpdatedAt:           assignment.UpdatedAt().Format(time.RFC3339),
			NotificationSent:    false, // stub
		}
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
	startDateStr := r.URL.Query().Get("start_date")
	endDateStr := r.URL.Query().Get("end_date")

	// 日付範囲のパース
	var startDate, endDate *time.Time
	if startDateStr != "" {
		parsed, err := time.Parse("2006-01-02", startDateStr)
		if err != nil {
			writeError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", "Invalid start_date format (expected YYYY-MM-DD)", nil)
			return
		}
		startDate = &parsed
	}
	if endDateStr != "" {
		parsed, err := time.Parse("2006-01-02", endDateStr)
		if err != nil {
			writeError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", "Invalid end_date format (expected YYYY-MM-DD)", nil)
			return
		}
		endDate = &parsed
	}

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
				resp, err := h.buildAssignmentResponse(ctx, a, startDate, endDate)
				if err != nil {
					writeError(w, http.StatusInternalServerError, "ERR_INTERNAL", "Failed to build assignment response", nil)
					return
				}
				if resp != nil {
					assignments = append(assignments, *resp)
				}
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
				resp, err := h.buildAssignmentResponse(ctx, a, startDate, endDate)
				if err != nil {
					writeError(w, http.StatusInternalServerError, "ERR_INTERNAL", "Failed to build assignment response", nil)
					return
				}
				if resp != nil {
					assignments = append(assignments, *resp)
				}
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
	resp, err := h.buildAssignmentResponse(ctx, assignment, nil, nil)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "ERR_INTERNAL", "Failed to build assignment response", nil)
		return
	}
	if resp == nil {
		writeError(w, http.StatusInternalServerError, "ERR_INTERNAL", "Failed to build assignment response", nil)
		return
	}

	writeSuccess(w, http.StatusOK, resp)
}

// buildAssignmentResponse builds a ShiftAssignmentResponse with JOIN data
func (h *ShiftAssignmentHandler) buildAssignmentResponse(ctx context.Context, assignment *shift.ShiftAssignment, startDate, endDate *time.Time) (*ShiftAssignmentResponse, error) {
	// JOIN クエリで member_display_name, slot_name, target_date, start_time, end_time を取得
	query := `
		SELECT
			m.display_name,
			ss.slot_name,
			ebd.target_date,
			ss.start_time,
			ss.end_time
		FROM shift_assignments sa
		INNER JOIN members m ON sa.member_id = m.member_id AND m.deleted_at IS NULL
		INNER JOIN shift_slots ss ON sa.slot_id = ss.slot_id AND ss.deleted_at IS NULL
		INNER JOIN event_business_days ebd ON ss.business_day_id = ebd.business_day_id AND ebd.deleted_at IS NULL
		WHERE sa.assignment_id = $1 AND sa.tenant_id = $2 AND sa.deleted_at IS NULL
	`

	var (
		memberDisplayName string
		slotName          string
		targetDate        time.Time
		startTime         time.Time
		endTime           time.Time
	)

	err := h.dbPool.QueryRow(ctx, query, assignment.AssignmentID().String(), assignment.TenantID().String()).Scan(
		&memberDisplayName,
		&slotName,
		&targetDate,
		&startTime,
		&endTime,
	)
	if err != nil {
		return nil, err
	}

	// 日付範囲フィルタ
	if startDate != nil && targetDate.Before(*startDate) {
		return nil, nil // フィルタで除外
	}
	if endDate != nil && targetDate.After(*endDate) {
		return nil, nil // フィルタで除外
	}

	var cancelledAtStr *string
	if assignment.CancelledAt() != nil {
		s := assignment.CancelledAt().Format(time.RFC3339)
		cancelledAtStr = &s
	}

	return &ShiftAssignmentResponse{
		AssignmentID:        assignment.AssignmentID().String(),
		TenantID:            assignment.TenantID().String(),
		SlotID:              assignment.SlotID().String(),
		MemberID:            assignment.MemberID().String(),
		MemberDisplayName:   memberDisplayName,
		SlotName:            slotName,
		TargetDate:          targetDate.Format("2006-01-02"),
		StartTime:           startTime.Format("15:04:05"),
		EndTime:             endTime.Format("15:04:05"),
		AssignmentStatus:    map[bool]string{true: "cancelled", false: "confirmed"}[assignment.IsCancelled()],
		AssignmentMethod:    string(assignment.AssignmentMethod()),
		IsOutsidePreference: assignment.IsOutsidePreference(),
		AssignedAt:          assignment.AssignedAt().Format(time.RFC3339),
		CancelledAt:         cancelledAtStr,
		CreatedAt:           assignment.CreatedAt().Format(time.RFC3339),
		UpdatedAt:           assignment.UpdatedAt().Format(time.RFC3339),
		NotificationSent:    false, // stub
	}, nil
}

// CancelAssignment handles DELETE /api/v1/shift-assignments/:assignment_id
func (h *ShiftAssignmentHandler) CancelAssignment(w http.ResponseWriter, r *http.Request) {
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

	// 割り当ての削除（論理削除）
	if err := h.assignmentRepo.Delete(ctx, tenantID, assignmentID); err != nil {
		if err.Error() == "shift assignment not found" {
			writeError(w, http.StatusNotFound, "ERR_NOT_FOUND", "Shift assignment not found", nil)
			return
		}
		writeError(w, http.StatusInternalServerError, "ERR_INTERNAL", "Failed to delete shift assignment", nil)
		return
	}

	writeSuccess(w, http.StatusNoContent, nil)
}
