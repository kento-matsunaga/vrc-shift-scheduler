package rest

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/application/usecase"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/tenant"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/infra/db"
	"github.com/jackc/pgx/v5/pgxpool"
)

// TenantHandler handles tenant-related HTTP requests
type TenantHandler struct {
	getTenantUC    *usecase.GetTenantUsecase
	updateTenantUC *usecase.UpdateTenantUsecase
}

// NewTenantHandler creates a new TenantHandler
func NewTenantHandler(dbPool *pgxpool.Pool) *TenantHandler {
	tenantRepo := db.NewTenantRepository(dbPool)
	return &TenantHandler{
		getTenantUC:    usecase.NewGetTenantUsecase(tenantRepo),
		updateTenantUC: usecase.NewUpdateTenantUsecase(tenantRepo),
	}
}

// TenantResponse represents a tenant in API responses
type TenantResponse struct {
	TenantID   string `json:"tenant_id"`
	TenantName string `json:"tenant_name"`
	Timezone   string `json:"timezone"`
	IsActive   bool   `json:"is_active"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
}

// UpdateTenantRequest represents the request body for updating a tenant
type UpdateTenantRequest struct {
	TenantName string `json:"tenant_name"`
}

// GetCurrentTenant handles GET /api/v1/tenants/me
func (h *TenantHandler) GetCurrentTenant(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// テナントIDの取得
	tenantID, ok := GetTenantID(ctx)
	if !ok {
		RespondBadRequest(w, "tenant_id is required")
		return
	}

	// Usecaseの実行
	input := usecase.GetTenantInput{
		TenantID: tenantID,
	}

	t, err := h.getTenantUC.Execute(ctx, input)
	if err != nil {
		RespondDomainError(w, err)
		return
	}

	// レスポンス
	RespondSuccess(w, toTenantResponse(t))
}

// UpdateCurrentTenant handles PUT /api/v1/tenants/me
func (h *TenantHandler) UpdateCurrentTenant(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// テナントIDの取得
	tenantID, ok := GetTenantID(ctx)
	if !ok {
		RespondBadRequest(w, "tenant_id is required")
		return
	}

	// リクエストボディのパース
	var req UpdateTenantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondBadRequest(w, "Invalid request body")
		return
	}

	// バリデーション
	if req.TenantName == "" {
		RespondBadRequest(w, "tenant_name is required")
		return
	}

	// Usecaseの実行
	input := usecase.UpdateTenantInput{
		TenantID:   tenantID,
		TenantName: req.TenantName,
	}

	t, err := h.updateTenantUC.Execute(ctx, input)
	if err != nil {
		RespondDomainError(w, err)
		return
	}

	// レスポンス
	RespondSuccess(w, toTenantResponse(t))
}

// toTenantResponse converts a Tenant entity to TenantResponse
func toTenantResponse(t *tenant.Tenant) TenantResponse {
	return TenantResponse{
		TenantID:   t.TenantID().String(),
		TenantName: t.TenantName(),
		Timezone:   t.Timezone(),
		IsActive:   t.IsActive(),
		CreatedAt:  t.CreatedAt().Format(time.RFC3339),
		UpdatedAt:  t.UpdatedAt().Format(time.RFC3339),
	}
}
