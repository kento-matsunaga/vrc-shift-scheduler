package attendance

import "time"

// TargetDateDTO represents a target date with ID
type TargetDateDTO struct {
	TargetDateID string    `json:"target_date_id"`
	TargetDate   time.Time `json:"target_date"`
	StartTime    *string   `json:"start_time,omitempty"` // 開始時間（HH:MM形式、任意）
	EndTime      *string   `json:"end_time,omitempty"`   // 終了時間（HH:MM形式、任意）
	DisplayOrder int       `json:"display_order"`
}

// TargetDateInput represents input for a target date when creating a collection
type TargetDateInput struct {
	TargetDate time.Time
	StartTime  *string // 開始時間（HH:MM形式、任意）
	EndTime    *string // 終了時間（HH:MM形式、任意）
}

// CreateCollectionInput represents the input for creating an attendance collection
type CreateCollectionInput struct {
	TenantID    string // from JWT context (管理API)
	Title       string
	Description string
	TargetType  string            // "event" or "business_day"
	TargetID    string            // event_id or business_day_id (optional)
	TargetDates []TargetDateInput // 対象日の配列（開始時間込み）
	Deadline    *time.Time
	GroupIDs    []string // 対象グループID（複数可）
	RoleIDs     []string // 対象ロールID（複数可）
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

// UpdateCollectionInput represents the input for updating an attendance collection
type UpdateCollectionInput struct {
	TenantID     string // from JWT context (管理API)
	CollectionID string
	Title        string
	Description  string
	Deadline     *time.Time
}

// UpdateCollectionOutput represents the output for updating an attendance collection
type UpdateCollectionOutput struct {
	CollectionID string     `json:"collection_id"`
	TenantID     string     `json:"tenant_id"`
	Title        string     `json:"title"`
	Description  string     `json:"description"`
	Status       string     `json:"status"`
	Deadline     *time.Time `json:"deadline,omitempty"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// SubmitResponseInput represents the input for submitting an attendance response
type SubmitResponseInput struct {
	PublicToken   string // from URL path (公開API)
	MemberID      string // from request body
	TargetDateID  string // from request body - 対象日ID
	Response      string // "attending" or "absent" or "undecided"
	Note          string
	AvailableFrom *string // 参加可能開始時間 (HH:MM)
	AvailableTo   *string // 参加可能終了時間 (HH:MM)
}

// SubmitResponseOutput represents the output for submitting an attendance response
type SubmitResponseOutput struct {
	ResponseID    string    `json:"response_id"`
	CollectionID  string    `json:"collection_id"`
	MemberID      string    `json:"member_id"`
	Response      string    `json:"response"`
	Note          string    `json:"note"`
	AvailableFrom *string   `json:"available_from,omitempty"`
	AvailableTo   *string   `json:"available_to,omitempty"`
	RespondedAt   time.Time `json:"responded_at"`
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
	RoleIDs      []string        `json:"role_ids,omitempty"`  // 対象ロールID
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
	ResponseID    string    `json:"response_id"`
	MemberID      string    `json:"member_id"`
	MemberName    string    `json:"member_name"`    // メンバー表示名
	TargetDateID  string    `json:"target_date_id"` // 対象日ID
	TargetDate    time.Time `json:"target_date"`    // 対象日
	Response      string    `json:"response"`
	Note          string    `json:"note"`
	AvailableFrom *string   `json:"available_from,omitempty"` // 参加可能開始時間
	AvailableTo   *string   `json:"available_to,omitempty"`   // 参加可能終了時間
	RespondedAt   time.Time `json:"responded_at"`
}

// GetResponsesOutput represents the output for getting all responses for a collection
type GetResponsesOutput struct {
	CollectionID string        `json:"collection_id"`
	Responses    []ResponseDTO `json:"responses"`
}

// GetMemberResponsesInput represents the input for getting a member's responses (public API)
type GetMemberResponsesInput struct {
	PublicToken string // from URL path
	MemberID    string // from URL path
}

// MemberResponseDTO represents a single response for a specific member
type MemberResponseDTO struct {
	TargetDateID  string  `json:"target_date_id"`
	Response      string  `json:"response"`
	Note          string  `json:"note"`
	AvailableFrom *string `json:"available_from,omitempty"`
	AvailableTo   *string `json:"available_to,omitempty"`
}

// GetMemberResponsesOutput represents the output for getting a member's responses
type GetMemberResponsesOutput struct {
	MemberID  string              `json:"member_id"`
	Responses []MemberResponseDTO `json:"responses"`
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

// AdminUpdateResponseInput represents the input for admin updating an attendance response
// 管理者による出欠回答の更新（締め切り後も可能）
type AdminUpdateResponseInput struct {
	TenantID      string // from JWT context (管理API)
	CollectionID  string // from URL path
	MemberID      string // from request body
	TargetDateID  string // from request body
	Response      string // "attending" or "absent" or "undecided"
	Note          string
	AvailableFrom *string // 参加可能開始時間 (HH:MM)
	AvailableTo   *string // 参加可能終了時間 (HH:MM)
}

// AdminUpdateResponseOutput represents the output for admin updating an attendance response
type AdminUpdateResponseOutput struct {
	ResponseID    string    `json:"response_id"`
	CollectionID  string    `json:"collection_id"`
	MemberID      string    `json:"member_id"`
	TargetDateID  string    `json:"target_date_id"`
	Response      string    `json:"response"`
	Note          string    `json:"note"`
	AvailableFrom *string   `json:"available_from,omitempty"`
	AvailableTo   *string   `json:"available_to,omitempty"`
	RespondedAt   time.Time `json:"responded_at"`
}
