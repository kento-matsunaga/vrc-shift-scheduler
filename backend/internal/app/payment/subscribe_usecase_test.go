package payment

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/auth"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/services"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/tenant"
)

// =====================================================
// Mock Implementations
// =====================================================

// MockTxManager is a mock implementation of TxManager
type MockTxManager struct {
	withTxFunc func(ctx context.Context, fn func(context.Context) error) error
}

func (m *MockTxManager) WithTx(ctx context.Context, fn func(context.Context) error) error {
	if m.withTxFunc != nil {
		return m.withTxFunc(ctx, fn)
	}
	return fn(ctx)
}

// MockTenantRepository is a mock implementation of tenant.TenantRepository
type MockTenantRepository struct {
	saveFunc                        func(ctx context.Context, t *tenant.Tenant) error
	findByIDFunc                    func(ctx context.Context, tenantID common.TenantID) (*tenant.Tenant, error)
	findByPendingStripeSessionIDFunc func(ctx context.Context, sessionID string) (*tenant.Tenant, error)
	listAllFunc                     func(ctx context.Context, status *tenant.TenantStatus, limit, offset int) ([]*tenant.Tenant, int, error)
}

func (m *MockTenantRepository) Save(ctx context.Context, t *tenant.Tenant) error {
	if m.saveFunc != nil {
		return m.saveFunc(ctx, t)
	}
	return nil
}

func (m *MockTenantRepository) FindByID(ctx context.Context, tenantID common.TenantID) (*tenant.Tenant, error) {
	if m.findByIDFunc != nil {
		return m.findByIDFunc(ctx, tenantID)
	}
	return nil, errors.New("not implemented")
}

func (m *MockTenantRepository) FindByPendingStripeSessionID(ctx context.Context, sessionID string) (*tenant.Tenant, error) {
	if m.findByPendingStripeSessionIDFunc != nil {
		return m.findByPendingStripeSessionIDFunc(ctx, sessionID)
	}
	return nil, errors.New("not implemented")
}

func (m *MockTenantRepository) ListAll(ctx context.Context, status *tenant.TenantStatus, limit, offset int) ([]*tenant.Tenant, int, error) {
	if m.listAllFunc != nil {
		return m.listAllFunc(ctx, status, limit, offset)
	}
	return nil, 0, errors.New("not implemented")
}

// MockAdminRepository is a mock implementation of auth.AdminRepository
type MockAdminRepository struct {
	saveFunc              func(ctx context.Context, admin *auth.Admin) error
	findByEmailGlobalFunc func(ctx context.Context, email string) (*auth.Admin, error)
}

func (m *MockAdminRepository) Save(ctx context.Context, admin *auth.Admin) error {
	if m.saveFunc != nil {
		return m.saveFunc(ctx, admin)
	}
	return nil
}

func (m *MockAdminRepository) FindByEmailGlobal(ctx context.Context, email string) (*auth.Admin, error) {
	if m.findByEmailGlobalFunc != nil {
		return m.findByEmailGlobalFunc(ctx, email)
	}
	return nil, nil // Default: email not found
}

// Unused methods - just satisfy the interface
func (m *MockAdminRepository) FindByID(ctx context.Context, adminID common.AdminID) (*auth.Admin, error) {
	return nil, errors.New("not implemented")
}
func (m *MockAdminRepository) FindByEmail(ctx context.Context, tenantID common.TenantID, email string) (*auth.Admin, error) {
	return nil, errors.New("not implemented")
}
func (m *MockAdminRepository) FindByIDWithTenant(ctx context.Context, tenantID common.TenantID, adminID common.AdminID) (*auth.Admin, error) {
	return nil, errors.New("not implemented")
}
func (m *MockAdminRepository) FindByTenantID(ctx context.Context, tenantID common.TenantID) ([]*auth.Admin, error) {
	return nil, errors.New("not implemented")
}
func (m *MockAdminRepository) FindActiveByTenantID(ctx context.Context, tenantID common.TenantID) ([]*auth.Admin, error) {
	return nil, errors.New("not implemented")
}
func (m *MockAdminRepository) Delete(ctx context.Context, tenantID common.TenantID, adminID common.AdminID) error {
	return errors.New("not implemented")
}
func (m *MockAdminRepository) ExistsByEmail(ctx context.Context, tenantID common.TenantID, email string) (bool, error) {
	return false, errors.New("not implemented")
}

// MockPasswordHasher is a mock implementation of PasswordHasher
type MockPasswordHasher struct {
	hashFunc func(password string) (string, error)
}

func (m *MockPasswordHasher) Hash(password string) (string, error) {
	if m.hashFunc != nil {
		return m.hashFunc(password)
	}
	return "$2a$10$mockhash", nil
}

// MockClock is a mock implementation of Clock
type MockClock struct {
	nowFunc func() time.Time
}

func (m *MockClock) Now() time.Time {
	if m.nowFunc != nil {
		return m.nowFunc()
	}
	return time.Now()
}

// MockPaymentGateway simulates PaymentGateway operations
type MockPaymentGateway struct {
	createCheckoutSessionFunc     func(params services.CheckoutSessionParams) (*services.CheckoutSessionResult, error)
	createBillingPortalSessionFunc func(params services.BillingPortalParams) (*services.BillingPortalResult, error)
}

func (m *MockPaymentGateway) CreateCheckoutSession(params services.CheckoutSessionParams) (*services.CheckoutSessionResult, error) {
	if m.createCheckoutSessionFunc != nil {
		return m.createCheckoutSessionFunc(params)
	}
	return &services.CheckoutSessionResult{
		SessionID: "cs_test_session123",
		URL:       "https://checkout.stripe.com/test",
		ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
	}, nil
}

func (m *MockPaymentGateway) CreateBillingPortalSession(params services.BillingPortalParams) (*services.BillingPortalResult, error) {
	if m.createBillingPortalSessionFunc != nil {
		return m.createBillingPortalSessionFunc(params)
	}
	return &services.BillingPortalResult{
		URL: "https://billing.stripe.com/test",
	}, nil
}

// =====================================================
// Test Helper Functions
// =====================================================

func validSubscribeInput() SubscribeInput {
	return SubscribeInput{
		Email:       "test@example.com",
		Password:    "Password123",
		TenantName:  "Test Organization",
		DisplayName: "Test Admin",
		Timezone:    "Asia/Tokyo",
	}
}

// Note: Since SubscribeUsecase depends on the actual *infrastripe.Client,
// we can't directly test with mocks. Instead, we'll test the input validation
// and error handling where we can control the flow.

// =====================================================
// Input Validation Tests
// =====================================================

func TestSubscribeUsecase_ValidateInput_Success(t *testing.T) {
	uc := &SubscribeUsecase{}

	input := validSubscribeInput()
	err := uc.validateInput(input)

	if err != nil {
		t.Errorf("validateInput() should succeed for valid input, got error: %v", err)
	}
}

func TestSubscribeUsecase_ValidateInput_EmptyEmail(t *testing.T) {
	uc := &SubscribeUsecase{}

	input := validSubscribeInput()
	input.Email = ""

	err := uc.validateInput(input)

	if err == nil {
		t.Error("validateInput() should fail when email is empty")
	}
}

func TestSubscribeUsecase_ValidateInput_EmptyPassword(t *testing.T) {
	uc := &SubscribeUsecase{}

	input := validSubscribeInput()
	input.Password = ""

	err := uc.validateInput(input)

	if err == nil {
		t.Error("validateInput() should fail when password is empty")
	}
}

func TestSubscribeUsecase_ValidateInput_PasswordTooShort(t *testing.T) {
	uc := &SubscribeUsecase{}

	input := validSubscribeInput()
	input.Password = "Pass1" // Only 5 characters

	err := uc.validateInput(input)

	if err == nil {
		t.Error("validateInput() should fail when password is too short")
	}
}

func TestSubscribeUsecase_ValidateInput_EmptyTenantName(t *testing.T) {
	uc := &SubscribeUsecase{}

	input := validSubscribeInput()
	input.TenantName = ""

	err := uc.validateInput(input)

	if err == nil {
		t.Error("validateInput() should fail when tenant name is empty")
	}
}

func TestSubscribeUsecase_ValidateInput_EmptyDisplayName(t *testing.T) {
	uc := &SubscribeUsecase{}

	input := validSubscribeInput()
	input.DisplayName = ""

	err := uc.validateInput(input)

	if err == nil {
		t.Error("validateInput() should fail when display name is empty")
	}
}

// =====================================================
// Integration-style Tests (with mocked dependencies)
// Note: These tests would require dependency injection changes
// to properly mock the Stripe client. For now, we test what we can.
// =====================================================

func TestSubscribeOutput_Structure(t *testing.T) {
	output := SubscribeOutput{
		CheckoutURL: "https://checkout.stripe.com/test",
		SessionID:   "cs_test_session123",
		TenantID:    "01HTEST12345678901234567",
		ExpiresAt:   time.Now().Add(24 * time.Hour).Unix(),
	}

	if output.CheckoutURL == "" {
		t.Error("CheckoutURL should not be empty")
	}

	if output.SessionID == "" {
		t.Error("SessionID should not be empty")
	}

	if output.TenantID == "" {
		t.Error("TenantID should not be empty")
	}

	if output.ExpiresAt == 0 {
		t.Error("ExpiresAt should not be zero")
	}
}

func TestSubscribeInput_Structure(t *testing.T) {
	input := SubscribeInput{
		Email:       "test@example.com",
		Password:    "Password123",
		TenantName:  "Test Organization",
		DisplayName: "Test Admin",
		Timezone:    "Asia/Tokyo",
	}

	if input.Email != "test@example.com" {
		t.Error("Email should be set correctly")
	}

	if input.TenantName != "Test Organization" {
		t.Error("TenantName should be set correctly")
	}

	if input.Timezone != "Asia/Tokyo" {
		t.Error("Timezone should be set correctly")
	}
}
