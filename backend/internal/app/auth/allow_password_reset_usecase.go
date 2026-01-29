package auth

import (
	"context"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/auth"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/services"
)

// AllowPasswordResetInput represents the input for allowing password reset
type AllowPasswordResetInput struct {
	CallerAdminID common.AdminID  // 実行者（Owner）のID
	CallerRole    auth.Role       // 実行者のロール
	TenantID      common.TenantID // テナントID
	TargetAdminID common.AdminID  // PWリセットを許可する対象のID
}

// AllowPasswordResetOutput represents the output for allowing password reset
type AllowPasswordResetOutput struct {
	TargetAdminID string `json:"target_admin_id"`
	TargetEmail   string `json:"target_email"`
	AllowedAt     string `json:"allowed_at"`
	ExpiresAt     string `json:"expires_at"`
	AllowedByName string `json:"allowed_by_name"`
}

// AllowPasswordResetUsecase handles the password reset allowance use case
type AllowPasswordResetUsecase struct {
	adminRepo auth.AdminRepository
	clock     services.Clock
}

// NewAllowPasswordResetUsecase creates a new AllowPasswordResetUsecase
func NewAllowPasswordResetUsecase(
	adminRepo auth.AdminRepository,
	clock services.Clock,
) *AllowPasswordResetUsecase {
	return &AllowPasswordResetUsecase{
		adminRepo: adminRepo,
		clock:     clock,
	}
}

// Execute executes the password reset allowance use case
func (u *AllowPasswordResetUsecase) Execute(ctx context.Context, input AllowPasswordResetInput) (*AllowPasswordResetOutput, error) {
	// 1. 実行者がOwnerであることを確認
	if input.CallerRole != auth.RoleOwner {
		return nil, ErrUnauthorized
	}

	// 2. 対象の管理者を取得
	targetAdmin, err := u.adminRepo.FindByIDWithTenant(ctx, input.TenantID, input.TargetAdminID)
	if err != nil {
		return nil, err
	}
	if targetAdmin == nil {
		return nil, ErrAdminNotFound
	}

	// 3. 実行者の情報を取得（ログ/通知用）
	callerAdmin, err := u.adminRepo.FindByIDWithTenant(ctx, input.TenantID, input.CallerAdminID)
	if err != nil {
		return nil, err
	}
	if callerAdmin == nil {
		return nil, ErrAdminNotFound
	}

	// 4. PWリセットを許可（ドメインルールで自分自身への許可は禁止）
	now := u.clock.Now()
	if err := targetAdmin.AllowPasswordReset(now, input.CallerAdminID); err != nil {
		return nil, err
	}

	// 5. 保存
	if err := u.adminRepo.Save(ctx, targetAdmin); err != nil {
		return nil, err
	}

	// 6. 出力を構築
	expiresAt := targetAdmin.PasswordResetExpiresAt()
	var expiresAtStr string
	if expiresAt != nil {
		expiresAtStr = expiresAt.Format("2006-01-02T15:04:05Z07:00")
	}

	return &AllowPasswordResetOutput{
		TargetAdminID: targetAdmin.AdminID().String(),
		TargetEmail:   targetAdmin.Email(),
		AllowedAt:     now.Format("2006-01-02T15:04:05Z07:00"),
		ExpiresAt:     expiresAtStr,
		AllowedByName: callerAdmin.DisplayName(),
	}, nil
}
