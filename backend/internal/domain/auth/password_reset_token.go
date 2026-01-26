package auth

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
)

// DefaultPasswordResetTokenExpiration はパスワードリセットトークンのデフォルト有効期限
const DefaultPasswordResetTokenExpiration = 1 * time.Hour

// PasswordResetToken はパスワードリセットトークンを表すエンティティ
type PasswordResetToken struct {
	tokenID   common.PasswordResetTokenID
	adminID   common.AdminID
	token     string // 64文字のhex文字列（32バイト）
	expiresAt time.Time
	usedAt    *time.Time
	createdAt time.Time
}

// NewPasswordResetToken は新しいパスワードリセットトークンを作成する
func NewPasswordResetToken(
	now time.Time,
	adminID common.AdminID,
	expirationDuration time.Duration,
) (*PasswordResetToken, error) {
	// セキュアなランダムトークン生成（32バイト = 64文字のhex）
	token, err := generateSecureToken(32)
	if err != nil {
		return nil, common.NewValidationError("failed to generate secure token", err)
	}

	prt := &PasswordResetToken{
		tokenID:   common.NewPasswordResetTokenIDWithTime(now),
		adminID:   adminID,
		token:     token,
		expiresAt: now.Add(expirationDuration),
		createdAt: now,
	}

	if err := prt.validate(); err != nil {
		return nil, err
	}

	return prt, nil
}

// ReconstructPasswordResetToken は永続化されたパスワードリセットトークンを再構築する
func ReconstructPasswordResetToken(
	tokenID common.PasswordResetTokenID,
	adminID common.AdminID,
	token string,
	expiresAt time.Time,
	usedAt *time.Time,
	createdAt time.Time,
) (*PasswordResetToken, error) {
	prt := &PasswordResetToken{
		tokenID:   tokenID,
		adminID:   adminID,
		token:     token,
		expiresAt: expiresAt,
		usedAt:    usedAt,
		createdAt: createdAt,
	}

	if err := prt.validate(); err != nil {
		return nil, err
	}

	return prt, nil
}

func (prt *PasswordResetToken) validate() error {
	if err := prt.tokenID.Validate(); err != nil {
		return err
	}
	if err := prt.adminID.Validate(); err != nil {
		return err
	}
	if prt.token == "" {
		return common.NewValidationError("token is required", nil)
	}
	if len(prt.token) != 64 {
		return common.NewValidationError("token must be 64 characters", nil)
	}
	if prt.expiresAt.Before(prt.createdAt) {
		return common.NewValidationError("expires_at must be after created_at", nil)
	}
	return nil
}

// IsExpired はトークンが期限切れかどうかを判定する
func (prt *PasswordResetToken) IsExpired(now time.Time) bool {
	return now.After(prt.expiresAt)
}

// IsUsed はトークンが使用済みかどうかを判定する
func (prt *PasswordResetToken) IsUsed() bool {
	return prt.usedAt != nil
}

// CanUse はトークンが使用可能かどうかを判定する
func (prt *PasswordResetToken) CanUse(now time.Time) error {
	if prt.IsUsed() {
		return common.NewValidationError("token already used", nil)
	}
	if prt.IsExpired(now) {
		return common.NewValidationError("token expired", nil)
	}
	return nil
}

// MarkAsUsed はトークンを使用済みにする
func (prt *PasswordResetToken) MarkAsUsed(now time.Time) error {
	if err := prt.CanUse(now); err != nil {
		return err
	}
	prt.usedAt = &now
	return nil
}

// Getters
func (prt *PasswordResetToken) TokenID() common.PasswordResetTokenID { return prt.tokenID }
func (prt *PasswordResetToken) AdminID() common.AdminID              { return prt.adminID }
func (prt *PasswordResetToken) Token() string                        { return prt.token }
func (prt *PasswordResetToken) ExpiresAt() time.Time                 { return prt.expiresAt }
func (prt *PasswordResetToken) UsedAt() *time.Time                   { return prt.usedAt }
func (prt *PasswordResetToken) CreatedAt() time.Time                 { return prt.createdAt }

// generateSecureRandomToken は暗号学的に安全なランダムトークンを生成する
// 注: invitation.go にも同様の関数があるが、パッケージ内で共有
func generateSecureRandomToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
