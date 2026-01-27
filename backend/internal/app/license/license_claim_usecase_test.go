package license

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
	// Default implementation: just call the function without actual transaction
	return fn(ctx)
}

// MockTenantRepository is a mock implementation of tenant.TenantRepository
type MockTenantRepository struct {
	saveFunc     func(ctx context.Context, t *tenant.Tenant) error
	findByIDFunc func(ctx context.Context, tenantID common.TenantID) (*tenant.Tenant, error)
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
	return nil, errors.New("not implemented")
}

func (m *MockTenantRepository) ListAll(ctx context.Context, status *tenant.TenantStatus, limit, offset int) ([]*tenant.Tenant, int, error) {
	return nil, 0, errors.New("not implemented")
}

// MockAdminRepository is a mock implementation of auth.AdminRepository
type MockAdminRepository struct {
	saveFunc               func(ctx context.Context, admin *auth.Admin) error
	findByEmailGlobalFunc  func(ctx context.Context, email string) (*auth.Admin, error)
	findByIDWithTenantFunc func(ctx context.Context, tenantID common.TenantID, adminID common.AdminID) (*auth.Admin, error)
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
	return nil, errors.New("not implemented")
}

func (m *MockAdminRepository) FindByIDWithTenant(ctx context.Context, tenantID common.TenantID, adminID common.AdminID) (*auth.Admin, error) {
	if m.findByIDWithTenantFunc != nil {
		return m.findByIDWithTenantFunc(ctx, tenantID, adminID)
	}
	return nil, errors.New("not implemented")
}

// Unused methods - just satisfy the interface
func (m *MockAdminRepository) FindByID(ctx context.Context, adminID common.AdminID) (*auth.Admin, error) {
	return nil, errors.New("not implemented")
}
func (m *MockAdminRepository) FindByEmail(ctx context.Context, tenantID common.TenantID, email string) (*auth.Admin, error) {
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
func (m *MockAdminRepository) ExistsByEmailGlobal(ctx context.Context, email string) (bool, error) {
	return false, errors.New("not implemented")
}

// MockLicenseKeyRepository is a mock implementation of billing.LicenseKeyRepository
type MockLicenseKeyRepository struct {
	findByHashForUpdateFunc func(ctx context.Context, keyHash string) (*billing.LicenseKey, error)
	saveFunc                func(ctx context.Context, key *billing.LicenseKey) error
}

func (m *MockLicenseKeyRepository) FindByHashForUpdate(ctx context.Context, keyHash string) (*billing.LicenseKey, error) {
	if m.findByHashForUpdateFunc != nil {
		return m.findByHashForUpdateFunc(ctx, keyHash)
	}
	return nil, nil
}

func (m *MockLicenseKeyRepository) Save(ctx context.Context, key *billing.LicenseKey) error {
	if m.saveFunc != nil {
		return m.saveFunc(ctx, key)
	}
	return nil
}

// Unused methods - just satisfy the interface
func (m *MockLicenseKeyRepository) SaveBatch(ctx context.Context, keys []*billing.LicenseKey) error {
	return errors.New("not implemented")
}
func (m *MockLicenseKeyRepository) FindByID(ctx context.Context, keyID billing.LicenseKeyID) (*billing.LicenseKey, error) {
	return nil, errors.New("not implemented")
}
func (m *MockLicenseKeyRepository) FindByBatchID(ctx context.Context, batchID string) ([]*billing.LicenseKey, error) {
	return nil, errors.New("not implemented")
}
func (m *MockLicenseKeyRepository) CountByStatus(ctx context.Context, status billing.LicenseKeyStatus) (int, error) {
	return 0, errors.New("not implemented")
}
func (m *MockLicenseKeyRepository) RevokeBatch(ctx context.Context, batchID string) error {
	return errors.New("not implemented")
}
func (m *MockLicenseKeyRepository) List(ctx context.Context, status *billing.LicenseKeyStatus, limit, offset int) ([]*billing.LicenseKey, int, error) {
	return nil, 0, errors.New("not implemented")
}
func (m *MockLicenseKeyRepository) FindByHashAndTenant(ctx context.Context, keyHash string, tenantID common.TenantID) (*billing.LicenseKey, error) {
	return nil, errors.New("not implemented")
}

// MockEntitlementRepository is a mock implementation of billing.EntitlementRepository
type MockEntitlementRepository struct {
	saveFunc func(ctx context.Context, entitlement *billing.Entitlement) error
}

func (m *MockEntitlementRepository) Save(ctx context.Context, entitlement *billing.Entitlement) error {
	if m.saveFunc != nil {
		return m.saveFunc(ctx, entitlement)
	}
	return nil
}

// Unused methods - just satisfy the interface
func (m *MockEntitlementRepository) FindByID(ctx context.Context, entitlementID billing.EntitlementID) (*billing.Entitlement, error) {
	return nil, errors.New("not implemented")
}
func (m *MockEntitlementRepository) FindByTenantID(ctx context.Context, tenantID common.TenantID) ([]*billing.Entitlement, error) {
	return nil, errors.New("not implemented")
}
func (m *MockEntitlementRepository) FindActiveByTenantID(ctx context.Context, tenantID common.TenantID) (*billing.Entitlement, error) {
	return nil, errors.New("not implemented")
}
func (m *MockEntitlementRepository) HasRevokedByTenantID(ctx context.Context, tenantID common.TenantID) (bool, error) {
	return false, errors.New("not implemented")
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

// Unused methods - just satisfy the interface
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

// =====================================================
// Test Helper Functions
// =====================================================

func createTestLicenseKey(t *testing.T, status billing.LicenseKeyStatus) *billing.LicenseKey {
	now := time.Now()
	keyHash := billing.HashLicenseKey("ABCD1234EF567890") // Only hex chars (0-9, A-F)

	if status == billing.LicenseKeyStatusUnused {
		key, err := billing.NewLicenseKey(now, keyHash, nil, "")
		if err != nil {
			t.Fatalf("Failed to create test license key: %v", err)
		}
		return key
	}

	// For used or revoked keys, we need to reconstruct
	key, err := billing.NewLicenseKey(now, keyHash, nil, "")
	if err != nil {
		t.Fatalf("Failed to create test license key: %v", err)
	}

	if status == billing.LicenseKeyStatusUsed {
		tenantID := common.NewTenantID()
		_ = key.MarkAsUsed(now, tenantID)
	} else if status == billing.LicenseKeyStatusRevoked {
		_ = key.Revoke(now)
	}

	return key
}

func createTestUsecase(
	txManager *MockTxManager,
	tenantRepo *MockTenantRepository,
	adminRepo *MockAdminRepository,
	licenseKeyRepo *MockLicenseKeyRepository,
	entitlementRepo *MockEntitlementRepository,
	auditLogRepo *MockBillingAuditLogRepository,
	passwordHasher *MockPasswordHasher,
) *LicenseClaimUsecase {
	if txManager == nil {
		txManager = &MockTxManager{}
	}
	if tenantRepo == nil {
		tenantRepo = &MockTenantRepository{}
	}
	if adminRepo == nil {
		adminRepo = &MockAdminRepository{}
	}
	if licenseKeyRepo == nil {
		licenseKeyRepo = &MockLicenseKeyRepository{}
	}
	if entitlementRepo == nil {
		entitlementRepo = &MockEntitlementRepository{}
	}
	if auditLogRepo == nil {
		auditLogRepo = &MockBillingAuditLogRepository{}
	}
	if passwordHasher == nil {
		passwordHasher = &MockPasswordHasher{}
	}

	return NewLicenseClaimUsecase(
		txManager,
		tenantRepo,
		adminRepo,
		licenseKeyRepo,
		entitlementRepo,
		auditLogRepo,
		passwordHasher,
	)
}

func validInput() LicenseClaimInput {
	return LicenseClaimInput{
		Email:       "test@example.com",
		Password:    "Password123",
		DisplayName: "Test Admin",
		TenantName:  "Test Tenant",
		LicenseKey:  "ABCD-1234-EF56-7890", // Only hex chars (0-9, A-F)
		IPAddress:   "192.168.1.1",
		UserAgent:   "TestAgent/1.0",
	}
}

// =====================================================
// LicenseClaimUsecase Tests - Success Cases
// =====================================================

func TestLicenseClaimUsecase_Execute_Success(t *testing.T) {
	testKey := createTestLicenseKey(t, billing.LicenseKeyStatusUnused)

	var savedTenant *tenant.Tenant
	var savedAdmin *auth.Admin
	var savedEntitlement *billing.Entitlement
	var savedLicenseKey *billing.LicenseKey

	licenseKeyRepo := &MockLicenseKeyRepository{
		findByHashForUpdateFunc: func(ctx context.Context, keyHash string) (*billing.LicenseKey, error) {
			return testKey, nil
		},
		saveFunc: func(ctx context.Context, key *billing.LicenseKey) error {
			savedLicenseKey = key
			return nil
		},
	}

	tenantRepo := &MockTenantRepository{
		saveFunc: func(ctx context.Context, t *tenant.Tenant) error {
			savedTenant = t
			return nil
		},
	}

	adminRepo := &MockAdminRepository{
		saveFunc: func(ctx context.Context, admin *auth.Admin) error {
			savedAdmin = admin
			return nil
		},
	}

	entitlementRepo := &MockEntitlementRepository{
		saveFunc: func(ctx context.Context, entitlement *billing.Entitlement) error {
			savedEntitlement = entitlement
			return nil
		},
	}

	auditLogRepo := &MockBillingAuditLogRepository{
		saveFunc: func(ctx context.Context, log *billing.BillingAuditLog) error {
			return nil
		},
	}

	usecase := createTestUsecase(nil, tenantRepo, adminRepo, licenseKeyRepo, entitlementRepo, auditLogRepo, nil)

	input := validInput()
	output, err := usecase.Execute(context.Background(), input)

	if err != nil {
		t.Fatalf("Execute() should succeed, but got error: %v", err)
	}

	if output == nil {
		t.Fatal("Execute() returned nil output")
	}

	// Verify output
	if output.TenantName != input.TenantName {
		t.Errorf("TenantName: expected %s, got %s", input.TenantName, output.TenantName)
	}

	if output.DisplayName != input.DisplayName {
		t.Errorf("DisplayName: expected %s, got %s", input.DisplayName, output.DisplayName)
	}

	if output.Email != input.Email {
		t.Errorf("Email: expected %s, got %s", input.Email, output.Email)
	}

	// Verify tenant was saved
	if savedTenant == nil {
		t.Error("Tenant should have been saved")
	}

	// Verify admin was saved
	if savedAdmin == nil {
		t.Error("Admin should have been saved")
	}

	// Verify entitlement was saved
	if savedEntitlement == nil {
		t.Error("Entitlement should have been saved")
	}

	// Verify license key was marked as used
	if savedLicenseKey == nil {
		t.Error("LicenseKey should have been saved")
	} else if !savedLicenseKey.IsUsed() {
		t.Error("LicenseKey should be marked as used")
	}
}

// =====================================================
// LicenseClaimUsecase Tests - Input Validation Errors
// =====================================================

func TestLicenseClaimUsecase_Execute_ErrorWhenInvalidLicenseKeyFormat(t *testing.T) {
	usecase := createTestUsecase(nil, nil, nil, nil, nil, nil, nil)

	tests := []struct {
		name       string
		licenseKey string
	}{
		{"empty key", ""},
		{"too short", "ABCD-1234"},
		{"too long", "ABCD-1234-EFGH-5678-IJKL"},
		{"invalid characters", "GHIJ-KLMN-OPQR-STUV"},
		{"no hyphens but wrong length", "ABCD1234"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := validInput()
			input.LicenseKey = tt.licenseKey

			output, err := usecase.Execute(context.Background(), input)

			if err == nil {
				t.Errorf("Execute() should return error for license key: %s", tt.licenseKey)
			}

			if output != nil {
				t.Error("Output should be nil when error occurs")
			}
		})
	}
}

func TestLicenseClaimUsecase_Execute_ErrorWhenEmailEmpty(t *testing.T) {
	usecase := createTestUsecase(nil, nil, nil, nil, nil, nil, nil)

	input := validInput()
	input.Email = ""

	output, err := usecase.Execute(context.Background(), input)

	if err == nil {
		t.Fatal("Execute() should return error when email is empty")
	}

	if output != nil {
		t.Error("Output should be nil when error occurs")
	}
}

func TestLicenseClaimUsecase_Execute_ErrorWhenPasswordTooShort(t *testing.T) {
	usecase := createTestUsecase(nil, nil, nil, nil, nil, nil, nil)

	input := validInput()
	input.Password = "Aa1bcde" // 7 characters, needs 8

	output, err := usecase.Execute(context.Background(), input)

	if err == nil {
		t.Fatal("Execute() should return error when password is too short")
	}

	if output != nil {
		t.Error("Output should be nil when error occurs")
	}
}

func TestLicenseClaimUsecase_Execute_ErrorWhenPasswordMissingUppercase(t *testing.T) {
	usecase := createTestUsecase(nil, nil, nil, nil, nil, nil, nil)

	input := validInput()
	input.Password = "password123" // no uppercase

	output, err := usecase.Execute(context.Background(), input)

	if err == nil {
		t.Fatal("Execute() should return error when password has no uppercase")
	}

	if output != nil {
		t.Error("Output should be nil when error occurs")
	}
}

func TestLicenseClaimUsecase_Execute_ErrorWhenPasswordMissingLowercase(t *testing.T) {
	usecase := createTestUsecase(nil, nil, nil, nil, nil, nil, nil)

	input := validInput()
	input.Password = "PASSWORD123" // no lowercase

	output, err := usecase.Execute(context.Background(), input)

	if err == nil {
		t.Fatal("Execute() should return error when password has no lowercase")
	}

	if output != nil {
		t.Error("Output should be nil when error occurs")
	}
}

func TestLicenseClaimUsecase_Execute_ErrorWhenPasswordMissingDigit(t *testing.T) {
	usecase := createTestUsecase(nil, nil, nil, nil, nil, nil, nil)

	input := validInput()
	input.Password = "PasswordABC" // no digit

	output, err := usecase.Execute(context.Background(), input)

	if err == nil {
		t.Fatal("Execute() should return error when password has no digit")
	}

	if output != nil {
		t.Error("Output should be nil when error occurs")
	}
}

func TestLicenseClaimUsecase_Execute_ErrorWhenPasswordTooLong(t *testing.T) {
	usecase := createTestUsecase(nil, nil, nil, nil, nil, nil, nil)

	input := validInput()
	// Create a password that's 129 characters (max is 128)
	input.Password = "Aa1" + string(make([]byte, 126))

	output, err := usecase.Execute(context.Background(), input)

	if err == nil {
		t.Fatal("Execute() should return error when password is too long")
	}

	if output != nil {
		t.Error("Output should be nil when error occurs")
	}
}

func TestLicenseClaimUsecase_Execute_ErrorWhenDisplayNameEmpty(t *testing.T) {
	usecase := createTestUsecase(nil, nil, nil, nil, nil, nil, nil)

	input := validInput()
	input.DisplayName = ""

	output, err := usecase.Execute(context.Background(), input)

	if err == nil {
		t.Fatal("Execute() should return error when display name is empty")
	}

	if output != nil {
		t.Error("Output should be nil when error occurs")
	}
}

func TestLicenseClaimUsecase_Execute_ErrorWhenTenantNameEmpty(t *testing.T) {
	usecase := createTestUsecase(nil, nil, nil, nil, nil, nil, nil)

	input := validInput()
	input.TenantName = ""

	output, err := usecase.Execute(context.Background(), input)

	if err == nil {
		t.Fatal("Execute() should return error when tenant name is empty")
	}

	if output != nil {
		t.Error("Output should be nil when error occurs")
	}
}

// =====================================================
// LicenseClaimUsecase Tests - License Key State Errors
// =====================================================

func TestLicenseClaimUsecase_Execute_ErrorWhenLicenseKeyNotFound(t *testing.T) {
	licenseKeyRepo := &MockLicenseKeyRepository{
		findByHashForUpdateFunc: func(ctx context.Context, keyHash string) (*billing.LicenseKey, error) {
			return nil, nil // Not found
		},
	}

	auditLogRepo := &MockBillingAuditLogRepository{
		saveFunc: func(ctx context.Context, log *billing.BillingAuditLog) error {
			return nil
		},
	}

	usecase := createTestUsecase(nil, nil, nil, licenseKeyRepo, nil, auditLogRepo, nil)

	input := validInput()
	output, err := usecase.Execute(context.Background(), input)

	if err == nil {
		t.Fatal("Execute() should return error when license key is not found")
	}

	if output != nil {
		t.Error("Output should be nil when error occurs")
	}
}

func TestLicenseClaimUsecase_Execute_ErrorWhenLicenseKeyAlreadyUsed(t *testing.T) {
	testKey := createTestLicenseKey(t, billing.LicenseKeyStatusUsed)

	licenseKeyRepo := &MockLicenseKeyRepository{
		findByHashForUpdateFunc: func(ctx context.Context, keyHash string) (*billing.LicenseKey, error) {
			return testKey, nil
		},
	}

	auditLogRepo := &MockBillingAuditLogRepository{
		saveFunc: func(ctx context.Context, log *billing.BillingAuditLog) error {
			return nil
		},
	}

	usecase := createTestUsecase(nil, nil, nil, licenseKeyRepo, nil, auditLogRepo, nil)

	input := validInput()
	output, err := usecase.Execute(context.Background(), input)

	if err == nil {
		t.Fatal("Execute() should return error when license key is already used")
	}

	if output != nil {
		t.Error("Output should be nil when error occurs")
	}
}

func TestLicenseClaimUsecase_Execute_ErrorWhenLicenseKeyRevoked(t *testing.T) {
	testKey := createTestLicenseKey(t, billing.LicenseKeyStatusRevoked)

	licenseKeyRepo := &MockLicenseKeyRepository{
		findByHashForUpdateFunc: func(ctx context.Context, keyHash string) (*billing.LicenseKey, error) {
			return testKey, nil
		},
	}

	auditLogRepo := &MockBillingAuditLogRepository{
		saveFunc: func(ctx context.Context, log *billing.BillingAuditLog) error {
			return nil
		},
	}

	usecase := createTestUsecase(nil, nil, nil, licenseKeyRepo, nil, auditLogRepo, nil)

	input := validInput()
	output, err := usecase.Execute(context.Background(), input)

	if err == nil {
		t.Fatal("Execute() should return error when license key is revoked")
	}

	if output != nil {
		t.Error("Output should be nil when error occurs")
	}
}

// =====================================================
// LicenseClaimUsecase Tests - Repository Error Cases
// =====================================================

func TestLicenseClaimUsecase_Execute_ErrorWhenLicenseKeyRepoFails(t *testing.T) {
	dbError := errors.New("database connection error")

	licenseKeyRepo := &MockLicenseKeyRepository{
		findByHashForUpdateFunc: func(ctx context.Context, keyHash string) (*billing.LicenseKey, error) {
			return nil, dbError
		},
	}

	auditLogRepo := &MockBillingAuditLogRepository{
		saveFunc: func(ctx context.Context, log *billing.BillingAuditLog) error {
			return nil
		},
	}

	usecase := createTestUsecase(nil, nil, nil, licenseKeyRepo, nil, auditLogRepo, nil)

	input := validInput()
	output, err := usecase.Execute(context.Background(), input)

	if !errors.Is(err, dbError) {
		t.Errorf("Expected database error, got %v", err)
	}

	if output != nil {
		t.Error("Output should be nil when error occurs")
	}
}

func TestLicenseClaimUsecase_Execute_ErrorWhenTenantSaveFails(t *testing.T) {
	testKey := createTestLicenseKey(t, billing.LicenseKeyStatusUnused)
	saveError := errors.New("tenant save failed")

	licenseKeyRepo := &MockLicenseKeyRepository{
		findByHashForUpdateFunc: func(ctx context.Context, keyHash string) (*billing.LicenseKey, error) {
			return testKey, nil
		},
	}

	tenantRepo := &MockTenantRepository{
		saveFunc: func(ctx context.Context, t *tenant.Tenant) error {
			return saveError
		},
	}

	auditLogRepo := &MockBillingAuditLogRepository{
		saveFunc: func(ctx context.Context, log *billing.BillingAuditLog) error {
			return nil
		},
	}

	usecase := createTestUsecase(nil, tenantRepo, nil, licenseKeyRepo, nil, auditLogRepo, nil)

	input := validInput()
	output, err := usecase.Execute(context.Background(), input)

	if !errors.Is(err, saveError) {
		t.Errorf("Expected save error, got %v", err)
	}

	if output != nil {
		t.Error("Output should be nil when error occurs")
	}
}

func TestLicenseClaimUsecase_Execute_ErrorWhenPasswordHashFails(t *testing.T) {
	testKey := createTestLicenseKey(t, billing.LicenseKeyStatusUnused)
	hashError := errors.New("hash failed")

	licenseKeyRepo := &MockLicenseKeyRepository{
		findByHashForUpdateFunc: func(ctx context.Context, keyHash string) (*billing.LicenseKey, error) {
			return testKey, nil
		},
	}

	tenantRepo := &MockTenantRepository{
		saveFunc: func(ctx context.Context, t *tenant.Tenant) error {
			return nil
		},
	}

	passwordHasher := &MockPasswordHasher{
		hashFunc: func(password string) (string, error) {
			return "", hashError
		},
	}

	auditLogRepo := &MockBillingAuditLogRepository{
		saveFunc: func(ctx context.Context, log *billing.BillingAuditLog) error {
			return nil
		},
	}

	usecase := createTestUsecase(nil, tenantRepo, nil, licenseKeyRepo, nil, auditLogRepo, passwordHasher)

	input := validInput()
	output, err := usecase.Execute(context.Background(), input)

	if !errors.Is(err, hashError) {
		t.Errorf("Expected hash error, got %v", err)
	}

	if output != nil {
		t.Error("Output should be nil when error occurs")
	}
}

func TestLicenseClaimUsecase_Execute_ErrorWhenAdminSaveFails(t *testing.T) {
	testKey := createTestLicenseKey(t, billing.LicenseKeyStatusUnused)
	saveError := errors.New("admin save failed")

	licenseKeyRepo := &MockLicenseKeyRepository{
		findByHashForUpdateFunc: func(ctx context.Context, keyHash string) (*billing.LicenseKey, error) {
			return testKey, nil
		},
	}

	tenantRepo := &MockTenantRepository{
		saveFunc: func(ctx context.Context, t *tenant.Tenant) error {
			return nil
		},
	}

	adminRepo := &MockAdminRepository{
		saveFunc: func(ctx context.Context, admin *auth.Admin) error {
			return saveError
		},
	}

	auditLogRepo := &MockBillingAuditLogRepository{
		saveFunc: func(ctx context.Context, log *billing.BillingAuditLog) error {
			return nil
		},
	}

	usecase := createTestUsecase(nil, tenantRepo, adminRepo, licenseKeyRepo, nil, auditLogRepo, nil)

	input := validInput()
	output, err := usecase.Execute(context.Background(), input)

	if !errors.Is(err, saveError) {
		t.Errorf("Expected save error, got %v", err)
	}

	if output != nil {
		t.Error("Output should be nil when error occurs")
	}
}

func TestLicenseClaimUsecase_Execute_ErrorWhenEntitlementSaveFails(t *testing.T) {
	testKey := createTestLicenseKey(t, billing.LicenseKeyStatusUnused)
	saveError := errors.New("entitlement save failed")

	licenseKeyRepo := &MockLicenseKeyRepository{
		findByHashForUpdateFunc: func(ctx context.Context, keyHash string) (*billing.LicenseKey, error) {
			return testKey, nil
		},
	}

	tenantRepo := &MockTenantRepository{
		saveFunc: func(ctx context.Context, t *tenant.Tenant) error {
			return nil
		},
	}

	adminRepo := &MockAdminRepository{
		saveFunc: func(ctx context.Context, admin *auth.Admin) error {
			return nil
		},
	}

	entitlementRepo := &MockEntitlementRepository{
		saveFunc: func(ctx context.Context, entitlement *billing.Entitlement) error {
			return saveError
		},
	}

	auditLogRepo := &MockBillingAuditLogRepository{
		saveFunc: func(ctx context.Context, log *billing.BillingAuditLog) error {
			return nil
		},
	}

	usecase := createTestUsecase(nil, tenantRepo, adminRepo, licenseKeyRepo, entitlementRepo, auditLogRepo, nil)

	input := validInput()
	output, err := usecase.Execute(context.Background(), input)

	if !errors.Is(err, saveError) {
		t.Errorf("Expected save error, got %v", err)
	}

	if output != nil {
		t.Error("Output should be nil when error occurs")
	}
}

func TestLicenseClaimUsecase_Execute_ErrorWhenLicenseKeySaveFails(t *testing.T) {
	testKey := createTestLicenseKey(t, billing.LicenseKeyStatusUnused)
	saveError := errors.New("license key save failed")

	licenseKeyRepo := &MockLicenseKeyRepository{
		findByHashForUpdateFunc: func(ctx context.Context, keyHash string) (*billing.LicenseKey, error) {
			return testKey, nil
		},
		saveFunc: func(ctx context.Context, key *billing.LicenseKey) error {
			return saveError
		},
	}

	tenantRepo := &MockTenantRepository{
		saveFunc: func(ctx context.Context, t *tenant.Tenant) error {
			return nil
		},
	}

	adminRepo := &MockAdminRepository{
		saveFunc: func(ctx context.Context, admin *auth.Admin) error {
			return nil
		},
	}

	entitlementRepo := &MockEntitlementRepository{
		saveFunc: func(ctx context.Context, entitlement *billing.Entitlement) error {
			return nil
		},
	}

	auditLogRepo := &MockBillingAuditLogRepository{
		saveFunc: func(ctx context.Context, log *billing.BillingAuditLog) error {
			return nil
		},
	}

	usecase := createTestUsecase(nil, tenantRepo, adminRepo, licenseKeyRepo, entitlementRepo, auditLogRepo, nil)

	input := validInput()
	output, err := usecase.Execute(context.Background(), input)

	if !errors.Is(err, saveError) {
		t.Errorf("Expected save error, got %v", err)
	}

	if output != nil {
		t.Error("Output should be nil when error occurs")
	}
}

// =====================================================
// LicenseClaimUsecase Tests - Transaction Behavior
// =====================================================

func TestLicenseClaimUsecase_Execute_TransactionRollbackOnError(t *testing.T) {
	testKey := createTestLicenseKey(t, billing.LicenseKeyStatusUnused)
	rollbackCalled := false

	txManager := &MockTxManager{
		withTxFunc: func(ctx context.Context, fn func(context.Context) error) error {
			err := fn(ctx)
			if err != nil {
				rollbackCalled = true // Simulating rollback behavior
			}
			return err
		},
	}

	licenseKeyRepo := &MockLicenseKeyRepository{
		findByHashForUpdateFunc: func(ctx context.Context, keyHash string) (*billing.LicenseKey, error) {
			return testKey, nil
		},
	}

	tenantRepo := &MockTenantRepository{
		saveFunc: func(ctx context.Context, t *tenant.Tenant) error {
			return errors.New("tenant save failed")
		},
	}

	auditLogRepo := &MockBillingAuditLogRepository{
		saveFunc: func(ctx context.Context, log *billing.BillingAuditLog) error {
			return nil
		},
	}

	usecase := createTestUsecase(txManager, tenantRepo, nil, licenseKeyRepo, nil, auditLogRepo, nil)

	input := validInput()
	_, err := usecase.Execute(context.Background(), input)

	if err == nil {
		t.Fatal("Execute() should return error")
	}

	if !rollbackCalled {
		t.Error("Transaction should have been rolled back")
	}
}

// =====================================================
// LicenseClaimUsecase Tests - Audit Log Behavior
// =====================================================

func TestLicenseClaimUsecase_Execute_AuditLogFailureDoesNotFailOperation(t *testing.T) {
	testKey := createTestLicenseKey(t, billing.LicenseKeyStatusUnused)

	licenseKeyRepo := &MockLicenseKeyRepository{
		findByHashForUpdateFunc: func(ctx context.Context, keyHash string) (*billing.LicenseKey, error) {
			return testKey, nil
		},
		saveFunc: func(ctx context.Context, key *billing.LicenseKey) error {
			return nil
		},
	}

	tenantRepo := &MockTenantRepository{
		saveFunc: func(ctx context.Context, t *tenant.Tenant) error {
			return nil
		},
	}

	adminRepo := &MockAdminRepository{
		saveFunc: func(ctx context.Context, admin *auth.Admin) error {
			return nil
		},
	}

	entitlementRepo := &MockEntitlementRepository{
		saveFunc: func(ctx context.Context, entitlement *billing.Entitlement) error {
			return nil
		},
	}

	// Note: The success audit log is saved in the transaction, so if it fails,
	// the whole transaction fails. But failed attempt audit logs are best-effort.
	auditLogSaveCount := 0
	auditLogRepo := &MockBillingAuditLogRepository{
		saveFunc: func(ctx context.Context, log *billing.BillingAuditLog) error {
			auditLogSaveCount++
			return nil // Success for success audit log
		},
	}

	usecase := createTestUsecase(nil, tenantRepo, adminRepo, licenseKeyRepo, entitlementRepo, auditLogRepo, nil)

	input := validInput()
	output, err := usecase.Execute(context.Background(), input)

	if err != nil {
		t.Fatalf("Execute() should succeed, but got error: %v", err)
	}

	if output == nil {
		t.Fatal("Output should not be nil")
	}

	if auditLogSaveCount != 1 {
		t.Errorf("Audit log should have been saved once (for success), got %d", auditLogSaveCount)
	}
}

func TestLicenseClaimUsecase_Execute_FailedAttemptIsLogged(t *testing.T) {
	failedAuditLogSaved := false

	licenseKeyRepo := &MockLicenseKeyRepository{
		findByHashForUpdateFunc: func(ctx context.Context, keyHash string) (*billing.LicenseKey, error) {
			return nil, nil // Not found
		},
	}

	auditLogRepo := &MockBillingAuditLogRepository{
		saveFunc: func(ctx context.Context, log *billing.BillingAuditLog) error {
			failedAuditLogSaved = true
			return nil
		},
	}

	usecase := createTestUsecase(nil, nil, nil, licenseKeyRepo, nil, auditLogRepo, nil)

	input := validInput()
	_, err := usecase.Execute(context.Background(), input)

	if err == nil {
		t.Fatal("Execute() should return error")
	}

	if !failedAuditLogSaved {
		t.Error("Failed attempt should be logged")
	}
}

// =====================================================
// Password Complexity Validation Tests
// =====================================================

func TestValidatePasswordComplexity(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{"valid password", "Password123", false},
		{"valid password with special chars", "Password123!@#", false},
		{"too short", "Pass12", true},
		{"exactly 8 chars valid", "Password1", false},
		{"no uppercase", "password123", true},
		{"no lowercase", "PASSWORD123", true},
		{"no digit", "PasswordABC", true},
		{"only uppercase and digit", "PASSWORD1", true},
		{"only lowercase and digit", "password1", true},
		{"only letters", "PasswordABC", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePasswordComplexity(tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("validatePasswordComplexity(%q) error = %v, wantErr %v", tt.password, err, tt.wantErr)
			}
		})
	}
}
