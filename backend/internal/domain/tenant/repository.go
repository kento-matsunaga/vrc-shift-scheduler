package tenant

import (
	"context"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
)

// TenantRepository defines the interface for tenant persistence
type TenantRepository interface {
	// FindByID finds a tenant by ID
	FindByID(ctx context.Context, tenantID common.TenantID) (*Tenant, error)

	// Save saves a tenant (INSERT or UPDATE)
	Save(ctx context.Context, tenant *Tenant) error
}
