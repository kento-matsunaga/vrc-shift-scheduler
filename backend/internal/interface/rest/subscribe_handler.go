package rest

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	apppayment "github.com/erenoa/vrc-shift-scheduler/backend/internal/app/payment"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	infrastripe "github.com/erenoa/vrc-shift-scheduler/backend/internal/infra/stripe"
)

// SubscribeHandler handles subscription requests via Stripe Checkout
type SubscribeHandler struct {
	subscribeUsecase *apppayment.SubscribeUsecase
	rateLimiter      *RateLimiter
}

// NewSubscribeHandler creates a new SubscribeHandler
func NewSubscribeHandler(subscribeUsecase *apppayment.SubscribeUsecase, rateLimiter *RateLimiter) *SubscribeHandler {
	return &SubscribeHandler{
		subscribeUsecase: subscribeUsecase,
		rateLimiter:      rateLimiter,
	}
}

// SubscribeRequest represents the request body for subscribing
type SubscribeRequest struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	TenantName  string `json:"tenant_name"`
	DisplayName string `json:"display_name"`
	Timezone    string `json:"timezone"`
}

// SubscribeResponse represents the response for a successful subscribe request
type SubscribeResponse struct {
	CheckoutURL string `json:"checkout_url"`
	SessionID   string `json:"session_id"`
	TenantID    string `json:"tenant_id"`
	ExpiresAt   int64  `json:"expires_at"`
	Message     string `json:"message"`
}

// Subscribe handles POST /api/v1/public/subscribe
func (h *SubscribeHandler) Subscribe(w http.ResponseWriter, r *http.Request) {
	// Extract client IP for rate limiting
	clientIP := getClientIP(r)

	// Check rate limit (5 requests per minute)
	if !h.rateLimiter.Allow(clientIP) {
		// Delay response to slow down attackers
		time.Sleep(1 * time.Second)
		RespondError(w, http.StatusTooManyRequests, "ERR_RATE_LIMITED",
			"リクエストが多すぎます。しばらくしてから再度お試しください。", nil)
		return
	}

	// Parse request body
	var req SubscribeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondBadRequest(w, "Invalid request body")
		return
	}

	// Validate required fields
	if req.Email == "" || req.Password == "" || req.DisplayName == "" || req.TenantName == "" {
		RespondBadRequest(w, "すべての項目を入力してください")
		return
	}

	// Default timezone if not provided
	if req.Timezone == "" {
		req.Timezone = "Asia/Tokyo"
	}

	// Execute subscribe usecase
	input := apppayment.SubscribeInput{
		Email:       req.Email,
		Password:    req.Password,
		TenantName:  req.TenantName,
		Timezone:    req.Timezone,
		DisplayName: req.DisplayName,
	}

	output, err := h.subscribeUsecase.Execute(r.Context(), input)
	if err != nil {
		// Delay response for failed attempts to prevent timing attacks
		time.Sleep(500 * time.Millisecond)

		// Check for validation errors
		if domainErr, ok := err.(*common.DomainError); ok {
			RespondError(w, http.StatusBadRequest, domainErr.Code(), domainErr.Message, nil)
			return
		}

		// Check for Stripe errors
		if stripeErr := infrastripe.GetStripeError(err); stripeErr != nil {
			log.Printf("[ERROR] Stripe error: %s - %v", stripeErr.Code, stripeErr.Err)
			RespondError(w, http.StatusBadGateway, stripeErr.Code, stripeErr.Message, nil)
			return
		}

		// Log the actual error for debugging
		log.Printf("[ERROR] Subscribe failed: %v", err)
		RespondInternalError(w)
		return
	}

	// Success response
	response := SubscribeResponse{
		CheckoutURL: output.CheckoutURL,
		SessionID:   output.SessionID,
		TenantID:    output.TenantID,
		ExpiresAt:   output.ExpiresAt,
		Message:     "決済ページにリダイレクトします。",
	}

	RespondJSON(w, http.StatusCreated, map[string]interface{}{
		"data": response,
	})
}
