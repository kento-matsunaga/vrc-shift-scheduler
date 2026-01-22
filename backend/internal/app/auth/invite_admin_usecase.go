package auth

import (
	"context"
	"log/slog"
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/auth"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/services"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/tenant"
)

// InviteAdminInput represents the input for inviting an admin
type InviteAdminInput struct {
	InviterAdminID string // JWTから取得
	Email          string
	Role           string
}

// InviteAdminOutput represents the output for inviting an admin
type InviteAdminOutput struct {
	InvitationID string    `json:"invitation_id"`
	Email        string    `json:"email"`
	Role         string    `json:"role"`
	Token        string    `json:"token"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// InviteAdminUsecase handles the admin invitation use case
type InviteAdminUsecase struct {
	adminRepo      auth.AdminRepository
	invitationRepo auth.InvitationRepository
	tenantRepo     tenant.TenantRepository
	emailService   services.EmailService
	clock          services.Clock
}

// NewInviteAdminUsecase creates a new InviteAdminUsecase
func NewInviteAdminUsecase(
	adminRepo auth.AdminRepository,
	invitationRepo auth.InvitationRepository,
	tenantRepo tenant.TenantRepository,
	emailService services.EmailService,
	clk services.Clock,
) *InviteAdminUsecase {
	return &InviteAdminUsecase{
		adminRepo:      adminRepo,
		invitationRepo: invitationRepo,
		tenantRepo:     tenantRepo,
		emailService:   emailService,
		clock:          clk,
	}
}

// Execute executes the invite admin use case
func (u *InviteAdminUsecase) Execute(ctx context.Context, input InviteAdminInput) (*InviteAdminOutput, error) {
	now := u.clock.Now()

	// 1. 招待者のAdmin取得
	inviterAdminID, err := common.ParseAdminID(input.InviterAdminID)
	if err != nil {
		return nil, common.NewValidationError("invalid inviter_admin_id", err)
	}

	inviterAdmin, err := u.adminRepo.FindByID(ctx, inviterAdminID)
	if err != nil {
		return nil, err
	}

	// 2. Role検証
	role, err := auth.NewRole(input.Role)
	if err != nil {
		return nil, err
	}

	// 3. 既に同じメールアドレスの管理者が存在するかチェック
	existsAdmin, _ := u.adminRepo.FindByEmailGlobal(ctx, input.Email)
	if existsAdmin != nil {
		return nil, common.NewValidationError("admin with this email already exists", nil)
	}

	// 4. 既に同じメールアドレスの未受理招待が存在するかチェック
	existsPending, err := u.invitationRepo.ExistsPendingByEmail(ctx, inviterAdmin.TenantID(), input.Email)
	if err != nil {
		return nil, err
	}
	if existsPending {
		return nil, common.NewValidationError("pending invitation for this email already exists", nil)
	}

	// 5. 招待作成（7日間有効）
	invitation, err := auth.NewInvitation(
		now,
		inviterAdmin, // Admin集約を渡す（tenantIDが自動設定される）
		input.Email,
		role,
		7*24*time.Hour, // 7日間
	)
	if err != nil {
		return nil, err
	}

	// 6. 招待を保存
	if err := u.invitationRepo.Save(ctx, invitation); err != nil {
		return nil, err
	}

	// 7. テナント情報を取得してメール送信
	tenantEntity, err := u.tenantRepo.FindByID(ctx, inviterAdmin.TenantID())
	if err != nil {
		return nil, err
	}

	emailInput := services.SendInvitationEmailInput{
		To:          invitation.Email(),
		InviterName: inviterAdmin.DisplayName(),
		TenantName:  tenantEntity.TenantName(),
		Role:        invitation.Role().String(),
		Token:       invitation.Token(),
		ExpiresAt:   invitation.ExpiresAt(),
	}

	if err := u.emailService.SendInvitationEmail(ctx, emailInput); err != nil {
		// メール送信失敗時は招待をロールバック
		if deleteErr := u.invitationRepo.Delete(ctx, invitation.InvitationID()); deleteErr != nil {
			slog.Error("failed to rollback invitation after email failure",
				"invitation_id", invitation.InvitationID().String(),
				"email_error", err,
				"delete_error", deleteErr)
		}
		return nil, common.NewDomainError("ERR_EMAIL_SEND_FAILED", "failed to send invitation email: "+err.Error())
	}

	return &InviteAdminOutput{
		InvitationID: invitation.InvitationID().String(),
		Email:        invitation.Email(),
		Role:         invitation.Role().String(),
		Token:        invitation.Token(),
		ExpiresAt:    invitation.ExpiresAt(),
	}, nil
}
