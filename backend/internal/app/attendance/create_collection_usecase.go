package attendance

import (
	"context"
	"fmt"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/attendance"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/role"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/services"
)

// CreateCollectionUsecase handles creating an attendance collection
type CreateCollectionUsecase struct {
	repo     attendance.AttendanceCollectionRepository
	roleRepo role.RoleRepository
	clock    services.Clock
}

// NewCreateCollectionUsecase creates a new CreateCollectionUsecase
func NewCreateCollectionUsecase(
	repo attendance.AttendanceCollectionRepository,
	roleRepo role.RoleRepository,
	clock services.Clock,
) *CreateCollectionUsecase {
	return &CreateCollectionUsecase{
		repo:     repo,
		roleRepo: roleRepo,
		clock:    clock,
	}
}

// Execute executes the create collection use case
func (u *CreateCollectionUsecase) Execute(ctx context.Context, input CreateCollectionInput) (*CreateCollectionOutput, error) {
	// 1. Parse TenantID
	tenantID, err := common.ParseTenantID(input.TenantID)
	if err != nil {
		return nil, err
	}

	// 2. Parse TargetType
	targetType, err := attendance.NewTargetType(input.TargetType)
	if err != nil {
		return nil, err
	}

	// 3. Create AttendanceCollection entity (Domain層)
	// Clock から now を取得して渡す
	now := u.clock.Now()
	collection, err := attendance.NewAttendanceCollection(
		now,
		tenantID,
		input.Title,
		input.Description,
		targetType,
		input.TargetID,
		input.Deadline,
	)
	if err != nil {
		return nil, err
	}

	// 4. Save to repository
	if err := u.repo.Save(ctx, collection); err != nil {
		return nil, err
	}

	// 5. Save target dates if provided
	if len(input.TargetDates) > 0 {
		var targetDates []*attendance.TargetDate
		for i, date := range input.TargetDates {
			td, err := attendance.NewTargetDate(now, collection.CollectionID(), date, i)
			if err != nil {
				return nil, err
			}
			targetDates = append(targetDates, td)
		}

		if err := u.repo.SaveTargetDates(ctx, collection.CollectionID(), targetDates); err != nil {
			return nil, err
		}
	}

	// 6. Save group assignments if provided
	if len(input.GroupIDs) > 0 {
		var assignments []*attendance.CollectionGroupAssignment
		for _, groupIDStr := range input.GroupIDs {
			groupID, err := common.ParseMemberGroupID(groupIDStr)
			if err != nil {
				return nil, err
			}
			assignment, err := attendance.NewCollectionGroupAssignment(now, collection.CollectionID(), groupID)
			if err != nil {
				return nil, err
			}
			assignments = append(assignments, assignment)
		}

		if err := u.repo.SaveGroupAssignments(ctx, collection.CollectionID(), assignments); err != nil {
			return nil, err
		}
	}

	// 7. Save role assignments if provided
	if len(input.RoleIDs) > 0 {
		// Parse all role IDs first
		roleIDs := make([]common.RoleID, 0, len(input.RoleIDs))
		for _, roleIDStr := range input.RoleIDs {
			roleID, err := common.ParseRoleID(roleIDStr)
			if err != nil {
				return nil, err
			}
			roleIDs = append(roleIDs, roleID)
		}

		// Batch fetch all roles to validate they exist and belong to this tenant (避免 N+1)
		foundRoles, err := u.roleRepo.FindByIDs(ctx, tenantID, roleIDs)
		if err != nil {
			return nil, err
		}

		// Check that all requested roles were found
		foundRoleMap := make(map[string]bool, len(foundRoles))
		for _, r := range foundRoles {
			foundRoleMap[r.RoleID().String()] = true
		}

		for _, roleIDStr := range input.RoleIDs {
			if !foundRoleMap[roleIDStr] {
				return nil, common.NewValidationError(
					fmt.Sprintf("role not found or not accessible: %s", roleIDStr),
					nil,
				)
			}
		}

		// Create role assignments
		var roleAssignments []*attendance.CollectionRoleAssignment
		for _, roleID := range roleIDs {
			assignment, err := attendance.NewCollectionRoleAssignment(now, collection.CollectionID(), roleID)
			if err != nil {
				return nil, err
			}
			roleAssignments = append(roleAssignments, assignment)
		}

		if err := u.repo.SaveRoleAssignments(ctx, collection.CollectionID(), roleAssignments); err != nil {
			return nil, err
		}
	}

	// 8. Return output DTO
	return &CreateCollectionOutput{
		CollectionID: collection.CollectionID().String(),
		TenantID:     collection.TenantID().String(),
		Title:        collection.Title(),
		Description:  collection.Description(),
		TargetType:   collection.TargetType().String(),
		TargetID:     collection.TargetID(),
		PublicToken:  collection.PublicToken().String(),
		Status:       collection.Status().String(),
		Deadline:     collection.Deadline(),
		CreatedAt:    collection.CreatedAt(),
		UpdatedAt:    collection.UpdatedAt(),
	}, nil
}
