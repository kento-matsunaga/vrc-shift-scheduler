package auth

import (
	"context"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
)

// PasswordResetTokenRepository はパスワードリセットトークンの永続化を担当する
type PasswordResetTokenRepository interface {
	// Save はパスワードリセットトークンを保存する（INSERT or UPDATE）
	Save(ctx context.Context, token *PasswordResetToken) error

	// FindByToken はトークン文字列でパスワードリセットトークンを検索する
	// 未使用のトークンのみを返す
	FindByToken(ctx context.Context, token string) (*PasswordResetToken, error)

	// FindValidByAdminID は指定管理者の有効な（未使用かつ未期限切れの）トークンを検索する
	FindValidByAdminID(ctx context.Context, adminID common.AdminID) (*PasswordResetToken, error)

	// InvalidateAllByAdminID は指定管理者の全てのトークンを無効化する
	// 新しいトークン発行時に古いトークンを無効化するために使用
	InvalidateAllByAdminID(ctx context.Context, adminID common.AdminID) error
}
