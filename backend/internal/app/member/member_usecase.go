package member

import (
	"context"
	"log"
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
	RoleIDs       []string // Optional role IDs to assign
}

// CreateMemberUsecase handles the member creation use case
type CreateMemberUsecase struct {
	memberRepo     MemberRepository
	memberRoleRepo MemberRoleAssigner
}

// NewCreateMemberUsecase creates a new CreateMemberUsecase
func NewCreateMemberUsecase(memberRepo MemberRepository, memberRoleRepo MemberRoleAssigner) *CreateMemberUsecase {
	return &CreateMemberUsecase{
		memberRepo:     memberRepo,
		memberRoleRepo: memberRoleRepo,
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

	// ロール割り当て
	if len(input.RoleIDs) > 0 && uc.memberRoleRepo != nil {
		roleIDs := make([]common.RoleID, len(input.RoleIDs))
		for i, rid := range input.RoleIDs {
			roleIDs[i] = common.RoleID(rid)
		}
		// ロール割り当て失敗はログに記録するが、メンバー作成は成功とする
		// （BulkImportMembersUsecaseと同じ方針）
		if err := uc.memberRoleRepo.SetMemberRoles(ctx, newMember.MemberID(), roleIDs); err != nil {
			log.Printf("[WARN] Failed to assign roles to member %s: %v", newMember.MemberID(), err)
		}
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

// DeleteMemberInput represents the input for deleting a member
type DeleteMemberInput struct {
	TenantID common.TenantID
	MemberID common.MemberID
}

// DeleteMemberUsecase handles the member deletion use case
type DeleteMemberUsecase struct {
	memberRepo MemberRepository
}

// NewDeleteMemberUsecase creates a new DeleteMemberUsecase
func NewDeleteMemberUsecase(memberRepo MemberRepository) *DeleteMemberUsecase {
	return &DeleteMemberUsecase{
		memberRepo: memberRepo,
	}
}

// Execute deletes a member (soft delete)
func (uc *DeleteMemberUsecase) Execute(ctx context.Context, input DeleteMemberInput) error {
	// メンバーの取得
	m, err := uc.memberRepo.FindByID(ctx, input.TenantID, input.MemberID)
	if err != nil {
		return err
	}

	// 削除（soft delete）
	now := time.Now()
	m.Delete(now)

	// 保存
	return uc.memberRepo.Save(ctx, m)
}

// BulkImportMemberInput represents a single member for bulk import
type BulkImportMemberInput struct {
	DisplayName string
	RoleIDs     []string
}

// BulkImportMembersInput represents the input for bulk importing members
type BulkImportMembersInput struct {
	TenantID common.TenantID
	Members  []BulkImportMemberInput
}

// BulkImportMemberResult represents the result of importing a single member
type BulkImportMemberResult struct {
	DisplayName string `json:"display_name"`
	Success     bool   `json:"success"`
	MemberID    string `json:"member_id,omitempty"`
	Error       string `json:"error,omitempty"`
}

// BulkImportMembersOutput represents the output of bulk importing members
type BulkImportMembersOutput struct {
	TotalCount   int                      `json:"total_count"`
	SuccessCount int                      `json:"success_count"`
	FailedCount  int                      `json:"failed_count"`
	Results      []BulkImportMemberResult `json:"results"`
}

// MemberRoleAssigner defines the interface for assigning roles to members
type MemberRoleAssigner interface {
	SetMemberRoles(ctx context.Context, memberID common.MemberID, roleIDs []common.RoleID) error
}

// BulkImportMembersUsecase handles the bulk member import use case
type BulkImportMembersUsecase struct {
	memberRepo     MemberRepository
	memberRoleRepo MemberRoleAssigner
}

// NewBulkImportMembersUsecase creates a new BulkImportMembersUsecase
func NewBulkImportMembersUsecase(memberRepo MemberRepository, memberRoleRepo MemberRoleAssigner) *BulkImportMembersUsecase {
	return &BulkImportMembersUsecase{
		memberRepo:     memberRepo,
		memberRoleRepo: memberRoleRepo,
	}
}

// Execute imports multiple members at once
func (uc *BulkImportMembersUsecase) Execute(ctx context.Context, input BulkImportMembersInput) (*BulkImportMembersOutput, error) {
	results := make([]BulkImportMemberResult, 0, len(input.Members))
	successCount := 0
	failedCount := 0

	for _, memberInput := range input.Members {
		result := BulkImportMemberResult{
			DisplayName: memberInput.DisplayName,
		}

		// バリデーション
		if memberInput.DisplayName == "" {
			result.Success = false
			result.Error = "display_name is required"
			failedCount++
			results = append(results, result)
			continue
		}

		if len(memberInput.DisplayName) > 50 {
			result.Success = false
			result.Error = "display_name must be 50 characters or less"
			failedCount++
			results = append(results, result)
			continue
		}

		// Member エンティティの作成
		newMember, err := member.NewMember(
			time.Now(),
			input.TenantID,
			memberInput.DisplayName,
			"", // discord_user_id は空
			"", // email は空
		)
		if err != nil {
			result.Success = false
			result.Error = err.Error()
			failedCount++
			results = append(results, result)
			continue
		}

		// 保存
		if err := uc.memberRepo.Save(ctx, newMember); err != nil {
			result.Success = false
			result.Error = "Failed to save member"
			failedCount++
			results = append(results, result)
			continue
		}

		// ロール割り当て
		if len(memberInput.RoleIDs) > 0 {
			roleIDs := make([]common.RoleID, len(memberInput.RoleIDs))
			for i, rid := range memberInput.RoleIDs {
				roleIDs[i] = common.RoleID(rid)
			}
			// ロール割り当て失敗でもメンバー作成は成功とする
			// エラーは無視する
			_ = uc.memberRoleRepo.SetMemberRoles(ctx, newMember.MemberID(), roleIDs)
		}

		result.Success = true
		result.MemberID = newMember.MemberID().String()
		successCount++
		results = append(results, result)
	}

	return &BulkImportMembersOutput{
		TotalCount:   len(input.Members),
		SuccessCount: successCount,
		FailedCount:  failedCount,
		Results:      results,
	}, nil
}
