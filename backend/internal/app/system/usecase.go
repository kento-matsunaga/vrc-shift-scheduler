package system

import (
	"context"
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/system"
)

// Usecase はシステム設定のユースケース
type Usecase struct {
	settingRepo system.SettingRepository
}

// NewUsecase は新しいシステム設定ユースケースを作成する
func NewUsecase(settingRepo system.SettingRepository) *Usecase {
	return &Usecase{
		settingRepo: settingRepo,
	}
}

// GetReleaseStatusOutput はリリース状態取得の出力
type GetReleaseStatusOutput struct {
	Released bool
}

// GetReleaseStatus はリリース状態を取得する
func (uc *Usecase) GetReleaseStatus(ctx context.Context) (*GetReleaseStatusOutput, error) {
	setting, err := uc.settingRepo.FindByKey(ctx, system.SettingKeyReleaseStatus)
	if err != nil {
		// 設定が見つからない場合はデフォルト（リリース前）を返す
		if common.IsNotFoundError(err) {
			return &GetReleaseStatusOutput{Released: false}, nil
		}
		// その他のエラー（DBエラーなど）は呼び出し元に伝播
		return nil, err
	}

	status, err := system.ParseReleaseStatus(setting.Value())
	if err != nil {
		return nil, err
	}

	return &GetReleaseStatusOutput{Released: status.Released}, nil
}

// UpdateReleaseStatusInput はリリース状態更新の入力
type UpdateReleaseStatusInput struct {
	Released bool
}

// UpdateReleaseStatus はリリース状態を更新する
func (uc *Usecase) UpdateReleaseStatus(ctx context.Context, input UpdateReleaseStatusInput) error {
	now := time.Now()

	status := system.NewReleaseStatus(input.Released)
	value, err := status.ToJSON()
	if err != nil {
		return err
	}

	// 既存の設定を取得するか、新規作成
	setting, err := uc.settingRepo.FindByKey(ctx, system.SettingKeyReleaseStatus)
	if err != nil {
		if common.IsNotFoundError(err) {
			// 設定が存在しない場合は新規作成
			setting, err = system.NewSetting(system.SettingKeyReleaseStatus, value, now)
			if err != nil {
				return err
			}
		} else {
			// その他のエラー（DBエラーなど）は呼び出し元に伝播
			return err
		}
	} else {
		// 既存の設定を更新
		if err := setting.UpdateValue(value, now); err != nil {
			return err
		}
	}

	return uc.settingRepo.Save(ctx, setting)
}
