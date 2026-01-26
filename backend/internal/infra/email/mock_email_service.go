package email

import (
	"context"
	"log/slog"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/services"
)

// MockEmailService is a mock implementation of EmailService for development
// It logs email content instead of actually sending
type MockEmailService struct {
	baseURL string
}

// NewMockEmailService creates a new MockEmailService
func NewMockEmailService(baseURL string) *MockEmailService {
	return &MockEmailService{
		baseURL: baseURL,
	}
}

// SendInvitationEmail logs the invitation email content
func (s *MockEmailService) SendInvitationEmail(ctx context.Context, input services.SendInvitationEmailInput) error {
	invitationURL := s.baseURL + "/invite/" + input.Token

	slog.Info("=== Mock Email Service: Invitation Email ===",
		"to", input.To,
		"inviter", input.InviterName,
		"tenant", input.TenantName,
		"role", input.Role,
		"expires_at", input.ExpiresAt.Format("2006-01-02 15:04"),
		"invitation_url", invitationURL,
	)

	slog.Info("Mock email content",
		"subject", "[VRC Shift Scheduler] 管理者として招待されました",
		"body_preview", "招待者: "+input.InviterName+" / テナント: "+input.TenantName,
	)

	return nil
}

// SendPasswordResetEmail logs the password reset email content
func (s *MockEmailService) SendPasswordResetEmail(ctx context.Context, input services.SendPasswordResetEmailInput) error {
	resetURL := s.baseURL + "/reset-password/" + input.Token

	slog.Info("=== Mock Email Service: Password Reset Email ===",
		"to", input.To,
		"expires_at", input.ExpiresAt.Format("2006-01-02 15:04"),
		"reset_url", resetURL,
	)

	slog.Info("Mock email content",
		"subject", "[VRC Shift Scheduler] パスワードリセット",
		"body_preview", "パスワードリセットのリクエストを受け付けました",
	)

	return nil
}

// Ensure MockEmailService implements EmailService
var _ services.EmailService = (*MockEmailService)(nil)
