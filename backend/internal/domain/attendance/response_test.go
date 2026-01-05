package attendance_test

import (
	"testing"
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/attendance"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
)

// =====================================================
// AttendanceResponse Tests
// =====================================================

func TestNewAttendanceResponse_Success(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	collectionID := common.NewCollectionID()
	memberID := common.NewMemberID()
	targetDateID := common.NewTargetDateID()

	response, err := attendance.NewAttendanceResponse(
		now,
		collectionID,
		tenantID,
		memberID,
		targetDateID,
		attendance.ResponseTypeAttending,
		"参加します",
		nil,
		nil,
	)

	if err != nil {
		t.Fatalf("NewAttendanceResponse() should succeed, got error: %v", err)
	}

	if response.ResponseID() == "" {
		t.Error("ResponseID should be set")
	}

	if response.TenantID() != tenantID {
		t.Errorf("TenantID mismatch: got %v, want %v", response.TenantID(), tenantID)
	}

	if response.CollectionID() != collectionID {
		t.Errorf("CollectionID mismatch: got %v, want %v", response.CollectionID(), collectionID)
	}

	if response.MemberID() != memberID {
		t.Errorf("MemberID mismatch: got %v, want %v", response.MemberID(), memberID)
	}

	if response.TargetDateID() != targetDateID {
		t.Errorf("TargetDateID mismatch: got %v, want %v", response.TargetDateID(), targetDateID)
	}

	if response.Response() != attendance.ResponseTypeAttending {
		t.Errorf("Response mismatch: got %v, want %v", response.Response(), attendance.ResponseTypeAttending)
	}

	if response.Note() != "参加します" {
		t.Errorf("Note mismatch: got %v, want '参加します'", response.Note())
	}
}

func TestNewAttendanceResponse_Absent(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	collectionID := common.NewCollectionID()
	memberID := common.NewMemberID()
	targetDateID := common.NewTargetDateID()

	response, err := attendance.NewAttendanceResponse(
		now,
		collectionID,
		tenantID,
		memberID,
		targetDateID,
		attendance.ResponseTypeAbsent,
		"予定があります",
		nil,
		nil,
	)

	if err != nil {
		t.Fatalf("NewAttendanceResponse() should succeed, got error: %v", err)
	}

	if response.Response() != attendance.ResponseTypeAbsent {
		t.Errorf("Response mismatch: got %v, want %v", response.Response(), attendance.ResponseTypeAbsent)
	}
}

func TestNewAttendanceResponse_EmptyNote(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	collectionID := common.NewCollectionID()
	memberID := common.NewMemberID()
	targetDateID := common.NewTargetDateID()

	response, err := attendance.NewAttendanceResponse(
		now,
		collectionID,
		tenantID,
		memberID,
		targetDateID,
		attendance.ResponseTypeAttending,
		"", // Empty note is allowed
		nil,
		nil,
	)

	if err != nil {
		t.Fatalf("NewAttendanceResponse() should succeed with empty note, got error: %v", err)
	}

	if response.Note() != "" {
		t.Errorf("Note should be empty: got %v", response.Note())
	}
}

func TestNewAttendanceResponse_ErrorWhenInvalidTenantID(t *testing.T) {
	now := time.Now()
	collectionID := common.NewCollectionID()
	memberID := common.NewMemberID()
	targetDateID := common.NewTargetDateID()

	_, err := attendance.NewAttendanceResponse(
		now,
		collectionID,
		common.TenantID(""), // Invalid tenant ID
		memberID,
		targetDateID,
		attendance.ResponseTypeAttending,
		"",
		nil,
		nil,
	)

	if err == nil {
		t.Fatal("NewAttendanceResponse() should fail when tenant ID is invalid")
	}
}

func TestNewAttendanceResponse_ErrorWhenInvalidCollectionID(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	memberID := common.NewMemberID()
	targetDateID := common.NewTargetDateID()

	_, err := attendance.NewAttendanceResponse(
		now,
		common.CollectionID(""), // Invalid collection ID
		tenantID,
		memberID,
		targetDateID,
		attendance.ResponseTypeAttending,
		"",
		nil,
		nil,
	)

	if err == nil {
		t.Fatal("NewAttendanceResponse() should fail when collection ID is invalid")
	}
}

func TestNewAttendanceResponse_ErrorWhenInvalidMemberID(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	collectionID := common.NewCollectionID()
	targetDateID := common.NewTargetDateID()

	_, err := attendance.NewAttendanceResponse(
		now,
		collectionID,
		tenantID,
		common.MemberID(""), // Invalid member ID
		targetDateID,
		attendance.ResponseTypeAttending,
		"",
		nil,
		nil,
	)

	if err == nil {
		t.Fatal("NewAttendanceResponse() should fail when member ID is invalid")
	}
}

func TestNewAttendanceResponse_ErrorWhenInvalidTargetDateID(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	collectionID := common.NewCollectionID()
	memberID := common.NewMemberID()

	_, err := attendance.NewAttendanceResponse(
		now,
		collectionID,
		tenantID,
		memberID,
		common.TargetDateID(""), // Invalid target date ID
		attendance.ResponseTypeAttending,
		"",
		nil,
		nil,
	)

	if err == nil {
		t.Fatal("NewAttendanceResponse() should fail when target date ID is invalid")
	}
}

func TestNewAttendanceResponse_ErrorWhenInvalidResponseType(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	collectionID := common.NewCollectionID()
	memberID := common.NewMemberID()
	targetDateID := common.NewTargetDateID()

	_, err := attendance.NewAttendanceResponse(
		now,
		collectionID,
		tenantID,
		memberID,
		targetDateID,
		attendance.ResponseType("invalid"), // Invalid response type
		"",
		nil,
		nil,
	)

	if err == nil {
		t.Fatal("NewAttendanceResponse() should fail when response type is invalid")
	}
}

func TestReconstructAttendanceResponse_Success(t *testing.T) {
	now := time.Now()
	responseID := common.NewResponseID()
	tenantID := common.NewTenantID()
	collectionID := common.NewCollectionID()
	memberID := common.NewMemberID()
	targetDateID := common.NewTargetDateID()

	response, err := attendance.ReconstructAttendanceResponse(
		responseID,
		tenantID,
		collectionID,
		memberID,
		targetDateID,
		attendance.ResponseTypeAttending,
		"テストノート",
		nil,
		nil,
		now,
		now,
		now,
	)

	if err != nil {
		t.Fatalf("ReconstructAttendanceResponse() should succeed, got error: %v", err)
	}

	if response.ResponseID() != responseID {
		t.Errorf("ResponseID mismatch: got %v, want %v", response.ResponseID(), responseID)
	}

	if response.TenantID() != tenantID {
		t.Errorf("TenantID mismatch: got %v, want %v", response.TenantID(), tenantID)
	}

	if response.CollectionID() != collectionID {
		t.Errorf("CollectionID mismatch: got %v, want %v", response.CollectionID(), collectionID)
	}

	if response.MemberID() != memberID {
		t.Errorf("MemberID mismatch: got %v, want %v", response.MemberID(), memberID)
	}

	if response.TargetDateID() != targetDateID {
		t.Errorf("TargetDateID mismatch: got %v, want %v", response.TargetDateID(), targetDateID)
	}

	if response.Response() != attendance.ResponseTypeAttending {
		t.Errorf("Response mismatch: got %v, want %v", response.Response(), attendance.ResponseTypeAttending)
	}

	if response.Note() != "テストノート" {
		t.Errorf("Note mismatch: got %v, want 'テストノート'", response.Note())
	}
}

func TestReconstructAttendanceResponse_ErrorWhenInvalid(t *testing.T) {
	now := time.Now()
	responseID := common.NewResponseID()
	collectionID := common.NewCollectionID()
	memberID := common.NewMemberID()
	targetDateID := common.NewTargetDateID()

	_, err := attendance.ReconstructAttendanceResponse(
		responseID,
		common.TenantID(""), // Invalid tenant ID
		collectionID,
		memberID,
		targetDateID,
		attendance.ResponseTypeAttending,
		"",
		nil,
		nil,
		now,
		now,
		now,
	)

	if err == nil {
		t.Fatal("ReconstructAttendanceResponse() should fail when tenant ID is invalid")
	}
}

