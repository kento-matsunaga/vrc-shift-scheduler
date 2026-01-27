package services

import (
	"context"
	"time"
)

// SendInvitationEmailInput represents the input for sending an invitation email
type SendInvitationEmailInput struct {
	To          string    // 招待先メールアドレス
	InviterName string    // 招待者の表示名
	TenantName  string    // テナント名
	Role        string    // 招待されたロール
	Token       string    // 招待トークン
	ExpiresAt   time.Time // 有効期限
}

// SendPasswordResetEmailInput represents the input for sending a password reset email
type SendPasswordResetEmailInput struct {
	To        string    // 送信先メールアドレス
	Token     string    // パスワードリセットトークン
	ExpiresAt time.Time // 有効期限
}

// EmailService defines the interface for sending emails
type EmailService interface {
	// SendInvitationEmail sends an invitation email to a new admin
	SendInvitationEmail(ctx context.Context, input SendInvitationEmailInput) error

	// SendPasswordResetEmail sends a password reset email
	SendPasswordResetEmail(ctx context.Context, input SendPasswordResetEmailInput) error
}
