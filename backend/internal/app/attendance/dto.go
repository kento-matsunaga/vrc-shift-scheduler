package attendance

import "time"

// TargetDateDTO represents a target date with ID
type TargetDateDTO struct {
	TargetDateID string    `json:"target_date_id"`
	TargetDate   time.Time `json:"target_date"`
	DisplayOrder int       `json:"display_order"`
}

// CreateCollectionInput represents the input for creating an attendance collection
type CreateCollectionInput struct {
	TenantID    string // from JWT context (管理API)
	Title       string
	Description string
	TargetType  string      // "event" or "business_day"
	TargetID    string      // event_id or business_day_id (optional)
	TargetDates []time.Time // 対象日の配列
	Deadline    *time.Time
	GroupIDs    []string // 対象グループID（複数可）
}

// CreateCollectionOutput represents the output for creating an attendance collection
type CreateCollectionOutput struct {
	CollectionID string     `json:"collection_id"`
	TenantID     string     `json:"tenant_id"`
	Title        string     `json:"title"`
	Description  string     `json:"description"`
	TargetType   string     `json:"target_type"`
	TargetID     string     `json:"target_id"`
	PublicToken  string     `json:"public_token"`
	Status       string     `json:"status"`
	Deadline     *time.Time `json:"deadline,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// SubmitResponseInput represents the input for submitting an attendance response
type SubmitResponseInput struct {
	PublicToken  string // from URL path (公開API)
	MemberID     string // from request body
	TargetDateID string // from request body - 対象日ID
	Response     string // "attending" or "absent"
	Note         string
}

// SubmitResponseOutput represents the output for submitting an attendance response
type SubmitResponseOutput struct {
	ResponseID   string    `json:"response_id"`
	CollectionID string    `json:"collection_id"`
	MemberID     string    `json:"member_id"`
	Response     string    `json:"response"`
	Note         string    `json:"note"`
	RespondedAt  time.Time `json:"responded_at"`
}

// CloseCollectionInput represents the input for closing an attendance collection
type CloseCollectionInput struct {
	TenantID     string // from JWT context (管理API)
	CollectionID string
}

// CloseCollectionOutput represents the output for closing an attendance collection
type CloseCollectionOutput struct {
	CollectionID string    `json:"collection_id"`
	Status       string    `json:"status"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// GetCollectionInput represents the input for getting a collection
type GetCollectionInput struct {
	TenantID     string // from JWT context (管理API)
	CollectionID string
}

// GetCollectionOutput represents the output for getting a collection
type GetCollectionOutput struct {
	CollectionID string          `json:"collection_id"`
	TenantID     string          `json:"tenant_id"`
	Title        string          `json:"title"`
	Description  string          `json:"description"`
	TargetType   string          `json:"target_type"`
	TargetID     string          `json:"target_id"`
	TargetDates  []TargetDateDTO `json:"target_dates,omitempty"` // 対象日の配列（IDあり）
	PublicToken  string          `json:"public_token"`
	Status       string          `json:"status"`
	Deadline     *time.Time      `json:"deadline,omitempty"`
	GroupIDs     []string        `json:"group_ids,omitempty"` // 対象グループID
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
}

// GetResponsesInput represents the input for getting all responses for a collection
type GetResponsesInput struct {
	TenantID     string // from JWT context (管理API)
	CollectionID string
}

// ResponseDTO represents a single attendance response in the output
type ResponseDTO struct {
	ResponseID   string    `json:"response_id"`
	MemberID     string    `json:"member_id"`
	MemberName   string    `json:"member_name"`    // メンバー表示名
	TargetDateID string    `json:"target_date_id"` // 対象日ID
	TargetDate   time.Time `json:"target_date"`    // 対象日
	Response     string    `json:"response"`
	Note         string    `json:"note"`
	RespondedAt  time.Time `json:"responded_at"`
}

// GetResponsesOutput represents the output for getting all responses for a collection
type GetResponsesOutput struct {
	CollectionID string        `json:"collection_id"`
	Responses    []ResponseDTO `json:"responses"`
}

// DeleteCollectionInput represents the input for deleting an attendance collection
type DeleteCollectionInput struct {
	TenantID     string // from JWT context (管理API)
	CollectionID string
}

// DeleteCollectionOutput represents the output for deleting a collection
type DeleteCollectionOutput struct {
	CollectionID string     `json:"collection_id"`
	Status       string     `json:"status"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty"`
	UpdatedAt    time.Time  `json:"updated_at"`
}
