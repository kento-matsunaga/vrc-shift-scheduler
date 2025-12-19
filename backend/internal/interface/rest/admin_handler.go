package rest

import (
	"encoding/json"
	"errors"
	"net/http"

	appAuth "github.com/erenoa/vrc-shift-scheduler/backend/internal/app/auth"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/infra/db"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/infra/security"
	"github.com/jackc/pgx/v5/pgxpool"
)

// AdminHandler handles admin-related HTTP requests
type AdminHandler struct {
	changePasswordUsecase *appAuth.ChangePasswordUsecase
}

// NewAdminHandler creates a new AdminHandler
func NewAdminHandler(dbPool *pgxpool.Pool) *AdminHandler {
	adminRepo := db.NewAdminRepository(dbPool)
	passwordHasher := security.NewBcryptHasher()
	return &AdminHandler{
		changePasswordUsecase: appAuth.NewChangePasswordUsecase(adminRepo, passwordHasher),
	}
}

// ChangePasswordRequest represents the request body for changing password
type ChangePasswordRequest struct {
	CurrentPassword    string `json:"current_password"`
	NewPassword        string `json:"new_password"`
	ConfirmNewPassword string `json:"confirm_new_password"`
}

// ChangePassword handles POST /api/v1/admins/me/change-password
func (h *AdminHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// テナントIDとAdminIDの取得
	tenantID, ok := GetTenantID(ctx)
	if !ok {
		RespondBadRequest(w, "tenant_id is required")
		return
	}

	adminID, ok := GetAdminID(ctx)
	if !ok {
		RespondBadRequest(w, "admin_id is required")
		return
	}

	// リクエストボディのパース
	var req ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondBadRequest(w, "Invalid request body")
		return
	}

	// バリデーション
	if req.CurrentPassword == "" {
		RespondBadRequest(w, "current_password is required")
		return
	}
	if req.NewPassword == "" {
		RespondBadRequest(w, "new_password is required")
		return
	}
	if req.ConfirmNewPassword == "" {
		RespondBadRequest(w, "confirm_new_password is required")
		return
	}
	if len(req.NewPassword) < 8 {
		RespondBadRequest(w, "new_password must be at least 8 characters")
		return
	}
	if req.NewPassword != req.ConfirmNewPassword {
		RespondBadRequest(w, "new_password and confirm_new_password do not match")
		return
	}
	if req.CurrentPassword == req.NewPassword {
		RespondBadRequest(w, "new_password must be different from current_password")
		return
	}

	// Usecaseの実行
	input := appAuth.ChangePasswordInput{
		AdminID:         adminID,
		TenantID:        tenantID,
		CurrentPassword: req.CurrentPassword,
		NewPassword:     req.NewPassword,
	}

	if err := h.changePasswordUsecase.Execute(ctx, input); err != nil {
		// エラーハンドリング
		if errors.Is(err, appAuth.ErrInvalidCredentials) {
			RespondError(w, http.StatusUnauthorized, "ERR_UNAUTHORIZED", "Current password is incorrect", nil)
			return
		}
		RespondDomainError(w, err)
		return
	}

	// 成功レスポンス
	RespondSuccess(w, map[string]string{
		"message": "Password changed successfully",
	})
}
