package services

// Checkout session expiration constants (Stripe API constraints)
const (
	// MinCheckoutExpireMinutes is the minimum expiration time for Stripe checkout sessions (30 minutes)
	MinCheckoutExpireMinutes = 30

	// MaxCheckoutExpireMinutes is the maximum expiration time for Stripe checkout sessions (24 hours)
	MaxCheckoutExpireMinutes = 1440

	// DefaultCheckoutExpireMinutes is the default expiration time for checkout sessions (24 hours)
	DefaultCheckoutExpireMinutes = 1440
)

// CheckoutSessionParams contains parameters for creating a checkout session
type CheckoutSessionParams struct {
	PriceID       string
	CustomerEmail string
	SuccessURL    string
	CancelURL     string
	TenantID      string
	TenantName    string
	ExpireMinutes int // Optional: minutes until session expires (default: 1440 = 24 hours, min: 30, max: 1440)
}

// CheckoutSessionResult contains the result of creating a checkout session
type CheckoutSessionResult struct {
	SessionID  string
	URL        string
	ExpiresAt  int64
	CustomerID string
}

// BillingPortalParams contains parameters for creating a billing portal session
type BillingPortalParams struct {
	CustomerID string
	ReturnURL  string
}

// BillingPortalResult contains the result of creating a billing portal session
type BillingPortalResult struct {
	URL string
}

// PaymentGateway defines the interface for payment provider operations.
// This abstraction allows the application layer to be independent of specific
// payment provider implementations (e.g., Stripe, PayPal).
//
// Implementations:
//   - infra/stripe.StripePaymentGateway: Stripe implementation
type PaymentGateway interface {
	// CreateCheckoutSession creates a new checkout session for subscription payment.
	// Returns the session details including the URL to redirect the user.
	CreateCheckoutSession(params CheckoutSessionParams) (*CheckoutSessionResult, error)

	// CreateBillingPortalSession creates a customer portal session.
	// The portal allows customers to manage their subscription, update payment methods, and view invoices.
	CreateBillingPortalSession(params BillingPortalParams) (*BillingPortalResult, error)
}
