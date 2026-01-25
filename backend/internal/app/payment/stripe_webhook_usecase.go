package payment

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/billing"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/services"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/tenant"
)

// StripeEvent represents a Stripe webhook event
type StripeEvent struct {
	ID   string          `json:"id"`
	Type string          `json:"type"`
	Data StripeEventData `json:"data"`
}

// StripeEventData represents the data object in a Stripe event
type StripeEventData struct {
	Object json.RawMessage `json:"object"`
}

// StripeInvoice represents a Stripe invoice object
type StripeInvoice struct {
	ID           string `json:"id"`
	Customer     string `json:"customer"`
	Subscription string `json:"subscription"`
	Status       string `json:"status"`
	Paid         bool   `json:"paid"`
}

// StripeSubscription represents a Stripe subscription object
type StripeSubscription struct {
	ID                string `json:"id"`
	Customer          string `json:"customer"`
	Status            string `json:"status"`
	CurrentPeriodEnd  int64  `json:"current_period_end"`
	CancelAtPeriodEnd bool   `json:"cancel_at_period_end"`
	CancelAt          int64  `json:"cancel_at"` // Unix timestamp, 0 if not set
}

// StripeCheckoutSession represents a Stripe Checkout Session object
type StripeCheckoutSession struct {
	ID           string            `json:"id"`
	Customer     string            `json:"customer"`
	Subscription string            `json:"subscription"`
	Status       string            `json:"status"`
	Mode         string            `json:"mode"`
	Metadata     map[string]string `json:"metadata"`
}

// StripeCheckoutSubscription represents Stripe subscription details from checkout
type StripeCheckoutSubscription struct {
	ID               string `json:"id"`
	Customer         string `json:"customer"`
	Status           string `json:"status"`
	CurrentPeriodEnd int64  `json:"current_period_end"`
}

// StripeWebhookUsecase handles Stripe webhook events.
//
// # Aggregate Coordination
//
// This usecase coordinates updates between two independent aggregates:
// Tenant and Subscription. While they are separate domain concepts,
// Stripe webhook events often require updating both within a single transaction.
//
// The coordination pattern is:
//  1. Receive Stripe webhook event
//  2. Begin transaction
//  3. Update Subscription state (if applicable)
//  4. Update Tenant state (if applicable)
//  5. Create audit log
//  6. Commit transaction
//
// This ensures consistency between Subscription and Tenant states.
// If any step fails, the entire transaction is rolled back.
//
// See also: domain/billing/subscription.go for aggregate relationship documentation.
type StripeWebhookUsecase struct {
	txManager        services.TxManager
	tenantRepo       tenant.TenantRepository
	subscriptionRepo billing.SubscriptionRepository
	entitlementRepo  billing.EntitlementRepository
	webhookEventRepo billing.WebhookEventRepository
	auditLogRepo     billing.BillingAuditLogRepository
}

// NewStripeWebhookUsecase creates a new StripeWebhookUsecase
func NewStripeWebhookUsecase(
	txManager services.TxManager,
	tenantRepo tenant.TenantRepository,
	subscriptionRepo billing.SubscriptionRepository,
	entitlementRepo billing.EntitlementRepository,
	webhookEventRepo billing.WebhookEventRepository,
	auditLogRepo billing.BillingAuditLogRepository,
) *StripeWebhookUsecase {
	return &StripeWebhookUsecase{
		txManager:        txManager,
		tenantRepo:       tenantRepo,
		subscriptionRepo: subscriptionRepo,
		entitlementRepo:  entitlementRepo,
		webhookEventRepo: webhookEventRepo,
		auditLogRepo:     auditLogRepo,
	}
}

// HandleWebhook processes a Stripe webhook event
// Returns (processed bool, error)
// processed=false means the event was already processed (duplicate)
func (uc *StripeWebhookUsecase) HandleWebhook(ctx context.Context, event StripeEvent, rawPayload string) (bool, error) {
	now := time.Now().UTC()

	// Try to insert the event for idempotency
	isNew, err := uc.webhookEventRepo.TryInsert(ctx, "stripe", event.ID, &rawPayload)
	if err != nil {
		return false, fmt.Errorf("failed to check webhook idempotency: %w", err)
	}

	if !isNew {
		// Event was already processed
		log.Printf("[Stripe Webhook] Duplicate event ignored: %s", event.ID)
		return false, nil
	}

	// Process based on event type
	switch event.Type {
	case "checkout.session.completed":
		return true, uc.handleCheckoutSessionCompleted(ctx, now, event)
	case "invoice.paid":
		return true, uc.handleInvoicePaid(ctx, now, event)
	case "invoice.payment_failed":
		return true, uc.handleInvoicePaymentFailed(ctx, now, event)
	case "customer.subscription.updated":
		return true, uc.handleSubscriptionUpdated(ctx, now, event)
	case "customer.subscription.deleted":
		return true, uc.handleSubscriptionDeleted(ctx, now, event)
	default:
		// Unknown event type - log and ignore
		log.Printf("[Stripe Webhook] Unknown event type: %s", event.Type)
		return true, nil
	}
}

// handleCheckoutSessionCompleted handles checkout.session.completed events
// Activates tenant, creates subscription and entitlement records
func (uc *StripeWebhookUsecase) handleCheckoutSessionCompleted(ctx context.Context, now time.Time, event StripeEvent) error {
	var session StripeCheckoutSession
	if err := json.Unmarshal(event.Data.Object, &session); err != nil {
		return fmt.Errorf("failed to parse checkout session: %w", err)
	}

	// Only handle subscription mode
	if session.Mode != "subscription" {
		log.Printf("[Stripe Webhook] Ignoring non-subscription checkout: %s", session.ID)
		return nil
	}

	// Find tenant by session ID
	t, err := uc.tenantRepo.FindByPendingStripeSessionID(ctx, session.ID)
	if err != nil {
		if common.IsNotFoundError(err) {
			log.Printf("[Stripe Webhook] No tenant found for session: %s", session.ID)
			return nil
		}
		return fmt.Errorf("failed to find tenant by session ID: %w", err)
	}

	// Verify tenant is in pending_payment status
	if t.Status() != tenant.TenantStatusPendingPayment {
		log.Printf("[Stripe Webhook] Tenant %s is not in pending_payment status: %s", t.TenantID(), t.Status())
		return nil
	}

	return uc.txManager.WithTx(ctx, func(txCtx context.Context) error {
		// Activate tenant
		previousStatus := t.Status()
		if err := t.SetStatusActive(now); err != nil {
			return fmt.Errorf("failed to set tenant status to active: %w", err)
		}
		if err := uc.tenantRepo.Save(txCtx, t); err != nil {
			return fmt.Errorf("failed to save tenant: %w", err)
		}

		// Create subscription record
		sub, err := billing.NewSubscription(
			now,
			t.TenantID(),
			session.Customer,
			session.Subscription,
			billing.SubscriptionStatusActive,
			nil, // current_period_end will be updated by invoice.paid webhook
		)
		if err != nil {
			return fmt.Errorf("failed to create subscription: %w", err)
		}
		if err := uc.subscriptionRepo.Save(txCtx, sub); err != nil {
			return fmt.Errorf("failed to save subscription: %w", err)
		}

		// Create entitlement record for SUB_200 plan
		entitlement, err := billing.NewEntitlement(
			now,
			t.TenantID(),
			"SUB_200",
			billing.EntitlementSourceStripe,
			nil, // No fixed end date for subscription entitlement
		)
		if err != nil {
			return fmt.Errorf("failed to create entitlement: %w", err)
		}
		if err := uc.entitlementRepo.Save(txCtx, entitlement); err != nil {
			return fmt.Errorf("failed to save entitlement: %w", err)
		}

		// Create audit log
		tenantIDStr := t.TenantID().String()
		afterJSON := fmt.Sprintf(`{"status":"active","previous_status":"%s","session_id":"%s","subscription_id":"%s"}`,
			previousStatus, session.ID, session.Subscription)
		auditLog, err := billing.NewBillingAuditLog(
			now,
			billing.ActorTypeStripe,
			nil,
			string(billing.BillingAuditActionSubscriptionCreate),
			strPtr("tenant"),
			&tenantIDStr,
			nil,
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

// handleInvoicePaid handles invoice.paid events
// Sets tenant to active and updates subscription
func (uc *StripeWebhookUsecase) handleInvoicePaid(ctx context.Context, now time.Time, event StripeEvent) error {
	var invoice StripeInvoice
	if err := json.Unmarshal(event.Data.Object, &invoice); err != nil {
		return fmt.Errorf("failed to parse invoice: %w", err)
	}

	if invoice.Subscription == "" {
		// Not a subscription invoice, ignore
		return nil
	}

	// Find subscription by Stripe ID
	sub, err := uc.subscriptionRepo.FindByStripeSubscriptionID(ctx, invoice.Subscription)
	if err != nil {
		return fmt.Errorf("failed to find subscription: %w", err)
	}
	if sub == nil {
		// Subscription not found, might be a new subscription
		log.Printf("[Stripe Webhook] Subscription not found: %s", invoice.Subscription)
		return nil
	}

	return uc.txManager.WithTx(ctx, func(txCtx context.Context) error {
		// Update subscription status
		if err := sub.UpdateStatus(now, billing.SubscriptionStatusActive, sub.CurrentPeriodEnd()); err != nil {
			return fmt.Errorf("failed to update subscription status: %w", err)
		}
		if err := uc.subscriptionRepo.Save(txCtx, sub); err != nil {
			return fmt.Errorf("failed to save subscription: %w", err)
		}

		// Update tenant status to active
		t, err := uc.tenantRepo.FindByID(txCtx, sub.TenantID())
		if err != nil {
			return fmt.Errorf("failed to find tenant: %w", err)
		}

		previousStatus := t.Status()
		if err := t.SetStatusActive(now); err != nil {
			return fmt.Errorf("failed to set tenant status to active: %w", err)
		}
		if err := uc.tenantRepo.Save(txCtx, t); err != nil {
			return fmt.Errorf("failed to save tenant: %w", err)
		}

		// Create audit log
		tenantIDStr := t.TenantID().String()
		afterJSON := fmt.Sprintf(`{"status":"active","previous_status":"%s"}`, previousStatus)
		auditLog, err := billing.NewBillingAuditLog(
			now,
			billing.ActorTypeStripe,
			nil,
			string(billing.BillingAuditActionPaymentSucceeded),
			strPtr("tenant"),
			&tenantIDStr,
			nil,
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

// handleInvoicePaymentFailed handles invoice.payment_failed events
// Sets tenant to grace period
func (uc *StripeWebhookUsecase) handleInvoicePaymentFailed(ctx context.Context, now time.Time, event StripeEvent) error {
	var invoice StripeInvoice
	if err := json.Unmarshal(event.Data.Object, &invoice); err != nil {
		return fmt.Errorf("failed to parse invoice: %w", err)
	}

	if invoice.Subscription == "" {
		return nil
	}

	sub, err := uc.subscriptionRepo.FindByStripeSubscriptionID(ctx, invoice.Subscription)
	if err != nil {
		return fmt.Errorf("failed to find subscription: %w", err)
	}
	if sub == nil {
		log.Printf("[Stripe Webhook] Subscription not found: %s", invoice.Subscription)
		return nil
	}

	return uc.txManager.WithTx(ctx, func(txCtx context.Context) error {
		// Update subscription status
		if err := sub.UpdateStatus(now, billing.SubscriptionStatusPastDue, sub.CurrentPeriodEnd()); err != nil {
			return fmt.Errorf("failed to update subscription status: %w", err)
		}
		if err := uc.subscriptionRepo.Save(txCtx, sub); err != nil {
			return fmt.Errorf("failed to save subscription: %w", err)
		}

		// Update tenant status to grace
		t, err := uc.tenantRepo.FindByID(txCtx, sub.TenantID())
		if err != nil {
			return fmt.Errorf("failed to find tenant: %w", err)
		}

		// 支払い失敗時はドメイン層で定義されたgrace期間を使用
		graceUntil := now.AddDate(0, 0, tenant.DefaultGracePeriodDays)
		previousStatus := t.Status()
		if err := t.SetStatusGrace(now, graceUntil); err != nil {
			return fmt.Errorf("failed to set tenant status to grace: %w", err)
		}
		if err := uc.tenantRepo.Save(txCtx, t); err != nil {
			return fmt.Errorf("failed to save tenant: %w", err)
		}

		// Create audit log
		tenantIDStr := t.TenantID().String()
		afterJSON := fmt.Sprintf(`{"status":"grace","grace_until":"%s","previous_status":"%s"}`,
			graceUntil.Format(time.RFC3339), previousStatus)
		auditLog, err := billing.NewBillingAuditLog(
			now,
			billing.ActorTypeStripe,
			nil,
			string(billing.BillingAuditActionPaymentFailed),
			strPtr("tenant"),
			&tenantIDStr,
			nil,
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

// handleSubscriptionUpdated handles customer.subscription.updated events
// Updates cancel_at_period_end flag when subscription is scheduled to cancel
func (uc *StripeWebhookUsecase) handleSubscriptionUpdated(ctx context.Context, now time.Time, event StripeEvent) error {
	var subscription StripeSubscription
	if err := json.Unmarshal(event.Data.Object, &subscription); err != nil {
		return fmt.Errorf("failed to parse subscription: %w", err)
	}

	sub, err := uc.subscriptionRepo.FindByStripeSubscriptionID(ctx, subscription.ID)
	if err != nil {
		return fmt.Errorf("failed to find subscription: %w", err)
	}
	if sub == nil {
		log.Printf("[Stripe Webhook] Subscription not found for update: %s", subscription.ID)
		return nil
	}

	// cancel_at_period_end フラグを更新
	var cancelAt *time.Time
	if subscription.CancelAt > 0 {
		t := time.Unix(subscription.CancelAt, 0).UTC()
		cancelAt = &t
	}

	sub.SetCancelAtPeriodEnd(now, subscription.CancelAtPeriodEnd, cancelAt)

	if err := uc.subscriptionRepo.Save(ctx, sub); err != nil {
		return fmt.Errorf("failed to save subscription: %w", err)
	}

	if subscription.CancelAtPeriodEnd {
		log.Printf("[Stripe Webhook] Subscription %s is now scheduled to cancel at period end", subscription.ID)
	} else {
		log.Printf("[Stripe Webhook] Subscription %s cancel schedule was removed", subscription.ID)
	}

	return nil
}

// handleSubscriptionDeleted handles customer.subscription.deleted events
// Marks tenant as suspended (or schedules it via grace_until check)
func (uc *StripeWebhookUsecase) handleSubscriptionDeleted(ctx context.Context, now time.Time, event StripeEvent) error {
	var subscription StripeSubscription
	if err := json.Unmarshal(event.Data.Object, &subscription); err != nil {
		return fmt.Errorf("failed to parse subscription: %w", err)
	}

	sub, err := uc.subscriptionRepo.FindByStripeSubscriptionID(ctx, subscription.ID)
	if err != nil {
		return fmt.Errorf("failed to find subscription: %w", err)
	}
	if sub == nil {
		log.Printf("[Stripe Webhook] Subscription not found: %s", subscription.ID)
		return nil
	}

	return uc.txManager.WithTx(ctx, func(txCtx context.Context) error {
		// Update subscription status
		periodEnd := time.Unix(subscription.CurrentPeriodEnd, 0).UTC()
		if err := sub.UpdateStatus(now, billing.SubscriptionStatusCanceled, &periodEnd); err != nil {
			return fmt.Errorf("failed to update subscription status: %w", err)
		}
		if err := uc.subscriptionRepo.Save(txCtx, sub); err != nil {
			return fmt.Errorf("failed to save subscription: %w", err)
		}

		// Check if period has already ended
		t, err := uc.tenantRepo.FindByID(txCtx, sub.TenantID())
		if err != nil {
			if common.IsNotFoundError(err) {
				return nil
			}
			return fmt.Errorf("failed to find tenant: %w", err)
		}

		previousStatus := t.Status()

		// サブスクリプション終了後、grace期間を設定
		// customer.subscription.deleted イベントは期間終了時に発火するため、
		// ドメイン層で定義されたビジネスルール（DefaultGracePeriodDays = 14日）に従って
		// grace_until を計算する
		//
		// 例: 1/31に期間終了 → grace_until = 2/14 → 2/15にsuspended
		if err := t.TransitionToGraceAfterSubscriptionEnd(now, periodEnd); err != nil {
			return fmt.Errorf("failed to transition tenant to grace: %w", err)
		}

		if err := uc.tenantRepo.Save(txCtx, t); err != nil {
			return fmt.Errorf("failed to save tenant: %w", err)
		}

		// Create audit log
		tenantIDStr := t.TenantID().String()
		afterJSON := fmt.Sprintf(`{"status":"%s","previous_status":"%s","period_end":"%s"}`,
			t.Status(), previousStatus, periodEnd.Format(time.RFC3339))
		auditLog, err := billing.NewBillingAuditLog(
			now,
			billing.ActorTypeStripe,
			nil,
			string(billing.BillingAuditActionSubscriptionUpdate),
			strPtr("tenant"),
			&tenantIDStr,
			nil,
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
