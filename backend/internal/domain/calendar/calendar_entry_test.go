package calendar_test

import (
	"testing"
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/calendar"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
)

func TestNewCalendarEntry_Success(t *testing.T) {
	now := time.Now()
	calendarID := common.NewCalendarID()
	tenantID := common.NewTenantID()
	date := time.Now().AddDate(0, 0, 7)

	entry, err := calendar.NewCalendarEntry(now, calendarID, tenantID, "テスト予定", date, nil, nil, "備考")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if entry.Title() != "テスト予定" {
		t.Errorf("expected title 'テスト予定', got '%s'", entry.Title())
	}
	if entry.Note() != "備考" {
		t.Errorf("expected note '備考', got '%s'", entry.Note())
	}
	if entry.CalendarID() != calendarID {
		t.Errorf("expected calendarID %s, got %s", calendarID, entry.CalendarID())
	}
	if entry.TenantID() != tenantID {
		t.Errorf("expected tenantID %s, got %s", tenantID, entry.TenantID())
	}
}

func TestNewCalendarEntry_SuccessWithTime(t *testing.T) {
	now := time.Now()
	calendarID := common.NewCalendarID()
	tenantID := common.NewTenantID()
	date := time.Now().AddDate(0, 0, 7)
	startTime := time.Date(2000, 1, 1, 20, 0, 0, 0, time.UTC)
	endTime := time.Date(2000, 1, 1, 22, 0, 0, 0, time.UTC)

	entry, err := calendar.NewCalendarEntry(now, calendarID, tenantID, "時間付き予定", date, &startTime, &endTime, "")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if entry.StartTime() == nil {
		t.Fatal("expected startTime to be set")
	}
	if entry.EndTime() == nil {
		t.Fatal("expected endTime to be set")
	}
}

func TestNewCalendarEntry_ErrorWhenTitleEmpty(t *testing.T) {
	now := time.Now()
	calendarID := common.NewCalendarID()
	tenantID := common.NewTenantID()
	date := time.Now()

	_, err := calendar.NewCalendarEntry(now, calendarID, tenantID, "", date, nil, nil, "")

	if err == nil {
		t.Fatal("expected error for empty title, got nil")
	}
}

func TestNewCalendarEntry_ErrorWhenTitleTooLong(t *testing.T) {
	now := time.Now()
	calendarID := common.NewCalendarID()
	tenantID := common.NewTenantID()
	date := time.Now()
	longTitle := string(make([]byte, 256)) // 256文字

	_, err := calendar.NewCalendarEntry(now, calendarID, tenantID, longTitle, date, nil, nil, "")

	if err == nil {
		t.Fatal("expected error for title too long, got nil")
	}
}

func TestCalendarEntry_Update_Success(t *testing.T) {
	now := time.Now()
	calendarID := common.NewCalendarID()
	tenantID := common.NewTenantID()
	date := time.Now()

	entry, _ := calendar.NewCalendarEntry(now, calendarID, tenantID, "初期タイトル", date, nil, nil, "")

	newDate := time.Now().AddDate(0, 0, 1)
	err := entry.Update(now, "更新タイトル", newDate, nil, nil, "更新備考")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if entry.Title() != "更新タイトル" {
		t.Errorf("expected title '更新タイトル', got '%s'", entry.Title())
	}
	if entry.Note() != "更新備考" {
		t.Errorf("expected note '更新備考', got '%s'", entry.Note())
	}
}

func TestCalendarEntry_Update_ErrorWhenTitleEmpty(t *testing.T) {
	now := time.Now()
	calendarID := common.NewCalendarID()
	tenantID := common.NewTenantID()
	date := time.Now()

	entry, _ := calendar.NewCalendarEntry(now, calendarID, tenantID, "初期タイトル", date, nil, nil, "")

	err := entry.Update(now, "", date, nil, nil, "")

	if err == nil {
		t.Fatal("expected error for empty title, got nil")
	}
	// タイトルは変更されていないはず
	if entry.Title() != "初期タイトル" {
		t.Errorf("expected title to remain '初期タイトル', got '%s'", entry.Title())
	}
}

func TestCalendarEntry_Update_ErrorWhenTitleTooLong(t *testing.T) {
	now := time.Now()
	calendarID := common.NewCalendarID()
	tenantID := common.NewTenantID()
	date := time.Now()

	entry, _ := calendar.NewCalendarEntry(now, calendarID, tenantID, "初期タイトル", date, nil, nil, "")

	longTitle := string(make([]byte, 256))
	err := entry.Update(now, longTitle, date, nil, nil, "")

	if err == nil {
		t.Fatal("expected error for title too long, got nil")
	}
}
