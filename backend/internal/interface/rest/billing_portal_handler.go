package rest

import (
	"log"
	"net/http"

	apppayment "github.com/erenoa/vrc-shift-scheduler/backend/internal/app/payment"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	infrastripe "github.com/erenoa/vrc-shift-scheduler/backend/internal/infra/stripe"
)

// BillingHandler handles billing-related requests
type BillingHandler struct {
	billingPortalUsecase *apppayment.BillingPortalUsecase
	billingStatusUsecase *apppayment.BillingStatusUsecase
}

// NewBillingHandler creates a new BillingHandler
func NewBillingHandler(
	billingPortalUsecase *apppayment.BillingPortalUsecase,
	billingStatusUsecase *apppayment.BillingStatusUsecase,
) *BillingHandler {
	return &BillingHandler{
		billingPortalUsecase: billingPortalUsecase,
		billingStatusUsecase: billingStatusUsecase,
	}
}

// GetStatus handles GET /api/v1/billing/status
// Returns the current billing status for the tenant
func (h *BillingHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	// Get tenant ID from context
	tenantID, ok := r.Context().Value(ContextKeyTenantID).(common.TenantID)
	if !ok {
		RespondError(w, http.StatusUnauthorized, "ERR_UNAUTHORIZED", "認証が必要です", nil)
		return
	}

	// Execute usecase
	output, err := h.billingStatusUsecase.Execute(r.Context(), apppayment.BillingStatusInput{
		TenantID: tenantID,
	})
	if err != nil {
		log.Printf("[ERROR] Failed to get billing status: %v", err)
		RespondInternalError(w)
		return
	}

	// Return status
	RespondJSON(w, http.StatusOK, map[string]interface{}{
		"data": output,
	})
}

// CreatePortalSession handles POST /api/v1/billing/portal
// Creates a Stripe Customer Portal session and returns the URL
func (h *BillingHandler) CreatePortalSession(w http.ResponseWriter, r *http.Request) {
	// Get tenant ID from context
	tenantID, ok := r.Context().Value(ContextKeyTenantID).(common.TenantID)
	if !ok {
		RespondError(w, http.StatusUnauthorized, "ERR_UNAUTHORIZED", "認証が必要です", nil)
		return
	}

	// Execute usecase
	output, err := h.billingPortalUsecase.Execute(r.Context(), apppayment.BillingPortalInput{
		TenantID: tenantID,
	})
	if err != nil {
		// Check for not found error
		if common.IsNotFoundError(err) {
			RespondError(w, http.StatusNotFound, "ERR_NOT_FOUND", "サブスクリプションが見つかりません。月額プランをご利用でない場合は、この機能は使用できません。", nil)
			return
		}

		// Check for domain error
		if domainErr, ok := err.(*common.DomainError); ok {
			RespondError(w, http.StatusBadRequest, domainErr.Code(), domainErr.Message, nil)
			return
		}

		// Check for Stripe error
		if stripeErr := infrastripe.GetStripeError(err); stripeErr != nil {
			log.Printf("[ERROR] Stripe error creating portal session: %s - %v", stripeErr.Code, stripeErr.Err)
			RespondError(w, http.StatusBadGateway, stripeErr.Code, stripeErr.Message, nil)
			return
		}

		log.Printf("[ERROR] Failed to create billing portal session: %v", err)
		RespondInternalError(w)
		return
	}

	// Return portal URL
	RespondJSON(w, http.StatusOK, map[string]interface{}{
		"data": map[string]string{
			"portal_url": output.PortalURL,
		},
	})
}
