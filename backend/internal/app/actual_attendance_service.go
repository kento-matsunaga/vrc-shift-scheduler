package app

import (
	"context"
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ActualAttendanceService は本出席（実際のシフト割り当て）を集計するサービス
// これはShiftAssignmentの読み取り専用ビューとして機能する
type ActualAttendanceService struct {
	dbPool *pgxpool.Pool
}

// NewActualAttendanceService creates a new ActualAttendanceService
func NewActualAttendanceService(dbPool *pgxpool.Pool) *ActualAttendanceService {
	return &ActualAttendanceService{
		dbPool: dbPool,
	}
}

// TargetDateInfo represents a business day in the actual attendance response
type TargetDateInfo struct {
	TargetDateID   string    `json:"target_date_id"`   // business_day_id
	TargetDate     time.Time `json:"target_date"`      // 営業日の日付
	DisplayOrder   int       `json:"display_order"`
}

// MemberAttendanceStatus represents a member's attendance status across multiple dates
type MemberAttendanceStatus struct {
	MemberID      string            `json:"member_id"`
	MemberName    string            `json:"member_name"`
	AttendanceMap map[string]string `json:"attendance_map"` // target_date_id -> "attended" | "absent"
}

// ActualAttendanceResponse represents the actual attendance data
type ActualAttendanceResponse struct {
	TargetDates        []TargetDateInfo         `json:"target_dates"`
	MemberAttendances  []MemberAttendanceStatus `json:"member_attendances"`
}

// GetRecentActualAttendance は直近の本出席データを取得する
// これは実際のシフト割り当て（shift_assignments）に基づく実績データ
//
// 実装ロジック:
//  1. 直近N日分の営業日を取得（最新順）
//  2. 全アクティブメンバーを取得
//  3. 各メンバー・各営業日について、シフト割り当てがあるか確認
//  4. 割り当てあり → "attended"、割り当てなし → "absent"
func (s *ActualAttendanceService) GetRecentActualAttendance(
	ctx context.Context,
	tenantID common.TenantID,
	limit int,
) (*ActualAttendanceResponse, error) {
	if limit <= 0 {
		limit = 30 // デフォルト
	}

	// 1. 直近N日分の営業日を取得（過去のみ、古い順）
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

	rows, err := s.dbPool.Query(ctx, targetDatesQuery, tenantID.String(), limit)
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

	// displayOrderを設定（古い順に1, 2, 3...）
	for i := range targetDates {
		targetDates[i].DisplayOrder = i + 1
	}

	// 2. 全アクティブメンバーを取得
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

	memberRows, err := s.dbPool.Query(ctx, membersQuery, tenantID.String())
	if err != nil {
		return nil, err
	}
	defer memberRows.Close()

	var members []struct {
		MemberID    string
		MemberName  string
	}
	for memberRows.Next() {
		var m struct {
			MemberID    string
			MemberName  string
		}
		if err := memberRows.Scan(&m.MemberID, &m.MemberName); err != nil {
			return nil, err
		}
		members = append(members, m)
	}

	// 3. 各メンバーの本出席データを構築
	var memberAttendances []MemberAttendanceStatus
	for _, member := range members {
		attendanceMap := make(map[string]string)

		// 各営業日について、シフト割り当てがあるか確認
		for _, td := range targetDates {
			// このメンバーがこの営業日のシフトに割り当てられているか確認
			// business_day_id -> shift_slots -> shift_assignments を JOIN
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
			err := s.dbPool.QueryRow(ctx, attendedQuery, tenantID.String(), member.MemberID, td.TargetDateID).Scan(&count)
			if err != nil {
				return nil, err
			}

			// シフト割り当てがある → "attended"、ない → "absent"
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

	return &ActualAttendanceResponse{
		TargetDates:       targetDates,
		MemberAttendances: memberAttendances,
	}, nil
}
