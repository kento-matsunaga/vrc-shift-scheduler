package license

import (
	"context"
	"time"
	"unicode"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/auth"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/billing"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/tenant"
)

// LicenseClaimInput represents the input for claiming a license
type LicenseClaimInput struct {
	Email       string
	Password    string
	DisplayName string
	TenantName  string
	LicenseKey  string
	IPAddress   string
	UserAgent   string
}

// LicenseClaimOutput represents the output of a successful claim
type LicenseClaimOutput struct {
	TenantID    common.TenantID
	AdminID     common.AdminID
	TenantName  string
	DisplayName string
	Email       string
}

// TxManager defines the interface for transaction management
type TxManager interface {
	WithTx(ctx context.Context, fn func(context.Context) error) error
}

// LicenseClaimUsecase handles the license claim process
type LicenseClaimUsecase struct {
	txManager       TxManager
	tenantRepo      tenant.TenantRepository
	adminRepo       auth.AdminRepository
	licenseKeyRepo  billing.LicenseKeyRepository
	entitlementRepo billing.EntitlementRepository
	auditLogRepo    billing.BillingAuditLogRepository
	passwordHasher  PasswordHasher
}

// PasswordHasher defines the interface for password hashing
type PasswordHasher interface {
	Hash(password string) (string, error)
}

// NewLicenseClaimUsecase creates a new LicenseClaimUsecase
func NewLicenseClaimUsecase(
	txManager TxManager,
	tenantRepo tenant.TenantRepository,
	adminRepo auth.AdminRepository,
	licenseKeyRepo billing.LicenseKeyRepository,
	entitlementRepo billing.EntitlementRepository,
	auditLogRepo billing.BillingAuditLogRepository,
	passwordHasher PasswordHasher,
) *LicenseClaimUsecase {
	return &LicenseClaimUsecase{
		txManager:       txManager,
		tenantRepo:      tenantRepo,
		adminRepo:       adminRepo,
		licenseKeyRepo:  licenseKeyRepo,
		entitlementRepo: entitlementRepo,
		auditLogRepo:    auditLogRepo,
		passwordHasher:  passwordHasher,
	}
}

// Execute claims a license key and creates tenant + admin
func (uc *LicenseClaimUsecase) Execute(ctx context.Context, input LicenseClaimInput) (*LicenseClaimOutput, error) {
	now := time.Now().UTC()

	// Validate license key format
	if !billing.ValidateLicenseKeyFormat(input.LicenseKey) {
		return nil, common.NewValidationError("ライセンスキーの形式が正しくありません", nil)
	}

	// Validate email
	if input.Email == "" {
		return nil, common.NewValidationError("メールアドレスを入力してください", nil)
	}

	// Validate password complexity
	if err := validatePasswordComplexity(input.Password); err != nil {
		return nil, err
	}

	// Validate display name
	if input.DisplayName == "" {
		return nil, common.NewValidationError("表示名を入力してください", nil)
	}

	// Validate tenant name
	if input.TenantName == "" {
		return nil, common.NewValidationError("テナント名を入力してください", nil)
	}

	// Normalize and hash the license key
	normalizedKey := billing.NormalizeLicenseKey(input.LicenseKey)
	keyHash := billing.HashLicenseKey(normalizedKey)

	var output *LicenseClaimOutput

	// Execute in transaction
	err := uc.txManager.WithTx(ctx, func(txCtx context.Context) error {
		// 1. Find and lock the license key
		licenseKey, err := uc.licenseKeyRepo.FindByHashForUpdate(txCtx, keyHash)
		if err != nil {
			return err
		}
		if licenseKey == nil {
			return common.NewValidationError("ライセンスキーが見つかりません", nil)
		}

		// 2. Verify the key is unused
		if !licenseKey.IsUnused() {
			if licenseKey.IsUsed() {
				return common.NewValidationError("このライセンスキーは既に使用されています", nil)
			}
			if licenseKey.IsRevoked() {
				return common.NewValidationError("このライセンスキーは無効化されています", nil)
			}
		}

		// 3. Create tenant
		newTenant, err := tenant.NewTenant(now, input.TenantName, "Asia/Tokyo")
		if err != nil {
			return err
		}

		if err := uc.tenantRepo.Save(txCtx, newTenant); err != nil {
			return err
		}

		// 4. Hash password and create admin
		passwordHash, err := uc.passwordHasher.Hash(input.Password)
		if err != nil {
			return err
		}

		newAdmin, err := auth.NewAdmin(
			now,
			newTenant.TenantID(),
			input.Email,
			passwordHash,
			input.DisplayName,
			auth.RoleOwner,
		)
		if err != nil {
			return err
		}

		if err := uc.adminRepo.Save(txCtx, newAdmin); err != nil {
			return err
		}

		// 5. Create entitlement
		entitlement, err := billing.NewEntitlement(
			now,
			newTenant.TenantID(),
			"LIFETIME",
			billing.EntitlementSourceBooth,
			nil, // lifetime has no end date
		)
		if err != nil {
			return err
		}

		if err := uc.entitlementRepo.Save(txCtx, entitlement); err != nil {
			return err
		}

		// 6. Mark license key as used
		if err := licenseKey.MarkAsUsed(now, newTenant.TenantID()); err != nil {
			return err
		}

		if err := uc.licenseKeyRepo.Save(txCtx, licenseKey); err != nil {
			return err
		}

		// 7. Create audit log
		tenantIDStr := newTenant.TenantID().String()
		afterJSON := `{"tenant_id":"` + tenantIDStr + `","plan_code":"LIFETIME","email":"` + input.Email + `"}`
		auditLog, err := billing.NewBillingAuditLog(
			now,
			billing.ActorTypeUser,
			nil,
			string(billing.BillingAuditActionLicenseClaim),
			strPtr("tenant"),
			&tenantIDStr,
			nil,
			&afterJSON,
			strPtr(input.IPAddress),
			strPtr(input.UserAgent),
		)
		if err != nil {
			return err
		}

		if err := uc.auditLogRepo.Save(txCtx, auditLog); err != nil {
			return err
		}

		output = &LicenseClaimOutput{
			TenantID:    newTenant.TenantID(),
			AdminID:     newAdmin.AdminID(),
			TenantName:  newTenant.TenantName(),
			DisplayName: newAdmin.DisplayName(),
			Email:       newAdmin.Email(),
		}

		return nil
	})

	if err != nil {
		// Log failed attempt
		uc.logFailedAttempt(ctx, input, now, err.Error())
		return nil, err
	}

	return output, nil
}

func (uc *LicenseClaimUsecase) logFailedAttempt(ctx context.Context, input LicenseClaimInput, now time.Time, reason string) {
	afterJSON := `{"email":"` + input.Email + `","reason":"` + reason + `"}`
	auditLog, err := billing.NewBillingAuditLog(
		now,
		billing.ActorTypeUser,
		nil,
		string(billing.BillingAuditActionLicenseClaimFailed),
		nil,
		nil,
		nil,
		&afterJSON,
		strPtr(input.IPAddress),
		strPtr(input.UserAgent),
	)
	if err != nil {
		return
	}

	// Best effort - don't fail the request if audit log fails
	_ = uc.auditLogRepo.Save(ctx, auditLog)
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// validatePasswordComplexity checks password meets security requirements
func validatePasswordComplexity(password string) error {
	if len(password) < 8 {
		return common.NewValidationError("パスワードは8文字以上で入力してください", nil)
	}

	if len(password) > 128 {
		return common.NewValidationError("パスワードは128文字以内で入力してください", nil)
	}

	var hasUpper, hasLower, hasDigit bool
	for _, r := range password {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsDigit(r):
			hasDigit = true
		}
	}

	if !hasUpper {
		return common.NewValidationError("パスワードには大文字を1文字以上含めてください", nil)
	}
	if !hasLower {
		return common.NewValidationError("パスワードには小文字を1文字以上含めてください", nil)
	}
	if !hasDigit {
		return common.NewValidationError("パスワードには数字を1文字以上含めてください", nil)
	}

	return nil
}
