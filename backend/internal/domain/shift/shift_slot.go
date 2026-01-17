package shift

import (
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/event"
)

// SlotID represents a shift slot identifier
type SlotID string

// NewSlotIDWithTime creates a new SlotID using the provided time.
func NewSlotIDWithTime(t time.Time) SlotID {
	return SlotID(common.NewULIDWithTime(t))
}

// NewSlotID creates a new SlotID using the current time.
// Deprecated: Use NewSlotIDWithTime for better testability.
func NewSlotID() SlotID {
	return SlotID(common.NewULID())
}

func (id SlotID) String() string {
	return string(id)
}

func (id SlotID) Validate() error {
	if id == "" {
		return common.NewValidationError("slot_id is required", nil)
	}
	return common.ValidateULID(string(id))
}

func ParseSlotID(s string) (SlotID, error) {
	if err := common.ValidateULID(s); err != nil {
		return "", err
	}
	return SlotID(s), nil
}

// ShiftSlot represents a shift slot entity (独立したエンティティ)
// EventBusinessDay に属するが、EventBusinessDay集約には含まれない
// Instance に紐づく（instanceID で参照、移行期間中は instanceName も保持）
//
// 表示順序のビジネスルール:
//   - priority が小さいほど優先的に表示される（昇順ソート）
//   - 同じ priority の場合は登録順（created_at 昇順）で表示
//   - 既存データとの互換性のため priority=0 を許容
//   - 新規作成時のデフォルト priority は 1（ユースケース層で設定）
type ShiftSlot struct {
	slotID        SlotID
	tenantID      common.TenantID
	businessDayID event.BusinessDayID
	instanceID    *InstanceID // インスタンスへの参照（FK）- nullable for migration
	slotName      string
	instanceName  string      // Deprecated: 移行完了後に削除予定。instanceID を使用してください。
	startTime     time.Time   // TIME型として扱う（HH:MM:SS）
	endTime       time.Time   // TIME型として扱う（HH:MM:SS）
	requiredCount int
	priority      int
	createdAt     time.Time
	updatedAt     time.Time
	deletedAt     *time.Time
}

// NewShiftSlot creates a new ShiftSlot entity
func NewShiftSlot(
	now time.Time,
	tenantID common.TenantID,
	businessDayID event.BusinessDayID,
	slotName string,
	instanceName string,
	startTime time.Time,
	endTime time.Time,
	requiredCount int,
	priority int,
) (*ShiftSlot, error) {
	slot := &ShiftSlot{
		slotID:        NewSlotIDWithTime(now),
		tenantID:      tenantID,
		businessDayID: businessDayID,
		slotName:      slotName,
		instanceName:  instanceName,
		startTime:     truncateToTime(startTime),
		endTime:       truncateToTime(endTime),
		requiredCount: requiredCount,
		priority:      priority,
		createdAt:     now,
		updatedAt:     now,
	}

	if err := slot.validate(); err != nil {
		return nil, err
	}

	return slot, nil
}

// ReconstructShiftSlot reconstructs a ShiftSlot from persistence
func ReconstructShiftSlot(
	slotID SlotID,
	tenantID common.TenantID,
	businessDayID event.BusinessDayID,
	instanceID *InstanceID,
	slotName string,
	instanceName string,
	startTime time.Time,
	endTime time.Time,
	requiredCount int,
	priority int,
	createdAt time.Time,
	updatedAt time.Time,
	deletedAt *time.Time,
) (*ShiftSlot, error) {
	slot := &ShiftSlot{
		slotID:        slotID,
		tenantID:      tenantID,
		businessDayID: businessDayID,
		instanceID:    instanceID,
		slotName:      slotName,
		instanceName:  instanceName,
		startTime:     truncateToTime(startTime),
		endTime:       truncateToTime(endTime),
		requiredCount: requiredCount,
		priority:      priority,
		createdAt:     createdAt,
		updatedAt:     updatedAt,
		deletedAt:     deletedAt,
	}

	if err := slot.validate(); err != nil {
		return nil, err
	}

	return slot, nil
}

func (s *ShiftSlot) validate() error {
	// TenantID の必須性チェック
	if err := s.tenantID.Validate(); err != nil {
		return common.NewValidationError("tenant_id is required", err)
	}

	// BusinessDayID の必須性チェック
	if err := s.businessDayID.Validate(); err != nil {
		return common.NewValidationError("business_day_id is required", err)
	}

	// SlotName の必須性チェック
	if s.slotName == "" {
		return common.NewValidationError("slot_name is required", nil)
	}

	if len(s.slotName) > 255 {
		return common.NewValidationError("slot_name must be less than 255 characters", nil)
	}

	// RequiredCount の範囲チェック（1以上）
	if s.requiredCount < 1 {
		return common.NewValidationError("required_count must be at least 1", nil)
	}

	// Priority の範囲チェック（0以上、小さいほど優先）
	// 注: 既存データとの互換性のため0を許容。新規作成時はデフォルト1が設定される
	if s.priority < 0 {
		return common.NewValidationError("priority must be at least 0", nil)
	}

	// 時刻の前後関係チェック（深夜営業対応）
	// start_time < end_time OR end_time < start_time のどちらかを満たす必要がある
	// （深夜営業の場合、end_time が start_time より前になる）

	return nil
}

// truncateToTime truncates a time to time only (HH:MM:SS)
func truncateToTime(t time.Time) time.Time {
	hour, min, sec := t.Clock()
	return time.Date(2000, 1, 1, hour, min, sec, 0, time.UTC)
}

// Getters

func (s *ShiftSlot) SlotID() SlotID {
	return s.slotID
}

func (s *ShiftSlot) TenantID() common.TenantID {
	return s.tenantID
}

func (s *ShiftSlot) BusinessDayID() event.BusinessDayID {
	return s.businessDayID
}

func (s *ShiftSlot) SlotName() string {
	return s.slotName
}

// InstanceID returns the instance ID if set
func (s *ShiftSlot) InstanceID() *InstanceID {
	return s.instanceID
}

// InstanceName returns the instance name (deprecated, use InstanceID instead)
// Deprecated: 移行完了後に削除予定。InstanceID() を使用してください。
func (s *ShiftSlot) InstanceName() string {
	return s.instanceName
}

func (s *ShiftSlot) StartTime() time.Time {
	return s.startTime
}

func (s *ShiftSlot) EndTime() time.Time {
	return s.endTime
}

func (s *ShiftSlot) RequiredCount() int {
	return s.requiredCount
}

func (s *ShiftSlot) Priority() int {
	return s.priority
}

func (s *ShiftSlot) CreatedAt() time.Time {
	return s.createdAt
}

func (s *ShiftSlot) UpdatedAt() time.Time {
	return s.updatedAt
}

func (s *ShiftSlot) DeletedAt() *time.Time {
	return s.deletedAt
}

func (s *ShiftSlot) IsDeleted() bool {
	return s.deletedAt != nil
}

// UpdateSlotName updates the slot name
func (s *ShiftSlot) UpdateSlotName(slotName string) error {
	if slotName == "" {
		return common.NewValidationError("slot_name is required", nil)
	}
	if len(slotName) > 255 {
		return common.NewValidationError("slot_name must be less than 255 characters", nil)
	}

	s.slotName = slotName
	s.updatedAt = time.Now()
	return nil
}

// UpdateRequiredCount updates the required count
func (s *ShiftSlot) UpdateRequiredCount(requiredCount int) error {
	if requiredCount < 1 {
		return common.NewValidationError("required_count must be at least 1", nil)
	}

	s.requiredCount = requiredCount
	s.updatedAt = time.Now()
	return nil
}

// UpdatePriority updates the priority (must be at least 0)
func (s *ShiftSlot) UpdatePriority(priority int) error {
	if priority < 0 {
		return common.NewValidationError("priority must be at least 0", nil)
	}

	s.priority = priority
	s.updatedAt = time.Now()
	return nil
}

// SetInstanceID sets the instance ID (for data migration)
func (s *ShiftSlot) SetInstanceID(instanceID InstanceID) {
	s.instanceID = &instanceID
	s.updatedAt = time.Now()
}

// Delete marks the slot as deleted (soft delete)
func (s *ShiftSlot) Delete() {
	now := time.Now()
	s.deletedAt = &now
	s.updatedAt = now
}

// IsOvernight returns true if the shift crosses midnight
func (s *ShiftSlot) IsOvernight() bool {
	return s.endTime.Before(s.startTime) || s.endTime.Equal(s.startTime)
}

// StartTimeString returns the start time as HH:MM string
func (s *ShiftSlot) StartTimeString() string {
	return s.startTime.Format("15:04")
}

// EndTimeString returns the end time as HH:MM string
func (s *ShiftSlot) EndTimeString() string {
	return s.endTime.Format("15:04")
}

