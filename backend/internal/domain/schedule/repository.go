package schedule

import (
	"context"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
)

// DateScheduleRepository defines the interface for DateSchedule persistence
type DateScheduleRepository interface {
	// Save saves a schedule (insert or update)
	Save(ctx context.Context, schedule *DateSchedule) error

	// FindByID finds a schedule by ID within a tenant
	FindByID(ctx context.Context, tenantID common.TenantID, id common.ScheduleID) (*DateSchedule, error)

	// FindByToken finds a schedule by public token
	FindByToken(ctx context.Context, token common.PublicToken) (*DateSchedule, error)

	// FindByTenantID finds all schedules within a tenant
	FindByTenantID(ctx context.Context, tenantID common.TenantID) ([]*DateSchedule, error)

	// UpsertResponse は回答を登録/更新する（ON CONFLICT DO UPDATE）
	// MVP方針: 回答の上書きはRepository層で行う
	UpsertResponse(ctx context.Context, response *DateScheduleResponse) error

	// FindResponsesByScheduleID は schedule の回答一覧を取得する
	FindResponsesByScheduleID(ctx context.Context, scheduleID common.ScheduleID) ([]*DateScheduleResponse, error)

	// FindCandidatesByScheduleID は schedule の候補日一覧を取得する
	FindCandidatesByScheduleID(ctx context.Context, scheduleID common.ScheduleID) ([]*CandidateDate, error)
}
