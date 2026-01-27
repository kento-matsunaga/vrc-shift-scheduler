package auth

import (
	"context"
	"testing"
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/auth"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
)

// =====================================================
// Mock Implementations for ResetPasswordWithTokenUsecase
// =====================================================

// MockTxManagerForPasswordReset is a mock implementation of TxManager
type MockTxManagerForPasswordReset struct {
	withTxFunc func(ctx context.Context, fn func(context.Context) error) error
}

func (m *MockTxManagerForPasswordReset) WithTx(ctx context.Context, fn func(context.Context) error) error {
	if m.withTxFunc != nil {
		return m.withTxFunc(ctx, fn)
	}
	// Default: just execute the function without transaction
	return fn(ctx)
}

// MockPasswordHasherForPasswordReset is a mock implementation of PasswordHasher
type MockPasswordHasherForPasswordReset struct {
	hashFunc    func(password string) (string, error)
	compareFunc func(hashedPassword, password string) error
}

func (m *MockPasswordHasherForPasswordReset) Hash(password string) (string, error) {
	if m.hashFunc != nil {
		return m.hashFunc(password)
	}
	return "hashed_" + password, nil
}

func (m *MockPasswordHasherForPasswordReset) Compare(hashedPassword, password string) error {
	if m.compareFunc != nil {
		return m.compareFunc(hashedPassword, password)
	}
	return nil
}

// =====================================================
// Tests for ResetPasswordWithTokenUsecase
// =====================================================

func TestResetPasswordWithTokenUsecase_Execute_Success(t *testing.T) {
	now := time.Now()
	testAdmin := createTestAdminForPasswordReset(t, now, true, false)

	// Create a valid token
	token, err := auth.NewPasswordResetToken(now, testAdmin.AdminID(), 1*time.Hour)
	if err != nil {
		t.Fatalf("Failed to create token: %v", err)
	}

	adminSaved := false
	tokenSaved := false
	tokensInvalidated := false

	adminRepo := &MockAdminRepositoryForPasswordReset{
		findByIDFunc: func(ctx context.Context, id common.AdminID) (*auth.Admin, error) {
			return testAdmin, nil
		},
		saveFunc: func(ctx context.Context, admin *auth.Admin) error {
			adminSaved = true
			return nil
		},
	}

	tokenRepo := &MockPasswordResetTokenRepository{
		findByTokenFunc: func(ctx context.Context, tokenStr string) (*auth.PasswordResetToken, error) {
			return token, nil
		},
		saveFunc: func(ctx context.Context, t *auth.PasswordResetToken) error {
			tokenSaved = true
			return nil
		},
		invalidateAllByAdminIDFunc: func(ctx context.Context, adminID common.AdminID) error {
			tokensInvalidated = true
			return nil
		},
	}

	passwordHasher := &MockPasswordHasherForPasswordReset{}
	clock := &MockPasswordResetClock{now: now}
	txManager := &MockTxManagerForPasswordReset{}

	usecase := NewResetPasswordWithTokenUsecase(adminRepo, tokenRepo, passwordHasher, clock, txManager)

	output, err := usecase.Execute(context.Background(), ResetPasswordWithTokenInput{
		Token:       token.Token(),
		NewPassword: "NewPassword123",
	})

	if err != nil {
		t.Fatalf("Execute() should succeed, got error: %v", err)
	}

	if !output.Success {
		t.Error("Output.Success should be true")
	}

	if !adminSaved {
		t.Error("Admin should have been saved")
	}

	if !tokenSaved {
		t.Error("Token should have been saved")
	}

	if !tokensInvalidated {
		t.Error("Other tokens should have been invalidated")
	}
}

func TestResetPasswordWithTokenUsecase_Execute_ExpiredToken(t *testing.T) {
	now := time.Now()
	testAdmin := createTestAdminForPasswordReset(t, now.Add(-2*time.Hour), true, false)

	// Create an expired token (created 2 hours ago, expires in 1 hour = expired 1 hour ago)
	token, err := auth.NewPasswordResetToken(now.Add(-2*time.Hour), testAdmin.AdminID(), 1*time.Hour)
	if err != nil {
		t.Fatalf("Failed to create token: %v", err)
	}

	adminRepo := &MockAdminRepositoryForPasswordReset{
		findByIDFunc: func(ctx context.Context, id common.AdminID) (*auth.Admin, error) {
			return testAdmin, nil
		},
	}

	tokenRepo := &MockPasswordResetTokenRepository{
		findByTokenFunc: func(ctx context.Context, tokenStr string) (*auth.PasswordResetToken, error) {
			return token, nil
		},
	}

	passwordHasher := &MockPasswordHasherForPasswordReset{}
	clock := &MockPasswordResetClock{now: now} // Current time is now, token expired 1 hour ago
	txManager := &MockTxManagerForPasswordReset{}

	usecase := NewResetPasswordWithTokenUsecase(adminRepo, tokenRepo, passwordHasher, clock, txManager)

	_, err = usecase.Execute(context.Background(), ResetPasswordWithTokenInput{
		Token:       token.Token(),
		NewPassword: "NewPassword123",
	})

	if err == nil {
		t.Error("Execute() should fail for expired token")
	}
}

func TestResetPasswordWithTokenUsecase_Execute_AlreadyUsedToken(t *testing.T) {
	now := time.Now()
	testAdmin := createTestAdminForPasswordReset(t, now, true, false)

	// Create a used token
	token, err := auth.NewPasswordResetToken(now, testAdmin.AdminID(), 1*time.Hour)
	if err != nil {
		t.Fatalf("Failed to create token: %v", err)
	}
	_ = token.MarkAsUsed(now.Add(10 * time.Minute))

	adminRepo := &MockAdminRepositoryForPasswordReset{
		findByIDFunc: func(ctx context.Context, id common.AdminID) (*auth.Admin, error) {
			return testAdmin, nil
		},
	}

	tokenRepo := &MockPasswordResetTokenRepository{
		findByTokenFunc: func(ctx context.Context, tokenStr string) (*auth.PasswordResetToken, error) {
			return token, nil
		},
	}

	passwordHasher := &MockPasswordHasherForPasswordReset{}
	clock := &MockPasswordResetClock{now: now.Add(20 * time.Minute)}
	txManager := &MockTxManagerForPasswordReset{}

	usecase := NewResetPasswordWithTokenUsecase(adminRepo, tokenRepo, passwordHasher, clock, txManager)

	_, err = usecase.Execute(context.Background(), ResetPasswordWithTokenInput{
		Token:       token.Token(),
		NewPassword: "NewPassword123",
	})

	if err == nil {
		t.Error("Execute() should fail for already used token")
	}
}

func TestResetPasswordWithTokenUsecase_Execute_InactiveAdmin(t *testing.T) {
	now := time.Now()
	testAdmin := createTestAdminForPasswordReset(t, now, false, false) // inactive

	token, err := auth.NewPasswordResetToken(now, testAdmin.AdminID(), 1*time.Hour)
	if err != nil {
		t.Fatalf("Failed to create token: %v", err)
	}

	adminRepo := &MockAdminRepositoryForPasswordReset{
		findByIDFunc: func(ctx context.Context, id common.AdminID) (*auth.Admin, error) {
			return testAdmin, nil
		},
	}

	tokenRepo := &MockPasswordResetTokenRepository{
		findByTokenFunc: func(ctx context.Context, tokenStr string) (*auth.PasswordResetToken, error) {
			return token, nil
		},
	}

	passwordHasher := &MockPasswordHasherForPasswordReset{}
	clock := &MockPasswordResetClock{now: now}
	txManager := &MockTxManagerForPasswordReset{}

	usecase := NewResetPasswordWithTokenUsecase(adminRepo, tokenRepo, passwordHasher, clock, txManager)

	_, err = usecase.Execute(context.Background(), ResetPasswordWithTokenInput{
		Token:       token.Token(),
		NewPassword: "NewPassword123",
	})

	if err == nil {
		t.Error("Execute() should fail for inactive admin")
	}
}

func TestResetPasswordWithTokenUsecase_Execute_DeletedAdmin(t *testing.T) {
	now := time.Now()
	testAdmin := createTestAdminForPasswordReset(t, now, true, true) // deleted

	token, err := auth.NewPasswordResetToken(now, testAdmin.AdminID(), 1*time.Hour)
	if err != nil {
		t.Fatalf("Failed to create token: %v", err)
	}

	adminRepo := &MockAdminRepositoryForPasswordReset{
		findByIDFunc: func(ctx context.Context, id common.AdminID) (*auth.Admin, error) {
			return testAdmin, nil
		},
	}

	tokenRepo := &MockPasswordResetTokenRepository{
		findByTokenFunc: func(ctx context.Context, tokenStr string) (*auth.PasswordResetToken, error) {
			return token, nil
		},
	}

	passwordHasher := &MockPasswordHasherForPasswordReset{}
	clock := &MockPasswordResetClock{now: now}
	txManager := &MockTxManagerForPasswordReset{}

	usecase := NewResetPasswordWithTokenUsecase(adminRepo, tokenRepo, passwordHasher, clock, txManager)

	_, err = usecase.Execute(context.Background(), ResetPasswordWithTokenInput{
		Token:       token.Token(),
		NewPassword: "NewPassword123",
	})

	if err == nil {
		t.Error("Execute() should fail for deleted admin")
	}
}

func TestResetPasswordWithTokenUsecase_Execute_WeakPassword(t *testing.T) {
	now := time.Now()
	testAdmin := createTestAdminForPasswordReset(t, now, true, false)

	token, err := auth.NewPasswordResetToken(now, testAdmin.AdminID(), 1*time.Hour)
	if err != nil {
		t.Fatalf("Failed to create token: %v", err)
	}

	adminRepo := &MockAdminRepositoryForPasswordReset{
		findByIDFunc: func(ctx context.Context, id common.AdminID) (*auth.Admin, error) {
			return testAdmin, nil
		},
	}

	tokenRepo := &MockPasswordResetTokenRepository{
		findByTokenFunc: func(ctx context.Context, tokenStr string) (*auth.PasswordResetToken, error) {
			return token, nil
		},
	}

	passwordHasher := &MockPasswordHasherForPasswordReset{}
	clock := &MockPasswordResetClock{now: now}
	txManager := &MockTxManagerForPasswordReset{}

	usecase := NewResetPasswordWithTokenUsecase(adminRepo, tokenRepo, passwordHasher, clock, txManager)

	testCases := []struct {
		name     string
		password string
	}{
		{"too short", "Abc123"},
		{"no uppercase", "password123"},
		{"no lowercase", "PASSWORD123"},
		{"no digit", "PasswordABC"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err = usecase.Execute(context.Background(), ResetPasswordWithTokenInput{
				Token:       token.Token(),
				NewPassword: tc.password,
			})

			if err == nil {
				t.Errorf("Execute() should fail for weak password: %s", tc.name)
			}
		})
	}
}

func TestResetPasswordWithTokenUsecase_Execute_TokenNotFound(t *testing.T) {
	now := time.Now()

	adminRepo := &MockAdminRepositoryForPasswordReset{}

	tokenRepo := &MockPasswordResetTokenRepository{
		findByTokenFunc: func(ctx context.Context, tokenStr string) (*auth.PasswordResetToken, error) {
			return nil, common.NewDomainError(common.ErrNotFound, "Token not found")
		},
	}

	passwordHasher := &MockPasswordHasherForPasswordReset{}
	clock := &MockPasswordResetClock{now: now}
	txManager := &MockTxManagerForPasswordReset{}

	usecase := NewResetPasswordWithTokenUsecase(adminRepo, tokenRepo, passwordHasher, clock, txManager)

	_, err := usecase.Execute(context.Background(), ResetPasswordWithTokenInput{
		Token:       "nonexistent_token_1234567890123456789012345678901234567890123456789012",
		NewPassword: "NewPassword123",
	})

	if err == nil {
		t.Error("Execute() should fail for non-existent token")
	}
}

func TestResetPasswordWithTokenUsecase_Execute_EmptyToken(t *testing.T) {
	now := time.Now()

	adminRepo := &MockAdminRepositoryForPasswordReset{}
	tokenRepo := &MockPasswordResetTokenRepository{}
	passwordHasher := &MockPasswordHasherForPasswordReset{}
	clock := &MockPasswordResetClock{now: now}
	txManager := &MockTxManagerForPasswordReset{}

	usecase := NewResetPasswordWithTokenUsecase(adminRepo, tokenRepo, passwordHasher, clock, txManager)

	_, err := usecase.Execute(context.Background(), ResetPasswordWithTokenInput{
		Token:       "",
		NewPassword: "NewPassword123",
	})

	if err == nil {
		t.Error("Execute() should fail for empty token")
	}
}
