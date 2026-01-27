package system

import "context"

// SettingRepository はシステム設定リポジトリのインターフェース
type SettingRepository interface {
	// FindByKey は指定されたキーの設定を取得する
	FindByKey(ctx context.Context, key SettingKey) (*Setting, error)

	// Save は設定を保存する（存在しない場合は作成、存在する場合は更新）
	Save(ctx context.Context, setting *Setting) error
}
