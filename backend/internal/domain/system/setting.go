package system

import (
	"encoding/json"
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
)

// SettingKey はシステム設定のキー
type SettingKey string

const (
	// SettingKeyReleaseStatus はリリース状態の設定キー
	SettingKeyReleaseStatus SettingKey = "release_status"
)

// Setting はシステム設定エンティティ
type Setting struct {
	key       SettingKey
	value     json.RawMessage
	updatedAt time.Time
}

// NewSetting は新しいシステム設定を作成する
func NewSetting(key SettingKey, value json.RawMessage, now time.Time) (*Setting, error) {
	setting := &Setting{
		key:       key,
		value:     value,
		updatedAt: now,
	}
	if err := setting.validate(); err != nil {
		return nil, err
	}
	return setting, nil
}

// ReconstructSetting は永続化から設定を復元する
func ReconstructSetting(key SettingKey, value json.RawMessage, updatedAt time.Time) (*Setting, error) {
	setting := &Setting{
		key:       key,
		value:     value,
		updatedAt: updatedAt,
	}
	if err := setting.validate(); err != nil {
		return nil, err
	}
	return setting, nil
}

func (s *Setting) validate() error {
	if s.key == "" {
		return common.NewValidationError("setting key is required", nil)
	}
	if len(s.value) == 0 {
		return common.NewValidationError("setting value is required", nil)
	}
	return nil
}

// Key は設定キーを返す
func (s *Setting) Key() SettingKey {
	return s.key
}

// Value は設定値を返す
func (s *Setting) Value() json.RawMessage {
	return s.value
}

// UpdatedAt は更新日時を返す
func (s *Setting) UpdatedAt() time.Time {
	return s.updatedAt
}

// UpdateValue は設定値を更新する
func (s *Setting) UpdateValue(value json.RawMessage, now time.Time) error {
	if len(value) == 0 {
		return common.NewValidationError("setting value is required", nil)
	}
	s.value = value
	s.updatedAt = now
	return nil
}

// ReleaseStatus はリリース状態を表す値オブジェクト
type ReleaseStatus struct {
	Released bool `json:"released"`
}

// NewReleaseStatus は新しいリリース状態を作成する
func NewReleaseStatus(released bool) *ReleaseStatus {
	return &ReleaseStatus{Released: released}
}

// ToJSON はリリース状態をJSONに変換する
func (rs *ReleaseStatus) ToJSON() (json.RawMessage, error) {
	return json.Marshal(rs)
}

// ParseReleaseStatus はJSONからリリース状態をパースする
func ParseReleaseStatus(data json.RawMessage) (*ReleaseStatus, error) {
	var rs ReleaseStatus
	if err := json.Unmarshal(data, &rs); err != nil {
		return nil, common.NewValidationError("invalid release status format", err)
	}
	return &rs, nil
}
