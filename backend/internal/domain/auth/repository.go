package auth

import (
	"context"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
)

// AdminRepository defines the interface for Admin persistence
// Multi-Tenant前提: 全メソッドで tenant_id を引数に取る
type AdminRepository interface {
	// Save saves an admin (insert or update)
	Save(ctx context.Context, admin *Admin) error

	// FindByID finds an admin by ID within a tenant
	FindByID(ctx context.Context, tenantID common.TenantID, adminID common.AdminID) (*Admin, error)

	// FindByEmail finds an admin by email within a tenant
	// ログイン時に使用
	FindByEmail(ctx context.Context, tenantID common.TenantID, email string) (*Admin, error)

	// FindByTenantID finds all admins within a tenant
	// deleted_at IS NULL のレコードのみ返す
	FindByTenantID(ctx context.Context, tenantID common.TenantID) ([]*Admin, error)

	// FindActiveByTenantID finds all active admins within a tenant
	FindActiveByTenantID(ctx context.Context, tenantID common.TenantID) ([]*Admin, error)

	// Delete deletes an admin (physical delete)
	// 通常は Admin.Delete() で論理削除を使用するため、このメソッドは稀に使用
	Delete(ctx context.Context, tenantID common.TenantID, adminID common.AdminID) error

	// ExistsByEmail checks if an admin with the given email exists within a tenant
	ExistsByEmail(ctx context.Context, tenantID common.TenantID, email string) (bool, error)
}
