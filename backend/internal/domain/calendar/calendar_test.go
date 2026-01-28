package calendar_test

import (
	"strings"
	"testing"
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/calendar"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
)

// =====================================================
// NewCalendar Tests
// =====================================================

func TestNewCalendar_Success(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	eventID := common.NewEventID()

	cal, err := calendar.NewCalendar(now, tenantID, "Test Calendar", "Test Description", []common.EventID{eventID})

	if err != nil {
		t.Fatalf("NewCalendar() should succeed, got error: %v", err)
	}

	if cal.CalendarID() == "" {
		t.Error("CalendarID should be set")
	}

	if cal.TenantID() != tenantID {
		t.Errorf("TenantID mismatch: got %v, want %v", cal.TenantID(), tenantID)
	}

	if cal.Title() != "Test Calendar" {
		t.Errorf("Title mismatch: got %v, want 'Test Calendar'", cal.Title())
	}

	if cal.Description() != "Test Description" {
		t.Errorf("Description mismatch: got %v, want 'Test Description'", cal.Description())
	}

	if cal.IsPublic() {
		t.Error("New calendar should not be public")
	}

	if cal.PublicToken() != nil {
		t.Error("New calendar should not have a public token")
	}

	if len(cal.EventIDs()) != 1 {
		t.Errorf("EventIDs length mismatch: got %v, want 1", len(cal.EventIDs()))
	}

	if cal.EventIDs()[0] != eventID {
		t.Errorf("EventID mismatch: got %v, want %v", cal.EventIDs()[0], eventID)
	}

	if cal.CreatedAt().IsZero() {
		t.Error("CreatedAt should be set")
	}

	if cal.UpdatedAt().IsZero() {
		t.Error("UpdatedAt should be set")
	}
}

func TestNewCalendar_SuccessWithEmptyDescription(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	eventID := common.NewEventID()

	cal, err := calendar.NewCalendar(now, tenantID, "Test Calendar", "", []common.EventID{eventID})

	if err != nil {
		t.Fatalf("NewCalendar() should succeed with empty description, got error: %v", err)
	}

	if cal.Description() != "" {
		t.Errorf("Description should be empty: got %v", cal.Description())
	}
}

func TestNewCalendar_SuccessWithMultipleEventIDs(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	eventIDs := []common.EventID{
		common.NewEventID(),
		common.NewEventID(),
		common.NewEventID(),
	}

	cal, err := calendar.NewCalendar(now, tenantID, "Multi-Event Calendar", "Description", eventIDs)

	if err != nil {
		t.Fatalf("NewCalendar() should succeed with multiple event IDs, got error: %v", err)
	}

	if len(cal.EventIDs()) != 3 {
		t.Errorf("EventIDs length mismatch: got %v, want 3", len(cal.EventIDs()))
	}
}

func TestNewCalendar_ErrorWhenTitleEmpty(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	eventID := common.NewEventID()

	_, err := calendar.NewCalendar(now, tenantID, "", "Description", []common.EventID{eventID})

	if err == nil {
		t.Fatal("NewCalendar() should fail when title is empty")
	}
}

func TestNewCalendar_ErrorWhenTitleTooLong(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	eventID := common.NewEventID()
	longTitle := strings.Repeat("a", 256)

	_, err := calendar.NewCalendar(now, tenantID, longTitle, "Description", []common.EventID{eventID})

	if err == nil {
		t.Fatal("NewCalendar() should fail when title exceeds 255 characters")
	}
}

func TestNewCalendar_SuccessWithMaxLengthTitle(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	eventID := common.NewEventID()
	maxTitle := strings.Repeat("a", 255)

	cal, err := calendar.NewCalendar(now, tenantID, maxTitle, "Description", []common.EventID{eventID})

	if err != nil {
		t.Fatalf("NewCalendar() should succeed with 255 character title, got error: %v", err)
	}

	if len(cal.Title()) != 255 {
		t.Errorf("Title length mismatch: got %v, want 255", len(cal.Title()))
	}
}

// =====================================================
// MakePublic Tests
// =====================================================

func TestCalendar_MakePublic(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	eventID := common.NewEventID()

	cal, _ := calendar.NewCalendar(now, tenantID, "Test Calendar", "Description", []common.EventID{eventID})

	if cal.IsPublic() {
		t.Error("New calendar should not be public initially")
	}

	updateTime := now.Add(1 * time.Hour)
	cal.MakePublic(updateTime)

	if !cal.IsPublic() {
		t.Error("Calendar should be public after MakePublic()")
	}

	if cal.PublicToken() == nil {
		t.Error("PublicToken should be set after MakePublic()")
	}

	if !cal.UpdatedAt().Equal(updateTime) {
		t.Errorf("UpdatedAt should be updated: got %v, want %v", cal.UpdatedAt(), updateTime)
	}
}

func TestCalendar_MakePublic_AlreadyPublic(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	eventID := common.NewEventID()

	cal, _ := calendar.NewCalendar(now, tenantID, "Test Calendar", "Description", []common.EventID{eventID})

	// First make public
	updateTime1 := now.Add(1 * time.Hour)
	cal.MakePublic(updateTime1)
	originalToken := cal.PublicToken()

	// Make public again
	updateTime2 := now.Add(2 * time.Hour)
	cal.MakePublic(updateTime2)

	if !cal.IsPublic() {
		t.Error("Calendar should remain public")
	}

	if cal.PublicToken() == nil {
		t.Error("PublicToken should still be set")
	}

	// Token should be the same (not regenerated)
	if cal.PublicToken().String() != originalToken.String() {
		t.Error("PublicToken should not change when already public")
	}

	if !cal.UpdatedAt().Equal(updateTime2) {
		t.Errorf("UpdatedAt should be updated: got %v, want %v", cal.UpdatedAt(), updateTime2)
	}
}

// =====================================================
// MakePrivate Tests
// =====================================================

func TestCalendar_MakePrivate(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	eventID := common.NewEventID()

	cal, _ := calendar.NewCalendar(now, tenantID, "Test Calendar", "Description", []common.EventID{eventID})

	// First make public
	updateTime1 := now.Add(1 * time.Hour)
	cal.MakePublic(updateTime1)
	originalToken := cal.PublicToken()

	if !cal.IsPublic() {
		t.Error("Calendar should be public after MakePublic()")
	}

	// Now make private
	updateTime2 := now.Add(2 * time.Hour)
	cal.MakePrivate(updateTime2)

	if cal.IsPublic() {
		t.Error("Calendar should be private after MakePrivate()")
	}

	// Token should be retained (not cleared)
	if cal.PublicToken() == nil {
		t.Error("PublicToken should be retained after MakePrivate()")
	}

	if cal.PublicToken().String() != originalToken.String() {
		t.Error("PublicToken should remain the same after MakePrivate()")
	}

	if !cal.UpdatedAt().Equal(updateTime2) {
		t.Errorf("UpdatedAt should be updated: got %v, want %v", cal.UpdatedAt(), updateTime2)
	}
}

func TestCalendar_MakePrivate_AlreadyPrivate(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	eventID := common.NewEventID()

	cal, _ := calendar.NewCalendar(now, tenantID, "Test Calendar", "Description", []common.EventID{eventID})

	// Calendar is already private by default
	updateTime := now.Add(1 * time.Hour)
	cal.MakePrivate(updateTime)

	if cal.IsPublic() {
		t.Error("Calendar should remain private")
	}

	if !cal.UpdatedAt().Equal(updateTime) {
		t.Errorf("UpdatedAt should be updated: got %v, want %v", cal.UpdatedAt(), updateTime)
	}
}

// =====================================================
// Update Tests
// =====================================================

func TestCalendar_Update_Success(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	eventID := common.NewEventID()

	cal, _ := calendar.NewCalendar(now, tenantID, "Original Title", "Original Description", []common.EventID{eventID})

	newEventIDs := []common.EventID{common.NewEventID(), common.NewEventID()}
	updateTime := now.Add(1 * time.Hour)

	err := cal.Update("New Title", "New Description", newEventIDs, updateTime)

	if err != nil {
		t.Fatalf("Update() should succeed, got error: %v", err)
	}

	if cal.Title() != "New Title" {
		t.Errorf("Title should be updated: got %v, want 'New Title'", cal.Title())
	}

	if cal.Description() != "New Description" {
		t.Errorf("Description should be updated: got %v, want 'New Description'", cal.Description())
	}

	if len(cal.EventIDs()) != 2 {
		t.Errorf("EventIDs should be updated: got %v, want 2", len(cal.EventIDs()))
	}

	if !cal.UpdatedAt().Equal(updateTime) {
		t.Errorf("UpdatedAt should be updated: got %v, want %v", cal.UpdatedAt(), updateTime)
	}
}

func TestCalendar_Update_ErrorWhenTitleEmpty(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	eventID := common.NewEventID()

	cal, _ := calendar.NewCalendar(now, tenantID, "Original Title", "Original Description", []common.EventID{eventID})

	updateTime := now.Add(1 * time.Hour)
	err := cal.Update("", "New Description", []common.EventID{eventID}, updateTime)

	if err == nil {
		t.Fatal("Update() should fail when title is empty")
	}

	// Original values should be preserved
	if cal.Title() != "Original Title" {
		t.Errorf("Title should remain unchanged: got %v, want 'Original Title'", cal.Title())
	}
}

func TestCalendar_Update_ErrorWhenTitleTooLong(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	eventID := common.NewEventID()

	cal, _ := calendar.NewCalendar(now, tenantID, "Original Title", "Original Description", []common.EventID{eventID})
	longTitle := strings.Repeat("a", 256)

	updateTime := now.Add(1 * time.Hour)
	err := cal.Update(longTitle, "New Description", []common.EventID{eventID}, updateTime)

	if err == nil {
		t.Fatal("Update() should fail when title exceeds 255 characters")
	}

	// Original values should be preserved
	if cal.Title() != "Original Title" {
		t.Errorf("Title should remain unchanged: got %v, want 'Original Title'", cal.Title())
	}
}

func TestCalendar_Update_SuccessWithEmptyDescription(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	eventID := common.NewEventID()

	cal, _ := calendar.NewCalendar(now, tenantID, "Original Title", "Original Description", []common.EventID{eventID})

	updateTime := now.Add(1 * time.Hour)
	err := cal.Update("New Title", "", []common.EventID{eventID}, updateTime)

	if err != nil {
		t.Fatalf("Update() should succeed with empty description, got error: %v", err)
	}

	if cal.Description() != "" {
		t.Errorf("Description should be empty: got %v", cal.Description())
	}

	if !cal.UpdatedAt().Equal(updateTime) {
		t.Errorf("UpdatedAt should be updated: got %v, want %v", cal.UpdatedAt(), updateTime)
	}
}

// =====================================================
// ReconstructCalendar Tests
// =====================================================

func TestReconstructCalendar_Success(t *testing.T) {
	calendarID := common.NewCalendarID()
	tenantID := common.NewTenantID()
	eventID := common.NewEventID()
	token := common.NewPublicToken()
	now := time.Now()

	cal, err := calendar.ReconstructCalendar(
		calendarID,
		tenantID,
		"Reconstructed Calendar",
		"Reconstructed Description",
		true,
		&token,
		[]common.EventID{eventID},
		now,
		now,
	)

	if err != nil {
		t.Fatalf("ReconstructCalendar() should succeed, got error: %v", err)
	}

	if cal.CalendarID() != calendarID {
		t.Errorf("CalendarID mismatch: got %v, want %v", cal.CalendarID(), calendarID)
	}

	if cal.Title() != "Reconstructed Calendar" {
		t.Errorf("Title mismatch: got %v, want 'Reconstructed Calendar'", cal.Title())
	}

	if !cal.IsPublic() {
		t.Error("Calendar should be public as reconstructed")
	}

	if cal.PublicToken() == nil {
		t.Error("PublicToken should be set as reconstructed")
	}
}

func TestReconstructCalendar_ErrorWhenTitleEmpty(t *testing.T) {
	calendarID := common.NewCalendarID()
	tenantID := common.NewTenantID()
	eventID := common.NewEventID()
	now := time.Now()

	_, err := calendar.ReconstructCalendar(
		calendarID,
		tenantID,
		"",
		"Description",
		false,
		nil,
		[]common.EventID{eventID},
		now,
		now,
	)

	if err == nil {
		t.Fatal("ReconstructCalendar() should fail when title is empty")
	}
}

// =====================================================
// Getter Tests
// =====================================================

func TestCalendar_Getters(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	eventID := common.NewEventID()

	cal, _ := calendar.NewCalendar(now, tenantID, "Test Calendar", "Test Description", []common.EventID{eventID})

	// Test all getters return expected values
	if cal.TenantID() != tenantID {
		t.Errorf("TenantID() mismatch")
	}

	if cal.Title() != "Test Calendar" {
		t.Errorf("Title() mismatch")
	}

	if cal.Description() != "Test Description" {
		t.Errorf("Description() mismatch")
	}

	if cal.IsPublic() != false {
		t.Errorf("IsPublic() should be false for new calendar")
	}

	if cal.PublicToken() != nil {
		t.Errorf("PublicToken() should be nil for new calendar")
	}

	if len(cal.EventIDs()) != 1 {
		t.Errorf("EventIDs() length mismatch")
	}

	if !cal.CreatedAt().Equal(now) {
		t.Errorf("CreatedAt() mismatch: got %v, want %v", cal.CreatedAt(), now)
	}

	if !cal.UpdatedAt().Equal(now) {
		t.Errorf("UpdatedAt() mismatch: got %v, want %v", cal.UpdatedAt(), now)
	}
}
