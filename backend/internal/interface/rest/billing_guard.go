package rest

import (
	"context"
	"net/http"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/billing"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/tenant"
)

// BillingGuardDeps defines the dependencies for the billing guard
type BillingGuardDeps struct {
	TenantRepo      tenant.TenantRepository
	EntitlementRepo billing.EntitlementRepository
}

// ContextKeyTenantStatus is the context key for tenant billing status
const ContextKeyTenantStatus ContextKey = "tenant_status"

// BillingGuard is a middleware that enforces billing-related access control
// Priority:
// 1. revoked entitlement → 403 (all operations blocked)
// 2. tenant.status IN ('grace', 'suspended') → 403 for write, OK for read
// 3. active → OK
func BillingGuard(deps BillingGuardDeps) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			// Get tenant ID from context
			tenantID, ok := GetTenantID(ctx)
			if !ok {
				// No tenant ID means not authenticated, let other middleware handle it
				next.ServeHTTP(w, r)
				return
			}

			// Check if any entitlement is revoked
			hasRevoked, err := deps.EntitlementRepo.HasRevokedByTenantID(ctx, tenantID)
			if err != nil {
				RespondInternalError(w)
				return
			}

			if hasRevoked {
				RespondError(w, http.StatusForbidden, "ERR_ACCESS_REVOKED",
					"Your access has been revoked. Please contact support.", nil)
				return
			}

			// Get tenant to check status
			t, err := deps.TenantRepo.FindByID(ctx, tenantID)
			if err != nil {
				// Tenant not found is handled elsewhere
				if common.IsNotFoundError(err) {
					next.ServeHTTP(w, r)
					return
				}
				RespondInternalError(w)
				return
			}

			// Store tenant status in context for handlers that need it
			ctx = context.WithValue(ctx, ContextKeyTenantStatus, t.Status())
			r = r.WithContext(ctx)

			// Check if this is a write operation
			isWriteOperation := r.Method == http.MethodPost ||
				r.Method == http.MethodPut ||
				r.Method == http.MethodPatch ||
				r.Method == http.MethodDelete

			// For write operations, check tenant status
			if isWriteOperation {
				switch t.Status() {
				case tenant.TenantStatusGrace:
					RespondError(w, http.StatusForbidden, "ERR_GRACE_PERIOD",
						"Your account is in grace period. Write operations are disabled. Please update your payment.", nil)
					return
				case tenant.TenantStatusSuspended:
					RespondError(w, http.StatusForbidden, "ERR_SUSPENDED",
						"Your account is suspended. Please contact support to restore access.", nil)
					return
				}
			}

			// Read operations are allowed for grace and suspended (unless revoked, which is already checked)
			next.ServeHTTP(w, r)
		})
	}
}

// GetTenantStatus extracts tenant status from context
func GetTenantStatus(ctx context.Context) (tenant.TenantStatus, bool) {
	status, ok := ctx.Value(ContextKeyTenantStatus).(tenant.TenantStatus)
	return status, ok
}

// RequireOwner is a middleware that requires the owner role
func RequireOwner(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		role, ok := ctx.Value(ContextKeyRole).(string)
		if !ok || role != "owner" {
			RespondError(w, http.StatusForbidden, "ERR_FORBIDDEN",
				"Owner permission required", nil)
			return
		}

		next.ServeHTTP(w, r)
	})
}
