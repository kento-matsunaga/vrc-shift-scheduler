package email

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/aws/aws-sdk-go-v2/service/ses/types"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/services"
)

// SESEmailService is an implementation of EmailService using AWS SES
type SESEmailService struct {
	client    *ses.Client
	fromEmail string
	baseURL   string
}

// NewSESEmailService creates a new SESEmailService
func NewSESEmailService(cfg aws.Config, fromEmail, baseURL string) *SESEmailService {
	return &SESEmailService{
		client:    ses.NewFromConfig(cfg),
		fromEmail: fromEmail,
		baseURL:   baseURL,
	}
}

// SendInvitationEmail sends an invitation email via AWS SES
func (s *SESEmailService) SendInvitationEmail(ctx context.Context, input services.SendInvitationEmailInput) error {
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

	sendInput := &ses.SendEmailInput{
		Source: aws.String(s.fromEmail),
		Destination: &types.Destination{
			ToAddresses: []string{input.To},
		},
		Message: &types.Message{
			Subject: &types.Content{
				Charset: aws.String("UTF-8"),
				Data:    aws.String(subject),
			},
			Body: &types.Body{
				Html: &types.Content{
					Charset: aws.String("UTF-8"),
					Data:    aws.String(htmlBody),
				},
				Text: &types.Content{
					Charset: aws.String("UTF-8"),
					Data:    aws.String(textBody),
				},
			},
		},
	}

	_, err = s.client.SendEmail(ctx, sendInput)
	if err != nil {
		return fmt.Errorf("failed to send email via SES: %w", err)
	}

	return nil
}

// Ensure SESEmailService implements EmailService
var _ services.EmailService = (*SESEmailService)(nil)
