package payment

import (
	"context"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/billing"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
)

// BillingStatusUsecase handles billing status retrieval
type BillingStatusUsecase struct {
	subscriptionRepo SubscriptionRepository
	entitlementRepo  EntitlementRepository
}

// EntitlementRepository defines the interface for entitlement data access
type EntitlementRepository interface {
	FindActiveByTenantID(ctx context.Context, tenantID common.TenantID) (*billing.Entitlement, error)
}

// NewBillingStatusUsecase creates a new BillingStatusUsecase
func NewBillingStatusUsecase(
	subscriptionRepo SubscriptionRepository,
	entitlementRepo EntitlementRepository,
) *BillingStatusUsecase {
	return &BillingStatusUsecase{
		subscriptionRepo: subscriptionRepo,
		entitlementRepo:  entitlementRepo,
	}
}

// BillingStatusInput represents the input for getting billing status
type BillingStatusInput struct {
	TenantID common.TenantID
}

// BillingStatusOutput represents the billing status
type BillingStatusOutput struct {
	PlanType         string  `json:"plan_type"`          // "subscription", "lifetime", "none"
	PlanName         string  `json:"plan_name"`          // 表示用プラン名
	Status           string  `json:"status"`             // "active", "canceled", "past_due", etc.
	CurrentPeriodEnd *string `json:"current_period_end"` // サブスクの場合、次回更新日
	CancelAtPeriodEnd bool   `json:"cancel_at_period_end"` // キャンセル予定
}

// Execute gets the billing status for a tenant
func (uc *BillingStatusUsecase) Execute(ctx context.Context, input BillingStatusInput) (*BillingStatusOutput, error) {
	// まずEntitlementを確認
	entitlement, err := uc.entitlementRepo.FindActiveByTenantID(ctx, input.TenantID)
	if err != nil {
		return nil, err
	}

	// Entitlementがない場合
	if entitlement == nil {
		return &BillingStatusOutput{
			PlanType: "none",
			PlanName: "プランなし",
			Status:   "inactive",
		}, nil
	}

	// LIFETIMEプランの場合
	if entitlement.PlanCode() == "LIFETIME" {
		return &BillingStatusOutput{
			PlanType: "lifetime",
			PlanName: "買い切りプラン",
			Status:   "active",
		}, nil
	}

	// サブスクリプションプランの場合
	subscription, err := uc.subscriptionRepo.FindByTenantID(ctx, input.TenantID)
	if err != nil {
		return nil, err
	}

	output := &BillingStatusOutput{
		PlanType: "subscription",
		PlanName: "月額プラン",
		Status:   "active",
	}

	if subscription != nil {
		output.Status = string(subscription.Status())
		if subscription.CurrentPeriodEnd() != nil {
			periodEnd := subscription.CurrentPeriodEnd().Format("2006-01-02")
			output.CurrentPeriodEnd = &periodEnd
		}
	}

	return output, nil
}
