package rest

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/application/usecase"
)

// StripeWebhookHandler handles Stripe webhook events
type StripeWebhookHandler struct {
	usecase       *usecase.StripeWebhookUsecase
	webhookSecret string
	enabled       bool
}

// NewStripeWebhookHandler creates a new StripeWebhookHandler
func NewStripeWebhookHandler(uc *usecase.StripeWebhookUsecase) *StripeWebhookHandler {
	secret := os.Getenv("STRIPE_WEBHOOK_SECRET")
	enabled := secret != ""

	if !enabled {
		log.Println("[INFO] Stripe integration disabled (STRIPE_WEBHOOK_SECRET not set)")
	} else {
		log.Println("[INFO] Stripe webhook integration enabled")
	}

	return &StripeWebhookHandler{
		usecase:       uc,
		webhookSecret: secret,
		enabled:       enabled,
	}
}

// HandleWebhook handles POST /api/v1/stripe/webhook
func (h *StripeWebhookHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	if !h.enabled {
		// Stripe integration disabled - return 200 to prevent retries
		log.Println("[WARN] Stripe webhook received but integration is disabled")
		RespondJSON(w, http.StatusOK, map[string]string{"status": "ignored"})
		return
	}

	// Read body
	const maxBodySize = 65536 // 64KB max
	body, err := io.ReadAll(io.LimitReader(r.Body, maxBodySize))
	if err != nil {
		log.Printf("[ERROR] Failed to read webhook body: %v", err)
		RespondBadRequest(w, "Failed to read request body")
		return
	}
	defer r.Body.Close()

	// Verify signature
	sigHeader := r.Header.Get("Stripe-Signature")
	if sigHeader == "" {
		log.Println("[WARN] Missing Stripe-Signature header")
		RespondError(w, http.StatusBadRequest, "ERR_MISSING_SIGNATURE", "Missing signature header", nil)
		return
	}

	if !h.verifySignature(body, sigHeader) {
		log.Println("[WARN] Invalid Stripe signature")
		RespondError(w, http.StatusUnauthorized, "ERR_INVALID_SIGNATURE", "Invalid signature", nil)
		return
	}

	// Parse event
	var event usecase.StripeEvent
	if err := json.Unmarshal(body, &event); err != nil {
		log.Printf("[ERROR] Failed to parse Stripe event: %v", err)
		RespondBadRequest(w, "Invalid event format")
		return
	}

	// Process event
	processed, err := h.usecase.HandleWebhook(r.Context(), event, string(body))
	if err != nil {
		log.Printf("[ERROR] Failed to process Stripe event %s: %v", event.ID, err)
		// Return 500 so Stripe will retry
		RespondInternalError(w)
		return
	}

	if processed {
		log.Printf("[INFO] Stripe event processed: %s (type: %s)", event.ID, event.Type)
	} else {
		log.Printf("[INFO] Stripe event already processed (duplicate): %s", event.ID)
	}

	// Always return 200 for successfully handled events (including duplicates)
	RespondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// verifySignature verifies the Stripe webhook signature
// Implements Stripe's signature verification algorithm:
// https://stripe.com/docs/webhooks/signatures
func (h *StripeWebhookHandler) verifySignature(payload []byte, sigHeader string) bool {
	// Parse signature header
	// Format: t=timestamp,v1=signature1,v1=signature2,...
	var timestamp string
	var signatures []string

	parts := strings.Split(sigHeader, ",")
	for _, part := range parts {
		kv := strings.SplitN(part, "=", 2)
		if len(kv) != 2 {
			continue
		}
		switch kv[0] {
		case "t":
			timestamp = kv[1]
		case "v1":
			signatures = append(signatures, kv[1])
		}
	}

	if timestamp == "" || len(signatures) == 0 {
		return false
	}

	// Check timestamp is within tolerance (5 minutes)
	ts, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return false
	}
	eventTime := time.Unix(ts, 0)
	tolerance := 5 * time.Minute
	if time.Since(eventTime) > tolerance {
		log.Printf("[WARN] Stripe webhook timestamp too old: %v", eventTime)
		return false
	}

	// Compute expected signature
	// signed_payload = timestamp + "." + payload
	signedPayload := timestamp + "." + string(payload)
	mac := hmac.New(sha256.New, []byte(h.webhookSecret))
	mac.Write([]byte(signedPayload))
	expectedSig := hex.EncodeToString(mac.Sum(nil))

	// Check if any of the provided signatures match
	for _, sig := range signatures {
		if hmac.Equal([]byte(sig), []byte(expectedSig)) {
			return true
		}
	}

	return false
}
