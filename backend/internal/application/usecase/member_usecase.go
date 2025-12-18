package usecase

import (
	"context"
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/member"
)

// MemberRepository defines the interface for member persistence
type MemberRepository interface {
	Save(ctx context.Context, member *member.Member) error
	FindByID(ctx context.Context, tenantID common.TenantID, memberID common.MemberID) (*member.Member, error)
	FindByTenantID(ctx context.Context, tenantID common.TenantID) ([]*member.Member, error)
	FindActiveByTenantID(ctx context.Context, tenantID common.TenantID) ([]*member.Member, error)
	FindByDiscordUserID(ctx context.Context, tenantID common.TenantID, discordUserID string) (*member.Member, error)
	FindByEmail(ctx context.Context, tenantID common.TenantID, email string) (*member.Member, error)
	ExistsByDiscordUserID(ctx context.Context, tenantID common.TenantID, discordUserID string) (bool, error)
	ExistsByEmail(ctx context.Context, tenantID common.TenantID, email string) (bool, error)
}

// MemberRoleRepository defines the interface for member-role association persistence
type MemberRoleRepository interface {
	FindRolesByMemberID(ctx context.Context, memberID common.MemberID) ([]common.RoleID, error)
}

// CreateMemberInput represents the input for creating a member
type CreateMemberInput struct {
	TenantID      common.TenantID
	DisplayName   string
	DiscordUserID string
	Email         string
}

// CreateMemberUsecase handles the member creation use case
type CreateMemberUsecase struct {
	memberRepo MemberRepository
}

// NewCreateMemberUsecase creates a new CreateMemberUsecase
func NewCreateMemberUsecase(memberRepo MemberRepository) *CreateMemberUsecase {
	return &CreateMemberUsecase{
		memberRepo: memberRepo,
	}
}

// Execute creates a new member
func (uc *CreateMemberUsecase) Execute(ctx context.Context, input CreateMemberInput) (*member.Member, error) {
	// Discord User ID の重複チェック
	if input.DiscordUserID != "" {
		exists, err := uc.memberRepo.ExistsByDiscordUserID(ctx, input.TenantID, input.DiscordUserID)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, common.NewConflictError("This discord_user_id is already registered")
		}
	}

	// Email の重複チェック
	if input.Email != "" {
		exists, err := uc.memberRepo.ExistsByEmail(ctx, input.TenantID, input.Email)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, common.NewConflictError("This email is already registered")
		}
	}

	// Member エンティティの作成
	newMember, err := member.NewMember(
		time.Now(),
		input.TenantID,
		input.DisplayName,
		input.DiscordUserID,
		input.Email,
	)
	if err != nil {
		return nil, err
	}

	// 保存
	if err := uc.memberRepo.Save(ctx, newMember); err != nil {
		return nil, err
	}

	return newMember, nil
}

// ListMembersInput represents the input for listing members
type ListMembersInput struct {
	TenantID common.TenantID
	IsActive *bool // optional: nil means no filter, true means active only, false means inactive only
}

// MemberWithRoles represents a member with their assigned roles
type MemberWithRoles struct {
	Member  *member.Member
	RoleIDs []common.RoleID
}

// ListMembersUsecase handles the member listing use case
type ListMembersUsecase struct {
	memberRepo     MemberRepository
	memberRoleRepo MemberRoleRepository
}

// NewListMembersUsecase creates a new ListMembersUsecase
func NewListMembersUsecase(
	memberRepo MemberRepository,
	memberRoleRepo MemberRoleRepository,
) *ListMembersUsecase {
	return &ListMembersUsecase{
		memberRepo:     memberRepo,
		memberRoleRepo: memberRoleRepo,
	}
}

// Execute retrieves members for a tenant with optional filtering and role aggregation
func (uc *ListMembersUsecase) Execute(ctx context.Context, input ListMembersInput) ([]*MemberWithRoles, error) {
	// メンバー一覧を取得
	members, err := uc.memberRepo.FindByTenantID(ctx, input.TenantID)
	if err != nil {
		return nil, err
	}

	// is_active フィルタ
	var filteredMembers []*member.Member
	if input.IsActive != nil {
		for _, m := range members {
			if *input.IsActive && m.IsActive() {
				filteredMembers = append(filteredMembers, m)
			} else if !*input.IsActive && !m.IsActive() {
				filteredMembers = append(filteredMembers, m)
			}
		}
	} else {
		filteredMembers = members
	}

	// 各メンバーのロールを取得
	result := make([]*MemberWithRoles, 0, len(filteredMembers))
	for _, m := range filteredMembers {
		roleIDs, err := uc.memberRoleRepo.FindRolesByMemberID(ctx, m.MemberID())
		if err != nil {
			// エラー時は空配列として継続
			roleIDs = []common.RoleID{}
		}

		result = append(result, &MemberWithRoles{
			Member:  m,
			RoleIDs: roleIDs,
		})
	}

	return result, nil
}

// GetMemberInput represents the input for getting a member
type GetMemberInput struct {
	TenantID common.TenantID
	MemberID common.MemberID
}

// GetMemberUsecase handles the member retrieval use case
type GetMemberUsecase struct {
	memberRepo     MemberRepository
	memberRoleRepo MemberRoleRepository
}

// NewGetMemberUsecase creates a new GetMemberUsecase
func NewGetMemberUsecase(
	memberRepo MemberRepository,
	memberRoleRepo MemberRoleRepository,
) *GetMemberUsecase {
	return &GetMemberUsecase{
		memberRepo:     memberRepo,
		memberRoleRepo: memberRoleRepo,
	}
}

// Execute retrieves a member by ID with role aggregation
func (uc *GetMemberUsecase) Execute(ctx context.Context, input GetMemberInput) (*MemberWithRoles, error) {
	// メンバーの取得
	m, err := uc.memberRepo.FindByID(ctx, input.TenantID, input.MemberID)
	if err != nil {
		return nil, err
	}

	// ロールの取得
	roleIDs, err := uc.memberRoleRepo.FindRolesByMemberID(ctx, m.MemberID())
	if err != nil {
		// エラー時は空配列として継続
		roleIDs = []common.RoleID{}
	}

	return &MemberWithRoles{
		Member:  m,
		RoleIDs: roleIDs,
	}, nil
}
