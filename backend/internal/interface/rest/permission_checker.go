package rest

import (
	"context"
	"net/http"

	apptenant "github.com/erenoa/vrc-shift-scheduler/backend/internal/app/tenant"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/tenant"
)

// PermissionChecker provides permission checking for handlers
type PermissionChecker struct {
	checkPermissionUC *apptenant.CheckManagerPermissionUsecase
}

// NewPermissionChecker creates a new PermissionChecker
func NewPermissionChecker(checkPermissionUC *apptenant.CheckManagerPermissionUsecase) *PermissionChecker {
	return &PermissionChecker{
		checkPermissionUC: checkPermissionUC,
	}
}

// RequirePermission returns a middleware that checks if the user has the required permission
// Owner always has all permissions, Manager permissions are checked against settings
func (pc *PermissionChecker) RequirePermission(permType tenant.PermissionType) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			// Get role from context
			role, ok := GetRole(ctx)
			if !ok {
				// No role in context - this means we're using X-Tenant-ID header (legacy mode)
				// For backward compatibility, allow the request
				next.ServeHTTP(w, r)
				return
			}

			// Owner has all permissions
			if role == "owner" {
				next.ServeHTTP(w, r)
				return
			}

			// For manager, check the specific permission
			if role == "manager" {
				tenantID, ok := GetTenantID(ctx)
				if !ok {
					RespondBadRequest(w, "tenant_id is required")
					return
				}

				hasPermission, err := pc.checkPermission(ctx, tenantID, permType)
				if err != nil {
					RespondInternalError(w)
					return
				}

				if !hasPermission {
					RespondError(w, http.StatusForbidden, "ERR_FORBIDDEN", "この操作を行う権限がありません", nil)
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

// checkPermission checks if a manager has the specified permission
func (pc *PermissionChecker) checkPermission(ctx context.Context, tenantID common.TenantID, permType tenant.PermissionType) (bool, error) {
	return pc.checkPermissionUC.Execute(ctx, apptenant.CheckManagerPermissionInput{
		TenantID:       tenantID,
		PermissionType: permType,
	})
}

// CheckPermission directly checks if the current user has the specified permission
// Returns true if user is owner or if manager has the permission
func (pc *PermissionChecker) CheckPermission(ctx context.Context, permType tenant.PermissionType) (bool, error) {
	role, ok := GetRole(ctx)
	if !ok {
		// No role - legacy mode, allow
		return true, nil
	}

	if role == "owner" {
		return true, nil
	}

	if role == "manager" {
		tenantID, ok := GetTenantID(ctx)
		if !ok {
			return false, nil
		}

		return pc.checkPermissionUC.Execute(ctx, apptenant.CheckManagerPermissionInput{
			TenantID:       tenantID,
			PermissionType: permType,
		})
	}

	return false, nil
}
