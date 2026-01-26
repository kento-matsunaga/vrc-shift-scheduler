package rest

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	apppayment "github.com/erenoa/vrc-shift-scheduler/backend/internal/app/payment"
)

// =====================================================
// Tests for verifySignature
// =====================================================

func TestVerifySignature_ValidSignature(t *testing.T) {
	secret := "whsec_test_secret_12345"
	handler := &StripeWebhookHandler{
		webhookSecret: secret,
		enabled:       true,
	}

	payload := []byte(`{"id":"evt_123","type":"checkout.session.completed"}`)
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)

	// Generate valid signature
	signedPayload := timestamp + "." + string(payload)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(signedPayload))
	signature := hex.EncodeToString(mac.Sum(nil))

	sigHeader := "t=" + timestamp + ",v1=" + signature

	if !handler.verifySignature(payload, sigHeader) {
		t.Error("verifySignature should return true for valid signature")
	}
}

func TestVerifySignature_InvalidSignature(t *testing.T) {
	secret := "whsec_test_secret_12345"
	handler := &StripeWebhookHandler{
		webhookSecret: secret,
		enabled:       true,
	}

	payload := []byte(`{"id":"evt_123","type":"checkout.session.completed"}`)
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)

	// Use wrong signature
	sigHeader := "t=" + timestamp + ",v1=invalid_signature_abc123"

	if handler.verifySignature(payload, sigHeader) {
		t.Error("verifySignature should return false for invalid signature")
	}
}

func TestVerifySignature_WrongSecret(t *testing.T) {
	handler := &StripeWebhookHandler{
		webhookSecret: "whsec_correct_secret",
		enabled:       true,
	}

	payload := []byte(`{"id":"evt_123","type":"checkout.session.completed"}`)
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)

	// Generate signature with wrong secret
	signedPayload := timestamp + "." + string(payload)
	mac := hmac.New(sha256.New, []byte("whsec_wrong_secret"))
	mac.Write([]byte(signedPayload))
	signature := hex.EncodeToString(mac.Sum(nil))

	sigHeader := "t=" + timestamp + ",v1=" + signature

	if handler.verifySignature(payload, sigHeader) {
		t.Error("verifySignature should return false when signed with wrong secret")
	}
}

func TestVerifySignature_ExpiredTimestamp(t *testing.T) {
	secret := "whsec_test_secret_12345"
	handler := &StripeWebhookHandler{
		webhookSecret: secret,
		enabled:       true,
	}

	payload := []byte(`{"id":"evt_123","type":"checkout.session.completed"}`)
	// Timestamp 10 minutes ago (exceeds 5 minute tolerance)
	timestamp := strconv.FormatInt(time.Now().Add(-10*time.Minute).Unix(), 10)

	// Generate valid signature with expired timestamp
	signedPayload := timestamp + "." + string(payload)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(signedPayload))
	signature := hex.EncodeToString(mac.Sum(nil))

	sigHeader := "t=" + timestamp + ",v1=" + signature

	if handler.verifySignature(payload, sigHeader) {
		t.Error("verifySignature should return false for expired timestamp (>5 minutes)")
	}
}

func TestVerifySignature_MissingTimestamp(t *testing.T) {
	handler := &StripeWebhookHandler{
		webhookSecret: "whsec_test_secret_12345",
		enabled:       true,
	}

	payload := []byte(`{"id":"evt_123"}`)
	sigHeader := "v1=some_signature"

	if handler.verifySignature(payload, sigHeader) {
		t.Error("verifySignature should return false when timestamp is missing")
	}
}

func TestVerifySignature_MissingSignature(t *testing.T) {
	handler := &StripeWebhookHandler{
		webhookSecret: "whsec_test_secret_12345",
		enabled:       true,
	}

	payload := []byte(`{"id":"evt_123"}`)
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	sigHeader := "t=" + timestamp

	if handler.verifySignature(payload, sigHeader) {
		t.Error("verifySignature should return false when signature is missing")
	}
}

func TestVerifySignature_EmptyHeader(t *testing.T) {
	handler := &StripeWebhookHandler{
		webhookSecret: "whsec_test_secret_12345",
		enabled:       true,
	}

	payload := []byte(`{"id":"evt_123"}`)

	if handler.verifySignature(payload, "") {
		t.Error("verifySignature should return false for empty header")
	}
}

func TestVerifySignature_MalformedHeader(t *testing.T) {
	handler := &StripeWebhookHandler{
		webhookSecret: "whsec_test_secret_12345",
		enabled:       true,
	}

	payload := []byte(`{"id":"evt_123"}`)

	testCases := []struct {
		name      string
		sigHeader string
	}{
		{"no equals sign", "t1234567890,v1abc123"},
		{"invalid timestamp format", "t=notanumber,v1=abc123"},
		{"empty values", "t=,v1="},
		{"just commas", ",,,"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if handler.verifySignature(payload, tc.sigHeader) {
				t.Errorf("verifySignature should return false for malformed header: %s", tc.sigHeader)
			}
		})
	}
}

func TestVerifySignature_MultipleSignatures(t *testing.T) {
	secret := "whsec_test_secret_12345"
	handler := &StripeWebhookHandler{
		webhookSecret: secret,
		enabled:       true,
	}

	payload := []byte(`{"id":"evt_123","type":"checkout.session.completed"}`)
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)

	// Generate valid signature
	signedPayload := timestamp + "." + string(payload)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(signedPayload))
	validSignature := hex.EncodeToString(mac.Sum(nil))

	// Header with multiple signatures (one invalid, one valid)
	sigHeader := "t=" + timestamp + ",v1=invalid_sig,v1=" + validSignature

	if !handler.verifySignature(payload, sigHeader) {
		t.Error("verifySignature should return true when at least one signature matches")
	}
}

func TestVerifySignature_TimestampAtBoundary(t *testing.T) {
	secret := "whsec_test_secret_12345"
	handler := &StripeWebhookHandler{
		webhookSecret: secret,
		enabled:       true,
	}

	payload := []byte(`{"id":"evt_123"}`)
	// Timestamp exactly 4 minutes 59 seconds ago (should pass)
	timestamp := strconv.FormatInt(time.Now().Add(-4*time.Minute-59*time.Second).Unix(), 10)

	signedPayload := timestamp + "." + string(payload)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(signedPayload))
	signature := hex.EncodeToString(mac.Sum(nil))

	sigHeader := "t=" + timestamp + ",v1=" + signature

	if !handler.verifySignature(payload, sigHeader) {
		t.Error("verifySignature should return true for timestamp within 5 minute tolerance")
	}
}

// =====================================================
// Tests for HandleWebhook
// =====================================================

// MockStripeWebhookUsecase is a mock for testing
type MockStripeWebhookUsecase struct {
	handleWebhookFunc func(event apppayment.StripeEvent, rawBody string) (bool, error)
}

func TestHandleWebhook_MissingSignatureHeader(t *testing.T) {
	handler := &StripeWebhookHandler{
		webhookSecret: "whsec_test_secret",
		enabled:       true,
	}

	body := []byte(`{"id":"evt_123","type":"test"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/stripe/webhook", bytes.NewReader(body))
	// Note: Not setting Stripe-Signature header

	rr := httptest.NewRecorder()
	handler.HandleWebhook(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if errData, ok := response["error"].(map[string]interface{}); ok {
		if errData["code"] != "ERR_MISSING_SIGNATURE" {
			t.Errorf("Expected error code ERR_MISSING_SIGNATURE, got %v", errData["code"])
		}
	} else {
		t.Error("Expected error response with code")
	}
}

func TestHandleWebhook_InvalidSignature(t *testing.T) {
	handler := &StripeWebhookHandler{
		webhookSecret: "whsec_test_secret",
		enabled:       true,
	}

	body := []byte(`{"id":"evt_123","type":"test"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/stripe/webhook", bytes.NewReader(body))
	req.Header.Set("Stripe-Signature", "t=1234567890,v1=invalid_signature")

	rr := httptest.NewRecorder()
	handler.HandleWebhook(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, rr.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if errData, ok := response["error"].(map[string]interface{}); ok {
		if errData["code"] != "ERR_INVALID_SIGNATURE" {
			t.Errorf("Expected error code ERR_INVALID_SIGNATURE, got %v", errData["code"])
		}
	} else {
		t.Error("Expected error response with code")
	}
}

func TestHandleWebhook_DisabledIntegration(t *testing.T) {
	handler := &StripeWebhookHandler{
		webhookSecret: "",
		enabled:       false,
	}

	body := []byte(`{"id":"evt_123","type":"test"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/stripe/webhook", bytes.NewReader(body))
	// No signature header needed when disabled

	rr := httptest.NewRecorder()
	handler.HandleWebhook(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status %d for disabled integration, got %d", http.StatusOK, rr.Code)
	}

	var response map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response["status"] != "ignored" {
		t.Errorf("Expected status 'ignored', got %s", response["status"])
	}
}

// =====================================================
// Helper functions for generating test signatures
// =====================================================

// generateStripeSignature creates a valid Stripe webhook signature for testing
func generateStripeSignature(payload []byte, secret string, timestamp time.Time) string {
	ts := strconv.FormatInt(timestamp.Unix(), 10)
	signedPayload := ts + "." + string(payload)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(signedPayload))
	signature := hex.EncodeToString(mac.Sum(nil))
	return "t=" + ts + ",v1=" + signature
}
