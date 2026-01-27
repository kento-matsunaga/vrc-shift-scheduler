package system_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	appSystem "github.com/erenoa/vrc-shift-scheduler/backend/internal/app/system"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/system"
)

// mockSettingRepository is a mock implementation of system.SettingRepository
type mockSettingRepository struct {
	findByKeyFunc func(ctx context.Context, key system.SettingKey) (*system.Setting, error)
	saveFunc      func(ctx context.Context, setting *system.Setting) error
}

func (m *mockSettingRepository) FindByKey(ctx context.Context, key system.SettingKey) (*system.Setting, error) {
	if m.findByKeyFunc != nil {
		return m.findByKeyFunc(ctx, key)
	}
	return nil, common.NewNotFoundError("setting", string(key))
}

func (m *mockSettingRepository) Save(ctx context.Context, setting *system.Setting) error {
	if m.saveFunc != nil {
		return m.saveFunc(ctx, setting)
	}
	return nil
}

func TestUsecase_GetReleaseStatus_Success(t *testing.T) {
	mockRepo := &mockSettingRepository{
		findByKeyFunc: func(ctx context.Context, key system.SettingKey) (*system.Setting, error) {
			value := json.RawMessage(`{"released": true}`)
			return system.ReconstructSetting(key, value, time.Now())
		},
	}

	uc := appSystem.NewUsecase(mockRepo)
	output, err := uc.GetReleaseStatus(context.Background())
	if err != nil {
		t.Fatalf("GetReleaseStatus failed: %v", err)
	}

	if !output.Released {
		t.Error("Expected Released = true, got false")
	}
}

func TestUsecase_GetReleaseStatus_DefaultWhenNotFound(t *testing.T) {
	mockRepo := &mockSettingRepository{
		findByKeyFunc: func(ctx context.Context, key system.SettingKey) (*system.Setting, error) {
			return nil, common.NewNotFoundError("setting", string(key))
		},
	}

	uc := appSystem.NewUsecase(mockRepo)
	output, err := uc.GetReleaseStatus(context.Background())
	if err != nil {
		t.Fatalf("GetReleaseStatus failed: %v", err)
	}

	if output.Released {
		t.Error("Expected Released = false (default), got true")
	}
}

func TestUsecase_GetReleaseStatus_PropagatesDBError(t *testing.T) {
	dbError := errors.New("database connection failed")
	mockRepo := &mockSettingRepository{
		findByKeyFunc: func(ctx context.Context, key system.SettingKey) (*system.Setting, error) {
			return nil, dbError
		},
	}

	uc := appSystem.NewUsecase(mockRepo)
	_, err := uc.GetReleaseStatus(context.Background())
	if err == nil {
		t.Error("Expected error to be propagated, got nil")
	}
	if !errors.Is(err, dbError) {
		t.Errorf("Expected dbError, got %v", err)
	}
}

func TestUsecase_UpdateReleaseStatus_Create(t *testing.T) {
	var savedSetting *system.Setting
	mockRepo := &mockSettingRepository{
		findByKeyFunc: func(ctx context.Context, key system.SettingKey) (*system.Setting, error) {
			return nil, common.NewNotFoundError("setting", string(key))
		},
		saveFunc: func(ctx context.Context, setting *system.Setting) error {
			savedSetting = setting
			return nil
		},
	}

	uc := appSystem.NewUsecase(mockRepo)
	err := uc.UpdateReleaseStatus(context.Background(), appSystem.UpdateReleaseStatusInput{
		Released: true,
	})
	if err != nil {
		t.Fatalf("UpdateReleaseStatus failed: %v", err)
	}

	if savedSetting == nil {
		t.Fatal("Expected setting to be saved, got nil")
	}
	if savedSetting.Key() != system.SettingKeyReleaseStatus {
		t.Errorf("Key = %v, want %v", savedSetting.Key(), system.SettingKeyReleaseStatus)
	}
}

func TestUsecase_UpdateReleaseStatus_Update(t *testing.T) {
	existingSetting, _ := system.NewSetting(
		system.SettingKeyReleaseStatus,
		json.RawMessage(`{"released": false}`),
		time.Now(),
	)

	var savedSetting *system.Setting
	mockRepo := &mockSettingRepository{
		findByKeyFunc: func(ctx context.Context, key system.SettingKey) (*system.Setting, error) {
			return existingSetting, nil
		},
		saveFunc: func(ctx context.Context, setting *system.Setting) error {
			savedSetting = setting
			return nil
		},
	}

	uc := appSystem.NewUsecase(mockRepo)
	err := uc.UpdateReleaseStatus(context.Background(), appSystem.UpdateReleaseStatusInput{
		Released: true,
	})
	if err != nil {
		t.Fatalf("UpdateReleaseStatus failed: %v", err)
	}

	if savedSetting == nil {
		t.Fatal("Expected setting to be saved, got nil")
	}

	// Verify the value was updated
	status, err := system.ParseReleaseStatus(savedSetting.Value())
	if err != nil {
		t.Fatalf("ParseReleaseStatus failed: %v", err)
	}
	if !status.Released {
		t.Error("Expected Released = true after update, got false")
	}
}

func TestUsecase_UpdateReleaseStatus_PropagatesDBError(t *testing.T) {
	dbError := errors.New("database connection failed")
	mockRepo := &mockSettingRepository{
		findByKeyFunc: func(ctx context.Context, key system.SettingKey) (*system.Setting, error) {
			return nil, dbError
		},
	}

	uc := appSystem.NewUsecase(mockRepo)
	err := uc.UpdateReleaseStatus(context.Background(), appSystem.UpdateReleaseStatusInput{
		Released: true,
	})
	if err == nil {
		t.Error("Expected error to be propagated, got nil")
	}
	if !errors.Is(err, dbError) {
		t.Errorf("Expected dbError, got %v", err)
	}
}
