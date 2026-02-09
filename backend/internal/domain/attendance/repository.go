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

	// FindResponsesByMemberID は member の回答一覧を取得する（出席率計算用）
	FindResponsesByMemberID(ctx context.Context, tenantID common.TenantID, memberID common.MemberID) ([]*AttendanceResponse, error)

	// FindResponsesByCollectionIDAndMemberID は collection 内の特定 member の回答一覧を取得する（公開ページ用）
	// tenant_id でスコープすることでクロステナントアクセスを防止
	FindResponsesByCollectionIDAndMemberID(ctx context.Context, tenantID common.TenantID, collectionID common.CollectionID, memberID common.MemberID) ([]*AttendanceResponse, error)

	// SaveTargetDates は対象日を保存する（全置換: DELETE ALL + INSERT ALL）
	SaveTargetDates(ctx context.Context, collectionID common.CollectionID, targetDates []*TargetDate) error

	// ReplaceTargetDates は対象日を差分更新する（既存IDの回答を保持）
	// targetDates 内の既存IDは UPDATE、新規IDは INSERT、
	// targetDates に含まれない既存IDは DELETE（CASCADE で回答も削除）
	ReplaceTargetDates(ctx context.Context, collectionID common.CollectionID, targetDates []*TargetDate) error

	// FindTargetDatesByCollectionID は collection の対象日一覧を取得する
	FindTargetDatesByCollectionID(ctx context.Context, collectionID common.CollectionID) ([]*TargetDate, error)

	// SaveGroupAssignments はグループ割り当てを保存する（既存のものを全て削除してから保存）
	SaveGroupAssignments(ctx context.Context, collectionID common.CollectionID, assignments []*CollectionGroupAssignment) error

	// FindGroupAssignmentsByCollectionID は collection のグループ割り当て一覧を取得する
	FindGroupAssignmentsByCollectionID(ctx context.Context, collectionID common.CollectionID) ([]*CollectionGroupAssignment, error)

	// SaveRoleAssignments はロール割り当てを保存する（既存のものを全て削除してから保存）
	SaveRoleAssignments(ctx context.Context, collectionID common.CollectionID, assignments []*CollectionRoleAssignment) error

	// FindRoleAssignmentsByCollectionID は collection のロール割り当て一覧を取得する
	FindRoleAssignmentsByCollectionID(ctx context.Context, collectionID common.CollectionID) ([]*CollectionRoleAssignment, error)
}
