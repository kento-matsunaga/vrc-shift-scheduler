package event

import (
	"testing"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
)

func TestNewEvent_Success(t *testing.T) {
	tenantID := common.NewTenantID()
	eventName := "週末VRChat集会"
	eventType := EventTypeNormal
	description := "毎週末に開催するVRChat集会イベント"

	event, err := NewEvent(tenantID, eventName, eventType, description)

	if err != nil {
		t.Fatalf("NewEvent() should succeed, but got error: %v", err)
	}

	if event == nil {
		t.Fatal("NewEvent() returned nil")
	}

	// 基本フィールドの検証
	if event.TenantID() != tenantID {
		t.Errorf("TenantID: expected %s, got %s", tenantID, event.TenantID())
	}

	if event.EventName() != eventName {
		t.Errorf("EventName: expected %s, got %s", eventName, event.EventName())
	}

	if event.EventType() != eventType {
		t.Errorf("EventType: expected %s, got %s", eventType, event.EventType())
	}

	if event.Description() != description {
		t.Errorf("Description: expected %s, got %s", description, event.Description())
	}

	// デフォルト値の検証
	if !event.IsActive() {
		t.Error("IsActive should be true by default")
	}

	if event.IsDeleted() {
		t.Error("IsDeleted should be false by default")
	}

	// ID が生成されているか
	if event.EventID() == "" {
		t.Error("EventID should not be empty")
	}

	// タイムスタンプが設定されているか
	if event.CreatedAt().IsZero() {
		t.Error("CreatedAt should not be zero")
	}

	if event.UpdatedAt().IsZero() {
		t.Error("UpdatedAt should not be zero")
	}
}

func TestNewEvent_ErrorWhenEventNameEmpty(t *testing.T) {
	tenantID := common.NewTenantID()
	eventName := "" // 空文字
	eventType := EventTypeNormal
	description := "説明"

	event, err := NewEvent(tenantID, eventName, eventType, description)

	if err == nil {
		t.Fatal("NewEvent() should return error when event_name is empty")
	}

	if event != nil {
		t.Error("NewEvent() should return nil when validation fails")
	}
}

func TestNewEvent_ErrorWhenEventNameTooLong(t *testing.T) {
	tenantID := common.NewTenantID()
	// 256文字の文字列（上限255を超える）
	eventName := string(make([]byte, 256))
	eventType := EventTypeNormal
	description := "説明"

	event, err := NewEvent(tenantID, eventName, eventType, description)

	if err == nil {
		t.Fatal("NewEvent() should return error when event_name is too long")
	}

	if event != nil {
		t.Error("NewEvent() should return nil when validation fails")
	}
}

func TestNewEvent_ErrorWhenInvalidEventType(t *testing.T) {
	tenantID := common.NewTenantID()
	eventName := "テストイベント"
	eventType := EventType("invalid") // 不正な値
	description := "説明"

	event, err := NewEvent(tenantID, eventName, eventType, description)

	if err == nil {
		t.Fatal("NewEvent() should return error when event_type is invalid")
	}

	if event != nil {
		t.Error("NewEvent() should return nil when validation fails")
	}
}

func TestNewEvent_ErrorWhenTenantIDEmpty(t *testing.T) {
	tenantID := common.TenantID("") // 空のテナントID
	eventName := "テストイベント"
	eventType := EventTypeNormal
	description := "説明"

	event, err := NewEvent(tenantID, eventName, eventType, description)

	if err == nil {
		t.Fatal("NewEvent() should return error when tenant_id is empty")
	}

	if event != nil {
		t.Error("NewEvent() should return nil when validation fails")
	}
}

func TestEvent_UpdateEventName(t *testing.T) {
	tenantID := common.NewTenantID()
	event, _ := NewEvent(tenantID, "元の名前", EventTypeNormal, "説明")

	newName := "新しい名前"
	err := event.UpdateEventName(newName)

	if err != nil {
		t.Fatalf("UpdateEventName() should succeed, but got error: %v", err)
	}

	if event.EventName() != newName {
		t.Errorf("EventName: expected %s, got %s", newName, event.EventName())
	}
}

func TestEvent_UpdateEventName_ErrorWhenEmpty(t *testing.T) {
	tenantID := common.NewTenantID()
	event, _ := NewEvent(tenantID, "元の名前", EventTypeNormal, "説明")

	err := event.UpdateEventName("")

	if err == nil {
		t.Fatal("UpdateEventName() should return error when name is empty")
	}
}

func TestEvent_ActivateDeactivate(t *testing.T) {
	tenantID := common.NewTenantID()
	event, _ := NewEvent(tenantID, "テストイベント", EventTypeNormal, "説明")

	// 初期状態はアクティブ
	if !event.IsActive() {
		t.Error("Event should be active by default")
	}

	// 非アクティブ化
	event.Deactivate()
	if event.IsActive() {
		t.Error("Event should be inactive after Deactivate()")
	}

	// 再アクティブ化
	event.Activate()
	if !event.IsActive() {
		t.Error("Event should be active after Activate()")
	}
}

func TestEvent_Delete(t *testing.T) {
	tenantID := common.NewTenantID()
	event, _ := NewEvent(tenantID, "テストイベント", EventTypeNormal, "説明")

	// 初期状態は削除されていない
	if event.IsDeleted() {
		t.Error("Event should not be deleted by default")
	}

	if event.DeletedAt() != nil {
		t.Error("DeletedAt should be nil by default")
	}

	// 削除
	event.Delete()

	if !event.IsDeleted() {
		t.Error("Event should be deleted after Delete()")
	}

	if event.DeletedAt() == nil {
		t.Error("DeletedAt should not be nil after Delete()")
	}
}

func TestEventType_Validate(t *testing.T) {
	tests := []struct {
		name      string
		eventType EventType
		wantError bool
	}{
		{"Normal type is valid", EventTypeNormal, false},
		{"Special type is valid", EventTypeSpecial, false},
		{"Invalid type returns error", EventType("invalid"), true},
		{"Empty type returns error", EventType(""), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.eventType.Validate()
			if (err != nil) != tt.wantError {
				t.Errorf("Validate() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

