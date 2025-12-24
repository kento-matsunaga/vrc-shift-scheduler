package schedule_test

import (
	"testing"
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/schedule"
)

// =====================================================
// DateSchedule Entity Tests
// =====================================================

func createTestCandidates(t *testing.T, scheduleID common.ScheduleID, now time.Time) []*schedule.CandidateDate {
	t.Helper()

	candidate1, err := schedule.NewCandidateDate(
		now,
		scheduleID,
		now.Add(24*time.Hour),
		nil, nil,
		1,
	)
	if err != nil {
		t.Fatalf("Failed to create candidate 1: %v", err)
	}

	candidate2, err := schedule.NewCandidateDate(
		now,
		scheduleID,
		now.Add(48*time.Hour),
		nil, nil,
		2,
	)
	if err != nil {
		t.Fatalf("Failed to create candidate 2: %v", err)
	}

	return []*schedule.CandidateDate{candidate1, candidate2}
}

func TestNewDateSchedule_Success(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	scheduleID := common.NewScheduleID()
	deadline := now.Add(7 * 24 * time.Hour)
	candidates := createTestCandidates(t, scheduleID, now)

	ds, err := schedule.NewDateSchedule(
		now,
		scheduleID,
		tenantID,
		"忘年会の日程調整",
		"2024年忘年会の日程を調整します",
		nil,
		candidates,
		&deadline,
	)

	if err != nil {
		t.Fatalf("NewDateSchedule() should succeed, got error: %v", err)
	}

	if ds.ScheduleID() != scheduleID {
		t.Errorf("ScheduleID mismatch: got %v, want %v", ds.ScheduleID(), scheduleID)
	}

	if ds.TenantID() != tenantID {
		t.Errorf("TenantID mismatch: got %v, want %v", ds.TenantID(), tenantID)
	}

	if ds.Title() != "忘年会の日程調整" {
		t.Errorf("Title mismatch: got %v, want '忘年会の日程調整'", ds.Title())
	}

	if ds.Status() != schedule.StatusOpen {
		t.Errorf("Initial status should be open: got %v", ds.Status())
	}

	if ds.PublicToken() == "" {
		t.Error("PublicToken should be set")
	}

	if len(ds.Candidates()) != 2 {
		t.Errorf("Candidates count mismatch: got %v, want 2", len(ds.Candidates()))
	}

	if ds.DecidedCandidateID() != nil {
		t.Error("DecidedCandidateID should be nil initially")
	}
}

func TestNewDateSchedule_ErrorWhenTitleEmpty(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	scheduleID := common.NewScheduleID()
	candidates := createTestCandidates(t, scheduleID, now)

	_, err := schedule.NewDateSchedule(
		now,
		scheduleID,
		tenantID,
		"", // Empty title
		"Description",
		nil,
		candidates,
		nil,
	)

	if err == nil {
		t.Fatal("NewDateSchedule() should fail when title is empty")
	}
}

func TestNewDateSchedule_ErrorWhenNoCandidates(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	scheduleID := common.NewScheduleID()

	_, err := schedule.NewDateSchedule(
		now,
		scheduleID,
		tenantID,
		"Test Schedule",
		"Description",
		nil,
		[]*schedule.CandidateDate{}, // No candidates
		nil,
	)

	if err == nil {
		t.Fatal("NewDateSchedule() should fail when no candidates")
	}
}

func TestDateSchedule_CanRespond_Success(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	scheduleID := common.NewScheduleID()
	deadline := now.Add(7 * 24 * time.Hour)
	candidates := createTestCandidates(t, scheduleID, now)

	ds, _ := schedule.NewDateSchedule(now, scheduleID, tenantID, "Test", "Desc", nil, candidates, &deadline)

	err := ds.CanRespond(now.Add(1 * time.Hour))
	if err != nil {
		t.Errorf("CanRespond() should succeed: %v", err)
	}
}

func TestDateSchedule_CanRespond_ErrorWhenClosed(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	scheduleID := common.NewScheduleID()
	candidates := createTestCandidates(t, scheduleID, now)

	ds, _ := schedule.NewDateSchedule(now, scheduleID, tenantID, "Test", "Desc", nil, candidates, nil)
	_ = ds.Close(now)

	err := ds.CanRespond(now)
	if err == nil {
		t.Error("CanRespond() should fail when schedule is closed")
	}
}

func TestDateSchedule_CanRespond_ErrorWhenDeadlinePassed(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	scheduleID := common.NewScheduleID()
	deadline := now.Add(1 * time.Hour) // 1 hour deadline
	candidates := createTestCandidates(t, scheduleID, now)

	ds, _ := schedule.NewDateSchedule(now, scheduleID, tenantID, "Test", "Desc", nil, candidates, &deadline)

	err := ds.CanRespond(now.Add(2 * time.Hour)) // After deadline
	if err == nil {
		t.Error("CanRespond() should fail when deadline has passed")
	}
}

func TestDateSchedule_Decide_Success(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	scheduleID := common.NewScheduleID()
	candidates := createTestCandidates(t, scheduleID, now)

	ds, _ := schedule.NewDateSchedule(now, scheduleID, tenantID, "Test", "Desc", nil, candidates, nil)

	candidateID := candidates[0].CandidateID()
	err := ds.Decide(candidateID, now)

	if err != nil {
		t.Fatalf("Decide() should succeed: %v", err)
	}

	if ds.Status() != schedule.StatusDecided {
		t.Errorf("Status should be decided: got %v", ds.Status())
	}

	if ds.DecidedCandidateID() == nil {
		t.Fatal("DecidedCandidateID should be set")
	}

	if *ds.DecidedCandidateID() != candidateID {
		t.Errorf("DecidedCandidateID mismatch: got %v, want %v", *ds.DecidedCandidateID(), candidateID)
	}
}

func TestDateSchedule_Decide_ErrorWhenAlreadyDecided(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	scheduleID := common.NewScheduleID()
	candidates := createTestCandidates(t, scheduleID, now)

	ds, _ := schedule.NewDateSchedule(now, scheduleID, tenantID, "Test", "Desc", nil, candidates, nil)

	// First decide
	_ = ds.Decide(candidates[0].CandidateID(), now)

	// Second decide attempt
	err := ds.Decide(candidates[1].CandidateID(), now)
	if err == nil {
		t.Error("Decide() should fail when already decided")
	}
}

func TestDateSchedule_Decide_ErrorWhenCandidateNotFound(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	scheduleID := common.NewScheduleID()
	candidates := createTestCandidates(t, scheduleID, now)

	ds, _ := schedule.NewDateSchedule(now, scheduleID, tenantID, "Test", "Desc", nil, candidates, nil)

	invalidCandidateID := common.NewCandidateID()
	err := ds.Decide(invalidCandidateID, now)

	if err == nil {
		t.Error("Decide() should fail when candidate not found")
	}
}

func TestDateSchedule_Close_Success(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	scheduleID := common.NewScheduleID()
	candidates := createTestCandidates(t, scheduleID, now)

	ds, _ := schedule.NewDateSchedule(now, scheduleID, tenantID, "Test", "Desc", nil, candidates, nil)

	err := ds.Close(now)

	if err != nil {
		t.Fatalf("Close() should succeed: %v", err)
	}

	if ds.Status() != schedule.StatusClosed {
		t.Errorf("Status should be closed: got %v", ds.Status())
	}
}

func TestDateSchedule_Close_ErrorWhenAlreadyClosed(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	scheduleID := common.NewScheduleID()
	candidates := createTestCandidates(t, scheduleID, now)

	ds, _ := schedule.NewDateSchedule(now, scheduleID, tenantID, "Test", "Desc", nil, candidates, nil)

	_ = ds.Close(now)
	err := ds.Close(now)

	if err == nil {
		t.Error("Close() should fail when already closed")
	}
}

func TestDateSchedule_Close_ErrorWhenDecided(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	scheduleID := common.NewScheduleID()
	candidates := createTestCandidates(t, scheduleID, now)

	ds, _ := schedule.NewDateSchedule(now, scheduleID, tenantID, "Test", "Desc", nil, candidates, nil)

	_ = ds.Decide(candidates[0].CandidateID(), now)
	err := ds.Close(now)

	if err == nil {
		t.Error("Close() should fail when already decided")
	}
}

// =====================================================
// Status Tests
// =====================================================

func TestStatus_Validate(t *testing.T) {
	tests := []struct {
		status  string
		wantErr bool
	}{
		{"open", false},
		{"closed", false},
		{"decided", false},
		{"invalid", true},
		{"", true},
		{"OPEN", true}, // Case sensitive
	}

	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			_, err := schedule.NewStatus(tt.status)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewStatus(%q) error = %v, wantErr %v", tt.status, err, tt.wantErr)
			}
		})
	}
}

func TestStatus_Methods(t *testing.T) {
	open, _ := schedule.NewStatus("open")
	if !open.IsOpen() {
		t.Error("IsOpen() should return true for 'open' status")
	}

	closed, _ := schedule.NewStatus("closed")
	if !closed.IsClosed() {
		t.Error("IsClosed() should return true for 'closed' status")
	}

	decided, _ := schedule.NewStatus("decided")
	if !decided.IsDecided() {
		t.Error("IsDecided() should return true for 'decided' status")
	}
}

// =====================================================
// Availability Tests
// =====================================================

func TestAvailability_Validate(t *testing.T) {
	tests := []struct {
		availability string
		wantErr      bool
	}{
		{"available", false},
		{"unavailable", false},
		{"maybe", false},
		{"invalid", true},
		{"", true},
	}

	for _, tt := range tests {
		t.Run(tt.availability, func(t *testing.T) {
			_, err := schedule.NewAvailability(tt.availability)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewAvailability(%q) error = %v, wantErr %v", tt.availability, err, tt.wantErr)
			}
		})
	}
}

// =====================================================
// CandidateDate Tests
// =====================================================

func TestNewCandidateDate_Success(t *testing.T) {
	now := time.Now()
	scheduleID := common.NewScheduleID()
	candidateDate := now.Add(24 * time.Hour)

	candidate, err := schedule.NewCandidateDate(
		now,
		scheduleID,
		candidateDate,
		nil, nil,
		1,
	)

	if err != nil {
		t.Fatalf("NewCandidateDate() should succeed: %v", err)
	}

	if candidate.CandidateID() == "" {
		t.Error("CandidateID should be set")
	}

	if candidate.ScheduleID() != scheduleID {
		t.Errorf("ScheduleID mismatch: got %v, want %v", candidate.ScheduleID(), scheduleID)
	}

	if candidate.DisplayOrder() != 1 {
		t.Errorf("DisplayOrder mismatch: got %v, want 1", candidate.DisplayOrder())
	}
}

func TestNewCandidateDate_WithTime(t *testing.T) {
	now := time.Now()
	scheduleID := common.NewScheduleID()
	candidateDate := now.Add(24 * time.Hour)
	startTime := time.Date(2024, 1, 1, 14, 0, 0, 0, time.UTC)
	endTime := time.Date(2024, 1, 1, 18, 0, 0, 0, time.UTC)

	candidate, err := schedule.NewCandidateDate(
		now,
		scheduleID,
		candidateDate,
		&startTime,
		&endTime,
		1,
	)

	if err != nil {
		t.Fatalf("NewCandidateDate() should succeed: %v", err)
	}

	if candidate.StartTime() == nil {
		t.Error("StartTime should be set")
	}

	if candidate.EndTime() == nil {
		t.Error("EndTime should be set")
	}
}

func TestReconstructCandidateDate_Success(t *testing.T) {
	now := time.Now()
	candidateID := common.NewCandidateID()
	scheduleID := common.NewScheduleID()
	candidateDate := now.Add(24 * time.Hour)

	candidate, err := schedule.ReconstructCandidateDate(
		candidateID,
		scheduleID,
		candidateDate,
		nil, nil,
		1,
		now,
	)

	if err != nil {
		t.Fatalf("ReconstructCandidateDate() should succeed: %v", err)
	}

	if candidate.CandidateID() != candidateID {
		t.Errorf("CandidateID mismatch: got %v, want %v", candidate.CandidateID(), candidateID)
	}
}
