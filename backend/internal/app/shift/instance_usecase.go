package shift

import (
	"context"
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/event"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/shift"
)

// CreateInstanceInput represents the input for creating an instance
type CreateInstanceInput struct {
	TenantID     common.TenantID
	EventID      common.EventID
	Name         string
	DisplayOrder int
	MaxMembers   *int
}

// CreateInstanceUsecase handles the instance creation use case
type CreateInstanceUsecase struct {
	instanceRepo shift.InstanceRepository
	eventRepo    event.EventRepository
}

// NewCreateInstanceUsecase creates a new CreateInstanceUsecase
func NewCreateInstanceUsecase(
	instanceRepo shift.InstanceRepository,
	eventRepo event.EventRepository,
) *CreateInstanceUsecase {
	return &CreateInstanceUsecase{
		instanceRepo: instanceRepo,
		eventRepo:    eventRepo,
	}
}

// Execute creates a new instance
func (uc *CreateInstanceUsecase) Execute(ctx context.Context, input CreateInstanceInput) (*shift.Instance, error) {
	// Event の存在確認
	_, err := uc.eventRepo.FindByID(ctx, input.TenantID, input.EventID)
	if err != nil {
		return nil, err
	}

	// 同名のインスタンスが既に存在しないか確認
	existing, err := uc.instanceRepo.FindByEventIDAndName(ctx, input.TenantID, input.EventID, input.Name)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, common.NewValidationError("instance with the same name already exists", nil)
	}

	// Instance エンティティの作成
	newInstance, err := shift.NewInstance(
		time.Now(),
		input.TenantID,
		input.EventID,
		input.Name,
		input.DisplayOrder,
		input.MaxMembers,
	)
	if err != nil {
		return nil, err
	}

	// 保存
	if err := uc.instanceRepo.Save(ctx, newInstance); err != nil {
		return nil, err
	}

	return newInstance, nil
}

// ListInstancesInput represents the input for listing instances
type ListInstancesInput struct {
	TenantID common.TenantID
	EventID  common.EventID
}

// ListInstancesUsecase handles the instance listing use case
type ListInstancesUsecase struct {
	instanceRepo shift.InstanceRepository
}

// NewListInstancesUsecase creates a new ListInstancesUsecase
func NewListInstancesUsecase(instanceRepo shift.InstanceRepository) *ListInstancesUsecase {
	return &ListInstancesUsecase{
		instanceRepo: instanceRepo,
	}
}

// Execute retrieves instances for an event
func (uc *ListInstancesUsecase) Execute(ctx context.Context, input ListInstancesInput) ([]*shift.Instance, error) {
	return uc.instanceRepo.FindByEventID(ctx, input.TenantID, input.EventID)
}

// GetInstanceInput represents the input for getting an instance
type GetInstanceInput struct {
	TenantID   common.TenantID
	InstanceID shift.InstanceID
}

// GetInstanceUsecase handles the instance retrieval use case
type GetInstanceUsecase struct {
	instanceRepo shift.InstanceRepository
}

// NewGetInstanceUsecase creates a new GetInstanceUsecase
func NewGetInstanceUsecase(instanceRepo shift.InstanceRepository) *GetInstanceUsecase {
	return &GetInstanceUsecase{
		instanceRepo: instanceRepo,
	}
}

// Execute retrieves an instance by ID
func (uc *GetInstanceUsecase) Execute(ctx context.Context, input GetInstanceInput) (*shift.Instance, error) {
	return uc.instanceRepo.FindByID(ctx, input.TenantID, input.InstanceID)
}

// UpdateInstanceInput represents the input for updating an instance
type UpdateInstanceInput struct {
	TenantID     common.TenantID
	InstanceID   shift.InstanceID
	Name         *string
	DisplayOrder *int
	MaxMembers   *int // nil means no change, pointer to nil means set to null
}

// UpdateInstanceUsecase handles the instance update use case
type UpdateInstanceUsecase struct {
	instanceRepo shift.InstanceRepository
}

// NewUpdateInstanceUsecase creates a new UpdateInstanceUsecase
func NewUpdateInstanceUsecase(instanceRepo shift.InstanceRepository) *UpdateInstanceUsecase {
	return &UpdateInstanceUsecase{
		instanceRepo: instanceRepo,
	}
}

// Execute updates an instance
func (uc *UpdateInstanceUsecase) Execute(ctx context.Context, input UpdateInstanceInput) (*shift.Instance, error) {
	// インスタンスの取得
	instance, err := uc.instanceRepo.FindByID(ctx, input.TenantID, input.InstanceID)
	if err != nil {
		return nil, err
	}

	// 名前の更新
	if input.Name != nil {
		// 同名のインスタンスが既に存在しないか確認
		existing, err := uc.instanceRepo.FindByEventIDAndName(ctx, input.TenantID, instance.EventID(), *input.Name)
		if err != nil {
			return nil, err
		}
		if existing != nil && existing.InstanceID() != instance.InstanceID() {
			return nil, common.NewValidationError("instance with the same name already exists", nil)
		}

		if err := instance.UpdateName(*input.Name); err != nil {
			return nil, err
		}
	}

	// 表示順の更新
	if input.DisplayOrder != nil {
		instance.UpdateDisplayOrder(*input.DisplayOrder)
	}

	// 最大人数の更新
	if input.MaxMembers != nil {
		if err := instance.UpdateMaxMembers(input.MaxMembers); err != nil {
			return nil, err
		}
	}

	// 保存
	if err := uc.instanceRepo.Save(ctx, instance); err != nil {
		return nil, err
	}

	return instance, nil
}

// DeleteInstanceInput represents the input for deleting an instance
type DeleteInstanceInput struct {
	TenantID   common.TenantID
	InstanceID shift.InstanceID
}

// DeleteInstanceUsecase handles the instance deletion use case
type DeleteInstanceUsecase struct {
	instanceRepo shift.InstanceRepository
}

// NewDeleteInstanceUsecase creates a new DeleteInstanceUsecase
func NewDeleteInstanceUsecase(instanceRepo shift.InstanceRepository) *DeleteInstanceUsecase {
	return &DeleteInstanceUsecase{
		instanceRepo: instanceRepo,
	}
}

// Execute deletes an instance
func (uc *DeleteInstanceUsecase) Execute(ctx context.Context, input DeleteInstanceInput) error {
	// インスタンスの存在確認
	_, err := uc.instanceRepo.FindByID(ctx, input.TenantID, input.InstanceID)
	if err != nil {
		return err
	}

	// 削除（シフト枠が紐付いている場合は FK 制約でエラーになる）
	return uc.instanceRepo.Delete(ctx, input.TenantID, input.InstanceID)
}

// FindOrCreateInstanceInput represents the input for finding or creating an instance
type FindOrCreateInstanceInput struct {
	TenantID     common.TenantID
	EventID      common.EventID
	Name         string
	DisplayOrder int
}

// FindOrCreateInstanceUsecase handles finding or creating an instance
// This is used by the template applier to ensure instances exist
type FindOrCreateInstanceUsecase struct {
	instanceRepo shift.InstanceRepository
}

// NewFindOrCreateInstanceUsecase creates a new FindOrCreateInstanceUsecase
func NewFindOrCreateInstanceUsecase(instanceRepo shift.InstanceRepository) *FindOrCreateInstanceUsecase {
	return &FindOrCreateInstanceUsecase{
		instanceRepo: instanceRepo,
	}
}

// Execute finds an instance by name or creates a new one if not found
func (uc *FindOrCreateInstanceUsecase) Execute(ctx context.Context, input FindOrCreateInstanceInput) (*shift.Instance, error) {
	// 既存のインスタンスを検索
	existing, err := uc.instanceRepo.FindByEventIDAndName(ctx, input.TenantID, input.EventID, input.Name)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return existing, nil
	}

	// 新規作成
	newInstance, err := shift.NewInstance(
		time.Now(),
		input.TenantID,
		input.EventID,
		input.Name,
		input.DisplayOrder,
		nil, // max_members はデフォルトでnull
	)
	if err != nil {
		return nil, err
	}

	if err := uc.instanceRepo.Save(ctx, newInstance); err != nil {
		return nil, err
	}

	return newInstance, nil
}
