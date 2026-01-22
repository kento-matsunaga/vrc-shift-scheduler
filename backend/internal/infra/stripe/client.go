package stripe

import (
	"errors"

	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/checkout/session"
)

// StripeError represents a Stripe-specific error with localized message
type StripeError struct {
	Code    string
	Message string
	Err     error
}

func (e *StripeError) Error() string {
	return e.Message
}

func (e *StripeError) Unwrap() error {
	return e.Err
}

// IsStripeError checks if an error is a StripeError
func IsStripeError(err error) bool {
	var stripeErr *StripeError
	return errors.As(err, &stripeErr)
}

// GetStripeError extracts StripeError from an error chain
func GetStripeError(err error) *StripeError {
	var stripeErr *StripeError
	if errors.As(err, &stripeErr) {
		return stripeErr
	}
	return nil
}

// wrapStripeError converts Stripe SDK errors to StripeError with Japanese messages
func wrapStripeError(err error) error {
	var stripeErr *stripe.Error
	if errors.As(err, &stripeErr) {
		switch stripeErr.Type {
		case stripe.ErrorTypeAPI:
			return &StripeError{
				Code:    "ERR_STRIPE_API",
				Message: "決済サービスでエラーが発生しました。しばらくしてから再度お試しください。",
				Err:     err,
			}
		case stripe.ErrorTypeCard:
			return &StripeError{
				Code:    "ERR_STRIPE_CARD",
				Message: "カード情報に問題があります。別のカードをお試しください。",
				Err:     err,
			}
		case stripe.ErrorTypeInvalidRequest:
			return &StripeError{
				Code:    "ERR_STRIPE_INVALID",
				Message: "リクエストに問題があります。入力内容をご確認ください。",
				Err:     err,
			}
		case stripe.ErrorTypeIdempotency:
			return &StripeError{
				Code:    "ERR_STRIPE_IDEMPOTENCY",
				Message: "重複したリクエストです。しばらくしてから再度お試しください。",
				Err:     err,
			}
		default:
			return &StripeError{
				Code:    "ERR_STRIPE_UNKNOWN",
				Message: "決済処理中にエラーが発生しました。しばらくしてから再度お試しください。",
				Err:     err,
			}
		}
	}
	// Not a Stripe error, return original
	return err
}

// Client wraps Stripe API operations
type Client struct {
	secretKey string
}

// NewClient creates a new Stripe client
func NewClient(secretKey string) *Client {
	stripe.Key = secretKey
	return &Client{
		secretKey: secretKey,
	}
}

// CheckoutSessionParams contains parameters for creating a checkout session
type CheckoutSessionParams struct {
	PriceID       string
	CustomerEmail string
	SuccessURL    string
	CancelURL     string
	TenantID      string
	TenantName    string
}

// CheckoutSessionResult contains the result of creating a checkout session
type CheckoutSessionResult struct {
	SessionID  string
	URL        string
	ExpiresAt  int64
	CustomerID string
}

// CreateCheckoutSession creates a new Stripe Checkout Session
func (c *Client) CreateCheckoutSession(params CheckoutSessionParams) (*CheckoutSessionResult, error) {
	sessionParams := &stripe.CheckoutSessionParams{
		Mode: stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(params.PriceID),
				Quantity: stripe.Int64(1),
			},
		},
		SuccessURL: stripe.String(params.SuccessURL),
		CancelURL:  stripe.String(params.CancelURL),
		Metadata: map[string]string{
			"tenant_id":   params.TenantID,
			"tenant_name": params.TenantName,
		},
		SubscriptionData: &stripe.CheckoutSessionSubscriptionDataParams{
			Metadata: map[string]string{
				"tenant_id":   params.TenantID,
				"tenant_name": params.TenantName,
			},
		},
	}

	// Set customer email if provided
	if params.CustomerEmail != "" {
		sessionParams.CustomerEmail = stripe.String(params.CustomerEmail)
	}

	// Allow promotion codes
	sessionParams.AllowPromotionCodes = stripe.Bool(true)

	// Set locale to Japanese
	sessionParams.Locale = stripe.String("ja")

	sess, err := session.New(sessionParams)
	if err != nil {
		return nil, wrapStripeError(err)
	}

	var customerID string
	if sess.Customer != nil {
		customerID = sess.Customer.ID
	}

	return &CheckoutSessionResult{
		SessionID:  sess.ID,
		URL:        sess.URL,
		ExpiresAt:  sess.ExpiresAt,
		CustomerID: customerID,
	}, nil
}

// RetrieveCheckoutSession retrieves a checkout session by ID
func (c *Client) RetrieveCheckoutSession(sessionID string) (*stripe.CheckoutSession, error) {
	sess, err := session.Get(sessionID, nil)
	if err != nil {
		return nil, wrapStripeError(err)
	}
	return sess, nil
}
