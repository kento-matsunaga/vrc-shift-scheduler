package event_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/event"
)

// =====================================================
// PatternType Tests
// =====================================================

func TestPatternType_Validate_Success(t *testing.T) {
	testCases := []struct {
		name        string
		patternType event.PatternType
	}{
		{"weekly", event.PatternTypeWeekly},
		{"monthly_date", event.PatternTypeMonthlyDate},
		{"custom", event.PatternTypeCustom},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.patternType.Validate()
			if err != nil {
				t.Errorf("Validate() should succeed for %s, got error: %v", tc.name, err)
			}
		})
	}
}

func TestPatternType_Validate_Error(t *testing.T) {
	testCases := []struct {
		name        string
		patternType event.PatternType
	}{
		{"empty", event.PatternType("")},
		{"invalid", event.PatternType("invalid")},
		{"WEEKLY", event.PatternType("WEEKLY")},
		{"daily", event.PatternType("daily")},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.patternType.Validate()
			if err == nil {
				t.Errorf("Validate() should fail for %s", tc.name)
			}
		})
	}
}

// =====================================================
// DayOfWeek Tests
// =====================================================

func TestDayOfWeek_Validate_Success(t *testing.T) {
	testCases := []struct {
		name string
		dow  event.DayOfWeek
	}{
		{"Monday", event.Monday},
		{"Tuesday", event.Tuesday},
		{"Wednesday", event.Wednesday},
		{"Thursday", event.Thursday},
		{"Friday", event.Friday},
		{"Saturday", event.Saturday},
		{"Sunday", event.Sunday},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.dow.Validate()
			if err != nil {
				t.Errorf("Validate() should succeed for %s, got error: %v", tc.name, err)
			}
		})
	}
}

func TestDayOfWeek_Validate_Error(t *testing.T) {
	testCases := []struct {
		name string
		dow  event.DayOfWeek
	}{
		{"empty", event.DayOfWeek("")},
		{"invalid", event.DayOfWeek("invalid")},
		{"lowercase", event.DayOfWeek("mon")},
		{"full_name", event.DayOfWeek("Monday")},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.dow.Validate()
			if err == nil {
				t.Errorf("Validate() should fail for %s", tc.name)
			}
		})
	}
}

// =====================================================
// WeeklyPatternConfig Tests
// =====================================================

func TestWeeklyPatternConfig_Validate_Success(t *testing.T) {
	config := &event.WeeklyPatternConfig{
		DayOfWeeks: []event.DayOfWeek{event.Saturday, event.Sunday},
		StartTime:  "20:00",
		EndTime:    "22:00",
	}

	err := config.Validate()
	if err != nil {
		t.Errorf("Validate() should succeed, got error: %v", err)
	}
}

func TestWeeklyPatternConfig_Validate_ErrorWhenNoDays(t *testing.T) {
	config := &event.WeeklyPatternConfig{
		DayOfWeeks: []event.DayOfWeek{},
		StartTime:  "20:00",
		EndTime:    "22:00",
	}

	err := config.Validate()
	if err == nil {
		t.Error("Validate() should fail when day_of_weeks is empty")
	}
}

func TestWeeklyPatternConfig_Validate_ErrorWhenTooManyDays(t *testing.T) {
	config := &event.WeeklyPatternConfig{
		DayOfWeeks: []event.DayOfWeek{
			event.Monday, event.Tuesday, event.Wednesday, event.Thursday,
			event.Friday, event.Saturday, event.Sunday, event.Monday,
		},
		StartTime: "20:00",
		EndTime:   "22:00",
	}

	err := config.Validate()
	if err == nil {
		t.Error("Validate() should fail when day_of_weeks exceeds 7")
	}
}

func TestWeeklyPatternConfig_Validate_ErrorWhenInvalidDayOfWeek(t *testing.T) {
	config := &event.WeeklyPatternConfig{
		DayOfWeeks: []event.DayOfWeek{event.DayOfWeek("INVALID")},
		StartTime:  "20:00",
		EndTime:    "22:00",
	}

	err := config.Validate()
	if err == nil {
		t.Error("Validate() should fail when day_of_week is invalid")
	}
}

func TestWeeklyPatternConfig_Validate_ErrorWhenNoStartTime(t *testing.T) {
	config := &event.WeeklyPatternConfig{
		DayOfWeeks: []event.DayOfWeek{event.Saturday},
		StartTime:  "",
		EndTime:    "22:00",
	}

	err := config.Validate()
	if err == nil {
		t.Error("Validate() should fail when start_time is empty")
	}
}

func TestWeeklyPatternConfig_Validate_ErrorWhenNoEndTime(t *testing.T) {
	config := &event.WeeklyPatternConfig{
		DayOfWeeks: []event.DayOfWeek{event.Saturday},
		StartTime:  "20:00",
		EndTime:    "",
	}

	err := config.Validate()
	if err == nil {
		t.Error("Validate() should fail when end_time is empty")
	}
}

func TestWeeklyPatternConfig_ToJSON(t *testing.T) {
	config := &event.WeeklyPatternConfig{
		DayOfWeeks: []event.DayOfWeek{event.Saturday},
		StartTime:  "20:00",
		EndTime:    "22:00",
	}

	jsonBytes, err := config.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON() should succeed, got error: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &parsed); err != nil {
		t.Fatalf("Should be valid JSON: %v", err)
	}

	if parsed["start_time"] != "20:00" {
		t.Errorf("start_time mismatch: got %v, want '20:00'", parsed["start_time"])
	}
}

// =====================================================
// MonthlyDatePatternConfig Tests
// =====================================================

func TestMonthlyDatePatternConfig_Validate_Success(t *testing.T) {
	config := &event.MonthlyDatePatternConfig{
		Dates:     []int{1, 15, 28},
		StartTime: "20:00",
		EndTime:   "22:00",
	}

	err := config.Validate()
	if err != nil {
		t.Errorf("Validate() should succeed, got error: %v", err)
	}
}

func TestMonthlyDatePatternConfig_Validate_ErrorWhenNoDates(t *testing.T) {
	config := &event.MonthlyDatePatternConfig{
		Dates:     []int{},
		StartTime: "20:00",
		EndTime:   "22:00",
	}

	err := config.Validate()
	if err == nil {
		t.Error("Validate() should fail when dates is empty")
	}
}

func TestMonthlyDatePatternConfig_Validate_ErrorWhenDateTooLow(t *testing.T) {
	config := &event.MonthlyDatePatternConfig{
		Dates:     []int{0},
		StartTime: "20:00",
		EndTime:   "22:00",
	}

	err := config.Validate()
	if err == nil {
		t.Error("Validate() should fail when date is less than 1")
	}
}

func TestMonthlyDatePatternConfig_Validate_ErrorWhenDateTooHigh(t *testing.T) {
	config := &event.MonthlyDatePatternConfig{
		Dates:     []int{32},
		StartTime: "20:00",
		EndTime:   "22:00",
	}

	err := config.Validate()
	if err == nil {
		t.Error("Validate() should fail when date is greater than 31")
	}
}

func TestMonthlyDatePatternConfig_Validate_ErrorWhenNoTime(t *testing.T) {
	config := &event.MonthlyDatePatternConfig{
		Dates:     []int{15},
		StartTime: "",
		EndTime:   "",
	}

	err := config.Validate()
	if err == nil {
		t.Error("Validate() should fail when times are empty")
	}
}

func TestMonthlyDatePatternConfig_ToJSON(t *testing.T) {
	config := &event.MonthlyDatePatternConfig{
		Dates:     []int{1, 15},
		StartTime: "20:00",
		EndTime:   "22:00",
	}

	jsonBytes, err := config.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON() should succeed, got error: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &parsed); err != nil {
		t.Fatalf("Should be valid JSON: %v", err)
	}

	dates, ok := parsed["dates"].([]interface{})
	if !ok {
		t.Fatal("dates should be an array")
	}
	if len(dates) != 2 {
		t.Errorf("dates length mismatch: got %d, want 2", len(dates))
	}
}

// =====================================================
// CustomPatternConfig Tests
// =====================================================

func TestCustomPatternConfig_Validate_Success(t *testing.T) {
	config := event.CustomPatternConfig{
		"key1": "value1",
		"key2": 123,
	}

	err := config.Validate()
	if err != nil {
		t.Errorf("Validate() should succeed, got error: %v", err)
	}
}

func TestCustomPatternConfig_Validate_SuccessEmpty(t *testing.T) {
	config := event.CustomPatternConfig{}

	err := config.Validate()
	if err != nil {
		t.Errorf("Validate() should succeed for empty config, got error: %v", err)
	}
}

func TestCustomPatternConfig_ToJSON(t *testing.T) {
	config := event.CustomPatternConfig{
		"key1": "value1",
	}

	jsonBytes, err := config.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON() should succeed, got error: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &parsed); err != nil {
		t.Fatalf("Should be valid JSON: %v", err)
	}

	if parsed["key1"] != "value1" {
		t.Errorf("key1 mismatch: got %v, want 'value1'", parsed["key1"])
	}
}

// =====================================================
// NewRecurringPattern Tests
// =====================================================

func TestNewRecurringPattern_Success_Weekly(t *testing.T) {
	tenantID := common.NewTenantID()
	eventID := common.NewEventID()

	config := &event.WeeklyPatternConfig{
		DayOfWeeks: []event.DayOfWeek{event.Saturday, event.Sunday},
		StartTime:  "20:00",
		EndTime:    "22:00",
	}

	pattern, err := event.NewRecurringPattern(tenantID, eventID, event.PatternTypeWeekly, config)
	if err != nil {
		t.Fatalf("NewRecurringPattern() should succeed, got error: %v", err)
	}

	if pattern.TenantID() != tenantID {
		t.Errorf("TenantID mismatch: got %v, want %v", pattern.TenantID(), tenantID)
	}

	if pattern.EventID() != eventID {
		t.Errorf("EventID mismatch: got %v, want %v", pattern.EventID(), eventID)
	}

	if pattern.PatternType() != event.PatternTypeWeekly {
		t.Errorf("PatternType mismatch: got %v, want weekly", pattern.PatternType())
	}

	if pattern.PatternID().String() == "" {
		t.Error("PatternID should be generated")
	}

	if pattern.IsDeleted() {
		t.Error("IsDeleted should be false for new pattern")
	}

	if pattern.CreatedAt().IsZero() {
		t.Error("CreatedAt should be set")
	}

	if pattern.UpdatedAt().IsZero() {
		t.Error("UpdatedAt should be set")
	}

	if pattern.DeletedAt() != nil {
		t.Error("DeletedAt should be nil for new pattern")
	}
}

func TestNewRecurringPattern_Success_MonthlyDate(t *testing.T) {
	tenantID := common.NewTenantID()
	eventID := common.NewEventID()

	config := &event.MonthlyDatePatternConfig{
		Dates:     []int{1, 15},
		StartTime: "20:00",
		EndTime:   "22:00",
	}

	pattern, err := event.NewRecurringPattern(tenantID, eventID, event.PatternTypeMonthlyDate, config)
	if err != nil {
		t.Fatalf("NewRecurringPattern() should succeed, got error: %v", err)
	}

	if pattern.PatternType() != event.PatternTypeMonthlyDate {
		t.Errorf("PatternType mismatch: got %v, want monthly_date", pattern.PatternType())
	}
}

func TestNewRecurringPattern_Success_Custom(t *testing.T) {
	tenantID := common.NewTenantID()
	eventID := common.NewEventID()

	config := event.CustomPatternConfig{
		"custom_key": "custom_value",
	}

	pattern, err := event.NewRecurringPattern(tenantID, eventID, event.PatternTypeCustom, config)
	if err != nil {
		t.Fatalf("NewRecurringPattern() should succeed, got error: %v", err)
	}

	if pattern.PatternType() != event.PatternTypeCustom {
		t.Errorf("PatternType mismatch: got %v, want custom", pattern.PatternType())
	}
}

func TestNewRecurringPattern_ErrorWhenInvalidTenantID(t *testing.T) {
	eventID := common.NewEventID()

	config := &event.WeeklyPatternConfig{
		DayOfWeeks: []event.DayOfWeek{event.Saturday},
		StartTime:  "20:00",
		EndTime:    "22:00",
	}

	_, err := event.NewRecurringPattern(common.TenantID(""), eventID, event.PatternTypeWeekly, config)
	if err == nil {
		t.Error("NewRecurringPattern() should fail when tenant_id is invalid")
	}
}

func TestNewRecurringPattern_ErrorWhenInvalidEventID(t *testing.T) {
	tenantID := common.NewTenantID()

	config := &event.WeeklyPatternConfig{
		DayOfWeeks: []event.DayOfWeek{event.Saturday},
		StartTime:  "20:00",
		EndTime:    "22:00",
	}

	_, err := event.NewRecurringPattern(tenantID, common.EventID(""), event.PatternTypeWeekly, config)
	if err == nil {
		t.Error("NewRecurringPattern() should fail when event_id is invalid")
	}
}

func TestNewRecurringPattern_ErrorWhenInvalidPatternType(t *testing.T) {
	tenantID := common.NewTenantID()
	eventID := common.NewEventID()

	config := &event.WeeklyPatternConfig{
		DayOfWeeks: []event.DayOfWeek{event.Saturday},
		StartTime:  "20:00",
		EndTime:    "22:00",
	}

	_, err := event.NewRecurringPattern(tenantID, eventID, event.PatternType("invalid"), config)
	if err == nil {
		t.Error("NewRecurringPattern() should fail when pattern_type is invalid")
	}
}

func TestNewRecurringPattern_ErrorWhenNilConfig(t *testing.T) {
	tenantID := common.NewTenantID()
	eventID := common.NewEventID()

	_, err := event.NewRecurringPattern(tenantID, eventID, event.PatternTypeWeekly, nil)
	if err == nil {
		t.Error("NewRecurringPattern() should fail when config is nil")
	}
}

func TestNewRecurringPattern_ErrorWhenInvalidConfig(t *testing.T) {
	tenantID := common.NewTenantID()
	eventID := common.NewEventID()

	config := &event.WeeklyPatternConfig{
		DayOfWeeks: []event.DayOfWeek{}, // Invalid: empty
		StartTime:  "20:00",
		EndTime:    "22:00",
	}

	_, err := event.NewRecurringPattern(tenantID, eventID, event.PatternTypeWeekly, config)
	if err == nil {
		t.Error("NewRecurringPattern() should fail when config is invalid")
	}
}

// =====================================================
// ReconstructRecurringPattern Tests
// =====================================================

func TestReconstructRecurringPattern_Success_Weekly(t *testing.T) {
	patternID := common.NewEventID()
	tenantID := common.NewTenantID()
	eventID := common.NewEventID()
	createdAt := time.Now().Add(-time.Hour)
	updatedAt := time.Now()

	configJSON := []byte(`{"day_of_weeks":["SAT","SUN"],"start_time":"20:00","end_time":"22:00"}`)

	pattern, err := event.ReconstructRecurringPattern(
		patternID,
		tenantID,
		eventID,
		event.PatternTypeWeekly,
		configJSON,
		createdAt,
		updatedAt,
		nil,
	)

	if err != nil {
		t.Fatalf("ReconstructRecurringPattern() should succeed, got error: %v", err)
	}

	if pattern.PatternID() != patternID {
		t.Errorf("PatternID mismatch: got %v, want %v", pattern.PatternID(), patternID)
	}

	if pattern.PatternType() != event.PatternTypeWeekly {
		t.Errorf("PatternType mismatch: got %v, want weekly", pattern.PatternType())
	}
}

func TestReconstructRecurringPattern_Success_MonthlyDate(t *testing.T) {
	patternID := common.NewEventID()
	tenantID := common.NewTenantID()
	eventID := common.NewEventID()
	createdAt := time.Now()
	updatedAt := time.Now()

	configJSON := []byte(`{"dates":[1,15],"start_time":"20:00","end_time":"22:00"}`)

	pattern, err := event.ReconstructRecurringPattern(
		patternID,
		tenantID,
		eventID,
		event.PatternTypeMonthlyDate,
		configJSON,
		createdAt,
		updatedAt,
		nil,
	)

	if err != nil {
		t.Fatalf("ReconstructRecurringPattern() should succeed, got error: %v", err)
	}

	if pattern.PatternType() != event.PatternTypeMonthlyDate {
		t.Errorf("PatternType mismatch: got %v, want monthly_date", pattern.PatternType())
	}
}

func TestReconstructRecurringPattern_Success_Custom(t *testing.T) {
	patternID := common.NewEventID()
	tenantID := common.NewTenantID()
	eventID := common.NewEventID()
	createdAt := time.Now()
	updatedAt := time.Now()

	configJSON := []byte(`{"custom_key":"custom_value"}`)

	pattern, err := event.ReconstructRecurringPattern(
		patternID,
		tenantID,
		eventID,
		event.PatternTypeCustom,
		configJSON,
		createdAt,
		updatedAt,
		nil,
	)

	if err != nil {
		t.Fatalf("ReconstructRecurringPattern() should succeed, got error: %v", err)
	}

	if pattern.PatternType() != event.PatternTypeCustom {
		t.Errorf("PatternType mismatch: got %v, want custom", pattern.PatternType())
	}
}

func TestReconstructRecurringPattern_Success_WithDeletedAt(t *testing.T) {
	patternID := common.NewEventID()
	tenantID := common.NewTenantID()
	eventID := common.NewEventID()
	createdAt := time.Now().Add(-time.Hour)
	updatedAt := time.Now()
	deletedAt := time.Now()

	configJSON := []byte(`{"day_of_weeks":["SAT"],"start_time":"20:00","end_time":"22:00"}`)

	pattern, err := event.ReconstructRecurringPattern(
		patternID,
		tenantID,
		eventID,
		event.PatternTypeWeekly,
		configJSON,
		createdAt,
		updatedAt,
		&deletedAt,
	)

	if err != nil {
		t.Fatalf("ReconstructRecurringPattern() should succeed, got error: %v", err)
	}

	if !pattern.IsDeleted() {
		t.Error("IsDeleted should be true when deletedAt is set")
	}

	if pattern.DeletedAt() == nil {
		t.Error("DeletedAt should not be nil")
	}
}

func TestReconstructRecurringPattern_ErrorWhenInvalidJSON(t *testing.T) {
	patternID := common.NewEventID()
	tenantID := common.NewTenantID()
	eventID := common.NewEventID()
	createdAt := time.Now()
	updatedAt := time.Now()

	configJSON := []byte(`{invalid json}`)

	_, err := event.ReconstructRecurringPattern(
		patternID,
		tenantID,
		eventID,
		event.PatternTypeWeekly,
		configJSON,
		createdAt,
		updatedAt,
		nil,
	)

	if err == nil {
		t.Error("ReconstructRecurringPattern() should fail when config JSON is invalid")
	}
}

func TestReconstructRecurringPattern_ErrorWhenUnknownPatternType(t *testing.T) {
	patternID := common.NewEventID()
	tenantID := common.NewTenantID()
	eventID := common.NewEventID()
	createdAt := time.Now()
	updatedAt := time.Now()

	configJSON := []byte(`{}`)

	_, err := event.ReconstructRecurringPattern(
		patternID,
		tenantID,
		eventID,
		event.PatternType("unknown"),
		configJSON,
		createdAt,
		updatedAt,
		nil,
	)

	if err == nil {
		t.Error("ReconstructRecurringPattern() should fail when pattern type is unknown")
	}
}

// =====================================================
// RecurringPattern Methods Tests
// =====================================================

func TestRecurringPattern_ConfigJSON(t *testing.T) {
	tenantID := common.NewTenantID()
	eventID := common.NewEventID()

	config := &event.WeeklyPatternConfig{
		DayOfWeeks: []event.DayOfWeek{event.Saturday},
		StartTime:  "20:00",
		EndTime:    "22:00",
	}

	pattern, _ := event.NewRecurringPattern(tenantID, eventID, event.PatternTypeWeekly, config)

	jsonBytes, err := pattern.ConfigJSON()
	if err != nil {
		t.Fatalf("ConfigJSON() should succeed, got error: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &parsed); err != nil {
		t.Fatalf("Should be valid JSON: %v", err)
	}
}

func TestRecurringPattern_UpdateConfig_Success(t *testing.T) {
	tenantID := common.NewTenantID()
	eventID := common.NewEventID()

	config := &event.WeeklyPatternConfig{
		DayOfWeeks: []event.DayOfWeek{event.Saturday},
		StartTime:  "20:00",
		EndTime:    "22:00",
	}

	pattern, _ := event.NewRecurringPattern(tenantID, eventID, event.PatternTypeWeekly, config)
	originalUpdatedAt := pattern.UpdatedAt()

	// Wait a bit to ensure time difference
	time.Sleep(time.Millisecond)

	newConfig := &event.WeeklyPatternConfig{
		DayOfWeeks: []event.DayOfWeek{event.Sunday},
		StartTime:  "21:00",
		EndTime:    "23:00",
	}

	err := pattern.UpdateConfig(newConfig)
	if err != nil {
		t.Fatalf("UpdateConfig() should succeed, got error: %v", err)
	}

	if pattern.Config() != newConfig {
		t.Error("Config should be updated")
	}

	if !pattern.UpdatedAt().After(originalUpdatedAt) {
		t.Error("UpdatedAt should be updated")
	}
}

func TestRecurringPattern_UpdateConfig_ErrorWhenNil(t *testing.T) {
	tenantID := common.NewTenantID()
	eventID := common.NewEventID()

	config := &event.WeeklyPatternConfig{
		DayOfWeeks: []event.DayOfWeek{event.Saturday},
		StartTime:  "20:00",
		EndTime:    "22:00",
	}

	pattern, _ := event.NewRecurringPattern(tenantID, eventID, event.PatternTypeWeekly, config)

	err := pattern.UpdateConfig(nil)
	if err == nil {
		t.Error("UpdateConfig() should fail when config is nil")
	}
}

func TestRecurringPattern_UpdateConfig_ErrorWhenInvalid(t *testing.T) {
	tenantID := common.NewTenantID()
	eventID := common.NewEventID()

	config := &event.WeeklyPatternConfig{
		DayOfWeeks: []event.DayOfWeek{event.Saturday},
		StartTime:  "20:00",
		EndTime:    "22:00",
	}

	pattern, _ := event.NewRecurringPattern(tenantID, eventID, event.PatternTypeWeekly, config)

	invalidConfig := &event.WeeklyPatternConfig{
		DayOfWeeks: []event.DayOfWeek{}, // Invalid: empty
		StartTime:  "20:00",
		EndTime:    "22:00",
	}

	err := pattern.UpdateConfig(invalidConfig)
	if err == nil {
		t.Error("UpdateConfig() should fail when config is invalid")
	}
}

func TestRecurringPattern_Delete(t *testing.T) {
	tenantID := common.NewTenantID()
	eventID := common.NewEventID()

	config := &event.WeeklyPatternConfig{
		DayOfWeeks: []event.DayOfWeek{event.Saturday},
		StartTime:  "20:00",
		EndTime:    "22:00",
	}

	pattern, _ := event.NewRecurringPattern(tenantID, eventID, event.PatternTypeWeekly, config)

	if pattern.IsDeleted() {
		t.Error("IsDeleted should be false before Delete()")
	}

	pattern.Delete()

	if !pattern.IsDeleted() {
		t.Error("IsDeleted should be true after Delete()")
	}

	if pattern.DeletedAt() == nil {
		t.Error("DeletedAt should be set after Delete()")
	}
}
