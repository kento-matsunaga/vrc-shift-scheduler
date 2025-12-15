package auth

import (
	"context"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
)

// InvitationRepository は招待の永続化を担当する
type InvitationRepository interface {
	// Save は招待を保存する（INSERT or UPDATE）
	Save(ctx context.Context, invitation *Invitation) error

	// FindByToken はトークンで招待を検索する
	FindByToken(ctx context.Context, token string) (*Invitation, error)

	// FindByTenantID はテナント内の招待一覧を取得する
	FindByTenantID(ctx context.Context, tenantID common.TenantID) ([]*Invitation, error)

	// ExistsPendingByEmail はテナント内の特定メールアドレスの未受理招待が存在するかチェック
	ExistsPendingByEmail(ctx context.Context, tenantID common.TenantID, email string) (bool, error)

	// Delete は招待を削除する（物理削除）
	Delete(ctx context.Context, invitationID InvitationID) error
}
