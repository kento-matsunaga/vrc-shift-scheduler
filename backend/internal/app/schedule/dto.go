package schedule

import "time"

// CandidateInput represents a candidate date input
type CandidateInput struct {
	Date      time.Time
	StartTime *time.Time
	EndTime   *time.Time
}

// CreateScheduleInput represents the input for creating a schedule
type CreateScheduleInput struct {
	TenantID    string // from JWT context (管理API)
	Title       string
	Description string
	EventID     *string // optional
	Candidates  []CandidateInput
	Deadline    *time.Time
	GroupIDs    []string // optional: target group IDs
}

// CreateScheduleOutput represents the output for creating a schedule
type CreateScheduleOutput struct {
	ScheduleID  string         `json:"schedule_id"`
	TenantID    string         `json:"tenant_id"`
	Title       string         `json:"title"`
	Description string         `json:"description"`
	EventID     *string        `json:"event_id,omitempty"`
	PublicToken string         `json:"public_token"`
	Status      string         `json:"status"`
	Deadline    *time.Time     `json:"deadline,omitempty"`
	Candidates  []CandidateDTO `json:"candidates"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

// CandidateDTO represents a candidate date in responses
type CandidateDTO struct {
	CandidateID string     `json:"candidate_id"`
	Date        time.Time  `json:"date"`
	StartTime   *time.Time `json:"start_time,omitempty"`
	EndTime     *time.Time `json:"end_time,omitempty"`
}

// SubmitResponseInput represents the input for submitting responses
type SubmitResponseInput struct {
	PublicToken string // from URL path (公開API)
	MemberID    string // from request body
	Responses   []ResponseInput
}

// ResponseInput represents a single response for a candidate
type ResponseInput struct {
	CandidateID  string
	Availability string // "available", "unavailable", "maybe"
	Note         string
}

// SubmitResponseOutput represents the output for submitting responses
type SubmitResponseOutput struct {
	ScheduleID  string    `json:"schedule_id"`
	MemberID    string    `json:"member_id"`
	RespondedAt time.Time `json:"responded_at"`
}

// DecideScheduleInput represents the input for deciding a schedule
type DecideScheduleInput struct {
	TenantID    string // from JWT context (管理API)
	ScheduleID  string
	CandidateID string
}

// DecideScheduleOutput represents the output for deciding a schedule
type DecideScheduleOutput struct {
	ScheduleID         string    `json:"schedule_id"`
	Status             string    `json:"status"`
	DecidedCandidateID string    `json:"decided_candidate_id"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// CloseScheduleInput represents the input for closing a schedule
type CloseScheduleInput struct {
	TenantID   string // from JWT context (管理API)
	ScheduleID string
}

// CloseScheduleOutput represents the output for closing a schedule
type CloseScheduleOutput struct {
	ScheduleID string    `json:"schedule_id"`
	Status     string    `json:"status"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// GetScheduleInput represents the input for getting a schedule
type GetScheduleInput struct {
	TenantID   string // from JWT context (管理API)
	ScheduleID string
}

// GetScheduleOutput represents the output for getting a schedule
type GetScheduleOutput struct {
	ScheduleID         string         `json:"schedule_id"`
	TenantID           string         `json:"tenant_id"`
	Title              string         `json:"title"`
	Description        string         `json:"description"`
	EventID            *string        `json:"event_id,omitempty"`
	PublicToken        string         `json:"public_token"`
	Status             string         `json:"status"`
	Deadline           *time.Time     `json:"deadline,omitempty"`
	DecidedCandidateID *string        `json:"decided_candidate_id,omitempty"`
	Candidates         []CandidateDTO `json:"candidates"`
	GroupIDs           []string       `json:"group_ids,omitempty"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`
}

// GetResponsesInput represents the input for getting responses
type GetResponsesInput struct {
	TenantID   string // from JWT context (管理API)
	ScheduleID string
}

// ScheduleResponseDTO represents a single response
type ScheduleResponseDTO struct {
	ResponseID   string    `json:"response_id"`
	MemberID     string    `json:"member_id"`
	CandidateID  string    `json:"candidate_id"`
	Availability string    `json:"availability"`
	Note         string    `json:"note"`
	RespondedAt  time.Time `json:"responded_at"`
}

// GetResponsesOutput represents the output for getting responses
type GetResponsesOutput struct {
	ScheduleID string                `json:"schedule_id"`
	Responses  []ScheduleResponseDTO `json:"responses"`
}

// DeleteScheduleInput represents the input for deleting a schedule
type DeleteScheduleInput struct {
	TenantID   string // from JWT context (管理API)
	ScheduleID string
}

// DeleteScheduleOutput represents the output for deleting a schedule
type DeleteScheduleOutput struct {
	ScheduleID string     `json:"schedule_id"`
	Status     string     `json:"status"`
	DeletedAt  *time.Time `json:"deleted_at,omitempty"`
	UpdatedAt  time.Time  `json:"updated_at"`
}
