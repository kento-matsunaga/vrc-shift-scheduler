package rest

import (
	"net/http"
	"strconv"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/app"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ActualAttendanceHandler handles actual attendance-related HTTP requests
type ActualAttendanceHandler struct {
	service *app.ActualAttendanceService
}

// NewActualAttendanceHandler creates a new ActualAttendanceHandler
func NewActualAttendanceHandler(dbPool *pgxpool.Pool) *ActualAttendanceHandler {
	return &ActualAttendanceHandler{
		service: app.NewActualAttendanceService(dbPool),
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
	limit := 30 // デフォルト
	if limitStr != "" {
		parsed, err := strconv.Atoi(limitStr)
		if err != nil || parsed <= 0 {
			writeError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", "Invalid limit parameter", nil)
			return
		}
		limit = parsed
	}

	// 本出席データの取得
	data, err := h.service.GetRecentActualAttendance(ctx, tenantID, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "ERR_INTERNAL", "Failed to fetch actual attendance data", nil)
		return
	}

	writeSuccess(w, http.StatusOK, data)
}
