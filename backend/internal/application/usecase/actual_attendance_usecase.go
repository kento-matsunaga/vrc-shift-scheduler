package usecase

import (
	"context"
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/jackc/pgx/v5/pgxpool"
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
	TenantID common.TenantID
	Limit    int
}

// GetRecentActualAttendanceOutput represents the output for getting recent actual attendance
type GetRecentActualAttendanceOutput struct {
	TargetDates       []TargetDateInfo
	MemberAttendances []MemberAttendanceStatus
}

// GetRecentActualAttendanceUsecase handles getting recent actual attendance data
type GetRecentActualAttendanceUsecase struct {
	dbPool *pgxpool.Pool
}

// NewGetRecentActualAttendanceUsecase creates a new GetRecentActualAttendanceUsecase
func NewGetRecentActualAttendanceUsecase(dbPool *pgxpool.Pool) *GetRecentActualAttendanceUsecase {
	return &GetRecentActualAttendanceUsecase{
		dbPool: dbPool,
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
		limit = 30 // デフォルト
	}

	// 1. Get recent N business days (past only, oldest first)
	targetDatesQuery := `
		SELECT
			business_day_id,
			target_date
		FROM event_business_days
		WHERE tenant_id = $1
		  AND deleted_at IS NULL
		  AND target_date <= CURRENT_DATE
		ORDER BY target_date ASC
		LIMIT $2
	`

	rows, err := uc.dbPool.Query(ctx, targetDatesQuery, input.TenantID.String(), limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var targetDates []TargetDateInfo
	for rows.Next() {
		var td TargetDateInfo
		if err := rows.Scan(&td.TargetDateID, &td.TargetDate); err != nil {
			return nil, err
		}
		targetDates = append(targetDates, td)
	}

	// Set display order (oldest first: 1, 2, 3...)
	for i := range targetDates {
		targetDates[i].DisplayOrder = i + 1
	}

	// 2. Get all active members
	membersQuery := `
		SELECT
			member_id,
			display_name
		FROM members
		WHERE tenant_id = $1
		  AND is_active = true
		  AND deleted_at IS NULL
		ORDER BY display_name
	`

	memberRows, err := uc.dbPool.Query(ctx, membersQuery, input.TenantID.String())
	if err != nil {
		return nil, err
	}
	defer memberRows.Close()

	var members []struct {
		MemberID   string
		MemberName string
	}
	for memberRows.Next() {
		var m struct {
			MemberID   string
			MemberName string
		}
		if err := memberRows.Scan(&m.MemberID, &m.MemberName); err != nil {
			return nil, err
		}
		members = append(members, m)
	}

	// 3. Build actual attendance data for each member
	var memberAttendances []MemberAttendanceStatus
	for _, member := range members {
		attendanceMap := make(map[string]string)

		// Check shift assignments for each business day
		for _, td := range targetDates {
			// Check if this member is assigned to any shift on this business day
			// JOIN: business_day_id -> shift_slots -> shift_assignments
			attendedQuery := `
				SELECT COUNT(*)
				FROM shift_assignments sa
				INNER JOIN shift_slots ss ON sa.slot_id = ss.slot_id AND ss.deleted_at IS NULL
				WHERE sa.tenant_id = $1
				  AND sa.member_id = $2
				  AND ss.business_day_id = $3
				  AND sa.assignment_status = 'confirmed'
				  AND sa.deleted_at IS NULL
			`

			var count int
			err := uc.dbPool.QueryRow(ctx, attendedQuery, input.TenantID.String(), member.MemberID, td.TargetDateID).Scan(&count)
			if err != nil {
				return nil, err
			}

			// Assignment exists → "attended", no assignment → "absent"
			if count > 0 {
				attendanceMap[td.TargetDateID] = "attended"
			} else {
				attendanceMap[td.TargetDateID] = "absent"
			}
		}

		memberAttendances = append(memberAttendances, MemberAttendanceStatus{
			MemberID:      member.MemberID,
			MemberName:    member.MemberName,
			AttendanceMap: attendanceMap,
		})
	}

	return &GetRecentActualAttendanceOutput{
		TargetDates:       targetDates,
		MemberAttendances: memberAttendances,
	}, nil
}
