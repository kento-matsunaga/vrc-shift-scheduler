package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/billing"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// SubscriptionRepository implements billing.SubscriptionRepository for PostgreSQL
type SubscriptionRepository struct {
	db *pgxpool.Pool
}

// NewSubscriptionRepository creates a new SubscriptionRepository
func NewSubscriptionRepository(db *pgxpool.Pool) *SubscriptionRepository {
	return &SubscriptionRepository{db: db}
}

// Save saves a subscription
func (r *SubscriptionRepository) Save(ctx context.Context, s *billing.Subscription) error {
	query := `
		INSERT INTO subscriptions (
			subscription_id, tenant_id, stripe_customer_id, stripe_subscription_id,
			status, current_period_end, cancel_at_period_end, cancel_at, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (subscription_id) DO UPDATE SET
			status = EXCLUDED.status,
			current_period_end = EXCLUDED.current_period_end,
			cancel_at_period_end = EXCLUDED.cancel_at_period_end,
			cancel_at = EXCLUDED.cancel_at,
			updated_at = EXCLUDED.updated_at
	`

	_, err := r.db.Exec(ctx, query,
		s.SubscriptionID().String(),
		s.TenantID().String(),
		s.StripeCustomerID(),
		s.StripeSubscriptionID(),
		s.Status().String(),
		s.CurrentPeriodEnd(),
		s.CancelAtPeriodEnd(),
		s.CancelAt(),
		s.CreatedAt(),
		s.UpdatedAt(),
	)

	if err != nil {
		return fmt.Errorf("failed to save subscription: %w", err)
	}

	return nil
}

// FindByTenantID finds a subscription by tenant ID
func (r *SubscriptionRepository) FindByTenantID(ctx context.Context, tenantID common.TenantID) (*billing.Subscription, error) {
	query := `
		SELECT
			subscription_id, tenant_id, stripe_customer_id, stripe_subscription_id,
			status, current_period_end, cancel_at_period_end, cancel_at, created_at, updated_at
		FROM subscriptions
		WHERE tenant_id = $1
	`

	return r.scanSubscription(ctx, query, tenantID.String())
}

// FindByStripeSubscriptionID finds a subscription by Stripe subscription ID
func (r *SubscriptionRepository) FindByStripeSubscriptionID(ctx context.Context, stripeSubID string) (*billing.Subscription, error) {
	query := `
		SELECT
			subscription_id, tenant_id, stripe_customer_id, stripe_subscription_id,
			status, current_period_end, cancel_at_period_end, cancel_at, created_at, updated_at
		FROM subscriptions
		WHERE stripe_subscription_id = $1
	`

	return r.scanSubscription(ctx, query, stripeSubID)
}

func (r *SubscriptionRepository) scanSubscription(ctx context.Context, query string, args ...interface{}) (*billing.Subscription, error) {
	var (
		subscriptionIDStr    string
		tenantIDStr          string
		stripeCustomerID     string
		stripeSubscriptionID string
		status               string
		currentPeriodEnd     sql.NullTime
		cancelAtPeriodEnd    bool
		cancelAt             sql.NullTime
		createdAt            time.Time
		updatedAt            time.Time
	)

	err := r.db.QueryRow(ctx, query, args...).Scan(
		&subscriptionIDStr,
		&tenantIDStr,
		&stripeCustomerID,
		&stripeSubscriptionID,
		&status,
		&currentPeriodEnd,
		&cancelAtPeriodEnd,
		&cancelAt,
		&createdAt,
		&updatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find subscription: %w", err)
	}

	subscriptionID, err := billing.ParseSubscriptionID(subscriptionIDStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse subscription_id: %w", err)
	}

	tenantID, err := common.ParseTenantID(tenantIDStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse tenant_id: %w", err)
	}

	var currentPeriodEndPtr *time.Time
	if currentPeriodEnd.Valid {
		currentPeriodEndPtr = &currentPeriodEnd.Time
	}

	var cancelAtPtr *time.Time
	if cancelAt.Valid {
		cancelAtPtr = &cancelAt.Time
	}

	return billing.ReconstructSubscription(
		subscriptionID,
		tenantID,
		stripeCustomerID,
		stripeSubscriptionID,
		billing.SubscriptionStatus(status),
		currentPeriodEndPtr,
		cancelAtPeriodEnd,
		cancelAtPtr,
		createdAt,
		updatedAt,
	)
}
