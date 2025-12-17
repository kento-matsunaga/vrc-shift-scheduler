package member

import (
	"context"
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/attendance"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/member"
)

// GetRecentAttendanceUsecase handles getting recent attendance status for all members
type GetRecentAttendanceUsecase struct {
	memberRepo     member.MemberRepository
	attendanceRepo attendance.AttendanceCollectionRepository
}

// NewGetRecentAttendanceUsecase creates a new GetRecentAttendanceUsecase
func NewGetRecentAttendanceUsecase(
	memberRepo member.MemberRepository,
	attendanceRepo attendance.AttendanceCollectionRepository,
) *GetRecentAttendanceUsecase {
	return &GetRecentAttendanceUsecase{
		memberRepo:     memberRepo,
		attendanceRepo: attendanceRepo,
	}
}

// MemberAttendanceStatus represents attendance status for a member
type MemberAttendanceStatus struct {
	MemberID      string            `json:"member_id"`
	MemberName    string            `json:"member_name"`
	AttendanceMap map[string]string `json:"attendance_map"` // target_date_id -> "attending" | "absent" | ""
}

// GetRecentAttendanceInput represents the input for getting recent attendance
type GetRecentAttendanceInput struct {
	TenantID string
	Limit    int // Number of recent target dates to fetch (default 10)
}

// GetRecentAttendanceOutput represents the output for getting recent attendance
type GetRecentAttendanceOutput struct {
	TargetDates       []TargetDateInfo         `json:"target_dates"`        // Recent target dates (newest first)
	MemberAttendances []MemberAttendanceStatus `json:"member_attendances"` // Attendance status per member
}

// TargetDateInfo represents a target date with its information
type TargetDateInfo struct {
	TargetDateID string    `json:"target_date_id"`
	TargetDate   time.Time `json:"target_date"`
	DisplayOrder int       `json:"display_order"`
}

// Execute executes the get recent attendance use case
func (u *GetRecentAttendanceUsecase) Execute(ctx context.Context, input GetRecentAttendanceInput) (*GetRecentAttendanceOutput, error) {
	// 1. Parse TenantID
	tenantID, err := common.ParseTenantID(input.TenantID)
	if err != nil {
		return nil, err
	}

	limit := input.Limit
	if limit <= 0 {
		limit = 10
	}

	// 2. Get all collections for this tenant
	collections, err := u.attendanceRepo.FindByTenantID(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	// 3. Collect all target dates from all collections
	type dateWithCollection struct {
		TargetDateID   string
		TargetDate     time.Time
		DisplayOrder   int
		CollectionID   string
	}

	var allDates []dateWithCollection
	for _, col := range collections {
		targetDates, err := u.attendanceRepo.FindTargetDatesByCollectionID(ctx, col.CollectionID())
		if err != nil {
			continue // Skip if error
		}
		for _, td := range targetDates {
			allDates = append(allDates, dateWithCollection{
				TargetDateID:   td.TargetDateID().String(),
				TargetDate:     td.TargetDateValue(),
				DisplayOrder:   td.DisplayOrder(),
				CollectionID:   col.CollectionID().String(),
			})
		}
	}

	// 4. Sort by date (newest first) and take the most recent N dates
	// Simple bubble sort for now
	for i := 0; i < len(allDates)-1; i++ {
		for j := i + 1; j < len(allDates); j++ {
			if allDates[i].TargetDate.Before(allDates[j].TargetDate) {
				allDates[i], allDates[j] = allDates[j], allDates[i]
			}
		}
	}

	// Take only the most recent dates up to limit
	if len(allDates) > limit {
		allDates = allDates[:limit]
	}

	// Convert to TargetDateInfo
	targetDateInfos := make([]TargetDateInfo, 0, len(allDates))
	targetDateIDs := make(map[string]bool)
	for _, d := range allDates {
		targetDateInfos = append(targetDateInfos, TargetDateInfo{
			TargetDateID: d.TargetDateID,
			TargetDate:   d.TargetDate,
			DisplayOrder: d.DisplayOrder,
		})
		targetDateIDs[d.TargetDateID] = true
	}

	// 5. Get all active members
	members, err := u.memberRepo.FindActiveByTenantID(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	// 6. For each member, get their responses and build attendance map
	memberAttendances := make([]MemberAttendanceStatus, 0, len(members))
	for _, m := range members {
		// Get all responses for this member
		responses, err := u.attendanceRepo.FindResponsesByMemberID(ctx, tenantID, m.MemberID())
		if err != nil {
			// If error, create empty map
			memberAttendances = append(memberAttendances, MemberAttendanceStatus{
				MemberID:      m.MemberID().String(),
				MemberName:    m.DisplayName(),
				AttendanceMap: make(map[string]string),
			})
			continue
		}

		// Build attendance map for the recent target dates only
		attendanceMap := make(map[string]string)
		for _, resp := range responses {
			targetDateID := resp.TargetDateID().String()
			// Only include if this target_date_id is in our recent dates
			if targetDateIDs[targetDateID] {
				attendanceMap[targetDateID] = resp.Response().String()
			}
		}

		memberAttendances = append(memberAttendances, MemberAttendanceStatus{
			MemberID:      m.MemberID().String(),
			MemberName:    m.DisplayName(),
			AttendanceMap: attendanceMap,
		})
	}

	return &GetRecentAttendanceOutput{
		TargetDates:       targetDateInfos,
		MemberAttendances: memberAttendances,
	}, nil
}
