package auth

import (
	"context"
	"log/slog"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/auth"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/services"
)

// RequestPasswordResetInput represents the input for requesting password reset
type RequestPasswordResetInput struct {
	Email string
}

// RequestPasswordResetOutput represents the output for requesting password reset
// Note: Always returns success to prevent user enumeration attacks
type RequestPasswordResetOutput struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// RequestPasswordResetUsecase handles the password reset request use case
type RequestPasswordResetUsecase struct {
	adminRepo              auth.AdminRepository
	passwordResetTokenRepo auth.PasswordResetTokenRepository
	emailService           services.EmailService
	clock                  services.Clock
}

// NewRequestPasswordResetUsecase creates a new RequestPasswordResetUsecase
func NewRequestPasswordResetUsecase(
	adminRepo auth.AdminRepository,
	passwordResetTokenRepo auth.PasswordResetTokenRepository,
	emailService services.EmailService,
	clock services.Clock,
) *RequestPasswordResetUsecase {
	return &RequestPasswordResetUsecase{
		adminRepo:              adminRepo,
		passwordResetTokenRepo: passwordResetTokenRepo,
		emailService:           emailService,
		clock:                  clock,
	}
}

// Execute executes the password reset request use case
// IMPORTANT: Always returns success to prevent user enumeration attacks
func (u *RequestPasswordResetUsecase) Execute(ctx context.Context, input RequestPasswordResetInput) (*RequestPasswordResetOutput, error) {
	now := u.clock.Now()

	// Validate email format
	if input.Email == "" {
		return &RequestPasswordResetOutput{
			Success: true,
			Message: "パスワードリセット用のメールを送信しました。メールをご確認ください。",
		}, nil
	}
	if !isValidEmail(input.Email) {
		return &RequestPasswordResetOutput{
			Success: true,
			Message: "パスワードリセット用のメールを送信しました。メールをご確認ください。",
		}, nil
	}

	// Find admin by email (global search)
	admin, err := u.adminRepo.FindByEmailGlobal(ctx, input.Email)
	if err != nil || admin == nil {
		// Admin not found - return success to prevent user enumeration
		slog.Info("Password reset requested for non-existent email")
		return &RequestPasswordResetOutput{
			Success: true,
			Message: "パスワードリセット用のメールを送信しました。メールをご確認ください。",
		}, nil
	}

	// Check if admin is active and not deleted
	if !admin.IsActive() || admin.IsDeleted() {
		slog.Info("Password reset requested for inactive/deleted admin",
			"admin_id", admin.AdminID().String(),
			"is_active", admin.IsActive(),
			"is_deleted", admin.IsDeleted())
		return &RequestPasswordResetOutput{
			Success: true,
			Message: "パスワードリセット用のメールを送信しました。メールをご確認ください。",
		}, nil
	}

	// Invalidate any existing tokens for this admin
	if err := u.passwordResetTokenRepo.InvalidateAllByAdminID(ctx, admin.AdminID()); err != nil {
		slog.Error("Failed to invalidate existing password reset tokens",
			"admin_id", admin.AdminID().String(),
			"error", err)
		// Continue anyway - not critical
	}

	// Create new password reset token (1 hour expiration)
	token, err := auth.NewPasswordResetToken(
		now,
		admin.AdminID(),
		auth.DefaultPasswordResetTokenExpiration,
	)
	if err != nil {
		slog.Error("Failed to create password reset token",
			"admin_id", admin.AdminID().String(),
			"error", err)
		return &RequestPasswordResetOutput{
			Success: true,
			Message: "パスワードリセット用のメールを送信しました。メールをご確認ください。",
		}, nil
	}

	// Save the token
	if err := u.passwordResetTokenRepo.Save(ctx, token); err != nil {
		slog.Error("Failed to save password reset token",
			"admin_id", admin.AdminID().String(),
			"error", err)
		return &RequestPasswordResetOutput{
			Success: true,
			Message: "パスワードリセット用のメールを送信しました。メールをご確認ください。",
		}, nil
	}

	// Send password reset email
	emailInput := services.SendPasswordResetEmailInput{
		To:        admin.Email(),
		Token:     token.Token(),
		ExpiresAt: token.ExpiresAt(),
	}

	if err := u.emailService.SendPasswordResetEmail(ctx, emailInput); err != nil {
		slog.Error("Failed to send password reset email",
			"admin_id", admin.AdminID().String(),
			"error", err)
		// Return success even if email fails to prevent timing attacks
		return &RequestPasswordResetOutput{
			Success: true,
			Message: "パスワードリセット用のメールを送信しました。メールをご確認ください。",
		}, nil
	}

	slog.Info("Password reset email sent successfully",
		"admin_id", admin.AdminID().String())

	return &RequestPasswordResetOutput{
		Success: true,
		Message: "パスワードリセット用のメールを送信しました。メールをご確認ください。",
	}, nil
}
