package attendance_test

import (
	"testing"
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/attendance"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
)

func TestNewTargetDate_Success(t *testing.T) {
	now := time.Now()
	collectionID := common.NewCollectionID()
	targetDate := time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC)
	startTime := "20:00"
	endTime := "23:00"

	td, err := attendance.NewTargetDate(now, collectionID, targetDate, &startTime, &endTime, 0)
	if err != nil {
		t.Fatalf("NewTargetDate() should succeed: %v", err)
	}

	if td.CollectionID() != collectionID {
		t.Errorf("CollectionID mismatch: got %v, want %v", td.CollectionID(), collectionID)
	}
	if td.TargetDateValue() != targetDate {
		t.Errorf("TargetDate mismatch: got %v, want %v", td.TargetDateValue(), targetDate)
	}
	if *td.StartTime() != startTime {
		t.Errorf("StartTime mismatch: got %v, want %v", *td.StartTime(), startTime)
	}
	if *td.EndTime() != endTime {
		t.Errorf("EndTime mismatch: got %v, want %v", *td.EndTime(), endTime)
	}
	if td.DisplayOrder() != 0 {
		t.Errorf("DisplayOrder mismatch: got %v, want 0", td.DisplayOrder())
	}
}

func TestNewTargetDate_WithoutTime(t *testing.T) {
	now := time.Now()
	collectionID := common.NewCollectionID()
	targetDate := time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC)

	td, err := attendance.NewTargetDate(now, collectionID, targetDate, nil, nil, 1)
	if err != nil {
		t.Fatalf("NewTargetDate() should succeed without time: %v", err)
	}
	if td.StartTime() != nil {
		t.Error("StartTime should be nil")
	}
	if td.EndTime() != nil {
		t.Error("EndTime should be nil")
	}
}

func TestNewTargetDate_ErrorWhenSameStartEndTime(t *testing.T) {
	now := time.Now()
	collectionID := common.NewCollectionID()
	targetDate := time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC)
	sameTime := "20:00"

	_, err := attendance.NewTargetDate(now, collectionID, targetDate, &sameTime, &sameTime, 0)
	if err == nil {
		t.Fatal("NewTargetDate() should fail when start_time == end_time")
	}
}

func TestNewTargetDate_ErrorWhenInvalidTimeFormat(t *testing.T) {
	now := time.Now()
	collectionID := common.NewCollectionID()
	targetDate := time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC)
	invalidTime := "25:00"

	_, err := attendance.NewTargetDate(now, collectionID, targetDate, &invalidTime, nil, 0)
	if err == nil {
		t.Fatal("NewTargetDate() should fail with invalid time format")
	}
}

func TestTargetDate_UpdateFields_Success(t *testing.T) {
	now := time.Now()
	collectionID := common.NewCollectionID()
	originalDate := time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC)
	startTime := "20:00"
	endTime := "23:00"

	td, err := attendance.NewTargetDate(now, collectionID, originalDate, &startTime, &endTime, 0)
	if err != nil {
		t.Fatalf("NewTargetDate() should succeed: %v", err)
	}

	originalID := td.TargetDateID()
	originalCreatedAt := td.CreatedAt()

	// フィールドを更新
	newDate := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	newStartTime := "18:00"
	newEndTime := "22:00"

	err = td.UpdateFields(newDate, &newStartTime, &newEndTime, 5)
	if err != nil {
		t.Fatalf("UpdateFields() should succeed: %v", err)
	}

	// ID と CreatedAt は保持される
	if td.TargetDateID() != originalID {
		t.Errorf("TargetDateID should be preserved: got %v, want %v", td.TargetDateID(), originalID)
	}
	if td.CreatedAt() != originalCreatedAt {
		t.Errorf("CreatedAt should be preserved: got %v, want %v", td.CreatedAt(), originalCreatedAt)
	}

	// フィールドが更新される
	if td.TargetDateValue() != newDate {
		t.Errorf("TargetDate should be updated: got %v, want %v", td.TargetDateValue(), newDate)
	}
	if *td.StartTime() != newStartTime {
		t.Errorf("StartTime should be updated: got %v, want %v", *td.StartTime(), newStartTime)
	}
	if *td.EndTime() != newEndTime {
		t.Errorf("EndTime should be updated: got %v, want %v", *td.EndTime(), newEndTime)
	}
	if td.DisplayOrder() != 5 {
		t.Errorf("DisplayOrder should be updated: got %v, want 5", td.DisplayOrder())
	}
}

func TestTargetDate_UpdateFields_RemoveTime(t *testing.T) {
	now := time.Now()
	collectionID := common.NewCollectionID()
	targetDate := time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC)
	startTime := "20:00"
	endTime := "23:00"

	td, err := attendance.NewTargetDate(now, collectionID, targetDate, &startTime, &endTime, 0)
	if err != nil {
		t.Fatalf("NewTargetDate() should succeed: %v", err)
	}

	// 時間を nil に更新（時間指定を解除）
	err = td.UpdateFields(targetDate, nil, nil, 0)
	if err != nil {
		t.Fatalf("UpdateFields() should succeed when removing time: %v", err)
	}

	if td.StartTime() != nil {
		t.Error("StartTime should be nil after update")
	}
	if td.EndTime() != nil {
		t.Error("EndTime should be nil after update")
	}
}

func TestTargetDate_UpdateFields_ErrorWhenSameStartEndTime(t *testing.T) {
	now := time.Now()
	collectionID := common.NewCollectionID()
	targetDate := time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC)
	startTime := "20:00"
	endTime := "23:00"

	td, err := attendance.NewTargetDate(now, collectionID, targetDate, &startTime, &endTime, 0)
	if err != nil {
		t.Fatalf("NewTargetDate() should succeed: %v", err)
	}

	sameTime := "20:00"
	err = td.UpdateFields(targetDate, &sameTime, &sameTime, 0)
	if err == nil {
		t.Fatal("UpdateFields() should fail when start_time == end_time")
	}

	// 元の値が保持されていないことを確認（バリデーション前に値が設定されるため）
	// ※ DDDの厳密な実装ではロールバックすべきだが、現在のvalidateパターンに合わせる
}

func TestTargetDate_UpdateFields_ErrorWhenInvalidTimeFormat(t *testing.T) {
	now := time.Now()
	collectionID := common.NewCollectionID()
	targetDate := time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC)

	td, err := attendance.NewTargetDate(now, collectionID, targetDate, nil, nil, 0)
	if err != nil {
		t.Fatalf("NewTargetDate() should succeed: %v", err)
	}

	invalidTime := "99:99"
	err = td.UpdateFields(targetDate, &invalidTime, nil, 0)
	if err == nil {
		t.Fatal("UpdateFields() should fail with invalid time format")
	}
}

func TestReconstructTargetDate_Success(t *testing.T) {
	targetDateID := common.NewTargetDateID()
	collectionID := common.NewCollectionID()
	targetDate := time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC)
	startTime := "20:00"
	endTime := "23:00"
	createdAt := time.Now()

	td, err := attendance.ReconstructTargetDate(targetDateID, collectionID, targetDate, &startTime, &endTime, 0, createdAt)
	if err != nil {
		t.Fatalf("ReconstructTargetDate() should succeed: %v", err)
	}

	if td.TargetDateID() != targetDateID {
		t.Errorf("TargetDateID mismatch: got %v, want %v", td.TargetDateID(), targetDateID)
	}
	if td.CreatedAt() != createdAt {
		t.Errorf("CreatedAt mismatch: got %v, want %v", td.CreatedAt(), createdAt)
	}
}
