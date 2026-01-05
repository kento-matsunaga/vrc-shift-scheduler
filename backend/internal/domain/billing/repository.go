package billing

import (
	"context"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
)

// PlanRepository defines the interface for plan persistence
type PlanRepository interface {
	// FindByCode finds a plan by its code
	FindByCode(ctx context.Context, planCode string) (*Plan, error)

	// FindAll retrieves all plans
	FindAll(ctx context.Context) ([]*Plan, error)
}

// EntitlementRepository defines the interface for entitlement persistence
type EntitlementRepository interface {
	// Save saves an entitlement
	Save(ctx context.Context, entitlement *Entitlement) error

	// FindByID finds an entitlement by ID
	FindByID(ctx context.Context, entitlementID EntitlementID) (*Entitlement, error)

	// FindByTenantID finds all entitlements for a tenant
	FindByTenantID(ctx context.Context, tenantID common.TenantID) ([]*Entitlement, error)

	// FindActiveByTenantID finds the active entitlement for a tenant (prioritized)
	FindActiveByTenantID(ctx context.Context, tenantID common.TenantID) (*Entitlement, error)

	// HasRevokedByTenantID checks if any entitlement for the tenant is revoked
	HasRevokedByTenantID(ctx context.Context, tenantID common.TenantID) (bool, error)
}

// LicenseKeyRepository defines the interface for license key persistence
type LicenseKeyRepository interface {
	// Save saves a license key
	Save(ctx context.Context, key *LicenseKey) error

	// SaveBatch saves multiple license keys
	SaveBatch(ctx context.Context, keys []*LicenseKey) error

	// FindByHash finds a license key by its hash (with row lock for claim)
	FindByHashForUpdate(ctx context.Context, keyHash string) (*LicenseKey, error)

	// FindByID finds a license key by ID
	FindByID(ctx context.Context, keyID LicenseKeyID) (*LicenseKey, error)

	// FindByBatchID finds all license keys in a batch
	FindByBatchID(ctx context.Context, batchID string) ([]*LicenseKey, error)

	// CountByStatus counts license keys by status
	CountByStatus(ctx context.Context, status LicenseKeyStatus) (int, error)

	// RevokeBatch revokes all keys in a batch
	RevokeBatch(ctx context.Context, batchID string) error

	// List returns license keys with optional status filter
	List(ctx context.Context, status *LicenseKeyStatus, limit, offset int) ([]*LicenseKey, int, error)

	// FindByHashAndTenant はハッシュとテナントIDで使用済みライセンスキーを検索
	// PWリセット時の本人確認に使用
	FindByHashAndTenant(ctx context.Context, keyHash string, tenantID common.TenantID) (*LicenseKey, error)
}

// SubscriptionRepository defines the interface for subscription persistence
type SubscriptionRepository interface {
	// Save saves a subscription
	Save(ctx context.Context, sub *Subscription) error

	// FindByTenantID finds a subscription by tenant ID
	FindByTenantID(ctx context.Context, tenantID common.TenantID) (*Subscription, error)

	// FindByStripeSubscriptionID finds a subscription by Stripe subscription ID
	FindByStripeSubscriptionID(ctx context.Context, stripeSubID string) (*Subscription, error)
}

// WebhookEventRepository defines the interface for webhook event persistence
type WebhookEventRepository interface {
	// TryInsert attempts to insert a webhook event, returns false if already exists
	TryInsert(ctx context.Context, provider string, eventID string, payloadJSON *string) (bool, error)

	// DeleteOlderThan deletes webhook events older than the specified time
	DeleteOlderThan(ctx context.Context, before int) (int64, error)
}

// BillingAuditLogRepository defines the interface for billing audit log persistence
type BillingAuditLogRepository interface {
	// Save saves a billing audit log
	Save(ctx context.Context, log *BillingAuditLog) error

	// FindByDateRange finds audit logs within a date range with pagination
	FindByDateRange(ctx context.Context, startDate, endDate string, limit, offset int) ([]*BillingAuditLog, error)

	// FindByAction finds audit logs by action
	FindByAction(ctx context.Context, action string, limit, offset int) ([]*BillingAuditLog, error)

	// CountByDateRange counts audit logs within a date range
	CountByDateRange(ctx context.Context, startDate, endDate string) (int, error)

	// List returns audit logs with optional action filter
	List(ctx context.Context, action *string, limit, offset int) ([]*BillingAuditLog, int, error)
}
