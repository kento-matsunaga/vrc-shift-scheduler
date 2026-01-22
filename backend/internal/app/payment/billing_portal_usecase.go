package payment

import (
	"context"
	"fmt"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/billing"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	infrastripe "github.com/erenoa/vrc-shift-scheduler/backend/internal/infra/stripe"
)

// BillingPortalUsecase handles billing portal session creation
type BillingPortalUsecase struct {
	subscriptionRepo SubscriptionRepository
	stripeClient     *infrastripe.Client
	returnURL        string
}

// SubscriptionRepository defines the interface for subscription data access
type SubscriptionRepository interface {
	FindByTenantID(ctx context.Context, tenantID common.TenantID) (*billing.Subscription, error)
}

// NewBillingPortalUsecase creates a new BillingPortalUsecase
func NewBillingPortalUsecase(
	subscriptionRepo SubscriptionRepository,
	stripeClient *infrastripe.Client,
	returnURL string,
) *BillingPortalUsecase {
	return &BillingPortalUsecase{
		subscriptionRepo: subscriptionRepo,
		stripeClient:     stripeClient,
		returnURL:        returnURL,
	}
}

// BillingPortalInput represents the input for creating a billing portal session
type BillingPortalInput struct {
	TenantID common.TenantID
}

// BillingPortalOutput represents the output of creating a billing portal session
type BillingPortalOutput struct {
	PortalURL string `json:"portal_url"`
}

// Execute creates a Stripe Customer Portal session for the tenant
func (uc *BillingPortalUsecase) Execute(ctx context.Context, input BillingPortalInput) (*BillingPortalOutput, error) {
	// Find the subscription for this tenant
	subscription, err := uc.subscriptionRepo.FindByTenantID(ctx, input.TenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to find subscription: %w", err)
	}

	if subscription == nil {
		return nil, common.NewNotFoundError("subscription", input.TenantID.String())
	}

	// Get the Stripe Customer ID from the subscription
	customerID := subscription.StripeCustomerID()
	if customerID == "" {
		return nil, common.NewValidationError("Stripe顧客情報が登録されていません", nil)
	}

	// Create the billing portal session
	result, err := uc.stripeClient.CreateBillingPortalSession(infrastripe.BillingPortalParams{
		CustomerID: customerID,
		ReturnURL:  uc.returnURL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create billing portal session: %w", err)
	}

	return &BillingPortalOutput{
		PortalURL: result.URL,
	}, nil
}
