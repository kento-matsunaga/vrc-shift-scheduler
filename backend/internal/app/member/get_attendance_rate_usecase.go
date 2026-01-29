package member

import (
	"context"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/attendance"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/member"
)

// GetAttendanceRateUsecase handles getting attendance rates for members
type GetAttendanceRateUsecase struct {
	memberRepo     member.MemberRepository
	attendanceRepo attendance.AttendanceCollectionRepository
}

// NewGetAttendanceRateUsecase creates a new GetAttendanceRateUsecase
func NewGetAttendanceRateUsecase(
	memberRepo member.MemberRepository,
	attendanceRepo attendance.AttendanceCollectionRepository,
) *GetAttendanceRateUsecase {
	return &GetAttendanceRateUsecase{
		memberRepo:     memberRepo,
		attendanceRepo: attendanceRepo,
	}
}

// MemberAttendanceRate represents attendance rate for a member
type MemberAttendanceRate struct {
	MemberID       string  `json:"member_id"`
	TotalResponses int     `json:"total_responses"` // 総回答数
	AttendingCount int     `json:"attending_count"` // 参加回答数
	AttendanceRate float64 `json:"attendance_rate"` // 出席率（0-100）
}

// GetAttendanceRatesInput represents the input for getting attendance rates
type GetAttendanceRatesInput struct {
	TenantID string
}

// GetAttendanceRatesOutput represents the output for getting attendance rates
type GetAttendanceRatesOutput struct {
	Rates []MemberAttendanceRate `json:"rates"`
}

// Execute executes the get attendance rates use case
func (u *GetAttendanceRateUsecase) Execute(ctx context.Context, input GetAttendanceRatesInput) (*GetAttendanceRatesOutput, error) {
	// 1. Parse TenantID
	tenantID, err := common.ParseTenantID(input.TenantID)
	if err != nil {
		return nil, err
	}

	// 2. Get all active members
	members, err := u.memberRepo.FindActiveByTenantID(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	// 3. Calculate attendance rate for each member
	rates := make([]MemberAttendanceRate, 0, len(members))
	for _, m := range members {
		// Get all responses for this member
		responses, err := u.attendanceRepo.FindResponsesByMemberID(ctx, tenantID, m.MemberID())
		if err != nil {
			// If error, skip this member or return 0 rate
			rates = append(rates, MemberAttendanceRate{
				MemberID:       m.MemberID().String(),
				TotalResponses: 0,
				AttendingCount: 0,
				AttendanceRate: 0.0,
			})
			continue
		}

		// Count attending responses
		totalResponses := len(responses)
		attendingCount := 0
		for _, resp := range responses {
			if resp.Response().String() == "attending" {
				attendingCount++
			}
		}

		// Calculate rate
		var attendanceRate float64
		if totalResponses > 0 {
			attendanceRate = float64(attendingCount) / float64(totalResponses) * 100.0
		}

		rates = append(rates, MemberAttendanceRate{
			MemberID:       m.MemberID().String(),
			TotalResponses: totalResponses,
			AttendingCount: attendingCount,
			AttendanceRate: attendanceRate,
		})
	}

	return &GetAttendanceRatesOutput{
		Rates: rates,
	}, nil
}
