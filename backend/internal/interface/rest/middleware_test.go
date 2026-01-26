package rest_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/tenant"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/interface/rest"
)

// MockTenantRepository is a mock implementation of tenant.TenantRepository
type MockTenantRepository struct {
	findByIDFunc func(ctx context.Context, id common.TenantID) (*tenant.Tenant, error)
}

func (m *MockTenantRepository) FindByID(ctx context.Context, id common.TenantID) (*tenant.Tenant, error) {
	if m.findByIDFunc != nil {
		return m.findByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockTenantRepository) FindByPendingStripeSessionID(ctx context.Context, sessionID string) (*tenant.Tenant, error) {
	return nil, nil
}

func (m *MockTenantRepository) Save(ctx context.Context, t *tenant.Tenant) error {
	return nil
}

func (m *MockTenantRepository) ListAll(ctx context.Context, status *tenant.TenantStatus, limit, offset int) ([]*tenant.Tenant, int, error) {
	return nil, 0, nil
}

// =====================================================
// TenantStatusMiddleware Tests
// =====================================================

func TestTenantStatusMiddleware_ActiveTenant_Passes(t *testing.T) {
	// Create an active tenant
	activeTenant, _ := tenant.NewTenant(time.Now(), "Test Tenant", "Asia/Tokyo")

	mockRepo := &MockTenantRepository{
		findByIDFunc: func(ctx context.Context, id common.TenantID) (*tenant.Tenant, error) {
			return activeTenant, nil
		},
	}

	// Create a handler that records if it was called
	handlerCalled := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	})

	// Wrap with middleware
	middleware := rest.TenantStatusMiddleware(mockRepo)
	wrappedHandler := middleware(handler)

	// Create request with tenant ID in context
	req := httptest.NewRequest("GET", "/test", nil)
	ctx := context.WithValue(req.Context(), rest.ContextKeyTenantID, activeTenant.TenantID())
	req = req.WithContext(ctx)

	// Execute
	rr := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(rr, req)

	// Verify
	if !handlerCalled {
		t.Error("Handler should have been called for active tenant")
	}
	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}
}

func TestTenantStatusMiddleware_GraceTenant_Passes(t *testing.T) {
	// Create a grace tenant (valid transition: active -> grace)
	graceTenant, _ := tenant.NewTenant(time.Now(), "Test Tenant", "Asia/Tokyo")
	graceUntil := time.Now().Add(14 * 24 * time.Hour)
	_ = graceTenant.SetStatusGrace(time.Now(), graceUntil)

	mockRepo := &MockTenantRepository{
		findByIDFunc: func(ctx context.Context, id common.TenantID) (*tenant.Tenant, error) {
			return graceTenant, nil
		},
	}

	handlerCalled := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	})

	middleware := rest.TenantStatusMiddleware(mockRepo)
	wrappedHandler := middleware(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	ctx := context.WithValue(req.Context(), rest.ContextKeyTenantID, graceTenant.TenantID())
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(rr, req)

	// Grace tenant should be allowed
	if !handlerCalled {
		t.Error("Handler should have been called for grace tenant")
	}
	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}
}

func TestTenantStatusMiddleware_SuspendedTenant_Blocked(t *testing.T) {
	// Create a suspended tenant (valid transition: active -> grace -> suspended)
	suspendedTenant, _ := tenant.NewTenant(time.Now(), "Test Tenant", "Asia/Tokyo")
	_ = suspendedTenant.SetStatusGrace(time.Now(), time.Now().Add(7*24*time.Hour))
	_ = suspendedTenant.SetStatusSuspended(time.Now())

	mockRepo := &MockTenantRepository{
		findByIDFunc: func(ctx context.Context, id common.TenantID) (*tenant.Tenant, error) {
			return suspendedTenant, nil
		},
	}

	handlerCalled := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	})

	middleware := rest.TenantStatusMiddleware(mockRepo)
	wrappedHandler := middleware(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	ctx := context.WithValue(req.Context(), rest.ContextKeyTenantID, suspendedTenant.TenantID())
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(rr, req)

	// Suspended tenant should be blocked
	if handlerCalled {
		t.Error("Handler should NOT have been called for suspended tenant")
	}
	if rr.Code != http.StatusForbidden {
		t.Errorf("Expected status 403, got %d", rr.Code)
	}

	// Check error message
	body := rr.Body.String()
	if body == "" {
		t.Error("Response body should contain error message")
	}
}

func TestTenantStatusMiddleware_NoTenantID_Passes(t *testing.T) {
	mockRepo := &MockTenantRepository{}

	handlerCalled := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	})

	middleware := rest.TenantStatusMiddleware(mockRepo)
	wrappedHandler := middleware(handler)

	// Request without tenant ID in context
	req := httptest.NewRequest("GET", "/test", nil)

	rr := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(rr, req)

	// Should pass through when no tenant ID
	if !handlerCalled {
		t.Error("Handler should have been called when no tenant ID")
	}
	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}
}
