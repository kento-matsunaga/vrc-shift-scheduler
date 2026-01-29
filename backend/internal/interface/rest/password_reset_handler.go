package rest

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	appAuth "github.com/erenoa/vrc-shift-scheduler/backend/internal/app/auth"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/auth"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/go-chi/chi/v5"
)

// PasswordResetHandler handles password reset related HTTP requests
type PasswordResetHandler struct {
	allowPasswordResetUsecase       *appAuth.AllowPasswordResetUsecase
	checkPasswordResetStatusUsecase *appAuth.CheckPasswordResetStatusUsecase
	verifyAndResetPasswordUsecase   *appAuth.VerifyAndResetPasswordUsecase
	requestPasswordResetUsecase     *appAuth.RequestPasswordResetUsecase
	resetPasswordWithTokenUsecase   *appAuth.ResetPasswordWithTokenUsecase
	rateLimiter                     *RateLimiter
}

// NewPasswordResetHandler creates a new PasswordResetHandler
func NewPasswordResetHandler(
	allowPasswordResetUsecase *appAuth.AllowPasswordResetUsecase,
	checkPasswordResetStatusUsecase *appAuth.CheckPasswordResetStatusUsecase,
	verifyAndResetPasswordUsecase *appAuth.VerifyAndResetPasswordUsecase,
	requestPasswordResetUsecase *appAuth.RequestPasswordResetUsecase,
	resetPasswordWithTokenUsecase *appAuth.ResetPasswordWithTokenUsecase,
	rateLimiter *RateLimiter,
) *PasswordResetHandler {
	return &PasswordResetHandler{
		allowPasswordResetUsecase:       allowPasswordResetUsecase,
		checkPasswordResetStatusUsecase: checkPasswordResetStatusUsecase,
		verifyAndResetPasswordUsecase:   verifyAndResetPasswordUsecase,
		requestPasswordResetUsecase:     requestPasswordResetUsecase,
		resetPasswordWithTokenUsecase:   resetPasswordWithTokenUsecase,
		rateLimiter:                     rateLimiter,
	}
}

// AllowPasswordResetRequest represents the request body for allowing password reset
type AllowPasswordResetRequest struct {
	// No body needed - target admin ID comes from URL path
}

// AllowPasswordResetResponse represents the response for allowing password reset
type AllowPasswordResetResponse struct {
	TargetAdminID string `json:"target_admin_id"`
	TargetEmail   string `json:"target_email"`
	AllowedAt     string `json:"allowed_at"`
	ExpiresAt     string `json:"expires_at"`
	AllowedByName string `json:"allowed_by_name"`
	Message       string `json:"message"`
}

// AllowPasswordReset handles POST /api/v1/admins/{admin_id}/allow-password-reset
// Allows an owner to permit a manager to reset their password
func (h *PasswordResetHandler) AllowPasswordReset(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get caller's tenant ID and admin ID from context (JWT)
	tenantID, ok := GetTenantID(ctx)
	if !ok {
		RespondBadRequest(w, "tenant_id is required")
		return
	}

	callerAdminID, ok := GetAdminID(ctx)
	if !ok {
		RespondBadRequest(w, "admin_id is required")
		return
	}

	roleStr, ok := GetRole(ctx)
	if !ok {
		RespondBadRequest(w, "role is required")
		return
	}

	// Parse caller's role
	callerRole, err := auth.NewRole(roleStr)
	if err != nil {
		RespondBadRequest(w, "invalid role")
		return
	}

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
	output, err := h.allowPasswordResetUsecase.Execute(ctx, appAuth.AllowPasswordResetInput{
		CallerAdminID: callerAdminID,
		CallerRole:    callerRole,
		TenantID:      tenantID,
		TargetAdminID: targetAdminID,
	})
	if err != nil {
		switch {
		case errors.Is(err, appAuth.ErrUnauthorized):
			RespondError(w, http.StatusForbidden, "ERR_FORBIDDEN", "この操作はオーナーのみ実行可能です", nil)
		case errors.Is(err, appAuth.ErrAdminNotFound):
			RespondError(w, http.StatusNotFound, "ERR_NOT_FOUND", "指定された管理者が見つかりません", nil)
		default:
			RespondDomainError(w, err)
		}
		return
	}

	RespondSuccess(w, AllowPasswordResetResponse{
		TargetAdminID: output.TargetAdminID,
		TargetEmail:   output.TargetEmail,
		AllowedAt:     output.AllowedAt,
		ExpiresAt:     output.ExpiresAt,
		AllowedByName: output.AllowedByName,
		Message:       "パスワードリセットを許可しました（24時間有効）",
	})
}

// CheckPasswordResetStatusResponse represents the response for checking password reset status
type CheckPasswordResetStatusResponse struct {
	Allowed   bool    `json:"allowed"`
	ExpiresAt *string `json:"expires_at,omitempty"`
	TenantID  string  `json:"tenant_id,omitempty"`
}

// CheckPasswordResetStatus handles GET /api/v1/auth/password-reset-status?email=xxx
// Public endpoint to check if password reset is allowed for an email
func (h *PasswordResetHandler) CheckPasswordResetStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Rate limiting check (if rate limiter is configured)
	if h.rateLimiter != nil {
		clientIP := getClientIP(r)
		if !h.rateLimiter.Allow(clientIP) {
			time.Sleep(1 * time.Second) // Delay to slow down attackers
			RespondError(w, http.StatusTooManyRequests, "ERR_RATE_LIMITED",
				"リクエストが多すぎます。しばらくしてから再度お試しください。", nil)
			return
		}
	}

	// Get email from query parameter
	email := r.URL.Query().Get("email")
	if email == "" {
		RespondBadRequest(w, "メールアドレスを入力してください")
		return
	}

	// Execute usecase
	output, err := h.checkPasswordResetStatusUsecase.Execute(ctx, appAuth.CheckPasswordResetStatusInput{
		Email: email,
	})
	if err != nil {
		// Don't reveal detailed errors to avoid information leakage
		RespondSuccess(w, CheckPasswordResetStatusResponse{
			Allowed: false,
		})
		return
	}

	RespondSuccess(w, CheckPasswordResetStatusResponse{
		Allowed:   output.Allowed,
		ExpiresAt: output.ExpiresAt,
		TenantID:  output.TenantID,
	})
}

// ResetPasswordRequest represents the request body for resetting password
type ResetPasswordRequest struct {
	Email              string `json:"email"`
	LicenseKey         string `json:"license_key"`
	NewPassword        string `json:"new_password"`
	ConfirmNewPassword string `json:"confirm_new_password"`
}

// ResetPasswordResponse represents the response for resetting password
type ResetPasswordResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// ResetPassword handles POST /api/v1/auth/reset-password
// Public endpoint to reset password with license key verification
func (h *PasswordResetHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Rate limiting check (if rate limiter is configured)
	if h.rateLimiter != nil {
		clientIP := getClientIP(r)
		if !h.rateLimiter.Allow(clientIP) {
			time.Sleep(1 * time.Second) // Delay to slow down attackers
			RespondError(w, http.StatusTooManyRequests, "ERR_RATE_LIMITED",
				"リクエストが多すぎます。しばらくしてから再度お試しください。", nil)
			return
		}
	}

	// Parse request body
	var req ResetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondBadRequest(w, "リクエストの形式が不正です")
		return
	}

	// Validation
	if req.Email == "" {
		RespondBadRequest(w, "メールアドレスを入力してください")
		return
	}
	if req.LicenseKey == "" {
		RespondBadRequest(w, "ライセンスキーを入力してください")
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

	// Execute usecase
	output, err := h.verifyAndResetPasswordUsecase.Execute(ctx, appAuth.VerifyAndResetPasswordInput{
		Email:       req.Email,
		LicenseKey:  req.LicenseKey,
		NewPassword: req.NewPassword,
	})
	if err != nil {
		switch {
		case errors.Is(err, appAuth.ErrAdminNotFound):
			RespondError(w, http.StatusNotFound, "ERR_NOT_FOUND", "指定されたメールアドレスの管理者が見つかりません", nil)
		case errors.Is(err, appAuth.ErrPasswordResetNotAllowed):
			RespondError(w, http.StatusForbidden, "ERR_FORBIDDEN", "パスワードリセットが許可されていないか、有効期限が切れています", nil)
		case errors.Is(err, appAuth.ErrInvalidLicenseKey):
			RespondError(w, http.StatusUnauthorized, "ERR_UNAUTHORIZED", "ライセンスキーが正しくありません", nil)
		default:
			RespondDomainError(w, err)
		}
		return
	}

	RespondSuccess(w, ResetPasswordResponse{
		Success: output.Success,
		Message: "パスワードをリセットしました",
	})
}

// ForgotPasswordRequest represents the request body for forgot password
type ForgotPasswordRequest struct {
	Email string `json:"email"`
}

// ForgotPasswordResponse represents the response for forgot password
type ForgotPasswordResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// ForgotPassword handles POST /api/v1/auth/forgot-password
// Public endpoint to request a password reset email
func (h *PasswordResetHandler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Rate limiting check (if rate limiter is configured)
	if h.rateLimiter != nil {
		clientIP := getClientIP(r)
		if !h.rateLimiter.Allow(clientIP) {
			time.Sleep(1 * time.Second) // Delay to slow down attackers
			RespondError(w, http.StatusTooManyRequests, "ERR_RATE_LIMITED",
				"リクエストが多すぎます。しばらくしてから再度お試しください。", nil)
			return
		}
	}

	// Parse request body
	var req ForgotPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondBadRequest(w, "リクエストの形式が不正です")
		return
	}

	// Execute usecase - always returns success to prevent user enumeration
	output, err := h.requestPasswordResetUsecase.Execute(ctx, appAuth.RequestPasswordResetInput{
		Email: req.Email,
	})
	if err != nil {
		// Should never happen as usecase always returns success
		RespondSuccess(w, ForgotPasswordResponse{
			Success: true,
			Message: "パスワードリセット用のメールを送信しました。メールをご確認ください。",
		})
		return
	}

	RespondSuccess(w, ForgotPasswordResponse{
		Success: output.Success,
		Message: output.Message,
	})
}

// ResetPasswordWithTokenRequest represents the request body for resetting password with token
type ResetPasswordWithTokenRequest struct {
	Token              string `json:"token"`
	NewPassword        string `json:"new_password"`
	ConfirmNewPassword string `json:"confirm_new_password"`
}

// ResetPasswordWithTokenResponse represents the response for resetting password with token
type ResetPasswordWithTokenResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// ResetPasswordWithToken handles POST /api/v1/auth/reset-password-with-token
// Public endpoint to reset password using a token received via email
func (h *PasswordResetHandler) ResetPasswordWithToken(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Rate limiting check (if rate limiter is configured)
	if h.rateLimiter != nil {
		clientIP := getClientIP(r)
		if !h.rateLimiter.Allow(clientIP) {
			time.Sleep(1 * time.Second) // Delay to slow down attackers
			RespondError(w, http.StatusTooManyRequests, "ERR_RATE_LIMITED",
				"リクエストが多すぎます。しばらくしてから再度お試しください。", nil)
			return
		}
	}

	// Parse request body
	var req ResetPasswordWithTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondBadRequest(w, "リクエストの形式が不正です")
		return
	}

	// Validation
	if req.Token == "" {
		RespondBadRequest(w, "トークンが必要です")
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
	if req.NewPassword != req.ConfirmNewPassword {
		RespondBadRequest(w, "新しいパスワードと確認用パスワードが一致しません")
		return
	}

	// Execute usecase
	output, err := h.resetPasswordWithTokenUsecase.Execute(ctx, appAuth.ResetPasswordWithTokenInput{
		Token:       req.Token,
		NewPassword: req.NewPassword,
	})
	if err != nil {
		RespondDomainError(w, err)
		return
	}

	RespondSuccess(w, ResetPasswordWithTokenResponse{
		Success: output.Success,
		Message: output.Message,
	})
}
