package auth

import (
	"context"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/auth"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/services"
)

// CheckPasswordResetStatusInput represents the input for checking password reset status
type CheckPasswordResetStatusInput struct {
	Email string // PWリセット対象のメールアドレス
}

// CheckPasswordResetStatusOutput represents the output for checking password reset status
type CheckPasswordResetStatusOutput struct {
	Allowed   bool    `json:"allowed"`              // PWリセットが許可されているか
	ExpiresAt *string `json:"expires_at,omitempty"` // 有効期限（許可されている場合）
	TenantID  string  `json:"tenant_id,omitempty"`  // テナントID（許可されている場合）
}

// CheckPasswordResetStatusUsecase handles the password reset status check use case
type CheckPasswordResetStatusUsecase struct {
	adminRepo auth.AdminRepository
	clock     services.Clock
}

// NewCheckPasswordResetStatusUsecase creates a new CheckPasswordResetStatusUsecase
func NewCheckPasswordResetStatusUsecase(
	adminRepo auth.AdminRepository,
	clock services.Clock,
) *CheckPasswordResetStatusUsecase {
	return &CheckPasswordResetStatusUsecase{
		adminRepo: adminRepo,
		clock:     clock,
	}
}

// Execute executes the password reset status check use case
func (u *CheckPasswordResetStatusUsecase) Execute(ctx context.Context, input CheckPasswordResetStatusInput) (*CheckPasswordResetStatusOutput, error) {
	// 1. メールアドレスで管理者を検索（グローバル検索）
	admin, err := u.adminRepo.FindByEmailGlobal(ctx, input.Email)
	if err != nil {
		// メールアドレスが存在しない場合は未許可として返す（攻撃者にヒントを与えない）
		return &CheckPasswordResetStatusOutput{
			Allowed: false,
		}, nil
	}
	if admin == nil {
		return &CheckPasswordResetStatusOutput{
			Allowed: false,
		}, nil
	}

	// 2. PWリセット可能かチェック（24時間以内かどうか）
	now := u.clock.Now()
	if !admin.CanResetPassword(now) {
		return &CheckPasswordResetStatusOutput{
			Allowed: false,
		}, nil
	}

	// 3. 有効期限を取得
	expiresAt := admin.PasswordResetExpiresAt()
	var expiresAtStr *string
	if expiresAt != nil {
		s := expiresAt.Format("2006-01-02T15:04:05Z07:00")
		expiresAtStr = &s
	}

	return &CheckPasswordResetStatusOutput{
		Allowed:   true,
		ExpiresAt: expiresAtStr,
		TenantID:  admin.TenantID().String(),
	}, nil
}
