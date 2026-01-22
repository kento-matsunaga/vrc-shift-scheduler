package payment

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/billing"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/tenant"
)

// =====================================================
// Mock Implementations for Webhook Tests
// =====================================================

// MockWebhookTxManager is a mock implementation of TxManager
type MockWebhookTxManager struct {
	withTxFunc func(ctx context.Context, fn func(context.Context) error) error
}

func (m *MockWebhookTxManager) WithTx(ctx context.Context, fn func(context.Context) error) error {
	if m.withTxFunc != nil {
		return m.withTxFunc(ctx, fn)
	}
	return fn(ctx)
}

// MockWebhookTenantRepository is a mock implementation of tenant.TenantRepository
type MockWebhookTenantRepository struct {
	findByIDFunc                    func(ctx context.Context, tenantID common.TenantID) (*tenant.Tenant, error)
	findByPendingStripeSessionIDFunc func(ctx context.Context, sessionID string) (*tenant.Tenant, error)
	saveFunc                        func(ctx context.Context, t *tenant.Tenant) error
	listAllFunc                     func(ctx context.Context, status *tenant.TenantStatus, limit, offset int) ([]*tenant.Tenant, int, error)
}

func (m *MockWebhookTenantRepository) FindByID(ctx context.Context, tenantID common.TenantID) (*tenant.Tenant, error) {
	if m.findByIDFunc != nil {
		return m.findByIDFunc(ctx, tenantID)
	}
	return nil, errors.New("not implemented")
}

func (m *MockWebhookTenantRepository) FindByPendingStripeSessionID(ctx context.Context, sessionID string) (*tenant.Tenant, error) {
	if m.findByPendingStripeSessionIDFunc != nil {
		return m.findByPendingStripeSessionIDFunc(ctx, sessionID)
	}
	return nil, errors.New("not implemented")
}

func (m *MockWebhookTenantRepository) Save(ctx context.Context, t *tenant.Tenant) error {
	if m.saveFunc != nil {
		return m.saveFunc(ctx, t)
	}
	return nil
}

func (m *MockWebhookTenantRepository) ListAll(ctx context.Context, status *tenant.TenantStatus, limit, offset int) ([]*tenant.Tenant, int, error) {
	if m.listAllFunc != nil {
		return m.listAllFunc(ctx, status, limit, offset)
	}
	return nil, 0, errors.New("not implemented")
}

// MockWebhookSubscriptionRepository is a mock implementation of billing.SubscriptionRepository
type MockWebhookSubscriptionRepository struct {
	saveFunc                       func(ctx context.Context, sub *billing.Subscription) error
	findByTenantIDFunc             func(ctx context.Context, tenantID common.TenantID) (*billing.Subscription, error)
	findByStripeSubscriptionIDFunc func(ctx context.Context, stripeSubID string) (*billing.Subscription, error)
}

func (m *MockWebhookSubscriptionRepository) Save(ctx context.Context, sub *billing.Subscription) error {
	if m.saveFunc != nil {
		return m.saveFunc(ctx, sub)
	}
	return nil
}

func (m *MockWebhookSubscriptionRepository) FindByTenantID(ctx context.Context, tenantID common.TenantID) (*billing.Subscription, error) {
	if m.findByTenantIDFunc != nil {
		return m.findByTenantIDFunc(ctx, tenantID)
	}
	return nil, errors.New("not implemented")
}

func (m *MockWebhookSubscriptionRepository) FindByStripeSubscriptionID(ctx context.Context, stripeSubID string) (*billing.Subscription, error) {
	if m.findByStripeSubscriptionIDFunc != nil {
		return m.findByStripeSubscriptionIDFunc(ctx, stripeSubID)
	}
	return nil, errors.New("not implemented")
}

// MockWebhookEntitlementRepository is a mock implementation of billing.EntitlementRepository
type MockWebhookEntitlementRepository struct {
	saveFunc func(ctx context.Context, entitlement *billing.Entitlement) error
}

func (m *MockWebhookEntitlementRepository) Save(ctx context.Context, entitlement *billing.Entitlement) error {
	if m.saveFunc != nil {
		return m.saveFunc(ctx, entitlement)
	}
	return nil
}

// Unused methods
func (m *MockWebhookEntitlementRepository) FindByID(ctx context.Context, entitlementID billing.EntitlementID) (*billing.Entitlement, error) {
	return nil, errors.New("not implemented")
}
func (m *MockWebhookEntitlementRepository) FindByTenantID(ctx context.Context, tenantID common.TenantID) ([]*billing.Entitlement, error) {
	return nil, errors.New("not implemented")
}
func (m *MockWebhookEntitlementRepository) FindActiveByTenantID(ctx context.Context, tenantID common.TenantID) (*billing.Entitlement, error) {
	return nil, errors.New("not implemented")
}
func (m *MockWebhookEntitlementRepository) HasRevokedByTenantID(ctx context.Context, tenantID common.TenantID) (bool, error) {
	return false, errors.New("not implemented")
}

// MockWebhookEventRepository is a mock implementation of billing.WebhookEventRepository
type MockWebhookEventRepository struct {
	tryInsertFunc       func(ctx context.Context, provider string, eventID string, payloadJSON *string) (bool, error)
	deleteOlderThanFunc func(ctx context.Context, before int) (int64, error)
}

func (m *MockWebhookEventRepository) TryInsert(ctx context.Context, provider string, eventID string, payloadJSON *string) (bool, error) {
	if m.tryInsertFunc != nil {
		return m.tryInsertFunc(ctx, provider, eventID, payloadJSON)
	}
	return true, nil // Default: new event
}

func (m *MockWebhookEventRepository) DeleteOlderThan(ctx context.Context, before int) (int64, error) {
	if m.deleteOlderThanFunc != nil {
		return m.deleteOlderThanFunc(ctx, before)
	}
	return 0, nil
}

// MockWebhookBillingAuditLogRepository is a mock implementation of billing.BillingAuditLogRepository
type MockWebhookBillingAuditLogRepository struct {
	saveFunc func(ctx context.Context, log *billing.BillingAuditLog) error
}

func (m *MockWebhookBillingAuditLogRepository) Save(ctx context.Context, log *billing.BillingAuditLog) error {
	if m.saveFunc != nil {
		return m.saveFunc(ctx, log)
	}
	return nil
}

// Unused methods
func (m *MockWebhookBillingAuditLogRepository) FindByDateRange(ctx context.Context, startDate, endDate string, limit, offset int) ([]*billing.BillingAuditLog, error) {
	return nil, errors.New("not implemented")
}
func (m *MockWebhookBillingAuditLogRepository) FindByAction(ctx context.Context, action string, limit, offset int) ([]*billing.BillingAuditLog, error) {
	return nil, errors.New("not implemented")
}
func (m *MockWebhookBillingAuditLogRepository) CountByDateRange(ctx context.Context, startDate, endDate string) (int, error) {
	return 0, errors.New("not implemented")
}
func (m *MockWebhookBillingAuditLogRepository) List(ctx context.Context, action *string, limit, offset int) ([]*billing.BillingAuditLog, int, error) {
	return nil, 0, errors.New("not implemented")
}

// =====================================================
// Test Helper Functions
// =====================================================

func createTestPendingPaymentTenant(t *testing.T) *tenant.Tenant {
	t.Helper()
	now := time.Now()
	expiresAt := now.Add(24 * time.Hour)

	testTenant, err := tenant.NewTenantPendingPayment(now, "Test Tenant", "Asia/Tokyo", "cs_test_session123", expiresAt)
	if err != nil {
		t.Fatalf("Failed to create test tenant: %v", err)
	}
	return testTenant
}

func createTestActiveTenant(t *testing.T) *tenant.Tenant {
	t.Helper()
	now := time.Now()

	testTenant, err := tenant.NewTenant(now, "Test Tenant", "Asia/Tokyo")
	if err != nil {
		t.Fatalf("Failed to create test tenant: %v", err)
	}
	return testTenant
}

func createTestWebhookSubscription(t *testing.T, tenantID common.TenantID) *billing.Subscription {
	t.Helper()
	now := time.Now()
	periodEnd := now.Add(30 * 24 * time.Hour)

	sub, err := billing.NewSubscription(now, tenantID, "cus_test123", "sub_test123", billing.SubscriptionStatusActive, &periodEnd)
	if err != nil {
		t.Fatalf("Failed to create test subscription: %v", err)
	}
	return sub
}

func createWebhookUsecase(
	txManager *MockWebhookTxManager,
	tenantRepo *MockWebhookTenantRepository,
	subscriptionRepo *MockWebhookSubscriptionRepository,
	entitlementRepo *MockWebhookEntitlementRepository,
	webhookEventRepo *MockWebhookEventRepository,
	auditLogRepo *MockWebhookBillingAuditLogRepository,
) *StripeWebhookUsecase {
	if txManager == nil {
		txManager = &MockWebhookTxManager{}
	}
	if tenantRepo == nil {
		tenantRepo = &MockWebhookTenantRepository{}
	}
	if subscriptionRepo == nil {
		subscriptionRepo = &MockWebhookSubscriptionRepository{}
	}
	if entitlementRepo == nil {
		entitlementRepo = &MockWebhookEntitlementRepository{}
	}
	if webhookEventRepo == nil {
		webhookEventRepo = &MockWebhookEventRepository{}
	}
	if auditLogRepo == nil {
		auditLogRepo = &MockWebhookBillingAuditLogRepository{}
	}

	return NewStripeWebhookUsecase(
		txManager,
		tenantRepo,
		subscriptionRepo,
		entitlementRepo,
		webhookEventRepo,
		auditLogRepo,
	)
}

// =====================================================
// HandleWebhook Tests - Idempotency
// =====================================================

func TestStripeWebhookUsecase_HandleWebhook_DuplicateEvent(t *testing.T) {
	webhookEventRepo := &MockWebhookEventRepository{
		tryInsertFunc: func(ctx context.Context, provider string, eventID string, payloadJSON *string) (bool, error) {
			return false, nil // Event already exists
		},
	}

	usecase := createWebhookUsecase(nil, nil, nil, nil, webhookEventRepo, nil)

	event := StripeEvent{
		ID:   "evt_test123",
		Type: "checkout.session.completed",
		Data: StripeEventData{Object: json.RawMessage(`{}`)},
	}

	processed, err := usecase.HandleWebhook(context.Background(), event, "{}")

	if err != nil {
		t.Errorf("HandleWebhook() should not return error for duplicate, got: %v", err)
	}

	if processed {
		t.Error("HandleWebhook() should return processed=false for duplicate event")
	}
}

func TestStripeWebhookUsecase_HandleWebhook_IdempotencyCheckFails(t *testing.T) {
	dbError := errors.New("database error")

	webhookEventRepo := &MockWebhookEventRepository{
		tryInsertFunc: func(ctx context.Context, provider string, eventID string, payloadJSON *string) (bool, error) {
			return false, dbError
		},
	}

	usecase := createWebhookUsecase(nil, nil, nil, nil, webhookEventRepo, nil)

	event := StripeEvent{
		ID:   "evt_test123",
		Type: "checkout.session.completed",
		Data: StripeEventData{Object: json.RawMessage(`{}`)},
	}

	processed, err := usecase.HandleWebhook(context.Background(), event, "{}")

	if err == nil {
		t.Error("HandleWebhook() should return error when idempotency check fails")
	}

	if processed {
		t.Error("HandleWebhook() should return processed=false when error occurs")
	}
}

// =====================================================
// HandleWebhook Tests - Unknown Event Type
// =====================================================

func TestStripeWebhookUsecase_HandleWebhook_UnknownEventType(t *testing.T) {
	webhookEventRepo := &MockWebhookEventRepository{
		tryInsertFunc: func(ctx context.Context, provider string, eventID string, payloadJSON *string) (bool, error) {
			return true, nil
		},
	}

	usecase := createWebhookUsecase(nil, nil, nil, nil, webhookEventRepo, nil)

	event := StripeEvent{
		ID:   "evt_test123",
		Type: "some.unknown.event",
		Data: StripeEventData{Object: json.RawMessage(`{}`)},
	}

	processed, err := usecase.HandleWebhook(context.Background(), event, "{}")

	if err != nil {
		t.Errorf("HandleWebhook() should not return error for unknown event type, got: %v", err)
	}

	if !processed {
		t.Error("HandleWebhook() should return processed=true for unknown event type (logged and ignored)")
	}
}

// =====================================================
// HandleWebhook Tests - checkout.session.completed
// =====================================================

func TestStripeWebhookUsecase_CheckoutSessionCompleted_Success(t *testing.T) {
	testTenant := createTestPendingPaymentTenant(t)

	var savedTenant *tenant.Tenant
	var savedSubscription *billing.Subscription
	var savedEntitlement *billing.Entitlement

	tenantRepo := &MockWebhookTenantRepository{
		findByPendingStripeSessionIDFunc: func(ctx context.Context, sessionID string) (*tenant.Tenant, error) {
			if sessionID == "cs_test_session123" {
				return testTenant, nil
			}
			return nil, common.NewNotFoundError("tenant", sessionID)
		},
		saveFunc: func(ctx context.Context, t *tenant.Tenant) error {
			savedTenant = t
			return nil
		},
	}

	subscriptionRepo := &MockWebhookSubscriptionRepository{
		saveFunc: func(ctx context.Context, sub *billing.Subscription) error {
			savedSubscription = sub
			return nil
		},
	}

	entitlementRepo := &MockWebhookEntitlementRepository{
		saveFunc: func(ctx context.Context, ent *billing.Entitlement) error {
			savedEntitlement = ent
			return nil
		},
	}

	webhookEventRepo := &MockWebhookEventRepository{
		tryInsertFunc: func(ctx context.Context, provider string, eventID string, payloadJSON *string) (bool, error) {
			return true, nil
		},
	}

	auditLogRepo := &MockWebhookBillingAuditLogRepository{
		saveFunc: func(ctx context.Context, log *billing.BillingAuditLog) error {
			return nil
		},
	}

	usecase := createWebhookUsecase(nil, tenantRepo, subscriptionRepo, entitlementRepo, webhookEventRepo, auditLogRepo)

	sessionData := StripeCheckoutSession{
		ID:           "cs_test_session123",
		Customer:     "cus_test123",
		Subscription: "sub_test123",
		Status:       "complete",
		Mode:         "subscription",
	}
	sessionJSON, _ := json.Marshal(sessionData)

	event := StripeEvent{
		ID:   "evt_test123",
		Type: "checkout.session.completed",
		Data: StripeEventData{Object: sessionJSON},
	}

	processed, err := usecase.HandleWebhook(context.Background(), event, string(sessionJSON))

	if err != nil {
		t.Fatalf("HandleWebhook() should succeed, got error: %v", err)
	}

	if !processed {
		t.Error("HandleWebhook() should return processed=true")
	}

	if savedTenant == nil {
		t.Error("Tenant should have been saved")
	} else if savedTenant.Status() != tenant.TenantStatusActive {
		t.Errorf("Tenant status should be active, got %s", savedTenant.Status())
	}

	if savedSubscription == nil {
		t.Error("Subscription should have been saved")
	}

	if savedEntitlement == nil {
		t.Error("Entitlement should have been saved")
	}
}

func TestStripeWebhookUsecase_CheckoutSessionCompleted_NonSubscriptionMode(t *testing.T) {
	webhookEventRepo := &MockWebhookEventRepository{
		tryInsertFunc: func(ctx context.Context, provider string, eventID string, payloadJSON *string) (bool, error) {
			return true, nil
		},
	}

	usecase := createWebhookUsecase(nil, nil, nil, nil, webhookEventRepo, nil)

	sessionData := StripeCheckoutSession{
		ID:     "cs_test_session123",
		Status: "complete",
		Mode:   "payment", // Not subscription
	}
	sessionJSON, _ := json.Marshal(sessionData)

	event := StripeEvent{
		ID:   "evt_test123",
		Type: "checkout.session.completed",
		Data: StripeEventData{Object: sessionJSON},
	}

	processed, err := usecase.HandleWebhook(context.Background(), event, string(sessionJSON))

	if err != nil {
		t.Errorf("HandleWebhook() should not return error for non-subscription mode, got: %v", err)
	}

	if !processed {
		t.Error("HandleWebhook() should return processed=true")
	}
}

func TestStripeWebhookUsecase_CheckoutSessionCompleted_TenantNotFound(t *testing.T) {
	tenantRepo := &MockWebhookTenantRepository{
		findByPendingStripeSessionIDFunc: func(ctx context.Context, sessionID string) (*tenant.Tenant, error) {
			return nil, common.NewNotFoundError("tenant", sessionID)
		},
	}

	webhookEventRepo := &MockWebhookEventRepository{
		tryInsertFunc: func(ctx context.Context, provider string, eventID string, payloadJSON *string) (bool, error) {
			return true, nil
		},
	}

	usecase := createWebhookUsecase(nil, tenantRepo, nil, nil, webhookEventRepo, nil)

	sessionData := StripeCheckoutSession{
		ID:           "cs_test_session123",
		Customer:     "cus_test123",
		Subscription: "sub_test123",
		Mode:         "subscription",
	}
	sessionJSON, _ := json.Marshal(sessionData)

	event := StripeEvent{
		ID:   "evt_test123",
		Type: "checkout.session.completed",
		Data: StripeEventData{Object: sessionJSON},
	}

	processed, err := usecase.HandleWebhook(context.Background(), event, string(sessionJSON))

	// Should not error - just log and continue (tenant not found is okay for webhooks)
	if err != nil {
		t.Errorf("HandleWebhook() should not return error when tenant not found, got: %v", err)
	}

	if !processed {
		t.Error("HandleWebhook() should return processed=true")
	}
}

// =====================================================
// HandleWebhook Tests - invoice.paid
// =====================================================

func TestStripeWebhookUsecase_InvoicePaid_Success(t *testing.T) {
	testTenant := createTestActiveTenant(t)
	testSubscription := createTestWebhookSubscription(t, testTenant.TenantID())

	var savedTenant *tenant.Tenant

	tenantRepo := &MockWebhookTenantRepository{
		findByIDFunc: func(ctx context.Context, tenantID common.TenantID) (*tenant.Tenant, error) {
			return testTenant, nil
		},
		saveFunc: func(ctx context.Context, t *tenant.Tenant) error {
			savedTenant = t
			return nil
		},
	}

	subscriptionRepo := &MockWebhookSubscriptionRepository{
		findByStripeSubscriptionIDFunc: func(ctx context.Context, stripeSubID string) (*billing.Subscription, error) {
			if stripeSubID == "sub_test123" {
				return testSubscription, nil
			}
			return nil, nil
		},
		saveFunc: func(ctx context.Context, sub *billing.Subscription) error {
			return nil
		},
	}

	webhookEventRepo := &MockWebhookEventRepository{
		tryInsertFunc: func(ctx context.Context, provider string, eventID string, payloadJSON *string) (bool, error) {
			return true, nil
		},
	}

	auditLogRepo := &MockWebhookBillingAuditLogRepository{
		saveFunc: func(ctx context.Context, log *billing.BillingAuditLog) error {
			return nil
		},
	}

	usecase := createWebhookUsecase(nil, tenantRepo, subscriptionRepo, nil, webhookEventRepo, auditLogRepo)

	invoiceData := StripeInvoice{
		ID:           "in_test123",
		Customer:     "cus_test123",
		Subscription: "sub_test123",
		Status:       "paid",
		Paid:         true,
	}
	invoiceJSON, _ := json.Marshal(invoiceData)

	event := StripeEvent{
		ID:   "evt_test123",
		Type: "invoice.paid",
		Data: StripeEventData{Object: invoiceJSON},
	}

	processed, err := usecase.HandleWebhook(context.Background(), event, string(invoiceJSON))

	if err != nil {
		t.Fatalf("HandleWebhook() should succeed, got error: %v", err)
	}

	if !processed {
		t.Error("HandleWebhook() should return processed=true")
	}

	if savedTenant == nil {
		t.Error("Tenant should have been saved")
	}
}

func TestStripeWebhookUsecase_InvoicePaid_NoSubscription(t *testing.T) {
	webhookEventRepo := &MockWebhookEventRepository{
		tryInsertFunc: func(ctx context.Context, provider string, eventID string, payloadJSON *string) (bool, error) {
			return true, nil
		},
	}

	usecase := createWebhookUsecase(nil, nil, nil, nil, webhookEventRepo, nil)

	invoiceData := StripeInvoice{
		ID:           "in_test123",
		Customer:     "cus_test123",
		Subscription: "", // Empty subscription
		Status:       "paid",
		Paid:         true,
	}
	invoiceJSON, _ := json.Marshal(invoiceData)

	event := StripeEvent{
		ID:   "evt_test123",
		Type: "invoice.paid",
		Data: StripeEventData{Object: invoiceJSON},
	}

	processed, err := usecase.HandleWebhook(context.Background(), event, string(invoiceJSON))

	if err != nil {
		t.Errorf("HandleWebhook() should not return error for non-subscription invoice, got: %v", err)
	}

	if !processed {
		t.Error("HandleWebhook() should return processed=true")
	}
}

// =====================================================
// HandleWebhook Tests - invoice.payment_failed
// =====================================================

func TestStripeWebhookUsecase_InvoicePaymentFailed_Success(t *testing.T) {
	testTenant := createTestActiveTenant(t)
	testSubscription := createTestWebhookSubscription(t, testTenant.TenantID())

	var savedTenant *tenant.Tenant

	tenantRepo := &MockWebhookTenantRepository{
		findByIDFunc: func(ctx context.Context, tenantID common.TenantID) (*tenant.Tenant, error) {
			return testTenant, nil
		},
		saveFunc: func(ctx context.Context, t *tenant.Tenant) error {
			savedTenant = t
			return nil
		},
	}

	subscriptionRepo := &MockWebhookSubscriptionRepository{
		findByStripeSubscriptionIDFunc: func(ctx context.Context, stripeSubID string) (*billing.Subscription, error) {
			if stripeSubID == "sub_test123" {
				return testSubscription, nil
			}
			return nil, nil
		},
		saveFunc: func(ctx context.Context, sub *billing.Subscription) error {
			return nil
		},
	}

	webhookEventRepo := &MockWebhookEventRepository{
		tryInsertFunc: func(ctx context.Context, provider string, eventID string, payloadJSON *string) (bool, error) {
			return true, nil
		},
	}

	auditLogRepo := &MockWebhookBillingAuditLogRepository{
		saveFunc: func(ctx context.Context, log *billing.BillingAuditLog) error {
			return nil
		},
	}

	usecase := createWebhookUsecase(nil, tenantRepo, subscriptionRepo, nil, webhookEventRepo, auditLogRepo)

	invoiceData := StripeInvoice{
		ID:           "in_test123",
		Customer:     "cus_test123",
		Subscription: "sub_test123",
		Status:       "open",
		Paid:         false,
	}
	invoiceJSON, _ := json.Marshal(invoiceData)

	event := StripeEvent{
		ID:   "evt_test123",
		Type: "invoice.payment_failed",
		Data: StripeEventData{Object: invoiceJSON},
	}

	processed, err := usecase.HandleWebhook(context.Background(), event, string(invoiceJSON))

	if err != nil {
		t.Fatalf("HandleWebhook() should succeed, got error: %v", err)
	}

	if !processed {
		t.Error("HandleWebhook() should return processed=true")
	}

	if savedTenant == nil {
		t.Error("Tenant should have been saved")
	} else if savedTenant.Status() != tenant.TenantStatusGrace {
		t.Errorf("Tenant status should be grace, got %s", savedTenant.Status())
	}
}

// =====================================================
// HandleWebhook Tests - customer.subscription.deleted
// =====================================================

func TestStripeWebhookUsecase_SubscriptionDeleted_PeriodEnded(t *testing.T) {
	testTenant := createTestActiveTenant(t)
	testSubscription := createTestWebhookSubscription(t, testTenant.TenantID())

	var savedTenant *tenant.Tenant

	tenantRepo := &MockWebhookTenantRepository{
		findByIDFunc: func(ctx context.Context, tenantID common.TenantID) (*tenant.Tenant, error) {
			return testTenant, nil
		},
		saveFunc: func(ctx context.Context, t *tenant.Tenant) error {
			savedTenant = t
			return nil
		},
	}

	subscriptionRepo := &MockWebhookSubscriptionRepository{
		findByStripeSubscriptionIDFunc: func(ctx context.Context, stripeSubID string) (*billing.Subscription, error) {
			if stripeSubID == "sub_test123" {
				return testSubscription, nil
			}
			return nil, nil
		},
		saveFunc: func(ctx context.Context, sub *billing.Subscription) error {
			return nil
		},
	}

	webhookEventRepo := &MockWebhookEventRepository{
		tryInsertFunc: func(ctx context.Context, provider string, eventID string, payloadJSON *string) (bool, error) {
			return true, nil
		},
	}

	auditLogRepo := &MockWebhookBillingAuditLogRepository{
		saveFunc: func(ctx context.Context, log *billing.BillingAuditLog) error {
			return nil
		},
	}

	usecase := createWebhookUsecase(nil, tenantRepo, subscriptionRepo, nil, webhookEventRepo, auditLogRepo)

	// Period ended in the past
	subscriptionData := StripeSubscription{
		ID:               "sub_test123",
		Customer:         "cus_test123",
		Status:           "canceled",
		CurrentPeriodEnd: time.Now().Add(-1 * time.Hour).Unix(), // Ended 1 hour ago
	}
	subscriptionJSON, _ := json.Marshal(subscriptionData)

	event := StripeEvent{
		ID:   "evt_test123",
		Type: "customer.subscription.deleted",
		Data: StripeEventData{Object: subscriptionJSON},
	}

	processed, err := usecase.HandleWebhook(context.Background(), event, string(subscriptionJSON))

	if err != nil {
		t.Fatalf("HandleWebhook() should succeed, got error: %v", err)
	}

	if !processed {
		t.Error("HandleWebhook() should return processed=true")
	}

	if savedTenant == nil {
		t.Error("Tenant should have been saved")
	} else if savedTenant.Status() != tenant.TenantStatusSuspended {
		t.Errorf("Tenant status should be suspended when period ended, got %s", savedTenant.Status())
	}
}

func TestStripeWebhookUsecase_SubscriptionDeleted_PeriodNotEnded(t *testing.T) {
	testTenant := createTestActiveTenant(t)
	testSubscription := createTestWebhookSubscription(t, testTenant.TenantID())

	var savedTenant *tenant.Tenant

	tenantRepo := &MockWebhookTenantRepository{
		findByIDFunc: func(ctx context.Context, tenantID common.TenantID) (*tenant.Tenant, error) {
			return testTenant, nil
		},
		saveFunc: func(ctx context.Context, t *tenant.Tenant) error {
			savedTenant = t
			return nil
		},
	}

	subscriptionRepo := &MockWebhookSubscriptionRepository{
		findByStripeSubscriptionIDFunc: func(ctx context.Context, stripeSubID string) (*billing.Subscription, error) {
			if stripeSubID == "sub_test123" {
				return testSubscription, nil
			}
			return nil, nil
		},
		saveFunc: func(ctx context.Context, sub *billing.Subscription) error {
			return nil
		},
	}

	webhookEventRepo := &MockWebhookEventRepository{
		tryInsertFunc: func(ctx context.Context, provider string, eventID string, payloadJSON *string) (bool, error) {
			return true, nil
		},
	}

	auditLogRepo := &MockWebhookBillingAuditLogRepository{
		saveFunc: func(ctx context.Context, log *billing.BillingAuditLog) error {
			return nil
		},
	}

	usecase := createWebhookUsecase(nil, tenantRepo, subscriptionRepo, nil, webhookEventRepo, auditLogRepo)

	// Period ends in the future
	subscriptionData := StripeSubscription{
		ID:               "sub_test123",
		Customer:         "cus_test123",
		Status:           "canceled",
		CurrentPeriodEnd: time.Now().Add(7 * 24 * time.Hour).Unix(), // Ends in 7 days
	}
	subscriptionJSON, _ := json.Marshal(subscriptionData)

	event := StripeEvent{
		ID:   "evt_test123",
		Type: "customer.subscription.deleted",
		Data: StripeEventData{Object: subscriptionJSON},
	}

	processed, err := usecase.HandleWebhook(context.Background(), event, string(subscriptionJSON))

	if err != nil {
		t.Fatalf("HandleWebhook() should succeed, got error: %v", err)
	}

	if !processed {
		t.Error("HandleWebhook() should return processed=true")
	}

	if savedTenant == nil {
		t.Error("Tenant should have been saved")
	} else if savedTenant.Status() != tenant.TenantStatusGrace {
		t.Errorf("Tenant status should be grace when period not ended, got %s", savedTenant.Status())
	}
}
