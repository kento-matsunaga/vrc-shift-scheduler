package tenant

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/auth"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/billing"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
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
	findByIDFunc                    func(ctx context.Context, tenantID common.TenantID) (*tenant.Tenant, error)
	findByPendingStripeSessionIDFunc func(ctx context.Context, sessionID string) (*tenant.Tenant, error)
	saveFunc                        func(ctx context.Context, t *tenant.Tenant) error
	listAllFunc                     func(ctx context.Context, status *tenant.TenantStatus, limit, offset int) ([]*tenant.Tenant, int, error)
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

func (m *MockTenantRepository) Save(ctx context.Context, t *tenant.Tenant) error {
	if m.saveFunc != nil {
		return m.saveFunc(ctx, t)
	}
	return nil
}

func (m *MockTenantRepository) ListAll(ctx context.Context, status *tenant.TenantStatus, limit, offset int) ([]*tenant.Tenant, int, error) {
	if m.listAllFunc != nil {
		return m.listAllFunc(ctx, status, limit, offset)
	}
	return nil, 0, errors.New("not implemented")
}

// MockAdminRepository is a mock implementation of auth.AdminRepository
type MockAdminRepository struct {
	findByTenantIDFunc func(ctx context.Context, tenantID common.TenantID) ([]*auth.Admin, error)
}

func (m *MockAdminRepository) FindByTenantID(ctx context.Context, tenantID common.TenantID) ([]*auth.Admin, error) {
	if m.findByTenantIDFunc != nil {
		return m.findByTenantIDFunc(ctx, tenantID)
	}
	return nil, errors.New("not implemented")
}

// Unused methods - just satisfy the interface
func (m *MockAdminRepository) Save(ctx context.Context, admin *auth.Admin) error {
	return errors.New("not implemented")
}
func (m *MockAdminRepository) FindByID(ctx context.Context, adminID common.AdminID) (*auth.Admin, error) {
	return nil, errors.New("not implemented")
}
func (m *MockAdminRepository) FindByEmail(ctx context.Context, tenantID common.TenantID, email string) (*auth.Admin, error) {
	return nil, errors.New("not implemented")
}
func (m *MockAdminRepository) FindByEmailGlobal(ctx context.Context, email string) (*auth.Admin, error) {
	return nil, errors.New("not implemented")
}
func (m *MockAdminRepository) FindByIDWithTenant(ctx context.Context, tenantID common.TenantID, adminID common.AdminID) (*auth.Admin, error) {
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

// MockEntitlementRepository is a mock implementation of billing.EntitlementRepository
type MockEntitlementRepository struct {
	findByTenantIDFunc func(ctx context.Context, tenantID common.TenantID) ([]*billing.Entitlement, error)
}

func (m *MockEntitlementRepository) FindByTenantID(ctx context.Context, tenantID common.TenantID) ([]*billing.Entitlement, error) {
	if m.findByTenantIDFunc != nil {
		return m.findByTenantIDFunc(ctx, tenantID)
	}
	return nil, errors.New("not implemented")
}

// Unused methods
func (m *MockEntitlementRepository) Save(ctx context.Context, entitlement *billing.Entitlement) error {
	return errors.New("not implemented")
}
func (m *MockEntitlementRepository) FindByID(ctx context.Context, entitlementID billing.EntitlementID) (*billing.Entitlement, error) {
	return nil, errors.New("not implemented")
}
func (m *MockEntitlementRepository) FindActiveByTenantID(ctx context.Context, tenantID common.TenantID) (*billing.Entitlement, error) {
	return nil, errors.New("not implemented")
}
func (m *MockEntitlementRepository) HasRevokedByTenantID(ctx context.Context, tenantID common.TenantID) (bool, error) {
	return false, errors.New("not implemented")
}

// MockSubscriptionRepository is a mock implementation of billing.SubscriptionRepository
type MockSubscriptionRepository struct {
	findByTenantIDFunc           func(ctx context.Context, tenantID common.TenantID) (*billing.Subscription, error)
	findByStripeSubscriptionIDFunc func(ctx context.Context, stripeSubID string) (*billing.Subscription, error)
	saveFunc                     func(ctx context.Context, sub *billing.Subscription) error
}

func (m *MockSubscriptionRepository) FindByTenantID(ctx context.Context, tenantID common.TenantID) (*billing.Subscription, error) {
	if m.findByTenantIDFunc != nil {
		return m.findByTenantIDFunc(ctx, tenantID)
	}
	return nil, errors.New("not implemented")
}

func (m *MockSubscriptionRepository) FindByStripeSubscriptionID(ctx context.Context, stripeSubID string) (*billing.Subscription, error) {
	if m.findByStripeSubscriptionIDFunc != nil {
		return m.findByStripeSubscriptionIDFunc(ctx, stripeSubID)
	}
	return nil, errors.New("not implemented")
}

func (m *MockSubscriptionRepository) Save(ctx context.Context, sub *billing.Subscription) error {
	if m.saveFunc != nil {
		return m.saveFunc(ctx, sub)
	}
	return nil
}

// MockBillingAuditLogRepository is a mock implementation of billing.BillingAuditLogRepository
type MockBillingAuditLogRepository struct {
	saveFunc func(ctx context.Context, log *billing.BillingAuditLog) error
}

func (m *MockBillingAuditLogRepository) Save(ctx context.Context, log *billing.BillingAuditLog) error {
	if m.saveFunc != nil {
		return m.saveFunc(ctx, log)
	}
	return nil
}

// Unused methods
func (m *MockBillingAuditLogRepository) FindByDateRange(ctx context.Context, startDate, endDate string, limit, offset int) ([]*billing.BillingAuditLog, error) {
	return nil, errors.New("not implemented")
}
func (m *MockBillingAuditLogRepository) FindByAction(ctx context.Context, action string, limit, offset int) ([]*billing.BillingAuditLog, error) {
	return nil, errors.New("not implemented")
}
func (m *MockBillingAuditLogRepository) CountByDateRange(ctx context.Context, startDate, endDate string) (int, error) {
	return 0, errors.New("not implemented")
}
func (m *MockBillingAuditLogRepository) List(ctx context.Context, action *string, limit, offset int) ([]*billing.BillingAuditLog, int, error) {
	return nil, 0, errors.New("not implemented")
}

// =====================================================
// Test Helper Functions
// =====================================================

func createTestTenant(t *testing.T) *tenant.Tenant {
	t.Helper()
	now := time.Now()

	testTenant, err := tenant.NewTenant(now, "Test Tenant", "Asia/Tokyo")
	if err != nil {
		t.Fatalf("Failed to create test tenant: %v", err)
	}
	return testTenant
}

func createTestAdmin(t *testing.T, tenantID common.TenantID) *auth.Admin {
	t.Helper()
	now := time.Now()

	admin, err := auth.NewAdmin(now, tenantID, "test@example.com", "hashedpassword", "Test Admin", auth.RoleOwner)
	if err != nil {
		t.Fatalf("Failed to create test admin: %v", err)
	}
	return admin
}

func createTestEntitlement(t *testing.T, tenantID common.TenantID) *billing.Entitlement {
	t.Helper()
	now := time.Now()

	entitlement, err := billing.NewEntitlement(now, tenantID, "SUB_200", billing.EntitlementSourceStripe, nil)
	if err != nil {
		t.Fatalf("Failed to create test entitlement: %v", err)
	}
	return entitlement
}

func createTestSubscription(t *testing.T, tenantID common.TenantID) *billing.Subscription {
	t.Helper()
	now := time.Now()
	periodEnd := now.Add(30 * 24 * time.Hour)

	sub, err := billing.NewSubscription(now, tenantID, "cus_test123", "sub_test123", billing.SubscriptionStatusActive, &periodEnd)
	if err != nil {
		t.Fatalf("Failed to create test subscription: %v", err)
	}
	return sub
}

func createTestAdminTenantUsecase(
	txManager *MockTxManager,
	tenantRepo *MockTenantRepository,
	adminRepo *MockAdminRepository,
	entitlementRepo *MockEntitlementRepository,
	subscriptionRepo *MockSubscriptionRepository,
	auditLogRepo *MockBillingAuditLogRepository,
) *AdminTenantUsecase {
	if txManager == nil {
		txManager = &MockTxManager{}
	}
	if tenantRepo == nil {
		tenantRepo = &MockTenantRepository{}
	}
	if adminRepo == nil {
		adminRepo = &MockAdminRepository{}
	}
	if entitlementRepo == nil {
		entitlementRepo = &MockEntitlementRepository{}
	}
	if subscriptionRepo == nil {
		subscriptionRepo = &MockSubscriptionRepository{}
	}
	if auditLogRepo == nil {
		auditLogRepo = &MockBillingAuditLogRepository{}
	}

	return NewAdminTenantUsecase(
		txManager,
		tenantRepo,
		adminRepo,
		entitlementRepo,
		subscriptionRepo,
		auditLogRepo,
	)
}

// =====================================================
// AdminTenantUsecase.List Tests
// =====================================================

func TestAdminTenantUsecase_List_Success(t *testing.T) {
	testTenant := createTestTenant(t)

	tenantRepo := &MockTenantRepository{
		listAllFunc: func(ctx context.Context, status *tenant.TenantStatus, limit, offset int) ([]*tenant.Tenant, int, error) {
			return []*tenant.Tenant{testTenant}, 1, nil
		},
	}

	usecase := createTestAdminTenantUsecase(nil, tenantRepo, nil, nil, nil, nil)

	input := TenantListInput{
		Limit:  50,
		Offset: 0,
	}

	output, err := usecase.List(context.Background(), input)

	if err != nil {
		t.Fatalf("List() should succeed, got error: %v", err)
	}

	if len(output.Tenants) != 1 {
		t.Errorf("Expected 1 tenant, got %d", len(output.Tenants))
	}

	if output.TotalCount != 1 {
		t.Errorf("Expected total count 1, got %d", output.TotalCount)
	}
}

func TestAdminTenantUsecase_List_WithStatusFilter(t *testing.T) {
	testTenant := createTestTenant(t)
	activeStatus := tenant.TenantStatusActive

	tenantRepo := &MockTenantRepository{
		listAllFunc: func(ctx context.Context, status *tenant.TenantStatus, limit, offset int) ([]*tenant.Tenant, int, error) {
			if status == nil || *status != activeStatus {
				t.Error("Expected active status filter")
			}
			return []*tenant.Tenant{testTenant}, 1, nil
		},
	}

	usecase := createTestAdminTenantUsecase(nil, tenantRepo, nil, nil, nil, nil)

	input := TenantListInput{
		Status: &activeStatus,
		Limit:  50,
		Offset: 0,
	}

	output, err := usecase.List(context.Background(), input)

	if err != nil {
		t.Fatalf("List() should succeed, got error: %v", err)
	}

	if len(output.Tenants) != 1 {
		t.Errorf("Expected 1 tenant, got %d", len(output.Tenants))
	}
}

func TestAdminTenantUsecase_List_LimitCapped(t *testing.T) {
	tenantRepo := &MockTenantRepository{
		listAllFunc: func(ctx context.Context, status *tenant.TenantStatus, limit, offset int) ([]*tenant.Tenant, int, error) {
			if limit > 100 {
				t.Errorf("Limit should be capped at 100, got %d", limit)
			}
			return []*tenant.Tenant{}, 0, nil
		},
	}

	usecase := createTestAdminTenantUsecase(nil, tenantRepo, nil, nil, nil, nil)

	input := TenantListInput{
		Limit: 200, // Should be capped to 100
	}

	_, err := usecase.List(context.Background(), input)

	if err != nil {
		t.Fatalf("List() should succeed, got error: %v", err)
	}
}

func TestAdminTenantUsecase_List_ErrorWhenRepoFails(t *testing.T) {
	dbError := errors.New("database error")

	tenantRepo := &MockTenantRepository{
		listAllFunc: func(ctx context.Context, status *tenant.TenantStatus, limit, offset int) ([]*tenant.Tenant, int, error) {
			return nil, 0, dbError
		},
	}

	usecase := createTestAdminTenantUsecase(nil, tenantRepo, nil, nil, nil, nil)

	input := TenantListInput{
		Limit: 50,
	}

	_, err := usecase.List(context.Background(), input)

	if err == nil {
		t.Fatal("List() should fail when repo fails")
	}
}

// =====================================================
// AdminTenantUsecase.GetDetail Tests
// =====================================================

func TestAdminTenantUsecase_GetDetail_Success(t *testing.T) {
	testTenant := createTestTenant(t)
	testAdmin := createTestAdmin(t, testTenant.TenantID())
	testEntitlement := createTestEntitlement(t, testTenant.TenantID())
	testSubscription := createTestSubscription(t, testTenant.TenantID())

	tenantRepo := &MockTenantRepository{
		findByIDFunc: func(ctx context.Context, tenantID common.TenantID) (*tenant.Tenant, error) {
			return testTenant, nil
		},
	}

	adminRepo := &MockAdminRepository{
		findByTenantIDFunc: func(ctx context.Context, tenantID common.TenantID) ([]*auth.Admin, error) {
			return []*auth.Admin{testAdmin}, nil
		},
	}

	entitlementRepo := &MockEntitlementRepository{
		findByTenantIDFunc: func(ctx context.Context, tenantID common.TenantID) ([]*billing.Entitlement, error) {
			return []*billing.Entitlement{testEntitlement}, nil
		},
	}

	subscriptionRepo := &MockSubscriptionRepository{
		findByTenantIDFunc: func(ctx context.Context, tenantID common.TenantID) (*billing.Subscription, error) {
			return testSubscription, nil
		},
	}

	usecase := createTestAdminTenantUsecase(nil, tenantRepo, adminRepo, entitlementRepo, subscriptionRepo, nil)

	output, err := usecase.GetDetail(context.Background(), testTenant.TenantID())

	if err != nil {
		t.Fatalf("GetDetail() should succeed, got error: %v", err)
	}

	if output.TenantID != testTenant.TenantID() {
		t.Errorf("TenantID mismatch: got %v, want %v", output.TenantID, testTenant.TenantID())
	}

	if len(output.Admins) != 1 {
		t.Errorf("Expected 1 admin, got %d", len(output.Admins))
	}

	if len(output.Entitlements) != 1 {
		t.Errorf("Expected 1 entitlement, got %d", len(output.Entitlements))
	}

	if output.Subscription == nil {
		t.Error("Expected subscription to be set")
	} else {
		if output.Subscription.StripeCustomerID != "cus_test123" {
			t.Errorf("StripeCustomerID mismatch: got %v", output.Subscription.StripeCustomerID)
		}
	}
}

func TestAdminTenantUsecase_GetDetail_WithoutSubscription(t *testing.T) {
	testTenant := createTestTenant(t)
	testAdmin := createTestAdmin(t, testTenant.TenantID())
	testEntitlement := createTestEntitlement(t, testTenant.TenantID())

	tenantRepo := &MockTenantRepository{
		findByIDFunc: func(ctx context.Context, tenantID common.TenantID) (*tenant.Tenant, error) {
			return testTenant, nil
		},
	}

	adminRepo := &MockAdminRepository{
		findByTenantIDFunc: func(ctx context.Context, tenantID common.TenantID) ([]*auth.Admin, error) {
			return []*auth.Admin{testAdmin}, nil
		},
	}

	entitlementRepo := &MockEntitlementRepository{
		findByTenantIDFunc: func(ctx context.Context, tenantID common.TenantID) ([]*billing.Entitlement, error) {
			return []*billing.Entitlement{testEntitlement}, nil
		},
	}

	subscriptionRepo := &MockSubscriptionRepository{
		findByTenantIDFunc: func(ctx context.Context, tenantID common.TenantID) (*billing.Subscription, error) {
			return nil, common.NewNotFoundError("subscription", tenantID.String())
		},
	}

	usecase := createTestAdminTenantUsecase(nil, tenantRepo, adminRepo, entitlementRepo, subscriptionRepo, nil)

	output, err := usecase.GetDetail(context.Background(), testTenant.TenantID())

	if err != nil {
		t.Fatalf("GetDetail() should succeed even without subscription, got error: %v", err)
	}

	if output.Subscription != nil {
		t.Error("Expected subscription to be nil for tenant without subscription")
	}
}

func TestAdminTenantUsecase_GetDetail_TenantNotFound(t *testing.T) {
	tenantID := common.NewTenantID()

	tenantRepo := &MockTenantRepository{
		findByIDFunc: func(ctx context.Context, tid common.TenantID) (*tenant.Tenant, error) {
			return nil, common.NewNotFoundError("tenant", tid.String())
		},
	}

	usecase := createTestAdminTenantUsecase(nil, tenantRepo, nil, nil, nil, nil)

	_, err := usecase.GetDetail(context.Background(), tenantID)

	if err == nil {
		t.Fatal("GetDetail() should fail when tenant not found")
	}

	domainErr, ok := err.(*common.DomainError)
	if !ok || domainErr.Code() != common.ErrNotFound {
		t.Errorf("Expected ErrNotFound, got %v", err)
	}
}

func TestAdminTenantUsecase_GetDetail_ErrorWhenEntitlementRepoFails(t *testing.T) {
	testTenant := createTestTenant(t)
	dbError := errors.New("database error")

	tenantRepo := &MockTenantRepository{
		findByIDFunc: func(ctx context.Context, tenantID common.TenantID) (*tenant.Tenant, error) {
			return testTenant, nil
		},
	}

	entitlementRepo := &MockEntitlementRepository{
		findByTenantIDFunc: func(ctx context.Context, tenantID common.TenantID) ([]*billing.Entitlement, error) {
			return nil, dbError
		},
	}

	usecase := createTestAdminTenantUsecase(nil, tenantRepo, nil, entitlementRepo, nil, nil)

	_, err := usecase.GetDetail(context.Background(), testTenant.TenantID())

	if err == nil {
		t.Fatal("GetDetail() should fail when entitlement repo fails")
	}
}

func TestAdminTenantUsecase_GetDetail_ErrorWhenSubscriptionRepoFails(t *testing.T) {
	testTenant := createTestTenant(t)
	testEntitlement := createTestEntitlement(t, testTenant.TenantID())
	dbError := errors.New("database error")

	tenantRepo := &MockTenantRepository{
		findByIDFunc: func(ctx context.Context, tenantID common.TenantID) (*tenant.Tenant, error) {
			return testTenant, nil
		},
	}

	entitlementRepo := &MockEntitlementRepository{
		findByTenantIDFunc: func(ctx context.Context, tenantID common.TenantID) ([]*billing.Entitlement, error) {
			return []*billing.Entitlement{testEntitlement}, nil
		},
	}

	subscriptionRepo := &MockSubscriptionRepository{
		findByTenantIDFunc: func(ctx context.Context, tenantID common.TenantID) (*billing.Subscription, error) {
			return nil, dbError
		},
	}

	usecase := createTestAdminTenantUsecase(nil, tenantRepo, nil, entitlementRepo, subscriptionRepo, nil)

	_, err := usecase.GetDetail(context.Background(), testTenant.TenantID())

	if err == nil {
		t.Fatal("GetDetail() should fail when subscription repo fails with non-NotFound error")
	}
}

// =====================================================
// AdminTenantUsecase.UpdateStatus Tests
// =====================================================

func TestAdminTenantUsecase_UpdateStatus_ToActive(t *testing.T) {
	testTenant := createTestTenant(t)
	adminID := common.NewAdminID()

	var savedTenant *tenant.Tenant
	tenantRepo := &MockTenantRepository{
		findByIDFunc: func(ctx context.Context, tenantID common.TenantID) (*tenant.Tenant, error) {
			return testTenant, nil
		},
		saveFunc: func(ctx context.Context, t *tenant.Tenant) error {
			savedTenant = t
			return nil
		},
	}

	auditLogRepo := &MockBillingAuditLogRepository{
		saveFunc: func(ctx context.Context, log *billing.BillingAuditLog) error {
			return nil
		},
	}

	usecase := createTestAdminTenantUsecase(nil, tenantRepo, nil, nil, nil, auditLogRepo)

	input := UpdateTenantStatusInput{
		TenantID: testTenant.TenantID(),
		Status:   tenant.TenantStatusActive,
		AdminID:  adminID,
	}

	err := usecase.UpdateStatus(context.Background(), input)

	if err != nil {
		t.Fatalf("UpdateStatus() should succeed, got error: %v", err)
	}

	if savedTenant == nil {
		t.Fatal("Tenant should have been saved")
	}

	if savedTenant.Status() != tenant.TenantStatusActive {
		t.Errorf("Expected status active, got %s", savedTenant.Status())
	}
}

func TestAdminTenantUsecase_UpdateStatus_ToGrace(t *testing.T) {
	testTenant := createTestTenant(t)
	adminID := common.NewAdminID()
	graceUntil := time.Now().Add(14 * 24 * time.Hour)

	var savedTenant *tenant.Tenant
	tenantRepo := &MockTenantRepository{
		findByIDFunc: func(ctx context.Context, tenantID common.TenantID) (*tenant.Tenant, error) {
			return testTenant, nil
		},
		saveFunc: func(ctx context.Context, t *tenant.Tenant) error {
			savedTenant = t
			return nil
		},
	}

	auditLogRepo := &MockBillingAuditLogRepository{
		saveFunc: func(ctx context.Context, log *billing.BillingAuditLog) error {
			return nil
		},
	}

	usecase := createTestAdminTenantUsecase(nil, tenantRepo, nil, nil, nil, auditLogRepo)

	input := UpdateTenantStatusInput{
		TenantID:   testTenant.TenantID(),
		Status:     tenant.TenantStatusGrace,
		GraceUntil: &graceUntil,
		AdminID:    adminID,
	}

	err := usecase.UpdateStatus(context.Background(), input)

	if err != nil {
		t.Fatalf("UpdateStatus() should succeed, got error: %v", err)
	}

	if savedTenant.Status() != tenant.TenantStatusGrace {
		t.Errorf("Expected status grace, got %s", savedTenant.Status())
	}
}

func TestAdminTenantUsecase_UpdateStatus_ToGraceWithoutGraceUntil(t *testing.T) {
	testTenant := createTestTenant(t)
	adminID := common.NewAdminID()

	tenantRepo := &MockTenantRepository{
		findByIDFunc: func(ctx context.Context, tenantID common.TenantID) (*tenant.Tenant, error) {
			return testTenant, nil
		},
	}

	usecase := createTestAdminTenantUsecase(nil, tenantRepo, nil, nil, nil, nil)

	input := UpdateTenantStatusInput{
		TenantID:   testTenant.TenantID(),
		Status:     tenant.TenantStatusGrace,
		GraceUntil: nil, // Missing required field
		AdminID:    adminID,
	}

	err := usecase.UpdateStatus(context.Background(), input)

	if err == nil {
		t.Fatal("UpdateStatus() should fail when grace_until is missing for grace status")
	}
}

func TestAdminTenantUsecase_UpdateStatus_ToSuspended(t *testing.T) {
	testTenant := createTestTenant(t)
	adminID := common.NewAdminID()

	var savedTenant *tenant.Tenant
	tenantRepo := &MockTenantRepository{
		findByIDFunc: func(ctx context.Context, tenantID common.TenantID) (*tenant.Tenant, error) {
			return testTenant, nil
		},
		saveFunc: func(ctx context.Context, t *tenant.Tenant) error {
			savedTenant = t
			return nil
		},
	}

	auditLogRepo := &MockBillingAuditLogRepository{
		saveFunc: func(ctx context.Context, log *billing.BillingAuditLog) error {
			return nil
		},
	}

	usecase := createTestAdminTenantUsecase(nil, tenantRepo, nil, nil, nil, auditLogRepo)

	input := UpdateTenantStatusInput{
		TenantID: testTenant.TenantID(),
		Status:   tenant.TenantStatusSuspended,
		AdminID:  adminID,
	}

	err := usecase.UpdateStatus(context.Background(), input)

	if err != nil {
		t.Fatalf("UpdateStatus() should succeed, got error: %v", err)
	}

	if savedTenant.Status() != tenant.TenantStatusSuspended {
		t.Errorf("Expected status suspended, got %s", savedTenant.Status())
	}
}

func TestAdminTenantUsecase_UpdateStatus_TenantNotFound(t *testing.T) {
	tenantID := common.NewTenantID()
	adminID := common.NewAdminID()

	tenantRepo := &MockTenantRepository{
		findByIDFunc: func(ctx context.Context, tid common.TenantID) (*tenant.Tenant, error) {
			return nil, common.NewNotFoundError("tenant", tid.String())
		},
	}

	usecase := createTestAdminTenantUsecase(nil, tenantRepo, nil, nil, nil, nil)

	input := UpdateTenantStatusInput{
		TenantID: tenantID,
		Status:   tenant.TenantStatusActive,
		AdminID:  adminID,
	}

	err := usecase.UpdateStatus(context.Background(), input)

	if err == nil {
		t.Fatal("UpdateStatus() should fail when tenant not found")
	}
}

func TestAdminTenantUsecase_UpdateStatus_InvalidStatus(t *testing.T) {
	testTenant := createTestTenant(t)
	adminID := common.NewAdminID()

	tenantRepo := &MockTenantRepository{
		findByIDFunc: func(ctx context.Context, tenantID common.TenantID) (*tenant.Tenant, error) {
			return testTenant, nil
		},
	}

	usecase := createTestAdminTenantUsecase(nil, tenantRepo, nil, nil, nil, nil)

	input := UpdateTenantStatusInput{
		TenantID: testTenant.TenantID(),
		Status:   tenant.TenantStatus("invalid_status"),
		AdminID:  adminID,
	}

	err := usecase.UpdateStatus(context.Background(), input)

	if err == nil {
		t.Fatal("UpdateStatus() should fail with invalid status")
	}
}
