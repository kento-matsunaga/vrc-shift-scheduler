package shift

import (
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
)

// InstanceID represents an instance identifier
type InstanceID string

// NewInstanceIDWithTime creates a new InstanceID using the provided time.
func NewInstanceIDWithTime(t time.Time) InstanceID {
	return InstanceID(common.NewULIDWithTime(t))
}

// NewInstanceID creates a new InstanceID using the current time.
// Deprecated: Use NewInstanceIDWithTime for better testability.
func NewInstanceID() InstanceID {
	return InstanceID(common.NewULID())
}

func (id InstanceID) String() string {
	return string(id)
}

func (id InstanceID) Validate() error {
	if id == "" {
		return common.NewValidationError("instance_id is required", nil)
	}
	return common.ValidateULID(string(id))
}

func ParseInstanceID(s string) (InstanceID, error) {
	if err := common.ValidateULID(s); err != nil {
		return "", err
	}
	return InstanceID(s), nil
}

// Instance represents an instance entity (独立したエンティティ)
// Event に属するが、Event集約には含まれない
// シフト枠（ShiftSlot）はこのインスタンスに紐づく
type Instance struct {
	instanceID   InstanceID
	tenantID     common.TenantID
	eventID      common.EventID
	name         string
	displayOrder int
	maxMembers   *int // NULL許容
	createdAt    time.Time
	updatedAt    time.Time
	deletedAt    *time.Time
}

// NewInstance creates a new Instance entity
func NewInstance(
	now time.Time,
	tenantID common.TenantID,
	eventID common.EventID,
	name string,
	displayOrder int,
	maxMembers *int,
) (*Instance, error) {
	instance := &Instance{
		instanceID:   NewInstanceIDWithTime(now),
		tenantID:     tenantID,
		eventID:      eventID,
		name:         name,
		displayOrder: displayOrder,
		maxMembers:   maxMembers,
		createdAt:    now,
		updatedAt:    now,
	}

	if err := instance.validate(); err != nil {
		return nil, err
	}

	return instance, nil
}

// ReconstructInstance reconstructs an Instance from persistence
func ReconstructInstance(
	instanceID InstanceID,
	tenantID common.TenantID,
	eventID common.EventID,
	name string,
	displayOrder int,
	maxMembers *int,
	createdAt time.Time,
	updatedAt time.Time,
	deletedAt *time.Time,
) (*Instance, error) {
	instance := &Instance{
		instanceID:   instanceID,
		tenantID:     tenantID,
		eventID:      eventID,
		name:         name,
		displayOrder: displayOrder,
		maxMembers:   maxMembers,
		createdAt:    createdAt,
		updatedAt:    updatedAt,
		deletedAt:    deletedAt,
	}

	if err := instance.validate(); err != nil {
		return nil, err
	}

	return instance, nil
}

func (i *Instance) validate() error {
	// TenantID の必須性チェック
	if err := i.tenantID.Validate(); err != nil {
		return common.NewValidationError("tenant_id is required", err)
	}

	// EventID の必須性チェック
	if err := i.eventID.Validate(); err != nil {
		return common.NewValidationError("event_id is required", err)
	}

	// Name の必須性チェック
	if i.name == "" {
		return common.NewValidationError("instance name is required", nil)
	}

	if len(i.name) > 255 {
		return common.NewValidationError("instance name must be less than 255 characters", nil)
	}

	// MaxMembers の範囲チェック（指定されている場合は1以上）
	if i.maxMembers != nil && *i.maxMembers < 1 {
		return common.NewValidationError("max_members must be at least 1", nil)
	}

	return nil
}

// Getters

func (i *Instance) InstanceID() InstanceID {
	return i.instanceID
}

func (i *Instance) TenantID() common.TenantID {
	return i.tenantID
}

func (i *Instance) EventID() common.EventID {
	return i.eventID
}

func (i *Instance) Name() string {
	return i.name
}

func (i *Instance) DisplayOrder() int {
	return i.displayOrder
}

func (i *Instance) MaxMembers() *int {
	return i.maxMembers
}

func (i *Instance) CreatedAt() time.Time {
	return i.createdAt
}

func (i *Instance) UpdatedAt() time.Time {
	return i.updatedAt
}

func (i *Instance) DeletedAt() *time.Time {
	return i.deletedAt
}

func (i *Instance) IsDeleted() bool {
	return i.deletedAt != nil
}

// UpdateName updates the instance name
func (i *Instance) UpdateName(name string) error {
	if name == "" {
		return common.NewValidationError("instance name is required", nil)
	}
	if len(name) > 255 {
		return common.NewValidationError("instance name must be less than 255 characters", nil)
	}

	i.name = name
	i.updatedAt = time.Now()
	return nil
}

// UpdateDisplayOrder updates the display order
func (i *Instance) UpdateDisplayOrder(displayOrder int) {
	i.displayOrder = displayOrder
	i.updatedAt = time.Now()
}

// UpdateMaxMembers updates the max members
func (i *Instance) UpdateMaxMembers(maxMembers *int) error {
	if maxMembers != nil && *maxMembers < 1 {
		return common.NewValidationError("max_members must be at least 1", nil)
	}

	i.maxMembers = maxMembers
	i.updatedAt = time.Now()
	return nil
}

// Delete marks the instance as deleted (soft delete)
func (i *Instance) Delete() {
	now := time.Now()
	i.deletedAt = &now
	i.updatedAt = now
}
