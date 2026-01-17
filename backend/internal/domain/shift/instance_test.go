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
	tests := []struct {
		name    string
		id      InstanceID
		wantErr bool
	}{
		{
			name:    "valid ULID format",
			id:      NewInstanceID(),
			wantErr: false,
		},
		{
			name:    "valid legacy format (26 uppercase alphanumeric)",
			id:      InstanceID("2827D68A90CD4DE28D4F8B717D"),
			wantErr: false,
		},
		{
			name:    "empty string",
			id:      InstanceID(""),
			wantErr: true,
		},
		{
			name:    "invalid - too short",
			id:      InstanceID("ABC123"),
			wantErr: true,
		},
		{
			name:    "invalid - too long",
			id:      InstanceID("2827D68A90CD4DE28D4F8B717DA"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.id.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("InstanceID.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestParseInstanceID(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "valid ULID format",
			input:   NewInstanceID().String(),
			wantErr: false,
		},
		{
			name:    "valid legacy format (26 uppercase alphanumeric)",
			input:   "2827D68A90CD4DE28D4F8B717D",
			wantErr: false,
		},
		{
			name:    "valid legacy format - another example",
			input:   "94E352BF22B94A6FAC8E27FB5D",
			wantErr: false,
		},
		{
			name:    "valid - lowercase (ULID spec allows case-insensitive)",
			input:   "2827d68a90cd4de28d4f8b717d",
			wantErr: false,
		},
		{
			name:    "invalid - empty string",
			input:   "",
			wantErr: true,
		},
		{
			name:    "invalid - too short",
			input:   "ABC123",
			wantErr: true,
		},
		{
			name:    "invalid - too long (27 chars)",
			input:   "2827D68A90CD4DE28D4F8B717DA",
			wantErr: true,
		},
		{
			name:    "invalid - too short (25 chars)",
			input:   "2827D68A90CD4DE28D4F8B71",
			wantErr: true,
		},
		{
			name:    "invalid - random string",
			input:   "invalid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed, err := ParseInstanceID(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseInstanceID() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err == nil && parsed.String() != tt.input {
				t.Errorf("ParseInstanceID() = %v, want %v", parsed, tt.input)
			}
		})
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
