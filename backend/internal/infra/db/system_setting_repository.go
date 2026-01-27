package db

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/system"
)

// SystemSettingRepository implements system.SettingRepository for PostgreSQL
type SystemSettingRepository struct {
	db *pgxpool.Pool
}

// NewSystemSettingRepository は新しいシステム設定リポジトリを作成する
func NewSystemSettingRepository(db *pgxpool.Pool) *SystemSettingRepository {
	return &SystemSettingRepository{db: db}
}

func (r *SystemSettingRepository) FindByKey(ctx context.Context, key system.SettingKey) (*system.Setting, error) {
	query := `
		SELECT key, value, updated_at
		FROM system_settings
		WHERE key = $1
	`

	var (
		keyStr    string
		value     json.RawMessage
		updatedAt time.Time
	)

	err := r.db.QueryRow(ctx, query, string(key)).Scan(&keyStr, &value, &updatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, common.NewDomainError(common.ErrNotFound, "setting not found")
		}
		return nil, err
	}

	return system.ReconstructSetting(system.SettingKey(keyStr), value, updatedAt)
}

func (r *SystemSettingRepository) Save(ctx context.Context, setting *system.Setting) error {
	query := `
		INSERT INTO system_settings (key, value, updated_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (key) DO UPDATE SET
			value = EXCLUDED.value,
			updated_at = EXCLUDED.updated_at
	`

	_, err := r.db.Exec(ctx, query,
		string(setting.Key()),
		setting.Value(),
		setting.UpdatedAt(),
	)
	if err != nil {
		return err
	}

	return nil
}
