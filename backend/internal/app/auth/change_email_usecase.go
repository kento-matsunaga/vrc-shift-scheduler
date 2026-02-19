package auth

import (
	"context"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/auth"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/services"
)

// ChangeEmailInput represents the input for changing email
type ChangeEmailInput struct {
	AdminID         common.AdminID
	TenantID        common.TenantID
	CurrentPassword string
	NewEmail        string
}

// ChangeEmailUsecase handles the email change use case
type ChangeEmailUsecase struct {
	adminRepo      auth.AdminRepository
	passwordHasher services.PasswordHasher
	clock          services.Clock
}

// NewChangeEmailUsecase creates a new ChangeEmailUsecase
func NewChangeEmailUsecase(
	adminRepo auth.AdminRepository,
	passwordHasher services.PasswordHasher,
	clock services.Clock,
) *ChangeEmailUsecase {
	return &ChangeEmailUsecase{
		adminRepo:      adminRepo,
		passwordHasher: passwordHasher,
		clock:          clock,
	}
}

// Execute executes the email change use case
func (u *ChangeEmailUsecase) Execute(ctx context.Context, input ChangeEmailInput) error {
	// 1. Admin取得
	admin, err := u.adminRepo.FindByIDWithTenant(ctx, input.TenantID, input.AdminID)
	if err != nil {
		return err
	}

	// 2. 現在のパスワードを検証
	if err := u.passwordHasher.Compare(admin.PasswordHash(), input.CurrentPassword); err != nil {
		return ErrInvalidCredentials
	}

	// 3. 現在のメールアドレスと同じでないことを確認
	if admin.Email() == input.NewEmail {
		return common.NewValidationError("new email must be different from current email", nil)
	}

	// 4. グローバルで重複チェック
	exists, err := u.adminRepo.ExistsByEmailGlobal(ctx, input.NewEmail)
	if err != nil {
		return err
	}
	if exists {
		return ErrEmailAlreadyExists
	}

	// 5. メールアドレスを更新（Domain層でフォーマット検証も行う）
	if err := admin.UpdateEmail(u.clock.Now(), input.NewEmail); err != nil {
		return err
	}

	// 6. 保存
	if err := u.adminRepo.Save(ctx, admin); err != nil {
		return err
	}

	return nil
}
