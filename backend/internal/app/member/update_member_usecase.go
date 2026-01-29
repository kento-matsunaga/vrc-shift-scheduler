package member

import (
	"context"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/member"
)

type UpdateMemberUsecase struct {
	memberRepo     member.MemberRepository
	memberRoleRepo member.MemberRoleRepository
}

func NewUpdateMemberUsecase(memberRepo member.MemberRepository, memberRoleRepo member.MemberRoleRepository) *UpdateMemberUsecase {
	return &UpdateMemberUsecase{
		memberRepo:     memberRepo,
		memberRoleRepo: memberRoleRepo,
	}
}

type UpdateMemberInput struct {
	TenantID      string // from JWT context
	MemberID      string
	DisplayName   string
	DiscordUserID string
	Email         string
	IsActive      bool
	RoleIDs       []string // Role IDs to assign
}

type UpdateMemberOutput struct {
	MemberID      string   `json:"member_id"`
	TenantID      string   `json:"tenant_id"`
	DisplayName   string   `json:"display_name"`
	DiscordUserID string   `json:"discord_user_id"`
	Email         string   `json:"email"`
	IsActive      bool     `json:"is_active"`
	RoleIDs       []string `json:"role_ids"`
	UpdatedAt     string   `json:"updated_at"`
}

func (u *UpdateMemberUsecase) Execute(ctx context.Context, input UpdateMemberInput) (*UpdateMemberOutput, error) {
	tenantID, err := common.ParseTenantID(input.TenantID)
	if err != nil {
		return nil, err
	}

	memberID, err := common.ParseMemberID(input.MemberID)
	if err != nil {
		return nil, err
	}

	// Find existing member
	m, err := u.memberRepo.FindByID(ctx, tenantID, memberID)
	if err != nil {
		return nil, err
	}

	// Update member details
	if err := m.UpdateDetails(input.DisplayName, input.DiscordUserID, input.Email, input.IsActive); err != nil {
		return nil, err
	}

	// Save member
	if err := u.memberRepo.Save(ctx, m); err != nil {
		return nil, err
	}

	// Update roles if provided
	if input.RoleIDs != nil {
		roleIDs := make([]common.RoleID, 0, len(input.RoleIDs))
		for _, roleIDStr := range input.RoleIDs {
			roleID, err := common.ParseRoleID(roleIDStr)
			if err != nil {
				return nil, err
			}
			roleIDs = append(roleIDs, roleID)
		}

		// Set member roles (replace existing)
		if err := u.memberRoleRepo.SetMemberRoles(ctx, memberID, roleIDs); err != nil {
			return nil, err
		}
	}

	// Get updated roles
	roleIDs, err := u.memberRoleRepo.FindRolesByMemberID(ctx, memberID)
	if err != nil {
		// ロール取得エラーは空配列で継続
		roleIDs = []common.RoleID{}
	}

	roleIDStrs := make([]string, len(roleIDs))
	for i, roleID := range roleIDs {
		roleIDStrs[i] = roleID.String()
	}

	// Build output
	return &UpdateMemberOutput{
		MemberID:      m.MemberID().String(),
		TenantID:      m.TenantID().String(),
		DisplayName:   m.DisplayName(),
		DiscordUserID: m.DiscordUserID(),
		Email:         m.Email(),
		IsActive:      m.IsActive(),
		RoleIDs:       roleIDStrs,
		UpdatedAt:     m.UpdatedAt().Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}
