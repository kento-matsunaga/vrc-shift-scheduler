package rest

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/application/usecase"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
)

// LicenseClaimHandler handles license claim requests
type LicenseClaimHandler struct {
	claimUsecase *usecase.LicenseClaimUsecase
	rateLimiter  *RateLimiter
}

// NewLicenseClaimHandler creates a new LicenseClaimHandler
func NewLicenseClaimHandler(claimUsecase *usecase.LicenseClaimUsecase, rateLimiter *RateLimiter) *LicenseClaimHandler {
	return &LicenseClaimHandler{
		claimUsecase: claimUsecase,
		rateLimiter:  rateLimiter,
	}
}

// ClaimLicenseRequest represents the request body for claiming a license
type ClaimLicenseRequest struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	DisplayName string `json:"display_name"`
	TenantName  string `json:"tenant_name"`
	LicenseKey  string `json:"license_key"`
}

// ClaimLicenseResponse represents the response for a successful claim
type ClaimLicenseResponse struct {
	TenantID    string `json:"tenant_id"`
	AdminID     string `json:"admin_id"`
	TenantName  string `json:"tenant_name"`
	DisplayName string `json:"display_name"`
	Email       string `json:"email"`
	Message     string `json:"message"`
}

// Claim handles POST /api/v1/public/license/claim
func (h *LicenseClaimHandler) Claim(w http.ResponseWriter, r *http.Request) {
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
	var req ClaimLicenseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondBadRequest(w, "Invalid request body")
		return
	}

	// Validate required fields
	if req.Email == "" || req.Password == "" || req.DisplayName == "" ||
		req.TenantName == "" || req.LicenseKey == "" {
		RespondBadRequest(w, "すべての項目を入力してください")
		return
	}

	// Execute claim usecase
	input := usecase.LicenseClaimInput{
		Email:       req.Email,
		Password:    req.Password,
		DisplayName: req.DisplayName,
		TenantName:  req.TenantName,
		LicenseKey:  req.LicenseKey,
		IPAddress:   clientIP,
		UserAgent:   r.UserAgent(),
	}

	output, err := h.claimUsecase.Execute(r.Context(), input)
	if err != nil {
		// Delay response for failed attempts to prevent timing attacks
		time.Sleep(1 * time.Second)

		// Check for validation errors
		if domainErr, ok := err.(*common.DomainError); ok {
			RespondError(w, http.StatusBadRequest, domainErr.Code(), domainErr.Message, nil)
			return
		}

		RespondInternalError(w)
		return
	}

	// Success response
	response := ClaimLicenseResponse{
		TenantID:    output.TenantID.String(),
		AdminID:     output.AdminID.String(),
		TenantName:  output.TenantName,
		DisplayName: output.DisplayName,
		Email:       output.Email,
		Message:     "登録が完了しました。ログインしてご利用ください。",
	}

	RespondJSON(w, http.StatusCreated, map[string]interface{}{
		"data": response,
	})
}

// getClientIP extracts the client IP address from the request
// Priority: CF-Connecting-IP (Cloudflare) > RemoteAddr
// Note: X-Forwarded-For and X-Real-IP are not trusted as they can be spoofed
func getClientIP(r *http.Request) string {
	// Cloudflare sets CF-Connecting-IP header with the original client IP
	// This header cannot be spoofed when behind Cloudflare
	if cfIP := r.Header.Get("CF-Connecting-IP"); cfIP != "" {
		return cfIP
	}

	// Fall back to RemoteAddr (direct connection or development)
	// RemoteAddr includes port, so extract just the IP
	host := r.RemoteAddr
	if idx := strings.LastIndex(host, ":"); idx != -1 {
		return host[:idx]
	}
	return host
}
