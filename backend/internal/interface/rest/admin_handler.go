package rest

import (
	"encoding/json"
	"errors"
	"net/http"

	appAuth "github.com/erenoa/vrc-shift-scheduler/backend/internal/app/auth"
)

// AdminHandler handles admin-related HTTP requests
type AdminHandler struct {
	changePasswordUsecase *appAuth.ChangePasswordUsecase
}

// NewAdminHandler creates a new AdminHandler with injected usecases
func NewAdminHandler(
	changePasswordUC *appAuth.ChangePasswordUsecase,
) *AdminHandler {
	return &AdminHandler{
		changePasswordUsecase: changePasswordUC,
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
		RespondBadRequest(w, "現在のパスワードを入力してください")
		return
	}
	if req.NewPassword == "" {
		RespondBadRequest(w, "新しいパスワードを入力してください")
		return
	}
	if req.ConfirmNewPassword == "" {
		RespondBadRequest(w, "確認用パスワードを入力してください")
		return
	}
	if len(req.NewPassword) < 8 {
		RespondBadRequest(w, "新しいパスワードは8文字以上で入力してください")
		return
	}
	if req.NewPassword != req.ConfirmNewPassword {
		RespondBadRequest(w, "新しいパスワードと確認用パスワードが一致しません")
		return
	}
	if req.CurrentPassword == req.NewPassword {
		RespondBadRequest(w, "新しいパスワードは現在のパスワードと異なる必要があります")
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
			RespondError(w, http.StatusUnauthorized, "ERR_UNAUTHORIZED", "現在のパスワードが正しくありません", nil)
			return
		}
		RespondDomainError(w, err)
		return
	}

	// 成功レスポンス
	RespondSuccess(w, map[string]string{
		"message": "パスワードを変更しました",
	})
}
