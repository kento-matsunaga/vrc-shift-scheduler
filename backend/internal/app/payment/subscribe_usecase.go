package payment

import (
	"context"
	"fmt"
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/auth"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/services"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/tenant"
)

// SubscribeInput represents the input for creating a new subscription
type SubscribeInput struct {
	Email       string
	Password    string
	TenantName  string
	Timezone    string
	DisplayName string
}

// SubscribeOutput represents the output of the subscribe usecase
type SubscribeOutput struct {
	CheckoutURL string `json:"checkout_url"`
	SessionID   string `json:"session_id"`
	TenantID    string `json:"tenant_id"`
	ExpiresAt   int64  `json:"expires_at"`
}

// SubscribeUsecase handles new subscription creation via Stripe Checkout
type SubscribeUsecase struct {
	txManager      services.TxManager
	tenantRepo     tenant.TenantRepository
	adminRepo      auth.AdminRepository
	passwordHasher services.PasswordHasher
	paymentGateway services.PaymentGateway
	clock          services.Clock
	successURL     string
	cancelURL      string
	stripePriceID  string
}

// NewSubscribeUsecase creates a new SubscribeUsecase
func NewSubscribeUsecase(
	txManager services.TxManager,
	tenantRepo tenant.TenantRepository,
	adminRepo auth.AdminRepository,
	passwordHasher services.PasswordHasher,
	paymentGateway services.PaymentGateway,
	clock services.Clock,
	successURL string,
	cancelURL string,
	stripePriceID string,
) *SubscribeUsecase {
	return &SubscribeUsecase{
		txManager:      txManager,
		tenantRepo:     tenantRepo,
		adminRepo:      adminRepo,
		passwordHasher: passwordHasher,
		paymentGateway: paymentGateway,
		clock:          clock,
		successURL:     successURL,
		cancelURL:      cancelURL,
		stripePriceID:  stripePriceID,
	}
}

// Execute creates a new tenant and admin in pending_payment status,
// then creates a Stripe Checkout Session
func (uc *SubscribeUsecase) Execute(ctx context.Context, input SubscribeInput) (*SubscribeOutput, error) {
	now := uc.clock.Now()

	// Validate input
	if err := uc.validateInput(input); err != nil {
		return nil, err
	}

	// Check if email already exists
	existingAdmin, _ := uc.adminRepo.FindByEmailGlobal(ctx, input.Email)
	if existingAdmin != nil {
		return nil, common.NewValidationError("このメールアドレスは既に登録されています", nil)
	}

	// Hash password
	passwordHash, err := uc.passwordHasher.Hash(input.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	var newTenant *tenant.Tenant
	var sessionResult *services.CheckoutSessionResult

	// NOTE: Stripe Checkout Session を先に作成し、その後 DB トランザクションを実行している。
	// もし DB トランザクションが失敗した場合、Stripe 側に孤立した Session が残るが、
	// Stripe Checkout Session は 24 時間で自動的に期限切れになるため、特別なクリーンアップ処理は不要。
	// この設計は意図的なものであり、Session ID を DB に保存してから Stripe API を呼ぶ方式より
	// シンプルで、失敗時のリカバリも容易である。
	//
	// Create Stripe Checkout Session first to get session ID and expiration
	// Stripe Checkout Session expires after 24 hours by default
	sessionResult, err = uc.paymentGateway.CreateCheckoutSession(services.CheckoutSessionParams{
		PriceID:       uc.stripePriceID,
		CustomerEmail: input.Email,
		SuccessURL:    uc.successURL,
		CancelURL:     uc.cancelURL,
		TenantID:      "", // Will be set after tenant creation
		TenantName:    input.TenantName,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create checkout session: %w", err)
	}

	// Use transaction to create tenant and admin atomically
	err = uc.txManager.WithTx(ctx, func(txCtx context.Context) error {
		// Create tenant in pending_payment status
		expiresAt := time.Unix(sessionResult.ExpiresAt, 0).UTC()
		newTenant, err = tenant.NewTenantPendingPayment(
			now,
			input.TenantName,
			input.Timezone,
			sessionResult.SessionID,
			expiresAt,
		)
		if err != nil {
			return fmt.Errorf("failed to create tenant: %w", err)
		}

		if err := uc.tenantRepo.Save(txCtx, newTenant); err != nil {
			return fmt.Errorf("failed to save tenant: %w", err)
		}

		// Create admin with owner role
		admin, err := auth.NewAdmin(
			now,
			newTenant.TenantID(),
			input.Email,
			passwordHash,
			input.DisplayName,
			auth.RoleOwner,
		)
		if err != nil {
			return fmt.Errorf("failed to create admin: %w", err)
		}

		if err := uc.adminRepo.Save(txCtx, admin); err != nil {
			return fmt.Errorf("failed to save admin: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &SubscribeOutput{
		CheckoutURL: sessionResult.URL,
		SessionID:   sessionResult.SessionID,
		TenantID:    newTenant.TenantID().String(),
		ExpiresAt:   sessionResult.ExpiresAt,
	}, nil
}

func (uc *SubscribeUsecase) validateInput(input SubscribeInput) error {
	if input.Email == "" {
		return common.NewValidationError("メールアドレスは必須です", nil)
	}
	if input.Password == "" {
		return common.NewValidationError("パスワードは必須です", nil)
	}
	if len(input.Password) < 8 {
		return common.NewValidationError("パスワードは8文字以上必要です", nil)
	}
	if input.TenantName == "" {
		return common.NewValidationError("組織名は必須です", nil)
	}
	if input.DisplayName == "" {
		return common.NewValidationError("表示名は必須です", nil)
	}
	if input.Timezone == "" {
		input.Timezone = "Asia/Tokyo"
	}
	return nil
}
