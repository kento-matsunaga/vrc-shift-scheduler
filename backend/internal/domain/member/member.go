package member

import (
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
)

// Member represents a member entity (aggregate root)
// メンバーはシフトに参加するメンバー情報を表す（真のMVP版: 最小限のフィールド）
type Member struct {
	memberID      common.MemberID
	tenantID      common.TenantID
	displayName   string
	discordUserID string // オプショナル
	email         string // オプショナル
	isActive      bool
	createdAt     time.Time
	updatedAt     time.Time
	deletedAt     *time.Time
}

// NewMember creates a new Member entity
func NewMember(
	tenantID common.TenantID,
	displayName string,
	discordUserID string,
	email string,
) (*Member, error) {
	member := &Member{
		memberID:      common.NewMemberID(),
		tenantID:      tenantID,
		displayName:   displayName,
		discordUserID: discordUserID,
		email:         email,
		isActive:      true,
		createdAt:     time.Now(),
		updatedAt:     time.Now(),
	}

	if err := member.validate(); err != nil {
		return nil, err
	}

	return member, nil
}

// ReconstructMember reconstructs a Member entity from persistence
func ReconstructMember(
	memberID common.MemberID,
	tenantID common.TenantID,
	displayName string,
	discordUserID string,
	email string,
	isActive bool,
	createdAt time.Time,
	updatedAt time.Time,
	deletedAt *time.Time,
) (*Member, error) {
	member := &Member{
		memberID:      memberID,
		tenantID:      tenantID,
		displayName:   displayName,
		discordUserID: discordUserID,
		email:         email,
		isActive:      isActive,
		createdAt:     createdAt,
		updatedAt:     updatedAt,
		deletedAt:     deletedAt,
	}

	if err := member.validate(); err != nil {
		return nil, err
	}

	return member, nil
}

func (m *Member) validate() error {
	// TenantID の必須性チェック
	if err := m.tenantID.Validate(); err != nil {
		return common.NewValidationError("tenant_id is required", err)
	}

	// DisplayName の必須性チェック
	if m.displayName == "" {
		return common.NewValidationError("display_name is required", nil)
	}

	if len(m.displayName) > 255 {
		return common.NewValidationError("display_name must be less than 255 characters", nil)
	}

	// Email の長さチェック（オプショナルだが、設定されている場合）
	if m.email != "" && len(m.email) > 255 {
		return common.NewValidationError("email must be less than 255 characters", nil)
	}

	// DiscordUserID の長さチェック（オプショナルだが、設定されている場合）
	if m.discordUserID != "" && len(m.discordUserID) > 100 {
		return common.NewValidationError("discord_user_id must be less than 100 characters", nil)
	}

	return nil
}

// Getters

func (m *Member) MemberID() common.MemberID {
	return m.memberID
}

func (m *Member) TenantID() common.TenantID {
	return m.tenantID
}

func (m *Member) DisplayName() string {
	return m.displayName
}

func (m *Member) DiscordUserID() string {
	return m.discordUserID
}

func (m *Member) Email() string {
	return m.email
}

func (m *Member) IsActive() bool {
	return m.isActive
}

func (m *Member) CreatedAt() time.Time {
	return m.createdAt
}

func (m *Member) UpdatedAt() time.Time {
	return m.updatedAt
}

func (m *Member) DeletedAt() *time.Time {
	return m.deletedAt
}

func (m *Member) IsDeleted() bool {
	return m.deletedAt != nil
}

// UpdateDetails updates multiple member details at once
func (m *Member) UpdateDetails(displayName, discordUserID, email string, isActive bool) error {
	// Update fields
	m.displayName = displayName
	m.discordUserID = discordUserID
	m.email = email
	m.isActive = isActive
	m.updatedAt = time.Now()

	// Validate after update
	return m.validate()
}

// UpdateDisplayName updates the display name
func (m *Member) UpdateDisplayName(displayName string) error {
	if displayName == "" {
		return common.NewValidationError("display_name is required", nil)
	}
	if len(displayName) > 255 {
		return common.NewValidationError("display_name must be less than 255 characters", nil)
	}

	m.displayName = displayName
	m.updatedAt = time.Now()
	return nil
}

// UpdateDiscordUserID updates the Discord user ID
func (m *Member) UpdateDiscordUserID(discordUserID string) error {
	if discordUserID != "" && len(discordUserID) > 100 {
		return common.NewValidationError("discord_user_id must be less than 100 characters", nil)
	}

	m.discordUserID = discordUserID
	m.updatedAt = time.Now()
	return nil
}

// UpdateEmail updates the email address
func (m *Member) UpdateEmail(email string) error {
	if email != "" && len(email) > 255 {
		return common.NewValidationError("email must be less than 255 characters", nil)
	}

	m.email = email
	m.updatedAt = time.Now()
	return nil
}

// Activate activates the member
func (m *Member) Activate() {
	m.isActive = true
	m.updatedAt = time.Now()
}

// Deactivate deactivates the member
func (m *Member) Deactivate() {
	m.isActive = false
	m.updatedAt = time.Now()
}

// Delete marks the member as deleted (soft delete)
func (m *Member) Delete() {
	now := time.Now()
	m.deletedAt = &now
	m.updatedAt = now
}

