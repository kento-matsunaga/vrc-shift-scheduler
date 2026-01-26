package stripe

import (
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/services"
)

// StripePaymentGateway implements services.PaymentGateway using Stripe API.
type StripePaymentGateway struct {
	client *Client
}

// NewStripePaymentGateway creates a new StripePaymentGateway.
func NewStripePaymentGateway(client *Client) *StripePaymentGateway {
	return &StripePaymentGateway{
		client: client,
	}
}

// Ensure StripePaymentGateway implements services.PaymentGateway
var _ services.PaymentGateway = (*StripePaymentGateway)(nil)

// CreateCheckoutSession creates a new Stripe Checkout Session.
func (g *StripePaymentGateway) CreateCheckoutSession(params services.CheckoutSessionParams) (*services.CheckoutSessionResult, error) {
	result, err := g.client.CreateCheckoutSession(CheckoutSessionParams{
		PriceID:       params.PriceID,
		CustomerEmail: params.CustomerEmail,
		SuccessURL:    params.SuccessURL,
		CancelURL:     params.CancelURL,
		TenantID:      params.TenantID,
		TenantName:    params.TenantName,
		ExpireMinutes: params.ExpireMinutes,
	})
	if err != nil {
		return nil, err
	}

	return &services.CheckoutSessionResult{
		SessionID:  result.SessionID,
		URL:        result.URL,
		ExpiresAt:  result.ExpiresAt,
		CustomerID: result.CustomerID,
	}, nil
}

// CreateBillingPortalSession creates a Stripe Customer Portal session.
func (g *StripePaymentGateway) CreateBillingPortalSession(params services.BillingPortalParams) (*services.BillingPortalResult, error) {
	result, err := g.client.CreateBillingPortalSession(BillingPortalParams{
		CustomerID: params.CustomerID,
		ReturnURL:  params.ReturnURL,
	})
	if err != nil {
		return nil, err
	}

	return &services.BillingPortalResult{
		URL: result.URL,
	}, nil
}
