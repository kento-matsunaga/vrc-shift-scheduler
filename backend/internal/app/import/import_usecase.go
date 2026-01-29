package importapp

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	importjob "github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/import"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/member"
)

// MemberRepository defines the interface for member persistence
type MemberRepository interface {
	Save(ctx context.Context, member *member.Member) error
	SaveBatch(ctx context.Context, members []*member.Member) error
	FindByTenantID(ctx context.Context, tenantID common.TenantID) ([]*member.Member, error)
	FindByDisplayName(ctx context.Context, tenantID common.TenantID, displayName string) (*member.Member, error)
}

// ImportMembersInput represents the input for importing members
type ImportMembersInput struct {
	TenantID common.TenantID
	AdminID  common.AdminID
	FileName string
	FileData []byte
	Options  importjob.ImportOptions
}

// ImportMembersOutput represents the output of member import
type ImportMembersOutput struct {
	ImportJobID  common.ImportJobID
	Status       importjob.ImportStatus
	TotalRows    int
	SuccessCount int
	ErrorCount   int
	Errors       []importjob.ErrorDetail
}

// ImportMembersUsecase handles the member import use case
type ImportMembersUsecase struct {
	importJobRepo importjob.ImportJobRepository
	memberRepo    MemberRepository
	csvParser     *importjob.CSVParser
}

// NewImportMembersUsecase creates a new ImportMembersUsecase
func NewImportMembersUsecase(
	importJobRepo importjob.ImportJobRepository,
	memberRepo MemberRepository,
) *ImportMembersUsecase {
	return &ImportMembersUsecase{
		importJobRepo: importJobRepo,
		memberRepo:    memberRepo,
		csvParser:     importjob.NewCSVParser(),
	}
}

// Execute imports members from CSV
func (uc *ImportMembersUsecase) Execute(ctx context.Context, input ImportMembersInput) (*ImportMembersOutput, error) {
	now := time.Now()

	// Create import job
	job, err := importjob.NewImportJob(
		now,
		input.TenantID,
		importjob.ImportTypeMembers,
		input.FileName,
		input.Options,
		input.AdminID,
	)
	if err != nil {
		return nil, err
	}

	// Save initial job
	if err := uc.importJobRepo.Save(ctx, job); err != nil {
		return nil, fmt.Errorf("failed to save import job: %w", err)
	}

	// Parse CSV
	reader := bytes.NewReader(input.FileData)
	rows, err := uc.csvParser.ParseMembersCSV(reader)
	if err != nil {
		_ = job.Fail(time.Now(), fmt.Sprintf("CSVパースエラー: %v", err))
		_ = uc.importJobRepo.Update(ctx, job)
		return &ImportMembersOutput{
			ImportJobID:  job.ImportJobID(),
			Status:       job.Status(),
			TotalRows:    0,
			SuccessCount: 0,
			ErrorCount:   1,
			Errors:       job.ErrorDetails(),
		}, nil
	}

	// Check row limit (max 10000 rows)
	const maxRows = 10000
	if len(rows) > maxRows {
		_ = job.Fail(time.Now(), fmt.Sprintf("行数が上限を超えています: %d行 (上限: %d行)", len(rows), maxRows))
		_ = uc.importJobRepo.Update(ctx, job)
		return &ImportMembersOutput{
			ImportJobID:  job.ImportJobID(),
			Status:       job.Status(),
			TotalRows:    len(rows),
			SuccessCount: 0,
			ErrorCount:   1,
			Errors:       job.ErrorDetails(),
		}, nil
	}

	// Start processing
	if err := job.Start(time.Now(), len(rows)); err != nil {
		return nil, err
	}
	if err := uc.importJobRepo.Update(ctx, job); err != nil {
		return nil, fmt.Errorf("failed to update import job: %w", err)
	}

	// Get existing members for duplicate check
	existingMembers, err := uc.memberRepo.FindByTenantID(ctx, input.TenantID)
	if err != nil {
		_ = job.Fail(time.Now(), fmt.Sprintf("既存メンバー取得エラー: %v", err))
		_ = uc.importJobRepo.Update(ctx, job)
		return &ImportMembersOutput{
			ImportJobID:  job.ImportJobID(),
			Status:       job.Status(),
			TotalRows:    len(rows),
			SuccessCount: 0,
			ErrorCount:   1,
			Errors:       job.ErrorDetails(),
		}, nil
	}

	// Build member matcher for duplicate check (with optional fuzzy matching)
	matcher := importjob.NewMemberMatcher(existingMembers, input.Options.FuzzyMemberMatch)

	// Build name index for tracking newly created members within this import
	existingNames := make(map[string]*member.Member)
	for _, m := range existingMembers {
		existingNames[m.DisplayName()] = m
	}

	// Collect new members for batch insert
	var newMembers []*member.Member

	// Process each row
	for _, row := range rows {
		// Validate row
		if err := row.Validate(); err != nil {
			job.RecordError(row.RowNumber, err.Error())
			continue
		}

		// Effective display name (already resolved by csv_parser)
		effectiveDisplayName := row.DisplayName

		// Check for duplicate using matcher (supports fuzzy matching)
		existing, _ := matcher.Match(effectiveDisplayName)
		// Also check newly created members in this import batch
		if existing == nil {
			existing = existingNames[effectiveDisplayName]
		}

		if existing != nil {
			if input.Options.SkipExisting {
				job.RecordSkip()
				continue
			}
			if input.Options.UpdateExisting {
				// Update existing member - mark as success (no actual changes for now)
				job.RecordSuccess()
				continue
			}
			// Neither skip nor update - record as error
			matchInfo := ""
			if input.Options.FuzzyMemberMatch && existing.DisplayName() != effectiveDisplayName {
				matchInfo = fmt.Sprintf(" (曖昧一致: '%s')", existing.DisplayName())
			}
			job.RecordError(row.RowNumber, fmt.Sprintf("メンバー '%s' は既に存在します%s", effectiveDisplayName, matchInfo))
			continue
		}

		// Create new member
		newMember, err := member.NewMember(
			time.Now(),
			input.TenantID,
			effectiveDisplayName,
			"", // discord_user_id - not used
			"", // email - not used
		)
		if err != nil {
			job.RecordError(row.RowNumber, fmt.Sprintf("メンバー作成エラー: %v", err))
			continue
		}

		// Add to batch for later insert
		newMembers = append(newMembers, newMember)

		// Add to existing names for duplicate detection within same import
		existingNames[effectiveDisplayName] = newMember
		job.RecordSuccess()
	}

	// Batch insert all new members
	if len(newMembers) > 0 {
		if err := uc.memberRepo.SaveBatch(ctx, newMembers); err != nil {
			_ = job.Fail(time.Now(), fmt.Sprintf("バッチ保存エラー: %v", err))
			_ = uc.importJobRepo.Update(ctx, job)
			return &ImportMembersOutput{
				ImportJobID:  job.ImportJobID(),
				Status:       job.Status(),
				TotalRows:    job.TotalRows(),
				SuccessCount: 0,
				ErrorCount:   job.ErrorCount() + 1,
				Errors:       job.ErrorDetails(),
			}, nil
		}
	}

	// Complete job
	if err := job.Complete(time.Now()); err != nil {
		return nil, fmt.Errorf("failed to complete import job: %w", err)
	}
	if err := uc.importJobRepo.Update(ctx, job); err != nil {
		return nil, fmt.Errorf("failed to update import job: %w", err)
	}

	return &ImportMembersOutput{
		ImportJobID:  job.ImportJobID(),
		Status:       job.Status(),
		TotalRows:    job.TotalRows(),
		SuccessCount: job.SuccessCount(),
		ErrorCount:   job.ErrorCount(),
		Errors:       job.ErrorDetails(),
	}, nil
}

// GetImportStatusInput represents the input for getting import status
type GetImportStatusInput struct {
	ImportJobID common.ImportJobID
	TenantID    common.TenantID
}

// GetImportStatusOutput represents the output of getting import status
type GetImportStatusOutput struct {
	ImportJobID   common.ImportJobID
	Status        importjob.ImportStatus
	ImportType    importjob.ImportType
	FileName      string
	TotalRows     int
	ProcessedRows int
	SuccessCount  int
	ErrorCount    int
	Progress      float64
	StartedAt     *time.Time
	CompletedAt   *time.Time
	CreatedAt     time.Time
}

// GetImportStatusUsecase handles getting import job status
type GetImportStatusUsecase struct {
	importJobRepo importjob.ImportJobRepository
}

// NewGetImportStatusUsecase creates a new GetImportStatusUsecase
func NewGetImportStatusUsecase(importJobRepo importjob.ImportJobRepository) *GetImportStatusUsecase {
	return &GetImportStatusUsecase{
		importJobRepo: importJobRepo,
	}
}

// Execute gets the status of an import job
func (uc *GetImportStatusUsecase) Execute(ctx context.Context, input GetImportStatusInput) (*GetImportStatusOutput, error) {
	job, err := uc.importJobRepo.FindByIDAndTenantID(ctx, input.ImportJobID, input.TenantID)
	if err != nil {
		return nil, err
	}

	return &GetImportStatusOutput{
		ImportJobID:   job.ImportJobID(),
		Status:        job.Status(),
		ImportType:    job.ImportType(),
		FileName:      job.FileName(),
		TotalRows:     job.TotalRows(),
		ProcessedRows: job.ProcessedRows(),
		SuccessCount:  job.SuccessCount(),
		ErrorCount:    job.ErrorCount(),
		Progress:      job.Progress(),
		StartedAt:     job.StartedAt(),
		CompletedAt:   job.CompletedAt(),
		CreatedAt:     job.CreatedAt(),
	}, nil
}

// GetImportResultInput represents the input for getting import result
type GetImportResultInput struct {
	ImportJobID common.ImportJobID
	TenantID    common.TenantID
}

// GetImportResultOutput represents the output of getting import result
type GetImportResultOutput struct {
	ImportJobID  common.ImportJobID
	Status       importjob.ImportStatus
	TotalRows    int
	SuccessCount int
	ErrorCount   int
	SkippedCount int
	Errors       []importjob.ErrorDetail
}

// GetImportResultUsecase handles getting import job result
type GetImportResultUsecase struct {
	importJobRepo importjob.ImportJobRepository
}

// NewGetImportResultUsecase creates a new GetImportResultUsecase
func NewGetImportResultUsecase(importJobRepo importjob.ImportJobRepository) *GetImportResultUsecase {
	return &GetImportResultUsecase{
		importJobRepo: importJobRepo,
	}
}

// Execute gets the result of an import job
func (uc *GetImportResultUsecase) Execute(ctx context.Context, input GetImportResultInput) (*GetImportResultOutput, error) {
	job, err := uc.importJobRepo.FindByIDAndTenantID(ctx, input.ImportJobID, input.TenantID)
	if err != nil {
		return nil, err
	}

	skippedCount := job.ProcessedRows() - job.SuccessCount() - job.ErrorCount()
	if skippedCount < 0 {
		skippedCount = 0
	}

	return &GetImportResultOutput{
		ImportJobID:  job.ImportJobID(),
		Status:       job.Status(),
		TotalRows:    job.TotalRows(),
		SuccessCount: job.SuccessCount(),
		ErrorCount:   job.ErrorCount(),
		SkippedCount: skippedCount,
		Errors:       job.ErrorDetails(),
	}, nil
}

// ListImportJobsInput represents the input for listing import jobs
type ListImportJobsInput struct {
	TenantID common.TenantID
	Limit    int
	Offset   int
}

// ListImportJobsOutput represents the output of listing import jobs
type ListImportJobsOutput struct {
	Jobs       []*GetImportStatusOutput
	TotalCount int
}

// ListImportJobsUsecase handles listing import jobs
type ListImportJobsUsecase struct {
	importJobRepo importjob.ImportJobRepository
}

// NewListImportJobsUsecase creates a new ListImportJobsUsecase
func NewListImportJobsUsecase(importJobRepo importjob.ImportJobRepository) *ListImportJobsUsecase {
	return &ListImportJobsUsecase{
		importJobRepo: importJobRepo,
	}
}

// Execute lists import jobs for a tenant
func (uc *ListImportJobsUsecase) Execute(ctx context.Context, input ListImportJobsInput) (*ListImportJobsOutput, error) {
	jobs, err := uc.importJobRepo.FindByTenantID(ctx, input.TenantID, input.Limit, input.Offset)
	if err != nil {
		return nil, err
	}

	count, err := uc.importJobRepo.CountByTenantID(ctx, input.TenantID)
	if err != nil {
		return nil, err
	}

	output := &ListImportJobsOutput{
		Jobs:       make([]*GetImportStatusOutput, len(jobs)),
		TotalCount: count,
	}

	for i, job := range jobs {
		output.Jobs[i] = &GetImportStatusOutput{
			ImportJobID:   job.ImportJobID(),
			Status:        job.Status(),
			ImportType:    job.ImportType(),
			FileName:      job.FileName(),
			TotalRows:     job.TotalRows(),
			ProcessedRows: job.ProcessedRows(),
			SuccessCount:  job.SuccessCount(),
			ErrorCount:    job.ErrorCount(),
			Progress:      job.Progress(),
			StartedAt:     job.StartedAt(),
			CompletedAt:   job.CompletedAt(),
			CreatedAt:     job.CreatedAt(),
		}
	}

	return output, nil
}

// Ensure io.Reader is used
var _ io.Reader = (*bytes.Reader)(nil)
