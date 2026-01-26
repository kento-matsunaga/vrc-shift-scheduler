package auth

import (
	"context"
	"log/slog"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/auth"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/services"
)

// ResetPasswordWithTokenInput represents the input for resetting password with token
type ResetPasswordWithTokenInput struct {
	Token       string // パスワードリセットトークン
	NewPassword string // 新しいパスワード（平文）
}

// ResetPasswordWithTokenOutput represents the output for resetting password with token
type ResetPasswordWithTokenOutput struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// ResetPasswordWithTokenUsecase handles the password reset with token use case
type ResetPasswordWithTokenUsecase struct {
	adminRepo              auth.AdminRepository
	passwordResetTokenRepo auth.PasswordResetTokenRepository
	passwordHasher         services.PasswordHasher
	clock                  services.Clock
}

// NewResetPasswordWithTokenUsecase creates a new ResetPasswordWithTokenUsecase
func NewResetPasswordWithTokenUsecase(
	adminRepo auth.AdminRepository,
	passwordResetTokenRepo auth.PasswordResetTokenRepository,
	passwordHasher services.PasswordHasher,
	clock services.Clock,
) *ResetPasswordWithTokenUsecase {
	return &ResetPasswordWithTokenUsecase{
		adminRepo:              adminRepo,
		passwordResetTokenRepo: passwordResetTokenRepo,
		passwordHasher:         passwordHasher,
		clock:                  clock,
	}
}

// Execute executes the password reset with token use case
func (u *ResetPasswordWithTokenUsecase) Execute(ctx context.Context, input ResetPasswordWithTokenInput) (*ResetPasswordWithTokenOutput, error) {
	now := u.clock.Now()

	// Validate token
	if input.Token == "" {
		return nil, common.NewValidationError("token is required", nil)
	}

	// Validate password complexity
	if err := validatePasswordComplexity(input.NewPassword); err != nil {
		return nil, err
	}

	// Find the password reset token
	resetToken, err := u.passwordResetTokenRepo.FindByToken(ctx, input.Token)
	if err != nil {
		slog.Info("Password reset token not found or already used",
			"token_prefix", input.Token[:min(8, len(input.Token))]+"...")
		return nil, common.NewValidationError("無効または期限切れのトークンです", nil)
	}

	// Check if token can be used (not expired, not used)
	if err := resetToken.CanUse(now); err != nil {
		slog.Info("Password reset token cannot be used",
			"token_id", resetToken.TokenID().String(),
			"reason", err.Error())
		return nil, common.NewValidationError("無効または期限切れのトークンです", nil)
	}

	// Find the admin
	admin, err := u.adminRepo.FindByID(ctx, resetToken.AdminID())
	if err != nil {
		slog.Error("Admin not found for password reset token",
			"token_id", resetToken.TokenID().String(),
			"admin_id", resetToken.AdminID().String())
		return nil, common.NewValidationError("無効または期限切れのトークンです", nil)
	}

	// Check if admin is active and not deleted
	if !admin.IsActive() || admin.IsDeleted() {
		slog.Info("Password reset attempted for inactive/deleted admin",
			"admin_id", admin.AdminID().String())
		return nil, common.NewValidationError("このアカウントは現在パスワードをリセットできません", nil)
	}

	// Hash the new password
	newPasswordHash, err := u.passwordHasher.Hash(input.NewPassword)
	if err != nil {
		slog.Error("Failed to hash password",
			"admin_id", admin.AdminID().String(),
			"error", err)
		return nil, common.NewDomainError("ERR_INTERNAL", "パスワードの処理中にエラーが発生しました")
	}

	// Update admin's password
	if err := admin.UpdatePasswordHash(now, newPasswordHash); err != nil {
		slog.Error("Failed to update admin password",
			"admin_id", admin.AdminID().String(),
			"error", err)
		return nil, err
	}

	// Save the admin
	if err := u.adminRepo.Save(ctx, admin); err != nil {
		slog.Error("Failed to save admin after password reset",
			"admin_id", admin.AdminID().String(),
			"error", err)
		return nil, common.NewDomainError("ERR_INTERNAL", "パスワードの保存中にエラーが発生しました")
	}

	// Mark the token as used
	if err := resetToken.MarkAsUsed(now); err != nil {
		slog.Error("Failed to mark token as used",
			"token_id", resetToken.TokenID().String(),
			"error", err)
		// Continue - password was already reset successfully
	}

	// Save the token to persist the used_at timestamp
	if err := u.passwordResetTokenRepo.Save(ctx, resetToken); err != nil {
		slog.Error("Failed to save token after marking as used",
			"token_id", resetToken.TokenID().String(),
			"error", err)
		// Continue - password was already reset successfully
	}

	// Invalidate all other tokens for this admin
	if err := u.passwordResetTokenRepo.InvalidateAllByAdminID(ctx, admin.AdminID()); err != nil {
		slog.Error("Failed to invalidate other tokens",
			"admin_id", admin.AdminID().String(),
			"error", err)
		// Continue - password was already reset successfully
	}

	slog.Info("Password reset successful",
		"admin_id", admin.AdminID().String())

	return &ResetPasswordWithTokenOutput{
		Success: true,
		Message: "パスワードが正常にリセットされました",
	}, nil
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
