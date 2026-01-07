package rest

import (
	"net/http"
	"strconv"

	appactual "github.com/erenoa/vrc-shift-scheduler/backend/internal/app/actual_attendance"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/event"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/member"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/shift"
)

// ActualAttendanceHandler handles actual attendance-related HTTP requests
type ActualAttendanceHandler struct {
	getRecentActualAttendanceUC *appactual.GetRecentActualAttendanceUsecase
}

// NewActualAttendanceHandler creates a new ActualAttendanceHandler
func NewActualAttendanceHandler(
	businessDayRepo event.EventBusinessDayRepository,
	memberRepo member.MemberRepository,
	assignmentRepo shift.ShiftAssignmentRepository,
) *ActualAttendanceHandler {
	return &ActualAttendanceHandler{
		getRecentActualAttendanceUC: appactual.NewGetRecentActualAttendanceUsecase(
			businessDayRepo,
			memberRepo,
			assignmentRepo,
		),
	}
}

// GetRecentActualAttendance handles GET /api/v1/actual-attendance
// 本出席データ（実際のシフト割り当て実績）を取得
func (h *ActualAttendanceHandler) GetRecentActualAttendance(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// テナントIDの取得
	tenantID, ok := getTenantIDFromContext(ctx)
	if !ok {
		writeError(w, http.StatusForbidden, "ERR_FORBIDDEN", "Tenant ID is required", nil)
		return
	}

	// クエリパラメータの取得
	limitStr := r.URL.Query().Get("limit")
	limit := 10 // デフォルト
	if limitStr != "" {
		parsed, err := strconv.Atoi(limitStr)
		if err != nil || parsed <= 0 {
			writeError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", "Invalid limit parameter", nil)
			return
		}
		limit = parsed
	}

	// event_id クエリパラメータの取得（オプション）
	var eventID *common.EventID
	eventIDStr := r.URL.Query().Get("event_id")
	if eventIDStr != "" {
		eid := common.EventID(eventIDStr)
		eventID = &eid
	}

	// Execute usecase
	input := appactual.GetRecentActualAttendanceInput{
		TenantID: tenantID,
		EventID:  eventID,
		Limit:    limit,
	}

	output, err := h.getRecentActualAttendanceUC.Execute(ctx, input)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "ERR_INTERNAL", "Failed to fetch actual attendance data", nil)
		return
	}

	// Build response
	response := map[string]interface{}{
		"target_dates":       output.TargetDates,
		"member_attendances": output.MemberAttendances,
	}

	writeSuccess(w, http.StatusOK, response)
}
