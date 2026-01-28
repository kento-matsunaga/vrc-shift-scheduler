package schedule

import (
	"context"
	"fmt"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/attendance"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/member"
	schedDomain "github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/schedule"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/services"
)

// ConvertToAttendanceUsecase handles converting a schedule to an attendance collection
type ConvertToAttendanceUsecase struct {
	scheduleRepo    schedDomain.DateScheduleRepository
	attendanceRepo  attendance.AttendanceCollectionRepository
	memberGroupRepo member.MemberGroupRepository
	txManager       services.TxManager
	clock           services.Clock
}

// NewConvertToAttendanceUsecase creates a new ConvertToAttendanceUsecase
func NewConvertToAttendanceUsecase(
	scheduleRepo schedDomain.DateScheduleRepository,
	attendanceRepo attendance.AttendanceCollectionRepository,
	memberGroupRepo member.MemberGroupRepository,
	txManager services.TxManager,
	clock services.Clock,
) *ConvertToAttendanceUsecase {
	return &ConvertToAttendanceUsecase{
		scheduleRepo:    scheduleRepo,
		attendanceRepo:  attendanceRepo,
		memberGroupRepo: memberGroupRepo,
		txManager:       txManager,
		clock:           clock,
	}
}

// Execute executes the convert to attendance use case
//
// 変換処理の概要：
// 1. 日程調整の候補日 → 出欠確認の対象日
// 2. グループ割り当てをコピー
// 3. 回答を変換（available→attending, unavailable→absent, maybe→undecided）
// 4. 未回答メンバーはundecidedとして追加
// 5. 元の回答時刻を保持
func (u *ConvertToAttendanceUsecase) Execute(ctx context.Context, input ConvertToAttendanceInput) (*ConvertToAttendanceOutput, error) {
	// 1. Parse and validate IDs
	tenantID, err := common.ParseTenantID(input.TenantID)
	if err != nil {
		return nil, common.NewValidationError("invalid tenant_id", err)
	}

	scheduleID, err := common.ParseScheduleID(input.ScheduleID)
	if err != nil {
		return nil, common.NewValidationError("invalid schedule_id", err)
	}

	if len(input.CandidateIDs) == 0 {
		return nil, common.NewValidationError("at least one candidate_id is required", nil)
	}

	candidateIDs := make([]common.CandidateID, len(input.CandidateIDs))
	for i, cidStr := range input.CandidateIDs {
		cid, err := common.ParseCandidateID(cidStr)
		if err != nil {
			return nil, common.NewValidationError(fmt.Sprintf("invalid candidate_id: %s", cidStr), err)
		}
		candidateIDs[i] = cid
	}

	// 2. Get schedule with candidates
	schedule, err := u.scheduleRepo.FindByID(ctx, tenantID, scheduleID)
	if err != nil {
		if common.IsNotFoundError(err) {
			return nil, common.NewNotFoundError("schedule", input.ScheduleID)
		}
		return nil, fmt.Errorf("failed to get schedule: %w", err)
	}

	// 3. Validate candidate IDs exist in schedule
	candidateMap := make(map[string]*schedDomain.CandidateDate)
	for _, c := range schedule.Candidates() {
		candidateMap[c.CandidateID().String()] = c
	}

	selectedCandidates := make([]*schedDomain.CandidateDate, 0, len(candidateIDs))
	for _, cid := range candidateIDs {
		c, ok := candidateMap[cid.String()]
		if !ok {
			return nil, common.NewValidationError(fmt.Sprintf("candidate_id not found in schedule: %s", cid.String()), nil)
		}
		selectedCandidates = append(selectedCandidates, c)
	}

	// 4. Get schedule group assignments
	groupAssignments, err := u.scheduleRepo.FindGroupAssignmentsByScheduleID(ctx, scheduleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get group assignments: %w", err)
	}

	// 5. Get schedule responses
	scheduleResponses, err := u.scheduleRepo.FindResponsesByScheduleID(ctx, scheduleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get schedule responses: %w", err)
	}

	// 6. Determine title
	title := input.Title
	if title == "" {
		title = schedule.Title()
	}

	now := u.clock.Now()

	// 7. Execute in transaction
	var output *ConvertToAttendanceOutput
	err = u.txManager.WithTx(ctx, func(ctx context.Context) error {
		// 7.1 Create AttendanceCollection
		targetType, err := attendance.NewTargetType("event")
		if err != nil {
			return fmt.Errorf("failed to create target type: %w", err)
		}
		var targetID string
		if schedule.EventID() != nil {
			targetID = schedule.EventID().String()
		}

		collection, err := attendance.NewAttendanceCollection(
			now,
			tenantID,
			title,
			schedule.Description(),
			targetType,
			targetID,
			schedule.Deadline(),
		)
		if err != nil {
			return fmt.Errorf("failed to create attendance collection: %w", err)
		}

		if err := u.attendanceRepo.Save(ctx, collection); err != nil {
			return fmt.Errorf("failed to save attendance collection: %w", err)
		}

		// 7.2 Create TargetDates from selected candidates
		targetDates := make([]*attendance.TargetDate, 0, len(selectedCandidates))
		candidateToTargetDate := make(map[string]common.TargetDateID)

		for i, candidate := range selectedCandidates {
			var startTime, endTime *string
			if candidate.StartTime() != nil {
				st := candidate.StartTime().Format("15:04")
				startTime = &st
			}
			if candidate.EndTime() != nil {
				et := candidate.EndTime().Format("15:04")
				endTime = &et
			}

			td, err := attendance.NewTargetDate(
				now,
				collection.CollectionID(),
				candidate.CandidateDateValue(),
				startTime,
				endTime,
				i,
			)
			if err != nil {
				return fmt.Errorf("failed to create target date: %w", err)
			}
			targetDates = append(targetDates, td)
			candidateToTargetDate[candidate.CandidateID().String()] = td.TargetDateID()
		}

		if err := u.attendanceRepo.SaveTargetDates(ctx, collection.CollectionID(), targetDates); err != nil {
			return fmt.Errorf("failed to save target dates: %w", err)
		}

		// 7.3 Copy group assignments
		if len(groupAssignments) > 0 {
			attendanceGroupAssignments := make([]*attendance.CollectionGroupAssignment, 0, len(groupAssignments))
			for _, ga := range groupAssignments {
				assignment, err := attendance.NewCollectionGroupAssignment(now, collection.CollectionID(), ga.GroupID())
				if err != nil {
					return fmt.Errorf("failed to create group assignment: %w", err)
				}
				attendanceGroupAssignments = append(attendanceGroupAssignments, assignment)
			}

			if err := u.attendanceRepo.SaveGroupAssignments(ctx, collection.CollectionID(), attendanceGroupAssignments); err != nil {
				return fmt.Errorf("failed to save group assignments: %w", err)
			}
		}

		// 7.4 Convert schedule responses to attendance responses
		// Build a set of selected candidate IDs for filtering
		selectedCandidateSet := make(map[string]bool)
		for _, cid := range candidateIDs {
			selectedCandidateSet[cid.String()] = true
		}

		// Build response map: candidateID -> memberID -> response
		responseMap := make(map[string]map[string]*schedDomain.DateScheduleResponse)
		respondedMemberIDs := make(map[string]bool)
		for _, resp := range scheduleResponses {
			if !selectedCandidateSet[resp.CandidateID().String()] {
				continue
			}
			candID := resp.CandidateID().String()
			if responseMap[candID] == nil {
				responseMap[candID] = make(map[string]*schedDomain.DateScheduleResponse)
			}
			responseMap[candID][resp.MemberID().String()] = resp
			respondedMemberIDs[resp.MemberID().String()] = true
		}

		// Get all group members for unresponded member handling
		allGroupMemberIDs := make(map[string]common.MemberID)
		for _, ga := range groupAssignments {
			memberIDs, err := u.memberGroupRepo.FindMemberIDsByGroupID(ctx, ga.GroupID())
			if err != nil {
				return fmt.Errorf("failed to get group members: %w", err)
			}
			for _, mid := range memberIDs {
				allGroupMemberIDs[mid.String()] = mid
			}
		}

		// Process each target date
		for _, cid := range candidateIDs {
			targetDateID, ok := candidateToTargetDate[cid.String()]
			if !ok {
				continue
			}

			candidateResponses := responseMap[cid.String()]

			// 1. Process responded members
			for memberIDStr, resp := range candidateResponses {
				responseType := mapAvailabilityToResponseType(resp.Availability())

				attendanceResp, err := attendance.NewAttendanceResponse(
					resp.RespondedAt(), // Keep original responded_at
					collection.CollectionID(),
					tenantID,
					resp.MemberID(),
					targetDateID,
					responseType,
					resp.Note(),
					nil, // available_from
					nil, // available_to
				)
				if err != nil {
					return fmt.Errorf("failed to create attendance response for member %s: %w", memberIDStr, err)
				}

				if err := u.attendanceRepo.UpsertResponse(ctx, attendanceResp); err != nil {
					return fmt.Errorf("failed to save attendance response: %w", err)
				}
			}

			// 2. Process unresponded members (add as undecided)
			for memberIDStr, memberID := range allGroupMemberIDs {
				// Skip if already responded for this candidate
				if candidateResponses != nil {
					if _, hasResponse := candidateResponses[memberIDStr]; hasResponse {
						continue
					}
				}

				attendanceResp, err := attendance.NewAttendanceResponse(
					now, // Use current time for unresponded members
					collection.CollectionID(),
					tenantID,
					memberID,
					targetDateID,
					attendance.ResponseTypeUndecided,
					"", // No note
					nil,
					nil,
				)
				if err != nil {
					return fmt.Errorf("failed to create attendance response for unresponded member %s: %w", memberIDStr, err)
				}

				if err := u.attendanceRepo.UpsertResponse(ctx, attendanceResp); err != nil {
					return fmt.Errorf("failed to save attendance response for unresponded member: %w", err)
				}
			}
		}

		output = &ConvertToAttendanceOutput{
			CollectionID: collection.CollectionID().String(),
			PublicToken:  collection.PublicToken().String(),
			Title:        collection.Title(),
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return output, nil
}

// mapAvailabilityToResponseType converts schedule availability to attendance response type
func mapAvailabilityToResponseType(availability schedDomain.Availability) attendance.ResponseType {
	switch availability {
	case schedDomain.AvailabilityAvailable:
		return attendance.ResponseTypeAttending
	case schedDomain.AvailabilityUnavailable:
		return attendance.ResponseTypeAbsent
	case schedDomain.AvailabilityMaybe:
		return attendance.ResponseTypeUndecided
	default:
		return attendance.ResponseTypeUndecided
	}
}
