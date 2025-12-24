package attendance_test

import (
	"testing"
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/attendance"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
)

// =====================================================
// AttendanceCollection Entity Tests
// =====================================================

func TestNewAttendanceCollection_Success(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	deadline := now.Add(7 * 24 * time.Hour)

	collection, err := attendance.NewAttendanceCollection(
		now,
		tenantID,
		"12月イベント出欠確認",
		"12月のイベントへの参加可否を回答してください",
		attendance.TargetTypeEvent,
		"event-123",
		&deadline,
	)

	if err != nil {
		t.Fatalf("NewAttendanceCollection() should succeed, got error: %v", err)
	}

	if collection.CollectionID() == "" {
		t.Error("CollectionID should be set")
	}

	if collection.TenantID() != tenantID {
		t.Errorf("TenantID mismatch: got %v, want %v", collection.TenantID(), tenantID)
	}

	if collection.Title() != "12月イベント出欠確認" {
		t.Errorf("Title mismatch: got %v, want '12月イベント出欠確認'", collection.Title())
	}

	if collection.Description() != "12月のイベントへの参加可否を回答してください" {
		t.Errorf("Description mismatch: got %v", collection.Description())
	}

	if collection.TargetType() != attendance.TargetTypeEvent {
		t.Errorf("TargetType mismatch: got %v, want %v", collection.TargetType(), attendance.TargetTypeEvent)
	}

	if collection.TargetID() != "event-123" {
		t.Errorf("TargetID mismatch: got %v, want 'event-123'", collection.TargetID())
	}

	if collection.Status() != attendance.StatusOpen {
		t.Errorf("Initial status should be open: got %v", collection.Status())
	}

	if collection.PublicToken() == "" {
		t.Error("PublicToken should be set")
	}

	if collection.Deadline() == nil {
		t.Error("Deadline should be set")
	}
}

func TestNewAttendanceCollection_WithBusinessDayTarget(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()

	collection, err := attendance.NewAttendanceCollection(
		now,
		tenantID,
		"営業日出欠確認",
		"",
		attendance.TargetTypeBusinessDay,
		"bd-456",
		nil,
	)

	if err != nil {
		t.Fatalf("NewAttendanceCollection() should succeed, got error: %v", err)
	}

	if collection.TargetType() != attendance.TargetTypeBusinessDay {
		t.Errorf("TargetType mismatch: got %v, want %v", collection.TargetType(), attendance.TargetTypeBusinessDay)
	}
}

func TestNewAttendanceCollection_ErrorWhenTitleEmpty(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()

	_, err := attendance.NewAttendanceCollection(
		now,
		tenantID,
		"", // Empty title
		"Description",
		attendance.TargetTypeEvent,
		"event-123",
		nil,
	)

	if err == nil {
		t.Fatal("NewAttendanceCollection() should fail when title is empty")
	}
}

func TestNewAttendanceCollection_ErrorWhenTitleTooLong(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	longTitle := make([]byte, 256)
	for i := range longTitle {
		longTitle[i] = 'a'
	}

	_, err := attendance.NewAttendanceCollection(
		now,
		tenantID,
		string(longTitle),
		"Description",
		attendance.TargetTypeEvent,
		"event-123",
		nil,
	)

	if err == nil {
		t.Fatal("NewAttendanceCollection() should fail when title is too long")
	}
}

func TestAttendanceCollection_CanRespond_Success(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	deadline := now.Add(7 * 24 * time.Hour)

	collection, _ := attendance.NewAttendanceCollection(
		now, tenantID, "Test", "Desc", attendance.TargetTypeEvent, "", &deadline,
	)

	err := collection.CanRespond(now.Add(1 * time.Hour))
	if err != nil {
		t.Errorf("CanRespond() should succeed: %v", err)
	}
}

func TestAttendanceCollection_CanRespond_ErrorWhenClosed(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()

	collection, _ := attendance.NewAttendanceCollection(
		now, tenantID, "Test", "Desc", attendance.TargetTypeEvent, "", nil,
	)

	_ = collection.Close(now)

	err := collection.CanRespond(now)
	if err == nil {
		t.Error("CanRespond() should fail when collection is closed")
	}
}

func TestAttendanceCollection_CanRespond_ErrorWhenDeadlinePassed(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	deadline := now.Add(1 * time.Hour)

	collection, _ := attendance.NewAttendanceCollection(
		now, tenantID, "Test", "Desc", attendance.TargetTypeEvent, "", &deadline,
	)

	err := collection.CanRespond(now.Add(2 * time.Hour))
	if err == nil {
		t.Error("CanRespond() should fail when deadline has passed")
	}
}

func TestAttendanceCollection_Close_Success(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()

	collection, _ := attendance.NewAttendanceCollection(
		now, tenantID, "Test", "Desc", attendance.TargetTypeEvent, "", nil,
	)

	err := collection.Close(now)

	if err != nil {
		t.Fatalf("Close() should succeed: %v", err)
	}

	if collection.Status() != attendance.StatusClosed {
		t.Errorf("Status should be closed: got %v", collection.Status())
	}
}

func TestAttendanceCollection_Close_ErrorWhenAlreadyClosed(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()

	collection, _ := attendance.NewAttendanceCollection(
		now, tenantID, "Test", "Desc", attendance.TargetTypeEvent, "", nil,
	)

	_ = collection.Close(now)
	err := collection.Close(now)

	if err == nil {
		t.Error("Close() should fail when already closed")
	}
}

func TestAttendanceCollection_IsDeleted(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()

	collection, _ := attendance.NewAttendanceCollection(
		now, tenantID, "Test", "Desc", attendance.TargetTypeEvent, "", nil,
	)

	if collection.IsDeleted() {
		t.Error("New collection should not be deleted")
	}

	if collection.DeletedAt() != nil {
		t.Error("DeletedAt should be nil for new collection")
	}
}

// =====================================================
// Status Tests
// =====================================================

func TestAttendanceStatus_Validate(t *testing.T) {
	tests := []struct {
		status  string
		wantErr bool
	}{
		{"open", false},
		{"closed", false},
		{"invalid", true},
		{"", true},
		{"OPEN", true}, // Case sensitive
	}

	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			_, err := attendance.NewStatus(tt.status)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewStatus(%q) error = %v, wantErr %v", tt.status, err, tt.wantErr)
			}
		})
	}
}

func TestAttendanceStatus_Methods(t *testing.T) {
	open, _ := attendance.NewStatus("open")
	if !open.IsOpen() {
		t.Error("IsOpen() should return true for 'open' status")
	}
	if open.IsClosed() {
		t.Error("IsClosed() should return false for 'open' status")
	}

	closed, _ := attendance.NewStatus("closed")
	if !closed.IsClosed() {
		t.Error("IsClosed() should return true for 'closed' status")
	}
	if closed.IsOpen() {
		t.Error("IsOpen() should return false for 'closed' status")
	}
}

// =====================================================
// TargetType Tests
// =====================================================

func TestTargetType_Validate(t *testing.T) {
	tests := []struct {
		targetType string
		wantErr    bool
	}{
		{"event", false},
		{"business_day", false},
		{"invalid", true},
		{"", true},
	}

	for _, tt := range tests {
		t.Run(tt.targetType, func(t *testing.T) {
			_, err := attendance.NewTargetType(tt.targetType)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewTargetType(%q) error = %v, wantErr %v", tt.targetType, err, tt.wantErr)
			}
		})
	}
}

// =====================================================
// ResponseType Tests
// =====================================================

func TestResponseType_Validate(t *testing.T) {
	tests := []struct {
		responseType string
		wantErr      bool
	}{
		{"attending", false},
		{"absent", false},
		{"invalid", true},
		{"", true},
	}

	for _, tt := range tests {
		t.Run(tt.responseType, func(t *testing.T) {
			_, err := attendance.NewResponseType(tt.responseType)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewResponseType(%q) error = %v, wantErr %v", tt.responseType, err, tt.wantErr)
			}
		})
	}
}

// =====================================================
// ReconstructAttendanceCollection Tests
// =====================================================

func TestReconstructAttendanceCollection_Success(t *testing.T) {
	now := time.Now()
	collectionID := common.NewCollectionID()
	tenantID := common.NewTenantID()
	publicToken := common.NewPublicToken()
	deadline := now.Add(7 * 24 * time.Hour)

	collection, err := attendance.ReconstructAttendanceCollection(
		collectionID,
		tenantID,
		"Reconstructed Collection",
		"Description",
		attendance.TargetTypeEvent,
		"event-123",
		publicToken,
		attendance.StatusOpen,
		&deadline,
		now,
		now,
		nil,
	)

	if err != nil {
		t.Fatalf("ReconstructAttendanceCollection() should succeed: %v", err)
	}

	if collection.CollectionID() != collectionID {
		t.Errorf("CollectionID mismatch: got %v, want %v", collection.CollectionID(), collectionID)
	}

	if collection.Title() != "Reconstructed Collection" {
		t.Errorf("Title mismatch: got %v, want 'Reconstructed Collection'", collection.Title())
	}

	if collection.PublicToken() != publicToken {
		t.Errorf("PublicToken mismatch: got %v, want %v", collection.PublicToken(), publicToken)
	}
}

func TestReconstructAttendanceCollection_WithClosedStatus(t *testing.T) {
	now := time.Now()
	collectionID := common.NewCollectionID()
	tenantID := common.NewTenantID()
	publicToken := common.NewPublicToken()

	collection, err := attendance.ReconstructAttendanceCollection(
		collectionID,
		tenantID,
		"Closed Collection",
		"Description",
		attendance.TargetTypeEvent,
		"event-123",
		publicToken,
		attendance.StatusClosed,
		nil,
		now,
		now,
		nil,
	)

	if err != nil {
		t.Fatalf("ReconstructAttendanceCollection() should succeed: %v", err)
	}

	if collection.Status() != attendance.StatusClosed {
		t.Errorf("Status mismatch: got %v, want %v", collection.Status(), attendance.StatusClosed)
	}
}

func TestReconstructAttendanceCollection_WithDeletedAt(t *testing.T) {
	now := time.Now()
	collectionID := common.NewCollectionID()
	tenantID := common.NewTenantID()
	publicToken := common.NewPublicToken()
	deletedAt := now.Add(-1 * time.Hour)

	collection, err := attendance.ReconstructAttendanceCollection(
		collectionID,
		tenantID,
		"Deleted Collection",
		"Description",
		attendance.TargetTypeEvent,
		"event-123",
		publicToken,
		attendance.StatusOpen,
		nil,
		now,
		now,
		&deletedAt,
	)

	if err != nil {
		t.Fatalf("ReconstructAttendanceCollection() should succeed: %v", err)
	}

	if !collection.IsDeleted() {
		t.Error("Collection should be deleted")
	}

	if collection.DeletedAt() == nil {
		t.Error("DeletedAt should be set")
	}
}
