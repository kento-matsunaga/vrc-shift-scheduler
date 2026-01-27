package system_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/system"
)

func TestNewSetting_Success(t *testing.T) {
	now := time.Now()
	value := json.RawMessage(`{"released": true}`)

	setting, err := system.NewSetting(system.SettingKeyReleaseStatus, value, now)
	if err != nil {
		t.Fatalf("NewSetting failed: %v", err)
	}

	if setting.Key() != system.SettingKeyReleaseStatus {
		t.Errorf("Key = %v, want %v", setting.Key(), system.SettingKeyReleaseStatus)
	}
	if string(setting.Value()) != string(value) {
		t.Errorf("Value = %v, want %v", string(setting.Value()), string(value))
	}
	if !setting.UpdatedAt().Equal(now) {
		t.Errorf("UpdatedAt = %v, want %v", setting.UpdatedAt(), now)
	}
}

func TestNewSetting_ErrorWhenKeyEmpty(t *testing.T) {
	now := time.Now()
	value := json.RawMessage(`{"released": true}`)

	_, err := system.NewSetting("", value, now)
	if err == nil {
		t.Error("Expected error when key is empty, got nil")
	}
}

func TestNewSetting_ErrorWhenValueEmpty(t *testing.T) {
	now := time.Now()

	_, err := system.NewSetting(system.SettingKeyReleaseStatus, nil, now)
	if err == nil {
		t.Error("Expected error when value is nil, got nil")
	}

	_, err = system.NewSetting(system.SettingKeyReleaseStatus, json.RawMessage{}, now)
	if err == nil {
		t.Error("Expected error when value is empty, got nil")
	}
}

func TestSetting_UpdateValue_Success(t *testing.T) {
	now := time.Now()
	initialValue := json.RawMessage(`{"released": false}`)
	setting, err := system.NewSetting(system.SettingKeyReleaseStatus, initialValue, now)
	if err != nil {
		t.Fatalf("NewSetting failed: %v", err)
	}

	newValue := json.RawMessage(`{"released": true}`)
	later := now.Add(time.Hour)
	err = setting.UpdateValue(newValue, later)
	if err != nil {
		t.Fatalf("UpdateValue failed: %v", err)
	}

	if string(setting.Value()) != string(newValue) {
		t.Errorf("Value = %v, want %v", string(setting.Value()), string(newValue))
	}
	if !setting.UpdatedAt().Equal(later) {
		t.Errorf("UpdatedAt = %v, want %v", setting.UpdatedAt(), later)
	}
}

func TestReconstructSetting_Success(t *testing.T) {
	now := time.Now()
	value := json.RawMessage(`{"released": true}`)

	setting, err := system.ReconstructSetting(system.SettingKeyReleaseStatus, value, now)
	if err != nil {
		t.Fatalf("ReconstructSetting failed: %v", err)
	}

	if setting.Key() != system.SettingKeyReleaseStatus {
		t.Errorf("Key = %v, want %v", setting.Key(), system.SettingKeyReleaseStatus)
	}
}

func TestParseReleaseStatus_Success(t *testing.T) {
	tests := []struct {
		name     string
		json     json.RawMessage
		expected bool
	}{
		{"released true", json.RawMessage(`{"released": true}`), true},
		{"released false", json.RawMessage(`{"released": false}`), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, err := system.ParseReleaseStatus(tt.json)
			if err != nil {
				t.Fatalf("ParseReleaseStatus failed: %v", err)
			}
			if status.Released != tt.expected {
				t.Errorf("Released = %v, want %v", status.Released, tt.expected)
			}
		})
	}
}

func TestParseReleaseStatus_ErrorWhenInvalidJSON(t *testing.T) {
	invalidJSON := json.RawMessage(`{invalid}`)

	_, err := system.ParseReleaseStatus(invalidJSON)
	if err == nil {
		t.Error("Expected error when JSON is invalid, got nil")
	}
}

func TestNewReleaseStatus_ToJSON(t *testing.T) {
	tests := []struct {
		released bool
	}{
		{true},
		{false},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			status := system.NewReleaseStatus(tt.released)
			jsonData, err := status.ToJSON()
			if err != nil {
				t.Fatalf("ToJSON failed: %v", err)
			}

			// Parse back and verify
			parsed, err := system.ParseReleaseStatus(jsonData)
			if err != nil {
				t.Fatalf("ParseReleaseStatus failed: %v", err)
			}
			if parsed.Released != tt.released {
				t.Errorf("Released = %v, want %v", parsed.Released, tt.released)
			}
		})
	}
}
