package auth

import (
	"context"
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/auth"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/infra/security"
)

// ChangePasswordInput represents the input for changing password
type ChangePasswordInput struct {
	AdminID         common.AdminID
	TenantID        common.TenantID
	CurrentPassword string
	NewPassword     string
}

// ChangePasswordUsecase handles the password change use case
type ChangePasswordUsecase struct {
	adminRepo      auth.AdminRepository
	passwordHasher security.PasswordHasher
}

// NewChangePasswordUsecase creates a new ChangePasswordUsecase
func NewChangePasswordUsecase(
	adminRepo auth.AdminRepository,
	passwordHasher security.PasswordHasher,
) *ChangePasswordUsecase {
	return &ChangePasswordUsecase{
		adminRepo:      adminRepo,
		passwordHasher: passwordHasher,
	}
}

// Execute executes the password change use case
func (u *ChangePasswordUsecase) Execute(ctx context.Context, input ChangePasswordInput) error {
	// 1. Admin取得
	admin, err := u.adminRepo.FindByIDWithTenant(ctx, input.TenantID, input.AdminID)
	if err != nil {
		return err
	}

	// 2. 現在のパスワードを検証
	if err := u.passwordHasher.Compare(admin.PasswordHash(), input.CurrentPassword); err != nil {
		return ErrInvalidCredentials
	}

	// 3. 新しいパスワードをハッシュ化
	newPasswordHash, err := u.passwordHasher.Hash(input.NewPassword)
	if err != nil {
		return common.NewValidationError("failed to hash password", err)
	}

	// 4. パスワードを更新
	if err := admin.UpdatePasswordHash(time.Now(), newPasswordHash); err != nil {
		return err
	}

	// 5. 保存
	if err := u.adminRepo.Save(ctx, admin); err != nil {
		return err
	}

	return nil
}
