package shift

import (
	"testing"
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/event"
)

// Helper function to create a test shift slot
func createTestSlot(t *testing.T, tenantID common.TenantID, slotName string, startTime, endTime time.Time, requiredCount int) *ShiftSlot {
	now := time.Now()
	businessDayID := event.NewBusinessDayID()
	slot, err := NewShiftSlot(now, tenantID, businessDayID, nil, slotName, "", startTime, endTime, requiredCount, 1)
	if err != nil {
		t.Fatalf("createTestSlot() failed: %v", err)
	}
	return slot
}

func TestNewShiftSlot_Success(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	businessDayID := event.NewBusinessDayID()
	slotName := "早番スタッフ"
	instanceName := "A班"
	startTime := time.Date(2000, 1, 1, 21, 30, 0, 0, time.UTC)
	endTime := time.Date(2000, 1, 1, 23, 0, 0, 0, time.UTC)
	requiredCount := 3
	priority := 10

	slot, err := NewShiftSlot(
		now,
		tenantID,
		businessDayID,
		nil, // instanceID is optional
		slotName,
		instanceName,
		startTime,
		endTime,
		requiredCount,
		priority,
	)

	if err != nil {
		t.Fatalf("NewShiftSlot() should succeed, but got error: %v", err)
	}

	if slot == nil {
		t.Fatal("NewShiftSlot() returned nil")
	}

	// 基本フィールドの検証
	if slot.TenantID() != tenantID {
		t.Errorf("TenantID: expected %s, got %s", tenantID, slot.TenantID())
	}

	if slot.BusinessDayID() != businessDayID {
		t.Errorf("BusinessDayID: expected %s, got %s", businessDayID, slot.BusinessDayID())
	}

	if slot.SlotName() != slotName {
		t.Errorf("SlotName: expected %s, got %s", slotName, slot.SlotName())
	}

	if slot.RequiredCount() != requiredCount {
		t.Errorf("RequiredCount: expected %d, got %d", requiredCount, slot.RequiredCount())
	}

	if slot.Priority() != priority {
		t.Errorf("Priority: expected %d, got %d", priority, slot.Priority())
	}

	// ID が生成されているか
	if slot.SlotID() == "" {
		t.Error("SlotID should not be empty")
	}

	// タイムスタンプが設定されているか
	if slot.CreatedAt().IsZero() {
		t.Error("CreatedAt should not be zero")
	}

	if slot.IsDeleted() {
		t.Error("IsDeleted should be false by default")
	}
}

func TestNewShiftSlot_ErrorWhenSlotNameEmpty(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	businessDayID := event.NewBusinessDayID()
	slotName := "" // 空文字
	startTime := time.Date(2000, 1, 1, 21, 30, 0, 0, time.UTC)
	endTime := time.Date(2000, 1, 1, 23, 0, 0, 0, time.UTC)

	slot, err := NewShiftSlot(
		now,
		tenantID,
		businessDayID,
		nil,
		slotName,
		"",
		startTime,
		endTime,
		1,
		1,
	)

	if err == nil {
		t.Fatal("NewShiftSlot() should return error when slot_name is empty")
	}

	if slot != nil {
		t.Error("NewShiftSlot() should return nil when validation fails")
	}
}

func TestNewShiftSlot_ErrorWhenRequiredCountZero(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	businessDayID := event.NewBusinessDayID()
	slotName := "スタッフ"
	startTime := time.Date(2000, 1, 1, 21, 30, 0, 0, time.UTC)
	endTime := time.Date(2000, 1, 1, 23, 0, 0, 0, time.UTC)
	requiredCount := 0 // 0は不正

	slot, err := NewShiftSlot(
		now,
		tenantID,
		businessDayID,
		nil,
		slotName,
		"",
		startTime,
		endTime,
		requiredCount,
		1,
	)

	if err == nil {
		t.Fatal("NewShiftSlot() should return error when required_count is 0")
	}

	if slot != nil {
		t.Error("NewShiftSlot() should return nil when validation fails")
	}
}

func TestNewShiftSlot_ErrorWhenTenantIDEmpty(t *testing.T) {
	now := time.Now()
	tenantID := common.TenantID("") // 空のテナントID
	businessDayID := event.NewBusinessDayID()
	slotName := "スタッフ"
	startTime := time.Date(2000, 1, 1, 21, 30, 0, 0, time.UTC)
	endTime := time.Date(2000, 1, 1, 23, 0, 0, 0, time.UTC)

	slot, err := NewShiftSlot(
		now,
		tenantID,
		businessDayID,
		nil,
		slotName,
		"",
		startTime,
		endTime,
		1,
		1,
	)

	if err == nil {
		t.Fatal("NewShiftSlot() should return error when tenant_id is empty")
	}

	if slot != nil {
		t.Error("NewShiftSlot() should return nil when validation fails")
	}
}

func TestNewShiftSlot_SuccessWhenPriorityZero(t *testing.T) {
	// priority=0 は既存データとの互換性のため許可される
	now := time.Now()
	tenantID := common.NewTenantID()
	businessDayID := event.NewBusinessDayID()
	slotName := "スタッフ"
	startTime := time.Date(2000, 1, 1, 21, 30, 0, 0, time.UTC)
	endTime := time.Date(2000, 1, 1, 23, 0, 0, 0, time.UTC)
	priority := 0 // 0は既存データとの互換性のため許可

	slot, err := NewShiftSlot(
		now,
		tenantID,
		businessDayID,
		nil,
		slotName,
		"",
		startTime,
		endTime,
		1,
		priority,
	)

	if err != nil {
		t.Fatalf("NewShiftSlot() should succeed when priority is 0, but got error: %v", err)
	}

	if slot == nil {
		t.Fatal("NewShiftSlot() should return slot when priority is 0")
	}

	if slot.Priority() != 0 {
		t.Errorf("Priority: expected 0, got %d", slot.Priority())
	}
}

func TestNewShiftSlot_ErrorWhenPriorityNegative(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	businessDayID := event.NewBusinessDayID()
	slotName := "スタッフ"
	startTime := time.Date(2000, 1, 1, 21, 30, 0, 0, time.UTC)
	endTime := time.Date(2000, 1, 1, 23, 0, 0, 0, time.UTC)
	priority := -1 // 負の値は不正

	slot, err := NewShiftSlot(
		now,
		tenantID,
		businessDayID,
		nil,
		slotName,
		"",
		startTime,
		endTime,
		1,
		priority,
	)

	if err == nil {
		t.Fatal("NewShiftSlot() should return error when priority is negative")
	}

	if slot != nil {
		t.Error("NewShiftSlot() should return nil when validation fails")
	}
}

func TestShiftSlot_UpdatePriority(t *testing.T) {
	tenantID := common.NewTenantID()
	startTime := time.Date(2000, 1, 1, 21, 30, 0, 0, time.UTC)
	endTime := time.Date(2000, 1, 1, 23, 0, 0, 0, time.UTC)

	slot := createTestSlot(t, tenantID, "スタッフ", startTime, endTime, 1)

	newPriority := 5
	err := slot.UpdatePriority(newPriority)

	if err != nil {
		t.Fatalf("UpdatePriority() should succeed, but got error: %v", err)
	}

	if slot.Priority() != newPriority {
		t.Errorf("Priority: expected %d, got %d", newPriority, slot.Priority())
	}
}

func TestShiftSlot_UpdatePriority_SuccessWhenZero(t *testing.T) {
	// priority=0 は既存データとの互換性のため許可される
	tenantID := common.NewTenantID()
	startTime := time.Date(2000, 1, 1, 21, 30, 0, 0, time.UTC)
	endTime := time.Date(2000, 1, 1, 23, 0, 0, 0, time.UTC)

	slot := createTestSlot(t, tenantID, "スタッフ", startTime, endTime, 1)

	err := slot.UpdatePriority(0)

	if err != nil {
		t.Fatalf("UpdatePriority() should succeed when priority is 0, but got error: %v", err)
	}

	if slot.Priority() != 0 {
		t.Errorf("Priority: expected 0, got %d", slot.Priority())
	}
}

func TestShiftSlot_UpdateSlotName(t *testing.T) {
	tenantID := common.NewTenantID()
	startTime := time.Date(2000, 1, 1, 21, 30, 0, 0, time.UTC)
	endTime := time.Date(2000, 1, 1, 23, 0, 0, 0, time.UTC)

	slot := createTestSlot(t, tenantID, "元の名前", startTime, endTime, 1)

	newName := "新しい名前"
	err := slot.UpdateSlotName(newName)

	if err != nil {
		t.Fatalf("UpdateSlotName() should succeed, but got error: %v", err)
	}

	if slot.SlotName() != newName {
		t.Errorf("SlotName: expected %s, got %s", newName, slot.SlotName())
	}
}

func TestShiftSlot_UpdateRequiredCount(t *testing.T) {
	tenantID := common.NewTenantID()
	startTime := time.Date(2000, 1, 1, 21, 30, 0, 0, time.UTC)
	endTime := time.Date(2000, 1, 1, 23, 0, 0, 0, time.UTC)

	slot := createTestSlot(t, tenantID, "スタッフ", startTime, endTime, 1)

	newCount := 5
	err := slot.UpdateRequiredCount(newCount)

	if err != nil {
		t.Fatalf("UpdateRequiredCount() should succeed, but got error: %v", err)
	}

	if slot.RequiredCount() != newCount {
		t.Errorf("RequiredCount: expected %d, got %d", newCount, slot.RequiredCount())
	}
}

func TestShiftSlot_UpdateRequiredCount_ErrorWhenZero(t *testing.T) {
	tenantID := common.NewTenantID()
	startTime := time.Date(2000, 1, 1, 21, 30, 0, 0, time.UTC)
	endTime := time.Date(2000, 1, 1, 23, 0, 0, 0, time.UTC)

	slot := createTestSlot(t, tenantID, "スタッフ", startTime, endTime, 1)

	err := slot.UpdateRequiredCount(0)

	if err == nil {
		t.Fatal("UpdateRequiredCount() should return error when count is 0")
	}
}

func TestShiftSlot_Delete(t *testing.T) {
	tenantID := common.NewTenantID()
	startTime := time.Date(2000, 1, 1, 21, 30, 0, 0, time.UTC)
	endTime := time.Date(2000, 1, 1, 23, 0, 0, 0, time.UTC)

	slot := createTestSlot(t, tenantID, "スタッフ", startTime, endTime, 1)

	// 初期状態は削除されていない
	if slot.IsDeleted() {
		t.Error("ShiftSlot should not be deleted by default")
	}

	// 削除
	slot.Delete()

	if !slot.IsDeleted() {
		t.Error("ShiftSlot should be deleted after Delete()")
	}

	if slot.DeletedAt() == nil {
		t.Error("DeletedAt should not be nil after Delete()")
	}
}

func TestShiftSlot_IsOvernight(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	businessDayID := event.NewBusinessDayID()

	tests := []struct {
		name      string
		startTime time.Time
		endTime   time.Time
		want      bool
	}{
		{
			name:      "通常シフト（深夜営業ではない）",
			startTime: time.Date(2000, 1, 1, 21, 30, 0, 0, time.UTC),
			endTime:   time.Date(2000, 1, 1, 23, 0, 0, 0, time.UTC),
			want:      false,
		},
		{
			name:      "深夜営業（日付をまたぐ）",
			startTime: time.Date(2000, 1, 1, 23, 30, 0, 0, time.UTC),
			endTime:   time.Date(2000, 1, 1, 2, 0, 0, 0, time.UTC),
			want:      true,
		},
		{
			name:      "同時刻（特殊ケース）",
			startTime: time.Date(2000, 1, 1, 21, 30, 0, 0, time.UTC),
			endTime:   time.Date(2000, 1, 1, 21, 30, 0, 0, time.UTC),
			want:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			slot, _ := NewShiftSlot(now, tenantID, businessDayID, nil, "スタッフ", "", tt.startTime, tt.endTime, 1, 1)
			got := slot.IsOvernight()
			if got != tt.want {
				t.Errorf("IsOvernight() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewShiftSlot_WithInstanceID(t *testing.T) {
	// instanceIDを指定した場合のテスト
	now := time.Now()
	tenantID := common.NewTenantID()
	businessDayID := event.NewBusinessDayID()
	instanceID := NewInstanceIDWithTime(now)
	slotName := "スタッフ"
	instanceName := "第一インスタンス"
	startTime := time.Date(2000, 1, 1, 21, 30, 0, 0, time.UTC)
	endTime := time.Date(2000, 1, 1, 23, 0, 0, 0, time.UTC)

	slot, err := NewShiftSlot(
		now,
		tenantID,
		businessDayID,
		&instanceID,
		slotName,
		instanceName,
		startTime,
		endTime,
		1,
		1,
	)

	if err != nil {
		t.Fatalf("NewShiftSlot() with instanceID should succeed, but got error: %v", err)
	}

	if slot.InstanceID() == nil {
		t.Error("InstanceID should not be nil when set")
	}

	if *slot.InstanceID() != instanceID {
		t.Errorf("InstanceID: expected %s, got %s", instanceID, *slot.InstanceID())
	}

	if slot.InstanceName() != instanceName {
		t.Errorf("InstanceName: expected %s, got %s", instanceName, slot.InstanceName())
	}
}

func TestShiftSlot_TimeString(t *testing.T) {
	tenantID := common.NewTenantID()
	startTime := time.Date(2000, 1, 1, 21, 30, 0, 0, time.UTC)
	endTime := time.Date(2000, 1, 1, 23, 45, 0, 0, time.UTC)

	slot := createTestSlot(t, tenantID, "スタッフ", startTime, endTime, 1)

	if slot.StartTimeString() != "21:30" {
		t.Errorf("StartTimeString() = %s, want 21:30", slot.StartTimeString())
	}

	if slot.EndTimeString() != "23:45" {
		t.Errorf("EndTimeString() = %s, want 23:45", slot.EndTimeString())
	}
}
