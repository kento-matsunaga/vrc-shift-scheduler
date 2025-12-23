package rest

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	appAuth "github.com/erenoa/vrc-shift-scheduler/backend/internal/app/auth"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/infra/clock"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/infra/db"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/infra/security"
)

// InvitationHandler handles invitation-related HTTP requests
type InvitationHandler struct {
	inviteAdminUsecase      *appAuth.InviteAdminUsecase
	acceptInvitationUsecase *appAuth.AcceptInvitationUsecase
}

// NewInvitationHandler creates a new InvitationHandler
func NewInvitationHandler(pool *pgxpool.Pool) *InvitationHandler {
	adminRepo := db.NewAdminRepository(pool)
	invitationRepo := db.NewInvitationRepository(pool)
	systemClock := &clock.RealClock{}
	passwordHasher := security.NewBcryptHasher()

	return &InvitationHandler{
		inviteAdminUsecase:      appAuth.NewInviteAdminUsecase(adminRepo, invitationRepo, systemClock),
		acceptInvitationUsecase: appAuth.NewAcceptInvitationUsecase(adminRepo, invitationRepo, passwordHasher, systemClock),
	}
}

// InviteAdminRequest represents the request body for inviting an admin
type InviteAdminRequest struct {
	Email string `json:"email"`
	Role  string `json:"role"`
}

// InviteAdminResponse represents the response body for inviting an admin
type InviteAdminResponse struct {
	InvitationID string `json:"invitation_id"`
	Email        string `json:"email"`
	Role         string `json:"role"`
	Token        string `json:"token"`
	ExpiresAt    string `json:"expires_at"`
}

// InviteAdmin handles POST /api/v1/invitations
func (h *InvitationHandler) InviteAdmin(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// JWTからadmin_idを取得
	adminID, ok := GetAdminID(ctx)
	if !ok {
		RespondError(w, http.StatusUnauthorized, "ERR_UNAUTHORIZED", "admin_id is required", nil)
		return
	}

	var req InviteAdminRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", "Invalid request body", nil)
		return
	}

	if req.Email == "" {
		RespondError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", "メールアドレスを入力してください", nil)
		return
	}
	if req.Role == "" {
		RespondError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", "権限を選択してください", nil)
		return
	}

	output, err := h.inviteAdminUsecase.Execute(ctx, appAuth.InviteAdminInput{
		InviterAdminID: adminID.String(),
		Email:          req.Email,
		Role:           req.Role,
	})
	if err != nil {
		RespondDomainError(w, err)
		return
	}

	RespondJSON(w, http.StatusCreated, SuccessResponse{
		Data: InviteAdminResponse{
			InvitationID: output.InvitationID,
			Email:        output.Email,
			Role:         output.Role,
			Token:        output.Token,
			ExpiresAt:    output.ExpiresAt.Format("2006-01-02T15:04:05Z07:00"),
		},
	})
}

// AcceptInvitationRequest represents the request body for accepting an invitation
type AcceptInvitationRequest struct {
	DisplayName string `json:"display_name"`
	Password    string `json:"password"`
}

// AcceptInvitationResponse represents the response body for accepting an invitation
type AcceptInvitationResponse struct {
	AdminID  string `json:"admin_id"`
	TenantID string `json:"tenant_id"`
	Email    string `json:"email"`
	Role     string `json:"role"`
}

// AcceptInvitation handles POST /api/v1/invitations/accept/{token}
func (h *InvitationHandler) AcceptInvitation(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	if token == "" {
		RespondError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", "token is required", nil)
		return
	}

	var req AcceptInvitationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", "Invalid request body", nil)
		return
	}

	if req.DisplayName == "" {
		RespondError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", "表示名を入力してください", nil)
		return
	}
	if req.Password == "" {
		RespondError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", "パスワードを入力してください", nil)
		return
	}

	output, err := h.acceptInvitationUsecase.Execute(r.Context(), appAuth.AcceptInvitationInput{
		Token:       token,
		DisplayName: req.DisplayName,
		Password:    req.Password,
	})
	if err != nil {
		switch {
		case errors.Is(err, appAuth.ErrInvalidInvitation):
			RespondError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", "招待が無効または期限切れです", nil)
		case errors.Is(err, appAuth.ErrEmailAlreadyExists):
			RespondError(w, http.StatusConflict, "ERR_CONFLICT", "このメールアドレスは既に登録されています", nil)
		default:
			RespondDomainError(w, err)
		}
		return
	}

	RespondJSON(w, http.StatusCreated, SuccessResponse{
		Data: AcceptInvitationResponse{
			AdminID:  output.AdminID,
			TenantID: output.TenantID,
			Email:    output.Email,
			Role:     output.Role,
		},
	})
}
