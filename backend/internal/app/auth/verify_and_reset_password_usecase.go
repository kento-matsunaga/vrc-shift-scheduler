package auth

import (
	"context"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/auth"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/billing"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/services"
)

// VerifyAndResetPasswordInput represents the input for password reset with license key verification
type VerifyAndResetPasswordInput struct {
	Email       string // PWリセット対象のメールアドレス
	LicenseKey  string // 本人確認用のライセンスキー（平文）
	NewPassword string // 新しいパスワード（平文）
}

// VerifyAndResetPasswordOutput represents the output for password reset
type VerifyAndResetPasswordOutput struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// VerifyAndResetPasswordUsecase handles the password reset with license key verification use case
type VerifyAndResetPasswordUsecase struct {
	adminRepo      auth.AdminRepository
	licenseKeyRepo billing.LicenseKeyRepository
	passwordHasher services.PasswordHasher
	clock          services.Clock
}

// NewVerifyAndResetPasswordUsecase creates a new VerifyAndResetPasswordUsecase
func NewVerifyAndResetPasswordUsecase(
	adminRepo auth.AdminRepository,
	licenseKeyRepo billing.LicenseKeyRepository,
	passwordHasher services.PasswordHasher,
	clock services.Clock,
) *VerifyAndResetPasswordUsecase {
	return &VerifyAndResetPasswordUsecase{
		adminRepo:      adminRepo,
		licenseKeyRepo: licenseKeyRepo,
		passwordHasher: passwordHasher,
		clock:          clock,
	}
}

// Execute executes the password reset with license key verification use case
func (u *VerifyAndResetPasswordUsecase) Execute(ctx context.Context, input VerifyAndResetPasswordInput) (*VerifyAndResetPasswordOutput, error) {
	// 1. メールアドレスで管理者を検索
	admin, err := u.adminRepo.FindByEmailGlobal(ctx, input.Email)
	if err != nil {
		return nil, ErrAdminNotFound
	}
	if admin == nil {
		return nil, ErrAdminNotFound
	}

	// 2. PWリセットが許可されているか確認（24時間以内）
	now := u.clock.Now()
	if !admin.CanResetPassword(now) {
		return nil, ErrPasswordResetNotAllowed
	}

	// 3. ライセンスキーを正規化してハッシュ化
	normalizedKey := billing.NormalizeLicenseKey(input.LicenseKey)
	keyHash := billing.HashLicenseKey(normalizedKey)

	// 4. ハッシュとテナントIDで使用済みライセンスキーを検索
	licenseKey, err := u.licenseKeyRepo.FindByHashAndTenant(ctx, keyHash, admin.TenantID())
	if err != nil {
		return nil, ErrInvalidLicenseKey
	}
	if licenseKey == nil {
		// ライセンスキーが見つからない（テナントと一致しない）
		return nil, ErrInvalidLicenseKey
	}

	// 5. ライセンスキーが使用済みであることを確認（used_tenant_idが一致）
	if !licenseKey.IsUsed() || licenseKey.UsedTenantID() == nil {
		return nil, ErrInvalidLicenseKey
	}
	if *licenseKey.UsedTenantID() != admin.TenantID() {
		return nil, ErrInvalidLicenseKey
	}

	// 6. 新しいパスワードをハッシュ化
	newPasswordHash, err := u.passwordHasher.Hash(input.NewPassword)
	if err != nil {
		return nil, err
	}

	// 7. パスワードをリセット（許可もクリア）
	if err := admin.ResetPassword(now, newPasswordHash); err != nil {
		return nil, err
	}

	// 8. 保存
	if err := u.adminRepo.Save(ctx, admin); err != nil {
		return nil, err
	}

	return &VerifyAndResetPasswordOutput{
		Success: true,
		Message: "password reset successfully",
	}, nil
}
