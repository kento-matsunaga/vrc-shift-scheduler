package rest

import (
	"errors"
	"log"
	"net/http"

	appAuth "github.com/erenoa/vrc-shift-scheduler/backend/internal/app/auth"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/go-chi/chi/v5"
)

// AdminAuthHandler handles admin auth-related HTTP requests
type AdminAuthHandler struct {
	adminAllowPasswordResetUsecase *appAuth.AdminAllowPasswordResetUsecase
}

// NewAdminAuthHandler creates a new AdminAuthHandler
func NewAdminAuthHandler(
	adminAllowPasswordResetUsecase *appAuth.AdminAllowPasswordResetUsecase,
) *AdminAuthHandler {
	return &AdminAuthHandler{
		adminAllowPasswordResetUsecase: adminAllowPasswordResetUsecase,
	}
}

// AllowPasswordResetResponse represents the response for allowing password reset
type AdminAllowPasswordResetResponse struct {
	TargetAdminID string `json:"target_admin_id"`
	TargetEmail   string `json:"target_email"`
	TenantID      string `json:"tenant_id"`
	AllowedAt     string `json:"allowed_at"`
	ExpiresAt     string `json:"expires_at"`
	Message       string `json:"message"`
}

// AllowPasswordReset handles POST /api/v1/admin/admins/{admin_id}/allow-password-reset
// Allows system admin to permit password reset for any tenant admin
func (h *AdminAuthHandler) AllowPasswordReset(w http.ResponseWriter, r *http.Request) {
	// Get system admin ID from context
	systemAdminID := getAdminIDFromContext(r)

	// Get target admin ID from URL path
	targetAdminIDStr := chi.URLParam(r, "admin_id")
	if targetAdminIDStr == "" {
		RespondBadRequest(w, "admin_id is required in path")
		return
	}

	targetAdminID, err := common.ParseAdminID(targetAdminIDStr)
	if err != nil {
		RespondBadRequest(w, "invalid admin_id format")
		return
	}

	// Execute usecase
	output, err := h.adminAllowPasswordResetUsecase.Execute(r.Context(), appAuth.AdminAllowPasswordResetInput{
		TargetAdminID: targetAdminID,
		SystemAdminID: systemAdminID,
	})
	if err != nil {
		log.Printf("[ERROR] AdminAllowPasswordReset failed: %v", err)
		if errors.Is(err, appAuth.ErrAdminNotFound) {
			RespondError(w, http.StatusNotFound, "ERR_NOT_FOUND", "指定された管理者が見つかりません", nil)
			return
		}
		RespondDomainError(w, err)
		return
	}

	RespondSuccess(w, AdminAllowPasswordResetResponse{
		TargetAdminID: output.TargetAdminID,
		TargetEmail:   output.TargetEmail,
		TenantID:      output.TenantID,
		AllowedAt:     output.AllowedAt,
		ExpiresAt:     output.ExpiresAt,
		Message:       "パスワードリセットを許可しました（24時間有効）",
	})
}
