package auth

import (
	"context"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/auth"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/infra/security"
)

// LoginUsecase handles the login use case
type LoginUsecase struct {
	adminRepo      auth.AdminRepository
	passwordHasher security.PasswordHasher
	tokenIssuer    security.TokenIssuer
}

// NewLoginUsecase creates a new LoginUsecase
func NewLoginUsecase(
	adminRepo auth.AdminRepository,
	passwordHasher security.PasswordHasher,
	tokenIssuer security.TokenIssuer,
) *LoginUsecase {
	return &LoginUsecase{
		adminRepo:      adminRepo,
		passwordHasher: passwordHasher,
		tokenIssuer:    tokenIssuer,
	}
}

// Execute executes the login use case
func (u *LoginUsecase) Execute(ctx context.Context, input LoginInput) (*LoginOutput, error) {
	// 1. Admin取得（グローバル検索）
	admin, err := u.adminRepo.FindByEmailGlobal(ctx, input.Email)
	if err != nil {
		// メールアドレスが存在しない場合も ErrInvalidCredentials を返す（攻撃者にヒントを与えない）
		return nil, ErrInvalidCredentials
	}

	// 2. ログイン可能かチェック（ドメインルール）
	if !admin.CanLogin() {
		return nil, ErrAccountDisabled
	}

	// 3. パスワード検証（Infra層に委譲）
	if err := u.passwordHasher.Compare(admin.PasswordHash(), input.Password); err != nil {
		// パスワードが違う場合も ErrInvalidCredentials を返す（攻撃者にヒントを与えない）
		return nil, ErrInvalidCredentials
	}

	// 4. JWT発行（Infra層に委譲）
	// TenantIDはAdminから自動取得
	token, expiresAt, err := u.tokenIssuer.Issue(
		admin.AdminID().String(),
		admin.TenantID().String(),
		admin.Role().String(),
	)
	if err != nil {
		return nil, err
	}

	return &LoginOutput{
		Token:     token,
		AdminID:   admin.AdminID().String(),
		TenantID:  admin.TenantID().String(), // TenantIDは返す（フロント用）
		Role:      admin.Role().String(),
		ExpiresAt: expiresAt,
	}, nil
}
