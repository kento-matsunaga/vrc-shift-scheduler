package auth_test

import (
	"testing"
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/auth"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
)

// =====================================================
// PasswordResetToken Entity Tests
// =====================================================

func TestNewPasswordResetToken_Success(t *testing.T) {
	now := time.Now()
	adminID := common.NewAdminID()

	token, err := auth.NewPasswordResetToken(
		now,
		adminID,
		1*time.Hour, // 1 hour expiration
	)

	if err != nil {
		t.Fatalf("NewPasswordResetToken() should succeed, got error: %v", err)
	}

	if token.TokenID() == "" {
		t.Error("TokenID should be set")
	}

	if token.AdminID() != adminID {
		t.Errorf("AdminID mismatch: got %v, want %v", token.AdminID(), adminID)
	}

	if token.Token() == "" {
		t.Error("Token should be generated")
	}

	if len(token.Token()) != 64 { // 32 bytes = 64 hex characters
		t.Errorf("Token should be 64 characters (32 bytes hex), got %d", len(token.Token()))
	}

	if token.UsedAt() != nil {
		t.Error("UsedAt should be nil initially")
	}

	if token.IsUsed() {
		t.Error("IsUsed() should return false initially")
	}

	// Check expiration is set correctly
	expectedExpiresAt := now.Add(1 * time.Hour)
	if token.ExpiresAt().Sub(expectedExpiresAt) > time.Second {
		t.Errorf("ExpiresAt mismatch: got %v, want approximately %v", token.ExpiresAt(), expectedExpiresAt)
	}
}

func TestPasswordResetToken_IsExpired(t *testing.T) {
	now := time.Now()
	adminID := common.NewAdminID()

	token, err := auth.NewPasswordResetToken(
		now,
		adminID,
		1*time.Hour, // 1 hour expiration
	)
	if err != nil {
		t.Fatalf("NewPasswordResetToken() failed: %v", err)
	}

	// Before expiration
	if token.IsExpired(now.Add(30 * time.Minute)) {
		t.Error("Token should not be expired 30 minutes after creation")
	}

	// After expiration
	if !token.IsExpired(now.Add(2 * time.Hour)) {
		t.Error("Token should be expired 2 hours after creation")
	}
}

func TestPasswordResetToken_CanUse_Success(t *testing.T) {
	now := time.Now()
	adminID := common.NewAdminID()

	token, _ := auth.NewPasswordResetToken(
		now,
		adminID,
		24*time.Hour,
	)

	err := token.CanUse(now.Add(1 * time.Hour))
	if err != nil {
		t.Errorf("CanUse() should succeed, got error: %v", err)
	}
}

func TestPasswordResetToken_CanUse_ErrorWhenExpired(t *testing.T) {
	now := time.Now()
	adminID := common.NewAdminID()

	token, _ := auth.NewPasswordResetToken(
		now,
		adminID,
		1*time.Hour,
	)

	err := token.CanUse(now.Add(2 * time.Hour))
	if err == nil {
		t.Error("CanUse() should fail when token is expired")
	}
}

func TestPasswordResetToken_CanUse_ErrorWhenAlreadyUsed(t *testing.T) {
	now := time.Now()
	adminID := common.NewAdminID()

	token, _ := auth.NewPasswordResetToken(
		now,
		adminID,
		24*time.Hour,
	)

	// First use
	err := token.MarkAsUsed(now.Add(1 * time.Hour))
	if err != nil {
		t.Fatalf("First MarkAsUsed() should succeed: %v", err)
	}

	// Try to check if can use again
	err = token.CanUse(now.Add(2 * time.Hour))
	if err == nil {
		t.Error("CanUse() should fail when token is already used")
	}
}

func TestPasswordResetToken_MarkAsUsed_Success(t *testing.T) {
	now := time.Now()
	adminID := common.NewAdminID()

	token, _ := auth.NewPasswordResetToken(
		now,
		adminID,
		24*time.Hour,
	)

	useTime := now.Add(1 * time.Hour)
	err := token.MarkAsUsed(useTime)

	if err != nil {
		t.Fatalf("MarkAsUsed() should succeed, got error: %v", err)
	}

	if !token.IsUsed() {
		t.Error("IsUsed() should return true after MarkAsUsed()")
	}

	if token.UsedAt() == nil {
		t.Error("UsedAt() should not be nil after MarkAsUsed()")
	}
}

func TestPasswordResetToken_MarkAsUsed_ErrorWhenExpired(t *testing.T) {
	now := time.Now()
	adminID := common.NewAdminID()

	token, _ := auth.NewPasswordResetToken(
		now,
		adminID,
		1*time.Hour,
	)

	err := token.MarkAsUsed(now.Add(2 * time.Hour))
	if err == nil {
		t.Error("MarkAsUsed() should fail when token is expired")
	}
}

func TestPasswordResetToken_MarkAsUsed_ErrorWhenAlreadyUsed(t *testing.T) {
	now := time.Now()
	adminID := common.NewAdminID()

	token, _ := auth.NewPasswordResetToken(
		now,
		adminID,
		24*time.Hour,
	)

	// First use
	_ = token.MarkAsUsed(now.Add(1 * time.Hour))

	// Second use attempt
	err := token.MarkAsUsed(now.Add(2 * time.Hour))
	if err == nil {
		t.Error("MarkAsUsed() should fail when token is already used")
	}
}

func TestReconstructPasswordResetToken_Success(t *testing.T) {
	tokenID := common.NewPasswordResetTokenID()
	adminID := common.NewAdminID()
	now := time.Now()
	expiresAt := now.Add(1 * time.Hour)

	token, err := auth.ReconstructPasswordResetToken(
		tokenID,
		adminID,
		"abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
		expiresAt,
		nil,
		now,
	)

	if err != nil {
		t.Fatalf("ReconstructPasswordResetToken() should succeed, got error: %v", err)
	}

	if token.TokenID() != tokenID {
		t.Errorf("TokenID mismatch: got %v, want %v", token.TokenID(), tokenID)
	}

	if token.AdminID() != adminID {
		t.Errorf("AdminID mismatch: got %v, want %v", token.AdminID(), adminID)
	}
}

func TestReconstructPasswordResetToken_ErrorWhenTokenEmpty(t *testing.T) {
	tokenID := common.NewPasswordResetTokenID()
	adminID := common.NewAdminID()
	now := time.Now()
	expiresAt := now.Add(1 * time.Hour)

	_, err := auth.ReconstructPasswordResetToken(
		tokenID,
		adminID,
		"", // Empty token
		expiresAt,
		nil,
		now,
	)

	if err == nil {
		t.Error("ReconstructPasswordResetToken() should fail when token is empty")
	}
}

func TestReconstructPasswordResetToken_ErrorWhenTokenWrongLength(t *testing.T) {
	tokenID := common.NewPasswordResetTokenID()
	adminID := common.NewAdminID()
	now := time.Now()
	expiresAt := now.Add(1 * time.Hour)

	_, err := auth.ReconstructPasswordResetToken(
		tokenID,
		adminID,
		"tooshort", // Wrong length
		expiresAt,
		nil,
		now,
	)

	if err == nil {
		t.Error("ReconstructPasswordResetToken() should fail when token has wrong length")
	}
}
