package auth

import (
	"context"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/auth"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/services"
)

// AdminAllowPasswordResetInput represents the input for system admin allowing password reset
type AdminAllowPasswordResetInput struct {
	TargetAdminID common.AdminID // PWリセットを許可する対象のID
	SystemAdminID common.AdminID // システム管理者のID（監査ログ用）
}

// AdminAllowPasswordResetOutput represents the output for system admin allowing password reset
type AdminAllowPasswordResetOutput struct {
	TargetAdminID string `json:"target_admin_id"`
	TargetEmail   string `json:"target_email"`
	TenantID      string `json:"tenant_id"`
	AllowedAt     string `json:"allowed_at"`
	ExpiresAt     string `json:"expires_at"`
}

// AdminAllowPasswordResetUsecase handles the password reset allowance by system admin
// This bypasses the Owner role check since it's a system admin operation
type AdminAllowPasswordResetUsecase struct {
	adminRepo auth.AdminRepository
	clock     services.Clock
}

// NewAdminAllowPasswordResetUsecase creates a new AdminAllowPasswordResetUsecase
func NewAdminAllowPasswordResetUsecase(
	adminRepo auth.AdminRepository,
	clock services.Clock,
) *AdminAllowPasswordResetUsecase {
	return &AdminAllowPasswordResetUsecase{
		adminRepo: adminRepo,
		clock:     clock,
	}
}

// Execute executes the password reset allowance by system admin
func (u *AdminAllowPasswordResetUsecase) Execute(ctx context.Context, input AdminAllowPasswordResetInput) (*AdminAllowPasswordResetOutput, error) {
	// 1. 対象の管理者を取得（テナントIDなしでグローバル検索）
	targetAdmin, err := u.adminRepo.FindByID(ctx, input.TargetAdminID)
	if err != nil {
		return nil, err
	}
	if targetAdmin == nil {
		return nil, ErrAdminNotFound
	}

	// 2. PWリセットを許可（システム管理者による操作なのでallowedByはNULL）
	// NOTE: システム管理者IDはadminsテーブルに存在しないため、
	// 外部キー制約を回避するためにNULLを設定
	now := u.clock.Now()
	targetAdmin.AllowPasswordResetBySystem(now)

	// 3. 保存
	if err := u.adminRepo.Save(ctx, targetAdmin); err != nil {
		return nil, err
	}

	// 4. 出力を構築
	expiresAt := targetAdmin.PasswordResetExpiresAt()
	var expiresAtStr string
	if expiresAt != nil {
		expiresAtStr = expiresAt.Format("2006-01-02T15:04:05Z07:00")
	}

	return &AdminAllowPasswordResetOutput{
		TargetAdminID: targetAdmin.AdminID().String(),
		TargetEmail:   targetAdmin.Email(),
		TenantID:      targetAdmin.TenantID().String(),
		AllowedAt:     now.Format("2006-01-02T15:04:05Z07:00"),
		ExpiresAt:     expiresAtStr,
	}, nil
}
