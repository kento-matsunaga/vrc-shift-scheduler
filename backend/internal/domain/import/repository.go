package importjob

import (
	"context"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
)

// ImportJobRepository defines the interface for import job persistence
type ImportJobRepository interface {
	// Save saves an import job
	Save(ctx context.Context, job *ImportJob) error

	// Update updates an existing import job
	Update(ctx context.Context, job *ImportJob) error

	// FindByID finds an import job by ID
	FindByID(ctx context.Context, id common.ImportJobID) (*ImportJob, error)

	// FindByIDAndTenantID finds an import job by ID and tenant ID (for authorization)
	FindByIDAndTenantID(ctx context.Context, id common.ImportJobID, tenantID common.TenantID) (*ImportJob, error)

	// FindByTenantID finds all import jobs for a tenant
	FindByTenantID(ctx context.Context, tenantID common.TenantID, limit, offset int) ([]*ImportJob, error)

	// CountByTenantID counts import jobs for a tenant
	CountByTenantID(ctx context.Context, tenantID common.TenantID) (int, error)
}

// ImportLogStatus represents the status of a single import log entry
type ImportLogStatus string

const (
	ImportLogStatusSuccess ImportLogStatus = "success"
	ImportLogStatusError   ImportLogStatus = "error"
	ImportLogStatusSkipped ImportLogStatus = "skipped"
)

// ImportJobLog represents a log entry for a single row in an import job
type ImportJobLog struct {
	LogID        common.ImportLogID
	ImportJobID  common.ImportJobID
	RowNumber    int
	Status       ImportLogStatus
	InputData    map[string]interface{}
	ErrorMessage string
}

// ImportJobLogRepository defines the interface for import job log persistence
type ImportJobLogRepository interface {
	// SaveBatch saves multiple log entries in a batch
	SaveBatch(ctx context.Context, logs []*ImportJobLog) error

	// FindByJobID finds all log entries for a job
	FindByJobID(ctx context.Context, jobID common.ImportJobID) ([]*ImportJobLog, error)

	// FindErrorsByJobID finds only error log entries for a job
	FindErrorsByJobID(ctx context.Context, jobID common.ImportJobID) ([]*ImportJobLog, error)
}
