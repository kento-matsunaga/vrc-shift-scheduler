package usecase

import (
	"context"
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/tenant"
)

// TenantRepository defines the interface for tenant persistence
type TenantRepository interface {
	FindByID(ctx context.Context, tenantID common.TenantID) (*tenant.Tenant, error)
	Save(ctx context.Context, tenant *tenant.Tenant) error
}

// GetTenantInput represents the input for getting a tenant
type GetTenantInput struct {
	TenantID common.TenantID
}

// GetTenantUsecase handles the tenant retrieval use case
type GetTenantUsecase struct {
	tenantRepo TenantRepository
}

// NewGetTenantUsecase creates a new GetTenantUsecase
func NewGetTenantUsecase(tenantRepo TenantRepository) *GetTenantUsecase {
	return &GetTenantUsecase{
		tenantRepo: tenantRepo,
	}
}

// Execute retrieves a tenant by ID
func (uc *GetTenantUsecase) Execute(ctx context.Context, input GetTenantInput) (*tenant.Tenant, error) {
	return uc.tenantRepo.FindByID(ctx, input.TenantID)
}

// UpdateTenantInput represents the input for updating a tenant
type UpdateTenantInput struct {
	TenantID   common.TenantID
	TenantName string
}

// UpdateTenantUsecase handles the tenant update use case
type UpdateTenantUsecase struct {
	tenantRepo TenantRepository
}

// NewUpdateTenantUsecase creates a new UpdateTenantUsecase
func NewUpdateTenantUsecase(tenantRepo TenantRepository) *UpdateTenantUsecase {
	return &UpdateTenantUsecase{
		tenantRepo: tenantRepo,
	}
}

// Execute updates a tenant
func (uc *UpdateTenantUsecase) Execute(ctx context.Context, input UpdateTenantInput) (*tenant.Tenant, error) {
	// テナントを取得
	t, err := uc.tenantRepo.FindByID(ctx, input.TenantID)
	if err != nil {
		return nil, err
	}

	// テナント名を更新
	if err := t.UpdateTenantName(time.Now(), input.TenantName); err != nil {
		return nil, err
	}

	// 保存
	if err := uc.tenantRepo.Save(ctx, t); err != nil {
		return nil, err
	}

	return t, nil
}
