package attendance

import (
	"context"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/attendance"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/services"
)

// CloseCollectionUsecase handles closing an attendance collection
type CloseCollectionUsecase struct {
	repo  attendance.AttendanceCollectionRepository
	clock services.Clock
}

// NewCloseCollectionUsecase creates a new CloseCollectionUsecase
func NewCloseCollectionUsecase(
	repo attendance.AttendanceCollectionRepository,
	clock services.Clock,
) *CloseCollectionUsecase {
	return &CloseCollectionUsecase{
		repo:  repo,
		clock: clock,
	}
}

// Execute executes the close collection use case
func (u *CloseCollectionUsecase) Execute(ctx context.Context, input CloseCollectionInput) (*CloseCollectionOutput, error) {
	// 1. Parse TenantID
	tenantID, err := common.ParseTenantID(input.TenantID)
	if err != nil {
		return nil, err
	}

	// 2. Parse CollectionID
	collectionID, err := common.ParseCollectionID(input.CollectionID)
	if err != nil {
		return nil, err
	}

	// 3. Find collection by ID (with tenant isolation)
	collection, err := u.repo.FindByID(ctx, tenantID, collectionID)
	if err != nil {
		return nil, err
	}

	// 4. Close collection (domain rule)
	now := u.clock.Now()
	if err := collection.Close(now); err != nil {
		return nil, err
	}

	// 5. Save updated collection
	if err := u.repo.Save(ctx, collection); err != nil {
		return nil, err
	}

	// 6. Return output DTO
	return &CloseCollectionOutput{
		CollectionID: collection.CollectionID().String(),
		Status:       collection.Status().String(),
		UpdatedAt:    collection.UpdatedAt(),
	}, nil
}
