package payment

import (
	"context"
	"fmt"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/billing"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/services"
)

// BillingPortalUsecase handles billing portal session creation
type BillingPortalUsecase struct {
	subscriptionRepo billing.SubscriptionRepository
	paymentGateway   services.PaymentGateway
	returnURL        string
}

// NewBillingPortalUsecase creates a new BillingPortalUsecase
func NewBillingPortalUsecase(
	subscriptionRepo billing.SubscriptionRepository,
	paymentGateway services.PaymentGateway,
	returnURL string,
) *BillingPortalUsecase {
	return &BillingPortalUsecase{
		subscriptionRepo: subscriptionRepo,
		paymentGateway:   paymentGateway,
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
	result, err := uc.paymentGateway.CreateBillingPortalSession(services.BillingPortalParams{
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
