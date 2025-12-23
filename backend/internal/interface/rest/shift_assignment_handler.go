package rest

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	appshift "github.com/erenoa/vrc-shift-scheduler/backend/internal/app/shift"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/shift"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/infra/db"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ShiftAssignmentHandler handles shift assignment-related HTTP requests
type ShiftAssignmentHandler struct {
	confirmAssignmentUC     *appshift.ConfirmManualAssignmentUsecase
	getAssignmentsUC        *appshift.GetAssignmentsUsecase
	getAssignmentDetailUC   *appshift.GetAssignmentDetailUsecase
	cancelAssignmentUC      *appshift.CancelAssignmentUsecase
}

// NewShiftAssignmentHandler creates a new ShiftAssignmentHandler
func NewShiftAssignmentHandler(dbPool *pgxpool.Pool) *ShiftAssignmentHandler {
	slotRepo := db.NewShiftSlotRepository(dbPool)
	assignmentRepo := db.NewShiftAssignmentRepository(dbPool)
	memberRepo := db.NewMemberRepository(dbPool)
	businessDayRepo := db.NewEventBusinessDayRepository(dbPool)

	return &ShiftAssignmentHandler{
		confirmAssignmentUC:   appshift.NewConfirmManualAssignmentUsecase(slotRepo, assignmentRepo, memberRepo),
		getAssignmentsUC:      appshift.NewGetAssignmentsUsecase(assignmentRepo, memberRepo, slotRepo, businessDayRepo),
		getAssignmentDetailUC: appshift.NewGetAssignmentDetailUsecase(assignmentRepo, memberRepo, slotRepo, businessDayRepo),
		cancelAssignmentUC:    appshift.NewCancelAssignmentUsecase(assignmentRepo),
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

	// Parse IDs
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

	// Execute usecase
	input := appshift.ConfirmManualAssignmentInput{
		TenantID: tenantID,
		SlotID:   slotID,
		MemberID: memberID,
		ActorID:  actorID,
		Note:     req.Note,
	}

	assignment, err := h.confirmAssignmentUC.Execute(ctx, input)
	if err != nil {
		log.Printf("ConfirmAssignment error: %+v", err)
		// Handle domain errors
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

	// Get assignment details with JOIN data
	detailInput := appshift.GetAssignmentDetailInput{
		TenantID:     tenantID,
		AssignmentID: assignment.AssignmentID(),
	}

	details, err := h.getAssignmentDetailUC.Execute(ctx, detailInput)
	if err != nil {
		// If JOIN fails, return minimal response
		resp := &ShiftAssignmentResponse{
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
			NotificationSent:    false,
		}
		writeSuccess(w, http.StatusCreated, resp)
		return
	}

	// Build full response with JOIN data
	resp := buildAssignmentResponse(details)
	writeSuccess(w, http.StatusCreated, resp)
}

// GetAssignments handles GET /api/v1/shift-assignments
func (h *ShiftAssignmentHandler) GetAssignments(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get tenant ID
	tenantID, ok := getTenantIDFromContext(ctx)
	if !ok {
		writeError(w, http.StatusForbidden, "ERR_FORBIDDEN", "Tenant ID is required", nil)
		return
	}

	// Parse query parameters
	memberIDStr := r.URL.Query().Get("member_id")
	slotIDStr := r.URL.Query().Get("slot_id")
	statusStr := r.URL.Query().Get("assignment_status")
	startDateStr := r.URL.Query().Get("start_date")
	endDateStr := r.URL.Query().Get("end_date")

	// Parse date range
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

	// Build usecase input
	input := appshift.GetAssignmentsInput{
		TenantID:  tenantID,
		Status:    statusStr,
		StartDate: startDate,
		EndDate:   endDate,
	}

	if memberIDStr != "" {
		memberID, err := common.ParseMemberID(memberIDStr)
		if err != nil {
			writeError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", "Invalid member_id format", nil)
			return
		}
		input.MemberID = &memberID
	} else if slotIDStr != "" {
		slotID, err := shift.ParseSlotID(slotIDStr)
		if err != nil {
			writeError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", "Invalid slot_id format", nil)
			return
		}
		input.SlotID = &slotID
	} else {
		writeError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", "member_id or slot_id is required", nil)
		return
	}

	// Execute usecase
	assignmentDetails, err := h.getAssignmentsUC.Execute(ctx, input)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "ERR_INTERNAL", "Failed to fetch assignments", nil)
		return
	}

	// Build response
	assignments := make([]ShiftAssignmentResponse, 0, len(assignmentDetails))
	for _, details := range assignmentDetails {
		assignments = append(assignments, buildAssignmentResponse(details))
	}

	writeSuccess(w, http.StatusOK, map[string]interface{}{
		"assignments": assignments,
		"count":       len(assignments),
	})
}

// GetAssignmentDetail handles GET /api/v1/shift-assignments/:assignment_id
func (h *ShiftAssignmentHandler) GetAssignmentDetail(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get tenant ID
	tenantID, ok := getTenantIDFromContext(ctx)
	if !ok {
		writeError(w, http.StatusForbidden, "ERR_FORBIDDEN", "Tenant ID is required", nil)
		return
	}

	// Get assignment_id from URL
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

	// Execute usecase
	input := appshift.GetAssignmentDetailInput{
		TenantID:     tenantID,
		AssignmentID: assignmentID,
	}

	details, err := h.getAssignmentDetailUC.Execute(ctx, input)
	if err != nil {
		if err.Error() == "shift assignment not found" {
			writeError(w, http.StatusNotFound, "ERR_NOT_FOUND", "Shift assignment not found", nil)
			return
		}
		writeError(w, http.StatusInternalServerError, "ERR_INTERNAL", "Failed to fetch shift assignment", nil)
		return
	}

	// Build response
	resp := buildAssignmentResponse(details)
	writeSuccess(w, http.StatusOK, resp)
}

// buildAssignmentResponse builds a ShiftAssignmentResponse from AssignmentWithDetails
func buildAssignmentResponse(details *appshift.AssignmentWithDetails) ShiftAssignmentResponse {
	var cancelledAtStr *string
	if details.Assignment.CancelledAt() != nil {
		s := details.Assignment.CancelledAt().Format(time.RFC3339)
		cancelledAtStr = &s
	}

	return ShiftAssignmentResponse{
		AssignmentID:        details.Assignment.AssignmentID().String(),
		TenantID:            details.Assignment.TenantID().String(),
		SlotID:              details.Assignment.SlotID().String(),
		MemberID:            details.Assignment.MemberID().String(),
		MemberDisplayName:   details.MemberDisplayName,
		SlotName:            details.SlotName,
		TargetDate:          details.TargetDate.Format("2006-01-02"),
		StartTime:           details.StartTime.Format("15:04:05"),
		EndTime:             details.EndTime.Format("15:04:05"),
		AssignmentStatus:    map[bool]string{true: "cancelled", false: "confirmed"}[details.Assignment.IsCancelled()],
		AssignmentMethod:    string(details.Assignment.AssignmentMethod()),
		IsOutsidePreference: details.Assignment.IsOutsidePreference(),
		AssignedAt:          details.Assignment.AssignedAt().Format(time.RFC3339),
		CancelledAt:         cancelledAtStr,
		CreatedAt:           details.Assignment.CreatedAt().Format(time.RFC3339),
		UpdatedAt:           details.Assignment.UpdatedAt().Format(time.RFC3339),
		NotificationSent:    false, // stub
	}
}

// CancelAssignment handles DELETE /api/v1/shift-assignments/:assignment_id
func (h *ShiftAssignmentHandler) CancelAssignment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get tenant ID
	tenantID, ok := getTenantIDFromContext(ctx)
	if !ok {
		writeError(w, http.StatusForbidden, "ERR_FORBIDDEN", "Tenant ID is required", nil)
		return
	}

	// Get assignment_id from URL
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

	// Execute usecase
	input := appshift.CancelAssignmentInput{
		TenantID:     tenantID,
		AssignmentID: assignmentID,
	}

	err = h.cancelAssignmentUC.Execute(ctx, input)
	if err != nil {
		if err.Error() == "shift assignment not found" {
			writeError(w, http.StatusNotFound, "ERR_NOT_FOUND", "Shift assignment not found", nil)
			return
		}
		writeError(w, http.StatusInternalServerError, "ERR_INTERNAL", "Failed to delete shift assignment", nil)
		return
	}

	writeSuccess(w, http.StatusNoContent, nil)
}
