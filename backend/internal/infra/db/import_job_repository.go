package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	importjob "github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/import"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ImportJobRepository implements importjob.ImportJobRepository for PostgreSQL
type ImportJobRepository struct {
	db *pgxpool.Pool
}

// NewImportJobRepository creates a new ImportJobRepository
func NewImportJobRepository(db *pgxpool.Pool) *ImportJobRepository {
	return &ImportJobRepository{db: db}
}

// Save saves an import job
func (r *ImportJobRepository) Save(ctx context.Context, job *importjob.ImportJob) error {
	errorDetailsJSON, err := json.Marshal(job.ErrorDetails())
	if err != nil {
		return fmt.Errorf("failed to marshal error_details: %w", err)
	}

	optionsJSON, err := job.Options().ToJSON()
	if err != nil {
		return fmt.Errorf("failed to marshal options: %w", err)
	}

	query := `
		INSERT INTO import_jobs (
			import_job_id, tenant_id, import_type, status, file_name,
			total_rows, processed_rows, success_count, error_count,
			error_details, options, started_at, completed_at, created_at, created_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
	`

	_, err = r.db.Exec(ctx, query,
		job.ImportJobID().String(),
		job.TenantID().String(),
		job.ImportType().String(),
		job.Status().String(),
		job.FileName(),
		job.TotalRows(),
		job.ProcessedRows(),
		job.SuccessCount(),
		job.ErrorCount(),
		errorDetailsJSON,
		optionsJSON,
		job.StartedAt(),
		job.CompletedAt(),
		job.CreatedAt(),
		job.CreatedBy().String(),
	)

	if err != nil {
		return fmt.Errorf("failed to save import job: %w", err)
	}

	return nil
}

// Update updates an existing import job
func (r *ImportJobRepository) Update(ctx context.Context, job *importjob.ImportJob) error {
	errorDetailsJSON, err := json.Marshal(job.ErrorDetails())
	if err != nil {
		return fmt.Errorf("failed to marshal error_details: %w", err)
	}

	query := `
		UPDATE import_jobs SET
			status = $2,
			total_rows = $3,
			processed_rows = $4,
			success_count = $5,
			error_count = $6,
			error_details = $7,
			started_at = $8,
			completed_at = $9
		WHERE import_job_id = $1
	`

	_, err = r.db.Exec(ctx, query,
		job.ImportJobID().String(),
		job.Status().String(),
		job.TotalRows(),
		job.ProcessedRows(),
		job.SuccessCount(),
		job.ErrorCount(),
		errorDetailsJSON,
		job.StartedAt(),
		job.CompletedAt(),
	)

	if err != nil {
		return fmt.Errorf("failed to update import job: %w", err)
	}

	return nil
}

// FindByID finds an import job by ID
func (r *ImportJobRepository) FindByID(ctx context.Context, id common.ImportJobID) (*importjob.ImportJob, error) {
	query := `
		SELECT
			import_job_id, tenant_id, import_type, status, file_name,
			total_rows, processed_rows, success_count, error_count,
			error_details, options, started_at, completed_at, created_at, created_by
		FROM import_jobs
		WHERE import_job_id = $1
	`

	var (
		importJobIDStr string
		tenantIDStr    string
		importTypeStr  string
		statusStr      string
		fileName       sql.NullString
		totalRows      int
		processedRows  int
		successCount   int
		errorCount     int
		errorDetails   []byte
		options        []byte
		startedAt      sql.NullTime
		completedAt    sql.NullTime
		createdAt      time.Time
		createdByStr   string
	)

	err := r.db.QueryRow(ctx, query, id.String()).Scan(
		&importJobIDStr,
		&tenantIDStr,
		&importTypeStr,
		&statusStr,
		&fileName,
		&totalRows,
		&processedRows,
		&successCount,
		&errorCount,
		&errorDetails,
		&options,
		&startedAt,
		&completedAt,
		&createdAt,
		&createdByStr,
	)

	if err == pgx.ErrNoRows {
		return nil, common.NewNotFoundError("ImportJob", id.String())
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find import job: %w", err)
	}

	return r.reconstructJob(
		importJobIDStr, tenantIDStr, importTypeStr, statusStr, fileName.String,
		totalRows, processedRows, successCount, errorCount,
		errorDetails, options, startedAt, completedAt, createdAt, createdByStr,
	)
}

// FindByIDAndTenantID finds an import job by ID and tenant ID (for authorization)
func (r *ImportJobRepository) FindByIDAndTenantID(ctx context.Context, id common.ImportJobID, tenantID common.TenantID) (*importjob.ImportJob, error) {
	query := `
		SELECT
			import_job_id, tenant_id, import_type, status, file_name,
			total_rows, processed_rows, success_count, error_count,
			error_details, options, started_at, completed_at, created_at, created_by
		FROM import_jobs
		WHERE import_job_id = $1 AND tenant_id = $2
	`

	var (
		importJobIDStr string
		tenantIDStr    string
		importTypeStr  string
		statusStr      string
		fileName       sql.NullString
		totalRows      int
		processedRows  int
		successCount   int
		errorCount     int
		errorDetails   []byte
		options        []byte
		startedAt      sql.NullTime
		completedAt    sql.NullTime
		createdAt      time.Time
		createdByStr   string
	)

	err := r.db.QueryRow(ctx, query, id.String(), tenantID.String()).Scan(
		&importJobIDStr,
		&tenantIDStr,
		&importTypeStr,
		&statusStr,
		&fileName,
		&totalRows,
		&processedRows,
		&successCount,
		&errorCount,
		&errorDetails,
		&options,
		&startedAt,
		&completedAt,
		&createdAt,
		&createdByStr,
	)

	if err == pgx.ErrNoRows {
		return nil, common.NewNotFoundError("ImportJob", id.String())
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find import job: %w", err)
	}

	return r.reconstructJob(
		importJobIDStr, tenantIDStr, importTypeStr, statusStr, fileName.String,
		totalRows, processedRows, successCount, errorCount,
		errorDetails, options, startedAt, completedAt, createdAt, createdByStr,
	)
}

// FindByTenantID finds all import jobs for a tenant
func (r *ImportJobRepository) FindByTenantID(ctx context.Context, tenantID common.TenantID, limit, offset int) ([]*importjob.ImportJob, error) {
	query := `
		SELECT
			import_job_id, tenant_id, import_type, status, file_name,
			total_rows, processed_rows, success_count, error_count,
			error_details, options, started_at, completed_at, created_at, created_by
		FROM import_jobs
		WHERE tenant_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, tenantID.String(), limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to find import jobs: %w", err)
	}
	defer rows.Close()

	var jobs []*importjob.ImportJob
	for rows.Next() {
		var (
			importJobIDStr string
			tenantIDStr    string
			importTypeStr  string
			statusStr      string
			fileName       sql.NullString
			totalRows      int
			processedRows  int
			successCount   int
			errorCount     int
			errorDetails   []byte
			options        []byte
			startedAt      sql.NullTime
			completedAt    sql.NullTime
			createdAt      time.Time
			createdByStr   string
		)

		err := rows.Scan(
			&importJobIDStr,
			&tenantIDStr,
			&importTypeStr,
			&statusStr,
			&fileName,
			&totalRows,
			&processedRows,
			&successCount,
			&errorCount,
			&errorDetails,
			&options,
			&startedAt,
			&completedAt,
			&createdAt,
			&createdByStr,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan import job: %w", err)
		}

		job, err := r.reconstructJob(
			importJobIDStr, tenantIDStr, importTypeStr, statusStr, fileName.String,
			totalRows, processedRows, successCount, errorCount,
			errorDetails, options, startedAt, completedAt, createdAt, createdByStr,
		)
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, job)
	}

	return jobs, nil
}

// CountByTenantID counts import jobs for a tenant
func (r *ImportJobRepository) CountByTenantID(ctx context.Context, tenantID common.TenantID) (int, error) {
	query := `SELECT COUNT(*) FROM import_jobs WHERE tenant_id = $1`

	var count int
	err := r.db.QueryRow(ctx, query, tenantID.String()).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count import jobs: %w", err)
	}

	return count, nil
}

func (r *ImportJobRepository) reconstructJob(
	importJobIDStr, tenantIDStr, importTypeStr, statusStr, fileName string,
	totalRows, processedRows, successCount, errorCount int,
	errorDetailsJSON, optionsJSON []byte,
	startedAt, completedAt sql.NullTime,
	createdAt time.Time, createdByStr string,
) (*importjob.ImportJob, error) {
	importJobID, err := common.ParseImportJobID(importJobIDStr)
	if err != nil {
		return nil, err
	}

	tenantID, err := common.ParseTenantID(tenantIDStr)
	if err != nil {
		return nil, err
	}

	createdBy, err := common.ParseAdminID(createdByStr)
	if err != nil {
		return nil, err
	}

	var errorDetails []importjob.ErrorDetail
	if len(errorDetailsJSON) > 0 {
		if err := json.Unmarshal(errorDetailsJSON, &errorDetails); err != nil {
			return nil, fmt.Errorf("failed to unmarshal error_details: %w", err)
		}
	}

	opts, err := importjob.ParseImportOptions(optionsJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to parse options: %w", err)
	}

	var startedAtPtr, completedAtPtr *time.Time
	if startedAt.Valid {
		startedAtPtr = &startedAt.Time
	}
	if completedAt.Valid {
		completedAtPtr = &completedAt.Time
	}

	return importjob.ReconstructImportJob(
		importJobID,
		tenantID,
		importjob.ImportType(importTypeStr),
		importjob.ImportStatus(statusStr),
		fileName,
		totalRows,
		processedRows,
		successCount,
		errorCount,
		errorDetails,
		opts,
		startedAtPtr,
		completedAtPtr,
		createdAt,
		createdBy,
	)
}

// ImportJobLogRepository implements importjob.ImportJobLogRepository for PostgreSQL
type ImportJobLogRepository struct {
	db *pgxpool.Pool
}

// NewImportJobLogRepository creates a new ImportJobLogRepository
func NewImportJobLogRepository(db *pgxpool.Pool) *ImportJobLogRepository {
	return &ImportJobLogRepository{db: db}
}

// SaveBatch saves multiple log entries in a batch
func (r *ImportJobLogRepository) SaveBatch(ctx context.Context, logs []*importjob.ImportJobLog) error {
	if len(logs) == 0 {
		return nil
	}

	// Use a transaction for batch insert
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	query := `
		INSERT INTO import_job_logs (
			log_id, import_job_id, row_number, status, input_data, error_message, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, NOW())
	`

	for _, log := range logs {
		inputDataJSON, err := json.Marshal(log.InputData)
		if err != nil {
			return fmt.Errorf("failed to marshal input_data: %w", err)
		}

		_, err = tx.Exec(ctx, query,
			log.LogID.String(),
			log.ImportJobID.String(),
			log.RowNumber,
			string(log.Status),
			inputDataJSON,
			log.ErrorMessage,
		)
		if err != nil {
			return fmt.Errorf("failed to save import job log: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// FindByJobID finds all log entries for a job
func (r *ImportJobLogRepository) FindByJobID(ctx context.Context, jobID common.ImportJobID) ([]*importjob.ImportJobLog, error) {
	query := `
		SELECT log_id, import_job_id, row_number, status, input_data, error_message
		FROM import_job_logs
		WHERE import_job_id = $1
		ORDER BY row_number
	`

	return r.findLogs(ctx, query, jobID.String())
}

// FindErrorsByJobID finds only error log entries for a job
func (r *ImportJobLogRepository) FindErrorsByJobID(ctx context.Context, jobID common.ImportJobID) ([]*importjob.ImportJobLog, error) {
	query := `
		SELECT log_id, import_job_id, row_number, status, input_data, error_message
		FROM import_job_logs
		WHERE import_job_id = $1 AND status = 'error'
		ORDER BY row_number
	`

	return r.findLogs(ctx, query, jobID.String())
}

func (r *ImportJobLogRepository) findLogs(ctx context.Context, query string, args ...interface{}) ([]*importjob.ImportJobLog, error) {
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to find import job logs: %w", err)
	}
	defer rows.Close()

	var logs []*importjob.ImportJobLog
	for rows.Next() {
		var (
			logIDStr       string
			importJobIDStr string
			rowNumber      int
			statusStr      string
			inputDataJSON  []byte
			errorMessage   sql.NullString
		)

		err := rows.Scan(
			&logIDStr,
			&importJobIDStr,
			&rowNumber,
			&statusStr,
			&inputDataJSON,
			&errorMessage,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan import job log: %w", err)
		}

		logID, err := common.ParseImportLogID(logIDStr)
		if err != nil {
			return nil, err
		}

		importJobID, err := common.ParseImportJobID(importJobIDStr)
		if err != nil {
			return nil, err
		}

		var inputData map[string]interface{}
		if len(inputDataJSON) > 0 {
			if err := json.Unmarshal(inputDataJSON, &inputData); err != nil {
				return nil, fmt.Errorf("failed to unmarshal input_data: %w", err)
			}
		}

		logs = append(logs, &importjob.ImportJobLog{
			LogID:        logID,
			ImportJobID:  importJobID,
			RowNumber:    rowNumber,
			Status:       importjob.ImportLogStatus(statusStr),
			InputData:    inputData,
			ErrorMessage: errorMessage.String,
		})
	}

	return logs, nil
}
