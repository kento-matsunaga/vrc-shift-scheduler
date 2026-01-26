package email

import (
	"context"
	"fmt"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/services"
	"github.com/resend/resend-go/v2"
)

// ResendEmailService is an implementation of EmailService using Resend
type ResendEmailService struct {
	client    *resend.Client
	fromEmail string
	baseURL   string
}

// NewResendEmailService creates a new ResendEmailService
func NewResendEmailService(apiKey, fromEmail, baseURL string) *ResendEmailService {
	return &ResendEmailService{
		client:    resend.NewClient(apiKey),
		fromEmail: fromEmail,
		baseURL:   baseURL,
	}
}

// SendInvitationEmail sends an invitation email via Resend
func (s *ResendEmailService) SendInvitationEmail(ctx context.Context, input services.SendInvitationEmailInput) error {
	invitationURL := s.baseURL + "/invitation/" + input.Token

	data := InvitationEmailData{
		InviterName:   input.InviterName,
		TenantName:    input.TenantName,
		Role:          input.Role,
		RoleJapanese:  RoleToJapanese(input.Role),
		ExpiresAt:     FormatExpiresAt(input.ExpiresAt),
		InvitationURL: invitationURL,
	}

	htmlBody, err := RenderInvitationHTML(data)
	if err != nil {
		return fmt.Errorf("failed to render HTML template: %w", err)
	}

	textBody, err := RenderInvitationText(data)
	if err != nil {
		return fmt.Errorf("failed to render text template: %w", err)
	}

	subject := "[VRC Shift Scheduler] 管理者として招待されました"

	params := &resend.SendEmailRequest{
		From:    s.fromEmail,
		To:      []string{input.To},
		Subject: subject,
		Html:    htmlBody,
		Text:    textBody,
	}

	_, err = s.client.Emails.Send(params)
	if err != nil {
		return fmt.Errorf("failed to send email via Resend: %w", err)
	}

	return nil
}

// Ensure ResendEmailService implements EmailService
var _ services.EmailService = (*ResendEmailService)(nil)
