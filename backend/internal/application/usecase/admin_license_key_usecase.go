package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/billing"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
)

// AdminLicenseKeyUsecase handles admin operations for license keys
type AdminLicenseKeyUsecase struct {
	txManager      TxManager
	licenseKeyRepo billing.LicenseKeyRepository
	auditLogRepo   billing.BillingAuditLogRepository
}

// NewAdminLicenseKeyUsecase creates a new AdminLicenseKeyUsecase
func NewAdminLicenseKeyUsecase(
	txManager TxManager,
	licenseKeyRepo billing.LicenseKeyRepository,
	auditLogRepo billing.BillingAuditLogRepository,
) *AdminLicenseKeyUsecase {
	return &AdminLicenseKeyUsecase{
		txManager:      txManager,
		licenseKeyRepo: licenseKeyRepo,
		auditLogRepo:   auditLogRepo,
	}
}

// GenerateLicenseKeyInput represents input for generating license keys
type GenerateLicenseKeyInput struct {
	Count     int
	ExpiresAt *time.Time
	Memo      string
	AdminID   common.AdminID
}

// GeneratedKey represents a generated license key
type GeneratedKey struct {
	KeyID     billing.LicenseKeyID
	Key       string // The actual key (only returned on generation)
	ExpiresAt *time.Time
	CreatedAt time.Time
}

// GenerateLicenseKeyOutput represents output from key generation
type GenerateLicenseKeyOutput struct {
	Keys []GeneratedKey
}

// Generate creates new license keys
func (uc *AdminLicenseKeyUsecase) Generate(ctx context.Context, input GenerateLicenseKeyInput) (*GenerateLicenseKeyOutput, error) {
	if input.Count <= 0 || input.Count > 100 {
		return nil, common.NewDomainError("ERR_INVALID_COUNT", "Count must be between 1 and 100")
	}

	now := time.Now().UTC()
	output := &GenerateLicenseKeyOutput{
		Keys: make([]GeneratedKey, 0, input.Count),
	}

	err := uc.txManager.WithTx(ctx, func(txCtx context.Context) error {
		for i := 0; i < input.Count; i++ {
			// Generate key
			key, err := billing.GenerateLicenseKey()
			if err != nil {
				return fmt.Errorf("failed to generate license key: %w", err)
			}
			// Normalize before hashing (remove hyphens) to ensure consistency with claim
			normalizedKey := billing.NormalizeLicenseKey(key)
			keyHash := billing.HashLicenseKey(normalizedKey)

			// Create entity
			licenseKey, err := billing.NewLicenseKey(
				now,
				keyHash,
				input.ExpiresAt,
				input.Memo,
			)
			if err != nil {
				return fmt.Errorf("failed to create license key: %w", err)
			}

			// Save
			if err := uc.licenseKeyRepo.Save(txCtx, licenseKey); err != nil {
				return fmt.Errorf("failed to save license key: %w", err)
			}

			output.Keys = append(output.Keys, GeneratedKey{
				KeyID:     licenseKey.KeyID(),
				Key:       key, // Return the actual key (only on generation)
				ExpiresAt: input.ExpiresAt,
				CreatedAt: now,
			})
		}

		// Audit log
		adminIDStr := input.AdminID.String()
		afterJSON := fmt.Sprintf(`{"count":%d,"memo":"%s"}`, input.Count, input.Memo)
		auditLog, err := billing.NewBillingAuditLog(
			now,
			billing.ActorTypeAdmin,
			&adminIDStr,
			string(billing.BillingAuditActionLicenseGenerated),
			nil,
			nil,
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

	if err != nil {
		return nil, err
	}

	return output, nil
}

// LicenseKeyListInput represents input for listing license keys
type LicenseKeyListInput struct {
	Status *billing.LicenseKeyStatus
	Limit  int
	Offset int
}

// LicenseKeyListItem represents a license key in the list
type LicenseKeyListItem struct {
	KeyID     billing.LicenseKeyID
	Status    billing.LicenseKeyStatus
	ExpiresAt *time.Time
	ClaimedAt *time.Time
	ClaimedBy *common.TenantID
	Memo      string
	CreatedAt time.Time
}

// LicenseKeyListOutput represents output from listing license keys
type LicenseKeyListOutput struct {
	Keys       []LicenseKeyListItem
	TotalCount int
}

// List returns a list of license keys
func (uc *AdminLicenseKeyUsecase) List(ctx context.Context, input LicenseKeyListInput) (*LicenseKeyListOutput, error) {
	if input.Limit <= 0 {
		input.Limit = 50
	}
	if input.Limit > 100 {
		input.Limit = 100
	}

	keys, totalCount, err := uc.licenseKeyRepo.List(ctx, input.Status, input.Limit, input.Offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list license keys: %w", err)
	}

	items := make([]LicenseKeyListItem, len(keys))
	for i, k := range keys {
		items[i] = LicenseKeyListItem{
			KeyID:     k.KeyID(),
			Status:    k.Status(),
			ExpiresAt: k.ExpiresAt(),
			ClaimedAt: k.ClaimedAt(),
			ClaimedBy: k.ClaimedBy(),
			Memo:      k.Memo(),
			CreatedAt: k.CreatedAt(),
		}
	}

	return &LicenseKeyListOutput{
		Keys:       items,
		TotalCount: totalCount,
	}, nil
}

// RevokeLicenseKeyInput represents input for revoking a license key
type RevokeLicenseKeyInput struct {
	KeyID   billing.LicenseKeyID
	AdminID common.AdminID
}

// Revoke revokes a license key
func (uc *AdminLicenseKeyUsecase) Revoke(ctx context.Context, input RevokeLicenseKeyInput) error {
	now := time.Now().UTC()

	return uc.txManager.WithTx(ctx, func(txCtx context.Context) error {
		key, err := uc.licenseKeyRepo.FindByID(txCtx, input.KeyID)
		if err != nil {
			return fmt.Errorf("failed to find license key: %w", err)
		}
		if key == nil {
			return common.NewDomainError(common.ErrNotFound, "License key not found")
		}

		if err := key.Revoke(now); err != nil {
			return err
		}

		if err := uc.licenseKeyRepo.Save(txCtx, key); err != nil {
			return fmt.Errorf("failed to save license key: %w", err)
		}

		// Audit log
		adminIDStr := input.AdminID.String()
		keyIDStr := input.KeyID.String()
		afterJSON := `{"status":"revoked"}`
		auditLog, err := billing.NewBillingAuditLog(
			now,
			billing.ActorTypeAdmin,
			&adminIDStr,
			string(billing.BillingAuditActionLicenseRevoked),
			strPtr("license_key"),
			&keyIDStr,
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
