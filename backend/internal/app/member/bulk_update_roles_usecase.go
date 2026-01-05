package member

import (
	"context"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/member"
)

// BulkUpdateRolesUsecase handles bulk role assignment/removal for members
type BulkUpdateRolesUsecase struct {
	memberRepo     member.MemberRepository
	memberRoleRepo member.MemberRoleRepository
}

// NewBulkUpdateRolesUsecase creates a new BulkUpdateRolesUsecase
func NewBulkUpdateRolesUsecase(
	memberRepo member.MemberRepository,
	memberRoleRepo member.MemberRoleRepository,
) *BulkUpdateRolesUsecase {
	return &BulkUpdateRolesUsecase{
		memberRepo:     memberRepo,
		memberRoleRepo: memberRoleRepo,
	}
}

// BulkUpdateRolesInput represents the input for bulk role update
type BulkUpdateRolesInput struct {
	TenantID      common.TenantID
	MemberIDs     []string
	AddRoleIDs    []string
	RemoveRoleIDs []string
}

// BulkUpdateRolesOutput represents the output for bulk role update
type BulkUpdateRolesOutput struct {
	TotalCount   int `json:"total_count"`
	SuccessCount int `json:"success_count"`
	FailedCount  int `json:"failed_count"`
}

// Execute executes the bulk role update use case
func (u *BulkUpdateRolesUsecase) Execute(ctx context.Context, input BulkUpdateRolesInput) (*BulkUpdateRolesOutput, error) {
	// Parse role IDs to add
	addRoleIDs := make([]common.RoleID, 0, len(input.AddRoleIDs))
	for _, roleIDStr := range input.AddRoleIDs {
		roleID, err := common.ParseRoleID(roleIDStr)
		if err != nil {
			return nil, common.NewValidationError("invalid add_role_id: "+roleIDStr, err)
		}
		addRoleIDs = append(addRoleIDs, roleID)
	}

	// Parse role IDs to remove
	removeRoleIDs := make([]common.RoleID, 0, len(input.RemoveRoleIDs))
	for _, roleIDStr := range input.RemoveRoleIDs {
		roleID, err := common.ParseRoleID(roleIDStr)
		if err != nil {
			return nil, common.NewValidationError("invalid remove_role_id: "+roleIDStr, err)
		}
		removeRoleIDs = append(removeRoleIDs, roleID)
	}

	successCount := 0
	failedCount := 0

	// Process each member
	for _, memberIDStr := range input.MemberIDs {
		memberID, err := common.ParseMemberID(memberIDStr)
		if err != nil {
			failedCount++
			continue
		}

		// Verify member exists and belongs to the tenant
		_, err = u.memberRepo.FindByID(ctx, input.TenantID, memberID)
		if err != nil {
			failedCount++
			continue
		}

		// Get current roles
		currentRoles, err := u.memberRoleRepo.FindRolesByMemberID(ctx, memberID)
		if err != nil {
			failedCount++
			continue
		}

		// Build new role set
		roleSet := make(map[common.RoleID]bool)
		for _, roleID := range currentRoles {
			roleSet[roleID] = true
		}

		// Add new roles
		for _, roleID := range addRoleIDs {
			roleSet[roleID] = true
		}

		// Remove roles
		for _, roleID := range removeRoleIDs {
			delete(roleSet, roleID)
		}

		// Convert back to slice
		newRoleIDs := make([]common.RoleID, 0, len(roleSet))
		for roleID := range roleSet {
			newRoleIDs = append(newRoleIDs, roleID)
		}

		// Set new roles
		if err := u.memberRoleRepo.SetMemberRoles(ctx, memberID, newRoleIDs); err != nil {
			failedCount++
			continue
		}

		successCount++
	}

	return &BulkUpdateRolesOutput{
		TotalCount:   len(input.MemberIDs),
		SuccessCount: successCount,
		FailedCount:  failedCount,
	}, nil
}
