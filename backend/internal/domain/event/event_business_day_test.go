package event_test

import (
	"testing"
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/event"
)

// =====================================================
// OccurrenceType Tests
// =====================================================

func TestOccurrenceType_Validate_Success(t *testing.T) {
	tests := []struct {
		name           string
		occurrenceType event.OccurrenceType
	}{
		{"recurring", event.OccurrenceTypeRecurring},
		{"special", event.OccurrenceTypeSpecial},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.occurrenceType.Validate()
			if err != nil {
				t.Errorf("Validate() should succeed for %s, got error: %v", tt.name, err)
			}
		})
	}
}

func TestOccurrenceType_Validate_Error(t *testing.T) {
	invalidTypes := []event.OccurrenceType{
		"",
		"invalid",
		"RECURRING",
		"manual",
	}

	for _, ot := range invalidTypes {
		t.Run(string(ot), func(t *testing.T) {
			err := ot.Validate()
			if err == nil {
				t.Errorf("Validate() should fail for '%s'", ot)
			}
		})
	}
}

// =====================================================
// BusinessDayID Tests
// =====================================================

func TestNewBusinessDayID(t *testing.T) {
	id := event.NewBusinessDayID()
	if id == "" {
		t.Error("NewBusinessDayID() should return non-empty ID")
	}

	if err := id.Validate(); err != nil {
		t.Errorf("NewBusinessDayID() should create valid ID: %v", err)
	}
}

func TestNewBusinessDayIDWithTime(t *testing.T) {
	now := time.Now()
	id := event.NewBusinessDayIDWithTime(now)

	if id == "" {
		t.Error("NewBusinessDayIDWithTime() should return non-empty ID")
	}

	if err := id.Validate(); err != nil {
		t.Errorf("NewBusinessDayIDWithTime() should create valid ID: %v", err)
	}
}

func TestBusinessDayID_String(t *testing.T) {
	id := event.NewBusinessDayID()
	str := id.String()

	if str == "" {
		t.Error("String() should return non-empty string")
	}

	if str != string(id) {
		t.Errorf("String() mismatch: got %v, want %v", str, string(id))
	}
}

func TestBusinessDayID_Validate_Error(t *testing.T) {
	tests := []struct {
		name string
		id   event.BusinessDayID
	}{
		{"empty", event.BusinessDayID("")},
		{"invalid format", event.BusinessDayID("invalid")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.id.Validate()
			if err == nil {
				t.Errorf("Validate() should fail for %s", tt.name)
			}
		})
	}
}

func TestParseBusinessDayID_Success(t *testing.T) {
	original := event.NewBusinessDayID()
	parsed, err := event.ParseBusinessDayID(original.String())

	if err != nil {
		t.Fatalf("ParseBusinessDayID() should succeed: %v", err)
	}

	if parsed != original {
		t.Errorf("ParseBusinessDayID() mismatch: got %v, want %v", parsed, original)
	}
}

func TestParseBusinessDayID_Error(t *testing.T) {
	_, err := event.ParseBusinessDayID("invalid")
	if err == nil {
		t.Error("ParseBusinessDayID() should fail for invalid ID")
	}
}

// =====================================================
// NewEventBusinessDay Tests
// =====================================================

func TestNewEventBusinessDay_Success_Special(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	eventID := common.NewEventID()
	targetDate := now.AddDate(0, 0, 7)
	startTime := time.Date(2000, 1, 1, 20, 0, 0, 0, time.UTC)
	endTime := time.Date(2000, 1, 1, 22, 0, 0, 0, time.UTC)

	bd, err := event.NewEventBusinessDay(
		now,
		tenantID,
		eventID,
		targetDate,
		startTime,
		endTime,
		event.OccurrenceTypeSpecial,
		nil,
	)

	if err != nil {
		t.Fatalf("NewEventBusinessDay() should succeed: %v", err)
	}

	if bd.TenantID() != tenantID {
		t.Errorf("TenantID mismatch: got %v, want %v", bd.TenantID(), tenantID)
	}

	if bd.EventID() != eventID {
		t.Errorf("EventID mismatch: got %v, want %v", bd.EventID(), eventID)
	}

	if bd.OccurrenceType() != event.OccurrenceTypeSpecial {
		t.Errorf("OccurrenceType mismatch: got %v, want %v", bd.OccurrenceType(), event.OccurrenceTypeSpecial)
	}

	if bd.RecurringPatternID() != nil {
		t.Error("RecurringPatternID should be nil for special occurrence")
	}

	if !bd.IsActive() {
		t.Error("New business day should be active")
	}

	if bd.IsDeleted() {
		t.Error("New business day should not be deleted")
	}
}

func TestNewEventBusinessDay_Success_Recurring(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	eventID := common.NewEventID()
	patternID := common.NewEventID()
	targetDate := now.AddDate(0, 0, 7)
	startTime := time.Date(2000, 1, 1, 20, 0, 0, 0, time.UTC)
	endTime := time.Date(2000, 1, 1, 22, 0, 0, 0, time.UTC)

	bd, err := event.NewEventBusinessDay(
		now,
		tenantID,
		eventID,
		targetDate,
		startTime,
		endTime,
		event.OccurrenceTypeRecurring,
		&patternID,
	)

	if err != nil {
		t.Fatalf("NewEventBusinessDay() should succeed: %v", err)
	}

	if bd.OccurrenceType() != event.OccurrenceTypeRecurring {
		t.Errorf("OccurrenceType mismatch: got %v, want %v", bd.OccurrenceType(), event.OccurrenceTypeRecurring)
	}

	if bd.RecurringPatternID() == nil {
		t.Error("RecurringPatternID should not be nil for recurring occurrence")
	}

	if *bd.RecurringPatternID() != patternID {
		t.Errorf("RecurringPatternID mismatch: got %v, want %v", *bd.RecurringPatternID(), patternID)
	}
}

func TestNewEventBusinessDay_ErrorWhenInvalidTenantID(t *testing.T) {
	now := time.Now()
	eventID := common.NewEventID()
	targetDate := now.AddDate(0, 0, 7)
	startTime := time.Date(2000, 1, 1, 20, 0, 0, 0, time.UTC)
	endTime := time.Date(2000, 1, 1, 22, 0, 0, 0, time.UTC)

	_, err := event.NewEventBusinessDay(
		now,
		common.TenantID(""), // Invalid
		eventID,
		targetDate,
		startTime,
		endTime,
		event.OccurrenceTypeSpecial,
		nil,
	)

	if err == nil {
		t.Error("NewEventBusinessDay() should fail when tenant ID is invalid")
	}
}

func TestNewEventBusinessDay_ErrorWhenInvalidEventID(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	targetDate := now.AddDate(0, 0, 7)
	startTime := time.Date(2000, 1, 1, 20, 0, 0, 0, time.UTC)
	endTime := time.Date(2000, 1, 1, 22, 0, 0, 0, time.UTC)

	_, err := event.NewEventBusinessDay(
		now,
		tenantID,
		common.EventID(""), // Invalid
		targetDate,
		startTime,
		endTime,
		event.OccurrenceTypeSpecial,
		nil,
	)

	if err == nil {
		t.Error("NewEventBusinessDay() should fail when event ID is invalid")
	}
}

func TestNewEventBusinessDay_ErrorWhenInvalidOccurrenceType(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	eventID := common.NewEventID()
	targetDate := now.AddDate(0, 0, 7)
	startTime := time.Date(2000, 1, 1, 20, 0, 0, 0, time.UTC)
	endTime := time.Date(2000, 1, 1, 22, 0, 0, 0, time.UTC)

	_, err := event.NewEventBusinessDay(
		now,
		tenantID,
		eventID,
		targetDate,
		startTime,
		endTime,
		event.OccurrenceType("invalid"),
		nil,
	)

	if err == nil {
		t.Error("NewEventBusinessDay() should fail when occurrence type is invalid")
	}
}

func TestNewEventBusinessDay_ErrorWhenSpecialWithPatternID(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	eventID := common.NewEventID()
	patternID := common.NewEventID()
	targetDate := now.AddDate(0, 0, 7)
	startTime := time.Date(2000, 1, 1, 20, 0, 0, 0, time.UTC)
	endTime := time.Date(2000, 1, 1, 22, 0, 0, 0, time.UTC)

	_, err := event.NewEventBusinessDay(
		now,
		tenantID,
		eventID,
		targetDate,
		startTime,
		endTime,
		event.OccurrenceTypeSpecial,
		&patternID, // Should be nil for special
	)

	if err == nil {
		t.Error("NewEventBusinessDay() should fail when special occurrence has pattern ID")
	}
}

// =====================================================
// ReconstructEventBusinessDay Tests
// =====================================================

func TestReconstructEventBusinessDay_Success(t *testing.T) {
	now := time.Now()
	businessDayID := event.NewBusinessDayID()
	tenantID := common.NewTenantID()
	eventID := common.NewEventID()
	targetDate := now.AddDate(0, 0, 7)
	startTime := time.Date(2000, 1, 1, 20, 0, 0, 0, time.UTC)
	endTime := time.Date(2000, 1, 1, 22, 0, 0, 0, time.UTC)

	bd, err := event.ReconstructEventBusinessDay(
		businessDayID,
		tenantID,
		eventID,
		targetDate,
		startTime,
		endTime,
		event.OccurrenceTypeSpecial,
		nil,
		true,
		nil,
		nil,
		now,
		now,
		nil,
	)

	if err != nil {
		t.Fatalf("ReconstructEventBusinessDay() should succeed: %v", err)
	}

	if bd.BusinessDayID() != businessDayID {
		t.Errorf("BusinessDayID mismatch: got %v, want %v", bd.BusinessDayID(), businessDayID)
	}
}

func TestReconstructEventBusinessDay_WithValidPeriod(t *testing.T) {
	now := time.Now()
	businessDayID := event.NewBusinessDayID()
	tenantID := common.NewTenantID()
	eventID := common.NewEventID()
	targetDate := now.AddDate(0, 0, 7)
	startTime := time.Date(2000, 1, 1, 20, 0, 0, 0, time.UTC)
	endTime := time.Date(2000, 1, 1, 22, 0, 0, 0, time.UTC)
	validFrom := now
	validTo := now.AddDate(0, 1, 0)

	bd, err := event.ReconstructEventBusinessDay(
		businessDayID,
		tenantID,
		eventID,
		targetDate,
		startTime,
		endTime,
		event.OccurrenceTypeSpecial,
		nil,
		true,
		&validFrom,
		&validTo,
		now,
		now,
		nil,
	)

	if err != nil {
		t.Fatalf("ReconstructEventBusinessDay() should succeed: %v", err)
	}

	if bd.ValidFrom() == nil || bd.ValidTo() == nil {
		t.Error("ValidFrom and ValidTo should be set")
	}
}

func TestReconstructEventBusinessDay_ErrorWhenValidFromAfterValidTo(t *testing.T) {
	now := time.Now()
	businessDayID := event.NewBusinessDayID()
	tenantID := common.NewTenantID()
	eventID := common.NewEventID()
	targetDate := now.AddDate(0, 0, 7)
	startTime := time.Date(2000, 1, 1, 20, 0, 0, 0, time.UTC)
	endTime := time.Date(2000, 1, 1, 22, 0, 0, 0, time.UTC)
	validFrom := now.AddDate(0, 1, 0) // After validTo
	validTo := now

	_, err := event.ReconstructEventBusinessDay(
		businessDayID,
		tenantID,
		eventID,
		targetDate,
		startTime,
		endTime,
		event.OccurrenceTypeSpecial,
		nil,
		true,
		&validFrom,
		&validTo,
		now,
		now,
		nil,
	)

	if err == nil {
		t.Error("ReconstructEventBusinessDay() should fail when validFrom is after validTo")
	}
}

func TestReconstructEventBusinessDay_ErrorWhenOnlyValidFromSet(t *testing.T) {
	now := time.Now()
	businessDayID := event.NewBusinessDayID()
	tenantID := common.NewTenantID()
	eventID := common.NewEventID()
	targetDate := now.AddDate(0, 0, 7)
	startTime := time.Date(2000, 1, 1, 20, 0, 0, 0, time.UTC)
	endTime := time.Date(2000, 1, 1, 22, 0, 0, 0, time.UTC)
	validFrom := now

	_, err := event.ReconstructEventBusinessDay(
		businessDayID,
		tenantID,
		eventID,
		targetDate,
		startTime,
		endTime,
		event.OccurrenceTypeSpecial,
		nil,
		true,
		&validFrom,
		nil, // Only validFrom set
		now,
		now,
		nil,
	)

	if err == nil {
		t.Error("ReconstructEventBusinessDay() should fail when only validFrom is set")
	}
}

// =====================================================
// EventBusinessDay Methods Tests
// =====================================================

func TestEventBusinessDay_Activate(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	eventID := common.NewEventID()
	targetDate := now.AddDate(0, 0, 7)
	startTime := time.Date(2000, 1, 1, 20, 0, 0, 0, time.UTC)
	endTime := time.Date(2000, 1, 1, 22, 0, 0, 0, time.UTC)

	bd, _ := event.NewEventBusinessDay(now, tenantID, eventID, targetDate, startTime, endTime, event.OccurrenceTypeSpecial, nil)

	bd.Deactivate()
	if bd.IsActive() {
		t.Error("Deactivate() should set isActive to false")
	}

	bd.Activate()
	if !bd.IsActive() {
		t.Error("Activate() should set isActive to true")
	}
}

func TestEventBusinessDay_Delete(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	eventID := common.NewEventID()
	targetDate := now.AddDate(0, 0, 7)
	startTime := time.Date(2000, 1, 1, 20, 0, 0, 0, time.UTC)
	endTime := time.Date(2000, 1, 1, 22, 0, 0, 0, time.UTC)

	bd, _ := event.NewEventBusinessDay(now, tenantID, eventID, targetDate, startTime, endTime, event.OccurrenceTypeSpecial, nil)

	if bd.IsDeleted() {
		t.Error("New business day should not be deleted")
	}

	bd.Delete()

	if !bd.IsDeleted() {
		t.Error("Delete() should mark business day as deleted")
	}

	if bd.DeletedAt() == nil {
		t.Error("DeletedAt() should not be nil after deletion")
	}
}

func TestEventBusinessDay_SetValidPeriod_Success(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	eventID := common.NewEventID()
	targetDate := now.AddDate(0, 0, 7)
	startTime := time.Date(2000, 1, 1, 20, 0, 0, 0, time.UTC)
	endTime := time.Date(2000, 1, 1, 22, 0, 0, 0, time.UTC)

	bd, _ := event.NewEventBusinessDay(now, tenantID, eventID, targetDate, startTime, endTime, event.OccurrenceTypeSpecial, nil)

	validFrom := now
	validTo := now.AddDate(0, 1, 0)

	err := bd.SetValidPeriod(&validFrom, &validTo)

	if err != nil {
		t.Fatalf("SetValidPeriod() should succeed: %v", err)
	}

	if bd.ValidFrom() == nil || bd.ValidTo() == nil {
		t.Error("ValidFrom and ValidTo should be set")
	}
}

func TestEventBusinessDay_SetValidPeriod_ClearPeriod(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	eventID := common.NewEventID()
	targetDate := now.AddDate(0, 0, 7)
	startTime := time.Date(2000, 1, 1, 20, 0, 0, 0, time.UTC)
	endTime := time.Date(2000, 1, 1, 22, 0, 0, 0, time.UTC)

	bd, _ := event.NewEventBusinessDay(now, tenantID, eventID, targetDate, startTime, endTime, event.OccurrenceTypeSpecial, nil)

	err := bd.SetValidPeriod(nil, nil)

	if err != nil {
		t.Fatalf("SetValidPeriod(nil, nil) should succeed: %v", err)
	}
}

func TestEventBusinessDay_SetValidPeriod_Error(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	eventID := common.NewEventID()
	targetDate := now.AddDate(0, 0, 7)
	startTime := time.Date(2000, 1, 1, 20, 0, 0, 0, time.UTC)
	endTime := time.Date(2000, 1, 1, 22, 0, 0, 0, time.UTC)

	bd, _ := event.NewEventBusinessDay(now, tenantID, eventID, targetDate, startTime, endTime, event.OccurrenceTypeSpecial, nil)

	// validFrom after validTo
	validFrom := now.AddDate(0, 1, 0)
	validTo := now

	err := bd.SetValidPeriod(&validFrom, &validTo)

	if err == nil {
		t.Error("SetValidPeriod() should fail when validFrom is after validTo")
	}
}

func TestEventBusinessDay_IsValidOn(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	eventID := common.NewEventID()
	targetDate := now.AddDate(0, 0, 7)
	startTime := time.Date(2000, 1, 1, 20, 0, 0, 0, time.UTC)
	endTime := time.Date(2000, 1, 1, 22, 0, 0, 0, time.UTC)

	bd, _ := event.NewEventBusinessDay(now, tenantID, eventID, targetDate, startTime, endTime, event.OccurrenceTypeSpecial, nil)

	// Without valid period, should always be valid when active
	if !bd.IsValidOn(now) {
		t.Error("Active business day without valid period should be valid on any date")
	}

	// Deactivated should not be valid
	bd.Deactivate()
	if bd.IsValidOn(now) {
		t.Error("Deactivated business day should not be valid")
	}
}

func TestEventBusinessDay_IsValidOn_WithPeriod(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	eventID := common.NewEventID()
	targetDate := now.AddDate(0, 0, 7)
	startTime := time.Date(2000, 1, 1, 20, 0, 0, 0, time.UTC)
	endTime := time.Date(2000, 1, 1, 22, 0, 0, 0, time.UTC)

	bd, _ := event.NewEventBusinessDay(now, tenantID, eventID, targetDate, startTime, endTime, event.OccurrenceTypeSpecial, nil)

	validFrom := now
	validTo := now.AddDate(0, 0, 14)
	_ = bd.SetValidPeriod(&validFrom, &validTo)

	// Within valid period
	if !bd.IsValidOn(now.AddDate(0, 0, 7)) {
		t.Error("Should be valid within the period")
	}

	// After valid period
	if bd.IsValidOn(now.AddDate(0, 1, 0)) {
		t.Error("Should not be valid after the period")
	}

	// Before valid period
	if bd.IsValidOn(now.AddDate(0, 0, -7)) {
		t.Error("Should not be valid before the period")
	}
}

func TestEventBusinessDay_DayOfWeek(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantID()
	eventID := common.NewEventID()

	// Create a business day for a known date (e.g., a Monday)
	monday := time.Date(2025, 12, 29, 0, 0, 0, 0, time.UTC) // This is a Monday
	startTime := time.Date(2000, 1, 1, 20, 0, 0, 0, time.UTC)
	endTime := time.Date(2000, 1, 1, 22, 0, 0, 0, time.UTC)

	bd, _ := event.NewEventBusinessDay(now, tenantID, eventID, monday, startTime, endTime, event.OccurrenceTypeSpecial, nil)

	if bd.DayOfWeek() != time.Monday {
		t.Errorf("DayOfWeek() mismatch: got %v, want %v", bd.DayOfWeek(), time.Monday)
	}

	if bd.DayOfWeekString() != event.Monday {
		t.Errorf("DayOfWeekString() mismatch: got %v, want %v", bd.DayOfWeekString(), event.Monday)
	}
}
