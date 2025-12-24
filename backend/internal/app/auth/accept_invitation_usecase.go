package auth

import (
	"context"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/auth"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/services"
)

// AcceptInvitationInput represents the input for accepting an invitation
type AcceptInvitationInput struct {
	Token       string
	DisplayName string
	Password    string
}

// AcceptInvitationOutput represents the output for accepting an invitation
type AcceptInvitationOutput struct {
	AdminID  string `json:"admin_id"`
	TenantID string `json:"tenant_id"`
	Email    string `json:"email"`
	Role     string `json:"role"`
}

// AcceptInvitationUsecase handles the accept invitation use case
type AcceptInvitationUsecase struct {
	adminRepo      auth.AdminRepository
	invitationRepo auth.InvitationRepository
	passwordHasher services.PasswordHasher
	clock          services.Clock
}

// NewAcceptInvitationUsecase creates a new AcceptInvitationUsecase
func NewAcceptInvitationUsecase(
	adminRepo auth.AdminRepository,
	invitationRepo auth.InvitationRepository,
	passwordHasher services.PasswordHasher,
	clk services.Clock,
) *AcceptInvitationUsecase {
	return &AcceptInvitationUsecase{
		adminRepo:      adminRepo,
		invitationRepo: invitationRepo,
		passwordHasher: passwordHasher,
		clock:          clk,
	}
}

// Execute executes the accept invitation use case
func (u *AcceptInvitationUsecase) Execute(ctx context.Context, input AcceptInvitationInput) (*AcceptInvitationOutput, error) {
	now := u.clock.Now()

	// 1. 招待を取得
	invitation, err := u.invitationRepo.FindByToken(ctx, input.Token)
	if err != nil {
		return nil, ErrInvalidInvitation
	}

	// 2. 招待を受理できるかチェック（ドメインルール）
	if err := invitation.CanAccept(now); err != nil {
		return nil, ErrInvalidInvitation
	}

	// 3. 既に同じメールアドレスの管理者が存在するかチェック
	existsAdmin, _ := u.adminRepo.FindByEmailGlobal(ctx, invitation.Email())
	if existsAdmin != nil {
		return nil, ErrEmailAlreadyExists
	}

	// 4. パスワードハッシュ化
	passwordHash, err := u.passwordHasher.Hash(input.Password)
	if err != nil {
		return nil, err
	}

	// 5. Admin作成
	admin, err := auth.NewAdmin(
		now,
		invitation.TenantID(),
		invitation.Email(),
		passwordHash,
		input.DisplayName,
		invitation.Role(),
	)
	if err != nil {
		return nil, err
	}

	// 6. Adminを保存
	if err := u.adminRepo.Save(ctx, admin); err != nil {
		return nil, err
	}

	// 7. 招待を受理済みに更新
	if err := invitation.Accept(now); err != nil {
		return nil, err
	}

	if err := u.invitationRepo.Save(ctx, invitation); err != nil {
		return nil, err
	}

	return &AcceptInvitationOutput{
		AdminID:  admin.AdminID().String(),
		TenantID: admin.TenantID().String(),
		Email:    admin.Email(),
		Role:     admin.Role().String(),
	}, nil
}
