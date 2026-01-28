package schedule

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/schedule"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/services"
)

type UpdateScheduleUsecase struct {
	repo  schedule.DateScheduleRepository
	clock services.Clock
}

func NewUpdateScheduleUsecase(repo schedule.DateScheduleRepository, clk services.Clock) *UpdateScheduleUsecase {
	return &UpdateScheduleUsecase{repo: repo, clock: clk}
}

func (u *UpdateScheduleUsecase) Execute(ctx context.Context, input UpdateScheduleInput) (*UpdateScheduleOutput, error) {
	tenantID, err := common.ParseTenantID(input.TenantID)
	if err != nil {
		return nil, err
	}

	scheduleID, err := common.ParseScheduleID(input.ScheduleID)
	if err != nil {
		return nil, err
	}

	sch, err := u.repo.FindByID(ctx, tenantID, scheduleID)
	if err != nil {
		return nil, err
	}

	now := u.clock.Now()

	var candidates []*schedule.CandidateDate
	if input.Candidates != nil {
		existingCandidates := make(map[string]*schedule.CandidateDate, len(sch.Candidates()))
		for _, c := range sch.Candidates() {
			existingCandidates[candidateKey(c.CandidateDateValue(), c.StartTime(), c.EndTime())] = c
		}

		removedCandidates, err := findRemovedCandidates(sch.Candidates(), input.Candidates)
		if err != nil {
			return nil, err
		}
		if len(removedCandidates) > 0 {
			hasResponses, err := u.hasResponsesForCandidates(ctx, scheduleID, removedCandidates)
			if err != nil {
				return nil, err
			}
			if hasResponses != nil && !input.ForceDeleteCandidateResponses {
				return nil, common.NewConflictError(candidateRemovalMessage(hasResponses))
			}
		}

		candidates = make([]*schedule.CandidateDate, 0, len(input.Candidates))
		for i, c := range input.Candidates {
			key := candidateKey(c.Date, c.StartTime, c.EndTime)
			if existing, ok := existingCandidates[key]; ok {
				updatedCandidate, err := schedule.ReconstructCandidateDate(
					existing.CandidateID(),
					scheduleID,
					c.Date,
					c.StartTime,
					c.EndTime,
					i,
					existing.CreatedAt(),
				)
				if err != nil {
					return nil, err
				}
				candidates = append(candidates, updatedCandidate)
				continue
			}

			candidate, err := schedule.NewCandidateDate(now, scheduleID, c.Date, c.StartTime, c.EndTime, i)
			if err != nil {
				return nil, err
			}
			candidates = append(candidates, candidate)
		}
	}

	if err := sch.Update(now, input.Title, input.Description, input.Deadline, candidates); err != nil {
		return nil, err
	}

	if err := u.repo.Save(ctx, sch); err != nil {
		return nil, err
	}
	log.Printf("[AUDIT] UpdateSchedule: tenant=%s schedule=%s", sch.TenantID().String(), sch.ScheduleID().String())

	candidateDTOs := make([]CandidateDTO, len(sch.Candidates()))
	for i, c := range sch.Candidates() {
		candidateDTOs[i] = CandidateDTO{
			CandidateID: c.CandidateID().String(),
			Date:        c.CandidateDateValue(),
			StartTime:   c.StartTime(),
			EndTime:     c.EndTime(),
		}
	}

	return &UpdateScheduleOutput{
		ScheduleID:  sch.ScheduleID().String(),
		TenantID:    sch.TenantID().String(),
		Title:       sch.Title(),
		Description: sch.Description(),
		Status:      sch.Status().String(),
		Deadline:    sch.Deadline(),
		Candidates:  candidateDTOs,
		UpdatedAt:   sch.UpdatedAt(),
	}, nil
}

func (u *UpdateScheduleUsecase) hasResponsesForCandidates(
	ctx context.Context,
	scheduleID common.ScheduleID,
	candidates []*schedule.CandidateDate,
) (*schedule.CandidateDate, error) {
	responses, err := u.repo.FindResponsesByScheduleID(ctx, scheduleID)
	if err != nil {
		return nil, err
	}

	responseCandidateIDs := make(map[string]struct{}, len(responses))
	for _, r := range responses {
		responseCandidateIDs[r.CandidateID().String()] = struct{}{}
	}

	for _, c := range candidates {
		if _, ok := responseCandidateIDs[c.CandidateID().String()]; ok {
			return c, nil
		}
	}

	return nil, nil
}

func findRemovedCandidates(
	existing []*schedule.CandidateDate,
	input []CandidateInput,
) ([]*schedule.CandidateDate, error) {
	incomingKeys := make(map[string]struct{}, len(input))
	for _, c := range input {
		incomingKeys[candidateKey(c.Date, c.StartTime, c.EndTime)] = struct{}{}
	}

	removed := make([]*schedule.CandidateDate, 0)
	for _, c := range existing {
		key := candidateKey(c.CandidateDateValue(), c.StartTime(), c.EndTime())
		if _, ok := incomingKeys[key]; !ok {
			removed = append(removed, c)
		}
	}

	return removed, nil
}

func candidateKey(date time.Time, startTime *time.Time, endTime *time.Time) string {
	key := date.Format(time.RFC3339)
	if startTime != nil {
		key = fmt.Sprintf("%s|%s", key, startTime.Format(time.RFC3339))
	} else {
		key = fmt.Sprintf("%s|", key)
	}
	if endTime != nil {
		key = fmt.Sprintf("%s|%s", key, endTime.Format(time.RFC3339))
	}
	return key
}

func candidateRemovalMessage(candidate *schedule.CandidateDate) string {
	date := candidate.CandidateDateValue().Format("2006-01-02")
	return fmt.Sprintf("削除しようとしている候補日(%s)に既存の回答が存在します。削除しますか？", date)
}
