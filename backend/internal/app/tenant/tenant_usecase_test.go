package tenant_test

import (
	"context"
	"errors"
	"testing"
	"time"

	apptenant "github.com/erenoa/vrc-shift-scheduler/backend/internal/app/tenant"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/tenant"
)

// =====================================================
// Mock Implementations
// =====================================================

// MockTenantRepository is a mock implementation of tenant.TenantRepository
type MockTenantRepository struct {
	findByIDFunc func(ctx context.Context, tenantID common.TenantID) (*tenant.Tenant, error)
	saveFunc     func(ctx context.Context, t *tenant.Tenant) error
}

func (m *MockTenantRepository) FindByID(ctx context.Context, tenantID common.TenantID) (*tenant.Tenant, error) {
	if m.findByIDFunc != nil {
		return m.findByIDFunc(ctx, tenantID)
	}
	return nil, errors.New("not implemented")
}

func (m *MockTenantRepository) Save(ctx context.Context, t *tenant.Tenant) error {
	if m.saveFunc != nil {
		return m.saveFunc(ctx, t)
	}
	return nil
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

// =====================================================
// GetTenantUsecase Tests
// =====================================================

func TestGetTenantUsecase_Execute_Success(t *testing.T) {
	testTenant := createTestTenant(t)

	repo := &MockTenantRepository{
		findByIDFunc: func(ctx context.Context, tenantID common.TenantID) (*tenant.Tenant, error) {
			return testTenant, nil
		},
	}

	usecase := apptenant.NewGetTenantUsecase(repo)

	input := apptenant.GetTenantInput{
		TenantID: testTenant.TenantID(),
	}

	result, err := usecase.Execute(context.Background(), input)

	if err != nil {
		t.Fatalf("Execute() should succeed, got error: %v", err)
	}

	if result.TenantID() != testTenant.TenantID() {
		t.Errorf("TenantID mismatch: got %v, want %v", result.TenantID(), testTenant.TenantID())
	}

	if result.TenantName() != "Test Tenant" {
		t.Errorf("TenantName mismatch: got %v, want 'Test Tenant'", result.TenantName())
	}
}

func TestGetTenantUsecase_Execute_NotFound(t *testing.T) {
	tenantID := common.NewTenantID()

	repo := &MockTenantRepository{
		findByIDFunc: func(ctx context.Context, tid common.TenantID) (*tenant.Tenant, error) {
			return nil, common.NewNotFoundError("tenant", tid.String())
		},
	}

	usecase := apptenant.NewGetTenantUsecase(repo)

	input := apptenant.GetTenantInput{
		TenantID: tenantID,
	}

	_, err := usecase.Execute(context.Background(), input)

	if err == nil {
		t.Fatal("Execute() should fail when tenant not found")
	}
}

func TestGetTenantUsecase_Execute_ErrorWhenFindFails(t *testing.T) {
	tenantID := common.NewTenantID()

	repo := &MockTenantRepository{
		findByIDFunc: func(ctx context.Context, tid common.TenantID) (*tenant.Tenant, error) {
			return nil, errors.New("database error")
		},
	}

	usecase := apptenant.NewGetTenantUsecase(repo)

	input := apptenant.GetTenantInput{
		TenantID: tenantID,
	}

	_, err := usecase.Execute(context.Background(), input)

	if err == nil {
		t.Fatal("Execute() should fail when find fails")
	}
}

// =====================================================
// UpdateTenantUsecase Tests
// =====================================================

func TestUpdateTenantUsecase_Execute_Success(t *testing.T) {
	testTenant := createTestTenant(t)

	repo := &MockTenantRepository{
		findByIDFunc: func(ctx context.Context, tenantID common.TenantID) (*tenant.Tenant, error) {
			return testTenant, nil
		},
		saveFunc: func(ctx context.Context, t *tenant.Tenant) error {
			return nil
		},
	}

	usecase := apptenant.NewUpdateTenantUsecase(repo)

	input := apptenant.UpdateTenantInput{
		TenantID:   testTenant.TenantID(),
		TenantName: "Updated Tenant Name",
	}

	result, err := usecase.Execute(context.Background(), input)

	if err != nil {
		t.Fatalf("Execute() should succeed, got error: %v", err)
	}

	if result.TenantName() != "Updated Tenant Name" {
		t.Errorf("TenantName should be updated: got %v, want 'Updated Tenant Name'", result.TenantName())
	}
}

func TestUpdateTenantUsecase_Execute_NotFound(t *testing.T) {
	tenantID := common.NewTenantID()

	repo := &MockTenantRepository{
		findByIDFunc: func(ctx context.Context, tid common.TenantID) (*tenant.Tenant, error) {
			return nil, common.NewNotFoundError("tenant", tid.String())
		},
	}

	usecase := apptenant.NewUpdateTenantUsecase(repo)

	input := apptenant.UpdateTenantInput{
		TenantID:   tenantID,
		TenantName: "Updated Name",
	}

	_, err := usecase.Execute(context.Background(), input)

	if err == nil {
		t.Fatal("Execute() should fail when tenant not found")
	}
}

func TestUpdateTenantUsecase_Execute_ErrorWhenSaveFails(t *testing.T) {
	testTenant := createTestTenant(t)

	repo := &MockTenantRepository{
		findByIDFunc: func(ctx context.Context, tenantID common.TenantID) (*tenant.Tenant, error) {
			return testTenant, nil
		},
		saveFunc: func(ctx context.Context, t *tenant.Tenant) error {
			return errors.New("database error")
		},
	}

	usecase := apptenant.NewUpdateTenantUsecase(repo)

	input := apptenant.UpdateTenantInput{
		TenantID:   testTenant.TenantID(),
		TenantName: "Updated Name",
	}

	_, err := usecase.Execute(context.Background(), input)

	if err == nil {
		t.Fatal("Execute() should fail when save fails")
	}
}

func TestUpdateTenantUsecase_Execute_ErrorWhenEmptyName(t *testing.T) {
	testTenant := createTestTenant(t)

	repo := &MockTenantRepository{
		findByIDFunc: func(ctx context.Context, tenantID common.TenantID) (*tenant.Tenant, error) {
			return testTenant, nil
		},
	}

	usecase := apptenant.NewUpdateTenantUsecase(repo)

	input := apptenant.UpdateTenantInput{
		TenantID:   testTenant.TenantID(),
		TenantName: "", // Empty name
	}

	_, err := usecase.Execute(context.Background(), input)

	if err == nil {
		t.Fatal("Execute() should fail when tenant name is empty")
	}
}
