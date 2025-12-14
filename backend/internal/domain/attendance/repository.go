package attendance

import (
	"context"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
)

// AttendanceCollectionRepository defines the interface for AttendanceCollection persistence
type AttendanceCollectionRepository interface {
	// Save saves a collection (insert or update)
	Save(ctx context.Context, collection *AttendanceCollection) error

	// FindByID finds a collection by ID within a tenant
	FindByID(ctx context.Context, tenantID common.TenantID, id common.CollectionID) (*AttendanceCollection, error)

	// FindByToken finds a collection by public token
	FindByToken(ctx context.Context, token common.PublicToken) (*AttendanceCollection, error)

	// FindByTenantID finds all collections within a tenant
	// deleted_at IS NULL のレコードのみ返す
	FindByTenantID(ctx context.Context, tenantID common.TenantID) ([]*AttendanceCollection, error)

	// UpsertResponse は回答を登録/更新する（ON CONFLICT DO UPDATE）
	// MVP方針: 回答の上書きはRepository層で行う
	UpsertResponse(ctx context.Context, response *AttendanceResponse) error

	// FindResponsesByCollectionID は collection の回答一覧を取得する
	FindResponsesByCollectionID(ctx context.Context, collectionID common.CollectionID) ([]*AttendanceResponse, error)
}
