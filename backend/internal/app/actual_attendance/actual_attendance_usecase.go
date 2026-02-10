package actual_attendance

import (
	"context"
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/event"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/member"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/shift"
)

// TargetDateInfo represents a business day in the actual attendance response
type TargetDateInfo struct {
	TargetDateID string    `json:"target_date_id"` // business_day_id
	TargetDate   time.Time `json:"target_date"`    // 営業日の日付
	DisplayOrder int       `json:"display_order"`
}

// MemberAttendanceStatus represents a member's attendance status across multiple dates
type MemberAttendanceStatus struct {
	MemberID      string            `json:"member_id"`
	MemberName    string            `json:"member_name"`
	AttendanceMap map[string]string `json:"attendance_map"` // target_date_id -> "attended" | "absent"
}

// GetRecentActualAttendanceInput represents the input for getting recent actual attendance
type GetRecentActualAttendanceInput struct {
	TenantID      common.TenantID
	EventID       *common.EventID // オプション: 指定された場合、このイベントの営業日のみフィルタリング
	Limit         int
	IncludeFuture bool // trueの場合、未来の営業日も含める
}

// GetRecentActualAttendanceOutput represents the output for getting recent actual attendance
type GetRecentActualAttendanceOutput struct {
	TargetDates       []TargetDateInfo
	MemberAttendances []MemberAttendanceStatus
}

// GetRecentActualAttendanceUsecase handles getting recent actual attendance data
type GetRecentActualAttendanceUsecase struct {
	businessDayRepo event.EventBusinessDayRepository
	memberRepo      member.MemberRepository
	assignmentRepo  shift.ShiftAssignmentRepository
}

// NewGetRecentActualAttendanceUsecase creates a new GetRecentActualAttendanceUsecase
func NewGetRecentActualAttendanceUsecase(
	businessDayRepo event.EventBusinessDayRepository,
	memberRepo member.MemberRepository,
	assignmentRepo shift.ShiftAssignmentRepository,
) *GetRecentActualAttendanceUsecase {
	return &GetRecentActualAttendanceUsecase{
		businessDayRepo: businessDayRepo,
		memberRepo:      memberRepo,
		assignmentRepo:  assignmentRepo,
	}
}

// Execute retrieves recent actual attendance data based on shift assignments
//
// Implementation logic:
//  1. Get recent N business days (past only, oldest first)
//  2. Get all active members
//  3. For each member and each business day, check if there are shift assignments
//  4. Assignment exists → "attended", no assignment → "absent"
func (uc *GetRecentActualAttendanceUsecase) Execute(
	ctx context.Context,
	input GetRecentActualAttendanceInput,
) (*GetRecentActualAttendanceOutput, error) {
	limit := input.Limit
	if limit <= 0 {
		limit = 10 // デフォルト
	}

	// 1. Get recent N business days
	var businessDays []*event.EventBusinessDay
	var err error
	if input.EventID != nil {
		// イベントIDが指定された場合、そのイベントの営業日のみ取得
		businessDays, err = uc.businessDayRepo.FindRecentByEventID(ctx, input.TenantID, *input.EventID, limit, input.IncludeFuture)
	} else {
		// 全営業日から取得（過去のみ）
		businessDays, err = uc.businessDayRepo.FindRecentByTenantID(ctx, input.TenantID, limit)
	}
	if err != nil {
		return nil, err
	}

	var targetDates []TargetDateInfo
	for i, bd := range businessDays {
		targetDates = append(targetDates, TargetDateInfo{
			TargetDateID: string(bd.BusinessDayID()),
			TargetDate:   bd.TargetDate(),
			DisplayOrder: i + 1,
		})
	}

	// 2. Get all active members
	activeMembers, err := uc.memberRepo.FindActiveByTenantID(ctx, input.TenantID)
	if err != nil {
		return nil, err
	}

	// 3. Build actual attendance data for each member
	var memberAttendances []MemberAttendanceStatus
	for _, m := range activeMembers {
		attendanceMap := make(map[string]string)

		// Check shift assignments for each business day
		for _, td := range targetDates {
			// Check if this member is assigned to any shift on this business day
			hasAssignment, err := uc.assignmentRepo.HasConfirmedByMemberAndBusinessDayID(
				ctx,
				input.TenantID,
				m.MemberID(),
				event.BusinessDayID(td.TargetDateID),
			)
			if err != nil {
				return nil, err
			}

			// Assignment exists → "attended", no assignment → "absent"
			if hasAssignment {
				attendanceMap[td.TargetDateID] = "attended"
			} else {
				attendanceMap[td.TargetDateID] = "absent"
			}
		}

		memberAttendances = append(memberAttendances, MemberAttendanceStatus{
			MemberID:      m.MemberID().String(),
			MemberName:    m.DisplayName(),
			AttendanceMap: attendanceMap,
		})
	}

	return &GetRecentActualAttendanceOutput{
		TargetDates:       targetDates,
		MemberAttendances: memberAttendances,
	}, nil
}
