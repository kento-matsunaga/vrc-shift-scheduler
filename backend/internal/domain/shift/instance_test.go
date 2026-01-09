package shift

import (
	"testing"
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
)

func TestNewInstance_Success(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	eventID := common.NewEventID()
	maxMembers := 10

	instance, err := NewInstance(now, tenantID, eventID, "第一インスタンス", 1, &maxMembers)
	if err != nil {
		t.Fatalf("NewInstance() should succeed, got error: %v", err)
	}

	if instance.TenantID() != tenantID {
		t.Errorf("TenantID mismatch: got %v, want %v", instance.TenantID(), tenantID)
	}

	if instance.EventID() != eventID {
		t.Errorf("EventID mismatch: got %v, want %v", instance.EventID(), eventID)
	}

	if instance.Name() != "第一インスタンス" {
		t.Errorf("Name mismatch: got %v, want %v", instance.Name(), "第一インスタンス")
	}

	if instance.DisplayOrder() != 1 {
		t.Errorf("DisplayOrder mismatch: got %v, want %v", instance.DisplayOrder(), 1)
	}

	if instance.MaxMembers() == nil || *instance.MaxMembers() != 10 {
		t.Errorf("MaxMembers mismatch: got %v, want %v", instance.MaxMembers(), &maxMembers)
	}

	if instance.IsDeleted() {
		t.Error("Instance should not be deleted")
	}
}

func TestNewInstance_WithNilMaxMembers(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	eventID := common.NewEventID()

	instance, err := NewInstance(now, tenantID, eventID, "第一インスタンス", 0, nil)
	if err != nil {
		t.Fatalf("NewInstance() should succeed with nil maxMembers, got error: %v", err)
	}

	if instance.MaxMembers() != nil {
		t.Errorf("MaxMembers should be nil, got %v", instance.MaxMembers())
	}
}

func TestNewInstance_EmptyName(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	eventID := common.NewEventID()

	_, err := NewInstance(now, tenantID, eventID, "", 0, nil)
	if err == nil {
		t.Fatal("NewInstance() should fail when name is empty")
	}
}

func TestNewInstance_InvalidMaxMembers(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	eventID := common.NewEventID()
	invalidMaxMembers := 0

	_, err := NewInstance(now, tenantID, eventID, "第一インスタンス", 0, &invalidMaxMembers)
	if err == nil {
		t.Fatal("NewInstance() should fail when maxMembers is 0")
	}
}

func TestInstance_UpdateName(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	eventID := common.NewEventID()

	instance, _ := NewInstance(now, tenantID, eventID, "第一インスタンス", 0, nil)

	err := instance.UpdateName("新しい名前")
	if err != nil {
		t.Fatalf("UpdateName() should succeed, got error: %v", err)
	}

	if instance.Name() != "新しい名前" {
		t.Errorf("Name mismatch: got %v, want %v", instance.Name(), "新しい名前")
	}
}

func TestInstance_UpdateName_Empty(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	eventID := common.NewEventID()

	instance, _ := NewInstance(now, tenantID, eventID, "第一インスタンス", 0, nil)

	err := instance.UpdateName("")
	if err == nil {
		t.Fatal("UpdateName() should fail when name is empty")
	}
}

func TestInstance_UpdateDisplayOrder(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	eventID := common.NewEventID()

	instance, _ := NewInstance(now, tenantID, eventID, "第一インスタンス", 0, nil)

	instance.UpdateDisplayOrder(5)

	if instance.DisplayOrder() != 5 {
		t.Errorf("DisplayOrder mismatch: got %v, want %v", instance.DisplayOrder(), 5)
	}
}

func TestInstance_UpdateMaxMembers(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	eventID := common.NewEventID()

	instance, _ := NewInstance(now, tenantID, eventID, "第一インスタンス", 0, nil)

	newMaxMembers := 20
	err := instance.UpdateMaxMembers(&newMaxMembers)
	if err != nil {
		t.Fatalf("UpdateMaxMembers() should succeed, got error: %v", err)
	}

	if instance.MaxMembers() == nil || *instance.MaxMembers() != 20 {
		t.Errorf("MaxMembers mismatch: got %v, want %v", instance.MaxMembers(), &newMaxMembers)
	}
}

func TestInstance_UpdateMaxMembers_Invalid(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	eventID := common.NewEventID()

	instance, _ := NewInstance(now, tenantID, eventID, "第一インスタンス", 0, nil)

	invalidMaxMembers := 0
	err := instance.UpdateMaxMembers(&invalidMaxMembers)
	if err == nil {
		t.Fatal("UpdateMaxMembers() should fail when maxMembers is 0")
	}
}

func TestInstance_Delete(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	eventID := common.NewEventID()

	instance, _ := NewInstance(now, tenantID, eventID, "第一インスタンス", 0, nil)

	instance.Delete()

	if !instance.IsDeleted() {
		t.Error("Instance should be deleted")
	}

	if instance.DeletedAt() == nil {
		t.Error("DeletedAt should not be nil")
	}
}

func TestInstanceID_Validate(t *testing.T) {
	validID := NewInstanceID()
	if err := validID.Validate(); err != nil {
		t.Errorf("Valid InstanceID should pass validation: %v", err)
	}

	invalidID := InstanceID("")
	if err := invalidID.Validate(); err == nil {
		t.Error("Empty InstanceID should fail validation")
	}
}

func TestParseInstanceID(t *testing.T) {
	validID := NewInstanceID()

	parsed, err := ParseInstanceID(validID.String())
	if err != nil {
		t.Fatalf("ParseInstanceID() should succeed for valid ID: %v", err)
	}

	if parsed != validID {
		t.Errorf("Parsed ID mismatch: got %v, want %v", parsed, validID)
	}

	_, err = ParseInstanceID("invalid")
	if err == nil {
		t.Error("ParseInstanceID() should fail for invalid ID")
	}
}

func TestReconstructInstance(t *testing.T) {
	instanceID := NewInstanceID()
	tenantID := common.NewTenantID()
	eventID := common.NewEventID()
	maxMembers := 10
	now := time.Now()

	instance, err := ReconstructInstance(
		instanceID,
		tenantID,
		eventID,
		"第一インスタンス",
		1,
		&maxMembers,
		now,
		now,
		nil,
	)
	if err != nil {
		t.Fatalf("ReconstructInstance() should succeed, got error: %v", err)
	}

	if instance.InstanceID() != instanceID {
		t.Errorf("InstanceID mismatch: got %v, want %v", instance.InstanceID(), instanceID)
	}
}
