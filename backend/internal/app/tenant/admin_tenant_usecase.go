package tenant

import (
	"context"
	"fmt"
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/auth"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/billing"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/services"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/tenant"
)

// AdminTenantUsecase handles admin operations for tenants
type AdminTenantUsecase struct {
	txManager        services.TxManager
	tenantRepo       tenant.TenantRepository
	adminRepo        auth.AdminRepository
	entitlementRepo  billing.EntitlementRepository
	subscriptionRepo billing.SubscriptionRepository
	auditLogRepo     billing.BillingAuditLogRepository
}

// NewAdminTenantUsecase creates a new AdminTenantUsecase
func NewAdminTenantUsecase(
	txManager services.TxManager,
	tenantRepo tenant.TenantRepository,
	adminRepo auth.AdminRepository,
	entitlementRepo billing.EntitlementRepository,
	subscriptionRepo billing.SubscriptionRepository,
	auditLogRepo billing.BillingAuditLogRepository,
) *AdminTenantUsecase {
	return &AdminTenantUsecase{
		txManager:        txManager,
		tenantRepo:       tenantRepo,
		adminRepo:        adminRepo,
		entitlementRepo:  entitlementRepo,
		subscriptionRepo: subscriptionRepo,
		auditLogRepo:     auditLogRepo,
	}
}

// TenantListInput represents input for listing tenants
type TenantListInput struct {
	Status *tenant.TenantStatus
	Limit  int
	Offset int
}

// TenantListItem represents a tenant in the list
type TenantListItem struct {
	TenantID   common.TenantID
	TenantName string
	Status     tenant.TenantStatus
	GraceUntil *time.Time
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// TenantListOutput represents output from listing tenants
type TenantListOutput struct {
	Tenants    []TenantListItem
	TotalCount int
}

// List returns a list of tenants
func (uc *AdminTenantUsecase) List(ctx context.Context, input TenantListInput) (*TenantListOutput, error) {
	if input.Limit <= 0 {
		input.Limit = 50
	}
	if input.Limit > 100 {
		input.Limit = 100
	}

	tenants, totalCount, err := uc.tenantRepo.ListAll(ctx, input.Status, input.Limit, input.Offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list tenants: %w", err)
	}

	items := make([]TenantListItem, len(tenants))
	for i, t := range tenants {
		items[i] = TenantListItem{
			TenantID:   t.TenantID(),
			TenantName: t.TenantName(),
			Status:     t.Status(),
			GraceUntil: t.GraceUntil(),
			CreatedAt:  t.CreatedAt(),
			UpdatedAt:  t.UpdatedAt(),
		}
	}

	return &TenantListOutput{
		Tenants:    items,
		TotalCount: totalCount,
	}, nil
}

// TenantDetailOutput represents detailed tenant information
type TenantDetailOutput struct {
	TenantID     common.TenantID
	TenantName   string
	Status       tenant.TenantStatus
	GraceUntil   *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
	Entitlements []EntitlementInfo
	Subscription *SubscriptionInfo
	Admins       []AdminInfo
}

// SubscriptionInfo represents subscription information for tenant detail
type SubscriptionInfo struct {
	SubscriptionID       billing.SubscriptionID
	StripeCustomerID     string
	StripeSubscriptionID string
	Status               billing.SubscriptionStatus
	CurrentPeriodEnd     *time.Time
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

// EntitlementInfo represents entitlement information
type EntitlementInfo struct {
	EntitlementID billing.EntitlementID
	PlanCode      string
	Source        billing.EntitlementSource
	StartsAt      time.Time
	RevokedAt     *time.Time
}

// AdminInfo represents admin information for tenant detail
type AdminInfo struct {
	AdminID     common.AdminID
	Email       string
	DisplayName string
	Role        auth.Role
}

// GetDetail returns detailed tenant information
func (uc *AdminTenantUsecase) GetDetail(ctx context.Context, tenantID common.TenantID) (*TenantDetailOutput, error) {
	t, err := uc.tenantRepo.FindByID(ctx, tenantID)
	if err != nil {
		if common.IsNotFoundError(err) {
			return nil, common.NewDomainError(common.ErrNotFound, "Tenant not found")
		}
		return nil, fmt.Errorf("failed to find tenant: %w", err)
	}

	entitlements, err := uc.entitlementRepo.FindByTenantID(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to find entitlements: %w", err)
	}

	entitlementInfos := make([]EntitlementInfo, len(entitlements))
	for i, e := range entitlements {
		entitlementInfos[i] = EntitlementInfo{
			EntitlementID: e.EntitlementID(),
			PlanCode:      e.PlanCode(),
			Source:        e.Source(),
			StartsAt:      e.StartsAt(),
			RevokedAt:     e.RevokedAt(),
		}
	}

	// Fetch subscription for the tenant
	var subscriptionInfo *SubscriptionInfo
	sub, err := uc.subscriptionRepo.FindByTenantID(ctx, tenantID)
	if err != nil && !common.IsNotFoundError(err) {
		return nil, fmt.Errorf("failed to find subscription: %w", err)
	}
	if sub != nil {
		subscriptionInfo = &SubscriptionInfo{
			SubscriptionID:       sub.SubscriptionID(),
			StripeCustomerID:     sub.StripeCustomerID(),
			StripeSubscriptionID: sub.StripeSubscriptionID(),
			Status:               sub.Status(),
			CurrentPeriodEnd:     sub.CurrentPeriodEnd(),
			CreatedAt:            sub.CreatedAt(),
			UpdatedAt:            sub.UpdatedAt(),
		}
	}

	// Fetch admins for the tenant
	admins, err := uc.adminRepo.FindByTenantID(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to find admins: %w", err)
	}

	adminInfos := make([]AdminInfo, len(admins))
	for i, a := range admins {
		adminInfos[i] = AdminInfo{
			AdminID:     a.AdminID(),
			Email:       a.Email(),
			DisplayName: a.DisplayName(),
			Role:        a.Role(),
		}
	}

	return &TenantDetailOutput{
		TenantID:     t.TenantID(),
		TenantName:   t.TenantName(),
		Status:       t.Status(),
		GraceUntil:   t.GraceUntil(),
		CreatedAt:    t.CreatedAt(),
		UpdatedAt:    t.UpdatedAt(),
		Entitlements: entitlementInfos,
		Subscription: subscriptionInfo,
		Admins:       adminInfos,
	}, nil
}

// UpdateTenantStatusInput represents input for updating tenant status
type UpdateTenantStatusInput struct {
	TenantID   common.TenantID
	Status     tenant.TenantStatus
	GraceUntil *time.Time // Required when status is "grace"
	AdminID    common.AdminID
}

// UpdateStatus updates a tenant's status
func (uc *AdminTenantUsecase) UpdateStatus(ctx context.Context, input UpdateTenantStatusInput) error {
	now := time.Now().UTC()

	return uc.txManager.WithTx(ctx, func(txCtx context.Context) error {
		t, err := uc.tenantRepo.FindByID(txCtx, input.TenantID)
		if err != nil {
			if common.IsNotFoundError(err) {
				return common.NewDomainError(common.ErrNotFound, "Tenant not found")
			}
			return fmt.Errorf("failed to find tenant: %w", err)
		}

		previousStatus := t.Status()

		switch input.Status {
		case tenant.TenantStatusActive:
			t.SetStatusActive(now)
		case tenant.TenantStatusGrace:
			if input.GraceUntil == nil {
				return common.NewDomainError("ERR_INVALID_INPUT", "grace_until is required for grace status")
			}
			t.SetStatusGrace(now, *input.GraceUntil)
		case tenant.TenantStatusSuspended:
			t.SetStatusSuspended(now)
		default:
			return common.NewDomainError("ERR_INVALID_STATUS", "Invalid status")
		}

		if err := uc.tenantRepo.Save(txCtx, t); err != nil {
			return fmt.Errorf("failed to save tenant: %w", err)
		}

		// Audit log
		adminIDStr := input.AdminID.String()
		tenantIDStr := input.TenantID.String()
		beforeJSON := fmt.Sprintf(`{"status":"%s"}`, previousStatus)
		afterJSON := fmt.Sprintf(`{"status":"%s"}`, input.Status)
		auditLog, err := billing.NewBillingAuditLog(
			now,
			billing.ActorTypeAdmin,
			&adminIDStr,
			string(billing.BillingAuditActionTenantStatusChange),
			strPtr("tenant"),
			&tenantIDStr,
			&beforeJSON,
			&afterJSON,
			nil,
			nil,
		)
		if err != nil {
			return err
		}
		return uc.auditLogRepo.Save(txCtx, auditLog)
	})
}

// strPtr returns a pointer to the given string
func strPtr(s string) *string {
	return &s
}
