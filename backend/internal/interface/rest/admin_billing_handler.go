package rest

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	appaudit "github.com/erenoa/vrc-shift-scheduler/backend/internal/app/audit"
	applicense "github.com/erenoa/vrc-shift-scheduler/backend/internal/app/license"
	apptenant "github.com/erenoa/vrc-shift-scheduler/backend/internal/app/tenant"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/billing"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/tenant"
	"github.com/go-chi/chi/v5"
)

// AdminBillingHandler handles admin billing endpoints
type AdminBillingHandler struct {
	licenseKeyUsecase *applicense.AdminLicenseKeyUsecase
	tenantUsecase     *apptenant.AdminTenantUsecase
	auditLogUsecase   *appaudit.AdminAuditLogUsecase
}

// NewAdminBillingHandler creates a new AdminBillingHandler
func NewAdminBillingHandler(
	licenseKeyUsecase *applicense.AdminLicenseKeyUsecase,
	tenantUsecase *apptenant.AdminTenantUsecase,
	auditLogUsecase *appaudit.AdminAuditLogUsecase,
) *AdminBillingHandler {
	return &AdminBillingHandler{
		licenseKeyUsecase: licenseKeyUsecase,
		tenantUsecase:     tenantUsecase,
		auditLogUsecase:   auditLogUsecase,
	}
}

// GenerateLicenseKeysRequest represents the request for generating license keys
type GenerateLicenseKeysRequest struct {
	Count     int        `json:"count"`
	ExpiresAt *time.Time `json:"expires_at"`
	Memo      string     `json:"memo"`
}

// getAdminIDFromContext extracts admin ID from context
// In local development, it uses a fixed system admin ID
func getAdminIDFromContext(r *http.Request) common.AdminID {
	// Try to get AdminID from JWT auth context first
	if adminID, ok := r.Context().Value(ContextKeyAdminID).(common.AdminID); ok {
		return adminID
	}

	// For Cloudflare Access / local dev, use a fixed system admin ID
	// The email is stored in context but we use a fixed ID for database compatibility
	// (actor_id column is char(26) for ULID format)
	return common.AdminID("01SYSTEM0ADMIN000000000000")
}

// GenerateLicenseKeys handles POST /api/v1/admin/license-keys
func (h *AdminBillingHandler) GenerateLicenseKeys(w http.ResponseWriter, r *http.Request) {
	adminID := getAdminIDFromContext(r)

	var req GenerateLicenseKeysRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondBadRequest(w, "Invalid request body")
		return
	}

	if req.Count <= 0 {
		req.Count = 1
	}

	input := applicense.GenerateLicenseKeyInput{
		Count:     req.Count,
		ExpiresAt: req.ExpiresAt,
		Memo:      req.Memo,
		AdminID:   adminID,
	}

	output, err := h.licenseKeyUsecase.Generate(r.Context(), input)
	if err != nil {
		if domainErr, ok := err.(*common.DomainError); ok {
			RespondError(w, http.StatusBadRequest, domainErr.Code(), domainErr.Message, nil)
			return
		}
		RespondInternalError(w)
		return
	}

	// Convert output to response
	keys := make([]map[string]interface{}, len(output.Keys))
	for i, k := range output.Keys {
		keys[i] = map[string]interface{}{
			"key_id":     k.KeyID.String(),
			"key":        k.Key,
			"expires_at": k.ExpiresAt,
			"created_at": k.CreatedAt,
		}
	}

	RespondJSON(w, http.StatusCreated, map[string]interface{}{
		"data": map[string]interface{}{
			"keys":  keys,
			"count": len(keys),
		},
	})
}

// ListLicenseKeys handles GET /api/v1/admin/license-keys
func (h *AdminBillingHandler) ListLicenseKeys(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	statusStr := r.URL.Query().Get("status")
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	var status *billing.LicenseKeyStatus
	if statusStr != "" {
		s := billing.LicenseKeyStatus(statusStr)
		if !s.IsValid() {
			RespondBadRequest(w, "Invalid status value")
			return
		}
		status = &s
	}

	limit := 50
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	offset := 0
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	input := applicense.LicenseKeyListInput{
		Status: status,
		Limit:  limit,
		Offset: offset,
	}

	output, err := h.licenseKeyUsecase.List(r.Context(), input)
	if err != nil {
		RespondInternalError(w)
		return
	}

	// Convert output to response
	keys := make([]map[string]interface{}, len(output.Keys))
	for i, k := range output.Keys {
		item := map[string]interface{}{
			"key_id":     k.KeyID.String(),
			"status":     string(k.Status),
			"memo":       k.Memo,
			"created_at": k.CreatedAt,
		}
		if k.ExpiresAt != nil {
			item["expires_at"] = k.ExpiresAt
		}
		if k.ClaimedAt != nil {
			item["claimed_at"] = k.ClaimedAt
		}
		if k.ClaimedBy != nil {
			item["claimed_by"] = k.ClaimedBy.String()
		}
		keys[i] = item
	}

	RespondJSON(w, http.StatusOK, map[string]interface{}{
		"data": map[string]interface{}{
			"keys":        keys,
			"total_count": output.TotalCount,
			"limit":       limit,
			"offset":      offset,
		},
	})
}

// RevokeLicenseKeyRequest represents the request for revoking a license key
type RevokeLicenseKeyRequest struct {
	Action string `json:"action"` // "revoke"
}

// UpdateLicenseKey handles PATCH /api/v1/admin/license-keys/{id}
func (h *AdminBillingHandler) UpdateLicenseKey(w http.ResponseWriter, r *http.Request) {
	adminID := getAdminIDFromContext(r)
	keyIDStr := chi.URLParam(r, "id")

	keyID, err := billing.ParseLicenseKeyID(keyIDStr)
	if err != nil {
		RespondBadRequest(w, "Invalid key ID")
		return
	}

	var req RevokeLicenseKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondBadRequest(w, "Invalid request body")
		return
	}

	switch req.Action {
	case "revoke":
		input := applicense.RevokeLicenseKeyInput{
			KeyID:   keyID,
			AdminID: adminID,
		}

		if err := h.licenseKeyUsecase.Revoke(r.Context(), input); err != nil {
			if domainErr, ok := err.(*common.DomainError); ok {
				if domainErr.Code() == common.ErrNotFound {
					RespondError(w, http.StatusNotFound, domainErr.Code(), domainErr.Message, nil)
					return
				}
				RespondError(w, http.StatusBadRequest, domainErr.Code(), domainErr.Message, nil)
				return
			}
			RespondInternalError(w)
			return
		}

		RespondJSON(w, http.StatusOK, map[string]interface{}{
			"data": map[string]string{
				"status": "revoked",
			},
		})
	default:
		RespondBadRequest(w, "Invalid action")
	}
}

// ListTenants handles GET /api/v1/admin/tenants
func (h *AdminBillingHandler) ListTenants(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	statusStr := r.URL.Query().Get("status")
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	var status *tenant.TenantStatus
	if statusStr != "" {
		s := tenant.TenantStatus(statusStr)
		status = &s
	}

	limit := 50
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	offset := 0
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	input := apptenant.TenantListInput{
		Status: status,
		Limit:  limit,
		Offset: offset,
	}

	output, err := h.tenantUsecase.List(r.Context(), input)
	if err != nil {
		RespondInternalError(w)
		return
	}

	// Convert output to response
	tenants := make([]map[string]interface{}, len(output.Tenants))
	for i, t := range output.Tenants {
		item := map[string]interface{}{
			"tenant_id":   t.TenantID.String(),
			"tenant_name": t.TenantName,
			"status":      string(t.Status),
			"created_at":  t.CreatedAt,
			"updated_at":  t.UpdatedAt,
		}
		if t.GraceUntil != nil {
			item["grace_until"] = t.GraceUntil
		}
		tenants[i] = item
	}

	RespondJSON(w, http.StatusOK, map[string]interface{}{
		"data": map[string]interface{}{
			"tenants":     tenants,
			"total_count": output.TotalCount,
			"limit":       limit,
			"offset":      offset,
		},
	})
}

// GetTenantDetail handles GET /api/v1/admin/tenants/{id}
func (h *AdminBillingHandler) GetTenantDetail(w http.ResponseWriter, r *http.Request) {
	tenantIDStr := chi.URLParam(r, "id")

	tenantID, err := common.ParseTenantID(tenantIDStr)
	if err != nil {
		RespondBadRequest(w, "Invalid tenant ID")
		return
	}

	output, err := h.tenantUsecase.GetDetail(r.Context(), tenantID)
	if err != nil {
		if domainErr, ok := err.(*common.DomainError); ok {
			if domainErr.Code() == common.ErrNotFound {
				RespondError(w, http.StatusNotFound, domainErr.Code(), domainErr.Message, nil)
				return
			}
		}
		RespondInternalError(w)
		return
	}

	// Convert entitlements
	entitlements := make([]map[string]interface{}, len(output.Entitlements))
	for i, e := range output.Entitlements {
		item := map[string]interface{}{
			"entitlement_id": e.EntitlementID.String(),
			"plan_code":      e.PlanCode,
			"source":         string(e.Source),
			"started_at":     e.StartsAt,
		}
		if e.RevokedAt != nil {
			item["revoked_at"] = e.RevokedAt
		}
		entitlements[i] = item
	}

	// Convert subscription
	var subscription map[string]interface{}
	if output.Subscription != nil {
		subscription = map[string]interface{}{
			"subscription_id":        output.Subscription.SubscriptionID.String(),
			"stripe_customer_id":     output.Subscription.StripeCustomerID,
			"stripe_subscription_id": output.Subscription.StripeSubscriptionID,
			"status":                 string(output.Subscription.Status),
			"created_at":             output.Subscription.CreatedAt,
			"updated_at":             output.Subscription.UpdatedAt,
		}
		if output.Subscription.CurrentPeriodEnd != nil {
			subscription["current_period_end"] = output.Subscription.CurrentPeriodEnd
		}
	}

	// Convert admins
	admins := make([]map[string]interface{}, len(output.Admins))
	for i, a := range output.Admins {
		admins[i] = map[string]interface{}{
			"admin_id":     a.AdminID.String(),
			"email":        a.Email,
			"display_name": a.DisplayName,
			"role":         string(a.Role),
		}
	}

	data := map[string]interface{}{
		"tenant_id":    output.TenantID.String(),
		"tenant_name":  output.TenantName,
		"status":       string(output.Status),
		"created_at":   output.CreatedAt,
		"updated_at":   output.UpdatedAt,
		"entitlements": entitlements,
		"admins":       admins,
	}
	if output.GraceUntil != nil {
		data["grace_until"] = output.GraceUntil
	}
	if subscription != nil {
		data["subscription"] = subscription
	}

	RespondJSON(w, http.StatusOK, map[string]interface{}{
		"data": data,
	})
}

// UpdateTenantStatusRequest represents the request for updating tenant status
type UpdateTenantStatusRequest struct {
	Status     string     `json:"status"`
	GraceUntil *time.Time `json:"grace_until"`
}

// UpdateTenantStatus handles PATCH /api/v1/admin/tenants/{id}/status
func (h *AdminBillingHandler) UpdateTenantStatus(w http.ResponseWriter, r *http.Request) {
	adminID := getAdminIDFromContext(r)
	tenantIDStr := chi.URLParam(r, "id")

	tenantID, err := common.ParseTenantID(tenantIDStr)
	if err != nil {
		RespondBadRequest(w, "Invalid tenant ID")
		return
	}

	var req UpdateTenantStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondBadRequest(w, "Invalid request body")
		return
	}

	status := tenant.TenantStatus(req.Status)
	if status != tenant.TenantStatusActive &&
		status != tenant.TenantStatusGrace &&
		status != tenant.TenantStatusSuspended {
		RespondBadRequest(w, "Invalid status value")
		return
	}

	input := apptenant.UpdateTenantStatusInput{
		TenantID:   tenantID,
		Status:     status,
		GraceUntil: req.GraceUntil,
		AdminID:    adminID,
	}

	if err := h.tenantUsecase.UpdateStatus(r.Context(), input); err != nil {
		if domainErr, ok := err.(*common.DomainError); ok {
			if domainErr.Code() == common.ErrNotFound {
				RespondError(w, http.StatusNotFound, domainErr.Code(), domainErr.Message, nil)
				return
			}
			RespondError(w, http.StatusBadRequest, domainErr.Code(), domainErr.Message, nil)
			return
		}
		RespondInternalError(w)
		return
	}

	RespondJSON(w, http.StatusOK, map[string]interface{}{
		"data": map[string]string{
			"status": req.Status,
		},
	})
}

// ListAuditLogs handles GET /api/v1/admin/audit-logs
func (h *AdminBillingHandler) ListAuditLogs(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	actionStr := r.URL.Query().Get("action")
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	var action *string
	if actionStr != "" {
		action = &actionStr
	}

	limit := 50
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	offset := 0
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	input := appaudit.AuditLogListInput{
		Action: action,
		Limit:  limit,
		Offset: offset,
	}

	output, err := h.auditLogUsecase.List(r.Context(), input)
	if err != nil {
		RespondInternalError(w)
		return
	}

	// Convert output to response
	logs := make([]map[string]interface{}, len(output.Logs))
	for i, l := range output.Logs {
		item := map[string]interface{}{
			"log_id":     l.LogID.String(),
			"actor_type": string(l.ActorType),
			"action":     l.Action,
			"created_at": l.CreatedAt,
		}
		if l.ActorID != nil {
			item["actor_id"] = *l.ActorID
		}
		if l.TargetType != nil {
			item["target_type"] = *l.TargetType
		}
		if l.TargetID != nil {
			item["target_id"] = *l.TargetID
		}
		if l.BeforeJSON != nil {
			item["before_json"] = *l.BeforeJSON
		}
		if l.AfterJSON != nil {
			item["after_json"] = *l.AfterJSON
		}
		if l.IPAddress != nil {
			item["ip_address"] = *l.IPAddress
		}
		if l.UserAgent != nil {
			item["user_agent"] = *l.UserAgent
		}
		logs[i] = item
	}

	RespondJSON(w, http.StatusOK, map[string]interface{}{
		"data": map[string]interface{}{
			"logs":        logs,
			"total_count": output.TotalCount,
			"limit":       limit,
			"offset":      offset,
		},
	})
}
