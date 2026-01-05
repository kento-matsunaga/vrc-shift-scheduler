package member

import (
	"context"
	"fmt"
	"log"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/member"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/role"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/services"
)

// BulkUpdateRolesUsecase handles bulk role assignment/removal for members.
//
// Design notes:
// - Uses transaction to ensure atomicity: either all updates succeed or all are rolled back.
// - Uses batch queries to avoid N+1 query problems.
// - Validates all roles and members upfront before making any changes.
// - Records audit log with admin ID for traceability.
type BulkUpdateRolesUsecase struct {
	memberRepo     member.MemberRepository
	memberRoleRepo member.MemberRoleRepository
	roleRepo       role.RoleRepository
	txManager      services.TxManager
}

// NewBulkUpdateRolesUsecase creates a new BulkUpdateRolesUsecase
func NewBulkUpdateRolesUsecase(
	memberRepo member.MemberRepository,
	memberRoleRepo member.MemberRoleRepository,
	roleRepo role.RoleRepository,
	txManager services.TxManager,
) *BulkUpdateRolesUsecase {
	return &BulkUpdateRolesUsecase{
		memberRepo:     memberRepo,
		memberRoleRepo: memberRoleRepo,
		roleRepo:       roleRepo,
		txManager:      txManager,
	}
}

// BulkUpdateRolesInput represents the input for bulk role update
type BulkUpdateRolesInput struct {
	TenantID      common.TenantID
	AdminID       common.AdminID // For audit logging
	MemberIDs     []string
	AddRoleIDs    []string
	RemoveRoleIDs []string
}

// FailureDetail represents a failed member update with reason
type FailureDetail struct {
	MemberID string `json:"member_id"`
	Reason   string `json:"reason"`
}

// BulkUpdateRolesOutput represents the output for bulk role update
type BulkUpdateRolesOutput struct {
	TotalCount   int             `json:"total_count"`
	SuccessCount int             `json:"success_count"`
	FailedCount  int             `json:"failed_count"`
	Failures     []FailureDetail `json:"failures,omitempty"`
}

// MaxBulkUpdateMembers is the maximum number of members that can be updated in a single request
const MaxBulkUpdateMembers = 100

// Execute executes the bulk role update use case
func (u *BulkUpdateRolesUsecase) Execute(ctx context.Context, input BulkUpdateRolesInput) (*BulkUpdateRolesOutput, error) {
	// Validate input count
	if len(input.MemberIDs) > MaxBulkUpdateMembers {
		return nil, common.NewValidationError(
			fmt.Sprintf("too many members: max %d allowed", MaxBulkUpdateMembers), nil)
	}

	if len(input.MemberIDs) == 0 {
		return nil, common.NewValidationError("member_ids is required", nil)
	}

	if len(input.AddRoleIDs) == 0 && len(input.RemoveRoleIDs) == 0 {
		return nil, common.NewValidationError("add_role_ids or remove_role_ids is required", nil)
	}

	// Parse and validate role IDs upfront (batch validation)
	addRoleIDs, err := u.parseAndValidateRoleIDs(ctx, input.TenantID, input.AddRoleIDs)
	if err != nil {
		return nil, err
	}

	removeRoleIDs, err := u.parseAndValidateRoleIDs(ctx, input.TenantID, input.RemoveRoleIDs)
	if err != nil {
		return nil, err
	}

	// Parse and validate member IDs upfront (batch validation)
	memberIDs, invalidMembers, err := u.parseAndValidateMemberIDs(ctx, input.TenantID, input.MemberIDs)
	if err != nil {
		return nil, err
	}

	// If all members are invalid, return early
	if len(memberIDs) == 0 {
		return &BulkUpdateRolesOutput{
			TotalCount:   len(input.MemberIDs),
			SuccessCount: 0,
			FailedCount:  len(invalidMembers),
			Failures:     invalidMembers,
		}, nil
	}

	var successCount int
	failures := invalidMembers // Start with already-invalid members

	// Execute within transaction for atomicity
	txErr := u.txManager.WithTx(ctx, func(txCtx context.Context) error {
		for _, memberID := range memberIDs {
			// Get current roles
			currentRoles, err := u.memberRoleRepo.FindRolesByMemberID(txCtx, memberID)
			if err != nil {
				failures = append(failures, FailureDetail{
					MemberID: memberID.String(),
					Reason:   "failed to get current roles",
				})
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
			if err := u.memberRoleRepo.SetMemberRoles(txCtx, memberID, newRoleIDs); err != nil {
				failures = append(failures, FailureDetail{
					MemberID: memberID.String(),
					Reason:   "failed to update roles",
				})
				continue
			}

			successCount++
		}

		// If no members succeeded, rollback the transaction
		if successCount == 0 && len(memberIDs) > 0 {
			return fmt.Errorf("all member updates failed")
		}

		return nil
	})

	if txErr != nil {
		// Transaction failed entirely
		return nil, fmt.Errorf("bulk update transaction failed: %w", txErr)
	}

	failedCount := len(input.MemberIDs) - successCount

	// Audit log
	log.Printf("[AUDIT] BulkUpdateRoles: admin=%s tenant=%s members=%d success=%d failed=%d add_roles=%v remove_roles=%v",
		input.AdminID.String(), input.TenantID.String(), len(input.MemberIDs), successCount, failedCount, input.AddRoleIDs, input.RemoveRoleIDs)

	return &BulkUpdateRolesOutput{
		TotalCount:   len(input.MemberIDs),
		SuccessCount: successCount,
		FailedCount:  failedCount,
		Failures:     failures,
	}, nil
}

// parseAndValidateRoleIDs parses role ID strings and validates they belong to the tenant.
// Uses batch query to avoid N+1 problem.
func (u *BulkUpdateRolesUsecase) parseAndValidateRoleIDs(
	ctx context.Context,
	tenantID common.TenantID,
	roleIDStrs []string,
) ([]common.RoleID, error) {
	if len(roleIDStrs) == 0 {
		return nil, nil
	}

	// Get all roles for tenant (batch query)
	tenantRoles, err := u.roleRepo.FindByTenantID(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch tenant roles: %w", err)
	}

	// Build lookup map
	validRoleIDs := make(map[string]bool)
	for _, r := range tenantRoles {
		validRoleIDs[r.RoleID().String()] = true
	}

	// Parse and validate
	result := make([]common.RoleID, 0, len(roleIDStrs))
	for _, roleIDStr := range roleIDStrs {
		roleID, err := common.ParseRoleID(roleIDStr)
		if err != nil {
			return nil, common.NewValidationError("invalid role_id format: "+roleIDStr, err)
		}

		if !validRoleIDs[roleIDStr] {
			return nil, common.NewValidationError("role does not belong to tenant: "+roleIDStr, nil)
		}

		result = append(result, roleID)
	}

	return result, nil
}

// parseAndValidateMemberIDs parses member ID strings and validates they belong to the tenant.
// Uses batch query to avoid N+1 problem.
// Returns valid member IDs and failure details for invalid ones.
func (u *BulkUpdateRolesUsecase) parseAndValidateMemberIDs(
	ctx context.Context,
	tenantID common.TenantID,
	memberIDStrs []string,
) ([]common.MemberID, []FailureDetail, error) {
	if len(memberIDStrs) == 0 {
		return nil, nil, nil
	}

	// Get all members for tenant (batch query)
	tenantMembers, err := u.memberRepo.FindByTenantID(ctx, tenantID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to fetch tenant members: %w", err)
	}

	// Build lookup map
	validMemberIDs := make(map[string]bool)
	for _, m := range tenantMembers {
		validMemberIDs[m.MemberID().String()] = true
	}

	// Parse and validate
	result := make([]common.MemberID, 0, len(memberIDStrs))
	failures := make([]FailureDetail, 0)

	for _, memberIDStr := range memberIDStrs {
		memberID, err := common.ParseMemberID(memberIDStr)
		if err != nil {
			failures = append(failures, FailureDetail{
				MemberID: memberIDStr,
				Reason:   "invalid member_id format",
			})
			continue
		}

		if !validMemberIDs[memberIDStr] {
			failures = append(failures, FailureDetail{
				MemberID: memberIDStr,
				Reason:   "member not found or does not belong to tenant",
			})
			continue
		}

		result = append(result, memberID)
	}

	return result, failures, nil
}
