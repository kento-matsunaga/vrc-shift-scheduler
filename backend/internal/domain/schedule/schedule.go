package schedule

import (
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
)

// DateSchedule は日程調整の集約ルート
// MVP方針: responses は集約内で保持しない（Repository側UPSERTで管理）
// candidates は集約内で保持する（作成時に確定）
type DateSchedule struct {
	scheduleID         common.ScheduleID
	tenantID           common.TenantID
	title              string
	description        string
	eventID            *common.EventID
	publicToken        common.PublicToken
	status             Status
	deadline           *time.Time
	decidedCandidateID *common.CandidateID
	candidates         []*CandidateDate // 候補日は集約内で保持
	createdAt          time.Time
	updatedAt          time.Time
	deletedAt          *time.Time
}

// NewDateSchedule creates a new DateSchedule entity
// NOTE: now は App層で clock.Now() を呼んで渡す（Domain層で time.Now() を呼ばない）
func NewDateSchedule(
	now time.Time,
	scheduleID common.ScheduleID,
	tenantID common.TenantID,
	title string,
	description string,
	eventID *common.EventID,
	candidates []*CandidateDate,
	deadline *time.Time,
) (*DateSchedule, error) {
	schedule := &DateSchedule{
		scheduleID:  scheduleID,
		tenantID:    tenantID,
		title:       title,
		description: description,
		eventID:     eventID,
		publicToken: common.NewPublicToken(),
		status:      StatusOpen,
		deadline:    deadline,
		candidates:  candidates,
		createdAt:   now,
		updatedAt:   now,
	}

	if err := schedule.validate(); err != nil {
		return nil, err
	}

	return schedule, nil
}

// ReconstructDateSchedule reconstructs a DateSchedule entity from persistence
func ReconstructDateSchedule(
	scheduleID common.ScheduleID,
	tenantID common.TenantID,
	title string,
	description string,
	eventID *common.EventID,
	publicToken common.PublicToken,
	status Status,
	deadline *time.Time,
	decidedCandidateID *common.CandidateID,
	candidates []*CandidateDate,
	createdAt time.Time,
	updatedAt time.Time,
	deletedAt *time.Time,
) (*DateSchedule, error) {
	schedule := &DateSchedule{
		scheduleID:         scheduleID,
		tenantID:           tenantID,
		title:              title,
		description:        description,
		eventID:            eventID,
		publicToken:        publicToken,
		status:             status,
		deadline:           deadline,
		decidedCandidateID: decidedCandidateID,
		candidates:         candidates,
		createdAt:          createdAt,
		updatedAt:          updatedAt,
		deletedAt:          deletedAt,
	}

	if err := schedule.validate(); err != nil {
		return nil, err
	}

	return schedule, nil
}

func (s *DateSchedule) validate() error {
	// TenantID の必須性チェック
	if err := s.tenantID.Validate(); err != nil {
		return common.NewValidationError("tenant_id is required", err)
	}

	// Title の必須性チェック
	if s.title == "" {
		return common.NewValidationError("title is required", nil)
	}
	if len(s.title) > 255 {
		return common.NewValidationError("title must be less than 255 characters", nil)
	}

	// PublicToken の検証
	if err := s.publicToken.Validate(); err != nil {
		return err
	}

	// Status の検証
	if err := s.status.Validate(); err != nil {
		return err
	}

	// Candidates の必須性チェック（少なくとも1つ必要）
	if len(s.candidates) == 0 {
		return common.NewValidationError("at least one candidate is required", nil)
	}

	return nil
}

// CanRespond は回答可能かを判定（ドメインルール）
// now は App層から Clock 経由で渡される
func (s *DateSchedule) CanRespond(now time.Time) error {
	if s.status != StatusOpen {
		return ErrScheduleClosed
	}
	if s.deadline != nil && now.After(*s.deadline) {
		return ErrDeadlinePassed
	}
	return nil
}

// Decide は開催日を決定する（ドメインルール）
// now は App層から Clock 経由で渡される
func (s *DateSchedule) Decide(candidateID common.CandidateID, now time.Time) error {
	if s.status == StatusDecided {
		return ErrAlreadyDecided
	}

	// 候補日が存在するかチェック
	found := false
	for _, c := range s.candidates {
		if c.CandidateID() == candidateID {
			found = true
			break
		}
	}
	if !found {
		return ErrCandidateNotFound
	}

	s.status = StatusDecided
	s.decidedCandidateID = &candidateID
	s.updatedAt = now
	return nil
}

// Close はステータスをclosedに変更（ドメインルール）
// now は App層から Clock 経由で渡される
func (s *DateSchedule) Close(now time.Time) error {
	if s.status == StatusClosed || s.status == StatusDecided {
		return ErrAlreadyClosed
	}
	s.status = StatusClosed
	s.updatedAt = now
	return nil
}

// Delete はスケジュールを削除済みにする（ソフトデリート）
// now は App層から Clock 経由で渡される（Domain層で time.Now() を呼ばない）
func (s *DateSchedule) Delete(now time.Time) error {
	if s.deletedAt != nil {
		return ErrAlreadyDeleted
	}
	s.deletedAt = &now
	s.updatedAt = now
	return nil
}

// Getters

func (s *DateSchedule) ScheduleID() common.ScheduleID {
	return s.scheduleID
}

func (s *DateSchedule) TenantID() common.TenantID {
	return s.tenantID
}

func (s *DateSchedule) Title() string {
	return s.title
}

func (s *DateSchedule) Description() string {
	return s.description
}

func (s *DateSchedule) EventID() *common.EventID {
	return s.eventID
}

func (s *DateSchedule) PublicToken() common.PublicToken {
	return s.publicToken
}

func (s *DateSchedule) Status() Status {
	return s.status
}

func (s *DateSchedule) Deadline() *time.Time {
	return s.deadline
}

func (s *DateSchedule) DecidedCandidateID() *common.CandidateID {
	return s.decidedCandidateID
}

func (s *DateSchedule) Candidates() []*CandidateDate {
	return s.candidates
}

func (s *DateSchedule) CreatedAt() time.Time {
	return s.createdAt
}

func (s *DateSchedule) UpdatedAt() time.Time {
	return s.updatedAt
}

func (s *DateSchedule) DeletedAt() *time.Time {
	return s.deletedAt
}

func (s *DateSchedule) IsDeleted() bool {
	return s.deletedAt != nil
}
