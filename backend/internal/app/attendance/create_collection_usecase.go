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
	repo      attendance.AttendanceCollectionRepository
	roleRepo  role.RoleRepository
	txManager services.TxManager
	clock     services.Clock
}

// NewCreateCollectionUsecase creates a new CreateCollectionUsecase
func NewCreateCollectionUsecase(
	repo attendance.AttendanceCollectionRepository,
	roleRepo role.RoleRepository,
	txManager services.TxManager,
	clock services.Clock,
) *CreateCollectionUsecase {
	return &CreateCollectionUsecase{
		repo:      repo,
		roleRepo:  roleRepo,
		txManager: txManager,
		clock:     clock,
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

	// 3. Parse GroupIDs upfront (validation before transaction)
	var groupIDs []common.MemberGroupID
	for _, groupIDStr := range input.GroupIDs {
		groupID, err := common.ParseMemberGroupID(groupIDStr)
		if err != nil {
			return nil, err
		}
		groupIDs = append(groupIDs, groupID)
	}

	// 4. Parse RoleIDs upfront (validation before transaction)
	var roleIDs []common.RoleID
	for _, roleIDStr := range input.RoleIDs {
		roleID, err := common.ParseRoleID(roleIDStr)
		if err != nil {
			return nil, err
		}
		roleIDs = append(roleIDs, roleID)
	}

	// 5. Create AttendanceCollection entity (Domain層)
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

	// 6. Create target dates entities upfront
	var targetDates []*attendance.TargetDate
	for i, dateInput := range input.TargetDates {
		td, err := attendance.NewTargetDate(now, collection.CollectionID(), dateInput.TargetDate, dateInput.StartTime, dateInput.EndTime, i)
		if err != nil {
			return nil, err
		}
		targetDates = append(targetDates, td)
	}

	// 7. Create group assignments entities upfront
	var groupAssignments []*attendance.CollectionGroupAssignment
	for _, groupID := range groupIDs {
		assignment, err := attendance.NewCollectionGroupAssignment(now, collection.CollectionID(), groupID)
		if err != nil {
			return nil, err
		}
		groupAssignments = append(groupAssignments, assignment)
	}

	// 8. Validate roles exist before transaction (read operation)
	if len(roleIDs) > 0 {
		foundRoles, err := u.roleRepo.FindByIDs(ctx, tenantID, roleIDs)
		if err != nil {
			return nil, err
		}

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
	}

	// 9. Create role assignments entities upfront
	var roleAssignments []*attendance.CollectionRoleAssignment
	for _, roleID := range roleIDs {
		assignment, err := attendance.NewCollectionRoleAssignment(now, collection.CollectionID(), roleID)
		if err != nil {
			return nil, err
		}
		roleAssignments = append(roleAssignments, assignment)
	}

	// 10. Execute all save operations within a transaction
	err = u.txManager.WithTx(ctx, func(txCtx context.Context) error {
		// Save collection
		if err := u.repo.Save(txCtx, collection); err != nil {
			return err
		}

		// Save target dates if provided
		if len(targetDates) > 0 {
			if err := u.repo.SaveTargetDates(txCtx, collection.CollectionID(), targetDates); err != nil {
				return err
			}
		}

		// Save group assignments if provided
		if len(groupAssignments) > 0 {
			if err := u.repo.SaveGroupAssignments(txCtx, collection.CollectionID(), groupAssignments); err != nil {
				return err
			}
		}

		// Save role assignments if provided
		if len(roleAssignments) > 0 {
			if err := u.repo.SaveRoleAssignments(txCtx, collection.CollectionID(), roleAssignments); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// 11. Return output DTO
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
