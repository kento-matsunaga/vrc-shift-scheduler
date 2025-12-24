package common

import (
	"testing"
	"time"
)

// =====================================================
// NewULID Tests
// =====================================================

func TestNewULID_Success(t *testing.T) {
	ulid := NewULID()

	if len(ulid) != 26 {
		t.Errorf("NewULID() length = %d, want 26", len(ulid))
	}

	if err := ValidateULID(ulid); err != nil {
		t.Errorf("NewULID() generated invalid ULID: %v", err)
	}
}

func TestNewULIDWithTime_Success(t *testing.T) {
	fixedTime := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	ulid := NewULIDWithTime(fixedTime)

	if len(ulid) != 26 {
		t.Errorf("NewULIDWithTime() length = %d, want 26", len(ulid))
	}

	if err := ValidateULID(ulid); err != nil {
		t.Errorf("NewULIDWithTime() generated invalid ULID: %v", err)
	}
}

func TestNewULID_Uniqueness(t *testing.T) {
	ulids := make(map[string]bool)
	iterations := 1000

	for i := 0; i < iterations; i++ {
		ulid := NewULID()
		if ulids[ulid] {
			t.Errorf("NewULID() generated duplicate ULID: %s", ulid)
		}
		ulids[ulid] = true
	}
}

// =====================================================
// ValidateULID Tests
// =====================================================

func TestValidateULID(t *testing.T) {
	tests := []struct {
		name    string
		ulid    string
		wantErr bool
	}{
		{
			name:    "valid ULID",
			ulid:    "01ARZ3NDEKTSV4RRFFQ69G5FAV",
			wantErr: false,
		},
		{
			name:    "valid ULID lowercase",
			ulid:    "01arz3ndektsv4rrffq69g5fav",
			wantErr: false,
		},
		{
			name:    "empty ULID",
			ulid:    "",
			wantErr: true,
		},
		{
			name:    "too short",
			ulid:    "01ARZ3NDEK",
			wantErr: true,
		},
		{
			name:    "too long",
			ulid:    "01ARZ3NDEKTSV4RRFFQ69G5FAVX",
			wantErr: true,
		},
		{
			name:    "UUID format (not ULID)",
			ulid:    "550e8400-e29b-41d4-a716-446655440000",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateULID(tt.ulid)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateULID(%q) error = %v, wantErr %v", tt.ulid, err, tt.wantErr)
			}
		})
	}
}

// =====================================================
// TenantID Tests
// =====================================================

func TestNewTenantID_Success(t *testing.T) {
	id := NewTenantID()

	if id == "" {
		t.Error("NewTenantID() should not return empty string")
	}

	if err := id.Validate(); err != nil {
		t.Errorf("NewTenantID() generated invalid ID: %v", err)
	}
}

func TestNewTenantIDWithTime_Success(t *testing.T) {
	fixedTime := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	id := NewTenantIDWithTime(fixedTime)

	if id == "" {
		t.Error("NewTenantIDWithTime() should not return empty string")
	}

	if err := id.Validate(); err != nil {
		t.Errorf("NewTenantIDWithTime() generated invalid ID: %v", err)
	}
}

func TestTenantID_Validate(t *testing.T) {
	tests := []struct {
		name    string
		id      TenantID
		wantErr bool
	}{
		{
			name:    "valid ID",
			id:      TenantID("01ARZ3NDEKTSV4RRFFQ69G5FAV"),
			wantErr: false,
		},
		{
			name:    "empty ID",
			id:      TenantID(""),
			wantErr: true,
		},
		{
			name:    "invalid format",
			id:      TenantID("invalid"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.id.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("TenantID.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestParseTenantID(t *testing.T) {
	t.Run("valid ID", func(t *testing.T) {
		validID := "01ARZ3NDEKTSV4RRFFQ69G5FAV"
		id, err := ParseTenantID(validID)
		if err != nil {
			t.Errorf("ParseTenantID(%q) unexpected error: %v", validID, err)
		}
		if id.String() != validID {
			t.Errorf("ParseTenantID(%q) = %q, want %q", validID, id.String(), validID)
		}
	})

	t.Run("invalid ID", func(t *testing.T) {
		invalidID := "invalid"
		_, err := ParseTenantID(invalidID)
		if err == nil {
			t.Errorf("ParseTenantID(%q) expected error, got nil", invalidID)
		}
	})
}

// =====================================================
// EventID Tests
// =====================================================

func TestNewEventID_Success(t *testing.T) {
	id := NewEventID()

	if id == "" {
		t.Error("NewEventID() should not return empty string")
	}

	if err := id.Validate(); err != nil {
		t.Errorf("NewEventID() generated invalid ID: %v", err)
	}
}

func TestEventID_Validate(t *testing.T) {
	tests := []struct {
		name    string
		id      EventID
		wantErr bool
	}{
		{
			name:    "valid ID",
			id:      EventID("01ARZ3NDEKTSV4RRFFQ69G5FAV"),
			wantErr: false,
		},
		{
			name:    "empty ID",
			id:      EventID(""),
			wantErr: true,
		},
		{
			name:    "invalid format",
			id:      EventID("invalid"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.id.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("EventID.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// =====================================================
// MemberID Tests
// =====================================================

func TestNewMemberID_Success(t *testing.T) {
	id := NewMemberID()

	if id == "" {
		t.Error("NewMemberID() should not return empty string")
	}

	if err := id.Validate(); err != nil {
		t.Errorf("NewMemberID() generated invalid ID: %v", err)
	}
}

func TestMemberID_Validate(t *testing.T) {
	tests := []struct {
		name    string
		id      MemberID
		wantErr bool
	}{
		{
			name:    "valid ID",
			id:      MemberID("01ARZ3NDEKTSV4RRFFQ69G5FAV"),
			wantErr: false,
		},
		{
			name:    "empty ID",
			id:      MemberID(""),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.id.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("MemberID.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// =====================================================
// AdminID Tests
// =====================================================

func TestNewAdminID_Success(t *testing.T) {
	id := NewAdminID()

	if id == "" {
		t.Error("NewAdminID() should not return empty string")
	}

	if err := id.Validate(); err != nil {
		t.Errorf("NewAdminID() generated invalid ID: %v", err)
	}
}

func TestAdminID_Validate(t *testing.T) {
	tests := []struct {
		name    string
		id      AdminID
		wantErr bool
	}{
		{
			name:    "valid ID",
			id:      AdminID("01ARZ3NDEKTSV4RRFFQ69G5FAV"),
			wantErr: false,
		},
		{
			name:    "empty ID",
			id:      AdminID(""),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.id.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("AdminID.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// =====================================================
// CollectionID Tests
// =====================================================

func TestNewCollectionID_Success(t *testing.T) {
	id := NewCollectionID()

	if id == "" {
		t.Error("NewCollectionID() should not return empty string")
	}

	if err := id.Validate(); err != nil {
		t.Errorf("NewCollectionID() generated invalid ID: %v", err)
	}
}

// =====================================================
// ScheduleID Tests
// =====================================================

func TestNewScheduleID_Success(t *testing.T) {
	id := NewScheduleID()

	if id == "" {
		t.Error("NewScheduleID() should not return empty string")
	}

	if err := id.Validate(); err != nil {
		t.Errorf("NewScheduleID() generated invalid ID: %v", err)
	}
}

// =====================================================
// RoleID Tests
// =====================================================

func TestNewRoleID_Success(t *testing.T) {
	id := NewRoleID()

	if id == "" {
		t.Error("NewRoleID() should not return empty string")
	}

	if err := id.Validate(); err != nil {
		t.Errorf("NewRoleID() generated invalid ID: %v", err)
	}
}

// =====================================================
// SlotID (via common) Tests
// =====================================================

func TestNewAssignmentID_Success(t *testing.T) {
	id := NewAssignmentID()

	if id == "" {
		t.Error("NewAssignmentID() should not return empty string")
	}

	if err := id.Validate(); err != nil {
		t.Errorf("NewAssignmentID() generated invalid ID: %v", err)
	}
}

// =====================================================
// MemberGroupID Tests
// =====================================================

func TestNewMemberGroupID_Success(t *testing.T) {
	id := NewMemberGroupID()

	if id == "" {
		t.Error("NewMemberGroupID() should not return empty string")
	}

	if err := id.Validate(); err != nil {
		t.Errorf("NewMemberGroupID() generated invalid ID: %v", err)
	}
}

// =====================================================
// RoleGroupID Tests
// =====================================================

func TestNewRoleGroupID_Success(t *testing.T) {
	id := NewRoleGroupID()

	if id == "" {
		t.Error("NewRoleGroupID() should not return empty string")
	}

	if err := id.Validate(); err != nil {
		t.Errorf("NewRoleGroupID() generated invalid ID: %v", err)
	}
}
