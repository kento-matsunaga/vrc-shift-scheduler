package auth

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
)

// InvitationID は招待のID
type InvitationID string

func (id InvitationID) String() string {
	return string(id)
}

func (id InvitationID) Validate() error {
	if id == "" {
		return common.NewValidationError("invitation_id is required", nil)
	}
	return nil
}

// Invitation は管理者招待を表す集約ルート
type Invitation struct {
	invitationID     InvitationID
	tenantID         common.TenantID    // 招待者のテナントに自動紐付け
	email            string
	role             Role
	token            string
	createdByAdminID common.AdminID     // 招待者（必須）
	expiresAt        time.Time
	acceptedAt       *time.Time
	createdAt        time.Time
}

// NewInvitation は新しい招待を作成する
// DDD原則: 集約ルート生成時にビジネスルールを適用
func NewInvitation(
	now time.Time,
	createdByAdmin *Admin, // ★ Admin集約を受け取る（tenantIDを自動抽出）
	email string,
	role Role,
	expirationDuration time.Duration,
) (*Invitation, error) {
	// ビジネスルール: 招待者が有効な管理者である必要がある
	if !createdByAdmin.IsActive() {
		return nil, common.NewValidationError("inviter must be active", nil)
	}
	if createdByAdmin.IsDeleted() {
		return nil, common.NewValidationError("inviter must not be deleted", nil)
	}

	// セキュアなランダムトークン生成（32バイト = 64文字のhex）
	token, err := generateSecureToken(32)
	if err != nil {
		return nil, common.NewValidationError("failed to generate secure token", err)
	}

	// ★ 招待者のテナントに自動紐付け
	inv := &Invitation{
		invitationID:     InvitationID(common.NewULID()),
		tenantID:         createdByAdmin.TenantID(), // ★ 招待者のテナントを自動設定
		email:            email,
		role:             role,
		token:            token,
		createdByAdminID: createdByAdmin.AdminID(), // ★ 招待者のIDを記録
		expiresAt:        now.Add(expirationDuration),
		createdAt:        now,
	}

	if err := inv.validate(); err != nil {
		return nil, err
	}

	return inv, nil
}

// ReconstructInvitation は永続化された招待を再構築する
func ReconstructInvitation(
	invitationID InvitationID,
	tenantID common.TenantID,
	email string,
	role Role,
	token string,
	createdByAdminID common.AdminID,
	expiresAt time.Time,
	acceptedAt *time.Time,
	createdAt time.Time,
) (*Invitation, error) {
	inv := &Invitation{
		invitationID:     invitationID,
		tenantID:         tenantID,
		email:            email,
		role:             role,
		token:            token,
		createdByAdminID: createdByAdminID,
		expiresAt:        expiresAt,
		acceptedAt:       acceptedAt,
		createdAt:        createdAt,
	}

	if err := inv.validate(); err != nil {
		return nil, err
	}

	return inv, nil
}

func (i *Invitation) validate() error {
	if err := i.invitationID.Validate(); err != nil {
		return err
	}
	if err := i.tenantID.Validate(); err != nil {
		return err
	}
	if err := i.createdByAdminID.Validate(); err != nil {
		return err
	}
	if i.email == "" {
		return common.NewValidationError("email is required", nil)
	}
	if len(i.email) > 255 {
		return common.NewValidationError("email must be less than 255 characters", nil)
	}
	if err := i.role.Validate(); err != nil {
		return err
	}
	if i.token == "" {
		return common.NewValidationError("token is required", nil)
	}
	if i.expiresAt.Before(i.createdAt) {
		return common.NewValidationError("expires_at must be after created_at", nil)
	}
	return nil
}

// IsExpired は招待が期限切れかどうかを判定する（ドメインルール）
func (i *Invitation) IsExpired(now time.Time) bool {
	return now.After(i.expiresAt)
}

// IsAccepted は招待が既に受理されているかを判定する（ドメインルール）
func (i *Invitation) IsAccepted() bool {
	return i.acceptedAt != nil
}

// CanAccept は招待を受理できるかを判定する（ドメインルール）
func (i *Invitation) CanAccept(now time.Time) error {
	if i.IsAccepted() {
		return common.NewValidationError("invitation already accepted", nil)
	}
	if i.IsExpired(now) {
		return common.NewValidationError("invitation expired", nil)
	}
	return nil
}

// Accept は招待を受理する（状態遷移）
func (i *Invitation) Accept(now time.Time) error {
	if err := i.CanAccept(now); err != nil {
		return err
	}
	i.acceptedAt = &now
	return nil
}

// Getters（不変性を保つ）
func (i *Invitation) InvitationID() InvitationID       { return i.invitationID }
func (i *Invitation) TenantID() common.TenantID        { return i.tenantID }
func (i *Invitation) Email() string                    { return i.email }
func (i *Invitation) Role() Role                       { return i.role }
func (i *Invitation) Token() string                    { return i.token }
func (i *Invitation) CreatedByAdminID() common.AdminID { return i.createdByAdminID }
func (i *Invitation) ExpiresAt() time.Time             { return i.expiresAt }
func (i *Invitation) AcceptedAt() *time.Time           { return i.acceptedAt }
func (i *Invitation) CreatedAt() time.Time             { return i.createdAt }

// generateSecureToken は暗号学的に安全なランダムトークンを生成する
func generateSecureToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
