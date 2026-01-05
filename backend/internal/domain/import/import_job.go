package importjob

import (
	"encoding/json"
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
)

// ImportType represents the type of import
type ImportType string

const (
	ImportTypeMembers          ImportType = "members"
	ImportTypeActualAttendance ImportType = "actual_attendance"
)

func (t ImportType) String() string {
	return string(t)
}

func (t ImportType) IsValid() bool {
	switch t {
	case ImportTypeMembers, ImportTypeActualAttendance:
		return true
	default:
		return false
	}
}

// ImportStatus represents the status of an import job
type ImportStatus string

const (
	ImportStatusPending    ImportStatus = "pending"
	ImportStatusProcessing ImportStatus = "processing"
	ImportStatusCompleted  ImportStatus = "completed"
	ImportStatusFailed     ImportStatus = "failed"
)

func (s ImportStatus) String() string {
	return string(s)
}

func (s ImportStatus) IsValid() bool {
	switch s {
	case ImportStatusPending, ImportStatusProcessing, ImportStatusCompleted, ImportStatusFailed:
		return true
	default:
		return false
	}
}

// ImportOptions represents the options for an import job
type ImportOptions struct {
	SkipExisting         bool     `json:"skip_existing"`
	UpdateExisting       bool     `json:"update_existing"`
	DefaultRoleIDs       []string `json:"default_role_ids,omitempty"`
	DefaultGroupIDs      []string `json:"default_group_ids,omitempty"`
	CreateMissingEvents  bool     `json:"create_missing_events,omitempty"`
	CreateMissingSlots   bool     `json:"create_missing_slots,omitempty"`
	FuzzyMemberMatch     bool     `json:"fuzzy_member_match,omitempty"`
}

// ToJSON converts ImportOptions to JSON bytes
func (o ImportOptions) ToJSON() ([]byte, error) {
	return json.Marshal(o)
}

// ParseImportOptions parses JSON bytes into ImportOptions
func ParseImportOptions(data []byte) (ImportOptions, error) {
	var opts ImportOptions
	if len(data) == 0 {
		return opts, nil
	}
	err := json.Unmarshal(data, &opts)
	return opts, err
}

// ErrorDetail represents a single error in the import process
type ErrorDetail struct {
	Row     int    `json:"row"`
	Message string `json:"message"`
}

// ImportJob represents an import job entity (aggregate root)
type ImportJob struct {
	importJobID   common.ImportJobID
	tenantID      common.TenantID
	importType    ImportType
	status        ImportStatus
	fileName      string
	totalRows     int
	processedRows int
	successCount  int
	errorCount    int
	errorDetails  []ErrorDetail
	options       ImportOptions
	startedAt     *time.Time
	completedAt   *time.Time
	createdAt     time.Time
	createdBy     common.AdminID
}

// NewImportJob creates a new ImportJob entity
func NewImportJob(
	now time.Time,
	tenantID common.TenantID,
	importType ImportType,
	fileName string,
	options ImportOptions,
	createdBy common.AdminID,
) (*ImportJob, error) {
	job := &ImportJob{
		importJobID:   common.NewImportJobIDWithTime(now),
		tenantID:      tenantID,
		importType:    importType,
		status:        ImportStatusPending,
		fileName:      fileName,
		totalRows:     0,
		processedRows: 0,
		successCount:  0,
		errorCount:    0,
		errorDetails:  []ErrorDetail{},
		options:       options,
		createdAt:     now,
		createdBy:     createdBy,
	}

	if err := job.validate(); err != nil {
		return nil, err
	}

	return job, nil
}

// ReconstructImportJob reconstructs an ImportJob entity from persistence
func ReconstructImportJob(
	importJobID common.ImportJobID,
	tenantID common.TenantID,
	importType ImportType,
	status ImportStatus,
	fileName string,
	totalRows int,
	processedRows int,
	successCount int,
	errorCount int,
	errorDetails []ErrorDetail,
	options ImportOptions,
	startedAt *time.Time,
	completedAt *time.Time,
	createdAt time.Time,
	createdBy common.AdminID,
) (*ImportJob, error) {
	job := &ImportJob{
		importJobID:   importJobID,
		tenantID:      tenantID,
		importType:    importType,
		status:        status,
		fileName:      fileName,
		totalRows:     totalRows,
		processedRows: processedRows,
		successCount:  successCount,
		errorCount:    errorCount,
		errorDetails:  errorDetails,
		options:       options,
		startedAt:     startedAt,
		completedAt:   completedAt,
		createdAt:     createdAt,
		createdBy:     createdBy,
	}

	return job, nil
}

func (j *ImportJob) validate() error {
	if err := j.tenantID.Validate(); err != nil {
		return common.NewValidationError("tenant_id is required", err)
	}

	if !j.importType.IsValid() {
		return common.NewValidationError("invalid import_type", nil)
	}

	if err := j.createdBy.Validate(); err != nil {
		return common.NewValidationError("created_by is required", err)
	}

	return nil
}

// Getters

func (j *ImportJob) ImportJobID() common.ImportJobID {
	return j.importJobID
}

func (j *ImportJob) TenantID() common.TenantID {
	return j.tenantID
}

func (j *ImportJob) ImportType() ImportType {
	return j.importType
}

func (j *ImportJob) Status() ImportStatus {
	return j.status
}

func (j *ImportJob) FileName() string {
	return j.fileName
}

func (j *ImportJob) TotalRows() int {
	return j.totalRows
}

func (j *ImportJob) ProcessedRows() int {
	return j.processedRows
}

func (j *ImportJob) SuccessCount() int {
	return j.successCount
}

func (j *ImportJob) ErrorCount() int {
	return j.errorCount
}

func (j *ImportJob) ErrorDetails() []ErrorDetail {
	return j.errorDetails
}

func (j *ImportJob) Options() ImportOptions {
	return j.options
}

func (j *ImportJob) StartedAt() *time.Time {
	return j.startedAt
}

func (j *ImportJob) CompletedAt() *time.Time {
	return j.completedAt
}

func (j *ImportJob) CreatedAt() time.Time {
	return j.createdAt
}

func (j *ImportJob) CreatedBy() common.AdminID {
	return j.createdBy
}

// Progress returns the progress percentage (0-100)
func (j *ImportJob) Progress() float64 {
	if j.totalRows == 0 {
		return 0
	}
	return float64(j.processedRows) / float64(j.totalRows) * 100
}

// Business logic methods

// Start starts the import job
func (j *ImportJob) Start(now time.Time, totalRows int) error {
	if j.status != ImportStatusPending {
		return common.NewValidationError("job is not in pending status", nil)
	}

	j.status = ImportStatusProcessing
	j.totalRows = totalRows
	j.startedAt = &now
	return nil
}

// RecordSuccess records a successful row processing
func (j *ImportJob) RecordSuccess() {
	j.processedRows++
	j.successCount++
}

// RecordError records an error for a specific row
func (j *ImportJob) RecordError(row int, message string) {
	j.processedRows++
	j.errorCount++
	j.errorDetails = append(j.errorDetails, ErrorDetail{
		Row:     row,
		Message: message,
	})
}

// RecordSkip records a skipped row
func (j *ImportJob) RecordSkip() {
	j.processedRows++
}

// Complete marks the job as completed
func (j *ImportJob) Complete(now time.Time) error {
	if j.status != ImportStatusProcessing {
		return common.NewValidationError("job is not in processing status", nil)
	}

	j.status = ImportStatusCompleted
	j.completedAt = &now
	return nil
}

// Fail marks the job as failed
func (j *ImportJob) Fail(now time.Time, reason string) error {
	j.status = ImportStatusFailed
	j.completedAt = &now
	j.errorDetails = append(j.errorDetails, ErrorDetail{
		Row:     0,
		Message: reason,
	})
	return nil
}
