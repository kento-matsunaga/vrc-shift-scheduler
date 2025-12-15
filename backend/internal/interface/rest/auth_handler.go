package rest

import (
	"encoding/json"
	"errors"
	"net/http"

	appAuth "github.com/erenoa/vrc-shift-scheduler/backend/internal/app/auth"
)

// AuthHandler handles authentication-related HTTP requests
type AuthHandler struct {
	loginUsecase *appAuth.LoginUsecase
}

// NewAuthHandler creates a new AuthHandler
func NewAuthHandler(loginUsecase *appAuth.LoginUsecase) *AuthHandler {
	return &AuthHandler{
		loginUsecase: loginUsecase,
	}
}

// LoginRequest represents the request body for login
type LoginRequest struct {
	// TenantID削除: email + password のみ
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginResponse represents the response body for login
type LoginResponse struct {
	Token     string `json:"token"`
	AdminID   string `json:"admin_id"`
	TenantID  string `json:"tenant_id"`
	Role      string `json:"role"`
	ExpiresAt string `json:"expires_at"`
}

// Login handles POST /api/v1/auth/login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	// 1. リクエスト解析
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", "Invalid request body", nil)
		return
	}

	// バリデーション
	if req.Email == "" {
		RespondError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", "email is required", nil)
		return
	}
	if req.Password == "" {
		RespondError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", "password is required", nil)
		return
	}

	// 2. Usecase呼び出し（ビジネスロジックはここにない）
	output, err := h.loginUsecase.Execute(r.Context(), appAuth.LoginInput{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		// エラーコード変換
		switch {
		case errors.Is(err, appAuth.ErrInvalidCredentials):
			RespondError(w, http.StatusUnauthorized, "ERR_UNAUTHORIZED", "Invalid email or password", nil)
		case errors.Is(err, appAuth.ErrAccountDisabled):
			RespondError(w, http.StatusForbidden, "ERR_FORBIDDEN", "Account is disabled", nil)
		default:
			RespondInternalError(w)
		}
		return
	}

	// 3. レスポンス変換
	RespondJSON(w, http.StatusOK, SuccessResponse{
		Data: LoginResponse{
			Token:     output.Token,
			AdminID:   output.AdminID,
			TenantID:  output.TenantID,
			Role:      output.Role,
			ExpiresAt: output.ExpiresAt.Format("2006-01-02T15:04:05Z07:00"),
		},
	})
}
